// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

// Package monitoring provides AWS CloudWatch integration for GenKit
package monitoring

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/genkit-aws/genkit-aws/internal/constants"
)

// CloudWatch implements GenKit monitoring using AWS CloudWatch
type CloudWatch struct {
	client    *cloudwatch.Client
	config    *Config
	namespace string

	// Metric buffering
	metricBuffer []types.MetricDatum
	bufferMutex  sync.Mutex
	bufferTicker *time.Ticker
	stopCh       chan struct{}
}

// NewCloudWatch creates a new CloudWatch monitoring instance
func NewCloudWatch(ctx context.Context, awsCfg aws.Config, config *Config) (*CloudWatch, error) {
	config.SetDefaults()

	cw := &CloudWatch{
		client:       cloudwatch.NewFromConfig(awsCfg),
		config:       config,
		namespace:    config.Namespace,
		metricBuffer: make([]types.MetricDatum, 0, config.MetricBufferSize),
		stopCh:       make(chan struct{}),
	}

	// Start metric buffer flush goroutine
	cw.bufferTicker = time.NewTicker(constants.MetricFlushInterval)
	go cw.flushLoop(ctx)

	return cw, nil
}

// Close stops the CloudWatch monitoring and flushes any remaining metrics
func (cw *CloudWatch) Close(ctx context.Context) error {
	close(cw.stopCh)
	cw.bufferTicker.Stop()

	// Flush any remaining metrics
	return cw.flushMetrics(ctx)
}

// OnFlowStart is called when a GenKit flow starts
func (cw *CloudWatch) OnFlowStart(ctx context.Context, flowName string, input interface{}) {
	if !cw.config.EnableFlowMetrics {
		return
	}

	dimensions := cw.buildDimensions(map[string]string{
		"FlowName": flowName,
	})

	cw.putMetric(ctx, "FlowStarted", 1.0, dimensions)
}

// OnFlowEnd is called when a GenKit flow completes successfully
func (cw *CloudWatch) OnFlowEnd(ctx context.Context, flowName string, duration time.Duration, output interface{}) {
	if !cw.config.EnableFlowMetrics {
		return
	}

	dimensions := cw.buildDimensions(map[string]string{
		"FlowName": flowName,
		"Status":   "Success",
	})

	// Record completion
	cw.putMetric(ctx, "FlowCompleted", 1.0, dimensions)

	// Record duration in milliseconds
	cw.putMetric(ctx, "FlowDuration", float64(duration.Milliseconds()), dimensions)
}

// OnFlowError is called when a GenKit flow fails
func (cw *CloudWatch) OnFlowError(ctx context.Context, flowName string, duration time.Duration, err error) {
	if !cw.config.EnableFlowMetrics {
		return
	}

	dimensions := cw.buildDimensions(map[string]string{
		"FlowName":  flowName,
		"Status":    "Error",
		"ErrorType": getErrorType(err),
	})

	cw.putMetric(ctx, "FlowError", 1.0, dimensions)
	cw.putMetric(ctx, "FlowDuration", float64(duration.Milliseconds()), dimensions)
}

// OnGenerate is called for each model generation
func (cw *CloudWatch) OnGenerate(ctx context.Context, modelID string, tokensUsed int, duration time.Duration) {
	if !cw.config.EnableModelMetrics {
		return
	}

	dimensions := cw.buildDimensions(map[string]string{
		"ModelID": modelID,
	})

	cw.putMetric(ctx, "TokensUsed", float64(tokensUsed), dimensions)
	cw.putMetric(ctx, "GenerationDuration", float64(duration.Milliseconds()), dimensions)
	cw.putMetric(ctx, "GenerationCount", 1.0, dimensions)
}

// putMetric adds a metric to the buffer
func (cw *CloudWatch) putMetric(ctx context.Context, metricName string, value float64, dimensions []types.Dimension) {
	metric := types.MetricDatum{
		MetricName: aws.String(metricName),
		Value:      aws.Float64(value),
		Unit:       types.StandardUnitCount,
		Timestamp:  aws.Time(time.Now()),
		Dimensions: dimensions,
	}

	cw.bufferMutex.Lock()
	defer cw.bufferMutex.Unlock()

	cw.metricBuffer = append(cw.metricBuffer, metric)

	// Flush if buffer is full
	if len(cw.metricBuffer) >= cw.config.MetricBufferSize {
		go func() {
			if err := cw.flushMetrics(ctx); err != nil {
				// Log error but don't block main operation
				_ = err // Acknowledge error handling
			}
		}()
	}
}

// flushLoop runs periodically to flush buffered metrics
func (cw *CloudWatch) flushLoop(ctx context.Context) {
	for {
		select {
		case <-cw.bufferTicker.C:
			if err := cw.flushMetrics(ctx); err != nil {
				// Log error but continue operation
				_ = err // Acknowledge error handling
			}
		case <-cw.stopCh:
			return
		}
	}
}

// flushMetrics sends buffered metrics to CloudWatch
func (cw *CloudWatch) flushMetrics(ctx context.Context) error {
	cw.bufferMutex.Lock()
	metricsToSend := make([]types.MetricDatum, len(cw.metricBuffer))
	copy(metricsToSend, cw.metricBuffer)
	cw.metricBuffer = cw.metricBuffer[:0]
	cw.bufferMutex.Unlock()

	if len(metricsToSend) == 0 {
		return nil
	}

	// CloudWatch PutMetricData has a limit of metrics per request
	const maxMetricsPerRequest = constants.MaxMetricsPerRequest

	for i := 0; i < len(metricsToSend); i += maxMetricsPerRequest {
		end := i + maxMetricsPerRequest
		if end > len(metricsToSend) {
			end = len(metricsToSend)
		}

		batch := metricsToSend[i:end]
		_, err := cw.client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
			Namespace:  aws.String(cw.namespace),
			MetricData: batch,
		})
		if err != nil {
			// Re-add failed metrics to buffer for retry
			cw.bufferMutex.Lock()
			cw.metricBuffer = append(cw.metricBuffer, batch...)
			cw.bufferMutex.Unlock()
			return err
		}
	}

	return nil
}

// buildDimensions creates CloudWatch dimensions with custom dimensions
func (cw *CloudWatch) buildDimensions(dimensions map[string]string) []types.Dimension {
	result := make([]types.Dimension, 0, len(dimensions)+len(cw.config.CustomDimensions))

	// Add custom dimensions first
	for name, value := range cw.config.CustomDimensions {
		result = append(result, types.Dimension{
			Name:  aws.String(name),
			Value: aws.String(value),
		})
	}

	// Add specific dimensions
	for name, value := range dimensions {
		result = append(result, types.Dimension{
			Name:  aws.String(name),
			Value: aws.String(value),
		})
	}

	return result
}

// getErrorType extracts error type for metrics
func getErrorType(err error) string {
	if err == nil {
		return "Unknown"
	}

	// Try to classify common error types
	errStr := err.Error()
	switch {
	case contains(errStr, "timeout"):
		return "Timeout"
	case contains(errStr, "connection"):
		return "Connection"
	case contains(errStr, "auth"):
		return "Authentication"
	case contains(errStr, "throttle"):
		return "Throttling"
	case contains(errStr, "rate"):
		return "RateLimit"
	default:
		return "GenericError"
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
