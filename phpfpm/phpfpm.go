// Copyright © 2018 Enrico Stahn <enrico.stahn@gmail.com>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package phpfpm provides convenient access to PHP-FPM pool data
package phpfpm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	fcgiclient "github.com/tomasen/fcgi_client"
)

var log logger

type logger interface {
	Info(ar ...interface{})
	Infof(string, ...interface{})
	Debug(ar ...interface{})
	Debugf(string, ...interface{})
	Error(ar ...interface{})
	Errorf(string, ...interface{})
}

// PoolManager manages all configured Pools
type PoolManager struct {
	Pools []Pool
}

type MemoryUsage struct {
	UsedMemory              uint64 `json:"used_memory"`
	FreeMemory              uint64 `json:"free_memory"`
	WastedMemory            uint64 `json:"wasted_memory"`
	CurrentWastedPercentage int    `json:"current_wasted_percentage"`
}

type InternedStringsUsage struct {
	BufferSize      int `json:"buffer_size"`
	UsedMemory      int `json:"used_memory"`
	FreeMemory      int `json:"free_memory"`
	NumberOfStrings int `json:"number_of_strings"`
}

type OpcacheStatistics struct {
	NumCachedScripts   int `json:"num_cached_scripts"`
	NumCachedKeys      int `json:"num_cached_keys"`
	MaxCachedKeys      int `json:"max_cached_keys"`
	Hits               int `json:"hits"`
	StartTime          int `json:"start_time"`
	LastRestartTime    int `json:"last_restart_time"`
	OomRestarts        int `json:"oom_restarts"`
	HashRestart        int `json:"hash_restarts"`
	ManualRestarts     int `json:"manual_restarts"`
	Misses             int `json:"misses"`
	BlacklistMisses    int `json:"blacklist_misses"`
	BlacklistMissRatio int `json:"blacklist_miss_ratio"`
	OpcacheHitRate     int `json:"opcache_hit_rate"`
}

// Pool describes a single PHP-FPM pool that can be reached via a Socket or TCP address
type Pool struct {
	// The address of the pool, e.g. tcp://127.0.0.1:9000 or unix:///tmp/php-fpm.sock
	Address              string `json:"-"`
	ScrapeError          error  `json:"-"`
	ScrapeFailures       int64  `json:"-"`
	Name                 string
	OpcacheEnabled       bool                 `json:"opcache_enabled"`
	CacheFull            bool                 `json:"cache_full"`
	RestartPending       bool                 `json:"restart_pending"`
	RestartInProgress    bool                 `json:"restart_in_progress"`
	MemoryUsage          MemoryUsage          `json:"memory_usage"`
	InternedStringsUsage InternedStringsUsage `json:"interned_string_usage"`
	OpcacheStatistics    OpcacheStatistics    `json:"opcache_statistics"`
}

type FPMPool struct {
	// The address of the pool, e.g. tcp://127.0.0.1:9000 or unix:///tmp/php-fpm.sock
	Address             string           `json:"-"`
	ScrapeError         error            `json:"-"`
	ScrapeFailures      int64            `json:"-"`
	Name                string           `json:"pool"`
	ProcessManager      string           `json:"process manager"`
	StartTime           timestamp        `json:"start time"`
	StartSince          int64            `json:"start since"`
	AcceptedConnections int64            `json:"accepted conn"`
	ListenQueue         int64            `json:"listen queue"`
	MaxListenQueue      int64            `json:"max listen queue"`
	ListenQueueLength   int64            `json:"listen queue len"`
	IdleProcesses       int64            `json:"idle processes"`
	ActiveProcesses     int64            `json:"active processes"`
	TotalProcesses      int64            `json:"total processes"`
	MaxActiveProcesses  int64            `json:"max active processes"`
	MaxChildrenReached  int64            `json:"max children reached"`
	SlowRequests        int64            `json:"slow requests"`
	Processes           []FPMPoolProcess `json:"processes"`
}

type FPMPoolProcess struct {
	PID               int64   `json:"pid"`
	State             string  `json:"state"`
	StartTime         int64   `json:"start time"`
	StartSince        int64   `json:"start since"`
	Requests          int64   `json:"requests"`
	RequestDuration   int64   `json:"request duration"`
	RequestMethod     string  `json:"request method"`
	RequestURI        string  `json:"request uri"`
	ContentLength     int64   `json:"content length"`
	User              string  `json:"user"`
	Script            string  `json:"script"`
	LastRequestCPU    float64 `json:"last request cpu"`
	LastRequestMemory int64   `json:"last request memory"`
}

// Add will add a pool to the pool manager based on the given URI.
func (pm *PoolManager) Add(uri string) Pool {
	p := Pool{Address: uri}
	pm.Pools = append(pm.Pools, p)
	return p
}

// Update will run the pool.Update() method concurrently on all Pools.
func (pm *PoolManager) Update() (err error) {
	wg := &sync.WaitGroup{}

	started := time.Now()

	for idx := range pm.Pools {
		wg.Add(1)
		go func(p *Pool) {
			defer wg.Done()
			p.Update()
		}(&pm.Pools[idx])
	}

	wg.Wait()

	ended := time.Now()

	log.Debugf("Updated %v pool(s) in %v", len(pm.Pools), ended.Sub(started))

	return nil
}

// Update will connect to PHP-FPM and retrieve the latest data for the pool.
func (p *Pool) Update() (err error) {
	p.ScrapeError = nil

	scheme, address, _, err := parseURL(p.Address)
	log.Debugf("%v, %v", scheme, address)
	if err != nil {
		return p.error(err)
	}

	fcgi, err := fcgiclient.DialTimeout(scheme, address, time.Duration(3)*time.Second)
	if err != nil {
		return p.error(err)
	}

	defer fcgi.Close()

	load := []byte("<?php echo json_encode(opcache_get_status())?>")
	tmpfile, err := ioutil.TempFile("", "php-opcache_exporter-*.php")
	if err != nil {
		return p.error(err)
	}

	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(load); err != nil {
		return p.error(err)
	}
	if err := tmpfile.Close(); err != nil {
		return p.error(err)
	}

	env := map[string]string{
		"SCRIPT_FILENAME": tmpfile.Name(),
		"SCRIPT_NAME":     tmpfile.Name(),
		"SERVER_SOFTWARE": "go / php-opcache_exporter",
		"REMOTE_ADDR":     "127.0.0.1",
	}

	resp, err := fcgi.Get(env)
	if err != nil {
		return p.error(err)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return p.error(err)
	}

	defer resp.Body.Close()

	log.Debugf("Pool[%v]: %v", p.Address, string(content))

	if err = json.Unmarshal(content, &p); err != nil {
		log.Errorf("Pool[%v]: %v", p.Address, string(content))
		return error(err)
	}

	p.GetPoolName()

	return nil
}

func (p *Pool) error(err error) error {
	p.ScrapeError = err
	p.ScrapeFailures++
	log.Error(err)
	return err
}

func (p *Pool) GetPoolName() (err error) {
	var fpm FPMPool
	scheme, address, path, err := parseURL(p.Address)

	if err != nil {
		return p.error(err)
	}

	fcgi, err := fcgiclient.DialTimeout(scheme, address, time.Duration(3)*time.Second)
	if err != nil {
		return p.error(err)
	}

	defer fcgi.Close()

	env_status := map[string]string{
		"SCRIPT_FILENAME": path,
		"SCRIPT_NAME":     path,
		"SERVER_SOFTWARE": "go / php-opcache_exporter",
		"REMOTE_ADDR":     "127.0.0.1",
		"QUERY_STRING":    "json&full",
	}
	resp, err := fcgi.Get(env_status)
	if err != nil {
		return p.error(err)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return p.error(err)
	}

	defer resp.Body.Close()
	content = JSONResponseFixer(content)
	if err = json.Unmarshal(content, &fpm); err != nil {
		log.Errorf("Pool[%v]: %v", p.Address, string(content))
		return error(err)
	}

	p.Name = fpm.Name
	return nil
}

// JSONResponseFixer resolves encoding issues with PHP-FPMs JSON response
func JSONResponseFixer(content []byte) []byte {
	c := string(content)
	re := regexp.MustCompile(`(,"request uri":)"(.*?)"(,"content length":)`)
	matches := re.FindAllStringSubmatch(c, -1)

	for _, match := range matches {
		requestURI, _ := json.Marshal(match[2])

		sold := match[0]
		snew := match[1] + string(requestURI) + match[3]

		c = strings.Replace(c, sold, snew, -1)
	}

	return []byte(c)
}

// parseURL creates elements to be passed into fcgiclient.DialTimeout
func parseURL(rawurl string) (scheme string, address string, path string, err error) {
	uri, err := url.Parse(rawurl)
	if err != nil {
		return uri.Scheme, uri.Host, uri.Path, err
	}

	scheme = uri.Scheme

	switch uri.Scheme {
	case "unix":
		result := strings.Split(uri.Path, ";")
		address = result[0]
		if len(result) > 1 {
			path = result[1]
		}
	default:
		address = uri.Host
		path = uri.Path
	}

	return
}

type timestamp time.Time

// MarshalJSON customise JSON for timestamp
func (t *timestamp) MarshalJSON() ([]byte, error) {
	ts := time.Time(*t).Unix()
	stamp := fmt.Sprint(ts)
	return []byte(stamp), nil
}

// UnmarshalJSON customise JSON for timestamp
func (t *timestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}
	*t = timestamp(time.Unix(int64(ts), 0))
	return nil
}

// SetLogger configures the used logger
func SetLogger(logger logger) {
	log = logger
}
