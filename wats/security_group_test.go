package wats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe("Security Groups", func() {
	type NoraCurlResponse struct {
		Stdout     string
		Stderr     string
		ReturnCode int `json:"return_code"`
	}

	AfterEach(func() {
		Eventually(runCf("logs", appName, "--recent")).Should(Succeed())
		Eventually(runCf("delete", appName, "-f")).Should(Succeed())
	})

	// this test assumes the default running security groups block access to the DEAs
	// the test takes advantage of the fact that the DEA ip address and internal container ip address
	//  are discoverable via the cc api and nora's myip endpoint
	It("allows traffic and then blocks traffic", func() {
		By("pushing it")
		Eventually(pushNora(appName), CF_PUSH_TIMEOUT).Should(Succeed())

		By("staging and running it on Diego")
		enableDiego(appName)
		Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())

		By("verifying it's up")
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))

		secureAddress := helpers.LoadConfig().SecureAddress
		secureHost, securePort, err := net.SplitHostPort(secureAddress)
		Expect(err).NotTo(HaveOccurred())

		// test app egress rules
		curlResponse := func() int {
			var noraCurlResponse NoraCurlResponse
			resp := helpers.CurlApp(appName, fmt.Sprintf("/curl/%s/%s", secureHost, securePort))
			json.Unmarshal([]byte(resp), &noraCurlResponse)
			return noraCurlResponse.ReturnCode
		}
		firstCurlError := curlResponse()

		// apply security group
		rules := fmt.Sprintf(`[{"destination":"%s","ports":"%s","protocol":"tcp"}]`, secureHost, securePort)

		file, _ := ioutil.TempFile(os.TempDir(), "DATS-sg-rules")
		defer os.Remove(file.Name())
		file.WriteString(rules)
		file.Close()

		rulesPath := file.Name()
		securityGroupName := fmt.Sprintf("DATS-SG-%s", generator.RandomName())

		cf.AsUser(context.AdminUserContext(), time.Minute, func() {
			Eventually(runCf("create-security-group", securityGroupName, rulesPath)).Should(Succeed())
			Eventually(
				runCf("bind-security-group",
					securityGroupName,
					context.RegularUserContext().Org,
					context.RegularUserContext().Space)).Should(Succeed())
		})
		defer func() {
			cf.AsUser(context.AdminUserContext(), time.Minute, func() {
				Eventually(runCf("delete-security-group", securityGroupName, "-f")).Should(Succeed())
			})
		}()

		Eventually(runCf("restart", appName), CF_PUSH_TIMEOUT).Should(Succeed())
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))

		// test app egress rules
		Eventually(curlResponse).Should(Equal(0))

		// unapply security group
		cf.AsUser(context.AdminUserContext(), time.Minute, func() {
			Eventually(
				runCf("unbind-security-group",
					securityGroupName, context.RegularUserContext().Org,
					context.RegularUserContext().Space)).
				Should(Succeed())
		})

		By("restarting it - without security group")
		Eventually(runCf("restart", appName), CF_PUSH_TIMEOUT).Should(Succeed())
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))

		// test app egress rules
		Eventually(curlResponse).Should(Equal(firstCurlError))
	})
})
