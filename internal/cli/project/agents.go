package project

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MagdielCAS/magi-cli/pkg/llm"
	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

// AnalysisResult represents the output of the ArchitectureAgent.
type AnalysisResult struct {
	Architecture string   `json:"architecture"`
	ProjectType  string   `json:"project_type"`
	Actions      []Action `json:"actions"`
}

// ArchitectureAgent analyzes the project structure.
type ArchitectureAgent struct {
	runtime *shared.RuntimeContext
}

func NewArchitectureAgent(runtime *shared.RuntimeContext) *ArchitectureAgent {
	return &ArchitectureAgent{runtime: runtime}
}

// Analyze scans the directory and uses LLM to identify architecture and suggested actions.
func (a *ArchitectureAgent) Analyze(rootPath string) (*AnalysisResult, error) {
	// 1. Gather file structure (simplified to avoid token limits)
	fileTree, err := getFileTree(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	// 2. Build LLM prompt
	systemPrompt := `You are an expert Software Architect. Analyze the provided file tree and identify:
1. The Architectural Pattern (e.g., Vertical Slice, MVC, Hexagonal, Clean Architecture).
2. The Project Type (e.g., Backend, Frontend, Monolith, Microservice, Event Driven).
3. A list of "Actions" that a developer would typically perform in this project.

CRITICAL RULES FOR ACTIONS:
- Action names MUST be in snake_case (e.g., "create_slice", "add_command", "add_middleware").
- You MUST identify necessary parameters for each action (e.g., "name" for creating a component, "method" for a handler).
- Parameters must have a name, description, type (string, bool, int), and required status.
- You MUST define the list of steps to execute this action.
    - Tools available: "create_file", "edit_file", "read_file", "search_replace", "run_command"
    - "create_file": instruction should describe the file purpose.
    - "edit_file": instruction should describe the change.
    - "run_command": 
        - "instruction": A brief description of what the command does (e.g., "Run all tests").
        - "parameters": MUST contain a key "command" with the EXACT executable shell command (e.g., "go test ./...").

Return the result in strictly valid JSON format matching this schema:
{
  "architecture": "string",
  "project_type": "string",
  "actions": [
    { 
      "name": "string", 
      "description": "string",
      "parameters": [
        { "name": "string", "description": "string", "type": "string", "required": true }
      ],
      "steps": [
        { "tool": "string", "instruction": "string", "parameters": { "key": "value" } }
      ]
    }
  ]
}
`

	userPrompt := fmt.Sprintf("Project Root: %s\n\nFile Tree:\n%s", filepath.Base(rootPath), fileTree)

	// 3. Call LLM
	service, err := llm.NewServiceBuilder(a.runtime).UseHeavyModel().Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build LLM service: %w", err)
	}

	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.1,
		// Using strict system prompt instead of ResponseFormat due to library version limitations/mismatch
	}

	resp, err := service.ChatCompletion(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// 4. Parse result
	var result AnalysisResult
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return &result, nil
}


// ValidatorAgent validates and corrects the AnalysisResult.
type ValidatorAgent struct {
	runtime *shared.RuntimeContext
}

func NewValidatorAgent(runtime *shared.RuntimeContext) *ValidatorAgent {
	return &ValidatorAgent{runtime: runtime}
}

// Validate checks the analysis result for common issues and attempts to fix them via LLM.
func (v *ValidatorAgent) Validate(result *AnalysisResult) (*AnalysisResult, error) {
    // 1. Check for invalid steps programmaticall first to save tokens
    needsFix := false
    var issues []string
    
    for _, action := range result.Actions {
        for i, step := range action.Steps {
            if step.Tool == "run_command" {
                if _, ok := step.Parameters["command"]; !ok {
                    needsFix = true
                    issues = append(issues, fmt.Sprintf("Action '%s' Step %d ('%s'): Missing 'command' parameter in run_command.", action.Name, i+1, step.Instruction))
                }
            }
        }
    }
    
    if !needsFix {
        return result, nil
    }

    // 2. Fix via LLM
    resultJSON, _ := json.Marshal(result)
    issuesStr := strings.Join(issues, "\n")
    
    systemPrompt := `You are a Strict Configuration Validator.
Your task is to FIX the provided Project Analysis JSON based on the reported validity issues.
Verify that all "run_command" steps have a "command" parameter with the actual executable shell command.
If the "instruction" contains the command, move it to "parameters.command" and keep "instruction" as a description.

Reported Issues:
` + issuesStr + `

Return the CORRECTED JSON strictly matching the input schema.`

    userPrompt := string(resultJSON)

	service, err := llm.NewServiceBuilder(v.runtime).UseHeavyModel().Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build LLM service: %w", err)
	}

	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.1,
	}

	resp, err := service.ChatCompletion(context.Background(), req)
	if err != nil {
		return nil, err
	}

	var fixedResult AnalysisResult
	if err := json.Unmarshal([]byte(resp), &fixedResult); err != nil {
		return nil, fmt.Errorf("failed to parse fixed LLM response: %w", err)
	}

	return &fixedResult, nil
}

// GeneratorAgent generates code based on actions.
type GeneratorAgent struct {
	runtime *shared.RuntimeContext
}

func NewGeneratorAgent(runtime *shared.RuntimeContext) *GeneratorAgent {
	return &GeneratorAgent{runtime: runtime}
}

type FileGenerationPlan struct {
	Files []GeneratedFile `json:"files"`
}

type GeneratedFile struct {
	Path        string `json:"path"`
	Description string `json:"description"`
}

type FileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// PlanGeneration asks LLM which files to create.
func (g *GeneratorAgent) PlanGeneration(rootPath string, architecture, projectType string, action Action, params map[string]string) (*FileGenerationPlan, error) {
	// 1. Build Prompt
	paramsJSON, _ := json.Marshal(params)
	systemPrompt := fmt.Sprintf(`You are an expert Software Architect for a %s project (%s).
Your task is to PLAN the creation of files for the action "%s" (%s).
Parameters: %s

Return a strictly valid JSON object listing the files that should be created.
Do not generate the content yet, just the paths and a brief description.
Schema:
{
  "files": [
    { "path": "path/to/file.go", "description": "Brief description" }
  ]
}`, architecture, projectType, action.Name, action.Description, string(paramsJSON))

	// 2. Call LLM (Light model likely enough for planning)
	service, err := llm.NewServiceBuilder(g.runtime).UseLightModel().Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build LLM service: %w", err)
	}

	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: "Plan the files for this action."},
		},
		Temperature: 0.1,
	}

	resp, err := service.ChatCompletion(context.Background(), req)
	if err != nil {
		return nil, err
	}

	var plan FileGenerationPlan
	if err := json.Unmarshal([]byte(resp), &plan); err != nil {
        // Fallback: try to find JSON block if strict JSON failed
		return nil, fmt.Errorf("failed to parse planning response: %w (%s)", err, resp)
	}

	return &plan, nil
}

// GenerateContent generates the content for a specific file.
func (g *GeneratorAgent) GenerateContent(rootPath, architecture, projectType string, action Action, params map[string]string, file GeneratedFile) (*FileContent, error) {
	paramsJSON, _ := json.Marshal(params)
	systemPrompt := fmt.Sprintf(`You are an expert Software Architect for a %s project (%s).
Your task is to GENERATE the content for the file "%s".
Action context: "%s" - %s.
Parameters: %s
File Description: %s

Return ONLY the code content for the file. No markdown code blocks, no explanations. 
If it is a go file, include the package declaration.`, architecture, projectType, file.Path, action.Name, action.Description, string(paramsJSON), file.Description)

	service, err := llm.NewServiceBuilder(g.runtime).UseHeavyModel().Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build LLM service: %w", err)
	}

	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: "Generate the file content."},
		},
		Temperature: 0.1,
	}

	resp, err := service.ChatCompletion(context.Background(), req)
	if err != nil {
		return nil, err
	}

    // Strip markdown code blocks if present
    content := strings.TrimSpace(resp)
    content = strings.TrimPrefix(content, "```go")
    content = strings.TrimPrefix(content, "```")
    content = strings.TrimSuffix(content, "```")

	return &FileContent{
		Path:    file.Path,
		Content: content,
	}, nil
}

// UpdateContent updates an existing file based on instructions.
func (g *GeneratorAgent) UpdateContent(filePath, originalContent, instruction, architecture, projectType string) (*FileContent, error) {
	systemPrompt := fmt.Sprintf(`You are an expert Software Architect for a %s project (%s).
Your task is to UPDATE the content of the file "%s" based on the user's instruction.

Original Content:
%s

User Instruction:
%s

Return ONLY the updated code content. No markdown code blocks, no explanations. 
Maintain the existing style and conventions.`, architecture, projectType, filePath, originalContent, instruction)

	service, err := llm.NewServiceBuilder(g.runtime).UseHeavyModel().Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build LLM service: %w", err)
	}

	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: "Update the file."},
		},
		Temperature: 0.1,
	}

	resp, err := service.ChatCompletion(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// Strip markdown code blocks if present
	content := strings.TrimSpace(resp)
	content = strings.TrimPrefix(content, "```go")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")

	return &FileContent{
		Path:    filePath,
		Content: content,
	}, nil
}

// ReviewerAgent checks project compliance.
type ReviewerAgent struct {
	runtime *shared.RuntimeContext
}

func NewReviewerAgent(runtime *shared.RuntimeContext) *ReviewerAgent {
	return &ReviewerAgent{runtime: runtime}
}

// ReviewCompliance checks if the project structure matches the rules.
func (r *ReviewerAgent) ReviewCompliance(rootPath string, rulesContent string) (string, error) {
   // 1. Get File Tree
    // We reuse a similar file tree function or extract it to a helper. 
    // Since getFileTree is a method of ArchitectureAgent, let's copy or refactor.
    // For simplicity I'll duplicate the walker logic here or make it a private function in agents.go
    // assuming I can access it if I make it a function not method, or just duplicate.
    // Let's refactor getFileTree to be a standalone function "getFileTree" in this package.
    
    fileTree, err := getFileTree(rootPath)
    if err != nil {
        return "", err
    }

	// 2. Build Prompt
	systemPrompt := `You are an expert Software Architect.
Your task is to REVIEW the provided file tree against the project rules definitions.
Identify any violations of the architecture or missing files/structure.

Rules:
` + rulesContent + `

Return a concise report of violations or "All checks passed" if everything looks good.
Format the output as a bulleted list of issues.`

	userPrompt := fmt.Sprintf("Project Root: %s\n\nFile Tree:\n%s", filepath.Base(rootPath), fileTree)

	// 3. Call LLM
	service, err := llm.NewServiceBuilder(r.runtime).UseHeavyModel().Build()
	if err != nil {
		return "", fmt.Errorf("failed to build LLM service: %w", err)
	}

	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.1,
	}

	resp, err := service.ChatCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}

	return resp, nil
}

// Helper function to be shared. 
// I need to change ArchitectureAgent.getFileTree to use this or be this.
func getFileTree(root string) (string, error) {
	var sb strings.Builder
	// Limit depth and exclude hidden/vendor dirs to save tokens
	maxDepth := 3
	// We might need deeper depth for review? Let's use 3 for now.
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if rel == "." {
			return nil
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			if strings.Count(rel, string(os.PathSeparator)) >= maxDepth {
				return filepath.SkipDir
			}
		}
		sb.WriteString(rel + "\n")
		return nil
	})
	return sb.String(), err
}

