package wats

import (
	"fmt"
	"io"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe("Application Lifecycle", func() {
	reportedComputerNames := func(instances int) map[string]bool {
		timer := time.NewTimer(time.Second * 120)
		defer timer.Stop()
		run := true
		go func() {
			<-timer.C
			run = false
		}()

		seenComputerNames := map[string]bool{}
		for len(seenComputerNames) != instances && run == true {
			seenComputerNames[helpers.CurlApp(config, appName, "/ENV/CF_INSTANCE_IP")] = true
		}

		return seenComputerNames
	}

	BeforeEach(func() {
		memLimit := config.GetNumWindowsCells() * 2 * 4
		if memLimit < 10 {
			memLimit = 10
		}
		setTotalMemoryLimit(fmt.Sprintf("%dG", memLimit))
	})

	AfterEach(func() {
		setTotalMemoryLimit("10G")
	})

	Describe("An app staged on Diego and running on Diego", func() {
		It("attempts to forkbomb the environment", func() {
			src, err := os.Open("../../assets/greenhouse-security-fixtures/bin/BreakoutBomb.exe")
			Expect(err).NotTo(HaveOccurred())
			defer src.Close()
			dst, err := os.Create("../../assets/nora/NoraPublished/bin/breakoutbomb.exe")
			Expect(err).NotTo(HaveOccurred())
			defer dst.Close()
			_, err = io.Copy(dst, src)
			Expect(err).NotTo(HaveOccurred())
			dst.Close()

			By("pushing it", func() {
				Expect(pushNoraWithOptions(appName, config.GetNumWindowsCells()*2, "2G").Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			})

			By("staging and running it on Diego", func() {
				enableDiego(appName)
				Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
			})

			By("verifying it's up", func() {
				Eventually(appRunning(appName, config.GetNumWindowsCells()*2, CF_PUSH_TIMEOUT), CF_PUSH_TIMEOUT).Should(Succeed())
				Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))
			})

			By("storing the current computer names")
			computerNames := reportedComputerNames(config.GetNumWindowsCells())
			Expect(len(computerNames)).To(Equal(config.GetNumWindowsCells()))

			By("Running fork bomb", func() {
				helpers.CurlApp(config, appName, "/run", "-f", "-X", "POST", "-d", "bin/breakoutbomb.exe")
			})

			time.Sleep(3 * time.Second)

			By("Making sure the bomb did not take down the machine", func() {
				newComputerNames := reportedComputerNames(config.GetNumWindowsCells())
				Expect(newComputerNames).To(Equal(computerNames))
			})
		})
	})
})
