package wats

import (
	"fmt"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe("Logs from apps hosted by Diego", func() {
	var appName string

	BeforeEach(func() {
		appName = generator.RandomName()

		Eventually(pushNora(appName), CF_PUSH_TIMEOUT).Should(Succeed())
		enableDiego(appName)
		Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	Context("when the app is running", func() {
		BeforeEach(func() {
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))
		})

		It("captures stdout logs with the correct tag", func() {
			var message string
			var logs *Session

			By("logging health checks")
			logs = cf.Cf("logs", appName, "--recent")
			Eventually(logs).Should(Exit(0))
			Expect(logs.Out).To(Say("\\[HEALTH/0\\]\\s+OUT healthcheck passed"))
			Expect(logs.Out).To(Say("\\[HEALTH/0\\]\\s+OUT Exit status 0"))

			By("logging application stdout")
			message = "message-from-stdout"
			helpers.CurlApp(appName, fmt.Sprintf("/print/%s", url.QueryEscape(message)))
			//TODO: make nora output message
			//			Eventually(helpers.CurlApp(appName, fmt.Sprintf("/print/%s", url.QueryEscape(message)))).Should(ContainSubstring(message))

			logs = cf.Cf("logs", appName, "--recent")
			Eventually(logs).Should(Exit(0))
			Expect(logs.Out).To(Say(fmt.Sprintf("\\[APP/0\\]\\s+OUT %s", message)))

		})

		XIt("captures stderr logs with the correct tag", func() {
			var message string
			var logs *Session

			By("logging health checks")
			logs = cf.Cf("logs", appName, "--recent")
			Eventually(logs).Should(Exit(0))
			Expect(logs.Out).To(Say("\\[HEALTH/0\\]\\s+OUT healthcheck passed"))
			Expect(logs.Out).To(Say("\\[HEALTH/0\\]\\s+OUT Exit status 0"))

			By("logging application stderr")
			message = "messag-from-stderr"
			By("logging application stderr")
			message = "A message from stderr"
			Eventually(helpers.CurlApp(appName, fmt.Sprintf("/print_err/%s", url.QueryEscape(message)))).Should(ContainSubstring(message))

			logs = cf.Cf("logs", appName, "--recent")
			Eventually(logs).Should(Exit(0))
			Expect(logs.Out).To(Say(fmt.Sprintf("\\[APP/0\\]\\s+ERR %s", message)))
		})
	})
})
