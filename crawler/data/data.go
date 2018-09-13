package data

import (
	"github.com/disiqueira/gotree"
	"github.com/joemcmahon/joe_macmahon_technical_test/crawler/cache"
	siteTree "github.com/joemcmahon/joe_macmahon_technical_test/crawler/shared-tree"
)

// Fetcher defines an interface that can fetch URLs.
type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// WorkRequest defines a unit of work to be done
type WorkRequest struct {
	URL    string
	Domain string
	Cache  cache.Cache
	Tree   siteTree.Tree
	Root   *gotree.Tree
}
