package wats

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ASP classic applications", func() {
	It("exercises the app through its lifecycle", func() {
		By("pushing it")
		Eventually(pushApp(appName, "../assets/asp-classic", 1, "1g")).Should(Succeed())

		By("staging and running it on Diego")
		enableDiego(appName)
		Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())

		By("verifying it's up")
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hello World!"))
	})
})
