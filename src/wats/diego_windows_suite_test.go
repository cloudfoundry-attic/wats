package wats

import (
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
	DEFAULT_TIMEOUT      = 45 * time.Second
	CF_PUSH_TIMEOUT      = 3 * time.Minute
	HWC_BUILDPACK_URL    = "https://github.com/greenhouse-org/hwc-buildpack/releases/download/v1.0.0/hwc-buildpack-1.0.0.zip"
	BINARY_BUILDPACK_URL = "https://github.com/greenhouse-org/binary-buildpack/releases/download/develop-v1.0.8-rc.1/binary_buildpack-develop-1.0.8-rc.1.zip"
)

var (
	appName     string
	config      *watsConfig
	environment *ReproducibleTestSuiteSetup
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
