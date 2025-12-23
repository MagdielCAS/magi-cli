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

func NewUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update [file]",
		Short: "Update existing file using AI",
		Long: `Updates a specific file based on natural language instructions using AI.
Requires the file path as an argument.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get cwd: %w", err)
			}

            // 1. Get Target File
            var targetFile string
            if len(args) > 0 {
                targetFile = args[0]
            } else {
                targetFile, _ = pterm.DefaultInteractiveTextInput.Show("Enter file path to update")
            }

            if targetFile == "" {
                return fmt.Errorf("file path is required")
            }

            fullPath := filepath.Join(cwd, targetFile)
            content, err := os.ReadFile(fullPath)
            if err != nil {
                return fmt.Errorf("failed to read file '%s': %w", targetFile, err)
            }

            // 2. Get Instruction
            instruction, _ := pterm.DefaultInteractiveTextInput.Show("What changes should be made?")
            if instruction == "" {
                 pterm.Warning.Println("No instruction provided.")
                 return nil
            }

            // 3. Load Config for Context
            configPath := filepath.Join(cwd, ".magi.yaml")
             architecture := "Go Project"
             projectType := "Standard"
			if _, err := os.Stat(configPath); err == nil {
				data, err := os.ReadFile(configPath)
				if err == nil {
					var config MagiConfig
					if err := yaml.Unmarshal(data, &config); err == nil {
                        if config.Architecture != "" {
						    architecture = config.Architecture
                        }
                        if config.ProjectType != "" {
                            projectType = config.ProjectType
                        }
					}
				}
			}

            // 4. Run Update
            pterm.Info.Println("Generating updates...")
            runtime, err := shared.BuildRuntimeContext()
			if err != nil {
				return err
			}

            agent := NewGeneratorAgent(runtime)
            updatedFile, err := agent.UpdateContent(targetFile, string(content), instruction, architecture, projectType)
            if err != nil {
                return fmt.Errorf("update failed: %w", err)
            }

            // 5. Confirm and Write
            pterm.Println()
            pterm.DefaultSection.Println("Proposed Changes")
            // TODO: In the future, show a helper diff here? For now, we rely on user trust/git.
            pterm.Info.Printf("File: %s\n", updatedFile.Path)
            pterm.Info.Println("Content Length:", len(updatedFile.Content))
            
            confirm, _ := pterm.DefaultInteractiveConfirm.Show("Apply changes?")
            if confirm {
                if err := os.WriteFile(fullPath, []byte(updatedFile.Content), 0644); err != nil {
                    return fmt.Errorf("failed to write file: %w", err)
                }
                pterm.Success.Println("File updated successfully.")
            } else {
                pterm.Warning.Println("Changes discarded.")
            }

			return nil
		},
	}
}
