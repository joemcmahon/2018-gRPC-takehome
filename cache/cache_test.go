package cache

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

const example = "https://www.example.com"
const missing = "http://missing.org"

var c Cache

func init() {
	if os.Getenv("TESTING") != "" {
		log.SetLevel(log.DebugLevel)
	}
}

var _ = Describe("", func() {
	Describe("add", func() {
		Context("add two different items", func() {
			BeforeEach(func() {
				c = New()
				c.Run()
				clear := Op{
					Clear: true,
				}
				c.Work <- clear
				add := Op{
					Add:   true,
					Key:   "alpha",
					Value: fmt.Errorf("one"),
				}
				c.Work <- add
				add = Op{
					Add:   true,
					Key:   "beta",
					Value: fmt.Errorf("two"),
				}
				c.Work <- add
				c.Quit <- true
			})
			It("added the items", func() {
				// Can whitebox because we're in the same package
				keys := reflect.ValueOf(c.urls).MapKeys()
				Expect(len(keys)).To(Equal(2))
				Expect(c.urls["alpha"]).To(Equal(fmt.Errorf("one")))
				Expect(c.urls["beta"]).To(Equal(fmt.Errorf("two")))
			})
		})
		Context("add two identical items", func() {
			BeforeEach(func() {
				c = New()
				c.Run()
				clear := Op{
					Clear: true,
				}
				c.Work <- clear
				add := Op{
					Add:   true,
					Key:   "alpha",
					Value: fmt.Errorf("one"),
				}
				c.Work <- add
				add = Op{
					Add:   true,
					Key:   "alpha",
					Value: fmt.Errorf("two"),
				}
				c.Work <- add
				c.Quit <- true
			})
			It("added only one item", func() {
				keys := reflect.ValueOf(c.urls).MapKeys()
				Expect(len(keys)).To(Equal(1))
				Expect(c.urls["alpha"]).To(Equal(fmt.Errorf("two")))
			})
		})
	})
	Describe("check", func() {
		BeforeEach(func() {
			c = New()
			c.Run()
			clear := Op{
				Clear: true,
			}
			c.Work <- clear
			add := Op{
				Add:   true,
				Key:   "gamma",
				Value: fmt.Errorf("three"),
			}
			c.Work <- add
		})
		It("sends back false for items not there", func() {
			z := make(chan bool)
			query := Op{
				Test:    true,
				Key:     "delta",
				InCache: z,
			}
			c.Work <- query
			there := <-z
			Expect(there).To(BeFalse())
		})
		It("sends back true for items that are there", func() {
			z := make(chan bool)
			query := Op{
				Test:    true,
				Key:     "gamma",
				InCache: z,
			}
			c.Work <- query
			there := <-z
			Expect(there).To(BeTrue())
		})
	})
})

var _ = func() {
	c.Quit <- true
}

func TestThings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Suite")
}
