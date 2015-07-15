package wats

import (
	"regexp"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe("Application Lifecycle", func() {
	var appName string

	apps := func() *Session {
		return cf.Cf("apps").Wait()
	}

	BeforeEach(func() {
		appName = generator.RandomName()
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	reportedIDs := func(instances int) map[string]bool {
		timer := time.NewTimer(time.Second * 120)
		defer timer.Stop()
		run := true
		go func() {
			<-timer.C
			run = false
		}()

		seenIDs := map[string]bool{}
		for len(seenIDs) != instances && run == true {
			seenIDs[helpers.CurlApp(appName, "/id")] = true
		}

		return seenIDs
	}

	differentIDsFrom := func(idsBefore map[string]bool) []string {
		differentIDs := []string{}

		for id, _ := range reportedIDs(len(idsBefore)) {
			if !idsBefore[id] {
				differentIDs = append(differentIDs, id)
			}
		}

		return differentIDs
	}

	Describe("An app staged on Diego and running on Diego", func() {
		It("exercises the app through its lifecycle", func() {
			By("pushing it", func() {
				Eventually(pushNora(appName), CF_PUSH_TIMEOUT).Should(Succeed())
			})

			By("staging and running it on Diego", func() {
				enableDiego(appName)
				disableSsh(appName)
				Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())
			})

			By("verifying it's up", func() {
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))
			})

			By("veriying reported disk/memory usage", func() {
				// #0   running   2015-06-10 02:22:39 PM   0.0%   48.7M of 2G   14M of 1G
				var metrics = regexp.MustCompile(`running.*(?:[\d\.]+)%\s+([\d\.]+)[MG]? of (?:[\d\.]+)[MG]\s+([\d\.]+)[MG]? of (?:[\d\.]+)[MG]`)
				memdisk := func() (mem, disk float64) {
					output, err := runCfWithOutput("app", appName)
					Expect(err).ToNot(HaveOccurred())
					arr := metrics.FindStringSubmatch(string(output.Contents()))
					mem, err = strconv.ParseFloat(arr[1], 64)
					Expect(err).ToNot(HaveOccurred())
					disk, err = strconv.ParseFloat(arr[2], 64)
					Expect(err).ToNot(HaveOccurred())
					return
				}
				Eventually(func() float64 { m, _ := memdisk(); return m }).Should(BeNumerically(">", 0.0))
				Expect(func() float64 { _, d := memdisk(); return d }()).Should(BeNumerically(">", 0.0))
			})

			By("stopping it", func() {
				Eventually(runCf("stop", appName)).Should(Succeed())
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("404"))
			})

			By("setting an environment variable", func() {
				Eventually(runCf("set-env", appName, "FOO", "bar")).Should(Succeed())
			})

			By("starting it", func() {
				Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))
			})

			By("checking custom env variables are available", func() {
				Eventually(func() string {
					return helpers.CurlAppWithTimeout(appName, "/env/FOO", 30*time.Second)
				}).Should(ContainSubstring("bar"))
			})

			By("scaling it", func() {
				Eventually(runCf("scale", appName, "-i", "2")).Should(Succeed())
				Eventually(apps).Should(Say("2/2"))
				Expect(cf.Cf("app", appName).Wait()).ToNot(Say("insufficient resources"))
			})

			By("restarting an instance", func() {
				idsBefore := reportedIDs(2)
				Expect(len(idsBefore)).To(Equal(2))
				Eventually(cf.Cf("restart-app-instance", appName, "1")).Should(Exit(0))
				Eventually(func() []string {
					return differentIDsFrom(idsBefore)
				}, time.Second*120).Should(HaveLen(1))
			})
		})
	})
})
