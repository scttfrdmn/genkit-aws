// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

package bedrock

import (
	"encoding/json"
	"testing"

	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				Models: []string{"anthropic.claude-3-sonnet-20240229-v1:0"},
			},
			wantErr: false,
		},
		{
			name: "no models",
			config: &Config{
				Models: []string{},
			},
			wantErr: true,
			errMsg:  "at least one model must be specified",
		},
		{
			name: "empty model ID",
			config: &Config{
				Models: []string{""},
			},
			wantErr: true,
			errMsg:  "model ID cannot be empty",
		},
		{
			name: "valid config with model configs",
			config: &Config{
				Models: []string{"anthropic.claude-3-sonnet-20240229-v1:0"},
				ModelConfigs: map[string]*ModelConfig{
					"anthropic.claude-3-sonnet-20240229-v1:0": {
						MaxTokens:   4096,
						Temperature: 0.7,
						TopP:        0.9,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestModelConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *ModelConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &ModelConfig{
				MaxTokens:   4096,
				Temperature: 0.7,
				TopP:        0.9,
			},
			wantErr: false,
		},
		{
			name: "negative max tokens",
			config: &ModelConfig{
				MaxTokens: -1,
			},
			wantErr: true,
			errMsg:  "max_tokens must be non-negative",
		},
		{
			name: "temperature too low",
			config: &ModelConfig{
				Temperature: -0.1,
			},
			wantErr: true,
			errMsg:  "temperature must be between 0.0 and 1.0",
		},
		{
			name: "temperature too high",
			config: &ModelConfig{
				Temperature: 1.1,
			},
			wantErr: true,
			errMsg:  "temperature must be between 0.0 and 1.0",
		},
		{
			name: "top_p too low",
			config: &ModelConfig{
				TopP: -0.1,
			},
			wantErr: true,
			errMsg:  "top_p must be between 0.0 and 1.0",
		},
		{
			name: "top_p too high",
			config: &ModelConfig{
				TopP: 1.1,
			},
			wantErr: true,
			errMsg:  "top_p must be between 0.0 and 1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_ModelConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		modelID     string
		expectedCfg *ModelConfig
	}{
		{
			name: "specific model config exists",
			config: &Config{
				ModelConfigs: map[string]*ModelConfig{
					"test-model": {
						MaxTokens:   2048,
						Temperature: 0.5,
					},
				},
				DefaultModelConfig: &ModelConfig{
					MaxTokens:   4096,
					Temperature: 0.7,
				},
			},
			modelID: "test-model",
			expectedCfg: &ModelConfig{
				MaxTokens:   2048,
				Temperature: 0.5,
			},
		},
		{
			name: "use default config",
			config: &Config{
				DefaultModelConfig: &ModelConfig{
					MaxTokens:   4096,
					Temperature: 0.7,
					TopP:        0.9,
				},
			},
			modelID: "unknown-model",
			expectedCfg: &ModelConfig{
				MaxTokens:   4096,
				Temperature: 0.7,
				TopP:        0.9,
			},
		},
		{
			name:    "use built-in defaults",
			config:  &Config{},
			modelID: "unknown-model",
			expectedCfg: &ModelConfig{
				MaxTokens:   4096,
				Temperature: 0.7,
				TopP:        0.9,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ModelConfig(tt.modelID)

			assert.Equal(t, tt.expectedCfg.MaxTokens, result.MaxTokens)
			assert.Equal(t, tt.expectedCfg.Temperature, result.Temperature)
			assert.Equal(t, tt.expectedCfg.TopP, result.TopP)
		})
	}
}

func TestModel_convertClaudeRequest(t *testing.T) {
	model := &Model{
		modelID: "anthropic.claude-3-sonnet-20240229-v1:0",
		config: &ModelConfig{
			MaxTokens:     4096,
			Temperature:   0.7,
			TopP:          0.9,
			StopSequences: []string{"STOP"},
		},
	}

	req := &ai.ModelRequest{
		Messages: []*ai.Message{
			{
				Role: "user",
				Content: []*ai.Part{
					{Text: "Hello, world!"},
				},
			},
		},
	}

	result, err := model.convertClaudeRequest(req)
	require.NoError(t, err)

	var claudeReq map[string]interface{}
	err = json.Unmarshal(result, &claudeReq)
	require.NoError(t, err)

	assert.Equal(t, "bedrock-2023-05-31", claudeReq["anthropic_version"])
	assert.Equal(t, float64(4096), claudeReq["max_tokens"])
	assert.Equal(t, 0.7, claudeReq["temperature"])
	assert.Equal(t, 0.9, claudeReq["top_p"])
	assert.Equal(t, []interface{}{"STOP"}, claudeReq["stop_sequences"])

	messages, ok := claudeReq["messages"].([]interface{})
	require.True(t, ok)
	require.Len(t, messages, 1)

	firstMessage, ok := messages[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "user", firstMessage["role"])
	assert.Equal(t, "Hello, world!", firstMessage["content"])
}

func TestModel_convertClaudeResponse(t *testing.T) {
	model := &Model{
		modelID: "anthropic.claude-3-sonnet-20240229-v1:0",
		config:  &ModelConfig{},
	}

	claudeResponse := map[string]interface{}{
		"content": []map[string]interface{}{
			{"text": "Hello! How can I help you today?"},
		},
		"usage": map[string]interface{}{
			"input_tokens":  10,
			"output_tokens": 8,
		},
	}

	responseBody, err := json.Marshal(claudeResponse)
	require.NoError(t, err)

	result, err := model.convertClaudeResponse(responseBody)
	require.NoError(t, err)

	require.NotNil(t, result.Message)
	assert.Equal(t, ai.Role("model"), result.Message.Role)
	require.Len(t, result.Message.Content, 1)
	assert.Equal(t, "Hello! How can I help you today?", result.Message.Content[0].Text)

	assert.Equal(t, 10, result.Usage.InputTokens)
	assert.Equal(t, 8, result.Usage.OutputTokens)
	assert.Equal(t, 18, result.Usage.TotalTokens)
}

func TestIsClaudeModel(t *testing.T) {
	tests := []struct {
		modelID  string
		expected bool
	}{
		{"anthropic.claude-3-sonnet-20240229-v1:0", true},
		{"anthropic.claude-3-haiku-20240307-v1:0", true},
		{"ANTHROPIC.CLAUDE-3-SONNET-20240229-V1:0", true},
		{"amazon.nova-pro-v1:0", false},
		{"meta.llama3-2-90b-instruct-v1:0", false},
		{"claude-instant-v1", true},
		{"anthropic", true},
		{"CLAUDE", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			result := isClaudeModel(tt.modelID)
			assert.Equal(t, tt.expected, result, "Model ID: %s", tt.modelID)
		})
	}
}

func TestIsNovaModel(t *testing.T) {
	tests := []struct {
		modelID  string
		expected bool
	}{
		{"amazon.nova-pro-v1:0", true},
		{"amazon.nova-lite-v1:0", true},
		{"amazon.nova-micro-v1:0", true},
		{"AMAZON.NOVA-PRO-V1:0", true},
		{"anthropic.claude-3-sonnet-20240229-v1:0", false},
		{"meta.llama3-2-90b-instruct-v1:0", false},
		{"nova", true},
		{"NOVA", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			result := isNovaModel(tt.modelID)
			assert.Equal(t, tt.expected, result, "Model ID: %s", tt.modelID)
		})
	}
}

func TestIsLlamaModel(t *testing.T) {
	tests := []struct {
		modelID  string
		expected bool
	}{
		{"meta.llama3-2-90b-instruct-v1:0", true},
		{"meta.llama3-2-11b-instruct-v1:0", true},
		{"META.LLAMA3-2-90B-INSTRUCT-V1:0", true},
		{"anthropic.claude-3-sonnet-20240229-v1:0", false},
		{"amazon.nova-pro-v1:0", false},
		{"llama", true},
		{"LLAMA", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			result := isLlamaModel(tt.modelID)
			assert.Equal(t, tt.expected, result, "Model ID: %s", tt.modelID)
		})
	}
}
