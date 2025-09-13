# Architecture Guide

Deep dive into how GenKit AWS plugins work and integrate with the GenKit framework.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Your GenKit Application                  │
├─────────────────────────────────────────────────────────────┤
│              Google's GenKit Framework                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐│
│  │   Flows     │ │   Models    │ │   Monitoring Hooks      ││
│  │             │ │             │ │                         ││
└──┴─────────────┘─┴─────────────┘─┴─────────────────────────┘│
   │                │                │                        │
   │                │                │                        │
┌──▼────────────────▼────────────────▼────────────────────────▼──┐
│                GenKit AWS Plugin                              │
│                                                               │
│  ┌─────────────────┐              ┌─────────────────────────┐ │
│  │ Bedrock Client  │              │ CloudWatch Monitoring   │ │
│  │                 │              │                         │ │
│  │ • Model Registry│              │ • Metric Collection     │ │
│  │ • Request Conv. │              │ • Buffering & Batching  │ │
│  │ • Response Conv.│              │ • Error Classification  │ │
│  └─────────────────┘              └─────────────────────────┘ │
└───────────────────┬─────────────────────────┬─────────────────┘
                    │                         │
                    │                         │
┌───────────────────▼─────────────────────────▼─────────────────┐
│                      AWS Services                            │
│                                                               │
│  ┌─────────────────┐              ┌─────────────────────────┐ │
│  │  AWS Bedrock    │              │    AWS CloudWatch       │ │
│  │                 │              │                         │ │
│  │ • Claude Models │              │ • Metrics Storage       │ │
│  │ • Nova Models   │              │ • Dashboards & Alarms   │ │
│  │ • Llama Models  │              │ • Log Aggregation       │ │
│  └─────────────────┘              └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Plugin Architecture

### Core Components

#### 1. Plugin Interface (`pkg/genkit-aws/plugin.go`)
```go
type Plugin struct {
    config  *Config
    bedrock *bedrock.Client
    monitor *monitoring.CloudWatch
}

// Implements api.Plugin interface
func (p *Plugin) Name() string
func (p *Plugin) Init(ctx context.Context) []api.Action
```

**Responsibilities**:
- Plugin lifecycle management
- AWS client initialization
- Model registration coordination
- Monitoring setup

#### 2. Bedrock Integration (`pkg/bedrock/`)
```go
type Client struct {
    runtime *bedrockruntime.Client
    config  *Config
}

type Model struct {
    client  *Client
    modelID string
    config  *ModelConfig
}
```

**Responsibilities**:
- AWS Bedrock API communication
- Model-specific request/response conversion
- Token usage tracking
- Error handling and retries

#### 3. CloudWatch Monitoring (`pkg/monitoring/`)
```go
type CloudWatch struct {
    client       *cloudwatch.Client
    config       *Config
    metricBuffer []types.MetricDatum
    bufferMutex  sync.Mutex
}
```

**Responsibilities**:
- Metric collection and buffering
- Automatic metric submission
- Error classification
- Resource cleanup

## Request Flow

### Model Generation Request
```
1. User calls genkit.Generate()
   ↓
2. GenKit routes to registered AWS model
   ↓  
3. Model.Generate() called
   ↓
4. Convert GenKit request → Bedrock format
   ↓
5. Call AWS Bedrock API
   ↓
6. Convert Bedrock response → GenKit format
   ↓
7. Return response to user
   ↓
8. Monitoring hooks collect metrics
```

### Metric Collection Flow
```
1. GenKit event occurs (flow start/end, generation, error)
   ↓
2. Monitoring hook called
   ↓
3. Metric added to buffer
   ↓
4. If buffer full OR timer expires:
   ↓
5. Batch metrics sent to CloudWatch
   ↓
6. Buffer cleared for next batch
```

## Model Conversion Details

### Request Conversion Strategy

Each model family has different API requirements:

#### Claude Models
```go
// GenKit → Claude Bedrock format
{
    "anthropic_version": "bedrock-2023-05-31",
    "max_tokens": 4096,
    "temperature": 0.7,
    "messages": [
        {"role": "user", "content": "Hello"}
    ]
}
```

#### Nova Models  
```go
// GenKit → Nova Bedrock format
{
    "messages": [
        {
            "role": "user", 
            "content": [{"text": "Hello"}]
        }
    ],
    "inferenceConfig": {
        "maxTokens": 4096,
        "temperature": 0.7
    }
}
```

#### Llama Models
```go
// GenKit → Llama Bedrock format
{
    "prompt": "Hello\n",
    "max_gen_len": 4096,
    "temperature": 0.7
}
```

### Response Conversion Strategy

All models convert to unified GenKit format:
```go
&ai.ModelResponse{
    Message: &ai.Message{
        Role: "model",
        Content: []*ai.Part{{Text: responseText}},
    },
    Usage: &ai.GenerationUsage{
        InputTokens:  inputTokens,
        OutputTokens: outputTokens,
        TotalTokens:  inputTokens + outputTokens,
    },
    FinishReason: "stop",
}
```

## Configuration Management

### Hierarchical Configuration
```
1. Built-in defaults (constants package)
   ↓
2. DefaultModelConfig (applies to all models)
   ↓  
3. ModelConfigs[modelID] (model-specific overrides)
   ↓
4. Runtime parameters (if provided)
```

### Configuration Validation
```go
// Validation pipeline
Plugin.New() → Config.Validate() → {
    ├── Basic validation (region, required fields)
    ├── Bedrock.Config.Validate() → ModelConfig validation
    └── CloudWatch.Config.Validate() → Namespace, buffer size
}
```

## Error Handling Strategy

### Error Types
1. **Configuration Errors**: Invalid config, missing required fields
2. **AWS Errors**: Authentication, authorization, service errors
3. **Model Errors**: Invalid model IDs, unsupported operations
4. **Network Errors**: Timeouts, connectivity issues

### Error Propagation
```go
// Errors are wrapped with context
return fmt.Errorf("failed to initialize Bedrock client: %w", err)

// Plugin initialization errors cause panics (fail-fast)
if err != nil {
    panic(fmt.Errorf("invalid genkit-aws config: %w", err))
}

// Runtime errors are returned to caller
if err != nil {
    return nil, fmt.Errorf("generation failed: %w", err)
}
```

## Performance Characteristics

### Latency
- **Plugin initialization**: ~100-500ms (one-time)
- **Model registration**: ~1-10ms per model (one-time)
- **Generation overhead**: ~5-15ms per request
- **Monitoring overhead**: ~1-5ms per metric

### Memory Usage
- **Base plugin**: ~1-5MB
- **Per model**: ~100KB-1MB
- **Metric buffering**: ~10-100KB (configurable)

### Concurrency
- **Thread-safe**: All components support concurrent access
- **Connection pooling**: AWS SDK handles connection reuse
- **Metric buffering**: Protected by mutex, async submission

## Security Model

### Credential Handling
```go
// Plugin uses standard AWS credential chain:
1. Environment variables (AWS_ACCESS_KEY_ID, etc.)
2. Shared credentials file (~/.aws/credentials)
3. IAM roles (EC2, Lambda, ECS)
4. Web identity token (EKS, Fargate)
```

### Permission Requirements
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": "arn:aws:bedrock:*::foundation-model/*"
    },
    {
      "Effect": "Allow", 
      "Action": [
        "cloudwatch:PutMetricData"
      ],
      "Resource": "*"
    }
  ]
}
```

### Data Privacy
- **No user data** stored or logged by plugin
- **Request/response content** not included in metrics
- **Only metadata** collected (tokens, duration, model ID)

## Extension Points

### Adding New Model Families
1. **Implement converter interface**:
   ```go
   func (m *Model) convertNewModelRequest(req *ai.ModelRequest) ([]byte, error)
   func (m *Model) convertNewModelResponse(body []byte) (*ai.ModelResponse, error)
   ```

2. **Add model detection**:
   ```go
   func isNewModel(modelID string) bool {
       return strings.Contains(strings.ToLower(modelID), "newmodel")
   }
   ```

3. **Update conversion dispatch**:
   ```go
   case isNewModel(m.modelID):
       return m.convertNewModelRequest(req)
   ```

### Adding New Monitoring Backends
1. **Implement Monitor interface**:
   ```go
   type Monitor interface {
       OnFlowStart(ctx context.Context, flowName string, input interface{})
       OnFlowEnd(ctx context.Context, flowName string, duration time.Duration, output interface{})
       // ... other methods
   }
   ```

2. **Add to plugin configuration**:
   ```go
   type Config struct {
       CloudWatch *monitoring.Config
       Datadog    *datadog.Config     // New backend
       Prometheus *prometheus.Config  // New backend
   }
   ```

## Testing Architecture

### Unit Tests
- **Mock AWS clients** using interfaces
- **Test conversion logic** with known inputs/outputs
- **Validate configuration** edge cases
- **Check error handling** paths

### Integration Tests
- **Real AWS services** with test accounts
- **End-to-end flows** with actual models
- **Monitoring verification** with CloudWatch
- **Performance measurement** under load

### Test Isolation
- **Separate namespaces** for test metrics
- **Cleanup after tests** to avoid state pollution
- **Parallel execution** safe with unique identifiers

---

This architecture provides a solid foundation for AWS integration while maintaining GenKit's patterns and performance characteristics.