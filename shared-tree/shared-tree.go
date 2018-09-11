package sharedTree

import (
	"github.com/disiqueira/gotree"
)

// addition is an individual tree item to be added at a specific point.
// InsertPoint is nil if we are building a new tree. The Response is
// indeed a pointer into the shared tree somewhere; the caller must
// treat this as an opaque token and not try to use it directly.
type addition struct {
	item        string
	insertPoint *gotree.Tree
	response    chan *gotree.Tree
}

// FormatReq is sent to request a formatted dump of the tree. The
// printed tree is written back to the Response channel.
type formatReq struct {
	response chan string
}

// SharedTree represents the tree management process itself. Send
// items to Tree.Add; send formatting requests to Format; send true
// to Tree.Quit to stop the process.
type SharedTree struct {
	tree   *gotree.Tree
	add    chan addition
	format chan formatReq
	quit   chan bool
}

// New creates a new tree.
func New() SharedTree {
	// Create and return the cache.
	worker := SharedTree{
		tree:   nil,
		add:    make(chan addition),
		format: make(chan formatReq),
		quit:   make(chan bool)}

	return worker
}

// Run runs the tree. We wait for work on our work queue, execute
// it, and wait for more work
func (t *SharedTree) Run() {
	go func() {
		for {
			select {
			case add := <-t.add:
				var newT gotree.Tree
				// Build a new tree if it is empty.
				if t.tree == nil {
					newT = gotree.New(add.item)
					t.tree = &newT
					// Insert at the insertion point if it is not.
				} else {
					newT = (*add.insertPoint).Add(add.item)
				}
				// Send back the new insert point.
				add.response <- &newT

			case req := <-t.format:
				req.response <- (*t.tree).Print()

			case <-t.quit:
				return
			}
		}
	}()
}

// AddAt inserts a new item at the specified insert point and
// returns the new item as an insert point. This allows us to
// build trees downward from the root.
func (t *SharedTree) AddAt(point *gotree.Tree, s string) *gotree.Tree {
	a := addition{
		item:        s,
		insertPoint: point,
		response:    make(chan *gotree.Tree),
	}
	t.add <- a
	p := <-a.response
	return p
}

// Format returns a formatted version of the tree, using gotree's Print().
func (t *SharedTree) Format() string {
	return (*t.tree).Print()
}

// Quit stops the process.
func (t *SharedTree) Quit() {
	t.quit <- true
}
