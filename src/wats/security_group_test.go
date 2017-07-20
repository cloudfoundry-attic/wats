package wats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

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
			_, err = runCfWithOutput("unbind-running-security-group", name)
			Expect(err).NotTo(HaveOccurred())
		}
	})
	return securityGroups
}

func bindSecurityGroups(groups []string) {
	AsUser(environment.AdminUserContext(), time.Minute, func() {
		for _, group := range groups {
			_, err := runCfWithOutput("bind-running-security-group", group)
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

	It("allows traffic and then blocks traffic", func() {
		groups := unbindSecurityGroups()
		defer bindSecurityGroups(groups)

		By("pushing it")
		Expect(pushNora(appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

		By("staging and running it on Diego")
		enableDiego(appName)
		Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

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
			Expect(cf.Cf("create-security-group", securityGroupName, rulesPath).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			Expect(cf.Cf("bind-security-group",
				securityGroupName,
				environment.RegularUserContext().Org,
				environment.RegularUserContext().Space).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
		})
		defer func() {
			AsUser(environment.AdminUserContext(), time.Minute, func() {
				Expect(cf.Cf("delete-security-group", securityGroupName, "-f").Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			})
		}()

		Expect(cf.Cf("restart", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
		Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))

		// test app egress rules
		Eventually(curlResponse).Should(Equal(0))

		// unapply security group
		AsUser(environment.AdminUserContext(), time.Minute, func() {
			Expect(cf.Cf("unbind-security-group",
				securityGroupName, environment.RegularUserContext().Org,
				environment.RegularUserContext().Space).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
		})

		By("restarting it - without security group")
		Expect(cf.Cf("restart", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
		Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))

		// test app egress rules
		Eventually(curlResponse).Should(Equal(firstCurlError))
	})

	Context("when an icmp rule is applied", func() {
		var (
			icmpRuleFile      string
			securityGroupName string
		)

		BeforeEach(func() {
			icmpRule := `[{"code": 1,"destination":"0.0.0.0/0","protocol":"icmp","type":0}]`
			securityGroupName = fmt.Sprintf("DATS-SG-%s", generator.PrefixedRandomName(config.GetNamePrefix(), "SECURITY-GROUP"))

			file, err := ioutil.TempFile("", securityGroupName)
			Expect(err).ToNot(HaveOccurred())
			_, err = file.WriteString(icmpRule)
			Expect(err).ToNot(HaveOccurred())
			Expect(file.Close()).To(Succeed())

			icmpRuleFile = file.Name()

			AsUser(environment.AdminUserContext(), 2*time.Minute, func() {
				Expect(cf.Cf("create-security-group", securityGroupName, icmpRuleFile).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
				Expect(cf.Cf("bind-security-group",
					securityGroupName,
					environment.RegularUserContext().Org,
					environment.RegularUserContext().Space).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			})
		})

		AfterEach(func() {
			Expect(os.Remove(icmpRuleFile)).To(Succeed())
			AsUser(environment.AdminUserContext(), 2*time.Minute, func() {
				Expect(cf.Cf("unbind-security-group",
					securityGroupName, environment.RegularUserContext().Org,
					environment.RegularUserContext().Space).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
				Expect(cf.Cf("delete-security-group", securityGroupName, "-f").Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			})
		})

		It("ignores the rule and can push an app", func() {
			By("pushing it", func() {
				Expect(pushNora(appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			})

			By("staging and running it on Diego", func() {
				enableDiego(appName)
				Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			})

			By("verifying it's up", func() {
				Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))
			})
		})
	})
})
