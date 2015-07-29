package wats

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe("An application printing a bunch of output", func() {
	var appName string

	BeforeEach(func() {
		appName = generator.RandomName()

		Eventually(pushNora(appName), CF_PUSH_TIMEOUT).Should(Succeed())
		enableDiego(appName)
		Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
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
