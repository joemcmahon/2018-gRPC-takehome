package Server

import log "github.com/sirupsen/logrus"

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
			log.Infof("Start requested for running crawl on %s: no action", url)
			c.state[url] = running
		case done:
			log.Infof("Start requested for completed crawl on %s: last crawl discarded, starting crawl", url)
			c.state[url] = running
		case stopped:
			log.Infof("Start requested for stopped crawl on %s: resuming crawl", url)
			c.state[url] = running
		case failed:
			log.Infof("Start requested for failed crawl on %s: retrying crawl", url)
			c.state[url] = running
		default:
			log.Infof("%s in invalid state %s: forcing stop", url, translate(state))
			c.state[url] = stopped
		}
	} else {
		log.Infof("Start requested for new URL %s: starting crawl", url)
		c.state[url] = running
	}
}

// Stop stops a crawl for a URL.
func (c *CrawlServer) Stop(url string) {
	if state, ok := c.state[url]; ok {
		switch state {
		case running:
			log.Infof("Stop requested for running crawl on %s: stopped crawl", url)
			c.state[url] = stopped
		case done, stopped, failed:
			log.Infof("Stop requested for %s crawl on %s: no action", translate(state), url)
		default:
			log.Infof("%s in invalid state %s: forcing stop", url, translate(state))
			c.state[url] = stopped
		}
	} else {
		log.Infof("Stop requested for unknown URL %s: no action", url)
	}
}

// Done marks a crawl as done for a URL.
func (c *CrawlServer) Done(url string) {
	if state, ok := c.state[url]; ok {
		switch state {
		case running:
			log.Infof("Marking crawl of %s as done", url)
			c.state[url] = stopped
		default:
			log.Infof("Attempt to mark crawl of %s in state %s as done: no action", url, translate(state))
			c.state[url] = stopped
		}
	} else {
		log.Infof("Stop requested for unknown URL %s: no action", url)
	}
	c.state[url] = done
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
