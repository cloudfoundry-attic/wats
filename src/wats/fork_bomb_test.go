package wats

import (
	"io"
	"os"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe("Application Lifecycle", func() {
	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

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
			seenComputerNames[helpers.CurlApp(appName, "/ENV/CF_INSTANCE_IP")] = true
		}

		return seenComputerNames
	}

	Describe("An app staged on Diego and running on Diego", func() {
		It("attempts to forkbomb the environment", func() {
			numWinCells, err := strconv.Atoi(os.Getenv("NUM_WIN_CELLS"))
			Expect(err).NotTo(HaveOccurred(), "Please provide NUM_WIN_CELLS (The number of windows cells in tested deployment)")

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
				Eventually(pushNoraWithOptions(appName, numWinCells*2, "2G"), CF_PUSH_TIMEOUT).Should(Succeed())
			})

			By("staging and running it on Diego", func() {
				enableDiego(appName)
				Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())
			})

			By("verifying it's up", func() {
				Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))
			})

			By("storing the current computer names")
			computerNames := reportedComputerNames(numWinCells)
			Expect(len(computerNames)).To(Equal(numWinCells))

			By("Running fork bomb", func() {
				helpers.CurlApp(appName, "/run", "-f", "-X", "POST", "-d", "bin/breakoutbomb.exe")
			})

			time.Sleep(3 * time.Second)

			By("Making sure the bomb did not take down the machine", func() {
				newComputerNames := reportedComputerNames(numWinCells)
				Expect(newComputerNames).To(Equal(computerNames))
			})
		})
	})
})
