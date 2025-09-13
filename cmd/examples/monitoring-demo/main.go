// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/scttfrdmn/genkit-aws/pkg/bedrock"
	genkitaws "github.com/scttfrdmn/genkit-aws/pkg/genkit-aws"
	"github.com/scttfrdmn/genkit-aws/pkg/monitoring"
)

func main() {
	ctx := context.Background()

	// Create AWS plugin with comprehensive monitoring
	awsPlugin, err := genkitaws.New(&genkitaws.Config{
		Region: "us-east-1",
		Bedrock: &bedrock.Config{
			Models: []string{
				"anthropic.claude-3-sonnet-20240229-v1:0",
				"amazon.nova-pro-v1:0",
				"amazon.nova-lite-v1:0",
			},
			DefaultModelConfig: &bedrock.ModelConfig{
				MaxTokens:   2048,
				Temperature: 0.8,
			},
		},
		CloudWatch: &monitoring.Config{
			Namespace:          "GenKitDemo/MonitoringExample",
			EnableFlowMetrics:  true,
			EnableModelMetrics: true,
			CustomDimensions: map[string]string{
				"Environment": "Development",
				"Application": "MonitoringDemo",
			},
			MetricBufferSize: 10,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize GenKit
	g := genkit.Init(ctx, genkit.WithPlugins(awsPlugin))

	// Define models
	awsPlugin.DefineModel(g, "anthropic.claude-3-sonnet-20240229-v1:0", nil)
	awsPlugin.DefineModel(g, "amazon.nova-pro-v1:0", nil)
	awsPlugin.DefineModel(g, "amazon.nova-lite-v1:0", nil)

	// Define multiple flows to demonstrate monitoring
	chatFlow := genkit.DefineFlow(g, "chatFlow", generateResponse)
	summarizeFlow := genkit.DefineFlow(g, "summarizeFlow", summarizeText)
	analyzeFlow := genkit.DefineFlow(g, "analyzeFlow", analyzeContent)

	// Define test scenarios
	testCases := []struct {
		name  string
		flow  *core.Flow[string, string, struct{}]
		input string
		model string
	}{
		{
			name:  "Chat with Claude",
			flow:  chatFlow,
			input: "Explain quantum computing in simple terms",
			model: "anthropic.claude-3-sonnet-20240229-v1:0",
		},
		{
			name:  "Chat with Nova Pro",
			flow:  chatFlow,
			input: "What are the benefits of renewable energy?",
			model: "amazon.nova-pro-v1:0",
		},
		{
			name:  "Chat with Nova Lite",
			flow:  chatFlow,
			input: "Write a haiku about technology",
			model: "amazon.nova-lite-v1:0",
		},
		{
			name:  "Summarize Text",
			flow:  summarizeFlow,
			input: "The field of artificial intelligence has grown exponentially over the past decade...",
			model: "anthropic.claude-3-sonnet-20240229-v1:0",
		},
		{
			name:  "Analyze Content",
			flow:  analyzeFlow,
			input: "This product has great features but the price is too high for most consumers.",
			model: "amazon.nova-pro-v1:0",
		},
	}

	fmt.Println("Starting GenKit AWS Monitoring Demo")
	fmt.Println("Running multiple flows to generate metrics...")
	fmt.Println()

	// Run test cases multiple times to generate interesting metrics
	for round := 1; round <= 3; round++ {
		fmt.Printf("=== Round %d ===\n", round)

		for i, tc := range testCases {
			fmt.Printf("[%d/%d] %s\n", i+1, len(testCases), tc.name)

			start := time.Now()

			// Set the model context for this test
			ctx = context.WithValue(ctx, modelKey("model"), tc.model)

			result, err := tc.flow.Run(ctx, tc.input)

			duration := time.Since(start)

			if err != nil {
				fmt.Printf("  ❌ Error: %v (took %v)\n", err, duration)
				continue
			}

			fmt.Printf("  ✅ Success (took %v)\n", duration)
			fmt.Printf("  📝 Response: %s...\n", truncateString(result, 80))

			// Add some random delay to make metrics more interesting
			time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
		}

		fmt.Println()
		time.Sleep(1 * time.Second)
	}

	// Simulate some errors for error metrics
	fmt.Println("=== Error Simulation ===")

	errorFlow := genkit.DefineFlow(g, "errorFlow", func(ctx context.Context, input string) (string, error) {
		return "", fmt.Errorf("simulated error: %s", input)
	})

	for i := 1; i <= 3; i++ {
		fmt.Printf("Generating error %d/3\n", i)
		_, err := errorFlow.Run(ctx, fmt.Sprintf("error-%d", i))
		if err != nil {
			fmt.Printf("  ❌ Expected error: %v\n", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println()
	fmt.Println("Demo completed!")
	fmt.Println("Check AWS CloudWatch metrics under namespace: GenKitDemo/MonitoringExample")
	fmt.Println("Expected metrics:")
	fmt.Println("  - FlowStarted, FlowCompleted, FlowError")
	fmt.Println("  - FlowDuration")
	fmt.Println("  - TokensUsed, GenerationDuration, GenerationCount")

	// Wait a bit for metrics to be flushed
	time.Sleep(2 * time.Second)
}

func generateResponse(ctx context.Context, input string) (string, error) {
	// For demo purposes, simulate a response generation
	// In a real implementation, this would use the GenKit registry
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	return fmt.Sprintf("Generated response for: %s", input), nil
}

func summarizeText(ctx context.Context, input string) (string, error) {
	// Simulate summarization
	time.Sleep(time.Duration(rand.Intn(800)) * time.Millisecond)
	return fmt.Sprintf("Summary of: %s", input[:min(50, len(input))]), nil
}

func analyzeContent(ctx context.Context, input string) (string, error) {
	// Simulate content analysis
	time.Sleep(time.Duration(rand.Intn(1200)) * time.Millisecond)
	return fmt.Sprintf("Analysis of: %s", input[:min(30, len(input))]), nil
}

type modelKey string

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
