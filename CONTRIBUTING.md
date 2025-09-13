# Contributing to GenKit AWS Plugins

Thank you for your interest in contributing to GenKit AWS! This document provides guidelines for contributing to the project.

## 🚀 Quick Start

1. **Fork and clone** the repository
2. **Set up development environment**:
   ```bash
   cd genkit-aws
   make deps
   ./scripts/install-hooks.sh
   ```
3. **Make your changes** following our coding standards
4. **Test thoroughly** with both unit and integration tests
5. **Submit a pull request** with a clear description

## 📋 Development Guidelines

### Prerequisites

- **Go 1.23+** installed
- **AWS CLI** configured with credentials
- **AWS Bedrock access** enabled in your test region
- **Pre-commit hooks** installed (`./scripts/install-hooks.sh`)

### Code Standards

#### **Go Code Quality**
- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- All code must pass: `make check` (fmt, vet, lint, test)
- Maintain **goreportcard.com A rating**
- Add comprehensive tests for new functionality
- Use meaningful variable and function names

#### **Commit Messages**
Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

<body>

<footer>
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

**Examples**:
```bash
feat(bedrock): add support for Amazon Nova models
fix(monitoring): resolve CloudWatch metric buffering race condition
docs(readme): improve installation instructions
test(integration): add coverage for error scenarios
```

#### **Pull Request Process**

1. **Create feature branch**: `git checkout -b feature/my-new-feature`
2. **Make changes** with tests and documentation
3. **Run quality checks**: `make check`
4. **Run integration tests**: `make test-integration` (if applicable)
5. **Commit with conventional format**
6. **Push and create PR** with detailed description

### Testing Requirements

#### **Unit Tests** (Required)
```bash
make test              # Must pass
make test-coverage     # Aim for >30% coverage
```

#### **Integration Tests** (For AWS changes)
```bash
make test-integration  # Must pass with real AWS
```

#### **Quality Checks** (Required)
```bash
make lint              # Zero issues
make check             # All quality gates pass
```

## 🏗️ Architecture Guidelines

### Package Structure
```
pkg/
├── genkit-aws/        # Main plugin interface
├── bedrock/           # AWS Bedrock integration
└── monitoring/        # CloudWatch monitoring

internal/
├── constants/         # Shared constants
└── version/          # Version information

cmd/examples/         # Usage examples
test/integration/     # Integration tests
```

### Interface Design
- **Prefer interfaces** over concrete types for testability
- **Return errors** instead of panicking
- **Use context.Context** for cancellation and timeouts
- **Follow GenKit patterns** established by Google's plugins

### Error Handling
```go
// ✅ Good: Return errors
func New(config *Config) (*Plugin, error) {
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    return &Plugin{config: config}, nil
}

// ❌ Bad: Panic in constructors  
func New(config *Config) *Plugin {
    if err := config.Validate(); err != nil {
        panic(err) // Don't do this
    }
    return &Plugin{config: config}
}
```

## 🧪 Testing Guidelines

### Unit Tests
- **Test public APIs** thoroughly
- **Mock external dependencies** (AWS services)
- **Cover edge cases** and error scenarios
- **Use table-driven tests** for multiple scenarios

```go
func TestConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  *Config
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid config",
            config: &Config{Region: "us-east-1"},
            wantErr: false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Integration Tests
- **Use build tags**: `//go:build integration`
- **Test real AWS services** with minimal costs
- **Clean up resources** after tests
- **Document AWS requirements** clearly

## 📝 Documentation Standards

### Code Documentation
- **Package documentation** for all public packages
- **Function documentation** for all public functions
- **Example usage** in documentation
- **Link to related concepts** when helpful

```go
// Package bedrock provides AWS Bedrock integration for Google's GenKit framework.
//
// This package enables GenKit applications to use AWS Bedrock's foundation models
// including Claude, Nova, and Llama families. It handles model-specific request/response
// formatting and provides a unified interface for all supported models.
//
// Example usage:
//   config := &bedrock.Config{
//       Models: []string{"anthropic.claude-3-sonnet-20240229-v1:0"},
//   }
//   client, err := bedrock.NewClient(ctx, awsConfig, config)
package bedrock
```

### README Updates
- Keep **installation instructions** current
- Update **example code** when APIs change
- Add **troubleshooting sections** for common issues
- Include **integration test instructions**

## 🔧 Common Development Tasks

### Adding New Model Support
1. **Update model detection** in `pkg/bedrock/client.go`
2. **Add conversion logic** for request/response format
3. **Add model to default configs** in constants
4. **Write tests** for the new model
5. **Update documentation** with supported models

### Adding New Monitoring Metrics
1. **Define metric in CloudWatch package**
2. **Add configuration options** if needed
3. **Update monitoring interface** if required
4. **Add tests** for metric collection
5. **Document metric** in README

### Performance Improvements
1. **Profile first** to identify bottlenecks
2. **Write benchmarks** for critical paths
3. **Optimize without breaking APIs**
4. **Verify with integration tests**

## 🐛 Bug Reports

When reporting bugs, please include:

- **Go version**: `go version`
- **AWS region** and services used
- **GenKit version** and configuration
- **Full error messages** and stack traces
- **Minimal reproduction case**
- **Expected vs actual behavior**

## 💡 Feature Requests

For new features, please provide:

- **Use case description** and motivation
- **Proposed API design** (if applicable)
- **Alternative solutions** considered
- **Breaking change assessment**

## 📦 Release Process

Releases follow **Semantic Versioning 2.0**:

- **MAJOR**: Breaking API changes
- **MINOR**: New features, backwards compatible
- **PATCH**: Bug fixes, backwards compatible

### Changelog
All changes are documented in [CHANGELOG.md](./CHANGELOG.md) following [Keep a Changelog](https://keepachangelog.com/) format.

## 🙏 Recognition

Contributors are recognized in:
- **GitHub contributors** page
- **Release notes** for significant contributions  
- **Documentation acknowledgments** where appropriate

## 📞 Getting Help

- **GitHub Issues**: Bug reports and feature requests
- **Discussions**: Questions and community support
- **Documentation**: Comprehensive guides in `/docs`

---

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.