package wats

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/sonde-go/events"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getOauthToken() string {
	session := cf.Cf("oauth-token")
	session.Wait()
	out := string(session.Out.Contents())
	authToken := strings.Split(out, "\n")[0]
	Expect(authToken).To(HavePrefix("bearer"))
	return authToken
}

func createNoaaClient(dopplerUrl, authToken string) (chan *events.Envelope, chan error) {
	connection := noaa.NewConsumer(dopplerUrl, &tls.Config{InsecureSkipVerify: true}, nil)
	msgChan := make(chan *events.Envelope)
	errorChan := make(chan error)

	go func() {
		defer close(msgChan)
		go connection.Firehose("firehose-a", authToken, msgChan, errorChan)

		for err := range errorChan {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		}
	}()

	return msgChan, errorChan
}

var _ = Describe("Metrics", func() {
	FIt("garden-windows emits metrics to the firehose", func() {
		duration, _ := time.ParseDuration("5s")
		AsUser(environment.AdminUserContext(), duration, func() {
			authToken := getOauthToken()
			msgChan, errorChan := createNoaaClient(DopplerUrl(), authToken)

			Consistently(errorChan).ShouldNot(Receive())

			sipTheStream := func() string {
				if envelope, ok := <-msgChan; ok {
					return *envelope.Origin
				}
				return ""
			}
			Eventually(sipTheStream, "1m", "5ms").Should(Equal("garden-windows"))
		})
	})
})
