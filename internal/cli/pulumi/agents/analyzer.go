package agents

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MagdielCAS/magi-cli/pkg/llm"
	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

// ArchitectureAnalyzer analyzes input and extracts infrastructure requirements
type ArchitectureAnalyzer struct {
	*llm.MCPAgent
}

// ArchitectureAnalysis represents the analysis result
type ArchitectureAnalysis struct {
	Services       []ServiceRequirement  `json:"services"`
	Networking     NetworkRequirement    `json:"networking"`
	Storage        StorageRequirement    `json:"storage"`
	Security       SecurityRequirement   `json:"security"`
	Monitoring     MonitoringRequirement `json:"monitoring"`
	Estimated_Cost string                `json:"estimated_cost"`
}

type ServiceRequirement struct {
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Description   string                 `json:"description"`
	Dependencies  []string               `json:"dependencies"`
	Configuration map[string]interface{} `json:"configuration"`
}

type NetworkRequirement struct {
	VPC          bool     `json:"vpc"`
	Subnets      []string `json:"subnets"`
	LoadBalancer bool     `json:"load_balancer"`
	CDN          bool     `json:"cdn"`
}

type StorageRequirement struct {
	Databases   []DatabaseRequirement    `json:"databases"`
	FileStorage []FileStorageRequirement `json:"file_storage"`
}

type DatabaseRequirement struct {
	Type   string `json:"type"`
	Engine string `json:"engine"`
	Size   string `json:"size"`
	Backup bool   `json:"backup"`
}

type FileStorageRequirement struct {
	Type        string `json:"type"`
	AccessLevel string `json:"access_level"`
	Encryption  bool   `json:"encryption"`
}

type SecurityRequirement struct {
	IAMRoles     []string `json:"iam_roles"`
	Encryption   bool     `json:"encryption"`
	VPCEndpoints bool     `json:"vpc_endpoints"`
	WAF          bool     `json:"waf"`
}

type MonitoringRequirement struct {
	CloudWatch bool `json:"cloudwatch"`
	Logging    bool `json:"logging"`
	Alerting   bool `json:"alerting"`
}

// NewArchitectureAnalyzer creates a new architecture analyzer
func NewArchitectureAnalyzer(mcpClient *llm.MCPClient, runtime *shared.RuntimeContext) *ArchitectureAnalyzer {
	config := llm.MCPAgentConfig{
		Name: "Architecture Analyzer",
		Task: `Analyze the provided architecture description and/or Mermaid diagram to extract detailed infrastructure requirements.

Your analysis should identify:
1. All required AWS services and their configurations
2. Networking requirements (VPC, subnets, load balancers, CDN)
3. Storage needs (databases, file storage, caching)
4. Security requirements (IAM, encryption, VPC endpoints)
5. Monitoring and logging requirements
6. Service dependencies and relationships
7. Estimated cost considerations

Use the MCP context to ensure recommendations follow current AWS and Pulumi best practices.

IMPORTANT: You must return the result as a valid JSON object with the following structure:
{
  "services": [
    {
      "name": "service_name",
      "type": "aws_service_type",
      "description": "description",
      "dependencies": ["dep1"],
      "configuration": {}
    }
  ],
  "networking": {
    "vpc": true,
    "subnets": ["public", "private"],
    "load_balancer": true,
    "cdn": false
  },
  "storage": {
    "databases": [{"type": "rds", "engine": "postgres", "size": "db.t3.micro", "backup": true}],
    "file_storage": [{"type": "s3", "access_level": "private", "encryption": true}]
  },
  "security": {
    "iam_roles": ["role1"],
    "encryption": true,
    "vpc_endpoints": false,
    "waf": false
  },
  "monitoring": {
    "cloudwatch": true,
    "logging": true,
    "alerting": false
  },
  "estimated_cost": "low/medium/high"
}
Do not wrap the JSON in markdown code blocks. Return raw JSON only.`,
		Personality: "Expert cloud architect with deep knowledge of AWS services, infrastructure patterns, and cost optimization. Skilled at translating business requirements into technical infrastructure specifications.",
		Tools:       []string{"get_resource_details"},
		MaxTokens:   4096,
	}

	mcpAgent := llm.NewMCPAgent(config, mcpClient, runtime)

	return &ArchitectureAnalyzer{
		MCPAgent: mcpAgent,
	}
}

// Analyze performs architecture analysis
func (a *ArchitectureAnalyzer) Analyze(input map[string]string) (*ArchitectureAnalysis, error) {
	// Extract resource types for MCP context
	resourceTypes := a.extractResourceTypes(input)
	input["resource_types"] = strings.Join(resourceTypes, ",")
	input["architecture_description"] = a.buildArchitectureDescription(input)

	result, err := a.AnalyzeWithMCP(input)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze architecture: %w", err)
	}

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

	// Parse the JSON response
	var analysis ArchitectureAnalysis
	if err := json.Unmarshal([]byte(result), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis result: %w\nResponse was: %s", err, result)
	}

	return &analysis, nil
}

// extractResourceTypes identifies AWS resource types from input
func (a *ArchitectureAnalyzer) extractResourceTypes(input map[string]string) []string {
	resourceTypes := []string{}

	text := strings.ToLower(input["text"] + " " + input["mermaid_content"])

	// Common AWS services mapping
	serviceMap := map[string]string{
		"database":      "rds",
		"web":           "ec2",
		"api":           "lambda",
		"storage":       "s3",
		"cache":         "elasticache",
		"queue":         "sqs",
		"notification":  "sns",
		"cdn":           "cloudfront",
		"load balancer": "elb",
	}

	for keyword, service := range serviceMap {
		if strings.Contains(text, keyword) {
			resourceTypes = append(resourceTypes, service)
		}
	}

	return resourceTypes
}

// buildArchitectureDescription creates a comprehensive description for analysis
func (a *ArchitectureAnalyzer) buildArchitectureDescription(input map[string]string) string {
	var parts []string

	if text := input["text"]; text != "" {
		parts = append(parts, fmt.Sprintf("Text Description: %s", text))
	}

	if mermaid := input["mermaid_content"]; mermaid != "" {
		parts = append(parts, fmt.Sprintf("Mermaid Diagram: %s", mermaid))
	}

	if region := input["aws_region"]; region != "" {
		parts = append(parts, fmt.Sprintf("Target AWS Region: %s", region))
	}

	return strings.Join(parts, "\n\n")
}
