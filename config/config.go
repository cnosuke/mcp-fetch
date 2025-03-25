package config

import (
	"github.com/jinzhu/configor"
)

// Config - Application configuration
type Config struct {
	Fetch struct {
		Timeout int `yaml:"timeout" default:"10" env:"FETCH_TIMEOUT"` // Timeout in seconds
		UserAgent string `yaml:"user_agent" default:"mcp-fetch/1.0" env:"FETCH_USER_AGENT"`
		MaxURLs int `yaml:"max_urls" default:"20" env:"FETCH_MAX_URLS"` // 一度に処理できるURLの最大数
		MaxWorkers int `yaml:"max_workers" default:"20" env:"FETCH_MAX_WORKERS"` // 並列処理に使用するワーカー数
	} `yaml:"fetch"`
}

// LoadConfig - Load configuration file
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	err := configor.New(&configor.Config{
		Debug:      false,
		Verbose:    false,
		Silent:     true,
		AutoReload: false,
	}).Load(cfg, path)
	return cfg, err
}
