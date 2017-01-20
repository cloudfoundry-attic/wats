package wats

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application environment", func() {
	Describe("And app staged on Diego and running on Diego", func() {
		It("should not have too many environment variable exposed", func() {
			By("pushing it")
			Eventually(pushNoraWithOptions(appName, 1, "2g"), CF_PUSH_TIMEOUT).Should(Succeed())
			By("staging and running it on Diego")
			enableDiego(appName)
			Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())

			excludedList := []string{
				"COMPUTERNAME",
				"ALLUSERSPROFILE",
				"FP_NO_HOST_CHECK",
				"GOPATH",
				"NUMBER_OF_PROCESSORS",
				"OS",
				"PATHEXT",
				"PROCESSOR_ARCHITECTURE",
				"PROCESSOR_IDENTIFIER",
				"PROCESSOR_LEVEL",
				"PROCESSOR_REVISION",
				"PSModulePath",
				"PUBLIC",
				"SystemDrive",
				"USERDOMAIN",
				"VS110COMNTOOLS",
				"VS120COMNTOOLS",
				"WIX",
			}
			response := helpers.CurlApp(config, appName, "/env")
			var env map[string]string
			json.Unmarshal([]byte(response), &env)
			for _, excludedKey := range excludedList {
				Expect(env).NotTo(HaveKey(excludedKey))
			}
		})
	})
})
