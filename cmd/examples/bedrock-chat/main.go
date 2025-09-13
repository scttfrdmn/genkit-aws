// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/genkit-aws/genkit-aws/pkg/bedrock"
	genkitaws "github.com/genkit-aws/genkit-aws/pkg/genkit-aws"
	"github.com/genkit-aws/genkit-aws/pkg/monitoring"
)

func main() {
	ctx := context.Background()

	// Create AWS plugin
	awsPlugin, err := genkitaws.New(&genkitaws.Config{
		Region: "us-east-1",
		Bedrock: &bedrock.Config{
			Models: []string{
				"anthropic.claude-3-sonnet-20240229-v1:0",
				"amazon.nova-pro-v1:0",
			},
			DefaultModelConfig: &bedrock.ModelConfig{
				MaxTokens:   4096,
				Temperature: 0.7,
				TopP:        0.9,
			},
		},
		CloudWatch: &monitoring.Config{
			Namespace:          "GenKitExample/BedrockChat",
			EnableFlowMetrics:  true,
			EnableModelMetrics: true,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize GenKit with the AWS plugin
	g := genkit.Init(ctx, genkit.WithPlugins(awsPlugin))

	// Define the Claude model
	awsPlugin.DefineModel(g, "anthropic.claude-3-sonnet-20240229-v1:0", nil)

	// Define a simple chat flow
	chatFlow := genkit.DefineFlow(g, "chatFlow", func(ctx context.Context, question string) (string, error) {
		resp, err := genkit.Generate(ctx, g,
			ai.WithModelName("anthropic.claude-3-sonnet-20240229-v1:0"),
			ai.WithMessages(ai.NewUserTextMessage(question)),
		)
		if err != nil {
			return "", fmt.Errorf("generation failed: %w", err)
		}

		if resp.Message == nil || len(resp.Message.Content) == 0 {
			return "", fmt.Errorf("no response generated")
		}

		return resp.Message.Content[0].Text, nil
	})

	// Test the flow
	question := "What is the capital of France?"
	if len(os.Args) > 1 {
		question = os.Args[1]
	}

	fmt.Printf("Question: %s\n", question)
	fmt.Println("Generating response...")

	result, err := chatFlow.Run(ctx, question)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Answer: %s\n", result)
}
