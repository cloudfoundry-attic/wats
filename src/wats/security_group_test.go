package wats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

type Destination struct {
	IP       string `json:"destination"`
	Port     string `json:"ports,omitempty"`
	Protocol string `json:"protocol"`
	Code     int    `json:"code,omitempty"`
	Type     int    `json:"type,omitempty"`
}

func createSecurityGroup(allowedDestinations ...Destination) string {
	file, _ := ioutil.TempFile(os.TempDir(), "WATS-sg-rules")
	defer os.Remove(file.Name())
	Expect(json.NewEncoder(file).Encode(allowedDestinations)).To(Succeed())

	rulesPath := file.Name()
	securityGroupName := fmt.Sprintf("WATS-SG-%s", generator.PrefixedRandomName(config.GetNamePrefix(), "SECURITY-GROUP"))

	workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		Expect(cf.Cf("create-security-group", securityGroupName, rulesPath).Wait(DEFAULT_TIMEOUT)).To(gexec.Exit(0))
	})

	return securityGroupName
}

func bindSecurityGroup(securityGroupName, orgName, spaceName string) {
	By("Applying security group")
	workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		Expect(cf.Cf("bind-security-group", securityGroupName, orgName, spaceName).Wait(DEFAULT_TIMEOUT)).To(gexec.Exit(0))
	})
}

func unbindSecurityGroup(securityGroupName, orgName, spaceName string) {
	By("Unapplying security group")
	workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		Expect(cf.Cf("unbind-security-group", securityGroupName, orgName, spaceName).Wait(DEFAULT_TIMEOUT)).To(gexec.Exit(0))
	})
}

func deleteSecurityGroup(securityGroupName string) {
	workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		Expect(cf.Cf("delete-security-group", securityGroupName, "-f").Wait(DEFAULT_TIMEOUT)).To(gexec.Exit(0))
	})
}

type NoraCurlResponse struct {
	Stdout     string
	Stderr     string
	ReturnCode int `json:"return_code"`
}

func noraCurlResponse(appName, host, port string) int {
	var noraCurlResponse NoraCurlResponse
	resp := helpers.CurlApp(config, appName, fmt.Sprintf("/curl/%s/%s", host, port))
	Expect(json.Unmarshal([]byte(resp), &noraCurlResponse)).To(Succeed())
	return noraCurlResponse.ReturnCode
}

var _ = Describe("Security Groups", func() {
	BeforeEach(func() {
		By("pushing it")
		Expect(pushNora(appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

		By("staging and running it on Diego")
		Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

		By("verifying it's up")
		Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))
	})

	Context("when a tcp (or udp) rule is applied", func() {
		var (
			securityGroupName string
			secureHost        string
			securePort        string
		)

		BeforeEach(func() {
			By("Asserting default running security group configuration for traffic to private ip addresses")
			var err error
			secureAddress := config.GetSecureAddress()
			secureHost, securePort, err = net.SplitHostPort(secureAddress)
			Expect(err).NotTo(HaveOccurred())
			Expect(noraCurlResponse(appName, secureHost, securePort)).ShouldNot(Equal(0))

			By("Asserting default running security group configuration from a running container to a public ip")
			Expect(noraCurlResponse(appName, "www.google.com", "80")).Should(Equal(0))
		})

		AfterEach(func() {
			deleteSecurityGroup(securityGroupName)
		})

		It("allows traffic to a private ip after applying a policy and blocks it when the policy is removed", func() {
			rule := Destination{IP: secureHost, Port: securePort, Protocol: "tcp"}
			securityGroupName = createSecurityGroup(rule)
			bindSecurityGroup(securityGroupName, environment.RegularUserContext().Org, environment.RegularUserContext().Space)

			Expect(cf.Cf("restart", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))

			Expect(noraCurlResponse(appName, secureHost, securePort)).Should(Equal(0))

			unbindSecurityGroup(securityGroupName, environment.RegularUserContext().Org, environment.RegularUserContext().Space)

			Expect(cf.Cf("restart", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))

			Expect(noraCurlResponse(appName, secureHost, securePort)).ShouldNot(Equal(0))
		})
	})

	Context("when an icmp rule is applied", func() {
		var (
			icmpGroupName string
		)

		BeforeEach(func() {
			rule := Destination{IP: "0.0.0.0/0", Protocol: "icmp", Code: -1, Type: -1}
			icmpGroupName = createSecurityGroup(rule)
			bindSecurityGroup(icmpGroupName, environment.RegularUserContext().Org, environment.RegularUserContext().Space)
		})

		AfterEach(func() {
			unbindSecurityGroup(icmpGroupName, environment.RegularUserContext().Org, environment.RegularUserContext().Space)
			deleteSecurityGroup(icmpGroupName)
		})

		It("ignores the rule and can push an app", func() {
			By("pushing it", func() {
				Expect(pushNora(appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			})

			By("staging and running it on Diego", func() {
				Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			})

			By("verifying it's up", func() {
				Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))
			})
		})
	})
})
