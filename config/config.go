package config

import (
	"os"
	"strconv"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config - Application configuration
type Config struct {
	Log   string `koanf:"log"`
	Debug bool   `koanf:"debug"`
	Fetch struct {
		Timeout          int    `koanf:"timeout"`
		UserAgent        string `koanf:"user_agent"`
		MaxURLs          int    `koanf:"max_urls"`
		MaxWorkers       int    `koanf:"max_workers"`
		DefaultMaxLength int    `koanf:"default_max_length"`
	} `koanf:"fetch"`
}

func defaultValues() map[string]any {
	return map[string]any{
		"log":                      "",
		"debug":                    false,
		"fetch.timeout":            10,
		"fetch.user_agent":         "mcp-fetch/1.0",
		"fetch.max_urls":           20,
		"fetch.max_workers":        20,
		"fetch.default_max_length": 5000,
	}
}

func loadEnvOverrides() (map[string]any, error) {
	overrides := map[string]any{}

	if v := os.Getenv("LOG_PATH"); v != "" {
		overrides["log"] = v
	}
	if v := os.Getenv("DEBUG"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
		overrides["debug"] = b
	}
	if v := os.Getenv("FETCH_TIMEOUT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		overrides["fetch.timeout"] = n
	}
	if v := os.Getenv("FETCH_USER_AGENT"); v != "" {
		overrides["fetch.user_agent"] = v
	}
	if v := os.Getenv("FETCH_MAX_URLS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		overrides["fetch.max_urls"] = n
	}
	if v := os.Getenv("FETCH_MAX_WORKERS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		overrides["fetch.max_workers"] = n
	}
	if v := os.Getenv("FETCH_DEFAULT_MAX_LENGTH"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		overrides["fetch.default_max_length"] = n
	}

	return overrides, nil
}

// LoadConfig - Load configuration file
func LoadConfig(path string) (*Config, error) {
	k := koanf.New(".")

	if err := k.Load(confmap.Provider(defaultValues(), "."), nil); err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); err == nil {
		if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
			return nil, err
		}
	}

	envOverrides, err := loadEnvOverrides()
	if err != nil {
		return nil, err
	}
	if err := k.Load(confmap.Provider(envOverrides, "."), nil); err != nil {
		return nil, err
	}

	cfg := &Config{}
	return cfg, k.Unmarshal("", cfg)
}
