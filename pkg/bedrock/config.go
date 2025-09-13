// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

package bedrock

import (
	"errors"
	"fmt"

	"github.com/genkit-aws/genkit-aws/internal/constants"
)

// Config holds configuration for Bedrock integration
type Config struct {
	// Models is a list of model IDs to make available
	Models []string `json:"models,omitempty"`

	// ModelConfigs allows per-model configuration overrides
	ModelConfigs map[string]*ModelConfig `json:"model_configs,omitempty"`

	// DefaultModelConfig provides default settings for all models
	DefaultModelConfig *ModelConfig `json:"default_model_config,omitempty"`
}

// ModelConfig holds configuration for a specific model
type ModelConfig struct {
	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature controls randomness in generation (0.0 to 1.0)
	Temperature float64 `json:"temperature,omitempty"`

	// TopP controls nucleus sampling
	TopP float64 `json:"top_p,omitempty"`

	// StopSequences are sequences that will stop generation
	StopSequences []string `json:"stop_sequences,omitempty"`
}

// Validate validates the Bedrock configuration
func (c *Config) Validate() error {
	if len(c.Models) == 0 {
		return errors.New("at least one model must be specified")
	}

	for _, modelID := range c.Models {
		if modelID == "" {
			return errors.New("model ID cannot be empty")
		}
	}

	// Validate model configs
	for modelID, config := range c.ModelConfigs {
		if err := config.Validate(); err != nil {
			return fmt.Errorf("invalid config for model %s: %w", modelID, err)
		}
	}

	if c.DefaultModelConfig != nil {
		if err := c.DefaultModelConfig.Validate(); err != nil {
			return fmt.Errorf("invalid default model config: %w", err)
		}
	}

	return nil
}

// Validate validates a model configuration
func (mc *ModelConfig) Validate() error {
	if mc.MaxTokens < 0 {
		return errors.New("max_tokens must be non-negative")
	}

	if mc.Temperature < 0.0 || mc.Temperature > 1.0 {
		return errors.New("temperature must be between 0.0 and 1.0")
	}

	if mc.TopP < 0.0 || mc.TopP > 1.0 {
		return errors.New("top_p must be between 0.0 and 1.0")
	}

	return nil
}

// ModelConfig returns the configuration for a specific model ID
func (c *Config) ModelConfig(modelID string) *ModelConfig {
	// Check for model-specific config first
	if config, exists := c.ModelConfigs[modelID]; exists {
		return c.mergeWithDefaults(config)
	}

	// Use default config
	if c.DefaultModelConfig != nil {
		return c.DefaultModelConfig
	}

	// Return sensible defaults
	return &ModelConfig{
		MaxTokens:   constants.DefaultMaxTokens,
		Temperature: constants.DefaultTemperature,
		TopP:        constants.DefaultTopP,
	}
}

// mergeWithDefaults merges a model config with defaults
func (c *Config) mergeWithDefaults(config *ModelConfig) *ModelConfig {
	if c.DefaultModelConfig == nil {
		return config
	}

	merged := *config

	if merged.MaxTokens == 0 {
		merged.MaxTokens = c.DefaultModelConfig.MaxTokens
	}
	if merged.Temperature == 0 {
		merged.Temperature = c.DefaultModelConfig.Temperature
	}
	if merged.TopP == 0 {
		merged.TopP = c.DefaultModelConfig.TopP
	}
	if len(merged.StopSequences) == 0 {
		merged.StopSequences = c.DefaultModelConfig.StopSequences
	}

	return &merged
}
