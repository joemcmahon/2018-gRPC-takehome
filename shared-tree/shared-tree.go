package sharedTree

import (
	"github.com/disiqueira/gotree"
)

// Addition is an individual tree item to be added at a specific point.
// InsertPoint is nil if we are building a new tree. The Response is
// indeed a pointer into the shared tree somewhere; the caller must
// treat this as an opaque token and not try to use it directly.
type Addition struct {
	Item        string
	InsertPoint *gotree.Tree
	Response    chan *gotree.Tree
}

// FormatReq is sent to request a formatted dump of the tree. The
// printed tree is written back to the Response channel.
type FormatReq struct {
	Response chan string
}

// SharedTree represents the tree management process itself. Send
// items to Tree.Add; send formatting requests to Format; send true
// to Tree.Quit to stop the process.
type SharedTree struct {
	tree   *gotree.Tree
	Add    chan Addition
	Format chan FormatReq
	Quit   chan bool
}

// New creates a new tree.
func New() SharedTree {
	// Create and return the cache.
	worker := SharedTree{
		tree:   nil,
		Add:    make(chan Addition),
		Format: make(chan FormatReq),
		Quit:   make(chan bool)}

	return worker
}

// Run runs the tree. We wait for work on our work queue, execute
// it, and wait for more work
func (t *SharedTree) Run() {
	go func() {
		for {
			select {
			case add := <-t.Add:
				var newT gotree.Tree
				// Build a new tree if it is empty.
				if t.tree == nil {
					newT = gotree.New(add.Item)
					t.tree = &newT
					// Insert at the insertion point if it is not.
				} else {
					newT = (*add.InsertPoint).Add(add.Item)
				}
				// Send back the new insert point.
				add.Response <- &newT

			case req := <-t.Format:
				req.Response <- (*t.tree).Print()

			case <-t.Quit:
				return
			}
		}
	}()
}
