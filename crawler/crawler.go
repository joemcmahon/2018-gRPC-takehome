package Crawler

import (
	"errors"
	"net/url"
	"os"
	"sync"

	"github.com/PuerkitoBio/purell"
	"github.com/disiqueira/gotree"
	log "github.com/sirupsen/logrus"
)

func init() {
	if os.Getenv("TESTING") != "" {
		log.SetLevel(log.DebugLevel)
	}
}

// State is the current state of the crawler.
type State struct {
	BaseURL  string
	domain   string
	cache    map[string]error
	tree     *gotree.Tree
	fetcher  Fetcher
	debug    bool
	done     bool
	failed   bool
	runState int
	control  chan int
	sync.Mutex
}

// Fetcher defines an interface that can fetch URLs.
type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Use this value to mark a URL as busy in the URL cache.
var errLoading = errors.New("url load in progress")

// Use this value to mark a URL as leading offsite
var errOffsite = errors.New("url points offsite")

// Run denotes the state in which we're actively crawling.
const Run = 0

// Pause denotes the state is which we're not crawling and
// are waiting to either be stopped or running again.
const Pause = 1

// Stop denotes the state in which all crawling stops for this URL.
const Stop = 2

// stoppedOrCrawling hides the details of the run/pause/stop
// mechanism from the crawl function.
func (state *State) stoppedOrCrawling() int {
	log.Debug("decide if stopped or crawling")
	state.Lock()
	defer (state.Unlock)()

	if state.runState != Pause {
		log.Debug("in proper active state")
		return state.runState
	}

	// Paused. Trap us inside this function until we receive
	// a new state.
	for {
		log.Debug("paused, wait for signal")
		// Wait until we receive a control signal.
		newState := <-state.control
		log.Debug("received signal", newState)
		if newState != Pause {
			// Either stopped or started; return new status
			// and stop monitoring.
			log.Debug("unpaused")
			state.runState = newState
			return newState
		}
		// Still paused. Go back and wait for another signal.
		log.Debug("still paused")
	}
}

// crawl uses Fetcher to recursively crawl pages starting with URL.
// Once all links that point to the same domain as the initial URL
// have been visited, crawling stops.
func (state *State) crawl(URL string, current *gotree.Tree) {
	// record URL under current SiteTree. If none exists, create.
	// If we detect that we're paused inside stoppedOrCrawling(), we
	// won't exit it until we are moved to stopped or running.
	log.Debug("Checking run state")
	if state.stoppedOrCrawling() == Stop {
		// Stopped. No more crawling, and skip this one too.
		log.Debug("terminating crawl due to stop")
		state.Unlock()
		return
	}

	// Skip empty links.
	if URL == "" {
		return
	}

	// Otherwise, we're running. Record this one in the tree.
	log.Debug("crawl running for ", URL)

	// See if this URL is valid.
	u, err := url.Parse(URL)
	if err != nil {
		state.Lock()
		state.cache[URL] = err
		state.Unlock()
		log.Debugf("Invalid URL %s: %s", URL, err.Error())
	}

	// We want to record the link, even if it's bad.
	state.Lock()
	var newT gotree.Tree
	if current == nil {
		// Tree's empty; build a new one.
		newT = gotree.New(URL)
		current = &newT
		state.tree = current
	} else {
		// Add this URL under the current node.
		newT = (*current).Add(URL)
	}
	state.Unlock()

	if err != nil {
		// Recorded. No further action.
		return
	}

	// Relative URLs are still on this domain. Make it so.
	if string(URL[0]) == "/" {
		// Make the next test know it lives here.
		u.Host = state.domain
		URL = u.String()
		URL, _ = purify(URL)
	}

	// Are we off our domain?
	if u.Host != state.domain {
		state.Lock()
		state.cache[URL] = errOffsite
		state.Unlock()
		log.Debugf("Offsite URL %s", URL)
		return
	}

	state.Lock()
	if _, ok := state.cache[URL]; ok {
		state.Unlock()
		log.Debugf("<- Done with %v, already fetched.\n", URL)
		return
	}
	// We mark the URL to be loading to avoid others reloading it at the same time.
	state.cache[URL] = errLoading
	state.Unlock()

	// We load it concurrently.
	body, urls, err := state.fetcher.Fetch(URL)

	// And update the status in a synced zone.
	state.Lock()
	state.cache[URL] = err
	state.Unlock()

	if err != nil {
		log.Debugf("<- Error on %v: %v\n", URL, err)
		return
	}
	log.Debugf("Found: %s %q\n", URL, body)
	done := make(chan bool)
	for i, u := range urls {
		// Ignoring the error because fetched URLs should already be
		// valid URLs of some sort.
		u, _ = purify(u)
		log.Debugf("-> Crawling child %v/%v of %v : %v.\n", i, len(urls), URL, u)
		go func(URL string) {
			state.crawl(URL, &newT)
			done <- true
		}(u)
	}
	for i := range urls {
		log.Debugf("<- [%v] %v/%v Waiting for child\n", URL, i, len(urls))
		<-done
	}
	log.Debugf("<- Done with %v\n", URL)
}

func purify(URL string) (string, error) {
	return purell.NormalizeURLString(URL,
		purell.FlagsAllNonGreedy&^purell.FlagRemoveDirectoryIndex&^purell.FlagForceHTTP&^purell.FlagAddWWW)
}

// New takes a URL and a Fetcher and returns a State.
func New(baseURL string, fetcher Fetcher) State {
	b, err := purify(baseURL)
	s := State{
		BaseURL:  b,
		cache:    make(map[string]error),
		fetcher:  fetcher,
		runState: Run,
		control:  make(chan int),
	}
	if err != nil {
		// bad initial URL. fail crawl right away.
		s.BaseURL = baseURL

		s.failed = true
		s.done = true
	}
	// purify() will have returned a valid URL.
	u, _ := url.Parse(baseURL)
	s.domain = u.Host
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

// Pause temporarily halts the crawl.
func (state *State) Pause() {
	state.control <- Pause
}

// Start resumes the crawl.
func (state *State) Start() {
	state.control <- Run
}

// Stop cancels the crawl.
func (state *State) Stop() {
	state.control <- Stop
}

// IsDone lets external entites safely check to see if the crawl is done.
func (state *State) IsDone() bool {
	if state == nil {
		return false
	}
	state.Lock()
	defer (state.Unlock)()
	log.Debug("check if done")
	switch {
	case state.done:
		log.Debug("done")
		return true
	case state.failed:
		log.Debug("failed, force done")
		state.failed = true
		return true
	default:
		log.Debug("not done")
		return false
	}
}

// HasFailed lets external entities see if the crawl failed.
func (state *State) HasFailed() bool {
	if state == nil {
		return false
	}
	state.Lock()
	defer (state.Unlock)()
	log.Debug("check if failed")
	switch {
	case state.failed:
		log.Debug("failed")
		return true
	default:
		log.Debug("still ok")
		return false
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
	state.Lock()
	state.cache = make(map[string]error)
	state.tree = nil
	state.Unlock()

	if os.Getenv("TESTING") != "" {
		Debug(true)
	}
	if state == nil {
		return
	}
	// Run the crawl asynchronously; when it terminates, set the done flag to true.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Debug("crawl panicked:", r)
				state.Lock()
				state.failed = true
				state.Unlock()
			}
		}()
		log.Debug("launch crawl")
		state.crawl(state.BaseURL, state.tree)
		log.Debug("Crawl complete")
		state.Lock()
		state.done = true
		state.Unlock()
	}()
}
