package config

import (
	"github.com/jinzhu/configor"
)

// Config - Application configuration
type Config struct {
	Log   string `yaml:"log" default:"" env:"LOG_PATH"`
	Debug bool   `yaml:"debug" default:"false" env:"DEBUG"` // Log file path
	Fetch struct {
		Timeout          int    `yaml:"timeout" default:"10" env:"FETCH_TIMEOUT"` // Timeout in seconds
		UserAgent        string `yaml:"user_agent" default:"mcp-fetch/1.0" env:"FETCH_USER_AGENT"`
		MaxURLs          int    `yaml:"max_urls" default:"20" env:"FETCH_MAX_URLS"`                       // Maximum number of URLs that can be processed at once
		MaxWorkers       int    `yaml:"max_workers" default:"20" env:"FETCH_MAX_WORKERS"`                 // Number of workers used for parallel processing
		DefaultMaxLength int    `yaml:"default_max_length" default:"5000" env:"FETCH_DEFAULT_MAX_LENGTH"` // Default maximum character count for returned content
	} `yaml:"fetch"`
	HTTP struct {
		Binding          string   `yaml:"binding" default:"localhost:8080" env:"HTTP_BINDING"`
		EndpointPath     string   `yaml:"endpoint_path" default:"/mcp" env:"HTTP_ENDPOINT_PATH"`
		HeartbeatSeconds int      `yaml:"heartbeat_seconds" default:"30" env:"HTTP_HEARTBEAT_SECONDS"`
		AuthToken        string   `yaml:"auth_token" default:"" env:"HTTP_AUTH_TOKEN"`
		AllowedOrigins   []string `yaml:"allowed_origins" env:"HTTP_ALLOWED_ORIGINS"`
	} `yaml:"http"`
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
