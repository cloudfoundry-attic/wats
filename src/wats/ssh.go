package wats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("SSH", func() {
	BeforeEach(func() {
		if config.GetStack() == "windows2012R2" {
			Skip("cf ssh does not work on windows2012R2")
		}

		Expect(pushNora(appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
		Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
		enableSSH(appName)
	})

	Describe("ssh", func() {
		Context("with multiple instances", func() {
			BeforeEach(func() {
				Expect(cf.Cf("scale", appName, "-i", "2").Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
				Eventually(func() string {
					return helpers.CurlApp(config, appName, "/env/INSTANCE_INDEX")
				}, CF_PUSH_TIMEOUT).Should(Equal(`"1"`))
			})

			It("can ssh to the second instance", func() {
				envCmd := cf.Cf("ssh", "-v", "-i", "1", appName, "-c", "set && set 1>&2")
				Expect(envCmd.Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

				output := string(envCmd.Out.Contents())
				stdErr := string(envCmd.Err.Contents())

				Expect(string(output)).To(MatchRegexp(fmt.Sprintf(`VCAP_APPLICATION=.*"application_name":"%s"`, appName)))
				Expect(string(output)).To(MatchRegexp("INSTANCE_INDEX=1"))

				Expect(string(stdErr)).To(MatchRegexp(fmt.Sprintf(`VCAP_APPLICATION=.*"application_name":"%s"`, appName)))
				Expect(string(stdErr)).To(MatchRegexp("INSTANCE_INDEX=1"))

				logCmd := cf.Cf("logs", appName, "--recent")
				Expect(logCmd.Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
				output = string(logCmd.Out.Contents())
				Expect(output).To(ContainSubstring("Successful remote access"))

				eventsCmd := cf.Cf("events", appName).Wait(CF_PUSH_TIMEOUT)
				Expect(eventsCmd.Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
				output = string(eventsCmd.Out.Contents())
				Expect(output).To(ContainSubstring("audit.app.ssh-authorized"))
			})
		})

		It("can execute a remote command in the container", func() {
			envCmd := cf.Cf("ssh", "-v", appName, "-c", "set && set 1>&2")
			Expect(envCmd.Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

			output := string(envCmd.Out.Contents())
			stdErr := string(envCmd.Err.Contents())

			Expect(string(output)).To(MatchRegexp(fmt.Sprintf(`VCAP_APPLICATION=.*"application_name":"%s"`, appName)))
			Expect(string(output)).To(MatchRegexp("INSTANCE_INDEX=0"))

			Expect(string(stdErr)).To(MatchRegexp(fmt.Sprintf(`VCAP_APPLICATION=.*"application_name":"%s"`, appName)))
			Expect(string(stdErr)).To(MatchRegexp("INSTANCE_INDEX=0"))

			logCmd := cf.Cf("logs", appName, "--recent")
			Expect(logCmd.Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			output = string(logCmd.Out.Contents())
			Expect(output).To(ContainSubstring("Successful remote access"))

			eventsCmd := cf.Cf("events", appName).Wait(CF_PUSH_TIMEOUT)
			Expect(eventsCmd.Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			output = string(eventsCmd.Out.Contents())
			Expect(output).To(ContainSubstring("audit.app.ssh-authorized"))
		})

		It("runs an interactive session when no command is provided", func() {
			envCmd := exec.Command("cf", "ssh", "-v", appName)

			stdin, err := envCmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			stdout, err := envCmd.StdoutPipe()
			Expect(err).NotTo(HaveOccurred())

			err = envCmd.Start()
			Expect(err).NotTo(HaveOccurred())

			_, err = stdin.Write([]byte("set\n"))
			Expect(err).NotTo(HaveOccurred())

			err = stdin.Close()
			Expect(err).NotTo(HaveOccurred())

			output, err := ioutil.ReadAll(stdout)
			Expect(err).NotTo(HaveOccurred())

			err = envCmd.Wait()
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(MatchRegexp(fmt.Sprintf(`VCAP_APPLICATION=.*"application_name":"%s"`, appName)))
			Expect(string(output)).To(MatchRegexp("INSTANCE_INDEX=0"))

			logCmd := cf.Cf("logs", appName, "--recent")
			Expect(logCmd.Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			output = logCmd.Out.Contents()
			Expect(string(output)).To(ContainSubstring("Successful remote access"))

			eventsCmd := cf.Cf("events", appName).Wait(CF_PUSH_TIMEOUT)
			Expect(eventsCmd.Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			output = eventsCmd.Out.Contents()
			Expect(string(output)).To(ContainSubstring("audit.app.ssh-authorized"))
		})

		It("allows local port forwarding", func() {
			listenCmd := exec.Command("cf", "ssh", "-v", "-L", "127.0.0.1:61007:localhost:8080", appName)

			stdin, err := listenCmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			err = listenCmd.Start()
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() string {
				curl := helpers.Curl(config, "http://127.0.0.1:61007/").Wait(CF_PUSH_TIMEOUT)
				return string(curl.Out.Contents())
			}, CF_PUSH_TIMEOUT).Should(ContainSubstring("hello i am nora"))

			err = stdin.Close()
			Expect(err).NotTo(HaveOccurred())

			err = listenCmd.Wait()
			Expect(err).NotTo(HaveOccurred())
		})

		It("records successful ssh attempts", func() {
			password := sshAccessCode()

			clientConfig := &ssh.ClientConfig{
				User:            fmt.Sprintf("cf:%s/%d", guidForAppName(appName), 0),
				Auth:            []ssh.AuthMethod{ssh.Password(password)},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}

			client, err := ssh.Dial("tcp", sshProxyAddress(), clientConfig)
			Expect(err).NotTo(HaveOccurred())

			session, err := client.NewSession()
			Expect(err).NotTo(HaveOccurred())

			output, err := session.CombinedOutput("set")
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(MatchRegexp(fmt.Sprintf(`VCAP_APPLICATION=.*"application_name":"%s"`, appName)))
			Expect(string(output)).To(MatchRegexp("INSTANCE_INDEX=0"))

			logCmd := cf.Cf("logs", appName, "--recent")
			Expect(logCmd.Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			output = logCmd.Out.Contents()
			Expect(string(output)).To(ContainSubstring("Successful remote access"))

			eventsCmd := cf.Cf("events", appName).Wait(CF_PUSH_TIMEOUT)
			Expect(eventsCmd.Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			output = eventsCmd.Out.Contents()
			Expect(string(output)).To(ContainSubstring("audit.app.ssh-authorized"))
		})

		It("records failed ssh attempts", func() {
			Eventually(cf.Cf("disable-ssh", appName), CF_PUSH_TIMEOUT).Should(gexec.Exit(0))

			password := sshAccessCode()
			clientConfig := &ssh.ClientConfig{
				User:            fmt.Sprintf("cf:%s/%d", guidForAppName(appName), 0),
				Auth:            []ssh.AuthMethod{ssh.Password(password)},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}

			_, err := ssh.Dial("tcp", sshProxyAddress(), clientConfig)
			Expect(err).To(HaveOccurred())

			eventsCmd := cf.Cf("events", appName).Wait(CF_PUSH_TIMEOUT)
			Expect(eventsCmd.Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			output := eventsCmd.Out.Contents()
			Expect(string(output)).To(ContainSubstring("audit.app.ssh-unauthorized"))
		})
	})

})

func enableSSH(appName string) {
	Expect(cf.Cf("enable-ssh", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
}

func sshAccessCode() string {
	getCode := cf.Cf("ssh-code")
	Eventually(getCode, CF_PUSH_TIMEOUT).Should(gexec.Exit(0))
	return strings.TrimSpace(string(getCode.Buffer().Contents()))
}

func sshProxyAddress() string {
	infoCommand := cf.Cf("curl", "/v2/info")
	Expect(infoCommand.Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

	type infoResponse struct {
		AppSSHEndpoint string `json:"app_ssh_endpoint"`
	}

	var response infoResponse
	err := json.Unmarshal(infoCommand.Buffer().Contents(), &response)
	Expect(err).NotTo(HaveOccurred())

	return response.AppSSHEndpoint
}
