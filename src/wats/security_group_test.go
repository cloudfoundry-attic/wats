package wats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

func unbindSecurityGroups() []string {
	var securityGroups []string

	AsUser(environment.AdminUserContext(), time.Minute, func() {
		out, err := runCfWithOutput("curl", "/v2/config/running_security_groups")
		Expect(err).NotTo(HaveOccurred())
		var result map[string]interface{}
		err = json.Unmarshal(out.Contents(), &result)
		Expect(err).NotTo(HaveOccurred())

		resources := result["resources"].([]interface{})
		for _, group := range resources {
			foo := group.(map[string]interface{})
			entity := foo["entity"].(map[string]interface{})
			name := entity["name"].(string)
			securityGroups = append(securityGroups, name)
			_, err = runCfWithOutput("unbind-security-group", name)
			Expect(err).NotTo(HaveOccurred())
		}
	})
	return securityGroups
}

func bindSecurityGroups(groups []string) {
	AsUser(environment.AdminUserContext(), time.Minute, func() {
		for _, group := range groups {
			_, err := runCfWithOutput("bind-security-group", group)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed to recreate running-security-group %s", group))
		}
	})
}

var _ = Describe("Security Groups", func() {
	type NoraCurlResponse struct {
		Stdout     string
		Stderr     string
		ReturnCode int `json:"return_code"`
	}

	// this test assumes the default running security groups block access to the DEAs
	// the test takes advantage of the fact that the DEA ip address and internal container ip address
	//  are discoverable via the cc api and nora's myip endpoint
	It("allows traffic and then blocks traffic", func() {
		groups := unbindSecurityGroups()
		defer bindSecurityGroups(groups)

		By("pushing it")
		Eventually(pushNora(appName), CF_PUSH_TIMEOUT).Should(Succeed())

		By("staging and running it on Diego")
		enableDiego(appName)
		Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())

		By("verifying it's up")
		Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))

		secureAddress := config.GetSecureAddress()
		secureHost, securePort, err := net.SplitHostPort(secureAddress)
		Expect(err).NotTo(HaveOccurred())

		// test app egress rules
		curlResponse := func() int {
			var noraCurlResponse NoraCurlResponse
			resp := helpers.CurlApp(config, appName, fmt.Sprintf("/curl/%s/%s", secureHost, securePort))
			json.Unmarshal([]byte(resp), &noraCurlResponse)
			return noraCurlResponse.ReturnCode
		}
		firstCurlError := curlResponse()
		Expect(firstCurlError).ShouldNot(Equal(0))

		// apply security group
		rules := fmt.Sprintf(`[{"destination":"%s","ports":"%s","protocol":"tcp"}]`, secureHost, securePort)

		file, _ := ioutil.TempFile(os.TempDir(), "DATS-sg-rules")
		defer os.Remove(file.Name())
		file.WriteString(rules)
		file.Close()

		rulesPath := file.Name()
		securityGroupName := fmt.Sprintf("DATS-SG-%s", generator.PrefixedRandomName(config.GetNamePrefix(), "SECURITY-GROUP"))

		AsUser(environment.AdminUserContext(), time.Minute, func() {
			Eventually(runCf("create-security-group", securityGroupName, rulesPath)).Should(Succeed())
			Eventually(
				runCf("bind-security-group",
					securityGroupName,
					environment.RegularUserContext().Org,
					environment.RegularUserContext().Space)).Should(Succeed())
		})
		defer func() {
			AsUser(environment.AdminUserContext(), time.Minute, func() {
				Eventually(runCf("delete-security-group", securityGroupName, "-f")).Should(Succeed())
			})
		}()

		Eventually(runCf("restart", appName), CF_PUSH_TIMEOUT).Should(Succeed())
		Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))

		// test app egress rules
		Eventually(curlResponse).Should(Equal(0))

		// unapply security group
		AsUser(environment.AdminUserContext(), time.Minute, func() {
			Eventually(
				runCf("unbind-security-group",
					securityGroupName, environment.RegularUserContext().Org,
					environment.RegularUserContext().Space)).
				Should(Succeed())
		})

		By("restarting it - without security group")
		Eventually(runCf("restart", appName), CF_PUSH_TIMEOUT).Should(Succeed())
		Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))

		// test app egress rules
		Eventually(curlResponse).Should(Equal(firstCurlError))
	})
})
