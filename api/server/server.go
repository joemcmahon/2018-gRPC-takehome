package Server

import (
	"context"
	"fmt"
	"sync"

	"github.com/joemcmahon/joe_macmahon_technical_test/api/crawl"
	"github.com/joemcmahon/joe_macmahon_technical_test/crawler"
	log "github.com/sirupsen/logrus"
)

// Define a dummy crawl server for the first iteration.
// This server will simply maintain state and not actually
// crawl anything; once the gRPC is working, we will add the
// Crawler package to allow us to do actual crawling.

// CrawlState defines the crawler states we'll maintain.
type CrawlState int

const (
	stopped CrawlState = iota
	running
	done
	unknown
	failed
)

// CrawlControl is the struct that minds a particular crawler.
type CrawlControl struct {
	State   CrawlState
	crawler *Crawler.State
}

// Fetcher defines an interface that can fetch URLs.
type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// CrawlServer defines the struct that holds the status of crawls
type CrawlServer struct {
	mutex sync.Mutex
	f     Fetcher
	// Crawler state for each URL
	state map[string]CrawlControl
}

// New creates and returns an empty CrawlServer.
func New(f Fetcher) *CrawlServer {
	return &CrawlServer{
		f:     f,
		state: make(map[string]CrawlControl),
	}
}

// Start starts a crawl for a URL.
func (c *CrawlServer) Start(url string) (string, CrawlState, error) {
	var status string
	var err error

	c.mutex.Lock()
	defer (c.mutex.Unlock)()
	log.Debug("selecting command")

	var newState CrawlControl
	log.Debug("Checking for over")
	c.checkForCrawlDoneOrFailed(url)
	log.Debug("not over")
	if state, ok := c.state[url]; ok {
		log.Debug("executing for", url)
		newState := state
		switch state.State {
		case running:
			status = changeState(url, "running", "running", "no action")
		case done:
			status = changeState(url, "done", "running", "last crawl discarded, restarting crawl")
			if newState.crawler != nil {
				newState.crawler.Run()
			} else {
				c := Crawler.New(url, c.f)
				newState.crawler = &c
				c.Run()
			}
			newState.State = running
		case stopped:
			status = changeState(url, "stopped", "running", "resuming crawl")
			if newState.crawler != nil {
				newState.crawler.Start()
				newState.State = running
			}
		case failed:
			status = changeState(url, "failed", "running", "retrying crawl")
			if newState.crawler != nil {
				newState.crawler.Run()
				newState.State = running
			}
		default:
			// This would be an entry in state 'unknown', which should not be possible.
			panic(changeState(url, "invalid state", "running", "panic!"))
		}
	} else {
		// Actually start a new crawl
		log.Debug("Start crawl")
		c := Crawler.New(url, c.f)
		newState.crawler = &c
		c.Run()
		status = changeState(url, translate(unknown), "running", "starting crawl")
		newState.State = running
	}
	c.state[url] = newState
	log.Debug("updating to", newState)
	log.Infof(status)
	return status, c.state[url].State, err
}

// Stop stops a crawl for a URL.
func (c *CrawlServer) Stop(url string) (string, CrawlState, error) {
	var status string
	var err error

	c.mutex.Lock()
	defer (c.mutex.Unlock)()

	var newState CrawlControl
	c.checkForCrawlDoneOrFailed(url)
	if newState, ok := c.state[url]; ok {
		switch newState.State {
		case running:
			status = changeState(url, "running", "stopped", "crawl paused")
			newState.State = stopped
			if newState.crawler != nil {
				newState.crawler.Stop()
			}
		case done, stopped, failed:
			status = changeState(url, translate(newState.State), "stopped", "no action")
		default:
			// This would be an entry in state 'unknown', which should not be possible.
			panic(changeState(url, "invalid state", "running", "panic!"))
		}
	} else {
		status = changeState(url, translate(unknown), "stopped", "no action")
	}
	log.Infof(status)
	c.state[url] = newState
	return status, c.state[url].State, err
}

// Probe checks the current state of a crawl without changing anything.
func (c *CrawlServer) Probe(url string) string {
	c.mutex.Lock()
	defer (c.mutex.Unlock)()

	log.Debug("probe state")
	c.checkForCrawlDoneOrFailed(url)
	log.Debug("not failed")
	if crawlerState, ok := c.state[url]; ok {
		return translate(crawlerState.State)
	}
	return translate(unknown)
}

// Show translates the crawl tree into a string and returns it.
// XXX: Note that this forces the output into a fixed format,
//      but since this is for the CLI, we can live with it for now.
//      Otherwise we need to extend the gotree interface and add
//      a custom formatter for JSON or whatever. (YAGNI)
func (c *CrawlServer) Show(url string) string {
	c.mutex.Lock()
	defer (c.mutex.Unlock)()

	var display string
	log.Debug("check state")
	c.checkForCrawlDoneOrFailed(url)
	log.Debug("not done")

	if state, ok := c.state[url]; ok {
		switch state.State {
		case running:
			// Stop it, run the formatter, start it.
			// Note we don't have to lock it while formatting
			// because the goroutine is paused, waiting for
			// the Start() call to resume.
			state.crawler.Pause()
			display = state.crawler.Format()
			state.crawler.Start()
		case stopped, done:
			display = state.crawler.Format()
		case failed:
			display = "Crawl failed; no valid results to show"
		}
	} else {
		// Unknown, so we've done nothing with it.
		return fmt.Sprintf("%s has not been crawled", url)
	}
	return display
}

var xlate = map[CrawlState]string{
	stopped: "stopped",
	running: "running",
	done:    "done",
	unknown: "unknown",
	failed:  "failed",
}

func translate(state CrawlState) string {
	return xlate[state]
}

func changeState(url, old, new, result string) string {
	return fmt.Sprintf("Change %s in state %s to %s: %s", url, old, new, result)
}

// CrawlSite starts, stops, or checks the status of a site.
func (c *CrawlServer) CrawlSite(ctx context.Context, req *crawl.URLRequest) (*crawl.URLState, error) {
	var status string
	var state CrawlState
	var err error

	switch req.State {
	case crawl.URLRequest_START:
		status, state, err = c.Start(req.URL)

	case crawl.URLRequest_STOP:
		status, state, err = c.Stop(req.URL)

	case crawl.URLRequest_CHECK:
		status = c.Probe(req.URL)
	}

	s := crawl.URLState{
		Status:  sendableState(state),
		Message: status,
	}
	return &s, err
}

var sendable = map[CrawlState]crawl.URLState_Status{
	stopped: crawl.URLState_STOPPED,
	running: crawl.URLState_RUNNING,
	unknown: crawl.URLState_UNKNOWN,
}

func sendableState(state CrawlState) crawl.URLState_Status {
	return sendable[state]
}

// CrawlResult sends the status of a given URL back over gRPC.
func (c *CrawlServer) CrawlResult(ctx context.Context, req *crawl.URLRequest) (*crawl.SiteNode, error) {
	status := c.Probe(req.URL)
	s := crawl.SiteNode{SiteURL: req.URL, TreeString: c.Show(req.URL), Status: status}
	return &s, nil
}

func (c *CrawlServer) checkForCrawlDoneOrFailed(url string) {
	// ONLY called after we've grabbed the mutex, so we don't grab it again.
	// We might get called for a URL that isn't there yet, so skip out in that case.
	if state, ok := c.state[url]; ok {
		if state.crawler.HasFailed() {
			state.State = failed
		} else if state.crawler.IsDone() {
			state.State = done
		}
		c.state[url] = state
	}
}
