package wats

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
)

var _ = Describe("Changing an app's start command", func() {
	var appName string

	BeforeEach(func() {
		appName = generator.RandomName()
	})

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
					"-c", "loop.bat"), CF_PUSH_TIMEOUT).Should(Exit(0))
			})

			By("staging and running it on Diego", func() {
				enableDiego(appName)
				disableSsh(appName)
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
				Eventually(output).Should(Say("Hi there!!!"))
			})
		})
	})
})
