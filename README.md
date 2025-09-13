# GenKit AWS Plugins

[![Go Report Card](https://goreportcard.com/badge/github.com/scottfriedman/genkit-aws)](https://goreportcard.com/report/github.com/scottfriedman/genkit-aws)
[![Go Reference](https://pkg.go.dev/badge/github.com/scottfriedman/genkit-aws.svg)](https://pkg.go.dev/github.com/scottfriedman/genkit-aws)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Coverage](https://img.shields.io/badge/Coverage-33.7%25-yellow.svg)](./coverage.html)

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
                    Models: []string{"anthropic.claude-3-sonnet-20240229-v1:0"},
                },
                CloudWatch: &monitoring.Config{
                    Namespace: "MyApp/GenKit",
                },
            }),
        },
    })
    if err != nil {
        log.Fatal(err)
    }

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
Our AWS Plugins (genkitaws.New())
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
err := genkit.Init(ctx, &genkit.Config{
    Plugins: []genkit.Plugin{
        genkitaws.New(&genkitaws.Config{
            Region: "us-east-1",
            Bedrock: &bedrock.Config{
                Models: []string{"anthropic.claude-3-sonnet-20240229-v1:0"},
            },
        }),
    },
})

// Use like any GenKit model
resp, _ := genkit.Generate(ctx, &ai.GenerateRequest{
    Model: genkit.Model("anthropic.claude-3-sonnet-20240229-v1:0"),
    Messages: []*ai.Message{{Role: ai.RoleUser, Content: []*ai.Part{{Text: "Hello"}}}},
})
```

### Multiple Providers
```go
// Mix Google AI, OpenAI, and AWS models in the same app
err := genkit.Init(ctx, &genkit.Config{
    Plugins: []genkit.Plugin{
        googleai.New(&googleai.Config{}),  // Google's plugin
        openai.New(&openai.Config{}),      // OpenAI plugin  
        genkitaws.New(&genkitaws.Config{  // Our AWS plugin
            Region: "us-east-1",
            Bedrock: &bedrock.Config{
                Models: []string{"anthropic.claude-3-sonnet-20240229-v1:0"},
            },
        }),
    },
})

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
    resp, err := genkit.Generate(ctx, &ai.GenerateRequest{
        Model: genkit.Model("anthropic.claude-3-sonnet-20240229-v1:0"),
        Messages: []*ai.Message{{Role: ai.RoleUser, Content: []*ai.Part{{Text: input}}}},
    })
    if err != nil {
        return "", err
    }
    return resp.Candidates[0].Message.Content[0].Text, nil
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
cd cmd/examples/bedrock-chat
go run main.go "What is AI?"

cd cmd/examples/monitoring-demo  
go run main.go
```

### Testing
```bash
make test              # Unit tests
make test-integration  # Integration tests (requires AWS credentials)
make lint             # Linting

# Integration testing with custom AWS profile/region
make test-integration-custom AWS_PROFILE=myprofile AWS_REGION=us-east-1

# Standalone integration test script
./scripts/integration-test.sh aws us-west-2    # Custom profile and region
./scripts/integration-test.sh                 # Use defaults (aws profile, us-west-2)
```

### Integration Testing

The project includes comprehensive integration tests that verify real AWS Bedrock and CloudWatch functionality:

#### **Quick Integration Test**
```bash
# Test with your default AWS profile and us-west-2 region
./scripts/integration-test.sh aws us-west-2
```

#### **Prerequisites for Integration Tests**
- AWS CLI configured with valid credentials
- AWS Bedrock model access enabled in your target region
- Sufficient AWS permissions for Bedrock and CloudWatch

#### **What Integration Tests Verify**
- ✅ AWS authentication and configuration
- ✅ Bedrock model availability and access
- ✅ Model generation with real Claude/Nova models
- ✅ CloudWatch metrics submission and buffering
- ✅ Error handling with invalid models
- ✅ Flow monitoring and instrumentation

#### **Integration Test Options**
```bash
# Using Go test directly
AWS_PROFILE=aws AWS_REGION=us-west-2 go test -tags=integration ./test/integration/

# Using Makefile
make test-integration                                    # Uses aws profile, us-west-2
make test-integration-custom AWS_PROFILE=prod AWS_REGION=us-east-1

# Using standalone script  
./scripts/integration-test.sh aws us-west-2 10m         # Custom timeout
```

**⚠️ Note**: Integration tests will incur small AWS charges for Bedrock API calls and CloudWatch metrics.

## Configuration

### Bedrock Configuration
```go
&bedrock.Config{
    Models: []string{
        "anthropic.claude-3-sonnet-20240229-v1:0",
        "amazon.nova-pro-v1:0",
    },
    DefaultModelConfig: &bedrock.ModelConfig{
        MaxTokens:     4096,
        Temperature:   0.7,
        TopP:          0.9,
        StopSequences: []string{"STOP"},
    },
    ModelConfigs: map[string]*bedrock.ModelConfig{
        "anthropic.claude-3-sonnet-20240229-v1:0": {
            MaxTokens:   8192,
            Temperature: 0.5,
        },
    },
}
```

### CloudWatch Configuration
```go
&monitoring.Config{
    Namespace:          "MyApp/GenKit",
    EnableFlowMetrics:  true,
    EnableModelMetrics: true,
    CustomDimensions: map[string]string{
        "Environment": "Production",
        "Version":     "1.0.0",
    },
    MetricBufferSize: 20,
}
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

### Submitting Issues
Please use GitHub Issues for bug reports and feature requests.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

---

**Note**: For GenKit framework documentation, see [Google's GenKit docs](https://firebase.google.com/docs/genkit). This repository only documents the AWS-specific plugins.