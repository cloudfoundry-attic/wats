package wats

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

const (
	DEFAULT_TIMEOUT = 45 * time.Second
	CF_PUSH_TIMEOUT = 3 * time.Minute
)

var (
	appName            string
	config             *watsConfig
	environment        *ReproducibleTestSuiteSetup
	hwcBuildPackURL    = "https://github.com/cloudfoundry-incubator/hwc-buildpack/releases/download/v2.1.0/hwc_buildpack-cached-v2.1.0.zip"
	binaryBuildPackURL = "https://github.com/cloudfoundry/binary-buildpack/releases/download/v1.0.8/binary_buildpack-cached-v1.0.8.zip"
)

func guidForAppName(appName string) string {
	cfApp := cf.Cf("app", appName, "--guid")
	Expect(cfApp.Wait()).To(Exit(0))

	appGuid := strings.TrimSpace(string(cfApp.Out.Contents()))
	Expect(appGuid).NotTo(Equal(""))
	return appGuid
}

func guidForSpaceName(spaceName string) string {
	cfSpace := cf.Cf("space", spaceName, "--guid")
	Expect(cfSpace.Wait()).To(Exit(0))

	spaceGuid := strings.TrimSpace(string(cfSpace.Out.Contents()))
	Expect(spaceGuid).NotTo(Equal(""))
	return spaceGuid
}

func enableDiego(appName string) {
	guid := guidForAppName(appName)
	Eventually(cf.Cf("curl", "/v2/apps/"+guid, "-X", "PUT", "-d", `{"diego": true}`)).Should(Exit(0))
}

func disableHealthCheck(appName string) {
	guid := guidForAppName(appName)
	Eventually(cf.Cf("curl", "/v2/apps/"+guid, "-X", "PUT", "-d", `{"health_check_type":"none"}`)).Should(Exit(0))
}

func TestDiegoWindows(t *testing.T) {
	RegisterFailHandler(Fail)

	SetDefaultEventuallyTimeout(time.Minute)
	SetDefaultEventuallyPollingInterval(time.Second)

	var err error
	config, err = LoadWatsConfig()
	if err != nil {
		t.Fatalf("could not load WATS config", err)
	}

	if config.NumWindowsCells == 0 {
		t.Fatalf("Please provide 'num_windows_cells' as a property in the integration config JSON (The number of windows cells in tested deployment)")
	}

	environment = NewTestSuiteSetup(config)

	BeforeSuite(func() {
		environment.Setup()
		binaryBuildpackVersion := getBuildpackVersion("binary_buildpack")
		if versionGreaterThan(binaryBuildpackVersion, 1, 0, 7) {
			binaryBuildPackURL = "binary_buildpack"
		}

		hwcBuildpackVersion := getBuildpackVersion("hwc_buildpack")
		if versionGreaterThan(hwcBuildpackVersion, 2, 0, 0) {
			hwcBuildPackURL = "hwc_buildpack"
		}
	})

	AfterSuite(func() {
		environment.Teardown()
	})

	BeforeEach(func() {
		Eventually(cf.Cf("apps").Out).Should(Say("No apps found"))
		appName = generator.PrefixedRandomName(config.GetNamePrefix(), "APP")
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	componentName := "DiegoWindows"

	rs := []Reporter{}

	if config.GetArtifactsDirectory() != "" {
		helpers.EnableCFTrace(config, componentName)
		rs = append(rs, helpers.NewJUnitReporter(config, componentName))
	}

	RunSpecsWithDefaultAndCustomReporters(t, componentName, rs)
}
func getBuildpackVersion(name string) string {
	buildpack := cf.Cf("curl", fmt.Sprintf("/v2/buildpacks?q=name:%s", name))
	Expect(buildpack.Wait()).To(Exit(0))
	type Buildpack struct {
		Resources []struct {
			Entity struct {
				FileName string
			}
		}
	}
	var b Buildpack
	Expect(json.Unmarshal(buildpack.Out.Contents(), &b)).To(Succeed())
	if len(b.Resources) == 0 {
		return "0.0.0"
	}
	re := regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+`)
	return re.FindString(b.Resources[0].Entity.FileName)
}

func versionGreaterThan(version string, inputMajor, inputMinor, inputPatch int) bool {
	versions := strings.Split(version, ".")
	major, _ := strconv.Atoi(versions[0])
	if major > inputMajor {
		return true
	}
	minor, _ := strconv.Atoi(versions[1])
	if minor > inputMinor {
		return true
	}
	patch, _ := strconv.Atoi(versions[2])
	if patch > inputPatch {
		return true
	}
	return false
}
