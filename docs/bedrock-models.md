# Supported AWS Bedrock Models

Complete list of AWS Bedrock foundation models supported by GenKit AWS plugins.

## Anthropic Claude Models

### Claude 3.5 Series
| Model ID | Model Name | Context Length | Best For |
|----------|------------|----------------|----------|
| `anthropic.claude-3-5-sonnet-20241022-v2:0` | Claude 3.5 Sonnet v2 | 200K | Balanced performance and capability |
| `anthropic.claude-3-5-sonnet-20240620-v1:0` | Claude 3.5 Sonnet v1 | 200K | Complex reasoning tasks |
| `anthropic.claude-3-5-haiku-20241022-v1:0` | Claude 3.5 Haiku | 200K | Fast, cost-effective |

### Claude 3 Series  
| Model ID | Model Name | Context Length | Best For |
|----------|------------|----------------|----------|
| `anthropic.claude-3-sonnet-20240229-v1:0` | Claude 3 Sonnet | 200K | Balanced tasks |
| `anthropic.claude-3-haiku-20240307-v1:0` | Claude 3 Haiku | 200K | Speed and efficiency |
| `anthropic.claude-3-opus-20240229-v1:0` | Claude 3 Opus | 200K | Complex analysis |

## Amazon Nova Models

### Nova Series (Latest)
| Model ID | Model Name | Context Length | Best For |
|----------|------------|----------------|----------|
| `amazon.nova-pro-v1:0` | Nova Pro | 300K | High-quality text generation |
| `amazon.nova-lite-v1:0` | Nova Lite | 300K | Fast, cost-effective |
| `amazon.nova-micro-v1:0` | Nova Micro | 128K | Ultra-fast responses |

## Meta Llama Models

### Llama 3.2 Series
| Model ID | Model Name | Context Length | Best For |
|----------|------------|----------------|----------|
| `meta.llama3-2-90b-instruct-v1:0` | Llama 3.2 90B | 128K | Large-scale reasoning |
| `meta.llama3-2-11b-instruct-v1:0` | Llama 3.2 11B | 128K | Efficient instruction following |
| `meta.llama3-2-3b-instruct-v1:0` | Llama 3.2 3B | 128K | Lightweight applications |
| `meta.llama3-2-1b-instruct-v1:0` | Llama 3.2 1B | 128K | Edge and mobile |

### Llama 3.1 Series
| Model ID | Model Name | Context Length | Best For |
|----------|------------|----------------|----------|
| `meta.llama3-1-405b-instruct-v1:0` | Llama 3.1 405B | 128K | Maximum capability |
| `meta.llama3-1-70b-instruct-v1:0` | Llama 3.1 70B | 128K | High performance |
| `meta.llama3-1-8b-instruct-v1:0` | Llama 3.1 8B | 128K | Balanced efficiency |

## AI21 Labs Models

### Jamba Series
| Model ID | Model Name | Context Length | Best For |
|----------|------------|----------------|----------|
| `ai21.jamba-1-5-large-v1:0` | Jamba 1.5 Large | 256K | Long context tasks |
| `ai21.jamba-1-5-mini-v1:0` | Jamba 1.5 Mini | 256K | Efficient processing |

## Cohere Models

### Command Series
| Model ID | Model Name | Context Length | Best For |
|----------|------------|----------------|----------|
| `cohere.command-r-plus-v1:0` | Command R+ | 128K | RAG and tool use |
| `cohere.command-r-v1:0` | Command R | 128K | Conversational AI |

## Mistral AI Models

### Mistral Series
| Model ID | Model Name | Context Length | Best For |
|----------|------------|----------------|----------|
| `mistral.mistral-7b-instruct-v0:2` | Mistral 7B | 32K | Efficient multilingual |
| `mistral.mixtral-8x7b-instruct-v0:1` | Mixtral 8x7B | 32K | High performance |
| `mistral.mistral-large-2402-v1:0` | Mistral Large | 32K | Complex reasoning |

## Model Configuration Examples

### Basic Model Setup
```go
&bedrock.Config{
    Models: []string{
        "anthropic.claude-3-sonnet-20240229-v1:0",
        "amazon.nova-pro-v1:0",
    },
}
```

### Per-Model Configuration
```go
&bedrock.Config{
    Models: []string{
        "anthropic.claude-3-sonnet-20240229-v1:0",
        "amazon.nova-lite-v1:0",
    },
    ModelConfigs: map[string]*bedrock.ModelConfig{
        "anthropic.claude-3-sonnet-20240229-v1:0": {
            MaxTokens:   8192,
            Temperature: 0.3, // More deterministic
            TopP:        0.8,
        },
        "amazon.nova-lite-v1:0": {
            MaxTokens:   2048,
            Temperature: 0.9, // More creative
            TopP:        0.95,
        },
    },
}
```

### Default Configuration for All Models
```go
&bedrock.Config{
    Models: []string{
        "anthropic.claude-3-sonnet-20240229-v1:0",
        "amazon.nova-pro-v1:0",
        "meta.llama3-2-11b-instruct-v1:0",
    },
    DefaultModelConfig: &bedrock.ModelConfig{
        MaxTokens:     4096,
        Temperature:   0.7,
        TopP:          0.9,
        StopSequences: []string{"STOP", "END"},
    },
}
```

## Regional Availability

### US Regions
- **us-east-1** (Virginia) - ✅ All models
- **us-west-2** (Oregon) - ✅ Most models

### EU Regions  
- **eu-west-1** (Ireland) - ✅ Select models
- **eu-central-1** (Frankfurt) - ✅ Select models

### Other Regions
- **ap-southeast-1** (Singapore) - ✅ Limited models
- **ap-northeast-1** (Tokyo) - ✅ Limited models

> **Note**: Model availability varies by region. Check [AWS Bedrock documentation](https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html) for current availability.

## Pricing Considerations

### Input vs Output Tokens
- **Input tokens**: Your prompt and context
- **Output tokens**: Generated response
- **Pricing varies** significantly between models

### Cost Optimization Tips
1. **Use smaller models** for simple tasks (Nova Lite, Claude Haiku)
2. **Set MaxTokens** appropriately to avoid over-generation
3. **Monitor usage** with CloudWatch metrics
4. **Cache responses** when possible

### Example Pricing (Approximate)
| Model Family | Input (per 1K tokens) | Output (per 1K tokens) |
|-------------|-------------------|-------------------|
| Claude 3 Haiku | $0.00025 | $0.00125 |
| Claude 3 Sonnet | $0.003 | $0.015 |
| Nova Lite | $0.0002 | $0.0008 |
| Nova Pro | $0.0008 | $0.0032 |

> **Note**: Prices are estimates. Check [AWS Bedrock pricing](https://aws.amazon.com/bedrock/pricing/) for current rates.

## Best Practices

### Model Selection
- **Claude Sonnet**: Balanced reasoning and speed
- **Nova Pro**: Cost-effective with good quality
- **Nova Lite**: Fastest and cheapest
- **Llama**: Open source, good for specialized tasks

### Configuration
- **Temperature 0.3-0.5**: Factual, deterministic responses
- **Temperature 0.7-0.9**: Creative, varied responses  
- **MaxTokens**: Set based on expected response length
- **TopP 0.8-0.95**: Good balance for most use cases

### Error Handling
```go
resp, err := genkit.Generate(ctx, g,
    ai.WithModelName("anthropic.claude-3-sonnet-20240229-v1:0"),
    ai.WithMessages(ai.NewUserTextMessage("Hello")),
)
if err != nil {
    // Handle errors appropriately
    log.Printf("Generation failed: %v", err)
    return
}
```

## Next Steps

- 📖 **[Configuration Guide](./configuration.md)** - Advanced options
- 🏗️ **[Architecture Guide](./architecture.md)** - How it works  
- 🧪 **[Integration Testing](./integration-testing.md)** - Test with real AWS
- 📊 **[Monitoring Setup](./monitoring-setup.md)** - CloudWatch metrics