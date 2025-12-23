package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MagdielCAS/magi-cli/pkg/shared"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize project rules and configuration",
		Long: `Analyzes the current project structure and creates/updates the .magi.yaml configuration
and AGENTS.md rules file. Uses AI to detect architecture and suggest actions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
            forceRules, _ := cmd.Flags().GetBool("force-rules")

			// 1. Safety Confirm
			confirm, _ := pterm.DefaultInteractiveConfirm.Show("This will analyze your project using LLM (consuming tokens) and may create/overwrite .magi.yaml. Proceed?")
			if !confirm {
				pterm.Info.Println("Aborted by user.")
				return nil
			}

			return RunAnalysisAndConfig(true, forceRules)
		},
	}
    cmd.Flags().Bool("force-rules", false, "Force creation/overwrite of AGENTS.md rules file")
    return cmd
}

// RunAnalysisAndConfig shared logic for init and redo.
func RunAnalysisAndConfig(createRules bool, forceRules bool) error {
	pterm.Info.Println("Initializing project analysis...")

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// 2. Run Architecture Analysis
	runtime, err := shared.BuildRuntimeContext()
	if err != nil {
		return fmt.Errorf("failed to load runtime context: %w", err)
	}

	agent := NewArchitectureAgent(runtime)
	spinner, _ := pterm.DefaultSpinner.Start("Analyzing project structure...")
	analysis, err := agent.Analyze(cwd)
	if err != nil {
		spinner.Fail("Analysis failed: " + err.Error())
		return err
	}
	spinner.Success("Analysis complete!")

    // 2.5 Validation
    validAgent := NewValidatorAgent(runtime)
    spinnerVal, _ := pterm.DefaultSpinner.Start("Validating analysis results...")
    analysis, err = validAgent.Validate(analysis)
    if err != nil {
        spinnerVal.Warning("Validation incomplete: " + err.Error())
        // Continue with original analysis instead of failing hard? 
        // Or fail? Let's log warning and proceed with potentially flawed analysis or original.
    } else {
        spinnerVal.Success("Validation complete!")
    }

	// 3. Log Results
	pterm.DefaultSection.Println("Project Analysis Result")
	pterm.Info.Printf("Architecture: %s\n", analysis.Architecture)
	pterm.Info.Printf("Project Type: %s\n", analysis.ProjectType)
	pterm.Info.Println("Identified Actions:")
	for _, action := range analysis.Actions {
		pterm.Println(pterm.Green("  - ") + action.Name + ": " + action.Description)
	}

	// 4. Update .magi.yaml
	configPath := filepath.Join(cwd, ".magi.yaml")
	var config MagiConfig

	// Read existing if any
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err == nil {
			_ = yaml.Unmarshal(data, &config)
		}
	}

	// Merge actions
	existingActions := make(map[string]Action)
	for _, a := range config.Actions {
		existingActions[a.Name] = a
	}

	for _, newAction := range analysis.Actions {
		if _, exists := existingActions[newAction.Name]; !exists {
			config.Actions = append(config.Actions, newAction)
		}
	}

	config.Architecture = analysis.Architecture
	config.ProjectType = analysis.ProjectType
	if config.RulesPath == "" {
		config.RulesPath = "AGENTS.md"
	}

	// Save
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write .magi.yaml: %w", err)
	}
	pterm.Success.Println("Updated .magi.yaml")

	// 5. Create AGENTS.md if missing
	if createRules || forceRules {
		rulesPath := filepath.Join(cwd, config.RulesPath)
        
        // Check existence
        rulesExist := false
		if _, err := os.Stat(rulesPath); err == nil {
            rulesExist = true
        }

        shouldCreate := false
        if forceRules {
            shouldCreate = true
            if rulesExist {
                pterm.Info.Println("Forcing recreation of AGENTS.md...")
            }
        } else if !rulesExist {
            createConfirm, _ := pterm.DefaultInteractiveConfirm.Show("Create default AGENTS.md?")
            if createConfirm {
                shouldCreate = true
            }
        }

		if shouldCreate {
            defaultRules := fmt.Sprintf("# %s Agent Rules\n\n## Architecture: %s\n## Type: %s\n\nAdd your project-specific rules here.", filepath.Base(cwd), analysis.Architecture, analysis.ProjectType)
            if err := os.WriteFile(rulesPath, []byte(defaultRules), 0644); err != nil {
                pterm.Error.Println("Failed to create AGENTS.md: " + err.Error())
            } else {
                pterm.Success.Println("Created/Updated " + config.RulesPath)
            }
		}
	}

	return nil
}
