package wats

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("A running application", func() {
	var appName string

	BeforeEach(func() {
		appName = generator.RandomName()
		pushAndStartNora(appName)
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(gexec.Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(gexec.Exit(0))
	})

	It("can show crash events", func() {
		helpers.CurlApp(appName, "/exit")

		Eventually(func() string {
			return string(cf.Cf("events", appName).Wait(CF_PUSH_TIMEOUT).Out.Contents())
		}, CF_PUSH_TIMEOUT).Should(ContainSubstring("Exited"))
	})
})
