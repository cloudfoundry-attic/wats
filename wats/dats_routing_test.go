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
	var appName string

	BeforeEach(func() {
		appName = generator.RandomName()
		pushAndStartNora(appName)
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	It("should be able to add and remove routes", func() {
		secondHost := generator.RandomName()

		By("changing the environment")
		Eventually(cf.Cf("set-env", appName, "WHY", "force-app-update")).Should(Exit(0))

		By("adding a route")
		Eventually(cf.Cf("map-route", appName, helpers.LoadConfig().AppsDomain, "-n", secondHost)).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))
		Eventually(helpers.CurlingAppRoot(secondHost)).Should(ContainSubstring("hello i am nora"))

		By("removing a route")
		Eventually(cf.Cf("unmap-route", appName, helpers.LoadConfig().AppsDomain, "-n", secondHost)).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(secondHost)).Should(ContainSubstring("404"))
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))

		By("deleting the original route")
		Eventually(cf.Cf("delete-route", helpers.LoadConfig().AppsDomain, "-n", appName, "-f")).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("404"))
	})
})
