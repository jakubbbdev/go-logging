package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the logger configuration
type Config struct {
	Level           string            `yaml:"level" json:"level"`
	Format          string            `yaml:"format" json:"format"`
	Output          string            `yaml:"output" json:"output"`
	IncludeCaller   bool              `yaml:"include_caller" json:"include_caller"`
	IncludeStack    bool              `yaml:"include_stack" json:"include_stack"`
	TimestampFormat string            `yaml:"timestamp_format" json:"timestamp_format"`
	UseColors       bool              `yaml:"use_colors" json:"use_colors"`
	DefaultFields   map[string]string `yaml:"default_fields" json:"default_fields"`

	// File handler specific
	FileConfig FileConfig `yaml:"file" json:"file"`

	// HTTP handler specific
	HTTPConfig HTTPConfig `yaml:"http" json:"http"`

	// Async handler specific
	AsyncConfig AsyncConfig `yaml:"async" json:"async"`

	// Metrics configuration
	MetricsConfig MetricsConfig `yaml:"metrics" json:"metrics"`
}

// FileConfig represents file handler configuration
type FileConfig struct {
	Path     string `yaml:"path" json:"path"`
	MaxSize  int64  `yaml:"max_size" json:"max_size"`
	MaxFiles int    `yaml:"max_files" json:"max_files"`
	Rotate   bool   `yaml:"rotate" json:"rotate"`
}

// HTTPConfig represents HTTP handler configuration
type HTTPConfig struct {
	URL     string            `yaml:"url" json:"url"`
	Headers map[string]string `yaml:"headers" json:"headers"`
	Timeout int               `yaml:"timeout" json:"timeout"`
}

// AsyncConfig represents async handler configuration
type AsyncConfig struct {
	BufferSize int `yaml:"buffer_size" json:"buffer_size"`
	Workers    int `yaml:"workers" json:"workers"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`
	Port      int    `yaml:"port" json:"port"`
	Path      string `yaml:"path" json:"path"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

// LoadConfigFromFile loads configuration from a YAML or JSON file
func LoadConfigFromFile(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}

	// Determine file type by extension
	if strings.HasSuffix(filepath, ".yaml") || strings.HasSuffix(filepath, ".yml") {
		err = yaml.Unmarshal(data, config)
	} else if strings.HasSuffix(filepath, ".json") {
		err = json.Unmarshal(data, config)
	} else {
		return nil, fmt.Errorf("unsupported config file format: %s", filepath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables
	config.overrideFromEnv()

	return config, nil
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() *Config {
	config := &Config{
		Level:           getEnv("LOG_LEVEL", "info"),
		Format:          getEnv("LOG_FORMAT", "text"),
		Output:          getEnv("LOG_OUTPUT", "console"),
		IncludeCaller:   getEnvBool("LOG_INCLUDE_CALLER", false),
		IncludeStack:    getEnvBool("LOG_INCLUDE_STACK", false),
		TimestampFormat: getEnv("LOG_TIMESTAMP_FORMAT", "2006-01-02 15:04:05"),
		UseColors:       getEnvBool("LOG_USE_COLORS", true),
		DefaultFields:   parseEnvFields("LOG_DEFAULT_FIELDS"),

		FileConfig: FileConfig{
			Path:     getEnv("LOG_FILE_PATH", "app.log"),
			MaxSize:  getEnvInt64("LOG_FILE_MAX_SIZE", 10*1024*1024),
			MaxFiles: getEnvInt("LOG_FILE_MAX_FILES", 5),
			Rotate:   getEnvBool("LOG_FILE_ROTATE", false),
		},

		HTTPConfig: HTTPConfig{
			URL:     getEnv("LOG_HTTP_URL", ""),
			Timeout: getEnvInt("LOG_HTTP_TIMEOUT", 30),
		},

		AsyncConfig: AsyncConfig{
			BufferSize: getEnvInt("LOG_ASYNC_BUFFER_SIZE", 1000),
			Workers:    getEnvInt("LOG_ASYNC_WORKERS", 4),
		},

		MetricsConfig: MetricsConfig{
			Enabled:   getEnvBool("LOG_METRICS_ENABLED", false),
			Port:      getEnvInt("LOG_METRICS_PORT", 8080),
			Path:      getEnv("LOG_METRICS_PATH", "/metrics"),
			Namespace: getEnv("LOG_METRICS_NAMESPACE", "logging"),
		},
	}

	return config
}

// overrideFromEnv overrides config values with environment variables
func (c *Config) overrideFromEnv() {
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		c.Level = val
	}
	if val := os.Getenv("LOG_FORMAT"); val != "" {
		c.Format = val
	}
	if val := os.Getenv("LOG_OUTPUT"); val != "" {
		c.Output = val
	}
	if val := os.Getenv("LOG_INCLUDE_CALLER"); val != "" {
		c.IncludeCaller = parseBool(val)
	}
	if val := os.Getenv("LOG_INCLUDE_STACK"); val != "" {
		c.IncludeStack = parseBool(val)
	}
}

// ToLogger creates a logger from the configuration
func (c *Config) ToLogger() (Logger, error) {
	// Parse level
	level, exists := ParseLevel(c.Level)
	if !exists {
		return nil, fmt.Errorf("invalid log level: %s", c.Level)
	}

	// Create formatter
	var formatter Formatter
	switch c.Format {
	case "json":
		formatter = NewJSONFormatter()
	case "text":
		formatter = NewTextFormatter()
	default:
		return nil, fmt.Errorf("invalid format: %s", c.Format)
	}

	// Create handler
	var handler Handler
	switch c.Output {
	case "console":
		handler = NewConsoleHandler()
	case "file":
		if c.FileConfig.Rotate {
			var err error
			handler, err = NewRotatingFileHandler(c.FileConfig.Path, c.FileConfig.MaxSize, c.FileConfig.MaxFiles)
			if err != nil {
				return nil, fmt.Errorf("failed to create rotating file handler: %w", err)
			}
		} else {
			var err error
			handler, err = NewFileHandler(c.FileConfig.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to create file handler: %w", err)
			}
		}
	case "http":
		if c.HTTPConfig.URL == "" {
			return nil, fmt.Errorf("HTTP URL is required for HTTP output")
		}
		handler = NewHTTPHandler(c.HTTPConfig.URL)
	default:
		return nil, fmt.Errorf("invalid output: %s", c.Output)
	}

	// Wrap with async if configured
	if c.AsyncConfig.BufferSize > 0 {
		handler = NewAsyncHandler(handler, c.AsyncConfig.BufferSize, c.AsyncConfig.Workers)
	}

	// Convert default fields
	fields := make(Fields)
	for k, v := range c.DefaultFields {
		fields[k] = v
	}

	// Create logger with options
	logger := NewLogger(
		WithLevel(level),
		WithHandler(handler),
		WithFormatter(formatter),
		WithCaller(c.IncludeCaller),
		WithStacktrace(c.IncludeStack),
		WithDefaultFields(fields),
	)

	return logger, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return parseBool(value)
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

func parseBool(value string) bool {
	switch strings.ToLower(value) {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}

func parseEnvFields(key string) map[string]string {
	fields := make(map[string]string)
	if value := os.Getenv(key); value != "" {
		// Format: "key1=value1,key2=value2"
		pairs := strings.Split(value, ",")
		for _, pair := range pairs {
			if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
				fields[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}
	return fields
}
