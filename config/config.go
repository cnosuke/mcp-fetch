package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config - Application configuration
type Config struct {
	Log   string `yaml:"log"`
	Debug bool   `yaml:"debug"`
	Fetch struct {
		Timeout          int    `yaml:"timeout"`
		UserAgent        string `yaml:"user_agent"`
		MaxURLs          int    `yaml:"max_urls"`
		MaxWorkers       int    `yaml:"max_workers"`
		DefaultMaxLength int    `yaml:"default_max_length"`
	} `yaml:"fetch"`
}

// LoadConfig - Load configuration file
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	cfg.Fetch.Timeout = 10
	cfg.Fetch.UserAgent = "mcp-fetch/1.0"
	cfg.Fetch.MaxURLs = 20
	cfg.Fetch.MaxWorkers = 20
	cfg.Fetch.DefaultMaxLength = 5000

	f, err := os.Open(path)
	if err == nil {
		defer f.Close()
		if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
			return nil, err
		}
	}

	if v := os.Getenv("LOG_PATH"); v != "" {
		cfg.Log = v
	}
	if v := os.Getenv("DEBUG"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Debug = b
		}
	}
	if v := os.Getenv("FETCH_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Fetch.Timeout = n
		}
	}
	if v := os.Getenv("FETCH_USER_AGENT"); v != "" {
		cfg.Fetch.UserAgent = v
	}
	if v := os.Getenv("FETCH_MAX_URLS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Fetch.MaxURLs = n
		}
	}
	if v := os.Getenv("FETCH_MAX_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Fetch.MaxWorkers = n
		}
	}
	if v := os.Getenv("FETCH_DEFAULT_MAX_LENGTH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Fetch.DefaultMaxLength = n
		}
	}

	return cfg, nil
}
