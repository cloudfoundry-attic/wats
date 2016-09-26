package wats

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe("An application printing a bunch of output", func() {

	BeforeEach(func() {
		Eventually(pushNora(appName), CF_PUSH_TIMEOUT).Should(Succeed())
		enableDiego(appName)
		Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())
	})

	It("doesn't die when printing 32MB", func() {
		beforeId := helpers.CurlApp(appName, "/id")

		Expect(helpers.CurlAppWithTimeout(appName, "/logspew/32000", DEFAULT_TIMEOUT)).
			To(ContainSubstring("Just wrote 32000 kbytes to the log"))

		Consistently(func() string {
			return helpers.CurlApp(appName, "/id")
		}, "10s").Should(Equal(beforeId))
	})
})
