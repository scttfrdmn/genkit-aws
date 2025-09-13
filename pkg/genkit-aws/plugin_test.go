// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

package genkitaws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Region: "us-east-1",
			},
			wantErr: false,
		},
		{
			name:    "nil config should use defaults",
			config:  nil,
			wantErr: true, // Will fail validation because region is required
		},
		{
			name:    "missing region",
			config:  &Config{},
			wantErr: true,
		},
		{
			name: "empty region",
			config: &Config{
				Region: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, err := New(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, plugin)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plugin)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid minimal config",
			config: &Config{
				Region: "us-east-1",
			},
			wantErr: false,
		},
		{
			name: "valid config with profile",
			config: &Config{
				Region:  "us-west-2",
				Profile: "my-profile",
			},
			wantErr: false,
		},
		{
			name: "missing region",
			config: &Config{
				Profile: "my-profile",
			},
			wantErr: true,
			errMsg:  "region is required",
		},
		{
			name: "empty region",
			config: &Config{
				Region: "",
			},
			wantErr: true,
			errMsg:  "region is required",
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

func TestPlugin_Init(t *testing.T) {
	// This test requires integration testing setup
	t.Skip("Integration test - requires AWS credentials and LocalStack setup")

	ctx := context.Background()

	config := &Config{
		Region: "us-east-1",
		// Add test-specific configuration
	}

	plugin, err := New(config)
	require.NoError(t, err)

	actions := plugin.Init(ctx)
	require.NotNil(t, actions)

	// Verify plugin is properly initialized
	assert.NotNil(t, plugin.config)
}
