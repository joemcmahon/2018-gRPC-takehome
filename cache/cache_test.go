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

var _ = Describe("cache", func() {
	Describe("add", func() {
		Context("add two different items", func() {
			BeforeEach(func() {
				c = New()
				c.Run()
				c.Clear()
				c.Add("alpha", fmt.Errorf("one"))
				c.Add("beta", fmt.Errorf("two"))
				c.Quit()
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
				c.Clear()
				c.Add("alpha", fmt.Errorf("one"))
				c.Add("alpha", fmt.Errorf("two"))
				c.Quit()
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
			c.Clear()
			c.Add("gamma", fmt.Errorf("three"))
		})
		It("sends back the right answers", func() {
			Expect(c.Check("delta")).To(BeFalse())
			Expect(c.Check("gamma")).To(BeTrue())
			c.Quit()
		})
	})
})

func TestThings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Suite")
}
