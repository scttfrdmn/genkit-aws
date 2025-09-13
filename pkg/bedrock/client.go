// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

// Package bedrock provides AWS Bedrock integration for GenKit
package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/firebase/genkit/go/ai"
)

// Client wraps AWS Bedrock runtime client for GenKit integration
type Client struct {
	runtime *bedrockruntime.Client
	config  *Config
}

// NewClient creates a new Bedrock client
func NewClient(ctx context.Context, awsCfg aws.Config, config *Config) (*Client, error) {
	return &Client{
		runtime: bedrockruntime.NewFromConfig(awsCfg),
		config:  config,
	}, nil
}

// Model returns a GenKit-compatible model interface for the given model ID
func (c *Client) Model(modelID string) *Model {
	return &Model{
		client:  c,
		modelID: modelID,
		config:  c.config.ModelConfig(modelID),
	}
}

// Model represents a Bedrock model compatible with GenKit
type Model struct {
	client  *Client
	modelID string
	config  *ModelConfig
}

// Generate implements GenKit's generation interface
func (m *Model) Generate(ctx context.Context, req *ai.ModelRequest, cb ai.ModelStreamCallback) (*ai.ModelResponse, error) {
	// Convert GenKit request to Bedrock format
	bedrockReq, err := m.convertRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Call Bedrock
	result, err := m.client.runtime.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(m.modelID),
		ContentType: aws.String("application/json"),
		Body:        bedrockReq,
	})
	if err != nil {
		return nil, fmt.Errorf("bedrock invoke failed: %w", err)
	}

	// Convert Bedrock response to GenKit format
	response, err := m.convertResponse(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response: %w", err)
	}

	// Call streaming callback if provided
	if cb != nil && response.Message != nil && len(response.Message.Content) > 0 {
		chunk := &ai.ModelResponseChunk{
			Content: []*ai.Part{{Text: response.Message.Content[0].Text}},
		}
		if err := cb(ctx, chunk); err != nil {
			return nil, fmt.Errorf("callback failed: %w", err)
		}
	}

	return response, nil
}

// convertRequest converts GenKit request to Bedrock-specific format
func (m *Model) convertRequest(req *ai.ModelRequest) ([]byte, error) {
	// Implementation varies by model family (Claude, Nova, etc.)
	switch {
	case isClaudeModel(m.modelID):
		return m.convertClaudeRequest(req)
	case isNovaModel(m.modelID):
		return m.convertNovaRequest(req)
	case isLlamaModel(m.modelID):
		return m.convertLlamaRequest(req)
	default:
		return nil, fmt.Errorf("unsupported model: %s", m.modelID)
	}
}

// convertResponse converts Bedrock response to GenKit format
func (m *Model) convertResponse(body []byte) (*ai.ModelResponse, error) {
	switch {
	case isClaudeModel(m.modelID):
		return m.convertClaudeResponse(body)
	case isNovaModel(m.modelID):
		return m.convertNovaResponse(body)
	case isLlamaModel(m.modelID):
		return m.convertLlamaResponse(body)
	default:
		return nil, fmt.Errorf("unsupported model: %s", m.modelID)
	}
}

// Claude-specific request conversion
func (m *Model) convertClaudeRequest(req *ai.ModelRequest) ([]byte, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("no messages in request")
	}

	var messages []map[string]interface{}
	for _, msg := range req.Messages {
		if len(msg.Content) == 0 {
			continue
		}

		var role string
		switch msg.Role {
		case "model":
			role = "assistant"
		case "system":
			role = "user" // Claude handles system messages differently
		default:
			role = "user"
		}

		messages = append(messages, map[string]interface{}{
			"role":    role,
			"content": msg.Content[0].Text,
		})
	}

	claudeReq := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        m.config.MaxTokens,
		"temperature":       m.config.Temperature,
		"messages":          messages,
	}

	if m.config.TopP > 0 {
		claudeReq["top_p"] = m.config.TopP
	}

	if len(m.config.StopSequences) > 0 {
		claudeReq["stop_sequences"] = m.config.StopSequences
	}

	return json.Marshal(claudeReq)
}

// Claude-specific response conversion
func (m *Model) convertClaudeResponse(body []byte) (*ai.ModelResponse, error) {
	var claudeResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Claude response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("no content in Claude response")
	}

	return &ai.ModelResponse{
		Message: &ai.Message{
			Role: "model",
			Content: []*ai.Part{
				{Text: claudeResp.Content[0].Text},
			},
		},
		Usage: &ai.GenerationUsage{
			InputTokens:  claudeResp.Usage.InputTokens,
			OutputTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:  claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
		FinishReason: "stop",
	}, nil
}

// Nova-specific request conversion
func (m *Model) convertNovaRequest(req *ai.ModelRequest) ([]byte, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("no messages in request")
	}

	var messages []map[string]interface{}
	for _, msg := range req.Messages {
		if len(msg.Content) == 0 {
			continue
		}

		var role string
		switch msg.Role {
		case "model":
			role = "assistant"
		default:
			role = "user"
		}

		messages = append(messages, map[string]interface{}{
			"role": role,
			"content": []map[string]interface{}{
				{"text": msg.Content[0].Text},
			},
		})
	}

	novaReq := map[string]interface{}{
		"messages": messages,
		"inferenceConfig": map[string]interface{}{
			"maxTokens":   m.config.MaxTokens,
			"temperature": m.config.Temperature,
		},
	}

	if m.config.TopP > 0 {
		novaReq["inferenceConfig"].(map[string]interface{})["topP"] = m.config.TopP
	}

	if len(m.config.StopSequences) > 0 {
		novaReq["inferenceConfig"].(map[string]interface{})["stopSequences"] = m.config.StopSequences
	}

	return json.Marshal(novaReq)
}

// Nova-specific response conversion
func (m *Model) convertNovaResponse(body []byte) (*ai.ModelResponse, error) {
	var novaResp struct {
		Output struct {
			Message struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			} `json:"message"`
		} `json:"output"`
		Usage struct {
			InputTokens  int `json:"inputTokens"`
			OutputTokens int `json:"outputTokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &novaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Nova response: %w", err)
	}

	if len(novaResp.Output.Message.Content) == 0 {
		return nil, fmt.Errorf("no content in Nova response")
	}

	return &ai.ModelResponse{
		Message: &ai.Message{
			Role: "model",
			Content: []*ai.Part{
				{Text: novaResp.Output.Message.Content[0].Text},
			},
		},
		Usage: &ai.GenerationUsage{
			InputTokens:  novaResp.Usage.InputTokens,
			OutputTokens: novaResp.Usage.OutputTokens,
			TotalTokens:  novaResp.Usage.InputTokens + novaResp.Usage.OutputTokens,
		},
		FinishReason: "stop",
	}, nil
}

// Llama-specific request conversion
func (m *Model) convertLlamaRequest(req *ai.ModelRequest) ([]byte, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("no messages in request")
	}

	// Llama uses a simple prompt format
	var prompt strings.Builder
	for _, msg := range req.Messages {
		if len(msg.Content) == 0 {
			continue
		}
		prompt.WriteString(msg.Content[0].Text)
		prompt.WriteString("\n")
	}

	llamaReq := map[string]interface{}{
		"prompt":      prompt.String(),
		"max_gen_len": m.config.MaxTokens,
		"temperature": m.config.Temperature,
	}

	if m.config.TopP > 0 {
		llamaReq["top_p"] = m.config.TopP
	}

	return json.Marshal(llamaReq)
}

// Llama-specific response conversion
func (m *Model) convertLlamaResponse(body []byte) (*ai.ModelResponse, error) {
	var llamaResp struct {
		Generation           string `json:"generation"`
		PromptTokenCount     int    `json:"prompt_token_count"`
		GenerationTokenCount int    `json:"generation_token_count"`
	}

	if err := json.Unmarshal(body, &llamaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Llama response: %w", err)
	}

	return &ai.ModelResponse{
		Message: &ai.Message{
			Role: "model",
			Content: []*ai.Part{
				{Text: llamaResp.Generation},
			},
		},
		Usage: &ai.GenerationUsage{
			InputTokens:  llamaResp.PromptTokenCount,
			OutputTokens: llamaResp.GenerationTokenCount,
			TotalTokens:  llamaResp.PromptTokenCount + llamaResp.GenerationTokenCount,
		},
		FinishReason: "stop",
	}, nil
}

// Helper functions for model identification
func isClaudeModel(modelID string) bool {
	return strings.Contains(strings.ToLower(modelID), "claude") ||
		strings.Contains(strings.ToLower(modelID), "anthropic")
}

func isNovaModel(modelID string) bool {
	return strings.Contains(strings.ToLower(modelID), "nova")
}

func isLlamaModel(modelID string) bool {
	return strings.Contains(strings.ToLower(modelID), "llama")
}
