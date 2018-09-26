package crawler

import (
	"errors"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/PuerkitoBio/purell"
	"github.com/disiqueira/gotree"
	"github.com/golang-collections/go-datastructures/queue"
	log "github.com/sirupsen/logrus"
)

func init() {
	if os.Getenv("TESTING") != "" {
		log.SetLevel(log.DebugLevel)
	}
}

type unprocessedItem struct {
	insertPoint *gotree.Tree
	URL         string
}

type controlFunc func()

// State is the current state of the crawler.
type State struct {
	BaseURL     string
	domain      string
	cache       map[string]error
	tree        *gotree.Tree
	fetcher     Fetcher
	debug       bool
	Done        bool
	unprocessed *queue.Queue
	Start       controlFunc
	Pause       controlFunc
	Resume      controlFunc
	Wait        controlFunc
	Quit        controlFunc

	sync.Mutex
}

// Fetcher defines an interface that can fetch URLs.
type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

const queueSize = 200

// Use this value to mark a URL as busy in the URL cache.
var errLoading = errors.New("url load in progress")

// Use this value to mark a URL as leading offsite
var errOffsite = errors.New("url points offsite")

func (state *State) crawlPage() {
	if state.unprocessed.Empty() {
		state.Quit()
	}
	z, _ := state.unprocessed.Get(1)
	item := z[0].(unprocessedItem)
	state.crawl(item.URL, item.insertPoint)
}

// crawl uses Fetcher to recursively crawl pages starting with URL.
// Once all links that point to the same domain as the initial URL
// have been visited, crawling stops.
func (state *State) crawl(URL string, current *gotree.Tree) {
	// record URL under current SiteTree. If none exists, create.
	// If we detect that we're paused inside stoppedOrCrawling(), we
	// won't exit it until we are moved to stopped or running.

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
	for i, u := range urls {
		// Ignoring the error because fetched URLs should already be
		// valid URLs of some sort.
		u, _ = purify(u)
		log.Debugf("-> Queuingchild %v/%v of %v : %v.\n", i, len(urls), URL, u)
		state.unprocessed.Put(unprocessedItem{URL: u, insertPoint: &newT})
	}
	log.Debugf("<- Done with %v\n", URL)
}

func purify(URL string) (string, error) {
	return purell.NormalizeURLString(URL,
		purell.FlagsAllNonGreedy&^purell.FlagRemoveDirectoryIndex&^purell.FlagForceHTTP&^purell.FlagAddWWW)
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

// New takes a URL and a Fetcher to fetch URLs.
// It initializes the crawler's data structures and returns a set of closures
// that can be used to start, pause, resume, and quit crawling. It also
// provides a wait() function to ensure that we can wait for the process to
// complete if we so desire. See https://stackoverflow.com/questions/38798863/golang-pause-a-loop-in-a-goroutine-with-channels
// holding the crawl state in the State pointer passed in.
func New(URL string, f Fetcher) *State {
	state := State{
		cache:       make(map[string]error),
		tree:        nil,
		fetcher:     f,
		unprocessed: queue.New(queueSize),
	}
	b, err := purify(URL)
	if err != nil {
		// bad initial URL. fail crawl right away.
		state.BaseURL = URL
		state.Done = true
		return &state
	}
	state.BaseURL = b
	// purify() will have returned a valid URL.
	u, _ := url.Parse(b)
	state.domain = u.Host

	state.unprocessed.Put(unprocessedItem{URL: URL, insertPoint: nil})

	if os.Getenv("TESTING") != "" {
		Debug(true)
	}
	state.Start, state.Pause, state.Resume, state.Quit, state.Wait = state.controls()
	return &state
}

// controls() controls the run/pause behavior for the crawl. It
// returns the controller functions needed to actually do the
// control operations on the crawl.
func (state *State) controls() (start, pause, resume, quit, wait func()) {
	var (
		chWork       <-chan struct{}
		chWorkBackup <-chan struct{}
		chControl    chan struct{}
		wg           sync.WaitGroup
	)

	// Routine encapsulates the logic to run one iteration of the
	// crawl, with run/pause controls.
	routine := func() {
		// Defer this so that if we quit, the waitgroup is closed out.
		defer wg.Done()

		for {
			select {
			case <-chWork:
				// crawl another URL, putting its sub-URLs on the queue,
				// then release the CPU.
				// If the queue is empty, crawlPage will quit().
				state.crawlPage()
				time.Sleep(100 * time.Millisecond)
			case _, ok := <-chControl:
				if ok {
					continue
				}
				return
			}
		}
	}

	start = func() {
		// chWork, chWorkBackup: two closed channels to
		// force a return when the read is done.
		ch := make(chan struct{})
		close(ch)
		chWork = ch
		chWorkBackup = ch

		// chControl is used to actually control whether we
		// run more goroutines.
		chControl = make(chan struct{})

		// wg
		wg = sync.WaitGroup{}
		wg.Add(1)

		// Run one more iteration of the crawl. Any URLs
		// found will be queued to be processed on the next
		// go-round.
		go routine()
	}

	pause = func() {
		// Used to disable the case that actually does work.
		// (Read from a nil channel in a select case causes
		// that case to be skipped.)
		chWork = nil
		chControl <- struct{}{}
	}

	resume = func() {
		// Restore the channel to re-enable the case.
		chWork = chWorkBackup
		chControl <- struct{}{}
	}

	quit = func() {
		// Read on a nil channel forces a return.
		chWork = nil
		close(chControl)
		state.Lock()
		state.Done = true
		state.Unlock()
	}

	wait = func() {
		// Wait for all operations to cease.
		wg.Wait()
	}

	return
}
