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
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "php_opcache"
)

// Exporter configures and exposes PHP-FPM metrics to Prometheus.
type Exporter struct {
	mutex       sync.Mutex
	PoolManager PoolManager

	CountProcessState bool

	up                           *prometheus.Desc
	scrapeFailures               *prometheus.Desc
	OpcacheEnabled               *prometheus.Desc
	CacheFull                    *prometheus.Desc
	RestartPending               *prometheus.Desc
	RestartInProgress            *prometheus.Desc
	CacheUsedMemory              *prometheus.Desc
	CacheFreeMemory              *prometheus.Desc
	CacheWastedMemory            *prometheus.Desc
	CacheCurrentWastedPercentage *prometheus.Desc
	StringsBufferSize            *prometheus.Desc
	StringsUsedMemory            *prometheus.Desc
	StringsFreeMemory            *prometheus.Desc
	StringsNumberOfStrings       *prometheus.Desc
	StatNumCachedScripts         *prometheus.Desc
	StatNumCachedKeys            *prometheus.Desc
	StatMaxCachedKeys            *prometheus.Desc
	StatHits                     *prometheus.Desc
	StatStartTime                *prometheus.Desc
	StatLastRestartTime          *prometheus.Desc
	StatOomRestarts              *prometheus.Desc
	StatHashRestart              *prometheus.Desc
	StatManualRestarts           *prometheus.Desc
	StatMisses                   *prometheus.Desc
	StatBlacklistMisses          *prometheus.Desc
	StatBlacklistMissRatio       *prometheus.Desc
	StatOpcacheHitRate           *prometheus.Desc
}

// NewExporter creates a new Exporter for a PoolManager and configures the necessary metrics.
func NewExporter(pm PoolManager) *Exporter {
	return &Exporter{
		PoolManager: pm,

		CountProcessState: false,

		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could opcache be reached?",
			[]string{"pool"},
			nil),

		OpcacheEnabled: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "enabled"),
			"Is opcache enabled?",
			[]string{"pool"},
			nil),

		scrapeFailures: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "scrape_failures"),
			"The number of failures scraping from opcache.",
			[]string{"pool"},
			nil),

		CacheFull: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "cache_full"),
			"Is opcache full?",
			[]string{"pool"},
			nil),

		RestartPending: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "restart_penging"),
			"Pending opcache restart",
			[]string{"pool"},
			nil),

		RestartInProgress: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "restart_in_progress"),
			"Opcache restart in progress",
			[]string{"pool"},
			nil),

		CacheUsedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "used_memory"),
			"The number of bytes used by opcache",
			[]string{"pool"},
			nil),

		CacheFreeMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "free_memory"),
			"The number of free bytes for opcache",
			[]string{"pool"},
			nil),

		CacheWastedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "wasted_memory"),
			"Opcache wasted memory",
			[]string{"pool"},
			nil),

		CacheCurrentWastedPercentage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "current_wasted_percentage"),
			"Percentage of wasted memory",
			[]string{"pool"},
			nil),

		StringsBufferSize: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "interned_strings_buffer_size"),
			"interned_strings_buffer_size",
			[]string{"pool"},
			nil),

		StringsUsedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "interned_strings_used_memory"),
			"interned_strings_used_memory",
			[]string{"pool"},
			nil),

		StringsFreeMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "interned_strings_free_memory"),
			"interned_strings_free_memory",
			[]string{"pool"},
			nil),

		StringsNumberOfStrings: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "interned_strings_count"),
			"interned_strings_count",
			[]string{"pool"},
			nil),

		StatNumCachedScripts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "stat_num_cached_scripts"),
			"stat_num_cached_scripts",
			[]string{"pool"},
			nil),

		StatNumCachedKeys: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "stat_num_cached_keys"),
			"stat_num_cached_scripts",
			[]string{"pool"},
			nil),

		StatMaxCachedKeys: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "stat_max_cached_scripts"),
			"stat_num_cached_scripts",
			[]string{"pool"},
			nil),

		StatHits: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "stat_hits"),
			"stat_hits",
			[]string{"pool"},
			nil),

		StatStartTime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "stat_start_time"),
			"stat_start_time",
			[]string{"pool"},
			nil),

		StatLastRestartTime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "stat_last_restart_time"),
			"stat_last_restart_time",
			[]string{"pool"},
			nil),

		StatOomRestarts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "stat_oom_restarts"),
			"stat_oom_restarts",
			[]string{"pool"},
			nil),

		StatHashRestart: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "stat_hash_restarts"),
			"stat_hash_restarts",
			[]string{"pool"},
			nil),

		StatManualRestarts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "stat_manual_restarts"),
			"stat_manual_restarts",
			[]string{"pool"},
			nil),

		StatMisses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "stat_misses"),
			"stat_misses",
			[]string{"pool"},
			nil),

		StatBlacklistMisses: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "stat_blacklist_misses"),
			"stat_blacklist_misses",
			[]string{"pool"},
			nil),

		StatBlacklistMissRatio: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "stat_blacklist_miss_ratio"),
			"stat_blacklist_miss_ratio",
			[]string{"pool"},
			nil),

		StatOpcacheHitRate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "stat_hit_rate"),
			"stat_hit_rate",
			[]string{"pool"},
			nil),
	}
}

// Collect updates the Pools and sends the collected metrics to Prometheus
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.PoolManager.Update()

	for _, pool := range e.PoolManager.Pools {
		ch <- prometheus.MustNewConstMetric(e.scrapeFailures, prometheus.CounterValue, float64(pool.ScrapeFailures), pool.Name)

		if pool.ScrapeError != nil {
			ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0, pool.Name)
			log.Errorf("Error scraping PHP-FPM: %v", pool.ScrapeError)
			continue
		}
		ch <- prometheus.MustNewConstMetric(e.OpcacheEnabled, prometheus.GaugeValue, float64(bool2int(pool.OpcacheEnabled)), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.CacheFull, prometheus.GaugeValue, float64(bool2int(pool.CacheFull)), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.RestartPending, prometheus.GaugeValue, float64(bool2int(pool.RestartPending)), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.RestartInProgress, prometheus.GaugeValue, float64(bool2int(pool.RestartInProgress)), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.CacheUsedMemory, prometheus.GaugeValue, float64(pool.MemoryUsage.UsedMemory), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.CacheFreeMemory, prometheus.GaugeValue, float64(pool.MemoryUsage.FreeMemory), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.CacheWastedMemory, prometheus.GaugeValue, float64(pool.MemoryUsage.WastedMemory), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.CacheCurrentWastedPercentage, prometheus.GaugeValue, float64(pool.MemoryUsage.CurrentWastedPercentage), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StringsBufferSize, prometheus.GaugeValue, float64(pool.InternedStringsUsage.BufferSize), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StringsUsedMemory, prometheus.GaugeValue, float64(pool.InternedStringsUsage.UsedMemory), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StringsFreeMemory, prometheus.GaugeValue, float64(pool.InternedStringsUsage.FreeMemory), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StringsNumberOfStrings, prometheus.GaugeValue, float64(pool.InternedStringsUsage.NumberOfStrings), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StatNumCachedScripts, prometheus.GaugeValue, float64(pool.OpcacheStatistics.NumCachedScripts), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StatNumCachedKeys, prometheus.GaugeValue, float64(pool.OpcacheStatistics.NumCachedKeys), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StatMaxCachedKeys, prometheus.GaugeValue, float64(pool.OpcacheStatistics.MaxCachedKeys), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StatHits, prometheus.GaugeValue, float64(pool.OpcacheStatistics.Hits), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StatStartTime, prometheus.GaugeValue, float64(pool.OpcacheStatistics.StartTime), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StatLastRestartTime, prometheus.GaugeValue, float64(pool.OpcacheStatistics.LastRestartTime), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StatOomRestarts, prometheus.GaugeValue, float64(pool.OpcacheStatistics.OomRestarts), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StatHashRestart, prometheus.GaugeValue, float64(pool.OpcacheStatistics.HashRestart), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StatManualRestarts, prometheus.GaugeValue, float64(pool.OpcacheStatistics.ManualRestarts), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StatMisses, prometheus.GaugeValue, float64(pool.OpcacheStatistics.Misses), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StatBlacklistMisses, prometheus.GaugeValue, float64(pool.OpcacheStatistics.BlacklistMisses), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StatBlacklistMissRatio, prometheus.GaugeValue, float64(pool.OpcacheStatistics.BlacklistMissRatio), pool.Name)
		ch <- prometheus.MustNewConstMetric(e.StatOpcacheHitRate, prometheus.GaugeValue, float64(pool.OpcacheStatistics.OpcacheHitRate), pool.Name)

	}
}

// Describe exposes the metric description to Prometheus
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.OpcacheEnabled
	ch <- e.scrapeFailures
	ch <- e.CacheFull
	ch <- e.RestartPending
	ch <- e.RestartInProgress
	ch <- e.CacheUsedMemory
	ch <- e.CacheFreeMemory
	ch <- e.CacheWastedMemory
	ch <- e.CacheCurrentWastedPercentage
	ch <- e.StringsBufferSize
	ch <- e.StringsUsedMemory
	ch <- e.StringsFreeMemory
	ch <- e.StringsNumberOfStrings
	ch <- e.StatNumCachedScripts
	ch <- e.StatNumCachedKeys
	ch <- e.StatMaxCachedKeys
	ch <- e.StatHits
	ch <- e.StatStartTime
	ch <- e.StatLastRestartTime
	ch <- e.StatOomRestarts
	ch <- e.StatHashRestart
	ch <- e.StatManualRestarts
	ch <- e.StatMisses
	ch <- e.StatBlacklistMisses
	ch <- e.StatBlacklistMissRatio
	ch <- e.StatOpcacheHitRate
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}
