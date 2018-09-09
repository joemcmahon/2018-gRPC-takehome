package Crawler

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joemcmahon/joe_macmahon_technical_test/crawler/test/mock_fetcher"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const fourOhFour = "http://example.com\n"
const knownURL = "http://golang.org/"
const unknownURL = "http://example.com"
const expectedStart = `http://golang.org/
├── http://golang.org/cmd/
└── http://golang.org/pkg/`

var _ = Describe("crawler", func() {
	if os.Getenv("TESTING") != "" {
		Debug(true)
	}
	Describe("have data", func() {
		state := runCrawler(knownURL)
		answer := state.Format()
		testPrint(answer)
		Context("scanning data we have", func() {
			It("scans the fake tree successfully", func() {
				// ExpectedStart here because the crawl items are
				// not always going to come back in the same order
				// past the top level.
				Expect(strings.HasPrefix(answer, expectedStart)).To(BeTrue())
			})
		})
	})
	Describe("don't have data", func() {
		state := runCrawler(unknownURL)
		answer := state.Format()
		testPrint(answer)
		Context("scanning data we don't have", func() {
			It("shows the empty tree as we expect it", func() {
				Expect(answer).To(Equal("http://example.com\n"))
			})
		})
	})
})

func testPrint(s string) {
	if os.Getenv("TESTING") != "" {
		fmt.Println(s)
	}
}

func runCrawler(url string) State {
	f := MockFetcher.New()
	state := New(url, f)
	state.Run()
	timeout := make(chan bool, 0)
	go func() {
		time.Sleep(10 * time.Second)
		timeout <- true
	}()
	select {
	case <-state.done:
		return state
	case <-state.failed:
		panic("Failed to crawl!")
	case <-timeout:
		panic("crawl timed out!")
	}
	return state
}
