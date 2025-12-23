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

func NewCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check compliance with project rules",
		Long: `Verifies if the current project structure complies with the rules defined in AGENTS.md.`,
        RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get cwd: %w", err)
			}
			
			// 1. Find Rules file
			configPath := filepath.Join(cwd, ".magi.yaml")
			var rulesFile string = "AGENTS.md" // Default
			if _, err := os.Stat(configPath); err == nil {
				data, err := os.ReadFile(configPath)
				if err == nil {
					var config MagiConfig
					if err := yaml.Unmarshal(data, &config); err == nil && config.RulesPath != "" {
						rulesFile = config.RulesPath
					}
				}
			}

			rulesPath := filepath.Join(cwd, rulesFile)
			if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
				pterm.Warning.Printf("Rules file '%s' not found. Cannot check compliance.\n", rulesFile)
				return nil
			}

			rulesContent, err := os.ReadFile(rulesPath)
			if err != nil {
				return fmt.Errorf("failed to read rules file: %w", err)
			}

			pterm.Info.Println("Checking project compliance...")
			runtime, err := shared.BuildRuntimeContext()
			if err != nil {
				return err
			}

			// 2. Run Review
			agent := NewReviewerAgent(runtime)
			spinner, _ := pterm.DefaultSpinner.Start("Analyzing project structure...")
			report, err := agent.ReviewCompliance(cwd, string(rulesContent))
			if err != nil {
				spinner.Fail("Review failed: " + err.Error())
				return err
			}
			spinner.Success("Analysis complete!")

			// 3. Output
			pterm.DefaultSection.Println("Compliance Report")
			pterm.Println(report)

			return nil
		},
	}
}
