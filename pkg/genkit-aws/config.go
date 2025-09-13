// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

package genkitaws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/scttfrdmn/genkit-aws/pkg/bedrock"
	"github.com/scttfrdmn/genkit-aws/pkg/monitoring"
)

// Config holds configuration for the GenKit AWS plugin
type Config struct {
	// Region specifies the AWS region to use
	Region string `json:"region,omitempty"`

	// Profile specifies the AWS profile to use (optional)
	Profile string `json:"profile,omitempty"`

	// Bedrock configuration (optional)
	Bedrock *bedrock.Config `json:"bedrock,omitempty"`

	// CloudWatch monitoring configuration (optional)
	CloudWatch *monitoring.Config `json:"cloudwatch,omitempty"`

	// Additional AWS config options
	AWSConfigOptions []func(*config.LoadOptions) error `json:"-"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Region == "" {
		return errors.New("region is required")
	}

	if c.Bedrock != nil {
		if err := c.Bedrock.Validate(); err != nil {
			return fmt.Errorf("bedrock config invalid: %w", err)
		}
	}

	if c.CloudWatch != nil {
		if err := c.CloudWatch.Validate(); err != nil {
			return fmt.Errorf("cloudwatch config invalid: %w", err)
		}
	}

	return nil
}

// AWSConfig creates an AWS config from the plugin configuration
func (c *Config) AWSConfig(ctx context.Context) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(c.Region),
	}

	if c.Profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(c.Profile))
	}

	opts = append(opts, c.AWSConfigOptions...)

	return config.LoadDefaultConfig(ctx, opts...)
}
