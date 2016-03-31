package wats

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getOauthToken() string {
	session := cf.Cf("oauth-token")
	session.Wait()
	out := string(session.Out.Contents())
	authToken := strings.Split(out, "\n")[3]
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
	It("garden-windows emits metrics to the firehose", func() {
		config := helpers.LoadConfig()

		duration, _ := time.ParseDuration("5s")
		cf.AsUser(context.AdminUserContext(), duration, func() {
			authToken := getOauthToken()
			msgChan, errorChan := createNoaaClient(DopplerUrl(config), authToken)

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
