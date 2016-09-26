package wats

import (
	"fmt"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe("Logs from apps hosted by Diego", func() {
	BeforeEach(func() {
		Eventually(pushNora(appName), CF_PUSH_TIMEOUT).Should(Succeed())
		enableDiego(appName)
		Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())
	})

	Context("when the app is running", func() {
		BeforeEach(func() {
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))
		})

		It("captures stdout logs with the correct tag", func() {
			var message string
			var logs *Session

			By("logging application stdout")
			message = "message-from-stdout"
			helpers.CurlApp(appName, fmt.Sprintf("/print/%s", url.QueryEscape(message)))
			//TODO: make nora output message
			//			Eventually(helpers.CurlApp(appName, fmt.Sprintf("/print/%s", url.QueryEscape(message)))).Should(ContainSubstring(message))

			logs = cf.Cf("logs", appName, "--recent")
			Eventually(logs).Should(Exit(0))
			Expect(logs.Out).To(Say(fmt.Sprintf("\\[APP/0\\]\\s+OUT %s", message)))

		})

		It("captures stderr logs with the correct tag", func() {
			var message string
			var logs *Session

			By("logging application stderr")
			message = "message-from-stderr"
			helpers.CurlApp(appName, fmt.Sprintf("/print_err/%s", url.QueryEscape(message)))

			logs = cf.Cf("logs", appName, "--recent")
			Eventually(logs).Should(Exit(0))
			Expect(logs.Out).To(Say(fmt.Sprintf("\\[APP/0\\]\\s+ERR %s", message)))
		})
	})
})
