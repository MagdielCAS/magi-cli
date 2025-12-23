package project

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func NewRedoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redo",
		Short: "Re-analyze project structure",
		Long: `Re-runs the project analysis to identify new structures or actions.
Updates .magi.yaml with findings.`,
        RunE: func(cmd *cobra.Command, args []string) error {
            forceRules, _ := cmd.Flags().GetBool("force-rules")

            // Safety Confirm
			confirm, _ := pterm.DefaultInteractiveConfirm.Show("This will re-analyze your project using LLM and update .magi.yaml. Proceed?")
			if !confirm {
				pterm.Info.Println("Aborted by user.")
				return nil
			}

            // Reuse init logic 
			return RunAnalysisAndConfig(false, forceRules)
		},
	}
    cmd.Flags().Bool("force-rules", false, "Force creation/overwrite of AGENTS.md rules file")
    return cmd
}
