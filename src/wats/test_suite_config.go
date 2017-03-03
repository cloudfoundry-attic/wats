package wats

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type watsConfig struct {
	ApiEndpoint          string `json:"api"`
	AdminUser            string `json:"admin_user"`
	AdminPassword        string `json:"admin_password"`
	SecureAddress        string `json:"secure_address"`
	AppsDomain           string `json:"apps_domain"`
	SkipSSLValidation    bool   `json:"skip_ssl_validation"`
	NumWindowsCells      int    `json:"num_windows_cells"`
	ArtifactsDirectory   string `json:"artifacts_directory"`
	UseHttp              bool   `json:"use_http"`
	ConsulMutualTls      bool   `json:"consul_mutual_tls"`
	HttpHealthcheck      bool   `json:"http_healthcheck"`
	IsolationSegmentName string `json:"isolation_segment_name"`
}

func LoadWatsConfig() (*watsConfig, error) {
	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		return &watsConfig{}, errors.New("Must set CONFIG to point to an integration config JSON file")
	}

	return LoadWatsConfigFromPath(configPath)
}

func LoadWatsConfigFromPath(configPath string) (*watsConfig, error) {
	configContents, err := ioutil.ReadFile(configPath)
	if err != nil {
		return &watsConfig{}, err
	}

	config := watsConfig{
		ArtifactsDirectory: filepath.Join("..", "results"),
		UseHttp:            true,
	}
	err = json.Unmarshal(configContents, &config)
	if err != nil {
		return &watsConfig{}, err
	}

	return &config, nil
}

func (w *watsConfig) GetApiEndpoint() string {
	return w.ApiEndpoint
}

func (w *watsConfig) GetConfigurableTestPassword() string {
	return ""
}

func (w *watsConfig) GetPersistentAppOrg() string {
	return ""
}

func (w *watsConfig) GetPersistentAppQuotaName() string {
	return ""
}

func (w *watsConfig) GetPersistentAppSpace() string {
	return ""
}

func (w *watsConfig) GetScaledTimeout(timeout time.Duration) time.Duration {
	return timeout
}

func (w *watsConfig) GetAdminPassword() string {
	return w.AdminPassword
}

func (w *watsConfig) GetExistingUser() string {
	return ""
}

func (w *watsConfig) GetExistingUserPassword() string {
	return ""
}

func (w *watsConfig) GetShouldKeepUser() bool {
	return false
}

func (w *watsConfig) GetUseExistingUser() bool {
	return false
}

func (w *watsConfig) GetAdminUser() string {
	return w.AdminUser
}

func (w *watsConfig) GetSkipSSLValidation() bool {
	return w.SkipSSLValidation
}

func (w *watsConfig) GetNamePrefix() string {
	return "WATS"
}

func (w *watsConfig) GetAppsDomain() string {
	return w.AppsDomain
}

func (w *watsConfig) GetNumWindowsCells() int {
	return w.NumWindowsCells
}

func (w *watsConfig) GetSecureAddress() string {
	return w.SecureAddress
}

func (w *watsConfig) GetArtifactsDirectory() string {
	return w.ArtifactsDirectory
}

func (w *watsConfig) Protocol() string {
	if w.UseHttp {
		return "http://"
	} else {
		return "https://"
	}
}

func (w *watsConfig) GetIsolationSegmentName() string {
	return w.IsolationSegmentName
}
