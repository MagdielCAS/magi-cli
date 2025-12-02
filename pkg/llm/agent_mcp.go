package llm

import (
	"fmt"
	"strings"

	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

// MCPAgent represents an agent that can use MCP tools
type MCPAgent struct {
	Agent
	MCPClient *MCPClient
	Tools     []string // Available MCP tools for this agent
}

// MCPAgentConfig configuration for MCP-enabled agents
type MCPAgentConfig struct {
	Name          string
	Task          string
	Personality   string
	Tools         []string
	MaxTokens     int
	UseTextFormat bool
}

// NewMCPAgent creates a new MCP-enabled agent
func NewMCPAgent(config MCPAgentConfig, mcpClient *MCPClient, runtime *shared.RuntimeContext) *MCPAgent {
	agent := Agent{
		Name:        config.Name,
		Task:        config.Task,
		Personality: config.Personality,
		Runtime:     runtime,
		CompletionRequest: CompletionRequest{
			ChatCompletionRequest: ChatCompletionRequest{
				MaxTokens: float64(config.MaxTokens),
			},
		},
	}

	return &MCPAgent{
		Agent:     agent,
		MCPClient: mcpClient,
		Tools:     config.Tools,
	}
}

// AnalyzeWithMCP performs analysis using both LLM and MCP tools
func (a *MCPAgent) AnalyzeWithMCP(input map[string]string) (string, error) {
	// Gather MCP context
	mcpContext, err := a.gatherMCPContext(input)
	if err != nil {
		return "", fmt.Errorf("failed to gather MCP context: %w", err)
	}

	// Enhance input with MCP context
	enhancedInput := make(map[string]string)
	for k, v := range input {
		enhancedInput[k] = v
	}
	enhancedInput["mcp_context"] = mcpContext

	// Use standard agent analysis with enhanced context
	return a.Agent.Analyze(enhancedInput)
}

// gatherMCPContext collects relevant context from MCP tools
func (a *MCPAgent) gatherMCPContext(input map[string]string) (string, error) {
	var contextParts []string

	for _, toolName := range a.Tools {
		if toolName == "get_resource_details" {
			if resourceTypes, exists := input["resource_types"]; exists {
				// Map common short names to Pulumi tokens
				tokenMap := map[string]string{
					"s3":         "aws:s3/bucket:Bucket",
					"bucket":     "aws:s3/bucket:Bucket",
					"vpc":        "awsx:ec2:Vpc", // awsx is often preferred for VPC
					"ec2":        "aws:ec2/instance:Instance",
					"instance":   "aws:ec2/instance:Instance",
					"rds":        "aws:rds/instance:Instance",
					"lambda":     "aws:lambda/function:Function",
					"dynamodb":   "aws:dynamodb/table:Table",
					"sqs":        "aws:sqs/queue:Queue",
					"sns":        "aws:sns/topic:Topic",
					"cloudfront": "aws:cloudfront/distribution:Distribution",
					"elb":        "aws:lb/loadBalancer:LoadBalancer",
					"alb":        "aws:lb/loadBalancer:LoadBalancer",
					"ecs":        "aws:ecs/cluster:Cluster",
					"eks":        "aws:eks/cluster:Cluster",
					"iam":        "aws:iam/role:Role",
				}

				for _, resourceType := range strings.Split(resourceTypes, ",") {
					resourceType = strings.TrimSpace(strings.ToLower(resourceType))
					if token, ok := tokenMap[resourceType]; ok {
						details, err := a.MCPClient.GetResourceDetails(token)
						if err == nil {
							contextParts = append(contextParts, fmt.Sprintf("Pulumi Resource Details for %s (%s):\n%s", resourceType, token, details))
						}
					}
				}
			}
		}
	}

	return strings.Join(contextParts, "\n\n"), nil
}

// SetAPIKey sets the API key for the underlying agent
func (a *MCPAgent) SetAPIKey(apiKey string) {
	a.Agent.CompletionRequest.ApiKey = apiKey
}
