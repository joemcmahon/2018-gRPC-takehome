package Server

import (
	"testing"

	"github.com/joemcmahon/joe_macmahon_technical_test/crawler/test/mock_fetcher"
	"github.com/joemcmahon/logcap"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const example = "https://www.example.com"
const missing = "http://missing.org"

var _ = Describe("matches logs", func() {
	f := MockFetcher.New()
	s := New(f)
	Context("running", func() {
		It("tries change running to stopped", func() {
			s.state[example] = CrawlControl{State: running}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Pause(example)
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
	})
	Context("stopped", func() {
		It("tries change stopped to stopped", func() {
			s.state[example] = CrawlControl{State: stopped}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Pause(example)
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
	})
	Context("unknown", func() {
		//  Note that we use missing for these tests to
		// trigger the "did not find the URL in the crawler state" cases.
		It("tries change unknown to stopped", func() {
			delete((*s).state, missing)
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Pause(missing)
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
	})
	Context("failed", func() {
		It("tries change failed to stopped", func() {
			s.state[example] = CrawlControl{State: failed}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Pause(example)
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
	})
	Context("done", func() {
		It("tries change done to stopped", func() {
			s.state[example] = CrawlControl{State: done}
			logHook := logcap.NewLogHook()
			logHook.Start()
			defer logHook.Stop()
			s.Pause(example)
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
	})
})

func TestThings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API server Suite")
}
