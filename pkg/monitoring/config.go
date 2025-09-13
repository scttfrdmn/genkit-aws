// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

package monitoring

import (
	"errors"

	"github.com/scttfrdmn/genkit-aws/internal/constants"
)

// Config holds configuration for CloudWatch monitoring
type Config struct {
	// Namespace is the CloudWatch namespace for metrics
	Namespace string `json:"namespace,omitempty"`

	// EnableFlowMetrics controls whether to track flow performance
	EnableFlowMetrics bool `json:"enable_flow_metrics,omitempty"`

	// EnableModelMetrics controls whether to track model usage
	EnableModelMetrics bool `json:"enable_model_metrics,omitempty"`

	// EnableXRayTracing controls whether to enable X-Ray tracing
	EnableXRayTracing bool `json:"enable_xray_tracing,omitempty"`

	// CustomDimensions are additional dimensions to add to all metrics
	CustomDimensions map[string]string `json:"custom_dimensions,omitempty"`

	// MetricBufferSize controls how many metrics to buffer before sending
	MetricBufferSize int `json:"metric_buffer_size,omitempty"`
}

// Validate validates the monitoring configuration
func (c *Config) Validate() error {
	if c.Namespace == "" {
		return errors.New("namespace is required")
	}

	if c.MetricBufferSize < 0 {
		return errors.New("metric_buffer_size must be non-negative")
	}

	return nil
}

// SetDefaults sets default values for the configuration
func (c *Config) SetDefaults() {
	if c.Namespace == "" {
		c.Namespace = constants.DefaultNamespace
	}

	// Enable flow and model metrics by default if they're not explicitly set
	if !c.EnableFlowMetrics && !c.EnableModelMetrics && !c.EnableXRayTracing {
		c.EnableFlowMetrics = true
		c.EnableModelMetrics = true
	} else if c.EnableXRayTracing && !c.EnableFlowMetrics && !c.EnableModelMetrics {
		// If only XRay is enabled, also enable the other metrics by default
		c.EnableFlowMetrics = true
		c.EnableModelMetrics = true
	}

	if c.MetricBufferSize == 0 {
		c.MetricBufferSize = constants.DefaultBufferSize
	}

	if c.CustomDimensions == nil {
		c.CustomDimensions = make(map[string]string)
	}
}
