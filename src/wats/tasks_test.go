package wats

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
)

var _ = Describe("Task Lifecycle", func() {
	It("exercises the task lifecycle on windows", func() {
		if !config.TestTask {
			Skip("Skipping tasks tests (requires diego-release v1.20.0 and above)")
		}
		By("pushing it", func() {
			Eventually(cf.Cf("push", appName, "-p", "../../assets/webapp", "-c", ".\\webapp.exe",
				"--no-start", "-b", binaryBuildPackURL, "-s", "windows2016"), CF_PUSH_TIMEOUT).Should(Exit(0))
		})

		By("staging and running it on Diego", func() {
			enableDiego(appName)
			session := cf.Cf("start", appName)
			Eventually(session, CF_PUSH_TIMEOUT).Should(Exit(0))
		})

		By("running a task", func() {
			session := cf.Cf("run-task", appName, "cmd /c echo 'hello world'")
			Eventually(session).Should(Exit(0))
		})

		By("checking the task has succeeded", func() {
			Eventually(func() *Session {
				taskSession := cf.Cf("tasks", appName)
				Expect(taskSession.Wait(DEFAULT_TIMEOUT)).To(Exit(0))
				return taskSession
			}).Should(Say("SUCCEEDED"))
		})
	})
})
