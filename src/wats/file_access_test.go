package wats

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("File ACLs", func() {

	var inaccessibleFiles = []string{
		"C:\\bosh",
		"C:\\containerizer",
		"C:\\var",
		"C:\\Windows\\System32\\Sysprep\\Panther\\IE\\setupact.log",
		"C:\\Windows\\System32\\Sysprep\\Panther\\IE\\setuperr.log",
		"C:\\Windows\\System32\\Sysprep\\Panther\\setupact.log",
		"C:\\Windows\\System32\\Sysprep\\Panther\\setuperr.log",
	}

	BeforeEach(func() {
		pushAndStartNora(appName)
		Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("hello i am nora"))

		// Windows file paths are case-insensitive
		for i, s := range inaccessibleFiles {
			inaccessibleFiles[i] = strings.ToLower(s)
		}
	})

	It("A Container user should not have permission to view sensitive files", func() {

		// The 'inaccessible_files' endpoint walks the
		// entire filesystem, which may take a while.
		response := helpers.CurlAppWithTimeout(appName, "/inaccessible_files", 5*time.Minute)
		var files []string
		Expect(json.Unmarshal([]byte(strings.ToLower(response)), &files)).To(Succeed())
		for _, inaccessibleFile := range inaccessibleFiles {
			Expect(files).To(ContainElement(inaccessibleFile))
		}
	})
})
