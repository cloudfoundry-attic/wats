package wats

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe("An application printing a bunch of output", func() {

	BeforeEach(func() {
		Expect(pushNoraWithOptions(appName, 1, "1G").Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
		Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
	})

	It("doesn't die when printing 32MB", func() {
		beforeId := helpers.CurlApp(config, appName, "/id")

		Expect(helpers.CurlAppWithTimeout(config, appName, "/logspew/32000", DEFAULT_LONG_TIMEOUT)).
			To(ContainSubstring("Just wrote 32000 kbytes to the log"))

		Consistently(func() string {
			return helpers.CurlApp(config, appName, "/id")
		}, "10s").Should(Equal(beforeId))
	})
})
