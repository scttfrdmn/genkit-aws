// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

// Package genkitaws provides AWS integrations for Google's GenKit framework
package genkitaws

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/genkit-aws/genkit-aws/pkg/bedrock"
	"github.com/genkit-aws/genkit-aws/pkg/monitoring"
)

// Plugin represents the main GenKit AWS plugin
type Plugin struct {
	config  *Config
	bedrock *bedrock.Client
	monitor *monitoring.CloudWatch
}

// New creates a new GenKit AWS plugin instance
// Returns an error if the configuration is invalid
func New(config *Config) (*Plugin, error) {
	if config == nil {
		config = &Config{}
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid genkit-aws config: %w", err)
	}

	return &Plugin{config: config}, nil
}

// Name implements the Plugin interface
func (p *Plugin) Name() string {
	return "genkit-aws"
}

// Init implements the Plugin interface
func (p *Plugin) Init(ctx context.Context) []api.Action {
	// Initialize AWS session
	awsCfg, err := p.config.AWSConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to create AWS config: %w", err))
	}

	// Initialize Bedrock client if configured
	if p.config.Bedrock != nil {
		client, err := bedrock.NewClient(ctx, awsCfg, p.config.Bedrock)
		if err != nil {
			panic(fmt.Errorf("failed to initialize Bedrock client: %w", err))
		}
		p.bedrock = client
	}

	// Initialize CloudWatch monitoring if configured
	if p.config.CloudWatch != nil {
		monitor, err := monitoring.NewCloudWatch(ctx, awsCfg, p.config.CloudWatch)
		if err != nil {
			panic(fmt.Errorf("failed to initialize CloudWatch monitoring: %w", err))
		}
		p.monitor = monitor
	}

	return []api.Action{}
}

// DefineModel defines a Bedrock model in the given registry
func (p *Plugin) DefineModel(g *genkit.Genkit, name string, opts *ai.ModelOptions) ai.Model {
	if p.bedrock == nil {
		panic("plugin not initialized or Bedrock not configured")
	}

	bedrockModel := p.bedrock.Model(name)

	if opts == nil {
		opts = &ai.ModelOptions{
			Label: fmt.Sprintf("AWS Bedrock - %s", name),
			Supports: &ai.ModelSupports{
				Output:     []string{"text"},
				Tools:      false,
				Media:      false,
				Multiturn:  true,
				SystemRole: true,
			},
		}
	}

	return genkit.DefineModel(g, name, opts, bedrockModel.Generate)
}

// GetMonitor returns the CloudWatch monitor instance
func (p *Plugin) GetMonitor() *monitoring.CloudWatch {
	return p.monitor
}
