package Dispatcher

import (
	"fmt"
	"net/url"

	"github.com/disiqueira/gotree"
	urlCache "github.com/joemcmahon/joe_macmahon_technical_test/crawler/cache"
	"github.com/joemcmahon/joe_macmahon_technical_test/crawler/data"
	siteTree "github.com/joemcmahon/joe_macmahon_technical_test/crawler/shared-tree"
	"github.com/joemcmahon/joe_macmahon_technical_test/crawler/worker"
	log "github.com/sirupsen/logrus"
)

// ControlValue is the type of constants sent to the control channel to pause and resume crawls.
type ControlValue int

const (
	// pause is used to pause the crawl
	pause ControlValue = iota
	// resume is used to continue the crawl.
	resume
	// quit quits crawling and terminates all workers.
	quit
)

// Dispatcher holds the work queue, workers, and items active for this
// particular dispatcher.
type Dispatcher struct {
	WorkQueue   chan data.WorkRequest
	WorkerQueue chan chan data.WorkRequest
	DoneQueue   chan bool
	itemsActive int
	control     chan ControlValue
	paused      bool
	workers     []worker.Worker
}

// StartWork initiates a crawl with the proper data
func (d *Dispatcher) StartWork(URL string) {
	log.Debug("Creating shared data processes")
	c := urlCache.New()
	t := siteTree.New()

	u, err := url.Parse(URL)
	if err != nil {
		log.Debugf("Invalid URL %s: %s", URL, err.Error())
	}

	log.Debugf("Start crawl for %s", URL)
	d.AddWork(URL, u.Host, c, t, nil)
	log.Debugf("Started")
}

// AddWork adds further URLs to be crawled to an active crawl.
// We pass along the cache for this crawl, the site tree we're building,
// and the address of the node we last inserted.
func (d *Dispatcher) AddWork(URL string, domain string, cache urlCache.Cache, tree siteTree.Tree, root *gotree.Tree) {
	work := data.WorkRequest{
		URL:    URL,
		Domain: domain,
		Cache:  cache,
		Tree:   tree,
		Root:   root,
	}

	// Push the work onto the queue.
	log.Debug("Add new work item: %s", URL)
	d.WorkQueue <- work
	log.Debug("Scheduled")
	return
}

// NewWorker creates and returns a new Worker object.
func (d *Dispatcher) NewWorker(id int) worker.Worker {
	// Create, and return the worker.
	worker := worker.Worker{
		ID:          id,
		Work:        make(chan data.WorkRequest),
		WorkerQueue: d.WorkerQueue,
		DoneQueue:   d.DoneQueue,
	}

	return worker
}

// Start creates the work queue for the work requests, starts the workers.
func Start(nworkers int, nrequests int) Dispatcher {
	log.Debug("Create dispatcher")
	d := Dispatcher{
		WorkerQueue: make(chan chan data.WorkRequest, nworkers),
		WorkQueue:   make(chan data.WorkRequest, nrequests),
		DoneQueue:   make(chan bool),
	}

	for i := 1; i <= nworkers; i++ {
		log.Debug("Starting worker", i)
		worker := d.NewWorker(i)
		worker.Start()
		d.workers = append(d.workers, worker)
	}

	log.Debug("Launch dispatcher")
	go func() {
		for {
			if d.paused {
				log.Debug("Paused - wait for event")
				if d.pausedDispatchHasQuit() {
					return
				}
			} else {
				log.Debug("Running - wait for event")
				if d.runningDispatchHasQuit() {
					return
				}
			}
		}
	}()
	return d
}

func (d *Dispatcher) pausedDispatchHasQuit() bool {
	select {
	case <-d.DoneQueue:
		log.Debug("Another item completed")
		// you're welcome!
		d.itemsActive--
		log.Debug("Items left: %d", d.itemsActive)
	case ctrl := <-d.control:
		switch ctrl {
		case pause:
			log.Debug("Pause signal detected, no action")
			d.paused = true
		case resume:
			log.Debug("Resume signal detected, resume crawl")
			d.paused = false
		case quit:
			log.Debug("Quit signal detected, exit")
			d.stopWorkers()
			return true
		}
	}
	return false
}

func (d *Dispatcher) runningDispatchHasQuit() bool {
	select {
	case work := <-d.WorkQueue:
		fmt.Println("Received work request")
		go func() {
			// Find an available worker on the worker list
			worker := <-d.WorkerQueue
			fmt.Println("Dispatching work request")
			// Send work unit to worker
			worker <- work

		}()
	case <-d.DoneQueue:
		// you're welcome!
		d.itemsActive--
	case ctrl := <-d.control:
		switch ctrl {
		case pause:
			log.Debug("Pause signal detected, pause execution")
			d.paused = true
		case resume:
			log.Debug("Resume signal detected, no action")
			d.paused = false
		case quit:
			log.Debug("Quit signal detected, exit")
			d.stopWorkers()
			return true
		}
	}
	return false
}

// PauseCrawl tells the dispatcher to stop scheduling work until a resume is seen.
func (d *Dispatcher) PauseCrawl() {
	d.control <- pause
}

// ResumeCrawl tells the dispatcher to schedule work until complete or a pause is seen
func (d *Dispatcher) ResumeCrawl() {
	d.control <- resume
}

// Done checks to see if this crawl has any more work to do.
func (d *Dispatcher) Done() bool {
	return d.itemsActive == 0
}

func (d *Dispatcher) stopWorkers() {
	for i := range d.workers {
		d.workers[i].Stop()
	}
}
