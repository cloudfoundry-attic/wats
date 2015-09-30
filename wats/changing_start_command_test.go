package wats

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
)

var _ = Describe("Changing an app's start command", func() {
	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	Describe("staged on Diego and running on Diego", func() {
		It("exercises the app through its lifecycle", func() {
			By("pushing it", func() {
				Eventually(cf.Cf(
					"push", appName,
					"-p", "../assets/batch-script",
					"--no-start",
					"--no-route",
					"-b", "https://github.com/ryandotsmith/null-buildpack.git",
					"-s", "windows2012R2",
					"-c", "loop.bat Hi there!!!"), CF_PUSH_TIMEOUT).Should(Exit(0))
			})

			By("staging and running it on Diego", func() {
				enableDiego(appName)
				disableHealthCheck(appName)
				session := cf.Cf("start", appName)
				Eventually(session, CF_PUSH_TIMEOUT).Should(Exit(0))
			})

			By("verifying it's up", func() {
				output := func() *Buffer {
					session := cf.Cf("logs", appName, "--recent")
					Eventually(session).Should(Exit(0))
					return session.Out
				}
				// OUT... to make sure we don't match the Launcher line: Running `loop.bat Hi there!!!'
				Eventually(output).Should(Say("OUT Hi there!!!"))
			})
		})
	})
})
