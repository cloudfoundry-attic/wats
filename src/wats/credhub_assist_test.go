package wats

import (
	"io/ioutil"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	archive_helpers "code.cloudfoundry.org/archiver/extractor/test_helper"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
)

type Config interface {
	GetStack() string
	GetAppsDomain() string
	GetAdminUser() string
	GetAdminPassword() string
}

var _ = CredhubDescribe("CredHub Integration", func() {
	var (
		chBrokerAppName string
		chServiceName   string
		instanceName    string
		appStartSession *Session
		config          Config
		err             error
	)

	BeforeEach(func() {
		config, err = LoadWatsConfig()
		Expect(err).NotTo(HaveOccurred())

		environment.RegularUserContext().TargetSpace()
		cf.Cf("target", "-o", environment.RegularUserContext().Org)
		Expect(string(cf.Cf("running-environment-variable-group").Wait(DEFAULT_TIMEOUT).Out.Contents())).To(ContainSubstring("CREDHUB_API"), "CredHub API environment not set")

		chBrokerAppName = generator.PrefixedRandomName("WATS", "BRKR-CH")

		Expect(cf.Cf(
			"push", chBrokerAppName,
			"-b", goBuildPackURL,
			"-m", "256m",
			"-p", "../../assets/credhub-service-broker",
			"-f", "../../assets/credhub-service-broker/manifest.yml",
			"-d", config.GetAppsDomain(),
		).Wait(CF_PUSH_TIMEOUT)).To(Exit(0), "failed pushing credhub-enabled service broker")

		chServiceName = generator.PrefixedRandomName("WATS", "SERVICE-NAME")
		Expect(cf.Cf(
			"set-env", chBrokerAppName,
			"SERVICE_NAME", chServiceName,
		).Wait(DEFAULT_TIMEOUT)).To(Exit(0), "failed setting SERVICE_NAME env var on credhub-enabled service broker")

		Expect(cf.Cf(
			"restart", chBrokerAppName,
		).Wait(CF_PUSH_TIMEOUT)).To(Exit(0), "failed restarting credhub-enabled service broker")

		workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
			serviceUrl := "https://" + chBrokerAppName + "." + config.GetAppsDomain()
			createServiceBroker := cf.Cf("create-service-broker", chBrokerAppName, config.GetAdminUser(), config.GetAdminPassword(), serviceUrl).Wait(DEFAULT_TIMEOUT)
			Expect(createServiceBroker).To(Exit(0), "failed creating credhub-enabled service broker")

			enableAccess := cf.Cf("enable-service-access", chServiceName, "-o", environment.RegularUserContext().Org).Wait(DEFAULT_TIMEOUT)
			Expect(enableAccess).To(Exit(0), "failed to enable service access for credhub-enabled broker")

			environment.RegularUserContext().TargetSpace()
			instanceName = generator.PrefixedRandomName("WATS", "SVIN-CH")
			createService := cf.Cf("create-service", chServiceName, "credhub-read-plan", instanceName).Wait(DEFAULT_TIMEOUT)
			Expect(createService).To(Exit(0), "failed creating credhub enabled service")
		})
	})

	AfterEach(func() {
		workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
			environment.RegularUserContext().TargetSpace()

			Expect(cf.Cf("purge-service-instance", instanceName, "-f").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
			Expect(cf.Cf("delete-service-broker", chBrokerAppName, "-f").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
			Expect(cf.Cf("delete", chBrokerAppName, "-f").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
		})
	})

	bindServiceAndStartApp := func(appName string) {
		Expect(chServiceName).ToNot(Equal(""))
		setServiceName := cf.Cf("set-env", appName, "SERVICE_NAME", chServiceName).Wait(DEFAULT_TIMEOUT)
		Expect(setServiceName).To(Exit(0), "failed setting SERVICE_NAME env var on app")

		workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
			environment.RegularUserContext().TargetSpace()

			bindService := cf.Cf("bind-service", appName, instanceName).Wait(DEFAULT_TIMEOUT)
			Expect(bindService).To(Exit(0), "failed binding app to service")
		})
		appStartSession = cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)
		Expect(appStartSession).To(Exit(0))
	}

	Context("during staging", func() {
		var (
			buildpackName string
			appName       string

			buildpackPath        string
			buildpackArchivePath string

			tmpdir string
		)
		BeforeEach(func() {
			workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
				buildpackName = generator.PrefixedRandomName("WATS", "BPK")
				appName = generator.PrefixedRandomName("WATS", "APP")

				var err error
				tmpdir, err = ioutil.TempDir("", "buildpack_env")
				Expect(err).ToNot(HaveOccurred())

				buildpackPath, err = ioutil.TempDir(tmpdir, "matching-buildpack")
				Expect(err).ToNot(HaveOccurred())

				buildpackArchivePath = path.Join(buildpackPath, "buildpack.zip")

				archive_helpers.CreateZipArchive(buildpackArchivePath, []archive_helpers.ArchiveFile{
					{
						Name: "bin/compile",
						Body: "",
					},
					{
						Name: "bin/detect",
						Body: ``,
					},
					{
						Name: "bin/compile.bat",
						Body: `echo COMPILING... really just dumping env...
cmd /C set
`,
					},
					{
						Name: "bin/detect.bat",
						Body: ``,
					},
					{
						Name: "bin/release.bat",
						Body: `@echo off
echo ---
echo default_process_types:
echo   web: webapp.exe
`,
					},
				})

				createBuildpack := cf.Cf("create-buildpack", buildpackName, buildpackArchivePath, "0").Wait(DEFAULT_TIMEOUT)
				Expect(createBuildpack).Should(Exit(0))
				Expect(createBuildpack).Should(Say("Creating"))
				Expect(createBuildpack).Should(Say("OK"))
				Expect(createBuildpack).Should(Say("Uploading"))
				Expect(createBuildpack).Should(Say("OK"))
			})
			Expect(cf.Cf("push", appName,
				"--no-start",
				"-b", buildpackName,
				"-m", "256m",
				"-p", "../../assets/webapp",
				"-d", config.GetAppsDomain(),
				"-s", config.GetStack(),
			).Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			bindServiceAndStartApp(appName)
		})

		AfterEach(func() {
			Expect(cf.Cf("delete", appName, "-f", "-r").Wait(DEFAULT_TIMEOUT)).To(Exit(0))

			workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
				Expect(cf.Cf("delete-buildpack", buildpackName, "-f").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
			})

			os.RemoveAll(tmpdir)
		})

		NonAssistedCredhubDescribe("", func() {
			It("still contains CredHub references in VCAP_SERVICES", func() {
				Expect(appStartSession).NotTo(Say("pinkyPie"))
				Expect(appStartSession).NotTo(Say("rainbowDash"))
				Expect(appStartSession).To(Say("credhub-ref"))
			})
		})

		AssistedCredhubDescribe("", func() {
			It("has CredHub references in VCAP_SERVICES interpolated", func() {
				Expect(appStartSession).To(Say(`{"password":"rainbowDash","user-name":"pinkyPie"}`))
				Expect(appStartSession).NotTo(Say("credhub-ref"))
			})
		})
	})

	Context("during runtime", func() {
		Describe("service bindings to credhub enabled broker", func() {
			var appName, appURL string
			BeforeEach(func() {
				appName = generator.PrefixedRandomName("WATS", "APP-CH")
				appURL = "https://" + appName + "." + config.GetAppsDomain()
			})

			AfterEach(func() {
				workflowhelpers.AsUser(environment.AdminUserContext(), DEFAULT_TIMEOUT, func() {
					environment.RegularUserContext().TargetSpace()
					unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(DEFAULT_TIMEOUT)
					Expect(unbindService).To(Exit(0), "failed unbinding app and service")

					Expect(cf.Cf("delete", appName, "-f", "-r").Wait(CF_PUSH_TIMEOUT)).To(Exit(0))
				})
			})

			AssistedCredhubDescribe("", func() {
				BeforeEach(func() {
					createApp := pushNora(appName).Wait(CF_PUSH_TIMEOUT)
					Expect(createApp).To(Exit(0), "failed creating credhub-enabled app")
					bindServiceAndStartApp(appName)
				})

				It("the broker returns credhub-ref in the credentials block", func() {
					appEnv := string(cf.Cf("env", appName).Wait(DEFAULT_TIMEOUT).Out.Contents())
					Expect(appEnv).To(ContainSubstring("credentials"), "credential block missing from service")
					Expect(appEnv).To(ContainSubstring("credhub-ref"), "credhub-ref not found")
				})

				It("the bound app gets CredHub refs in VCAP_SERVICES interpolated", func() {
					curlCmd := helpers.CurlSkipSSL(true, appURL+"/env/VCAP_SERVICES").Wait(DEFAULT_TIMEOUT)
					Expect(curlCmd).To(Exit(0))

					bytes := curlCmd.Out.Contents()
					Expect(string(bytes)).To(ContainSubstring(`rainbowDash`))
					Expect(string(bytes)).To(ContainSubstring(`pinkyPie`))
				})
			})
		})
	})
})
