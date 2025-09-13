// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

package monitoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
				Namespace: "GenKit/Test",
			},
			wantErr: false,
		},
		{
			name:    "missing namespace",
			config:  &Config{},
			wantErr: true,
			errMsg:  "namespace is required",
		},
		{
			name: "empty namespace",
			config: &Config{
				Namespace: "",
			},
			wantErr: true,
			errMsg:  "namespace is required",
		},
		{
			name: "negative buffer size",
			config: &Config{
				Namespace:        "GenKit/Test",
				MetricBufferSize: -1,
			},
			wantErr: true,
			errMsg:  "metric_buffer_size must be non-negative",
		},
		{
			name: "valid config with all options",
			config: &Config{
				Namespace:          "GenKit/Test",
				EnableFlowMetrics:  true,
				EnableModelMetrics: true,
				EnableXRayTracing:  true,
				CustomDimensions: map[string]string{
					"Environment": "Test",
				},
				MetricBufferSize: 50,
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

func TestConfig_SetDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    *Config
		expected *Config
	}{
		{
			name:  "empty config gets defaults",
			input: &Config{},
			expected: &Config{
				Namespace:          "GenKit/AWS",
				EnableFlowMetrics:  true,
				EnableModelMetrics: true,
				EnableXRayTracing:  false,
				CustomDimensions:   map[string]string{},
				MetricBufferSize:   20,
			},
		},
		{
			name: "partial config preserves existing values",
			input: &Config{
				Namespace:         "Custom/Namespace",
				EnableXRayTracing: true,
			},
			expected: &Config{
				Namespace:          "Custom/Namespace",
				EnableFlowMetrics:  true,
				EnableModelMetrics: true,
				EnableXRayTracing:  true,
				CustomDimensions:   map[string]string{},
				MetricBufferSize:   20,
			},
		},
		{
			name: "config with only XRay enabled",
			input: &Config{
				Namespace:         "Test/Namespace",
				EnableXRayTracing: true,
			},
			expected: &Config{
				Namespace:          "Test/Namespace",
				EnableFlowMetrics:  true,
				EnableModelMetrics: true,
				EnableXRayTracing:  true,
				CustomDimensions:   map[string]string{},
				MetricBufferSize:   20,
			},
		},
		{
			name: "config with custom dimensions",
			input: &Config{
				CustomDimensions: map[string]string{
					"Environment": "Production",
				},
			},
			expected: &Config{
				Namespace:          "GenKit/AWS",
				EnableFlowMetrics:  true,
				EnableModelMetrics: true,
				EnableXRayTracing:  false,
				CustomDimensions: map[string]string{
					"Environment": "Production",
				},
				MetricBufferSize: 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.SetDefaults()

			assert.Equal(t, tt.expected.Namespace, tt.input.Namespace)
			assert.Equal(t, tt.expected.EnableFlowMetrics, tt.input.EnableFlowMetrics)
			assert.Equal(t, tt.expected.EnableModelMetrics, tt.input.EnableModelMetrics)
			assert.Equal(t, tt.expected.EnableXRayTracing, tt.input.EnableXRayTracing)
			assert.Equal(t, tt.expected.MetricBufferSize, tt.input.MetricBufferSize)
			assert.Equal(t, tt.expected.CustomDimensions, tt.input.CustomDimensions)
		})
	}
}

func TestGetErrorType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "Unknown",
		},
		{
			name:     "timeout error",
			err:      assert.AnError, // This will be treated as generic
			expected: "GenericError",
		},
		{
			name:     "connection error",
			err:      &testError{"connection failed"},
			expected: "Connection",
		},
		{
			name:     "authentication error",
			err:      &testError{"auth failed"},
			expected: "Authentication",
		},
		{
			name:     "throttling error",
			err:      &testError{"throttle limit exceeded"},
			expected: "Throttling",
		},
		{
			name:     "rate limit error",
			err:      &testError{"rate limit exceeded"},
			expected: "RateLimit",
		},
		{
			name:     "timeout with case variation",
			err:      &testError{"request TIMEOUT occurred"},
			expected: "Timeout",
		},
		{
			name:     "generic error",
			err:      &testError{"something went wrong"},
			expected: "GenericError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getErrorType(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "exact match",
			s:        "timeout",
			substr:   "timeout",
			expected: true,
		},
		{
			name:     "case insensitive match",
			s:        "TIMEOUT ERROR",
			substr:   "timeout",
			expected: true,
		},
		{
			name:     "substring match",
			s:        "connection timeout occurred",
			substr:   "timeout",
			expected: true,
		},
		{
			name:     "no match",
			s:        "connection failed",
			substr:   "timeout",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "any string",
			substr:   "",
			expected: true,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "timeout",
			expected: false,
		},
		{
			name:     "both empty",
			s:        "",
			substr:   "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// testError is a simple error implementation for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
