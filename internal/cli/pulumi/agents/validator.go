package agents

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MagdielCAS/magi-cli/pkg/llm"
	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

// InfrastructureValidator validates the generated infrastructure code
type InfrastructureValidator struct {
	*llm.MCPAgent
}

// ValidationResult represents the result of the validation
type ValidationResult struct {
	IsValid       bool     `json:"is_valid"`
	Issues        []string `json:"issues"`
	Suggestions   []string `json:"suggestions"`
	SecurityRisks []string `json:"security_risks"`
}

// NewInfrastructureValidator creates a new infrastructure validator
func NewInfrastructureValidator(mcpClient *llm.MCPClient, runtime *shared.RuntimeContext) *InfrastructureValidator {
	config := llm.MCPAgentConfig{
		Name: "Infrastructure Validator",
		Task: `Validate the provided Pulumi TypeScript project code for correctness, security, and best practices.

Review the code for:
1. Syntax errors or logical flaws
2. AWS security best practices (IAM, Security Groups, Encryption)
3. Pulumi best practices (Resource options, Config, Stack references)
4. Resource tagging and naming conventions
5. Potential cost issues

Use the MCP context to check against current AWS and Pulumi recommendations.

Return a JSON structure indicating validity and listing any issues or risks.

IMPORTANT: You must return the result as a valid JSON object with the following structure:
{
  "is_valid": true,
  "issues": ["issue1", "issue2"],
  "suggestions": ["suggestion1"],
  "security_risks": ["risk1"]
}
Do not wrap the JSON in markdown code blocks. Return raw JSON only.`,
		Personality: "Security-conscious Infrastructure Auditor. Meticulous about security, compliance, and code quality. detailed and critical.",
		Tools:       []string{"get_resource_details"},
		MaxTokens:   4096,
	}

	mcpAgent := llm.NewMCPAgent(config, mcpClient, runtime)

	return &InfrastructureValidator{
		MCPAgent: mcpAgent,
	}
}

// Validate performs validation on the generated project
func (v *InfrastructureValidator) Validate(project *GeneratedProject) (*ValidationResult, error) {
	// Serialize project files for analysis
	var filesBuilder strings.Builder
	for name, content := range project.ProjectFiles {
		filesBuilder.WriteString(fmt.Sprintf("--- %s ---\n%s\n\n", name, content))
	}

	input := map[string]string{
		"project_code": filesBuilder.String(),
	}

	// Extract resource types for MCP context
	resourceTypes := v.extractResourceTypes(filesBuilder.String())
	if len(resourceTypes) > 0 {
		input["resource_types"] = strings.Join(resourceTypes, ",")
	}

	result, err := v.AnalyzeWithMCP(input)
	if err != nil {
		return nil, fmt.Errorf("validation analysis failed: %w", err)
	}

	return v.parseValidationResult(result)
}

func (v *InfrastructureValidator) extractResourceTypes(code string) []string {
	// Simple heuristic to find AWS resources
	// Looks for patterns like: new aws.s3.Bucket
	// or: aws.ec2.Instance

	types := make(map[string]bool)

	// We'll just look for "aws.<service>" to be safe and broad
	// This avoids complex regex for now
	services := []string{
		"s3", "ec2", "rds", "lambda", "dynamodb", "sqs", "sns",
		"cloudfront", "route53", "iam", "vpc", "eks", "ecs",
	}

	lowerCode := strings.ToLower(code)
	for _, svc := range services {
		// Check for:
		// 1. aws.<service>
		// 2. @pulumi/aws/<service>
		// 3. awsx.<service> (common for ec2, ecs, lb)
		if strings.Contains(lowerCode, "aws."+svc) ||
			strings.Contains(lowerCode, "@pulumi/aws/"+svc) ||
			strings.Contains(lowerCode, "awsx."+svc) {
			types[svc] = true
		}
	}

	var result []string
	for t := range types {
		result = append(result, t)
	}
	return result
}

func (v *InfrastructureValidator) parseValidationResult(result string) (*ValidationResult, error) {
	// Clean up result if it contains markdown code blocks
	result = strings.TrimSpace(result)
	if strings.HasPrefix(result, "```json") {
		result = strings.TrimPrefix(result, "```json")
		result = strings.TrimSuffix(result, "```")
	} else if strings.HasPrefix(result, "```") {
		result = strings.TrimPrefix(result, "```")
		result = strings.TrimSuffix(result, "```")
	}
	result = strings.TrimSpace(result)

	var validation ValidationResult
	if err := json.Unmarshal([]byte(result), &validation); err != nil {
		return nil, fmt.Errorf("failed to parse validation result: %w", err)
	}

	return &validation, nil
}
