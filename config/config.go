package config

import (
	"errors"
	"io/fs"
	"os"
	"strconv"
	"strings"

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
	HTTP struct {
		Binding          string   `koanf:"binding"`
		EndpointPath     string   `koanf:"endpoint_path"`
		HeartbeatSeconds int      `koanf:"heartbeat_seconds"`
		AuthToken        string   `koanf:"auth_token"`
		AllowedOrigins   []string `koanf:"allowed_origins"`
	} `koanf:"http"`
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
		"http.binding":             "localhost:8080",
		"http.endpoint_path":       "/mcp",
		"http.heartbeat_seconds":   30,
		"http.auth_token":          "",
		"http.allowed_origins":     []string{},
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
	if v := os.Getenv("HTTP_BINDING"); v != "" {
		overrides["http.binding"] = v
	}
	if v := os.Getenv("HTTP_ENDPOINT_PATH"); v != "" {
		overrides["http.endpoint_path"] = v
	}
	if v := os.Getenv("HTTP_HEARTBEAT_SECONDS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		overrides["http.heartbeat_seconds"] = n
	}
	if v := os.Getenv("HTTP_AUTH_TOKEN"); v != "" {
		overrides["http.auth_token"] = v
	}
	if v := os.Getenv("HTTP_ALLOWED_ORIGINS"); v != "" {
		overrides["http.allowed_origins"] = strings.Split(v, ",")
	}

	return overrides, nil
}

// LoadConfig - Load configuration file
func LoadConfig(path string) (*Config, error) {
	k := koanf.New(".")

	if err := k.Load(confmap.Provider(defaultValues(), "."), nil); err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	} else {
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
