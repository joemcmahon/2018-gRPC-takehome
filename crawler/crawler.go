package Crawler

import (
	"errors"
	"sync"

	"github.com/disiqueira/gotree"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

// State is the current state of the crawler.
type State struct {
	BaseURL string
	cache   map[string]error
	tree    *gotree.Tree
	fetcher Fetcher
	debug   bool
	sync.Mutex
}

// Fetcher defines an interface that can fetch URLs.
type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// fetched tracks URLs that have been (or are being) fetched.
// The lock must be held while reading from or writing to the map.
// See http://golang.org/ref/spec#Struct_types section on embedded types.

var errLoading = errors.New("url load in progress") // sentinel value

// crawl uses Fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func crawl(url string, depth int, fetched *State, current *gotree.Tree) {
	// record url under current SiteTree. If none exists, create.
	fetched.Lock()
	var newT gotree.Tree
	if current == nil {
		newT = gotree.New(url)
		current = &newT
		fetched.tree = current
	} else {
		newT = (*current).Add(url)
	}
	fetched.Unlock()

	if depth <= 0 {
		log.Debugf("<- Done with %v, depth 0.\n", url)
		return
	}

	fetched.Lock()
	if _, ok := fetched.cache[url]; ok {
		fetched.Unlock()
		log.Debugf("<- Done with %v, already fetched.\n", url)
		return
	}
	// We mark the url to be loading to avoid others reloading it at the same time.
	fetched.cache[url] = errLoading
	fetched.Unlock()

	// We load it concurrently.
	body, urls, err := fetched.fetcher.Fetch(url)

	// And update the status in a synced zone.
	fetched.Lock()
	fetched.cache[url] = err
	fetched.Unlock()

	if err != nil {
		log.Debugf("<- Error on %v: %v\n", url, err)
		return
	}
	log.Debugf("Found: %s %q\n", url, body)
	done := make(chan bool)
	for i, u := range urls {
		log.Debugf("-> Crawling child %v/%v of %v : %v.\n", i, len(urls), url, u)
		go func(url string) {
			crawl(url, depth-1, fetched, &newT)
			done <- true
		}(u)
	}
	for i := range urls {
		log.Debugf("<- [%v] %v/%v Waiting for child\n", url, i, len(urls))
		<-done
	}
	log.Debugf("<- Done with %v\n", url)
}

// New takes a URL and a Fetcher and returns a State.
func New(baseURL string, fetcher Fetcher) State {
	s := State{
		BaseURL: baseURL,
		cache:   make(map[string]error),
		fetcher: fetcher,
	}
	return s
}

// Debug turns debug logging on or off.
func Debug(state bool) {
	if state == true {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.FatalLevel)
	}
}

// Format formats the crawl tree as it stands and returns it.
func (state *State) Format() string {
	if state == nil || state.tree == nil {
		log.Debug("tree is not initialized")
		return ""
	}
	return (*state.tree).Print()
}

// Run takes a URL and a Fetcher to fetch URLs, crawls the tree,
// holding the crawl state in the State pointer passed in.
func (state *State) Run() {
	// TODO: add purell to canonicize URLs, extract domain for regex
	//       check instead of depth check
	// TODO: Have crawl do the worker state check on launch as in
	// https://stackoverflow.com/questions/16101409/is-there-some-elegant-way-to-pause-resume-any-other-goroutine-in-golang
	crawl(state.BaseURL, 4, state, state.tree)
}
