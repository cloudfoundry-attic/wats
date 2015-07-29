package wats

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("When staging fails", func() {
	var appName string

	Context("due to insufficient resources", func() {
		BeforeEach(func() {
			context.SetRunawayQuota()

			appName = generator.RandomName()
			Eventually(cf.Cf("push", appName, "--no-start",
				"-m", helpers.RUNAWAY_QUOTA_MEM_LIMIT,
				"-p", "../assets/nora/NoraPublished",
				"-s", "windows2012R2",
				"-b", "https://github.com/ryandotsmith/null-buildpack.git",
			), CF_PUSH_TIMEOUT).Should(Exit(0))
			enableDiego(appName)
		})

		AfterEach(func() {
			Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
			Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
		})

		It("informs the user in the CLI output and the logs", func() {
			logs := cf.Cf("logs", appName)

			start := cf.Cf("start", appName)
			Eventually(start, CF_PUSH_TIMEOUT).Should(Exit(1))
			Expect(start.Out).To(gbytes.Say("InsufficientResources"))

			Eventually(logs.Out).Should(gbytes.Say("Failed to stage application: insufficient resources"))

			app := cf.Cf("app", appName)
			Eventually(app).Should(Exit(0))
			Expect(app.Out).To(gbytes.Say("requested state: stopped"))
		})
	})
})
