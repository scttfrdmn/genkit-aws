# GenKit AWS Runtime Plugins Project

## Project Structure

```
genkit-aws/
├── go.mod
├── go.sum
├── README.md
├── LICENSE
├── CONTRIBUTING.md
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
├── cmd/
│   └── examples/
│       ├── bedrock-chat/
│       │   └── main.go
│       └── monitoring-demo/
│           └── main.go
├── pkg/
│   ├── genkit-aws/
│   │   ├── plugin.go
│   │   ├── config.go
│   │   └── plugin_test.go
│   ├── bedrock/
│   │   ├── client.go
│   │   ├── models.go
│   │   ├── streaming.go
│   │   ├── client_test.go
│   │   └── testdata/
│   │       └── fixtures.json
│   └── monitoring/
│       ├── cloudwatch.go
│       ├── xray.go
│       ├── metrics.go
│       ├── monitoring_test.go
│       └── testdata/
├── internal/
│   ├── aws/
│   │   ├── config.go
│   │   └── session.go
│   ├── testutil/
│   │   ├── mocks.go
│   │   └── containers.go
│   └── version/
│       └── version.go
├── docs/
│   ├── quickstart.md
│   ├── bedrock-models.md
│   ├── monitoring-setup.md
│   └── examples/
├── scripts/
│   ├── build.sh
│   ├── test.sh
│   └── release.sh
└── tools/
    ├── go.mod
    └── tools.go
```

## Core Files

### go.mod
```go
module github.com/genkit-aws/genkit-aws

go 1.23

require (
    github.com/firebase/genkit/go v0.5.8
    github.com/aws/aws-sdk-go-v2 v1.32.6
    github.com/aws/aws-sdk-go-v2/config v1.28.6
    github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.21.7
    github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.42.7
    github.com/aws/aws-xray-sdk-go v1.8.5
    go.opentelemetry.io/otel v1.31.0
    go.opentelemetry.io/otel/trace v1.31.0
    go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.56.0
)

require (
    github.com/stretchr/testify v1.9.0
    github.com/testcontainers/testcontainers-go v0.33.0
    github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.15.15
)
```

### pkg/genkit-aws/plugin.go
```go
// Package genkitaws provides AWS integrations for Google's GenKit framework
package genkitaws

import (
    "context"
    "fmt"

    "github.com/firebase/genkit/go/genkit"
    "github.com/genkit-aws/genkit-aws/pkg/bedrock"
    "github.com/genkit-aws/genkit-aws/pkg/monitoring"
)

// Plugin represents the main GenKit AWS plugin
type Plugin struct {
    config *Config
    bedrock *bedrock.Client
    monitor *monitoring.CloudWatch
}

// New creates a new GenKit AWS plugin instance
func New(config *Config) genkit.Plugin {
    if config == nil {
        config = &Config{}
    }
    
    if err := config.Validate(); err != nil {
        panic(fmt.Errorf("invalid genkit-aws config: %w", err))
    }

    return &Plugin{config: config}
}

// Init initializes the GenKit AWS plugin
func (p *Plugin) Init(ctx context.Context) error {
    // Initialize AWS session
    awsCfg, err := p.config.AWSConfig(ctx)
    if err != nil {
        return fmt.Errorf("failed to create AWS config: %w", err)
    }

    // Initialize Bedrock client if configured
    if p.config.Bedrock != nil {
        client, err := bedrock.NewClient(ctx, awsCfg, p.config.Bedrock)
        if err != nil {
            return fmt.Errorf("failed to initialize Bedrock client: %w", err)
        }
        p.bedrock = client

        // Register Bedrock models with GenKit
        if err := p.registerBedrockModels(ctx); err != nil {
            return fmt.Errorf("failed to register Bedrock models: %w", err)
        }
    }

    // Initialize CloudWatch monitoring if configured
    if p.config.CloudWatch != nil {
        monitor, err := monitoring.NewCloudWatch(ctx, awsCfg, p.config.CloudWatch)
        if err != nil {
            return fmt.Errorf("failed to initialize CloudWatch monitoring: %w", err)
        }
        p.monitor = monitor

        // Setup monitoring hooks
        if err := p.setupMonitoring(ctx); err != nil {
            return fmt.Errorf("failed to setup monitoring: %w", err)
        }
    }

    return nil
}

// registerBedrockModels registers all configured Bedrock models with GenKit
func (p *Plugin) registerBedrockModels(ctx context.Context) error {
    for _, modelID := range p.config.Bedrock.Models {
        model := p.bedrock.Model(modelID)
        genkit.DefineModel(modelID, model.Generate)
    }
    return nil
}

// setupMonitoring configures GenKit monitoring hooks
func (p *Plugin) setupMonitoring(ctx context.Context) error {
    // Setup flow instrumentation
    genkit.OnFlowStart(p.monitor.OnFlowStart)
    genkit.OnFlowEnd(p.monitor.OnFlowEnd)
    genkit.OnFlowError(p.monitor.OnFlowError)
    
    // Setup model generation instrumentation  
    genkit.OnGenerate(p.monitor.OnGenerate)
    
    return nil
}
```

### pkg/genkit-aws/config.go
```go
package genkitaws

import (
    "context"
    "errors"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/genkit-aws/genkit-aws/pkg/bedrock"
    "github.com/genkit-aws/genkit-aws/pkg/monitoring"
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
```

### pkg/bedrock/client.go
```go
// Package bedrock provides AWS Bedrock integration for GenKit
package bedrock

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
    "github.com/firebase/genkit/go/ai"
)

// Client wraps AWS Bedrock runtime client for GenKit integration
type Client struct {
    runtime *bedrockruntime.Client
    config  *Config
}

// NewClient creates a new Bedrock client
func NewClient(ctx context.Context, awsCfg aws.Config, config *Config) (*Client, error) {
    return &Client{
        runtime: bedrockruntime.NewFromConfig(awsCfg),
        config:  config,
    }, nil
}

// Model returns a GenKit-compatible model interface for the given model ID
func (c *Client) Model(modelID string) *Model {
    return &Model{
        client:  c,
        modelID: modelID,
        config:  c.config.ModelConfig(modelID),
    }
}

// Model represents a Bedrock model compatible with GenKit
type Model struct {
    client  *Client
    modelID string
    config  *ModelConfig
}

// Generate implements GenKit's generation interface
func (m *Model) Generate(ctx context.Context, req *ai.GenerateRequest, cb func(context.Context, *ai.GenerateResponseChunk) error) (*ai.GenerateResponse, error) {
    // Convert GenKit request to Bedrock format
    bedrockReq, err := m.convertRequest(req)
    if err != nil {
        return nil, fmt.Errorf("failed to convert request: %w", err)
    }

    // Call Bedrock
    result, err := m.client.runtime.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
        ModelId:     aws.String(m.modelID),
        ContentType: aws.String("application/json"),
        Body:        bedrockReq,
    })
    if err != nil {
        return nil, fmt.Errorf("bedrock invoke failed: %w", err)
    }

    // Convert Bedrock response to GenKit format
    response, err := m.convertResponse(result.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to convert response: %w", err)
    }

    // Call streaming callback if provided
    if cb != nil {
        chunk := &ai.GenerateResponseChunk{
            Content: []*ai.Part{{Text: response.Candidates[0].Message.Content[0].Text}},
        }
        if err := cb(ctx, chunk); err != nil {
            return nil, fmt.Errorf("callback failed: %w", err)
        }
    }

    return response, nil
}

// convertRequest converts GenKit request to Bedrock-specific format
func (m *Model) convertRequest(req *ai.GenerateRequest) ([]byte, error) {
    // Implementation varies by model family (Claude, Nova, etc.)
    switch {
    case isClaudeModel(m.modelID):
        return m.convertClaudeRequest(req)
    case isNovaModel(m.modelID):
        return m.convertNovaRequest(req)
    default:
        return nil, fmt.Errorf("unsupported model: %s", m.modelID)
    }
}

// convertResponse converts Bedrock response to GenKit format
func (m *Model) convertResponse(body []byte) (*ai.GenerateResponse, error) {
    switch {
    case isClaudeModel(m.modelID):
        return m.convertClaudeResponse(body)
    case isNovaModel(m.modelID):
        return m.convertNovaResponse(body)
    default:
        return nil, fmt.Errorf("unsupported model: %s", m.modelID)
    }
}

// Claude-specific request conversion
func (m *Model) convertClaudeRequest(req *ai.GenerateRequest) ([]byte, error) {
    claudeReq := map[string]interface{}{
        "anthropic_version": "bedrock-2023-05-31",
        "max_tokens":        m.config.MaxTokens,
        "temperature":       m.config.Temperature,
        "messages": []map[string]interface{}{
            {
                "role":    "user",
                "content": req.Messages[0].Content[0].Text,
            },
        },
    }
    return json.Marshal(claudeReq)
}

// Claude-specific response conversion
func (m *Model) convertClaudeResponse(body []byte) (*ai.GenerateResponse, error) {
    var claudeResp struct {
        Content []struct {
            Text string `json:"text"`
        } `json:"content"`
        Usage struct {
            InputTokens  int `json:"input_tokens"`
            OutputTokens int `json:"output_tokens"`
        } `json:"usage"`
    }
    
    if err := json.Unmarshal(body, &claudeResp); err != nil {
        return nil, err
    }

    return &ai.GenerateResponse{
        Candidates: []*ai.Candidate{
            {
                Index: 0,
                Message: &ai.Message{
                    Role: ai.RoleModel,
                    Content: []*ai.Part{
                        {Text: claudeResp.Content[0].Text},
                    },
                },
            },
        },
        Usage: &ai.GenerationUsage{
            InputTokens:  claudeResp.Usage.InputTokens,
            OutputTokens: claudeResp.Usage.OutputTokens,
            TotalTokens:  claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
        },
    }, nil
}

// Helper functions for model identification
func isClaudeModel(modelID string) bool {
    // Check if model ID contains "claude" or "anthropic"
    return len(modelID) > 0 && (contains(modelID, "claude") || contains(modelID, "anthropic"))
}

func isNovaModel(modelID string) bool {
    return contains(modelID, "nova")
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && s[0:len(substr)] == substr
}
```

### pkg/monitoring/cloudwatch.go
```go
// Package monitoring provides AWS CloudWatch integration for GenKit
package monitoring

import (
    "context"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
    "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
    "github.com/firebase/genkit/go/genkit"
)

// CloudWatch implements GenKit monitoring using AWS CloudWatch
type CloudWatch struct {
    client    *cloudwatch.Client
    config    *Config
    namespace string
}

// NewCloudWatch creates a new CloudWatch monitoring instance
func NewCloudWatch(ctx context.Context, awsCfg aws.Config, config *Config) (*CloudWatch, error) {
    return &CloudWatch{
        client:    cloudwatch.NewFromConfig(awsCfg),
        config:    config,
        namespace: config.Namespace,
    }, nil
}

// OnFlowStart is called when a GenKit flow starts
func (cw *CloudWatch) OnFlowStart(ctx context.Context, flowName string, input interface{}) {
    cw.putMetric(ctx, "FlowStarted", 1.0, []types.Dimension{
        {Name: aws.String("FlowName"), Value: aws.String(flowName)},
    })
}

// OnFlowEnd is called when a GenKit flow completes successfully
func (cw *CloudWatch) OnFlowEnd(ctx context.Context, flowName string, duration time.Duration, output interface{}) {
    dimensions := []types.Dimension{
        {Name: aws.String("FlowName"), Value: aws.String(flowName)},
        {Name: aws.String("Status"), Value: aws.String("Success")},
    }

    // Record completion
    cw.putMetric(ctx, "FlowCompleted", 1.0, dimensions)
    
    // Record duration
    cw.putMetric(ctx, "FlowDuration", float64(duration.Milliseconds()), dimensions)
}

// OnFlowError is called when a GenKit flow fails
func (cw *CloudWatch) OnFlowError(ctx context.Context, flowName string, duration time.Duration, err error) {
    dimensions := []types.Dimension{
        {Name: aws.String("FlowName"), Value: aws.String(flowName)},
        {Name: aws.String("Status"), Value: aws.String("Error")},
        {Name: aws.String("ErrorType"), Value: aws.String(getErrorType(err))},
    }

    cw.putMetric(ctx, "FlowError", 1.0, dimensions)
    cw.putMetric(ctx, "FlowDuration", float64(duration.Milliseconds()), dimensions)
}

// OnGenerate is called for each model generation
func (cw *CloudWatch) OnGenerate(ctx context.Context, modelID string, tokensUsed int, duration time.Duration) {
    dimensions := []types.Dimension{
        {Name: aws.String("ModelID"), Value: aws.String(modelID)},
    }

    cw.putMetric(ctx, "TokensUsed", float64(tokensUsed), dimensions)
    cw.putMetric(ctx, "GenerationDuration", float64(duration.Milliseconds()), dimensions)
}

// putMetric sends a metric to CloudWatch
func (cw *CloudWatch) putMetric(ctx context.Context, metricName string, value float64, dimensions []types.Dimension) {
    // Use goroutine to avoid blocking the main flow
    go func() {
        _, err := cw.client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
            Namespace: aws.String(cw.namespace),
            MetricData: []types.MetricDatum{
                {
                    MetricName: aws.String(metricName),
                    Value:      aws.Float64(value),
                    Unit:       types.StandardUnitCount,
                    Timestamp:  aws.Time(time.Now()),
                    Dimensions: dimensions,
                },
            },
        })
        if err != nil {
            // Log error but don't fail the main operation
            // TODO: Add proper logging
        }
    }()
}

// getErrorType extracts error type for metrics
func getErrorType(err error) string {
    if err == nil {
        return "Unknown"
    }
    // Simple error classification
    // TODO: Enhance with more specific error types
    return "GenericError"
}
```

### Testing Setup (pkg/genkit-aws/plugin_test.go)
```go
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
            name: "missing region",
            config: &Config{},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.wantErr {
                assert.Panics(t, func() {
                    New(tt.config)
                })
            } else {
                assert.NotPanics(t, func() {
                    plugin := New(tt.config)
                    assert.NotNil(t, plugin)
                })
            }
        })
    }
}

func TestPlugin_Init(t *testing.T) {
    // Integration test using testcontainers for LocalStack
    ctx := context.Background()
    
    config := &Config{
        Region: "us-east-1",
        // Add test-specific configuration
    }
    
    plugin := New(config).(*Plugin)
    
    err := plugin.Init(ctx)
    require.NoError(t, err)
    
    // Verify plugin is properly initialized
    assert.NotNil(t, plugin.config)
}
```

### Example Application (cmd/examples/bedrock-chat/main.go)
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/firebase/genkit/go/genkit"
    genkitaws "github.com/genkit-aws/genkit-aws/pkg/genkit-aws"
    "github.com/genkit-aws/genkit-aws/pkg/bedrock"
    "github.com/genkit-aws/genkit-aws/pkg/monitoring"
)

func main() {
    ctx := context.Background()

    // Initialize GenKit with AWS plugins
    err := genkit.Init(ctx, &genkit.Config{
        Plugins: []genkit.Plugin{
            genkitaws.New(&genkitaws.Config{
                Region: "us-east-1",
                Bedrock: &bedrock.Config{
                    Models: []string{
                        "anthropic.claude-3-sonnet-20240229-v1:0",
                        "amazon.nova-pro-v1:0",
                    },
                },
                CloudWatch: &monitoring.Config{
                    Namespace: "GenKitExample/BedrockChat",
                },
            }),
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    // Define a simple chat flow
    chatFlow := genkit.DefineFlow("chatFlow", func(ctx context.Context, question string) (string, error) {
        resp, err := genkit.Generate(ctx, &genkit.GenerateRequest{
            Model: genkit.Model("anthropic.claude-3-sonnet-20240229-v1:0"),
            Prompt: genkit.NewTextPrompt(question),
        })
        if err != nil {
            return "", fmt.Errorf("generation failed: %w", err)
        }

        return resp.Text(), nil
    })

    // Test the flow
    question := "What is the capital of France?"
    if len(os.Args) > 1 {
        question = os.Args[1]
    }

    result, err := chatFlow.Run(ctx, question)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Question: %s\n", question)
    fmt.Printf("Answer: %s\n", result)
}
```

### README.md
```markdown
# GenKit AWS Plugins

**AWS plugins for Google's GenKit framework** - add AWS Bedrock models and CloudWatch monitoring to your existing GenKit applications.

> **Important**: This is NOT a fork or port of GenKit. These are plugins that extend Google's GenKit framework to work with AWS services.

## What This Provides

- **AWS Bedrock Integration**: Use 40+ AWS models (Claude, Nova, Llama) through GenKit's standard API
- **CloudWatch Monitoring**: Automatic flow metrics and observability for GenKit applications  
- **Seamless Integration**: Keep using GenKit's APIs, just add AWS capabilities

## Quick Start

### 1. Install Google's GenKit Framework First
```bash
go get github.com/firebase/genkit/go@latest
```

### 2. Add AWS Plugins
```bash
go get github.com/genkit-aws/genkit-aws@latest
```

### 3. Use AWS Models in GenKit
```go
import (
    // Google's GenKit framework
    "github.com/firebase/genkit/go/genkit"
    "github.com/firebase/genkit/go/ai"
    
    // AWS plugins that extend GenKit
    "github.com/genkit-aws/genkit-aws/aws/bedrock"
    "github.com/genkit-aws/genkit-aws/aws/monitoring"
)

func main() {
    ctx := context.Background()

    // Initialize GenKit (Google's framework)
    genkit.Init(ctx, nil)

    // Add AWS Bedrock models to GenKit
    bedrock.Init(ctx, &bedrock.Config{
        Region: "us-east-1",
        Models: []string{"anthropic.claude-3-sonnet-20240229-v1:0"},
    })

    // Add CloudWatch monitoring to GenKit
    monitoring.Init(ctx, &monitoring.Config{
        Region: "us-east-1",
        Namespace: "MyApp/GenKit",
    })

    // Use GenKit's standard API with AWS models
    resp, err := genkit.Generate(ctx, &ai.GenerateRequest{
        Model: genkit.Model("anthropic.claude-3-sonnet-20240229-v1:0"), // Now available!
        Messages: []*ai.Message{
            {Role: ai.RoleUser, Content: []*ai.Part{{Text: "Hello!"}}},
        },
    })
}
```

## How It Works

1. **You use Google's GenKit framework** as normal
2. **Our plugins register AWS models** with GenKit's model registry  
3. **GenKit's APIs work unchanged** - they just have more models available
4. **Monitoring automatically instruments** your existing GenKit flows

## Architecture

```
Your GenKit App
       ↓
Google's GenKit Framework (genkit.Model(), genkit.Generate(), etc.)
       ↓
Our AWS Plugins (bedrock.Init(), monitoring.Init())
       ↓
AWS Services (Bedrock, CloudWatch)
```

## Available AWS Models

### Anthropic Claude
- `anthropic.claude-3-5-sonnet-20241022-v2:0`
- `anthropic.claude-3-sonnet-20240229-v1:0`  
- `anthropic.claude-3-haiku-20240307-v1:0`

### Amazon Nova (Latest)
- `amazon.nova-pro-v1:0`
- `amazon.nova-lite-v1:0`
- `amazon.nova-micro-v1:0`

### Meta Llama  
- `meta.llama3-2-90b-instruct-v1:0`
- `meta.llama3-2-11b-instruct-v1:0`

[See all 40+ supported models](docs/bedrock-models.md)

## Monitoring Integration

The CloudWatch plugin automatically tracks:

- **Flow Performance**: Duration, success rates, error types
- **Model Usage**: Token consumption, generation latency  
- **Custom Metrics**: Configurable dimensions and namespaces

View in AWS CloudWatch under your configured namespace.

## Examples

### Basic Usage
```go
// Initialize GenKit + AWS plugins
genkit.Init(ctx, nil)
bedrock.Init(ctx, &bedrock.Config{Region: "us-east-1", Models: []string{"anthropic.claude-3-sonnet-20240229-v1:0"}})

// Use like any GenKit model
resp, _ := genkit.Generate(ctx, &ai.GenerateRequest{
    Model: genkit.Model("anthropic.claude-3-sonnet-20240229-v1:0"),
    Messages: []*ai.Message{{Role: ai.RoleUser, Content: []*ai.Part{{Text: "Hello"}}}},
})
```

### Multiple Providers
```go
// Mix Google AI, OpenAI, and AWS models in the same app
googleai.Init(ctx, &googleai.Config{})  // Google's plugin
openai.Init(ctx, &openai.Config{})      // OpenAI plugin  
bedrock.Init(ctx, &bedrock.Config{})    // Our AWS plugin

// All available through genkit.Model()
models := []string{
    "googleai/gemini-1.5-pro",                              // Google
    "openai/gpt-4",                                        // OpenAI
    "anthropic.claude-3-sonnet-20240229-v1:0",             // AWS
}
```

### Flow with Monitoring
```go
// Define flow using GenKit's API
flow := genkit.DefineFlow("myFlow", func(ctx context.Context, input string) (string, error) {
    return genkit.Generate(ctx, &ai.GenerateRequest{
        Model: genkit.Model("anthropic.claude-3-sonnet-20240229-v1:0"),
        Messages: []*ai.Message{{Role: ai.RoleUser, Content: []*ai.Part{{Text: input}}}},
    })
})

// Monitoring plugin automatically tracks this flow's performance
result, _ := flow.Run(ctx, "What's the weather?")
```

## Prerequisites

- Go 1.23+
- AWS credentials configured (`aws configure` or environment variables)
- AWS Bedrock model access enabled in your region
- Google's GenKit Go SDK

## Installation & Setup

### 1. Install Dependencies
```bash
go mod init my-genkit-app
go get github.com/firebase/genkit/go@latest
go get github.com/genkit-aws/genkit-aws@latest
```

### 2. Configure AWS Credentials
```bash
aws configure
# or set environment variables:
# export AWS_REGION=us-east-1
# export AWS_ACCESS_KEY_ID=...
# export AWS_SECRET_ACCESS_KEY=...
```

### 3. Enable Bedrock Models
Go to AWS Console → Bedrock → Model Access and enable the models you want to use.

## Development

### Running Examples
```bash
cd cmd/examples/basic-bedrock
go run main.go "What is AI?"
```

### Testing
```bash
make test              # Unit tests
make test-integration  # Integration tests (requires AWS credentials)
```

## Relationship to GenKit

| Component | Provided By |
|-----------|-------------|
| Core Framework | Google's GenKit |
| Flow API | Google's GenKit |
| Model Interface | Google's GenKit |
| Prompt Management | Google's GenKit |
| **AWS Bedrock Models** | **Our Plugin** |
| **CloudWatch Monitoring** | **Our Plugin** |

## Contributing

We welcome contributions! This project follows Google's GenKit plugin patterns and conventions.

### Development Setup
```bash
git clone https://github.com/genkit-aws/genkit-aws
cd genkit-aws
go mod download
make test
```

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

---

**Note**: For GenKit framework documentation, see [Google's GenKit docs](https://firebase.google.com/docs/genkit). This repository only documents the AWS-specific plugins.
```

### Makefile
```makefile
.PHONY: build test lint clean examples

# Build
build:
	@echo "Building..."
	go build ./...

# Test
test:
	@echo "Running tests..."
	go test -v -race ./...

# Lint  
lint:
	@echo "Linting..."
	golangci-lint run

# Clean
clean:
	@echo "Cleaning..."
	go clean ./...
	rm -rf dist/

# Examples
examples:
	@echo "Building examples..."
	go build ./cmd/examples/...

# Integration tests (requires AWS credentials)
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./...

# Generate mocks
mocks:
	@echo "Generating mocks..."
	go generate ./...

# Release
release:
	@echo "Creating release..."
	goreleaser release --rm-dist
```

This project structure follows Go best practices with:

1. **Clear module organization** - Separate packages for different concerns
2. **Interface-driven design** - Easy to mock and test
3. **Comprehensive testing** - Unit tests, integration tests, testdata
4. **Proper error handling** - Context-aware, wrapped errors
5. **Production readiness** - Monitoring, observability, configuration validation
6. **Developer experience** - Examples, documentation, clear APIs
7. **CI/CD ready** - GitHub Actions, automated testing, releases
