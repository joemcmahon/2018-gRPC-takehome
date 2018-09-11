package sharedTree

import (
	"os"
	"testing"

	"github.com/disiqueira/gotree"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var t SharedTree

func init() {
	if os.Getenv("TESTING") != "" {
		log.SetLevel(log.DebugLevel)
	}
}

const expected = `root
├── a
└── b
`

var _ = Describe("shared tree", func() {
	Context("Insert into empty tree", func() {
		var r *gotree.Tree
		BeforeEach(func() {
			t = New()
			t.Run()
			a := Addition{
				Item:     "root",
				Response: make(chan *gotree.Tree),
			}
			t.Add <- a
			r = <-a.Response
			t.Quit <- true
		})
		It("added the item", func() {
			// Can whitebox because we're not sharing yet
			Expect(r).ToNot(BeNil())
			Expect((*r).Print()).To(Equal("root\n"))
		})
		Context("insert multiple items", func() {
			BeforeEach(func() {
				t = New()
				t.Run()
				a := Addition{
					Item:     "root",
					Response: make(chan *gotree.Tree),
				}
				t.Add <- a
				root := <-a.Response
				a = Addition{
					Item:        "a",
					InsertPoint: root,
					Response:    make(chan *gotree.Tree),
				}
				t.Add <- a
				<-a.Response
				a = Addition{
					Item:        "b",
					InsertPoint: root,
					Response:    make(chan *gotree.Tree),
				}
				t.Add <- a
				<-a.Response
				t.Quit <- true
			})
			It("added all the items properly", func() {
				Expect(t.tree).ToNot(BeNil())
				Expect((*t.tree).Print()).To(Equal(expected))
			})
		})
	})
	Describe("Format output", func() {
		BeforeEach(func() {
		})
		It("sends back a correctly-formatted tree", func() {
		})
	})
})

func TestThings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Suite")
}
