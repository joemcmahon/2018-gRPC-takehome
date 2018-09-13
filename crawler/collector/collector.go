package collector

import (
	dispatcher "github.com/joemcmahon/joe_macmahon_technical_test/crawler/dispatcher"
	log "github.com/sirupsen/logrus"
)

// Collector global data
type Collector struct {
	Dispatcher dispatcher.Dispatcher
	NWorkers   int
	NRequests  int
	URL        string
}

// New creates a new collector for the gRPC server.
func New(workerCount, requestCount int) Collector {
	return Collector{
		NWorkers:  workerCount,
		NRequests: requestCount,
	}
}

// Crawl schedules the crawl on the server.
func (c *Collector) Crawl(url string) {
	log.Debug("Starting the dispatcher")
	c.Dispatcher = dispatcher.Start(c.NWorkers, c.NRequests)
	c.Dispatcher.StartWork(url)
}

// Done checks with the dispatcher to see if it's done crawling.
func (c *Collector) Done() bool {
	return c.Dispatcher.Done()
}
