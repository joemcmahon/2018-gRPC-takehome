package MockFetcher

import (
	"fmt"
)

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

// MockFetcher is a URL fetcher that uses a data structure to
// simulate pulling pages off the web.
type MockFetcher struct {
	fake *fakeFetcher
}

// New creates a properly-initialized MockFetcher.
func New() *MockFetcher {
	m := MockFetcher{fake: fetcher}
	return &m
}

// Fetch looks up a URL in the MockFetcher's map.
// Returns a fake page and URLs from it if found, error if not.
func (m *MockFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := (*m.fake)[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = &fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
