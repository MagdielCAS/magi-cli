package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/MagdielCAS/magi-cli/pkg/shared"
	"github.com/pterm/pterm"
)

// Executor handles the execution of action steps.
type Executor struct {
	Agent          *GeneratorAgent
	Runtime        *shared.RuntimeContext
	Cwd            string
	Architecture   string
	ProjectType    string
	CurrentAction  Action
	CurrentParams  map[string]string
}

// NewExecutor creates a new Executor.
func NewExecutor(runtime *shared.RuntimeContext, cwd, arch, pType string, action Action, params map[string]string) *Executor {
	return &Executor{
		Agent:          NewGeneratorAgent(runtime),
		Runtime:        runtime,
		Cwd:            cwd,
		Architecture:   arch,
		ProjectType:    pType,
		CurrentAction:  action,
		CurrentParams:  params,
	}
}

// ExecuteSteps runs the defined steps sequentially.
func (e *Executor) ExecuteSteps(steps []ActionStep) error {
	pterm.Info.Printf("Executing %d steps for action '%s'...\n", len(steps), e.CurrentAction.Name)

	for i, step := range steps {
		stepNum := i + 1
		pterm.DefaultSection.Printf("Step %d: %s (%s)\n", stepNum, step.Instruction, step.Tool)

		var err error
		switch step.Tool {
		case "create_file":
			err = e.handleCreateFile(step)
		case "edit_file":
			err = e.handleEditFile(step)
		case "run_command":
			err = e.handleRunCommand(step)
		case "search_replace":
			err = e.handleSearchReplace(step)
		case "read_file":
			err = e.handleReadFile(step)
		default:
			pterm.Warning.Printf("Unknown tool '%s', skipping.\n", step.Tool)
		}

		if err != nil {
			return fmt.Errorf("step %d failed: %w", stepNum, err)
		}
	}
	return nil
}

// handleCreateFile handles file creation.
func (e *Executor) handleCreateFile(step ActionStep) error {
	// Plan the file path and description based on instruction
	stepAction := Action{Name: e.CurrentAction.Name, Description: step.Instruction}
	plan, err := e.Agent.PlanGeneration(e.Cwd, e.Architecture, e.ProjectType, stepAction, e.CurrentParams)
	if err != nil {
		return fmt.Errorf("failed to plan file creation: %w", err)
	}

	for _, f := range plan.Files {
		pterm.Info.Printf("Proposed File: %s\n", f.Path)
		if confirm, _ := pterm.DefaultInteractiveConfirm.Show("Generate this file?"); confirm {
			content, err := e.Agent.GenerateContent(e.Cwd, e.Architecture, e.ProjectType, stepAction, e.CurrentParams, f)
			if err != nil {
				return fmt.Errorf("failed generation: %w", err)
			}
			fullPath := filepath.Join(e.Cwd, f.Path)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(fullPath, []byte(content.Content), 0644); err != nil {
				return err
			}
			pterm.Success.Println("File created: " + f.Path)
		}
	}
	return nil
}

// handleEditFile handles editing existing files.
func (e *Executor) handleEditFile(step ActionStep) error {
	targetFile := e.resolveVariable(step.Parameters["target"])
	
	// If target is missing, try to resolve it via LLM or interactive prompt
	if targetFile == "" {
		// Use instruction to hint at the file. For now, interactive fallback.
		// Future: Agent that queries file tree to find best match.
		var err error
		targetFile, err = pterm.DefaultInteractiveTextInput.Show(fmt.Sprintf("Target file for '%s' (leave empty to skip)", step.Instruction))
		if err != nil || targetFile == "" {
			pterm.Warning.Println("Skipping edit step: no target file provided.")
			return nil
		}
	}

	fullPath := filepath.Join(e.Cwd, targetFile)
	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read target file '%s': %w", targetFile, err)
	}

	updated, err := e.Agent.UpdateContent(targetFile, string(contentBytes), step.Instruction, e.Architecture, e.ProjectType)
	if err != nil {
		return fmt.Errorf("failed to generate updates: %w", err)
	}

	// Show diff (simplified) or just confirm
	pterm.Info.Println("Proposed changes generated.")
	if confirm, _ := pterm.DefaultInteractiveConfirm.Show("Apply changes to " + targetFile + "?"); confirm {
		if err := os.WriteFile(fullPath, []byte(updated.Content), 0644); err != nil {
			return err
		}
		pterm.Success.Println("File updated.")
	}
	return nil
}

// handleRunCommand executes a shell command.
func (e *Executor) handleRunCommand(step ActionStep) error {
	cmdStr := e.resolveVariable(step.Instruction)
	// Also check if command is in params
	if cmdParam, ok := step.Parameters["command"]; ok {
		cmdStr = e.resolveVariable(cmdParam)
	}

	pterm.Info.Printf("Command: %s\n", cmdStr)
	
	if confirm, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).Show("Run this command?"); !confirm {
		pterm.Info.Println("Skipped command execution.")
		return nil
	}

	// Execute
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return nil
	}
	
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = e.Cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin // Allow interactivity

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	pterm.Success.Println("Command executed successfully.")
	return nil
}

// handleSearchReplace handles simple search and replace or agentic replacement.
func (e *Executor) handleSearchReplace(step ActionStep) error {
	// Treat as edit_file with specific instruction if no explicit search/replace params
	return e.handleEditFile(step)
}

// handleReadFile reads a file and logs logic (mostly for context).
func (e *Executor) handleReadFile(step ActionStep) error {
	targetFile := e.resolveVariable(step.Parameters["target"])
	if targetFile == "" {
		return nil
	}
	fullPath := filepath.Join(e.Cwd, targetFile)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		pterm.Error.Printf("Failed to read file %s: %v\n", targetFile, err)
		return nil // Don't block flow for read error?
	}
	pterm.Info.Printf("Read %s (%d bytes)\n", targetFile, len(content))
	return nil
}

// resolveVariable replaces {var} with values from CurrentParams.
func (e *Executor) resolveVariable(input string) string {
	if input == "" {
		return ""
	}
	output := input
	for k, v := range e.CurrentParams {
		key := fmt.Sprintf("{%s}", k)
		output = strings.ReplaceAll(output, key, v)
	}
	return output
}
