package cache

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

// Op is an individual cache operation.
// - Add true adds the key and value to the cache.
// - Clear true empties the cache.
// - Test true sees if key is in the cache and sends
//   the result back over the InCache channel.
type Op struct {
	Add     bool
	Clear   bool
	Test    bool
	Key     string
	Value   error
	InCache chan bool
}

// Cache represents the cache process itself. Send queue actions
// to Cache.Work; send true to Cache.QuitChan to stop the process.
type Cache struct {
	urls map[string]error
	Work chan Op
	Quit chan bool
}

// New creates a new cache.
func New() Cache {
	// Create and return the cache.
	worker := Cache{
		urls: make(map[string]error),
		Work: make(chan Op),
		Quit: make(chan bool)}

	return worker
}

// Run runs the cache. We wait for work on our work queue, execute
// it, and wait for more work
func (c *Cache) Run() {
	go func() {
		for {
			select {
			case op := <-c.Work:
				switch {
				case op.Clear:
					c.urls = make(map[string]error)
				case op.Add:
					c.urls[op.Key] = op.Value
				case op.Test:
					x, ok := c.urls[op.Key]
					spew.Dump(x)
					fmt.Println("send", ok)
					op.InCache <- ok
				}
			case <-c.Quit:
				return
			}
		}
	}()
}
