package wats

import (
	"time"

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
				Eventually(pushNora(appName), CF_PUSH_TIMEOUT).Should(Succeed())
			})

			By("staging and running it on Diego", func() {
				enableDiego(appName)
				Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())
			})

			// FIXME: Something about stdout logging

			By("verifying it's up", func() {
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))
			})

			By("stopping it", func() {
				Eventually(runCf("stop", appName)).Should(Succeed())
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("404"))
			})

			By("setting an environment variable", func() {
				Eventually(runCf("set-env", appName, "FOO", "bar")).Should(Succeed())
			})

			By("starting it", func() {
				Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))
			})

			By("checking custom env variables are available", func() {
				Eventually(func() string {
					return helpers.CurlAppWithTimeout(appName, "/env/FOO", 30*time.Second)
				}).Should(ContainSubstring("bar"))
			})

			By("scaling it", func() {
				Eventually(runCf("scale", appName, "-i", "2")).Should(Succeed())
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
