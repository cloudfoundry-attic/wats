package wats

import (
	"errors"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
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
