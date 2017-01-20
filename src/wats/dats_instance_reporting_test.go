package wats

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Getting instance information", func() {
	BeforeEach(func() {
		Eventually(cf.Cf("push", appName, "-m", "2Gb", "-p", "../../assets/webapp", "--no-start", "-b", "binary_buildpack", "-s", "windows2012R2"), CF_PUSH_TIMEOUT).Should(Exit(0))
		enableDiego(appName)
		session := cf.Cf("start", appName)
		Eventually(session, CF_PUSH_TIMEOUT).Should(Exit(0))
	})

	Context("scaling memory", func() {
		BeforeEach(func() {
			setTotalMemoryLimit(RUNAWAY_QUOTA_MEM_LIMIT)

			scale := cf.Cf("scale", appName, "-m", EXCEED_CELL_MEMORY, "-f")
			Eventually(scale, CF_PUSH_TIMEOUT).Should(Say("insufficient resources|down"))
			scale.Kill()
		})

		AfterEach(func() {
			setTotalMemoryLimit("10G")
		})

		It("fails with insufficient resources", func() {
			app := cf.Cf("app", appName)
			Eventually(app).Should(Exit(0))
			Expect(app.Out).To(Say("insufficient resources"))
		})
	})
})
