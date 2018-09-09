package Crawler

import (
	"errors"
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
	cache    map[string]error
	tree     *gotree.Tree
	fetcher  Fetcher
	debug    bool
	done     chan bool
	failed   chan bool
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

// crawl uses Fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func (state *State) crawl(url string, depth int, current *gotree.Tree) {
	// record url under current SiteTree. If none exists, create.
	// If we detect that we're paused inside stoppedOrCrawling(), we
	// won't exit it until we are moved to stopped or running.
	log.Debug("Checking run state")
	if state.stoppedOrCrawling() == Stop {
		// Stopped. No more crawling, and skip this one too.
		log.Debug("terminating crawl due to stop")
		state.Unlock()
		return
	}

	// Otherwise, we're running. Record this one in the tree.
	log.Debug("crawl running for ", url)

	state.Lock()
	var newT gotree.Tree
	if current == nil {
		// Tree's empty; build a new one.
		newT = gotree.New(url)
		current = &newT
		state.tree = current
	} else {
		// Add this URL under the current node.
		newT = (*current).Add(url)
	}
	state.Unlock()

	// See if this URL goes offsite.
	if depth <= 0 {
		log.Debugf("<- Done with %v, depth 0.\n", url)
		return
	}

	state.Lock()
	if _, ok := state.cache[url]; ok {
		state.Unlock()
		log.Debugf("<- Done with %v, already fetched.\n", url)
		return
	}
	// We mark the url to be loading to avoid others reloading it at the same time.
	state.cache[url] = errLoading
	state.Unlock()

	// We load it concurrently.
	body, urls, err := state.fetcher.Fetch(url)

	// And update the status in a synced zone.
	state.Lock()
	state.cache[url] = err
	state.Unlock()

	if err != nil {
		log.Debugf("<- Error on %v: %v\n", url, err)
		return
	}
	log.Debugf("Found: %s %q\n", url, body)
	done := make(chan bool)
	for i, u := range urls {
		// Ignoring the error because fetched URLs should already be
		// valid URLs of some sort.
		u, _ = purify(u)
		log.Debugf("-> Crawling child %v/%v of %v : %v.\n", i, len(urls), url, u)
		go func(url string) {
			state.crawl(url, depth-1, &newT)
			done <- true
		}(u)
	}
	for i := range urls {
		log.Debugf("<- [%v] %v/%v Waiting for child\n", url, i, len(urls))
		<-done
	}
	log.Debugf("<- Done with %v\n", url)
}

func purify(url string) (string, error) {
	return purell.NormalizeURLString(url,
		purell.FlagsAllNonGreedy&^purell.FlagRemoveDirectoryIndex&^purell.FlagForceHTTP&^purell.FlagAddWWW)
}

// New takes a URL and a Fetcher and returns a State.
func New(baseURL string, fetcher Fetcher) State {
	b, err := purify(baseURL)
	s := State{
		BaseURL:  b,
		cache:    make(map[string]error),
		fetcher:  fetcher,
		done:     make(chan bool),
		failed:   make(chan bool),
		runState: Run,
		control:  make(chan int),
	}
	if err != nil {
		// bad initial URL. fail crawl right away.
		s.BaseURL = baseURL
		s.failed <- true
		s.done <- true
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

// Stop temporarily halts the crawl.
func (state *State) Stop() {
	log.Debug("stop attempted")
}

// Start resumes the crawl. (Note that Run implies a Start.)
func (state *State) Start() {
	log.Debug("start attempted")
}

// IsDone lets external entites safely check to see if the crawl is done.
func (state *State) IsDone() bool {
	if state == nil {
		return false
	}
	state.Lock()
	defer (state.Unlock)()
	log.Debug("check if done")
	select {
	case <-state.done:
		log.Debug("done")
		return true
	case <-state.failed:
		log.Debug("failed, force done")
		state.failed <- true
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
	select {
	case <-state.failed:
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
	// TODO: add purell to canonicize URLs, extract domain for regex
	//       check instead of depth check
	// TODO: Have crawl do the worker state check on launch as in
	// https://stackoverflow.com/questions/16101409/is-there-some-elegant-way-to-pause-resume-any-other-goroutine-in-golang
	Debug(true)
	if state == nil {
		return
	}
	// Run the crawl asynchronously; when it terminates, set the done flag to true.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Debug("crawl panicked:", r)
				state.failed <- true
			}
		}()
		log.Debug("launch crawl")
		state.crawl(state.BaseURL, 4, state.tree)
		log.Debug("Crawl complete")
		state.done <- true
	}()
}
