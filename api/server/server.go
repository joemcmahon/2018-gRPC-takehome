package Server

import (
	"context"
	"fmt"
	"sync"

	"github.com/joemcmahon/joe_macmahon_technical_test/api/crawl"
	"github.com/joemcmahon/joe_macmahon_technical_test/crawler"
	"github.com/joemcmahon/joe_macmahon_technical_test/crawler/test/mock_fetcher"
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

// CrawlServer defines the struct that holds the status of crawls
type CrawlServer struct {
	mutex sync.Mutex
	// Crawler state for each URL
	state map[string]CrawlControl
}

// New creates and returns an empty CrawlServer.
func New() *CrawlServer {
	return &CrawlServer{
		state: make(map[string]CrawlControl),
	}
}

// Start starts a crawl for a URL.
func (c *CrawlServer) Start(url string) (string, CrawlState, error) {
	var status string
	var err error

	c.mutex.Lock()
	defer (c.mutex.Unlock)()

	var newState CrawlControl
	c.checkForCrawlDoneOrFailed(url)
	if state, ok := c.state[url]; ok {
		newState := state
		switch state.State {
		case running:
			status = changeState(url, "running", "running", "no action")
		case done:
			status = changeState(url, "done", "running", "last crawl discarded, restarting crawl")
			newState.crawler.Run()
			newState.State = running
		case stopped:
			status = changeState(url, "stopped", "running", "resuming crawl")
			newState.crawler.Start()
			newState.State = running
		case failed:
			status = changeState(url, "failed", "running", "retrying crawl")
			newState.crawler.Run()
			newState.State = running
		default:
			// This would be an entry in state 'unknown', which should not be possible.
			panic(changeState(url, "invalid state", "running", "panic!"))
		}
	} else {
		// Actually start a new crawl
		f := MockFetcher.New()
		c := Crawler.New(url, f)
		newState.crawler = &c
		status = changeState(url, translate(unknown), "running", "starting crawl")
		newState.State = running
	}
	c.state[url] = newState
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
	if state, ok := c.state[url]; ok {
		switch state.State {
		case running:
			status = changeState(url, "running", "stopped", "crawl paused")
			newState.State = stopped
			newState.crawler.Stop()
		case done, stopped, failed:
			status = changeState(url, translate(state.State), "stopped", "no action")
		default:
			// This would be an entry in state 'unknown', which should not be possible.
			panic(changeState(url, "invalid state", "running", "panic!"))
		}
	} else {
		status = changeState(url, translate(unknown), "stopped", "no action")
	}
	log.Infof(status)
	return status, c.state[url].State, err
}

// Done marks a crawl as done for a URL.The Crawler will be given a pointer to
// the CrawlServer so that it can call Done when it finishes.
func (c *CrawlServer) Done(url string) (string, CrawlState) {
	var status string

	c.mutex.Lock()
	defer (c.mutex.Unlock)()

	var newState CrawlControl
	c.checkForCrawlDoneOrFailed(url)
	if state, ok := c.state[url]; ok {
		switch state.State {
		case running:
			status = changeState(url, "running", "done", "recording completed crawl")
			newState.State = done
		default:
			status = changeState(url, translate(state.State), "done", "no action")
		}
	} else {
		status = changeState(url, translate(unknown), "done", "no action")
	}
	log.Infof(status)
	c.state[url] = newState
	return status, c.state[url].State
}

// Failed marks a crawl as failed for a URL. Also a callback from Run, but handled
// by a panic trap.
func (c *CrawlServer) Failed(url string) (string, CrawlState) {
	var status string

	c.mutex.Lock()
	defer (c.mutex.Unlock)()

	var newState CrawlControl
	c.checkForCrawlDoneOrFailed(url)
	if state, ok := c.state[url]; ok {
		switch state.State {
		case running:
			status = changeState(url, translate(state.State), "failed", "marked failed")
			newState.State = failed
		default:
			status = changeState(url, translate(state.State), "failed", "no action")
		}
	} else {
		// This would be an entry in state 'unknown', which should not be possible.
		panic(changeState(url, "invalid state", "failed", "panic!"))
	}
	log.Infof(status)
	c.state[url] = newState
	return status, c.state[url].State
}

// Probe checks the current state of a crawl without changing anything.
func (c *CrawlServer) Probe(url string) string {
	c.mutex.Lock()
	defer (c.mutex.Unlock)()

	c.checkForCrawlDoneOrFailed(url)
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
	c.checkForCrawlDoneOrFailed(url)

	if state, ok := c.state[url]; ok {
		switch state.State {
		case running:
			// Stop it, run the formatter, start it.
			c.Stop(url)
			display = "look! halted to show a formatted tree!"
			c.Start(url)
		case stopped, done:
			display = "look! formatted the resulting tree!"
		case failed:
			display = "Crawl failed; no valid results to show"
		}
	} else {
		// Unknown, so we've done nothing with it.
		return "%s has not been crawled"
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
	done:    crawl.URLState_DONE,
	unknown: crawl.URLState_UNKNOWN,
	failed:  crawl.URLState_FAILED,
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
