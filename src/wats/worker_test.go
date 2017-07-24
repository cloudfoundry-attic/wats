package wats

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
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

var _ = Describe("apps without a port", func() {
	var logs *gexec.Session

	BeforeEach(func() {
		workerPath, err := gexec.BuildWithEnvironment("../../assets/worker/worker.go", []string{"GOARCH=amd64", "GOOS=windows"})
		Expect(err).NotTo(HaveOccurred())
		Expect(cf.Cf("push", appName, "-p", filepath.Dir(workerPath), "-c", ".\\worker.go",
			"--no-start", "-b", binaryBuildPackURL, "-s", "windows2016").Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
		enableDiego(appName)
		disableHealthCheck(appName)
		logs = cf.Cf("logs", appName)
		// if healthcheck ran, the following will fail. `cf start` will wait
		// for the heathcheck to succeed.
		Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
	})

	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
	})

	It("run (and don't run healthcheck)", func() {
		Eventually(logs.Out).Should(Say("Running Worker 1"))
		Eventually(logs.Out).Should(Say("Running Worker 10"))
		Expect(logs.Out).ToNot(Say("healthcheck"))
	})
})
