# GenKit AWS Examples

Complete examples demonstrating various use cases and integration patterns.

## Available Examples

### 1. [Basic Chat Bot](../../cmd/examples/bedrock-chat/)
**File**: `cmd/examples/bedrock-chat/main.go`
**What it demonstrates**:
- Basic GenKit AWS plugin setup
- Claude 3 Sonnet model usage
- Simple chat flow implementation
- Command-line argument handling

**Run it**:
```bash
cd cmd/examples/bedrock-chat
go run main.go "What is artificial intelligence?"
```

### 2. [Monitoring Demo](../../cmd/examples/monitoring-demo/)
**File**: `cmd/examples/monitoring-demo/main.go`
**What it demonstrates**:
- CloudWatch monitoring setup
- Multiple flow definitions
- Error simulation and metric collection
- Custom dimensions and namespaces

**Run it**:
```bash
cd cmd/examples/monitoring-demo
go run main.go
```

## Tutorial Examples

### Example 1: Multi-Model Application

Create an application that uses different models for different tasks:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/firebase/genkit/go/ai" 
    "github.com/firebase/genkit/go/genkit"
    genkitaws "github.com/scttfrdmn/genkit-aws/pkg/genkit-aws"
    "github.com/scttfrdmn/genkit-aws/pkg/bedrock"
)

func main() {
    ctx := context.Background()

    // Configure multiple models for different purposes
    awsPlugin, err := genkitaws.New(&genkitaws.Config{
        Region: "us-east-1",
        Bedrock: &bedrock.Config{
            Models: []string{
                "anthropic.claude-3-sonnet-20240229-v1:0", // Complex reasoning
                "amazon.nova-lite-v1:0",                   // Fast responses
                "meta.llama3-2-11b-instruct-v1:0",        // Code generation
            },
            ModelConfigs: map[string]*bedrock.ModelConfig{
                "anthropic.claude-3-sonnet-20240229-v1:0": {
                    MaxTokens:   4096,
                    Temperature: 0.3, // More deterministic
                },
                "amazon.nova-lite-v1:0": {
                    MaxTokens:   500,  // Quick responses
                    Temperature: 0.8,
                },
                "meta.llama3-2-11b-instruct-v1:0": {
                    MaxTokens:   2048,
                    Temperature: 0.1, // Very deterministic for code
                },
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    g := genkit.Init(ctx, genkit.WithPlugins(awsPlugin))

    // Define models
    claude := awsPlugin.DefineModel(g, "anthropic.claude-3-sonnet-20240229-v1:0", &ai.ModelOptions{
        Label: "Claude 3 Sonnet - Reasoning",
    })
    nova := awsPlugin.DefineModel(g, "amazon.nova-lite-v1:0", &ai.ModelOptions{
        Label: "Nova Lite - Fast",
    })
    llama := awsPlugin.DefineModel(g, "meta.llama3-2-11b-instruct-v1:0", &ai.ModelOptions{
        Label: "Llama 3.2 - Code",
    })

    // Define specialized flows
    analyzeFlow := genkit.DefineFlow(g, "analyze", func(ctx context.Context, text string) (string, error) {
        return genkit.GenerateText(ctx, g,
            ai.WithModel(claude),
            ai.WithMessages(ai.NewUserTextMessage(
                fmt.Sprintf("Analyze this text for key themes and sentiment: %s", text),
            )),
        )
    })

    quickResponseFlow := genkit.DefineFlow(g, "quickResponse", func(ctx context.Context, question string) (string, error) {
        return genkit.GenerateText(ctx, g,
            ai.WithModel(nova),
            ai.WithMessages(ai.NewUserTextMessage(question)),
        )
    })

    codeFlow := genkit.DefineFlow(g, "generateCode", func(ctx context.Context, description string) (string, error) {
        return genkit.GenerateText(ctx, g,
            ai.WithModel(llama),
            ai.WithMessages(ai.NewUserTextMessage(
                fmt.Sprintf("Generate Go code for: %s", description),
            )),
        )
    })

    // Use the flows
    analysis, _ := analyzeFlow.Run(ctx, "This product is amazing but expensive")
    fmt.Println("Analysis:", analysis)

    response, _ := quickResponseFlow.Run(ctx, "What's 2+2?")
    fmt.Println("Quick response:", response)

    code, _ := codeFlow.Run(ctx, "a function that reverses a string")
    fmt.Println("Generated code:", code)
}
```

### Example 2: Production Application with Monitoring

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/genkit"
    genkitaws "github.com/scttfrdmn/genkit-aws/pkg/genkit-aws"
    "github.com/scttfrdmn/genkit-aws/pkg/bedrock"
    "github.com/scttfrdmn/genkit-aws/pkg/monitoring"
)

func main() {
    ctx := context.Background()
    
    // Production-ready configuration
    awsPlugin, err := genkitaws.New(&genkitaws.Config{
        Region:  os.Getenv("AWS_REGION"),
        Profile: os.Getenv("AWS_PROFILE"), // Optional: use specific profile
        Bedrock: &bedrock.Config{
            Models: []string{
                "anthropic.claude-3-sonnet-20240229-v1:0",
                "amazon.nova-pro-v1:0",
            },
            DefaultModelConfig: &bedrock.ModelConfig{
                MaxTokens:     4096,
                Temperature:   0.7,
                TopP:          0.9,
                StopSequences: []string{"STOP", "END"},
            },
        },
        CloudWatch: &monitoring.Config{
            Namespace:          fmt.Sprintf("MyApp-%s/GenKit", os.Getenv("ENVIRONMENT")),
            EnableFlowMetrics:  true,
            EnableModelMetrics: true,
            CustomDimensions: map[string]string{
                "Environment": os.Getenv("ENVIRONMENT"),
                "Version":     os.Getenv("APP_VERSION"),
                "Region":      os.Getenv("AWS_REGION"),
            },
            MetricBufferSize: 50, // Larger buffer for production
        },
    })
    if err != nil {
        log.Fatalf("Failed to initialize AWS plugin: %v", err)
    }

    g := genkit.Init(ctx, genkit.WithPlugins(awsPlugin))

    // Define models with production settings
    awsPlugin.DefineModel(g, "anthropic.claude-3-sonnet-20240229-v1:0", &ai.ModelOptions{
        Label: "Claude 3 Sonnet - Production",
    })

    // Production flow with error handling
    chatFlow := genkit.DefineFlow(g, "productionChat", func(ctx context.Context, input string) (string, error) {
        resp, err := genkit.Generate(ctx, g,
            ai.WithModelName("anthropic.claude-3-sonnet-20240229-v1:0"),
            ai.WithMessages(ai.NewUserTextMessage(input)),
        )
        if err != nil {
            // Log error for debugging
            log.Printf("Generation failed for input '%s': %v", input, err)
            return "", fmt.Errorf("failed to generate response: %w", err)
        }

        if resp.Message == nil || len(resp.Message.Content) == 0 {
            return "", fmt.Errorf("empty response from model")
        }

        // Log success metrics
        log.Printf("Generated %d tokens for input length %d", 
                   resp.Usage.OutputTokens, len(input))

        return resp.Message.Content[0].Text, nil
    })

    // Example usage
    result, err := chatFlow.Run(ctx, "Explain quantum computing")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Result:", result)
    
    // Graceful shutdown - flush metrics
    if monitor := awsPlugin.GetMonitor(); monitor != nil {
        if err := monitor.Close(ctx); err != nil {
            log.Printf("Warning: failed to close monitor: %v", err)
        }
    }
}
```

### Example 3: Error Handling and Retries

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/genkit"
    genkitaws "github.com/scttfrdmn/genkit-aws/pkg/genkit-aws"
    "github.com/scttfrdmn/genkit-aws/pkg/bedrock"
)

// RetryableFlow demonstrates robust error handling
func createRetryableFlow(g *genkit.Genkit, modelName string) *genkit.Flow {
    return genkit.DefineFlow(g, "retryableGeneration", func(ctx context.Context, input string) (string, error) {
        const maxRetries = 3
        var lastErr error

        for attempt := 1; attempt <= maxRetries; attempt++ {
            resp, err := genkit.Generate(ctx, g,
                ai.WithModelName(modelName),
                ai.WithMessages(ai.NewUserTextMessage(input)),
            )
            
            if err == nil && resp.Message != nil && len(resp.Message.Content) > 0 {
                return resp.Message.Content[0].Text, nil
            }
            
            lastErr = err
            log.Printf("Attempt %d failed: %v", attempt, err)
            
            if attempt < maxRetries {
                // Exponential backoff
                delay := time.Duration(attempt) * time.Second
                log.Printf("Retrying in %v...", delay)
                time.Sleep(delay)
            }
        }
        
        return "", fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
    })
}

func main() {
    ctx := context.Background()

    awsPlugin, err := genkitaws.New(&genkitaws.Config{
        Region: "us-east-1",
        Bedrock: &bedrock.Config{
            Models: []string{"anthropic.claude-3-sonnet-20240229-v1:0"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    g := genkit.Init(ctx, genkit.WithPlugins(awsPlugin))
    awsPlugin.DefineModel(g, "anthropic.claude-3-sonnet-20240229-v1:0", nil)

    // Create robust flow
    flow := createRetryableFlow(g, "anthropic.claude-3-sonnet-20240229-v1:0")

    // Test with various inputs
    inputs := []string{
        "What is machine learning?",
        "Explain blockchain technology",
        "Invalid very long prompt that might cause issues...",
    }

    for _, input := range inputs {
        fmt.Printf("\nTesting: %s\n", input)
        result, err := flow.Run(ctx, input)
        if err != nil {
            fmt.Printf("❌ Failed: %v\n", err)
        } else {
            fmt.Printf("✅ Success: %s\n", result[:min(100, len(result))])
        }
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

## Running Examples

### Prerequisites
```bash
# Ensure AWS credentials are configured
aws configure

# Or set environment variables
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=your_key
export AWS_SECRET_ACCESS_KEY=your_secret
```

### Basic Examples
```bash
# Chat bot example
cd cmd/examples/bedrock-chat
go run main.go "Tell me about AI"

# Monitoring demo
cd cmd/examples/monitoring-demo  
go run main.go
```

### Custom Examples
```bash
# Create your own example
mkdir my-example
cd my-example
go mod init my-example
go get github.com/firebase/genkit/go@latest
go get github.com/scttfrdmn/genkit-aws@latest

# Copy and modify one of the examples above
```

## Common Patterns

### 1. Model Fallback Chain
```go
func generateWithFallback(ctx context.Context, g *genkit.Genkit, input string) (string, error) {
    models := []string{
        "anthropic.claude-3-sonnet-20240229-v1:0", // Primary
        "amazon.nova-pro-v1:0",                    // Fallback 1
        "amazon.nova-lite-v1:0",                   // Fallback 2
    }
    
    for _, model := range models {
        resp, err := genkit.Generate(ctx, g,
            ai.WithModelName(model),
            ai.WithMessages(ai.NewUserTextMessage(input)),
        )
        if err == nil {
            return resp.Text(), nil
        }
        log.Printf("Model %s failed: %v", model, err)
    }
    
    return "", fmt.Errorf("all models failed")
}
```

### 2. Streaming Responses
```go
func streamingGeneration(ctx context.Context, g *genkit.Genkit, input string) error {
    return genkit.Generate(ctx, g,
        ai.WithModelName("anthropic.claude-3-sonnet-20240229-v1:0"),
        ai.WithMessages(ai.NewUserTextMessage(input)),
        ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
            if len(chunk.Content) > 0 {
                fmt.Print(chunk.Content[0].Text)
            }
            return nil
        }),
    )
}
```

### 3. Configuration from Environment
```go
func createConfigFromEnv() *genkitaws.Config {
    return &genkitaws.Config{
        Region:  os.Getenv("AWS_REGION"),
        Profile: os.Getenv("AWS_PROFILE"),
        Bedrock: &bedrock.Config{
            Models: strings.Split(os.Getenv("BEDROCK_MODELS"), ","),
            DefaultModelConfig: &bedrock.ModelConfig{
                MaxTokens:   getEnvInt("MAX_TOKENS", 4096),
                Temperature: getEnvFloat("TEMPERATURE", 0.7),
            },
        },
        CloudWatch: &monitoring.Config{
            Namespace:          fmt.Sprintf("%s/GenKit", os.Getenv("APP_NAME")),
            EnableFlowMetrics:  getEnvBool("ENABLE_FLOW_METRICS", true),
            EnableModelMetrics: getEnvBool("ENABLE_MODEL_METRICS", true),
        },
    }
}
```

## Performance Examples

### 1. Concurrent Generations
```go
func concurrentGeneration(ctx context.Context, g *genkit.Genkit, inputs []string) []string {
    results := make([]string, len(inputs))
    var wg sync.WaitGroup
    
    for i, input := range inputs {
        wg.Add(1)
        go func(idx int, prompt string) {
            defer wg.Done()
            
            resp, err := genkit.Generate(ctx, g,
                ai.WithModelName("amazon.nova-lite-v1:0"), // Fast model
                ai.WithMessages(ai.NewUserTextMessage(prompt)),
            )
            if err != nil {
                results[idx] = fmt.Sprintf("Error: %v", err)
                return
            }
            results[idx] = resp.Text()
        }(i, input)
    }
    
    wg.Wait()
    return results
}
```

### 2. Response Caching
```go
type CachedGenerator struct {
    cache map[string]string
    mutex sync.RWMutex
    flow  *genkit.Flow
}

func (cg *CachedGenerator) Generate(ctx context.Context, input string) (string, error) {
    // Check cache first
    cg.mutex.RLock()
    if cached, exists := cg.cache[input]; exists {
        cg.mutex.RUnlock()
        return cached, nil
    }
    cg.mutex.RUnlock()
    
    // Generate new response
    result, err := cg.flow.Run(ctx, input)
    if err != nil {
        return "", err
    }
    
    // Cache result
    cg.mutex.Lock()
    cg.cache[input] = result
    cg.mutex.Unlock()
    
    return result, nil
}
```

## Deployment Examples

### 1. Docker Container
```dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o genkit-app ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/genkit-app .

# Environment configuration
ENV AWS_REGION=us-east-1
ENV BEDROCK_MODELS=anthropic.claude-3-sonnet-20240229-v1:0

CMD ["./genkit-app"]
```

### 2. AWS Lambda
```go
package main

import (
    "context"
    "encoding/json"
    
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/genkit"
    genkitaws "github.com/scttfrdmn/genkit-aws/pkg/genkit-aws"
)

type Request struct {
    Prompt string `json:"prompt"`
    Model  string `json:"model,omitempty"`
}

type Response struct {
    Result string `json:"result"`
    Usage  *ai.GenerationUsage `json:"usage,omitempty"`
}

var g *genkit.Genkit
var awsPlugin *genkitaws.Plugin

func init() {
    ctx := context.Background()
    
    var err error
    awsPlugin, err = genkitaws.New(&genkitaws.Config{
        Region: os.Getenv("AWS_REGION"),
        Bedrock: &bedrock.Config{
            Models: []string{
                "anthropic.claude-3-sonnet-20240229-v1:0",
                "amazon.nova-pro-v1:0",
            },
        },
        CloudWatch: &monitoring.Config{
            Namespace: "Lambda/GenKit",
            EnableFlowMetrics: true,
            EnableModelMetrics: true,
        },
    })
    if err != nil {
        panic(err)
    }
    
    g = genkit.Init(ctx, genkit.WithPlugins(awsPlugin))
    awsPlugin.DefineModel(g, "anthropic.claude-3-sonnet-20240229-v1:0", nil)
    awsPlugin.DefineModel(g, "amazon.nova-pro-v1:0", nil)
}

func handleRequest(ctx context.Context, req Request) (Response, error) {
    modelName := req.Model
    if modelName == "" {
        modelName = "anthropic.claude-3-sonnet-20240229-v1:0"
    }
    
    resp, err := genkit.Generate(ctx, g,
        ai.WithModelName(modelName),
        ai.WithMessages(ai.NewUserTextMessage(req.Prompt)),
    )
    if err != nil {
        return Response{}, err
    }
    
    return Response{
        Result: resp.Text(),
        Usage:  resp.Usage,
    }, nil
}

func main() {
    lambda.Start(handleRequest)
}
```

## Testing Examples

### 1. Example Tests
```go
func TestExampleIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    ctx := context.Background()
    awsPlugin, err := genkitaws.New(&genkitaws.Config{
        Region: "us-east-1",
        Bedrock: &bedrock.Config{
            Models: []string{"anthropic.claude-3-sonnet-20240229-v1:0"},
        },
    })
    require.NoError(t, err)
    
    g := genkit.Init(ctx, genkit.WithPlugins(awsPlugin))
    awsPlugin.DefineModel(g, "anthropic.claude-3-sonnet-20240229-v1:0", nil)
    
    // Your test logic here
}
```

## Best Practices

1. **Environment Configuration**: Use environment variables for deployment-specific settings
2. **Error Handling**: Always handle generation errors gracefully  
3. **Resource Cleanup**: Close monitoring connections when shutting down
4. **Cost Monitoring**: Set appropriate token limits for your use case
5. **Security**: Never log or metric sensitive user data

---

💡 **Need more examples?** Check out the [cmd/examples/](../../cmd/examples/) directory or [open an issue](https://github.com/scttfrdmn/genkit-aws/issues) requesting specific examples!