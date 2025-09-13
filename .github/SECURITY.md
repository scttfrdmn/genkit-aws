# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability in GenKit AWS plugins, please follow these steps:

### 🚨 For Security Issues

**DO NOT** open a public GitHub issue for security vulnerabilities.

Instead, please report security issues by:

1. **Email**: Send details to `security@scottfriedman.dev`
2. **GitHub Security**: Use [GitHub Security Advisories](https://github.com/scttfrdmn/genkit-aws/security/advisories/new)

### 📋 What to Include

Please include the following information:
- **Description** of the vulnerability
- **Steps to reproduce** the issue
- **Potential impact** and severity assessment
- **Suggested fix** (if you have one)
- **Your contact information** for follow-up

### ⏱️ Response Timeline

- **Initial response**: Within 48 hours
- **Detailed assessment**: Within 1 week
- **Security fix**: Depends on severity (critical: days, others: weeks)
- **Public disclosure**: After fix is released and users have time to update

### 🛡️ Security Best Practices

When using GenKit AWS plugins:

#### **Credential Security**
- ✅ Use IAM roles instead of access keys when possible
- ✅ Rotate credentials regularly
- ✅ Use least-privilege permissions
- ❌ Never commit AWS credentials to code

#### **Data Privacy**
- ✅ Be aware that prompts/responses go to AWS Bedrock
- ✅ Review AWS Bedrock data usage policies
- ✅ Implement data classification for sensitive content
- ❌ Don't send PII, secrets, or confidential data without proper controls

#### **Network Security**
- ✅ Use VPC endpoints for AWS services when possible
- ✅ Enable AWS CloudTrail for API monitoring
- ✅ Use TLS 1.2+ for all communications
- ❌ Don't disable TLS certificate verification

#### **Monitoring Security**
- ✅ Monitor for unusual API usage patterns
- ✅ Set up CloudWatch alarms for error rates
- ✅ Review access logs regularly
- ❌ Don't include sensitive data in CloudWatch metrics

### 🔒 Minimum IAM Permissions

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": [
        "arn:aws:bedrock:*::foundation-model/anthropic.claude*",
        "arn:aws:bedrock:*::foundation-model/amazon.nova*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "cloudwatch:PutMetricData"
      ],
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "cloudwatch:namespace": "YourApp/GenKit"
        }
      }
    }
  ]
}
```

### 🚫 What NOT to Report

The following are **not** security vulnerabilities:
- Questions about usage or configuration
- AWS service outages or issues
- Performance problems
- Feature requests
- General bug reports

For these issues, please use [GitHub Issues](https://github.com/scttfrdmn/genkit-aws/issues) instead.

### 🏆 Recognition

Security researchers who responsibly disclose vulnerabilities will be:
- Credited in security advisories (if desired)
- Listed in our security hall of fame
- Eligible for acknowledgment in release notes

Thank you for helping keep GenKit AWS plugins secure! 🙏