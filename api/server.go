package Server

import (
	"context"
	"fmt"
	"sync"

	"github.com/joemcmahon/joe_macmahon_technical_test/api/crawl"
	log "github.com/sirupsen/logrus"
)

// Define a dummy crawl server for the first iteration.
// This server will simply maintain state and not actually
// crawl anything; once the gRPC is working, we will add the
// Crawler package to allow us to do actual crawling.

// Define the crawler states we'll maintain.
type crawlState int

const (
	stopped crawlState = iota
	running
	done
	unknown
	failed
)

// CrawlServer defines the struct that holds the status of crawls
type CrawlServer struct {
	mutex sync.Mutex
	// Crawler state for each URL
	state map[string]crawlState
}

// New creates and returns an empty CrawlServer.
func New() *CrawlServer {
	return &CrawlServer{
		state: make(map[string]crawlState),
	}
}

// Start starts a crawl for a URL.
func (c *CrawlServer) Start(url string) string {
	var status string

	if state, ok := c.state[url]; ok {
		switch state {
		case running:
			status = changeState(url, "running", "running", "no action")
			c.state[url] = running
		case done:
			status = changeState(url, "done", "running", "last crawl discarded, restarting crawl")
			c.state[url] = running
		case stopped:
			status = changeState(url, "stopped", "running", "resuming crawl")
			c.state[url] = running
		case failed:
			status = changeState(url, "failed", "running", "retrying crawl")
			c.state[url] = running
		default:
			status = changeState(url, "invalid state", "running", "forcing stop")
			c.state[url] = stopped
		}
	} else {
		status = changeState(url, translate(unknown), "running", "starting crawl")
		c.state[url] = running
	}
	log.Infof(status)
	return status
}

// Stop stops a crawl for a URL.
func (c *CrawlServer) Stop(url string) string {
	var status string

	if state, ok := c.state[url]; ok {
		switch state {
		case running:
			status = changeState(url, "running", "stopped", "crawl paused")
			c.state[url] = stopped
		case done, stopped, failed:
			status = changeState(url, translate(state), "stopped", "no action")
		default:
			status = fmt.Sprintf("%s in invalid state %s: forcing stop", url, translate(state))
			c.state[url] = stopped
		}
	} else {
		status = changeState(url, translate(unknown), "stopped", "no action")
	}
	log.Infof(status)
	return status
}

// Done marks a crawl as done for a URL.
func (c *CrawlServer) Done(url string) string {
	var status string

	if state, ok := c.state[url]; ok {
		switch state {
		case running:
			status = changeState(url, "running", "done", "recording completed crawl")
			c.state[url] = done
		default:
			status = changeState(url, translate(state), "done", "no action")
		}
	} else {
		status = changeState(url, translate(unknown), "done", "no action")
	}
	log.Infof(status)
	return status
}

// Failed marks a crawl as failed for a URL.
func (c *CrawlServer) Failed(url string) string {
	var status string

	if state, ok := c.state[url]; ok {
		switch state {
		case running:
			status = changeState(url, translate(state), "failed", "marked failed")
			c.state[url] = failed
		default:
			status = changeState(url, translate(state), "failed", "no action")
		}
	} else {
		status = changeState(url, translate(unknown), "failed", "no action")
	}
	log.Infof(status)
	return status
}

// Probe checks the current state of a crawl without changing anything.
func (c *CrawlServer) Probe(url string) string {
	if crawlerState, ok := c.state[url]; ok {
		return translate(crawlerState)
	}
	return translate(unknown)
}

// Show translates the crawl tree into a string and returns it.
// XXX: Note that this forces the output into a fixed format,
//      but since this is for the CLI, we can live with it for now.
//      Otherwise we need to extend the gotree interface and add
//      a custom formatter for JSON or whatever. (YAGNI)
func (c *CrawlServer) Show(url string) string {
	return "(nothing to show yet)"
}

var xlate = map[crawlState]string{
	stopped: "stopped",
	running: "running",
	done:    "done",
	unknown: "unknown",
	failed:  "failed",
}

func translate(state crawlState) string {
	return xlate[state]
}

func changeState(url, old, new, result string) string {
	return fmt.Sprintf("Change %s in state %s to %s: %s", url, old, new, result)
}

func (c *CrawlServer) CrawlSite(ctx context.Context, req *crawl.URLRequest) (*crawl.URLState, error) {
	var status string

	switch req.State {
	case crawl.URLRequest_START:
		status, err = c.Start(req.URL)

	case crawl.URLRequest_STOP:
		status, err = c.Start(req.URL)

	case crawl.URLRequest_CHECK:
		status, err = c.Probe(req.URL)
	}

	return status, err
}

func (c *CrawlServer) URLStatus(ctx context.Context, req *crawl.URLRequest) (*crawl.SiteNode, error) {
	status, err := c.Probe(req.URL)
	s := crawl.SiteNode{SiteURL: req.URL, TreeString: c.Show(req.URL), Status: status}
	return &s, err
}
