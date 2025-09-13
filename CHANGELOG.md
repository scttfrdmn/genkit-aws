# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-09-12

### Added
- Initial release of GenKit AWS Runtime Plugins
- AWS Bedrock integration with support for Claude, Nova, and Llama models
- CloudWatch monitoring with automatic flow and model metrics
- Comprehensive plugin architecture following GenKit patterns
- Support for model configuration (temperature, max tokens, etc.)
- Buffered metric submission with error handling
- Example applications demonstrating usage patterns
- Complete test coverage for all core components
- Documentation and development setup guides

### Features
- **Models Supported:**
  - Anthropic Claude (3-sonnet, 3-haiku)
  - Amazon Nova (pro, lite, micro)  
  - Meta Llama (3-2 series)
- **Monitoring:**
  - Flow performance metrics
  - Model usage tracking
  - Custom dimensions support
  - Error classification
- **Developer Experience:**
  - Idiomatic Go APIs
  - Comprehensive examples
  - Production-ready configuration
  - CI/CD integration ready

[Unreleased]: https://github.com/scottfriedman/genkit-aws/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/scottfriedman/genkit-aws/releases/tag/v0.1.0