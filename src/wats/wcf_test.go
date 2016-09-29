package wats

import (
	"crypto/tls"
	"encoding/xml"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const soapBody = `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
	<s:Body>
		<Echo xmlns="http://tempuri.org/">
			<msg>test</msg>
		</Echo>
	</s:Body>
</s:Envelope>
`

var _ = Describe("WCF", func() {
	Describe("A WCF application", func() {
		It("can have multiple routable instances on the same cell", func() {
			numWinCells, err := strconv.Atoi(os.Getenv("NUM_WIN_CELLS"))
			Expect(err).NotTo(HaveOccurred(), "Please provide NUM_WIN_CELLS (The number of windows cells in tested deployment)")

			By("pushing multiple instances of it", func() {
				Eventually(pushApp(appName, "../../assets/wcf/Hello.Service.IIS", numWinCells+1, "256m"), CF_PUSH_TIMEOUT).Should(Succeed())
			})

			enableDiego(appName)
			Eventually(runCf("start", appName), CF_PUSH_TIMEOUT).Should(Succeed())

			By("verifying it's up")
			type WCFResponse struct {
				Msg          string
				InstanceGuid string
				CFInstanceIp string
			}

			wcfRequest := func(appName string) WCFResponse {
				uri := helpers.AppUri(appName, "/Hello.svc?wsdl")

				helloMsg := `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><Echo xmlns="http://tempuri.org/"><msg>WATS!!!</msg></Echo></s:Body></s:Envelope>`
				buf := strings.NewReader(helloMsg)
				req, err := http.NewRequest("POST", uri, buf)
				req.Header.Add("Content-Type", "text/xml")
				req.Header.Add("SOAPAction", "http://tempuri.org/IHelloService/Echo")
				client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				xmlDecoder := xml.NewDecoder(resp.Body)
				type SoapResponse struct {
					XMLResult string `xml:"Body>EchoResponse>EchoResult"`
				}
				xmlResponse := SoapResponse{}
				Expect(xmlDecoder.Decode(&xmlResponse)).To(BeNil())
				results := strings.Split(xmlResponse.XMLResult, ",")
				Expect(len(results)).To(Equal(3))
				return WCFResponse{
					Msg:          results[0],
					CFInstanceIp: results[1],
					InstanceGuid: results[2],
				}
			}

			Eventually(wcfRequest(appName).Msg).Should(Equal("WATS!!!"))
			isServiceRunningOnTheSameCell := func(appName string) bool {
				// Keep track of the IDs of the instances we have reached
				output := map[string]string{}
				for i := 0; i < numWinCells*5; i++ {
					res := wcfRequest(appName)
					guids := output[res.CFInstanceIp]
					if guids != "" && !strings.Contains(guids, res.InstanceGuid) {
						return true
					}
					output[res.CFInstanceIp] = res.InstanceGuid
				}
				return false
			}

			Expect(isServiceRunningOnTheSameCell(appName)).To(BeTrue())
		})
	})
})
