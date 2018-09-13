package worker

import (
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/purell"
	"github.com/joemcmahon/joe_macmahon_technical_test/crawler/data"
	log "github.com/sirupsen/logrus"
)

// Worker represents an individual worker process. It uses the global
// work queue to find work and to queue more work, it uses the worker
// queue to mark itself as ready for work, and it uses the done queue
// to let the dispatcher know a work unit is complete. The fetcher is
// passed from the dispatcher so that the worker knows how to fetch a URL.
type Worker struct {
	ID          int
	Work        chan data.WorkRequest
	WorkerQueue chan chan data.WorkRequest
	Quit        chan bool
	DoneQueue   chan bool
	Fetcher     *data.Fetcher
}

var errLoading = fmt.Errorf("loading")
var errOffsite = fmt.Errorf("offsite")

// Start the worker by starting a goroutine that's
// an infinite "for-select" loop.
func (w *Worker) Start() {
	go func() {
		for {
			w.WorkerQueue <- w.Work
			select {
			case work := <-w.Work:
				// Receive a work request.
				fmt.Printf("worker%d: Received work request\n", w.ID)
				w.Process(work)
				fmt.Printf("worker%d: Complete\n", w.ID)
				w.DoneQueue <- true

			case <-w.Quit:
				// We have been asked to stop.
				fmt.Printf("worker%d stopping\n", w.ID)
				return
			}
		}
	}()
}

// Process actually stores the URL in the tree, fetches it if possible,
// stores it in the cache, and then schedules crawls for the URLs in the
// fetched page.
func (w *Worker) Process(req data.WorkRequest) {
	// Skip empty URLs.
	if req.URL == "" {
		return
	}

	log.Debug("crawl running for ", req.URL)

	// We want to record the link in the site tree, even if it's bad.
	newRoot := req.Tree.AddAt(req.Root, req.URL)

	// See if this URL is valid.
	u, err := url.Parse(req.URL)
	if err != nil {
		log.Debugf("Invalid URL %s: %s", req.URL, err.Error())
		return
	}

	// Set host to current domain if this is a relative URL.
	if string(req.URL[0]) == "/" {
		u.Host = req.Domain
		req.URL = u.String()
		req.URL, _ = purify(req.URL)
	}

	// Are we off our domain?
	if u.Host != req.Domain {
		req.Cache.Add(req.URL, errOffsite)
		log.Debugf("Offsite URL %s", req.URL)
		return
	}

	if req.Cache.Check(req.URL) {
		log.Debugf("<- Done with %v, already fetched.\n", req.URL)
		return
	}

	// We mark the URL to be loading to avoid others reloading it at the same time.
	req.Cache.Add(req.URL, errLoading)

	// We load it concurrently.
	body, urls, err := (*w.Fetcher).Fetch(req.URL)

	// And update the status in a synced zone.
	req.Cache.Add(req.URL, err)

	if err != nil {
		log.Debugf("<- Error on %v: %v\n", req.URL, err)
		return
	}
	log.Debugf("Found: %s %q\n", req.URL, body)

	for i, u := range urls {
		// Ignoring the error because fetched URLs should already be
		// valid URLs of some sort.
		u, _ = purify(u)
		log.Debugf("-> Crawling child %v/%v of %v : %v.\n", i, len(urls), req.URL, u)
		new := data.WorkRequest{
			URL:    u,
			Domain: req.Domain,
			Cache:  req.Cache,
			Tree:   req.Tree,
			Root:   newRoot,
		}
		w.Work <- new
	}
	log.Debugf("<- Done with %v\n", req.URL)
}

// Stop tells the worker to stop listening for work requests.
// Note that the worker will only stop *after* it has finished its work.
func (w *Worker) Stop() {
	go func() {
		w.Quit <- true
	}()
}

func purify(URL string) (string, error) {
	return purell.NormalizeURLString(URL,
		purell.FlagsAllNonGreedy&^purell.FlagRemoveDirectoryIndex&^purell.FlagForceHTTP&^purell.FlagAddWWW)
}
