package Server

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const example = "https://www.example.com"
const missing = "http://missing.org"

var _ = Describe("matches logs", func() {
	s := New()
	Context("running", func() {
		s.state[example] = running
		It("tries change to stopped", func() {
		})
		It("tries change to running", func() {
		})
		It("tries change to failed", func() {
		})
		It("tries change to done", func() {
		})
	})
	Context("stopped", func() {
		s.state[example] = stopped
		It("tries change to stopped", func() {
		})
		It("tries change to running", func() {
		})
		It("tries change to failed", func() {
		})
		It("tries change to done", func() {
		})
	})
	Context("unknown", func() {
		//  Note that we use missing for these tests to
		// trigger the "did not find the URL in the crawler state" cases.
		It("tries change to stopped", func() {
		})
		It("tries change to running", func() {
		})
		It("tries change to failed", func() {
		})
		It("tries change to done", func() {
		})
	})
	Context("failed", func() {
		s.state[example] = failed
		It("tries change to stopped", func() {
		})
		It("tries change to running", func() {
		})
		It("tries change to failed", func() {
		})
		It("tries change to done", func() {
		})
	})
	Context("done", func() {
		s.state[example] = done
		It("tries change to stopped", func() {
		})
		It("tries change to running", func() {
		})
		It("tries change to failed", func() {
		})
		It("tries change to done", func() {
		})
	})

	//It("matcher", func() {
	//	logHook := logcap.NewLogHook()
	//	logHook.Start()
	//	defer logHook.Stop()
	//	logrus.Info("This is a log entry")
	//	Ω(logHook).Should(logcap.HaveLogs("This is a log entry"))
	//})
})

func TestThings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logcap Suite")
}
