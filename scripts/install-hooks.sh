#!/bin/bash
# Install development tools and pre-commit hooks

set -e

echo "Installing GenKit AWS development tools..."

# Install Go tools
echo "Installing Go development tools..."
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
go install github.com/client9/misspell/cmd/misspell@latest
go install github.com/gordonklaus/ineffassign@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

echo "✅ Go tools installed"

# Copy pre-commit hook
echo "Installing pre-commit hook..."
cp scripts/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

echo "✅ Pre-commit hook installed"

echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "Available commands:"
echo "  make test              - Run tests"
echo "  make lint              - Run linting"
echo "  make build             - Build project"
echo "  make check             - Run all checks"
echo ""
echo "Pre-commit hook will automatically run quality checks before each commit."