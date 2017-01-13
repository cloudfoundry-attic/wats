package wats

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

func pushNoraWithOptions(appName string, instances int, memory string) func() error {
	return pushApp(appName, "../../assets/nora/NoraPublished", instances, memory)
}

func pushNora(appName string) func() error {
	return pushNoraWithOptions(appName, 1, "256m")
}

func appRunning(appName string, instances int, timeout time.Duration) func() error {
	return func() error {
		type StatsResponse map[string]struct {
			State string `json:"state"`
		}

		buf, err := runCfWithOutput("app", appName, "--guid")
		if err != nil {
			return err
		}
		appGuid := strings.Replace(string(buf.Contents()), "\n", "", -1)

		endpoint := fmt.Sprintf("/v2/apps/%s/stats", appGuid)

		var response StatsResponse
		cf.ApiRequest("GET", endpoint, &response, timeout)

		err = nil
		for k, v := range response {
			if v.State != "RUNNING" {
				err = errors.New(fmt.Sprintf("App %s instance %s is not running: State = %s", appName, k, v.State))
			}
		}
		return err
	}
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
		cfInfoBuffer, err := runCfWithOutput("curl", "/v2/info")
		Expect(err).NotTo(HaveOccurred())

		var cfInfo struct {
			DopplerLoggingEndpoint string `json:"doppler_logging_endpoint"`
		}

		err = json.NewDecoder(bytes.NewReader(cfInfoBuffer.Contents())).Decode(&cfInfo)
		Expect(err).NotTo(HaveOccurred())

		doppler = cfInfo.DopplerLoggingEndpoint
	}
	return doppler
}

func pushAndStartNora(appName string) {
	By("pushing it")
	Eventually(pushNora(appName), CF_PUSH_TIMEOUT).Should(Succeed())

	By("staging and running it on Diego")
	enableDiego(appName)
	Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())

	By("verifying it's up")
	Eventually(CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))
}

func pushApp(appName, path string, instances int, memory string) func() error {
	return runCf(
		"push", appName,
		"-p", path,
		"--no-start",
		"-i", strconv.Itoa(instances),
		"-m", memory,
		"-b", "binary_buildpack",
		"-s", "windows2012R2")
}
