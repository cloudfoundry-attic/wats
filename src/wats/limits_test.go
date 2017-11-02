package wats

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Application Lifecycle", func() {
	Describe("An app staged on Diego and running on Diego", func() {
		It("exercises the app through its lifecycle", func() {
			By("pushing it")
			Expect(pushNoraWithOptions(appName, 1, "256m").Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

			By("staging and running it on Diego")
			Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

			By("attempting to leak more memory than allowed")
			// leak 300mb
			response := helpers.CurlApp(config, appName, "/leakmemory/300")
			Expect(response).To(ContainSubstring("Insufficient memory"))
		})
	})
})
