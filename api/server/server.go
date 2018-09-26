package Server

import (
	"context"
	"fmt"
	"strings"
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
	crawler *crawler.State
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
	crawlers map[string]CrawlControl
}

// New creates and returns an empty CrawlServer.
func New(f Fetcher) *CrawlServer {
	return &CrawlServer{
		f:        f,
		crawlers: make(map[string]CrawlControl),
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
	if state, ok := c.crawlers[url]; ok {
		log.Debug("executing for", url)
		newState = state
		switch state.State {
		case running:
			status = c.changeState(url, "running", "running", "no action")
		case done:
			status = c.changeState(url, "done", "running", "last crawl discarded, restarting crawl")
			newState.crawler = crawler.New(url, c.f)
			newState.crawler.Start()
			newState.State = running
		case stopped:
			status = c.changeState(url, "stopped", "running", "resuming crawl")
			if newState.crawler != nil {
				newState.crawler.Resume()
				newState.State = running
			}
		case failed:
			status = c.changeState(url, "failed", "running", "retrying crawl")
			if newState.crawler != nil {
				newState.crawler.Start()
				newState.State = running
			}
		default:
			// This would be an entry in state 'unknown', which should not be possible.
			panic(c.changeState(url, "invalid state", "running", "panic!"))
		}
	} else {
		// Actually start a new crawl
		log.Debug("Start crawl")
		status = c.changeState(url, translate(unknown), "running", "starting crawl")
		newState.crawler = crawler.New(url, c.f)
		newState.crawler.Start()
		newState.State = running
	}
	c.crawlers[url] = newState
	log.Infof(status)
	return status, c.crawlers[url].State, err
}

// Pause pauses a crawl for a URL.
func (c *CrawlServer) Pause(url string) (string, CrawlState, error) {
	var status string
	var err error

	c.mutex.Lock()
	defer (c.mutex.Unlock)()

	if newState, ok := c.crawlers[url]; ok {
		switch newState.State {
		case running:
			status = c.changeState(url, "running", "stopped", "crawl paused")
			newState.State = stopped
			if newState.crawler != nil {
				newState.crawler.Pause()
			}
		case done, stopped, failed:
			status = c.changeState(url, translate(newState.State), "stopped", "no action")
		default:
			// This would be an entry in state 'unknown', which should not be possible.
			panic(c.changeState(url, "invalid state", "running", "panic!"))
		}
	} else {
		status = c.changeState(url, translate(unknown), "stopped", "no action")
	}
	log.Infof(status)
	return status, c.crawlers[url].State, err
}

// Probe checks the current state of a crawl without changing anything.
func (c *CrawlServer) Probe(url string) string {
	c.mutex.Lock()
	defer (c.mutex.Unlock)()

	if crawlerState, ok := c.crawlers[url]; ok {
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
	if state, ok := c.crawlers[url]; ok {
		switch state.State {
		case running:
			state.crawler.Lock()
			display = state.crawler.Format()
			state.crawler.Unlock()
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

func (c *CrawlServer) changeState(url, old, new, result string) string {
	z := c.crawlers[url]
	z.State = saveableState(new)
	c.crawlers[url] = z
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
		status, state, err = c.Pause(req.URL)

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

var saveable = map[string]CrawlState{
	"stopped": stopped,
	"running": running,
	"done":    done,
	"unknown": unknown,
	"failed":  failed,
}

func saveableState(state string) CrawlState {
	return saveable[state]
}

// CrawlResult sends the status of a given URL back over gRPC.
func (c *CrawlServer) CrawlResult(req *crawl.URLRequest, stream crawl.Crawl_CrawlResultServer) error {
	status := c.Probe(req.URL)
	result := c.Show(req.URL)
	for _, s := range strings.Split(result, "\n") {
		n := crawl.SiteNode{SiteURL: req.URL, TreeString: s, Status: status}
		if fail := stream.Send(&n); fail != nil {
			return fail
		}
	}
	return nil
}
