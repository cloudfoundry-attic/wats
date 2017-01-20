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
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/onsi/gomega/gbytes"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		ApiRequest("GET", endpoint, &response, timeout)

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

func DopplerUrl() string {
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
	Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))
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

func setTotalMemoryLimit(memoryLimit string) {
	type quotaDefinitions struct {
		Resources []struct {
			Entity struct {
				Name string `json:"name"`
			} `json:"entity"`
		} `json:"resources"`
	}

	var response quotaDefinitions
	ApiRequest("GET", "/v2/quota_definitions", &response, DEFAULT_TIMEOUT)

	var quotaDefinitionName string
	for _, r := range response.Resources {
		if r.Entity.Name != "default" && r.Entity.Name != "runaway" {
			quotaDefinitionName = r.Entity.Name
		}
	}
	Expect(quotaDefinitionName).ToNot(BeEmpty())

	AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		cf.Cf("update-quota", quotaDefinitionName, "-m", memoryLimit)
	})
}
