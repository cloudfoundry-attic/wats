package wats

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
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
			Expect(pushNoraWithOptions(appName, 1, "2g").Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			By("staging and running it on Diego")
			Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

			response := helpers.CurlApp(config, appName, "/curl/127.0.0.1/8500")
			Expect(response).To(ContainSubstring("The server committed a protocol violation. Section=ResponseStatusLine"))
		})
	})
})
