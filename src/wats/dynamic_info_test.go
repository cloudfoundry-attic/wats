package wats

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("A running application", func() {
	BeforeEach(func() {
		pushAndStartNora(appName)
	})

	It("can show crash events", func() {
		helpers.CurlApp(config, appName, "/exit")

		Eventually(func() string {
			return string(cf.Cf("events", appName).Wait(CF_PUSH_TIMEOUT).Out.Contents())
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Exited"))
	})
})
