package cache

// Op is an individual cache operation.
// - Add true adds the key and value to the cache.
// - Clear true empties the cache.
// - Test true sees if key is in the cache and sends
//   the result back over the InCache channel.
type op struct {
	add     bool
	clear   bool
	test    bool
	key     string
	value   error
	inCache chan bool
}

// Cache represents the cache process itself. Send queue actions
// to Cache.Work; send true to Cache.QuitChan to stop the process.
type Cache struct {
	urls  map[string]error
	cWork chan op
	cQuit chan bool
}

// New creates a new cache.
func New() Cache {
	// Create and return the cache.
	worker := Cache{
		urls:  make(map[string]error),
		cWork: make(chan op),
		cQuit: make(chan bool)}

	return worker
}

// Clear clears the cache. Sync is hidden in this function.
func (c *Cache) Clear() {
	clear := op{clear: true}
	c.cWork <- clear
}

// Add creates the necessary data structure to add an item and adds it.
// The synchronization is hidden inside this function.
func (c *Cache) Add(item string, value error) {
	add := op{
		add:   true,
		key:   item,
		value: value,
	}
	c.cWork <- add
}

// Check sees if an item is in the cache or not. Sync is done inside
// this function.
func (c *Cache) Check(item string) bool {
	z := make(chan bool)
	query := op{
		test:    true,
		key:     item,
		inCache: z,
	}
	c.cWork <- query
	there := <-z
	return there
}

// Run runs the cache. We wait for work on our work queue, execute
// it, and wait for more work
func (c *Cache) Run() {
	go func() {
		for {
			select {
			case op := <-c.cWork:
				switch {
				case op.clear:
					c.urls = make(map[string]error)
				case op.add:
					c.urls[op.key] = op.value
				case op.test:
					_, ok := c.urls[op.key]
					op.inCache <- ok
				}
			case <-c.cQuit:
				return
			}
		}
	}()
}

// Quit stops the process.
func (c *Cache) Quit() {
	c.cQuit <- true
}
