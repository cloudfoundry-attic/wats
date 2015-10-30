package wats

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe("A standalone webapp", func() {

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	Describe("staged on Diego and running on Diego", func() {
		It("exercises the app through its lifecycle", func() {
			By("pushing it", func() {
				Eventually(cf.Cf("push", appName, "-p", "../assets/webapp", "--no-start", "-b", "binary_buildpack", "-s", "windows2012R2"), CF_PUSH_TIMEOUT).Should(Exit(0))
			})

			By("staging and running it on Diego", func() {
				enableDiego(appName)
				session := cf.Cf("start", appName)
				Eventually(session, CF_PUSH_TIMEOUT).Should(Exit(0))
			})

			By("verifying it's up", func() {
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hi i am a standalone webapp"))
			})
		})
	})
})
