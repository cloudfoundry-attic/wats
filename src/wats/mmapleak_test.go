package wats

import (
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe("Application Lifecycle", func() {
	Describe("An app staged on Diego and running on Diego", func() {
		XIt("attempts to leak mmap", func() {
			By("pushing it", func() {
				Expect(pushNora(appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
				Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			})

			By("verifying it's up", func() {
				Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))
			})

			By("requesting current memory commit charge")
			commitCharge := helpers.CurlApp(config, appName, "/commitcharge")
			commitChargeValue, err := strconv.ParseInt(commitCharge, 10, 64)
			Expect(err).NotTo(HaveOccurred())

			By("Running mmapleak (3Gb)", func() {
				helpers.CurlApp(config, appName, fmt.Sprintf("/mmapleak/{%#v}", int64(3)*1024*1024*1024))
			})

			By("Commit Charge should not have changed by more than container max (2Gb)", func() {
				newCommitCharge := helpers.CurlApp(config, appName, "/commitcharge")
				newCommitChargeValue, err := strconv.ParseInt(newCommitCharge, 10, 64)
				Expect(err).NotTo(HaveOccurred())

				Expect(newCommitChargeValue - commitChargeValue).To(BeNumerically("<", (int64(2) * 1024 * 1024 * 1024)))
			})
		})
	})
})
