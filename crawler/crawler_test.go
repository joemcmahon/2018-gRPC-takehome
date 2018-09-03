package Crawler

import (
	"fmt"
	"os"
	"strings"

	"github.com/joemcmahon/joe_macmahon_technical_test/crawler/test/mock_fetcher"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

const fourOhFour = "http://example.com\n"
const knownURL = "http://golang.org/"
const unknownURL = "http://example.com"

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
				Expect(answer).ToNot(Equal(""))
			})
			It("looks about right", func() {
				Expect(strings.HasPrefix(answer, "http://golang.org/\n├── http://golang")).To(BeTrue())
			})
		})
	})
	Describe("don't have data", func() {
		state := runCrawler(unknownURL)
		answer := state.Format()
		testPrint(answer)
		Context("scanning data we don't have", func() {
			It("shows the empty tree as we expect it", func() {
				Expect(answer).To(Equal(fourOhFour))
			})
		})
	})
})

func testPrint(s string) {
	if os.Getenv("TESTING") != "" {
		fmt.Println(s)
	}
}

func init() {
	log.SetLevel(log.DebugLevel)
}

func runCrawler(url string) State {
	f := MockFetcher.New()
	state := New(url, f)
	state.Run()
	return state
}
