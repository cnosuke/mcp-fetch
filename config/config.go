package config

import (
	"github.com/jinzhu/configor"
)

// Config - Application configuration
type Config struct {
	Greeting struct {
		DefaultMessage string `yaml:"default_message" default:"Hello!" env:"GREETING_DEFAULT_MESSAGE"`
	} `yaml:"greeting"`
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
