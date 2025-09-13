# CloudWatch Monitoring Setup

Complete guide for setting up AWS CloudWatch monitoring with GenKit AWS plugins.

## Overview

The GenKit AWS plugin automatically collects and sends metrics to CloudWatch for:
- **Flow Performance**: Duration, success rates, error types
- **Model Usage**: Token consumption, generation latency
- **Custom Metrics**: Configurable dimensions and namespaces

## Basic Setup

```go
awsPlugin, err := genkitaws.New(&genkitaws.Config{
    Region: "us-east-1",
    CloudWatch: &monitoring.Config{
        Namespace: "MyApp/GenKit",
        EnableFlowMetrics: true,
        EnableModelMetrics: true,
    },
})
```

## Configuration Options

### Namespace Organization
```go
&monitoring.Config{
    // Recommended: App/Component format
    Namespace: "MyApp/GenKit",
    
    // For multiple environments:
    Namespace: "MyApp-Production/GenKit",
    Namespace: "MyApp-Staging/GenKit",
}
```

### Metric Types
```go
&monitoring.Config{
    EnableFlowMetrics:  true, // Flow duration, success/error rates
    EnableModelMetrics: true, // Token usage, generation latency
    EnableXRayTracing:  true, // X-Ray distributed tracing (future)
}
```

### Custom Dimensions
```go
&monitoring.Config{
    Namespace: "MyApp/GenKit",
    CustomDimensions: map[string]string{
        "Environment": "Production",
        "Version":     "1.2.3", 
        "Region":      "us-east-1",
        "Team":        "AI-Platform",
    },
}
```

### Buffering Configuration
```go
&monitoring.Config{
    MetricBufferSize: 20, // Metrics to buffer before sending
    // Metrics are also flushed every 30 seconds automatically
}
```

## Available Metrics

### Flow Metrics

| Metric Name | Type | Description | Dimensions |
|------------|------|-------------|------------|
| `FlowStarted` | Count | Number of flows started | FlowName |
| `FlowCompleted` | Count | Number of flows completed successfully | FlowName, Status=Success |
| `FlowError` | Count | Number of flows that failed | FlowName, Status=Error, ErrorType |
| `FlowDuration` | Duration | Time taken for flow execution (ms) | FlowName, Status |

### Model Metrics

| Metric Name | Type | Description | Dimensions |
|------------|------|-------------|------------|
| `TokensUsed` | Count | Total tokens consumed | ModelID |
| `GenerationDuration` | Duration | Model generation time (ms) | ModelID |
| `GenerationCount` | Count | Number of generation requests | ModelID |

### Error Classification

Error types automatically detected:
- **Timeout**: Network or processing timeouts
- **Connection**: Network connectivity issues
- **Authentication**: AWS credential problems
- **Throttling**: AWS rate limiting
- **RateLimit**: API rate limits exceeded
- **GenericError**: Other unclassified errors

## Viewing Metrics in AWS Console

### 1. Navigate to CloudWatch
1. Open [AWS Console](https://console.aws.amazon.com)
2. Go to **CloudWatch** service
3. Click **Metrics** → **All metrics**

### 2. Find Your Namespace
1. Look for your namespace (e.g., `MyApp/GenKit`)
2. Explore available metrics by dimension

### 3. Create Dashboards
```json
{
  "widgets": [
    {
      "type": "metric",
      "properties": {
        "metrics": [
          ["MyApp/GenKit", "FlowCompleted", "FlowName", "chatFlow"],
          [".", "FlowError", ".", "."]
        ],
        "period": 300,
        "stat": "Sum",
        "region": "us-east-1",
        "title": "Flow Success/Error Rate"
      }
    }
  ]
}
```

## Advanced Configuration

### Environment-Specific Setup
```go
func createMonitoringConfig(env string) *monitoring.Config {
    config := &monitoring.Config{
        Namespace:          fmt.Sprintf("MyApp-%s/GenKit", env),
        EnableFlowMetrics:  true,
        EnableModelMetrics: true,
        CustomDimensions: map[string]string{
            "Environment": env,
            "GitCommit":   os.Getenv("GIT_COMMIT"),
            "BuildTime":   os.Getenv("BUILD_TIME"),
        },
    }
    
    // Production gets more detailed monitoring
    if env == "production" {
        config.MetricBufferSize = 50
        config.CustomDimensions["AlertingEnabled"] = "true"
    }
    
    return config
}
```

### Performance Monitoring
```go
// Monitor model performance across different configurations
configs := map[string]*bedrock.ModelConfig{
    "fast": {MaxTokens: 100, Temperature: 0.3},
    "balanced": {MaxTokens: 1000, Temperature: 0.7},
    "creative": {MaxTokens: 2000, Temperature: 0.9},
}

for name, config := range configs {
    plugin.DefineModel(g, fmt.Sprintf("claude-sonnet-%s", name), &ai.ModelOptions{
        Label: fmt.Sprintf("Claude Sonnet (%s)", name),
    })
    
    // Each gets separate metrics by ModelID
}
```

## CloudWatch Alarms

### High Error Rate Alert
```bash
aws cloudwatch put-metric-alarm \
  --alarm-name "GenKit-High-Error-Rate" \
  --alarm-description "GenKit flow error rate > 5%" \
  --metric-name "FlowError" \
  --namespace "MyApp/GenKit" \
  --statistic "Sum" \
  --period 300 \
  --evaluation-periods 2 \
  --threshold 5 \
  --comparison-operator "GreaterThanThreshold"
```

### High Token Usage Alert
```bash
aws cloudwatch put-metric-alarm \
  --alarm-name "GenKit-High-Token-Usage" \
  --alarm-description "Token usage exceeds budget" \
  --metric-name "TokensUsed" \
  --namespace "MyApp/GenKit" \
  --statistic "Sum" \
  --period 3600 \
  --evaluation-periods 1 \
  --threshold 1000000 \
  --comparison-operator "GreaterThanThreshold"
```

## Cost Monitoring

### Track Costs by Dimension
Monitor costs using custom dimensions:

```go
&monitoring.Config{
    CustomDimensions: map[string]string{
        "CostCenter": "AI-Research",
        "Project":    "ChatBot-v2",
        "Owner":      "team-ai@company.com",
    },
}
```

### Usage Analytics Queries
```sql
-- Top models by usage
SELECT ModelID, SUM(TokensUsed) as TotalTokens
FROM METRICS 
WHERE MetricName = 'TokensUsed' 
GROUP BY ModelID 
ORDER BY TotalTokens DESC;

-- Error rate by flow
SELECT FlowName, 
       SUM(FlowError) as Errors,
       SUM(FlowCompleted) as Success,
       (SUM(FlowError) / (SUM(FlowError) + SUM(FlowCompleted))) * 100 as ErrorRate
FROM METRICS 
GROUP BY FlowName;
```

## Troubleshooting

### ❌ No Metrics Appearing
**Possible causes**:
1. **IAM Permissions**: Ensure `cloudwatch:PutMetricData` permission
2. **Region Mismatch**: Verify CloudWatch region matches plugin region
3. **Namespace**: Check correct namespace in CloudWatch console

**Solution**:
```bash
# Check IAM permissions
aws iam simulate-principal-policy \
  --policy-source-arn arn:aws:iam::ACCOUNT:user/USERNAME \
  --action-names cloudwatch:PutMetricData \
  --resource-arns "*"
```

### ❌ Metrics Delayed
**Cause**: Normal CloudWatch ingestion delay (1-2 minutes)

**Solution**: Wait 2-5 minutes for metrics to appear

### ❌ High CloudWatch Costs
**Causes**:
1. Too many custom dimensions
2. High-frequency metric submission
3. Long retention periods

**Solutions**:
1. Reduce custom dimensions to essential ones
2. Increase `MetricBufferSize` to batch metrics
3. Use CloudWatch Logs for detailed debugging instead

## Integration with Other Tools

### Grafana Dashboard
```json
{
  "dashboard": {
    "title": "GenKit AWS Metrics",
    "panels": [
      {
        "title": "Generation Success Rate",
        "targets": [
          {
            "namespace": "MyApp/GenKit",
            "metricName": "FlowCompleted",
            "dimensions": {"Status": "Success"}
          }
        ]
      }
    ]
  }
}
```

### Prometheus Integration
Export CloudWatch metrics to Prometheus using:
- [CloudWatch Exporter](https://github.com/prometheus/cloudwatch_exporter)
- AWS CloudWatch Agent with Prometheus output

## Security Best Practices

### IAM Permissions (Minimal)
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cloudwatch:PutMetricData"
      ],
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "cloudwatch:namespace": "MyApp/GenKit"
        }
      }
    }
  ]
}
```

### Avoid Sensitive Data
- **Never include** API keys, user data, or PII in metric dimensions
- **Sanitize** error messages before including in metrics
- **Use generic identifiers** instead of specific user/session IDs

## Performance Tips

1. **Buffer metrics** with appropriate `MetricBufferSize`
2. **Limit dimensions** to essential ones (CloudWatch charges per unique combination)
3. **Use async collection** (metrics don't block your application)
4. **Monitor your monitoring** - track CloudWatch API costs

---

💡 **Pro tip**: Start with basic monitoring, then add custom dimensions and alarms as you understand your usage patterns.