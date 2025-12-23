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

// NewExecCmd creates the exec command
func NewExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec [action]",
		Short: "Execute a defined action",
		Long:  `Executes a project action (e.g., create a slice, add a feature) defined in .magi.yaml.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
            cwd, err := os.Getwd()
            if err != nil {
                return fmt.Errorf("failed to get cwd: %w", err)
            }

            // 1. Load Actions Logic
            configPath := filepath.Join(cwd, ".magi.yaml")
            if _, err := os.Stat(configPath); os.IsNotExist(err) {
                pterm.Warning.Println(".magi.yaml not found. Run 'magi project init' first.")
                return nil
            }
            
            data, err := os.ReadFile(configPath)
            if err != nil {
                return fmt.Errorf("failed to read config: %w", err)
            }
            var config MagiConfig
            if err := yaml.Unmarshal(data, &config); err != nil {
                return fmt.Errorf("failed to parse config: %w", err)
            }

            if len(config.Actions) == 0 {
                pterm.Warning.Println("No actions defined in .magi.yaml. Run 'magi project init' to detect actions.")
                return nil
            }

            // 2. Select Action
            var actionName string
            if len(args) > 0 {
                actionName = args[0]
            } else {
                 var options []string
                 for _, a := range config.Actions {
                     options = append(options, a.Name)
                 }
                 actionName, _ = pterm.DefaultInteractiveSelect.WithOptions(options).Show("Select action to perform")
            }

            var selectedAction *Action
            for _, a := range config.Actions {
                if a.Name == actionName {
                    selectedAction = &a
                    break
                }
            }
            if selectedAction == nil {
                return fmt.Errorf("action '%s' not found", actionName)
            }

            // 3. Collect Parameters
            params := make(map[string]string)
            for _, p := range selectedAction.Parameters {
                val, _ := pterm.DefaultInteractiveTextInput.Show(fmt.Sprintf("%s (%s)", p.Name, p.Description))
                params[p.Name] = val
            }

            // 4. Execution Logic
            runtime, err := shared.BuildRuntimeContext()
             if err != nil {
                 return err
             }
             agent := NewGeneratorAgent(runtime)
             
             architecture := config.Architecture
             if architecture == "" { architecture = "Go Project" }
             projectType := config.ProjectType
             if projectType == "" { projectType = "Standard" }

            // If action has defined steps, execute them
            if len(selectedAction.Steps) > 0 {
                executor := NewExecutor(runtime, cwd, architecture, projectType, *selectedAction, params)
                if err := executor.ExecuteSteps(selectedAction.Steps); err != nil {
                    return err
                }
                pterm.Success.Println("Action completed successfully!")
                return nil
            }
            
            // Fallback to legacy Plan/Geneate
            pterm.Info.Println("No steps defined. specific steps. Fallback to auto-planning...")

             // ... Legacy Logic ...
             agent = NewGeneratorAgent(runtime)
             plan, err := agent.PlanGeneration(cwd, architecture, projectType, *selectedAction, params)
             if err != nil {
                 return fmt.Errorf("planning failed: %w", err)
             }

             // 5. Confirm Plan
             pterm.DefaultSection.Println("Generation Plan")
             for _, f := range plan.Files {
                 pterm.Println(pterm.Green("  + ") + f.Path + pterm.Gray(" ("+f.Description+")"))
             }
             
             confirm, _ := pterm.DefaultInteractiveConfirm.Show("Proceed with generation?")
             if !confirm {
                 pterm.Info.Println("Aborted.")
                 return nil
             }

             // 6. Generate and Write
             progressBar, _ := pterm.DefaultProgressbar.WithTotal(len(plan.Files)).Start()
             for _, f := range plan.Files {
                 progressBar.UpdateTitle("Generating " + f.Path)
                 content, err := agent.GenerateContent(cwd, architecture, projectType, *selectedAction, params, f)
                 if err != nil {
                     pterm.Error.Printf("Failed to generate %s: %v\n", f.Path, err)
                     continue 
                 }
                 
                 fullPath := filepath.Join(cwd, f.Path)
                 if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
                      pterm.Error.Printf("Failed to create dir for %s: %v\n", f.Path, err)
                      continue
                 }
                 
                 // Check overwrite
                 if _, err := os.Stat(fullPath); err == nil {
                     // In non-interactive mode or simply proceed for now as user confirmed plan.
                     // Ideally we verify individually but bulk confirm is standard for "create".
                 }

                 if err := os.WriteFile(fullPath, []byte(content.Content), 0644); err != nil {
                     pterm.Error.Printf("Failed to write %s: %v\n", f.Path, err)
                 }
                 progressBar.Increment()
             }
             progressBar.Stop()
             pterm.Success.Println("Generation complete!")

			return nil
		},
	}
    return cmd 
}
