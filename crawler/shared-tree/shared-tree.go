package sharedTree

import (
	"github.com/disiqueira/gotree"
	log "github.com/sirupsen/logrus"
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

// Tree represents the tree management process itself. Send
// items to Tree.Add; send formatting requests to Format; send true
// to Tree.Quit to stop the process.
type Tree struct {
	tree   *gotree.Tree
	add    chan addition
	format chan formatReq
	quit   chan bool
}

// New creates a new tree.
func New() *Tree {
	// Create and return the tree.
	t := Tree{
		tree:   nil,
		add:    make(chan addition, 1),
		format: make(chan formatReq, 1),
		quit:   make(chan bool, 1),
	}

	return &t
}

// Run runs the tree. We wait for work on our work queue, execute
// it, and wait for more work
func (t *Tree) Run() {
	go func() {
		for {
			select {
			case add := <-t.add:
				log.Debugf("adding item to tree")
				var newT gotree.Tree
				// Build a new tree if it is empty.
				if t.tree == nil {
					log.Debugf("building new tree")
					newT = gotree.New(add.item)
					t.tree = &newT
					// Insert at the insertion point if it is not.
				} else {
					log.Debugf("adding to existing tree")
					newT = (*add.insertPoint).Add(add.item)
				}
				log.Debug((*t.tree).Print())
				// Send back the new insert point.
				log.Debugf("sending back the insert point")
				add.response <- &newT

			case req := <-t.format:
				log.Debugf("formatting tree")
				req.response <- (*t.tree).Print()

			case <-t.quit:
				log.Debugf("tree exits")
				return
			}
		}
	}()
}

// AddAt inserts a new item at the specified insert point and
// returns the new item as an insert point. This allows us to
// build trees downward from the root.
func (t *Tree) AddAt(point *gotree.Tree, s string) *gotree.Tree {
	a := addition{
		item:        s,
		insertPoint: point,
		response:    make(chan *gotree.Tree, 1),
	}
	log.Debugf("adding %s", a.item)
	t.add <- a
	log.Debugf("wait for response")
	p := <-a.response
	log.Debugf("completed add of %s", a.item)
	return p
}

// Format returns a formatted version of the tree, using gotree's Print().
func (t *Tree) Format() string {
	return (*t.tree).Print()
}

// Quit stops the process.
func (t *Tree) Quit() {
	t.quit <- true
}
