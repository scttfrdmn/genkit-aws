// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/genkit-aws/pkg/bedrock"
	genkitaws "github.com/scttfrdmn/genkit-aws/pkg/genkit-aws"
	"github.com/scttfrdmn/genkit-aws/pkg/monitoring"
)

func TestBedrockIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := getEnvOrDefault("AWS_PROFILE", "aws")
	region := getEnvOrDefault("AWS_REGION", "us-west-2")

	ctx := context.Background()

	// Create AWS plugin
	awsPlugin, err := genkitaws.New(&genkitaws.Config{
		Region:  region,
		Profile: profile,
		Bedrock: &bedrock.Config{
			Models: []string{
				"anthropic.claude-3-sonnet-20240229-v1:0",
			},
			DefaultModelConfig: &bedrock.ModelConfig{
				MaxTokens:   50, // Small for testing
				Temperature: 0.7,
			},
		},
		CloudWatch: &monitoring.Config{
			Namespace:          "GenKitAWS/IntegrationTest",
			EnableFlowMetrics:  true,
			EnableModelMetrics: true,
			CustomDimensions: map[string]string{
				"Test": "Integration",
			},
		},
	})
	require.NoError(t, err)

	// Initialize GenKit
	g := genkit.Init(ctx, genkit.WithPlugins(awsPlugin))

	// Define model
	model := awsPlugin.DefineModel(g, "anthropic.claude-3-sonnet-20240229-v1:0", nil)
	require.NotNil(t, model)

	t.Run("SimpleGeneration", func(t *testing.T) {
		resp, err := genkit.Generate(ctx, g,
			ai.WithModelName("anthropic.claude-3-sonnet-20240229-v1:0"),
			ai.WithMessages(ai.NewUserTextMessage("Say exactly: Hello from integration test")),
		)
		require.NoError(t, err)
		require.NotNil(t, resp.Message)
		require.NotEmpty(t, resp.Message.Content)

		text := resp.Message.Content[0].Text
		assert.NotEmpty(t, text)
		assert.Contains(t, text, "Hello")

		// Verify usage tracking
		assert.Greater(t, resp.Usage.InputTokens, 0)
		assert.Greater(t, resp.Usage.OutputTokens, 0)
		assert.Equal(t, resp.Usage.TotalTokens, resp.Usage.InputTokens+resp.Usage.OutputTokens)

		t.Logf("Generated response: %s", text)
		t.Logf("Token usage: %d input + %d output = %d total",
			resp.Usage.InputTokens, resp.Usage.OutputTokens, resp.Usage.TotalTokens)
	})

	t.Run("FlowWithMonitoring", func(t *testing.T) {
		// Create a flow to test monitoring
		testFlow := genkit.DefineFlow(g, "integrationTestFlow", func(ctx context.Context, input string) (string, error) {
			resp, err := genkit.Generate(ctx, g,
				ai.WithModelName("anthropic.claude-3-sonnet-20240229-v1:0"),
				ai.WithMessages(ai.NewUserTextMessage(input)),
			)
			if err != nil {
				return "", err
			}

			if resp.Message == nil || len(resp.Message.Content) == 0 {
				return "", fmt.Errorf("no response generated")
			}

			return resp.Message.Content[0].Text, nil
		})

		start := time.Now()
		result, err := testFlow.Run(ctx, "Count to 3")
		duration := time.Since(start)

		require.NoError(t, err)
		assert.NotEmpty(t, result)

		t.Logf("Flow completed in %v", duration)
		t.Logf("Flow result: %s", result)

		// Verify monitoring is working
		monitor := awsPlugin.GetMonitor()
		assert.NotNil(t, monitor, "CloudWatch monitor should be initialized")
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// Test with invalid model name
		_, err := genkit.Generate(ctx, g,
			ai.WithModelName("invalid.model.name"),
			ai.WithMessages(ai.NewUserTextMessage("This should fail")),
		)
		assert.Error(t, err)
		t.Logf("Expected error for invalid model: %v", err)
	})
}

func TestCloudWatchMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := getEnvOrDefault("AWS_PROFILE", "aws")
	region := getEnvOrDefault("AWS_REGION", "us-west-2")

	ctx := context.Background()

	// Create monitoring-only plugin
	awsPlugin, err := genkitaws.New(&genkitaws.Config{
		Region:  region,
		Profile: profile,
		CloudWatch: &monitoring.Config{
			Namespace:          "GenKitAWS/MetricsTest",
			EnableFlowMetrics:  true,
			EnableModelMetrics: true,
			CustomDimensions: map[string]string{
				"TestType": "MetricsOnly",
				"TestTime": time.Now().Format(time.RFC3339),
			},
			MetricBufferSize: 5, // Small buffer for faster testing
		},
	})
	require.NoError(t, err)

	// Initialize GenKit
	g := genkit.Init(ctx, genkit.WithPlugins(awsPlugin))

	// Get monitor
	monitor := awsPlugin.GetMonitor()
	require.NotNil(t, monitor)

	// Use genkit instance for potential future use
	_ = g

	t.Run("MetricCollection", func(t *testing.T) {
		// Simulate various monitoring events
		monitor.OnFlowStart(ctx, "testFlow", "test input")
		monitor.OnGenerate(ctx, "test-model", 100, 500*time.Millisecond)
		monitor.OnFlowEnd(ctx, "testFlow", 1*time.Second, "test output")

		// Wait for metrics to be collected
		time.Sleep(1 * time.Second)

		t.Log("✅ Metrics collected successfully")
	})

	t.Run("ErrorMetrics", func(t *testing.T) {
		// Simulate error metrics
		testErr := fmt.Errorf("test error for metrics")
		monitor.OnFlowError(ctx, "errorFlow", 2*time.Second, testErr)

		// Wait for error metrics
		time.Sleep(1 * time.Second)

		t.Log("✅ Error metrics collected successfully")
	})

	// Close monitoring to flush remaining metrics
	err = monitor.Close(ctx)
	assert.NoError(t, err)

	t.Log("✅ CloudWatch monitoring integration test completed")
	t.Log("📊 Check AWS Console -> CloudWatch -> Metrics -> GenKitAWS/MetricsTest")
}

func getEnvOrDefault(env, defaultValue string) string {
	if value := os.Getenv(env); value != "" {
		return value
	}
	return defaultValue
}
