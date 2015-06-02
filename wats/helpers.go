package wats

import (
	"errors"
	"os"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
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

func runCf(values ...string) func() error {
	return func() error {
		session := cf.Cf(values...)
		session.Wait()
		if session.ExitCode() == 0 {
			return nil
		}

		return errors.New("non zero exit code")
	}
}

func DopplerUrl(c Config) string {
	doppler := os.Getenv("DOPPLER_URL")
	if doppler == "" {
		doppler = "wss://doppler." + c.AppsDomain
	}
	return doppler
}
