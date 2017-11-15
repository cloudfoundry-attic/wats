package wats

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("File ACLs", func() {

	var inaccessibleFiles = []string{
		"C:\\bosh",
		"C:\\containerizer",
		"C:\\var",
		"C:\\Windows\\Panther\\Unattend",
	}

	var client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	BeforeEach(func() {
		if config.GetStack() == "windows2016" {
			Skip("n/a on windows2016")
		}

		pushAndStartNora(appName)
		Eventually(helpers.CurlingAppRoot(config, appName)).Should(ContainSubstring("hello i am nora"))
	})

	permission := func(path string) (string, error) {
		uri := helpers.AppUri(appName, "/inaccessible_file", config)
		res, err := client.Post(uri, "text/plain", strings.NewReader(path))
		if err != nil {
			return "", err
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		return string(body), err
	}

	It("A Container user should not have permission to view sensitive files", func() {
		for _, path := range inaccessibleFiles {
			response, err := permission(path)
			Expect(err).To(Succeed(), path)

			response, err = strconv.Unquote(response)
			Expect(err).To(Succeed())
			Expect(response).To(Or(Equal("ACCESS_DENIED"), Equal("NOT_EXIST")), path)
		}
	})
})
