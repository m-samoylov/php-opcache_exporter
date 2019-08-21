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

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/m-samoylov/php-opcache_exporter/phpfpm"
	"github.com/spf13/cobra"
)

// Configuration variables
var (
	output string
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Returns opcache metrics without running as a server",
	Long: `"get" fetches opcache metrics from php-fpm. Multiple addresses can be specified as follows:

* php-opcache_exporter get --phpfpm.scrape-uri 127.0.0.1:9000 --phpfpm.scrape-uri 127.0.0.1:9001 [...]
* php-opcache_exporter get --phpfpm.scrape-uri 127.0.0.1:9000,127.0.0.1:9001,[...]
`,
	Run: func(cmd *cobra.Command, args []string) {
		pm := phpfpm.PoolManager{}

		for _, uri := range scrapeURIs {
			pm.Add(uri)
		}

		if err := pm.Update(); err != nil {
			log.Fatal("Could not update pool.", err)
		}

		switch output {
		case "json":
			content, err := json.Marshal(pm)
			if err != nil {
				log.Fatal("Cannot encode to JSON ", err)
			}
			fmt.Print(string(content))
		case "spew":
			spew.Dump(pm)
		default:
			log.Error("Output format not valid.")
		}
	},
}

func init() {
	RootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	getCmd.Flags().StringSliceVar(&scrapeURIs, "phpfpm.scrape-uri", []string{"tcp://127.0.0.1:9000/status"}, "FastCGI address, e.g. unix:///tmp/php.sock;/status or tcp://127.0.0.1:9000/status")
	getCmd.Flags().StringVar(&output, "out", "json", "Output format. One of: json, spew")
}
