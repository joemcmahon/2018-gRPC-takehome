package Server

import (
	"fmt"

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
func (c *CrawlServer) Start(url string) {
	if state, ok := c.state[url]; ok {
		switch state {
		case running:
			log.Infof(changeState(url, "running", "running", "no action"))
			c.state[url] = running
		case done:
			log.Infof(changeState(url, "done", "running", "last crawl discarded, restarting crawl"))
			c.state[url] = running
		case stopped:
			log.Infof(changeState(url, "stopped", "running", "resuming crawl"))
			c.state[url] = running
		case failed:
			log.Infof(changeState(url, "failed", "running", "retrying crawl"))
			c.state[url] = running
		default:
			log.Infof(changeState(url, "invalid state", "running", "forcing stop"))
			c.state[url] = stopped
		}
	} else {
		log.Infof(changeState(url, translate(unknown), "running", "starting crawl"))
		c.state[url] = running
	}
}

// Stop stops a crawl for a URL.
func (c *CrawlServer) Stop(url string) {
	if state, ok := c.state[url]; ok {
		switch state {
		case running:
			log.Infof(changeState(url, "running", "stopped", "crawl paused"))
			c.state[url] = stopped
		case done, stopped, failed:
			log.Infof(changeState(url, translate(state), "stopped", "no action"))
		default:
			log.Infof("%s in invalid state %s: forcing stop", url, translate(state))
			c.state[url] = stopped
		}
	} else {
		log.Infof(changeState(url, translate(unknown), "stopped", "no action"))
	}
}

// Done marks a crawl as done for a URL.
func (c *CrawlServer) Done(url string) {
	if state, ok := c.state[url]; ok {
		switch state {
		case running:
			log.Infof(changeState(url, "running", "done", "recording completed crawl"))
			c.state[url] = done
		default:
			log.Infof(changeState(url, translate(state), "done", "no action"))
		}
	} else {
		log.Infof(changeState(url, translate(unknown), "done", "no action"))
	}
}

// Failed marks a crawl as failed for a URL.
func (c *CrawlServer) Failed(url string) {
	if state, ok := c.state[url]; ok {
		switch state {
		case running:
			log.Infof(changeState(url, translate(state), "failed", "marked failed"))
			c.state[url] = failed
		default:
			log.Infof(changeState(url, translate(state), "failed", "no action"))
		}
	} else {
		log.Infof(changeState(url, translate(unknown), "failed", "no action"))
	}
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
