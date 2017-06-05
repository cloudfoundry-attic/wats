package wats

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Http Healthcheck", func() {
	BeforeEach(func() {
		if !config.HttpHealthcheck {
			Skip("Skipping Http Healthcheck tests (requires capi-release v1.12.0 and above)")
		}
	})
	testHealthCheck := func(healthCheckType, healthCheckEndpoint string) {
		healthcheck := cf.Cf("curl", fmt.Sprintf("/v2/apps?q=name:%s", appName))
		Expect(healthcheck.Wait()).To(gexec.Exit(0))
		type HealthCheck struct {
			Resources []struct {
				Entity struct {
					HealthCheckEndpoint string `json:"health_check_http_endpoint"`
					HealthCheckType     string `json:"health_check_type"`
				}
			}
		}
		var h HealthCheck
		Expect(json.Unmarshal(healthcheck.Out.Contents(), &h)).To(Succeed())
		Expect(h.Resources).ToNot(BeEmpty())
		Expect(h.Resources[0].Entity.HealthCheckType).To(Equal(healthCheckType))
		Expect(h.Resources[0].Entity.HealthCheckEndpoint).To(Equal(healthCheckEndpoint))
	}

	cfLogs := func(appName string) func() *Buffer {
		return func() *Buffer {
			out, _ := runCfWithOutput("logs", appName, "--recent")
			return out
		}
	}

	Describe("An app staged on Diego and running on Diego", func() {
		It("should not start with an invalid healthcheck endpoint", func() {
			By("pushing it")
			Expect(pushNoraWithOptions(appName, 1, "2g").Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

			By("setting an invalid healthcheck endpoint")
			cf.Cf("set-health-check", appName, "http", "--endpoint", "/invalidhealthcheck")

			By("staging and running it on Diego")
			enableDiego(appName)
			start := cf.Cf("start", appName)
			defer start.Kill()
			Eventually(cfLogs(appName), CF_PUSH_TIMEOUT).Should(Say("health check never passed."))
		})

		It("starts with a valid http healthcheck endpoint", func() {
			By("pushing it")
			Expect(pushNoraWithOptions(appName, 1, "2g").Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

			By("setting the healthcheck endpoint")
			cf.Cf("set-health-check", appName, "http", "--endpoint", "/healthcheck")

			By("staging and running it on Diego")
			enableDiego(appName)
			Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

			By("ensuring the healthcheck endpoint is set")
			testHealthCheck("http", "/healthcheck")
		})

		It("starts with a http healthcheck endpoint that is a redirect", func() {
			By("pushing it")
			Expect(pushNoraWithOptions(appName, 1, "2g").Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

			By("setting the healthcheck endpoint")
			cf.Cf("set-health-check", appName, "http", "--endpoint", "/redirect/healthcheck")

			By("staging and running it on Diego")
			enableDiego(appName)
			Expect(cf.Cf("start", appName).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

			By("ensuring the healthcheck endpoint is set")
			testHealthCheck("http", "/redirect/healthcheck")
		})

		It("does not start with a http healthcheck endpoint that is an invalid redirect", func() {
			By("pushing it")
			Expect(pushNoraWithOptions(appName, 1, "2g").Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

			By("setting the healthcheck endpoint")
			cf.Cf("set-health-check", appName, "http", "--endpoint", "/redirect/invalidhealthcheck")

			By("staging and running it on Diego")
			enableDiego(appName)
			start := cf.Cf("start", appName)
			defer start.Kill()
			Eventually(cfLogs(appName), CF_PUSH_TIMEOUT).Should(Say("health check never passed."))
		})
	})
})
