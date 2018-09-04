// Copyright © 2018 Joe McMahon <joe.mcmahon@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/joemcmahon/joe_macmahon_technical_test/api/client"
	pb "github.com/joemcmahon/joe_macmahon_technical_test/api/crawl"
	"github.com/spf13/cobra"
)

const usage = `Usage client start <url>

Starts a crawl on the supplied URL; the URL is required.
`
const addr = "127.0.0.1:10000"

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start crawling a URL",
	Long: `Adds the URL to the crawl list and starts crawling it. The
crawl will continue until all URLS in this URL's domain reachable from
this root URL are visited, or the crawl is explicitly stopped.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println(usage)
			return
		}
		url := args[0]

		c := Client.New(addr)
		defer c.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		req := pb.URLRequest{URL: url, State: pb.URLRequest_START}
		state, err := c.CrawlSite(ctx, &req)
		if err != nil {
			fmt.Println("Failed to start crawl: %s", err.Error())
			return
		}
		fmt.Println(state.Status.String(), state.Message)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
