package wats

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe("Adding and removing routes", func() {
	BeforeEach(func() {
		pushAndStartNora(appName)
		Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))
	})

	It("should be able to add and remove routes", func() {
		secondHost := generator.PrefixedRandomName(config.GetNamePrefix(), "ROUTE")

		By("changing the environment")
		Eventually(cf.Cf("set-env", appName, "WHY", "force-app-update")).Should(Exit(0))

		By("adding a route")
		Eventually(cf.Cf("map-route", appName, config.GetAppsDomain(), "-n", secondHost)).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))
		Eventually(helpers.CurlingAppRoot(config, secondHost)).Should(ContainSubstring("hello i am nora"))

		By("removing a route")
		Eventually(cf.Cf("unmap-route", appName, config.GetAppsDomain(), "-n", secondHost)).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(config, secondHost)).Should(ContainSubstring("404"))
		Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))

		By("deleting the original route")
		Eventually(cf.Cf("delete-route", config.GetAppsDomain(), "-n", appName, "-f")).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("404"))
	})
})
