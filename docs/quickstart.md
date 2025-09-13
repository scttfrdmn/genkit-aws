# GenKit AWS Quick Start Guide

Get up and running with AWS Bedrock models in your GenKit application in under 5 minutes.

## Prerequisites

- **Go 1.23+** installed
- **AWS CLI** configured (`aws configure`)
- **AWS Bedrock access** enabled in your region
- **GenKit knowledge** (see [GenKit docs](https://genkit.dev))

## Step 1: Install Dependencies

```bash
go mod init my-genkit-app
go get github.com/firebase/genkit/go@latest
go get github.com/scttfrdmn/genkit-aws@latest
```

## Step 2: Enable AWS Bedrock Models

Visit [AWS Console → Bedrock → Model Access](https://console.aws.amazon.com/bedrock/home#/modelaccess) and enable:
- ✅ Anthropic Claude 3 Sonnet
- ✅ Amazon Nova Pro
- ✅ Amazon Nova Lite

## Step 3: Basic Usage

Create `main.go`:

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

    // Create AWS plugin
    awsPlugin, err := genkitaws.New(&genkitaws.Config{
        Region: "us-east-1", // Your AWS region
        Bedrock: &bedrock.Config{
            Models: []string{
                "anthropic.claude-3-sonnet-20240229-v1:0",
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    // Initialize GenKit with AWS plugin
    g := genkit.Init(ctx, genkit.WithPlugins(awsPlugin))

    // Define the model
    awsPlugin.DefineModel(g, "anthropic.claude-3-sonnet-20240229-v1:0", nil)

    // Generate text
    resp, err := genkit.Generate(ctx, g,
        ai.WithModelName("anthropic.claude-3-sonnet-20240229-v1:0"),
        ai.WithMessages(ai.NewUserTextMessage("What is GenKit?")),
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Response:", resp.Text())
}
```

## Step 4: Run Your Application

```bash
go run main.go
```

**Expected output:**
```
Response: GenKit is an open-source AI development framework created by Google...
```

## Step 5: Add Monitoring (Optional)

```go
awsPlugin, err := genkitaws.New(&genkitaws.Config{
    Region: "us-east-1",
    Bedrock: &bedrock.Config{
        Models: []string{"anthropic.claude-3-sonnet-20240229-v1:0"},
    },
    CloudWatch: &monitoring.Config{
        Namespace: "MyApp/GenKit",
        EnableFlowMetrics: true,
        EnableModelMetrics: true,
    },
})
```

View metrics in **AWS Console → CloudWatch → Metrics → MyApp/GenKit**.

## Troubleshooting

### ❌ "region is required"
**Solution**: Add region to your config:
```go
&genkitaws.Config{
    Region: "us-east-1", // Add this
}
```

### ❌ "model not found"
**Solution**: Define the model after initializing GenKit:
```go
g := genkit.Init(ctx, genkit.WithPlugins(awsPlugin))
awsPlugin.DefineModel(g, "anthropic.claude-3-sonnet-20240229-v1:0", nil) // Add this
```

### ❌ "operation error Bedrock Runtime"
**Solutions**:
1. Enable model access in AWS Console → Bedrock → Model Access
2. Check your AWS credentials: `aws sts get-caller-identity`
3. Verify region supports Bedrock: use `us-east-1` or `us-west-2`

### ❌ "no AWS config sources"
**Solutions**:
1. Run `aws configure` to set up credentials
2. Set environment variables:
   ```bash
   export AWS_REGION=us-east-1
   export AWS_ACCESS_KEY_ID=your_key
   export AWS_SECRET_ACCESS_KEY=your_secret
   ```

## Next Steps

- 📖 **[Configuration Guide](./configuration.md)** - Advanced configuration options
- 🏗️ **[Architecture Guide](./architecture.md)** - How the plugin works
- 🧪 **[Integration Testing](./integration-testing.md)** - Test with real AWS
- 📊 **[Monitoring Setup](./monitoring-setup.md)** - CloudWatch configuration
- 📝 **[Examples](../cmd/examples/)** - Complete example applications

## Need Help?

- 🐛 **[Report Issues](https://github.com/scttfrdmn/genkit-aws/issues/new/choose)**
- ❓ **[Ask Questions](https://github.com/scttfrdmn/genkit-aws/discussions)**
- 📚 **[GenKit Documentation](https://genkit.dev)** 
- 🔧 **[AWS Bedrock Docs](https://docs.aws.amazon.com/bedrock/)**