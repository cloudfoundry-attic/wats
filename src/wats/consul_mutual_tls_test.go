package wats

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application security", func() {
	BeforeEach(func() {
		if !config.ConsulMutualTls {
			Skip("Skipping Consul mutual TLS tests")
		}
	})

	Describe("And app staged on Diego and running on Diego", func() {
		It("should block access to consul", func() {
			By("pushing it")
			Eventually(pushNoraWithOptions(appName, 1, "2g"), CF_PUSH_TIMEOUT).Should(Succeed())
			By("staging and running it on Diego")
			enableDiego(appName)
			Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())

			response := helpers.CurlApp(config, appName, "/curl/127.0.0.1/8500")
			Expect(response).To(ContainSubstring("The server committed a protocol violation. Section=ResponseStatusLine"))
		})
	})
})
