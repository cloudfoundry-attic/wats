package wats

import (
	. "github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/onsi/gomega/gbytes"
)

const (
	EXCEED_CELL_MEMORY = "900g"
)

var _ = Describe("When staging fails", func() {
	Context("due to insufficient resources", func() {
		BeforeEach(func() {
			setTotalMemoryLimit(RUNAWAY_QUOTA_MEM_LIMIT)

			Eventually(cf.Cf("push", appName, "--no-start",
				"-m", EXCEED_CELL_MEMORY,
				"-p", "../../assets/nora/NoraPublished",
				"-s", "windows2012R2",
				"-b", HWC_BUILDPACK_URL,
			), CF_PUSH_TIMEOUT).Should(Exit(0))
			enableDiego(appName)
		})

		AfterEach(func() {
			setTotalMemoryLimit("10G")
		})

		It("informs the user in the CLI output and the logs", func() {

			start := cf.Cf("start", appName)
			Eventually(start, CF_PUSH_TIMEOUT).Should(Exit(1))

			app := cf.Cf("app", appName)
			Eventually(app).Should(Exit(0))
			Expect(app.Out).To(gbytes.Say("requested state: stopped"))
		})
	})
})
