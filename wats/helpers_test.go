package wats

import (
	"errors"
	"os"
	"strconv"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

func pushNoraWithOptions(appName string, instances int, memory string) func() error {
	return runCf(
		"push", appName,
		"-p", "../assets/nora/NoraPublished",
		"--no-start",
		"-i", strconv.Itoa(instances),
		"-m", memory,
		"-b", "https://github.com/ryandotsmith/null-buildpack.git",
		"-s", "windows2012R2")
}

func pushNora(appName string) func() error {
	return pushNoraWithOptions(appName, 1, "2g")
}

func runCfWithOutput(values ...string) (*gbytes.Buffer, error) {
	session := cf.Cf(values...)
	session.Wait(CF_PUSH_TIMEOUT)
	if session.ExitCode() == 0 {
		return session.Out, nil
	}

	return nil, errors.New("non zero exit code")
}

func runCf(values ...string) func() error {
	return func() error {
		_, err := runCfWithOutput(values...)
		return err
	}
}

func DopplerUrl(c Config) string {
	doppler := os.Getenv("DOPPLER_URL")
	if doppler == "" {
		doppler = "wss://doppler." + c.AppsDomain + ":4443"
	}
	return doppler
}

func pushAndStartNora(appName string) {
	By("pushing it")
	Eventually(pushNora(appName), CF_PUSH_TIMEOUT).Should(Succeed())

	By("staging and running it on Diego")
	enableDiego(appName)
	disableSsh(appName)
	Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())

	By("verifying it's up")
	Eventually(CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))
}
