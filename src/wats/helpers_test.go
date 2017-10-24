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
	"github.com/onsi/gomega/gexec"
)

const SkipCredhubMessage = `Skipping this test because Config.CredhubMode is not set to either 'assisted' or 'non-assisted'.
NOTE: Ensure instance identity credential is turned on and CredHub is deployed before enabling this test`
const SkipAssistedCredhubMessage = `Skipping this test because Config.CredhubMode is not set to 'assisted'.
NOTE: Ensure instance identity credential is turned on and CredHub is deployed before enabling this test`
const SkipNonAssistedCredhubMessage = `Skipping this test because Config.CredhubMode is not set to 'non-assisted'.
NOTE: Ensure instance identity credential is turned on and CredHub is deployed before enabling this test`

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

	return session.Out, fmt.Errorf("non zero exit code %d", session.ExitCode())
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

func CredhubDescribe(description string, callback func()) bool {
	return Describe("[credhub]", func() {
		BeforeEach(func() {
			if config, err := LoadWatsConfig(); err == nil && !config.GetIncludeCredhubAssisted() && !config.GetIncludeCredhubNonAssisted() {
				Skip(SkipCredhubMessage)
			}
		})
		Describe(description, callback)
	})
}

func AssistedCredhubDescribe(description string, callback func()) bool {
	return Describe("[assisted credhub]", func() {
		BeforeEach(func() {
			if config, err := LoadWatsConfig(); err == nil && !config.GetIncludeCredhubAssisted() {
				Skip(SkipAssistedCredhubMessage)
			}
		})
		Describe(description, callback)
	})
}

func NonAssistedCredhubDescribe(description string, callback func()) bool {
	return Describe("[non-assisted credhub]", func() {
		BeforeEach(func() {
			if config, err := LoadWatsConfig(); err == nil && !config.GetIncludeCredhubNonAssisted() {
				Skip(SkipNonAssistedCredhubMessage)
			}
		})
		Describe(description, callback)
	})
}

func pushAndStartNora(appName string) {
	By("pushing it")
	Expect(pushNora(appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

	By("staging and running it on Diego")
	enableDiego(appName)
	Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

	By("verifying it's up")
	Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))
}

func pushNora(appName string) *gexec.Session {
	return pushNoraWithOptions(appName, 1, "256m")
}

func pushNoraWithOptions(appName string, instances int, memory string) *gexec.Session {
	return pushApp(appName, "../../assets/nora/NoraPublished", instances, memory, hwcBuildPackURL)
}

func pushApp(appName, path string, instances int, memory, buildpack string) *gexec.Session {
	return cf.Cf(
		"push", appName,
		"-p", path,
		"--no-start",
		"-i", strconv.Itoa(instances),
		"-m", memory,
		"-b", buildpack,
		"-s", config.GetStack())
}

func setTotalMemoryLimit(memoryLimit string) {
	type quotaDefinitionUrl struct {
		Resources []struct {
			Entity struct {
				QuotaDefinitionUrl string `json:"quota_definition_url"`
			} `json:"entity"`
		} `json:"resources"`
	}

	orgEndpoint := fmt.Sprintf("/v2/organizations?q=name%%3A%s", environment.GetOrganizationName())
	var org quotaDefinitionUrl
	ApiRequest("GET", orgEndpoint, &org, DEFAULT_TIMEOUT)
	Expect(org.Resources).ToNot(BeEmpty())

	type quotaDefinition struct {
		Entity struct {
			Name string `json:"name"`
		} `json:"entity"`
	}
	var quota quotaDefinition
	ApiRequest("GET", org.Resources[0].Entity.QuotaDefinitionUrl, &quota, DEFAULT_TIMEOUT)
	Expect(quota.Entity.Name).ToNot(BeEmpty())

	AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		Expect(cf.Cf("update-quota", quota.Entity.Name, "-m", memoryLimit).Wait(DEFAULT_TIMEOUT)).To(gexec.Exit(0))
	})
}
