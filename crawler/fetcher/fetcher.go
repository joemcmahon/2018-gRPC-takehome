package Fetcher

import (
	"net/url"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

// Fetcher is a URL fetcher that takes a URL (as a string),
// fetches the web page corresponding to it, and returns
// a list of the URLs on the page (as strings) and any HTTP
// error occurring while trying to follow the link, parse the HTML, etc.
type Fetcher struct {
}

// New creates a properly-initialized Fetcher.
func New() *Fetcher {
	f := Fetcher{}
	return &f
}

// Fetch actually does all the work.
func (m *Fetcher) Fetch(URL string) (string, []string, error) {
	u, err := url.Parse(URL)
	links := []string{}
	if err != nil {
		return "", links, err
	}

	var text string

	// Set up scraper
	c := colly.NewCollector(
		colly.AllowedDomains(u.Host),
	)

	// Capture all the body text
	c.OnHTML("body", func(e *colly.HTMLElement) {
		text = e.Text
	})
	// Extract links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		links = append(links, e.Attr("href"))
	})
	// Log a debug message for each page visit
	c.OnRequest(func(r *colly.Request) {
		log.Debugf("VISIT> %s", r.URL.String())
	})
	// Actually do it.
	c.Visit(URL)

	return string(text), links, nil
}
