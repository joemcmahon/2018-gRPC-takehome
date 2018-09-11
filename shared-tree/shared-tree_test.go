package sharedTree

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

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
		var t SharedTree
		BeforeEach(func() {
			t = New()
			t.Run()
			t.AddAt(nil, "root")
			t.Quit()
		})
		It("added the item", func() {
			// Can whitebox because we're not sharing yet
			Expect(t.tree).ToNot(BeNil())
			Expect((*t.tree).Print()).To(Equal("root\n"))
		})
		Context("insert multiple items", func() {
			var t SharedTree
			var answer string
			BeforeEach(func() {
				t = New()
				t.Run()
				root := t.AddAt(nil, "root")
				_ = t.AddAt(root, "a")
				_ = t.AddAt(root, "b")
				answer = t.Format()
				t.Quit()
			})
			It("added all the items properly", func() {
				Expect(t.tree).ToNot(BeNil())
				Expect((*t.tree).Print()).To(Equal(expected))
			})
			It("sends back a correctly-formatted tree", func() {
				Expect(answer).To(Equal(expected))
			})
		})
	})
})

func TestThings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Suite")
}
