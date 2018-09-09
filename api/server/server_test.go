package Server

import (
	"testing"

	"github.com/joemcmahon/logcap"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const example = "https://www.example.com"
const missing = "http://missing.org"

var _ = Describe("matches logs", func() {
	s := New()
	Context("running", func() {
		It("tries change running to stopped", func() {
			s.state[example] = CrawlControl{State: running}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Stop(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "running", "stopped", "crawl paused")))
		})
		It("tries change running to running", func() {
			s.state[example] = CrawlControl{State: running}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Start(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "running", "running", "no action")))
		})
		It("tries change running to failed", func() {
			s.state[example] = CrawlControl{State: running}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Failed(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "running", "failed", "marked failed")))
		})
		It("tries change running to done", func() {
			s.state[example] = CrawlControl{State: running}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Done(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "running", "done", "recording completed crawl")))
		})
	})
	Context("stopped", func() {
		It("tries change stopped to stopped", func() {
			s.state[example] = CrawlControl{State: stopped}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Stop(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "stopped", "stopped", "no action")))
		})
		It("tries change stopped to running", func() {
			s.state[example] = CrawlControl{State: stopped}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Start(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "stopped", "running", "resuming crawl")))
		})
		It("tries change stopped to failed", func() {
			s.state[example] = CrawlControl{State: stopped}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Failed(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "stopped", "failed", "no action")))
		})
		It("tries change stopped to done", func() {
			s.state[example] = CrawlControl{State: stopped}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Done(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "stopped", "done", "no action")))
		})
	})
	Context("unknown", func() {
		//  Note that we use missing for these tests to
		// trigger the "did not find the URL in the crawler state" cases.
		It("tries change unknown to stopped", func() {
			delete((*s).state, missing)
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Stop(missing)
			Ω(logHook).Should(logcap.HaveLogs(changeState(missing, "unknown", "stopped", "no action")))
		})
		It("tries change unknown to running", func() {
			delete((*s).state, missing)
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Start(missing)
			Ω(logHook).Should(logcap.HaveLogs(changeState(missing, "unknown", "running", "starting crawl")))
		})
		It("tries change unknown to failed", func() {
			delete((*s).state, missing)
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Failed(missing)
			Ω(logHook).Should(logcap.HaveLogs(changeState(missing, "unknown", "failed", "no action")))
		})
		It("tries change unknown to done", func() {
			delete((*s).state, missing)
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Done(missing)
			Ω(logHook).Should(logcap.HaveLogs(changeState(missing, "unknown", "done", "no action")))
		})
	})
	Context("failed", func() {
		It("tries change failed to stopped", func() {
			s.state[example] = CrawlControl{State: failed}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Stop(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "failed", "stopped", "no action")))
		})
		It("tries change failed to running", func() {
			s.state[example] = CrawlControl{State: failed}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Start(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "failed", "running", "retrying crawl")))
		})
		It("tries change failed to failed", func() {
			s.state[example] = CrawlControl{State: failed}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Failed(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "failed", "failed", "no action")))
		})
		It("tries change failed to done", func() {
			s.state[example] = CrawlControl{State: failed}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Done(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "failed", "done", "no action")))
		})
	})
	Context("done", func() {
		It("tries change done to stopped", func() {
			s.state[example] = CrawlControl{State: done}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Stop(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "done", "stopped", "no action")))
		})
		It("tries change done to running", func() {
			s.state[example] = CrawlControl{State: done}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Start(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "done", "running", "last crawl discarded, restarting crawl")))
		})
		It("tries change done to failed", func() {
			s.state[example] = CrawlControl{State: done}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Failed(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "done", "failed", "no action")))
		})
		It("tries change done to done", func() {
			s.state[example] = CrawlControl{State: done}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Done(example)
			Ω(logHook).Should(logcap.HaveLogs(changeState(example, "done", "done", "no action")))
		})
	})
})

func TestThings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API server Suite")
}
