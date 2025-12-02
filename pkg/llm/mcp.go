package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// MCPClient represents a Model Context Protocol client
type MCPClient struct {
	ServerURL  string
	HTTPClient *http.Client
	SessionID  string
	Tools      map[string]MCPTool
}

// MCPTool represents an available MCP tool
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCPRequest represents a request to an MCP server
type MCPRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

// MCPResponse represents a response from an MCP server
type MCPResponse struct {
	Result interface{} `json:"result"`
	Error  *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewMCPClient creates a new MCP client
func NewMCPClient(serverURL string) *MCPClient {
	return &MCPClient{
		ServerURL: serverURL,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		Tools: make(map[string]MCPTool),
	}
}

// Connect establishes connection to MCP server and initializes session
func (c *MCPClient) Connect() error {
	// Initialize session
	req := MCPRequest{
		Method: "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "magi-cli",
				"version": "1.0.0",
			},
		},
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("failed to initialize MCP session: %w", err)
	}

	// Extract session info and available tools
	if result, ok := resp.Result.(map[string]interface{}); ok {
		if serverInfo, ok := result["serverInfo"].(map[string]interface{}); ok {
			c.SessionID = fmt.Sprintf("%v", serverInfo["name"])
		}
	}

	// List available tools
	return c.listTools()
}

// listTools retrieves available tools from the MCP server
func (c *MCPClient) listTools() error {
	req := MCPRequest{
		Method: "tools/list",
		Params: map[string]interface{}{},
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	if result, ok := resp.Result.(map[string]interface{}); ok {
		if tools, ok := result["tools"].([]interface{}); ok {
			for _, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					mcpTool := MCPTool{
						Name:        fmt.Sprintf("%v", toolMap["name"]),
						Description: fmt.Sprintf("%v", toolMap["description"]),
					}
					if schema, ok := toolMap["inputSchema"].(map[string]interface{}); ok {
						mcpTool.InputSchema = schema
					}
					c.Tools[mcpTool.Name] = mcpTool
				}
			}
		}
	}

	return nil
}

// CallTool executes a tool on the MCP server
func (c *MCPClient) CallTool(toolName string, arguments map[string]interface{}) (interface{}, error) {
	if _, exists := c.Tools[toolName]; !exists {
		return nil, fmt.Errorf("tool %s not available", toolName)
	}

	req := MCPRequest{
		Method: "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool %s: %w", toolName, err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("MCP tool error: %s", resp.Error.Message)
	}

	return resp.Result, nil
}

// GetResourceDetails retrieves detailed information about a Pulumi Registry resource
func (c *MCPClient) GetResourceDetails(token string) (string, error) {
	result, err := c.CallTool("get-resource", map[string]interface{}{
		"token": token,
	})
	if err != nil {
		return "", err
	}

	// The result structure depends on the tool implementation, but typically it returns a JSON string or object
	// We'll try to marshal it to string if it's an object
	if resultStr, ok := result.(string); ok {
		return resultStr, nil
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal resource details: %w", err)
	}

	return string(jsonBytes), nil
}

// sendRequest sends an HTTP request to the MCP server
func (c *MCPClient) sendRequest(req MCPRequest) (*MCPResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.ServerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer httpResp.Body.Close()

	var mcpResp MCPResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&mcpResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &mcpResp, nil
}

// Close closes the MCP client connection
func (c *MCPClient) Close() error {
	// No active connection to close for HTTP client
	return nil
}
