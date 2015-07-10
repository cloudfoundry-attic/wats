package wats

import (
	"errors"
	"os"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/onsi/gomega/gbytes"
)

func pushNora(appName string) func() error {
	return runCf(
		"push", appName,
		"-p", "../assets/nora/NoraPublished",
		"--no-start",
		"-m", "2g",
		"-b", "https://github.com/ryandotsmith/null-buildpack.git",
		"-s", "windows2012R2")
}

func runCfWithOutput(values ...string) (*gbytes.Buffer, error) {
	session := cf.Cf(values...)
	session.Wait(CF_PUSH_TIMEOUT)
	if session.ExitCode() == 0 {
		return session.Out, nil
	}

	return nil, errors.New("non zero exit code")
}

func runCf(values ...string) func() error {
	return func() error {
		_, err := runCfWithOutput(values...)
		return err
	}
}

func DopplerUrl(c Config) string {
	doppler := os.Getenv("DOPPLER_URL")
	if doppler == "" {
		doppler = "wss://doppler." + c.AppsDomain + ":4443"
	}
	return doppler
}
