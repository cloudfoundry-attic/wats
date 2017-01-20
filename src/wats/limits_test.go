package wats

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Lifecycle", func() {
	Describe("An app staged on Diego and running on Diego", func() {
		It("exercises the app through its lifecycle", func() {
			By("pushing it")
			Eventually(pushNoraWithOptions(appName, 1, "256m"), CF_PUSH_TIMEOUT).Should(Succeed())

			By("staging and running it on Diego")
			enableDiego(appName)
			Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())

			By("attempting to leak more memory than allowed")
			// leak 300mb
			response := helpers.CurlApp(config, appName, "/leakmemory/300")
			Expect(response).To(ContainSubstring("Insufficient memory"))
		})
	})
})
