package wats

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
)

// Copied from : cf-test-helpers/runner/run.go
const timeFormat = "2006-01-02 15:04:05.00 (MST)"

func Run(executable string, env []string, args ...string) *gexec.Session {
	cmd := exec.Command(executable, args...)
	cmd.Env = env

	fmt.Fprintf(GinkgoWriter, "\n[%s]> %s %s\n", time.Now().UTC().Format(timeFormat), executable, strings.Join(cmd.Args, " "))

	// innerRun
	sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return sess
}

func pushWorker(appName string) func() error {
	return runCf(
		"push", appName,
		"-p", "../assets/worker",
		"--no-start",
		"--no-route",
		"-m", "2g",
		"-b", "https://github.com/ryandotsmith/null-buildpack.git",
		"-s", "windows2012R2")
}

var _ = Describe("apps without a port", func() {
	var appName string
	var logs *gexec.Session

	BeforeEach(func() {
		appName = generator.RandomName()

		Eventually(pushWorker(appName), CF_PUSH_TIMEOUT).Should(Succeed())
		enableDiego(appName)
		disableSsh(appName)
		disableHealthCheck(appName)
		logs = cf.Cf("logs", appName)
		// if healthcheck ran, the following will fail. `cf start` will wait
		// for the heathcheck to succeed.
		Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(gexec.Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(gexec.Exit(0))
	})

	It("run (and don't run healthcheck)", func() {
		Eventually(logs.Out).Should(Say("Running Worker 1"))
		Eventually(logs.Out).Should(Say("Running Worker 10"))
		Expect(logs.Out).ToNot(Say("healthcheck"))
	})
})
