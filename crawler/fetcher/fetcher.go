package Fetcher

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
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
	_, err := url.Parse(URL)
	links := []string{}
	if err != nil {
		return "", links, err
	}

	page, err := http.Get(URL)
	if err != nil {
		return "", links, err
	}

	defer page.Body.Close()
	var buf bytes.Buffer
	tee := io.TeeReader(page.Body, &buf)
	text, err := ioutil.ReadAll(tee)

	if page.StatusCode >= 400 {
		return string(text), links, fmt.Errorf("%s %d", URL, page.StatusCode)
	}

	doc, err := html.Parse(page.Body)
	if err != nil {
		return string(text), links, fmt.Errorf("failed to parse: %v", err)
	}

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					links = append(links, a.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return string(text), links, nil
}
