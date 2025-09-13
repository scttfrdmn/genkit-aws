// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

// Package constants defines common constants used throughout the GenKit AWS project
package constants

import "time"

// Default configuration values
const (
	// DefaultMaxTokens is the default maximum number of tokens to generate
	DefaultMaxTokens = 4096

	// DefaultTemperature is the default temperature for generation
	DefaultTemperature = 0.7

	// DefaultTopP is the default top-p value for nucleus sampling
	DefaultTopP = 0.9

	// DefaultNamespace is the default CloudWatch namespace
	DefaultNamespace = "GenKit/AWS"

	// DefaultBufferSize is the default number of metrics to buffer before sending
	DefaultBufferSize = 20

	// MetricFlushInterval is how often buffered metrics are flushed
	MetricFlushInterval = 30 * time.Second

	// MaxMetricsPerRequest is the CloudWatch limit for metrics per API call
	MaxMetricsPerRequest = 20

	// MaxRetries is the maximum number of times to retry failed operations
	MaxRetries = 3
)
