package config

import (
	"github.com/jinzhu/configor"
)

// Config - Application configuration
type Config struct {
	Fetch struct {
		Timeout int `yaml:"timeout" default:"10" env:"FETCH_TIMEOUT"` // Timeout in seconds
		UserAgent string `yaml:"user_agent" default:"mcp-fetch/1.0" env:"FETCH_USER_AGENT"`
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
