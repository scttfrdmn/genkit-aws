#!/bin/bash
# Copyright 2025 Scott Friedman
# Licensed under the Apache License, Version 2.0

# Standalone integration test script for GenKit AWS Runtime Plugins
# Tests real AWS Bedrock and CloudWatch integration

set -e

# Default values
DEFAULT_PROFILE="aws"
DEFAULT_REGION="us-west-2"
DEFAULT_TIMEOUT="5m"

# Parse command line arguments
PROFILE="${1:-$DEFAULT_PROFILE}"
REGION="${2:-$DEFAULT_REGION}"
TIMEOUT="${3:-$DEFAULT_TIMEOUT}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE} GenKit AWS Integration Tests${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo -e "Profile: ${YELLOW}$PROFILE${NC}"
    echo -e "Region:  ${YELLOW}$REGION${NC}"
    echo -e "Timeout: ${YELLOW}$TIMEOUT${NC}"
    echo ""
}

print_section() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

check_prerequisites() {
    print_section "Checking Prerequisites"
    
    # Check AWS CLI
    if ! command -v aws &> /dev/null; then
        print_error "AWS CLI not found. Please install it first."
        exit 1
    fi
    print_success "AWS CLI found"
    
    # Check if profile exists
    if ! aws configure list-profiles | grep -q "^$PROFILE$"; then
        print_error "AWS profile '$PROFILE' not found. Available profiles:"
        aws configure list-profiles
        exit 1
    fi
    print_success "AWS profile '$PROFILE' found"
    
    # Check AWS credentials
    if ! aws sts get-caller-identity --profile "$PROFILE" --region "$REGION" &> /dev/null; then
        print_error "Cannot authenticate with AWS profile '$PROFILE' in region '$REGION'"
        exit 1
    fi
    
    CALLER_IDENTITY=$(aws sts get-caller-identity --profile "$PROFILE" --region "$REGION" --output text --query 'Account')
    print_success "AWS authentication successful (Account: $CALLER_IDENTITY)"
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        print_error "Go not found. Please install Go 1.23 or later."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go found ($GO_VERSION)"
}

check_bedrock_access() {
    print_section "Checking Bedrock Model Access"
    
    # List available foundation models
    echo "Checking available Bedrock models..."
    
    # Test specific models we use
    MODELS=(
        "anthropic.claude-3-sonnet-20240229-v1:0"
        "amazon.nova-pro-v1:0"
        "amazon.nova-lite-v1:0"
    )
    
    for model in "${MODELS[@]}"; do
        if aws bedrock list-foundation-models --profile "$PROFILE" --region "$REGION" \
           --query "modelSummaries[?modelId=='$model'].modelId" --output text | grep -q "$model"; then
            print_success "Model $model is available"
        else
            print_warning "Model $model not available (may need to request access)"
        fi
    done
}

build_integration_test() {
    print_section "Building Integration Test"
    
    # Create integration test executable
    cat > integration_test_main.go << 'EOF'
// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/genkit-aws/genkit-aws/pkg/bedrock"
	genkitaws "github.com/genkit-aws/genkit-aws/pkg/genkit-aws"
	"github.com/genkit-aws/genkit-aws/pkg/monitoring"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: integration-test <profile> <region>")
		os.Exit(1)
	}
	
	profile := os.Args[1]
	region := os.Args[2]
	
	ctx := context.Background()
	
	fmt.Printf("🧪 Starting integration tests with profile=%s, region=%s\n\n", profile, region)
	
	// Create AWS plugin
	awsPlugin, err := genkitaws.New(&genkitaws.Config{
		Region:  region,
		Profile: profile,
		Bedrock: &bedrock.Config{
			Models: []string{
				"anthropic.claude-3-sonnet-20240229-v1:0",
			},
			DefaultModelConfig: &bedrock.ModelConfig{
				MaxTokens:   100, // Small for testing
				Temperature: 0.7,
			},
		},
		CloudWatch: &monitoring.Config{
			Namespace:          "GenKitAWS/IntegrationTest",
			EnableFlowMetrics:  true,
			EnableModelMetrics: true,
			CustomDimensions: map[string]string{
				"Environment": "Integration",
				"TestRun":     fmt.Sprintf("%d", time.Now().Unix()),
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create AWS plugin: %v", err)
	}

	// Initialize GenKit
	g := genkit.Init(ctx, genkit.WithPlugins(awsPlugin))

	// Define model
	model := awsPlugin.DefineModel(g, "anthropic.claude-3-sonnet-20240229-v1:0", nil)
	if model == nil {
		log.Fatal("Failed to define model")
	}
	fmt.Println("✅ Model defined successfully")

	// Test basic generation
	fmt.Println("🧪 Testing model generation...")
	
	testFlow := genkit.DefineFlow(g, "integrationTest", func(ctx context.Context, input string) (string, error) {
		resp, err := genkit.Generate(ctx, g,
			ai.WithModelName("anthropic.claude-3-sonnet-20240229-v1:0"),
			ai.WithMessages(ai.NewUserTextMessage(input)),
		)
		if err != nil {
			return "", fmt.Errorf("generation failed: %w", err)
		}

		if resp.Message == nil || len(resp.Message.Content) == 0 {
			return "", fmt.Errorf("no response generated")
		}

		return resp.Message.Content[0].Text, nil
	})

	// Run test
	start := time.Now()
	result, err := testFlow.Run(ctx, "Say 'Hello from GenKit AWS!' in exactly those words.")
	duration := time.Since(start)
	
	if err != nil {
		log.Fatalf("Integration test failed: %v", err)
	}
	
	fmt.Printf("✅ Generation successful (took %v)\n", duration)
	fmt.Printf("📝 Response: %s\n", result)
	
	// Verify monitoring
	monitor := awsPlugin.GetMonitor()
	if monitor != nil {
		fmt.Println("✅ CloudWatch monitoring initialized")
		
		// Wait for metrics to be potentially sent
		fmt.Println("⏳ Waiting for metrics to flush...")
		time.Sleep(2 * time.Second)
	}
	
	fmt.Println("\n🎉 Integration tests completed successfully!")
	fmt.Printf("📊 Check CloudWatch metrics in namespace: GenKitAWS/IntegrationTest\n")
}
EOF

    echo "Building integration test..."
    if ! go build -o integration-test integration_test_main.go; then
        print_error "Failed to build integration test"
        return 1
    fi
    
    print_success "Integration test built successfully"
}

run_integration_tests() {
    print_section "Running Integration Tests"
    
    echo "Executing integration test with AWS profile '$PROFILE' and region '$REGION'..."
    echo ""
    
    # Set timeout and run test
    timeout "$TIMEOUT" ./integration-test "$PROFILE" "$REGION" || {
        exit_code=$?
        if [ $exit_code -eq 124 ]; then
            print_error "Integration test timed out after $TIMEOUT"
        else
            print_error "Integration test failed with exit code $exit_code"
        fi
        return $exit_code
    }
    
    print_success "Integration tests completed successfully"
}

cleanup() {
    print_section "Cleanup"
    
    # Remove temporary files
    [ -f integration_test_main.go ] && rm integration_test_main.go
    [ -f integration-test ] && rm integration-test
    
    print_success "Cleanup completed"
}

main() {
    print_header
    
    # Trap to ensure cleanup happens
    trap cleanup EXIT
    
    check_prerequisites
    echo ""
    
    check_bedrock_access
    echo ""
    
    build_integration_test
    echo ""
    
    run_integration_tests
    echo ""
    
    print_section "Integration Test Summary"
    print_success "All integration tests passed"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo "1. Check CloudWatch metrics in AWS Console"
    echo "2. Verify model responses in your application"
    echo "3. Monitor costs in AWS Billing dashboard"
    echo ""
    echo -e "${YELLOW}Note: This test may incur small AWS charges for Bedrock and CloudWatch usage.${NC}"
}

# Show usage if help requested
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    echo "Usage: $0 [PROFILE] [REGION] [TIMEOUT]"
    echo ""
    echo "Parameters:"
    echo "  PROFILE   AWS profile name (default: $DEFAULT_PROFILE)"
    echo "  REGION    AWS region (default: $DEFAULT_REGION)"  
    echo "  TIMEOUT   Test timeout (default: $DEFAULT_TIMEOUT)"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Use defaults"
    echo "  $0 aws us-west-2                     # Custom profile and region"
    echo "  $0 aws us-west-2 10m                 # Custom timeout"
    echo ""
    echo "Prerequisites:"
    echo "  - AWS CLI configured with valid credentials"
    echo "  - AWS Bedrock model access enabled in target region"
    echo "  - Go 1.23+ installed"
    echo ""
    exit 0
fi

# Run main function
main