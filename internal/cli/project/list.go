package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available actions",
		Long:  `Lists all actions defined in .magi.yaml along with their descriptions and parameters.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get cwd: %w", err)
			}

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

			pterm.DefaultSection.Println("Available Actions")

			if len(config.Actions) == 0 {
				pterm.Info.Println("No actions found.")
				return nil
			}

			tableData := [][]string{{"Name", "Description", "Parameters"}}
			for _, action := range config.Actions {
				params := ""
				for _, p := range action.Parameters {
					req := ""
					if p.Required {
						req = "*"
					}
					params += fmt.Sprintf("%s%s (%s), ", p.Name, req, p.Type)
				}
				if len(params) > 2 {
					params = params[:len(params)-2]
				}
				tableData = append(tableData, []string{action.Name, action.Description, params})
			}

			pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
			return nil
		},
	}
	return cmd
}
