package diego

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe("Application Lifecycle", func() {
	var appName string

	apps := func() *Session {
		return cf.Cf("apps").Wait()
	}

	BeforeEach(func() {
		appName = generator.RandomName()
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	Describe("An app staged on Diego and running on Diego", func() {
		It("exercises the app through its lifecycle", func() {
			By("pushing it", func() {
				Eventually(cf.Cf("push", appName, "-p", "../assets/nora/NoraPublished", "--no-start", "-b", "java_buildpack", "-s", "windows2012R2"), CF_PUSH_TIMEOUT).Should(Exit(0))
			})

			By("staging and running it on Diego", func() {
				enableDiego(appName)
				session := cf.Cf("start", appName)
				Eventually(session, CF_PUSH_TIMEOUT).Should(Exit(0))
			})

			// FIXME: Something about stdout logging

			By("verifying it's up", func() {
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))
			})

			By("stopping it", func() {
				Eventually(cf.Cf("stop", appName)).Should(Exit(0))
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("404"))
			})

			By("setting an environment variable", func() {
				Eventually(cf.Cf("set-env", appName, "FOO", "bar")).Should(Exit(0))
			})

			By("starting it", func() {
				Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))
			})

			By("checking custom env variables are available", func() {
				Eventually(helpers.CurlApp(appName, "/env/FOO")).Should(ContainSubstring("bar"))
			})

			By("scaling it", func() {
				Eventually(cf.Cf("scale", appName, "-i", "2")).Should(Exit(0))
				Eventually(apps).Should(Say("2/2"))
			})

			// By("restarting an instance", func() {
			// 	idsBefore := reportedIDs(2)
			// 	Eventually(cf.Cf("restart-app-instance", appName, "1")).Should(Exit(0))
			// 	Eventually(func() []string {
			// 		return differentIDsFrom(idsBefore)
			// 	}).Should(HaveLen(1))
			// })
		})
	})
})
