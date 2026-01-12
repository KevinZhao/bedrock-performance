package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the complete configuration for the benchmark tool
type Config struct {
	AWS         AWSConfig         `json:"aws"`
	Model       ModelConfig       `json:"model"`
	Test        TestConfig        `json:"test"`
	Concurrency ConcurrencyConfig `json:"concurrency"`
	Output      OutputConfig      `json:"output"`
}

// AWSConfig contains AWS credentials and region
type AWSConfig struct {
	Region          string `json:"region"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
}

// ModelConfig contains Bedrock model configuration
type ModelConfig struct {
	ID    string `json:"id"`
	Quota int    `json:"quota"`
}

// TestConfig contains test parameters
type TestConfig struct {
	PromptSize     int     `json:"prompt_size"`
	PromptTemplate string  `json:"prompt_template"`
	Streaming      bool    `json:"streaming"`
	NonStreaming   bool    `json:"non_streaming"`
	MaxTokens      int     `json:"max_tokens"`
	Temperature    float64 `json:"temperature"`
	ServiceTier    string  `json:"service_tier"`
}

// ConcurrencyConfig defines the concurrency test parameters
type ConcurrencyConfig struct {
	Start            int `json:"start"`
	End              int `json:"end"`
	Step             int `json:"step"`
	DurationSeconds  int `json:"duration_seconds"`
}

// OutputConfig defines output settings
type OutputConfig struct {
	ReportFile string `json:"report_file"`
}

// LoadConfig reads and parses the configuration file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.AWS.Region == "" {
		return fmt.Errorf("aws.region is required")
	}
	// access_key_id and secret_access_key are optional
	// if empty, the SDK will use default credential chain
	if c.Model.ID == "" {
		return fmt.Errorf("model.id is required")
	}
	if c.Model.Quota <= 0 {
		return fmt.Errorf("model.quota must be positive")
	}
	if c.Test.PromptSize <= 0 {
		return fmt.Errorf("test.prompt_size must be positive")
	}
	if !c.Test.Streaming && !c.Test.NonStreaming {
		return fmt.Errorf("at least one of streaming or non_streaming must be enabled")
	}
	if c.Test.MaxTokens <= 0 {
		return fmt.Errorf("test.max_tokens must be positive")
	}
	if c.Concurrency.Start <= 0 {
		return fmt.Errorf("concurrency.start must be positive")
	}
	if c.Concurrency.End < c.Concurrency.Start {
		return fmt.Errorf("concurrency.end must be >= concurrency.start")
	}
	if c.Concurrency.Step <= 0 {
		return fmt.Errorf("concurrency.step must be positive")
	}
	if c.Concurrency.DurationSeconds <= 0 {
		return fmt.Errorf("concurrency.duration_seconds must be positive")
	}
	if c.Output.ReportFile == "" {
		return fmt.Errorf("output.report_file is required")
	}

	return nil
}
