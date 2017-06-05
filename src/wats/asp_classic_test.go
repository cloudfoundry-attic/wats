package wats

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("ASP classic applications", func() {
	It("exercises the app through its lifecycle", func() {
		By("pushing it")
		Expect(pushApp(appName, "../../assets/asp-classic", 1, "256m", hwcBuildPackURL).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

		By("staging and running it on Diego")
		enableDiego(appName)
		Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

		By("verifying it's up")
		Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("Hello World!"))
	})
})
