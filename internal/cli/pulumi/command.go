package pulumi

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MagdielCAS/magi-cli/internal/cli/pulumi/agents"
	"github.com/MagdielCAS/magi-cli/internal/cli/pulumi/parsers"
	"github.com/MagdielCAS/magi-cli/pkg/llm"
	"github.com/MagdielCAS/magi-cli/pkg/shared"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type PulumiFlags struct {
	InputText      string
	MermaidFile    string
	OutputDir      string
	ProjectName    string
	AwsRegion      string
	SkipValidation bool
	AutoConfirm    bool
	UseLocalMCP    bool
	MCPServerURL   string
}

func NewPulumiCommand() *cobra.Command {
	flags := &PulumiFlags{}

	cmd := &cobra.Command{
		Use:   "pulumi",
		Short: "Generate Pulumi infrastructure as code from architecture descriptions",
		Long: `Generate production-ready Pulumi TypeScript projects for AWS infrastructure.

This command transforms natural language descriptions and/or Mermaid architecture 
diagrams into complete Pulumi projects with proper AWS resource configurations.

FEATURES:
• Natural language to infrastructure translation
• Mermaid diagram parsing and interpretation  
• MCP integration for real-time Pulumi documentation
• AWS best practices and security configurations
• Production-ready TypeScript code generation
• Infrastructure validation and optimization
• Project scaffolding with proper structure

INPUT OPTIONS:
• Free text description via --text flag or interactive prompt
• Mermaid architecture file via --mermaid flag
• Combined text + diagram for enhanced context

OUTPUT:
• Complete Pulumi TypeScript project
• Proper project structure and dependencies
• AWS resource configurations with best practices
• Documentation and deployment instructions

EXAMPLES:
  # Generate from text description
  magi pulumi --text "Create a web app with RDS database and S3 storage"
  
  # Generate from Mermaid file
  magi pulumi --mermaid architecture.mmd
  
  # Interactive mode with custom output directory
  magi pulumi --output ./my-infrastructure
  
  # Specify AWS region and project name
  magi pulumi --region us-west-2 --project my-app-infra
  
  # Use local MCP server
  magi pulumi --use-local-mcp --mcp-server http://localhost:3000

The command uses MCP servers to access up-to-date Pulumi documentation and 
AWS best practices, ensuring generated code follows current standards.`,
		Run: func(cmd *cobra.Command, args []string) {
			runPulumi(flags)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&flags.InputText, "text", "t", "", "Natural language description of infrastructure")
	cmd.Flags().StringVarP(&flags.MermaidFile, "mermaid", "m", "", "Path to Mermaid architecture diagram file")
	cmd.Flags().StringVarP(&flags.OutputDir, "output", "o", "./pulumi-infrastructure", "Output directory for generated project")
	cmd.Flags().StringVarP(&flags.ProjectName, "project", "p", "", "Pulumi project name (auto-generated if not provided)")
	cmd.Flags().StringVarP(&flags.AwsRegion, "region", "r", "us-east-1", "AWS region for resources")
	cmd.Flags().BoolVar(&flags.SkipValidation, "skip-validation", false, "Skip infrastructure validation")
	cmd.Flags().BoolVarP(&flags.AutoConfirm, "yes", "y", false, "Auto-confirm all prompts")
	cmd.Flags().BoolVar(&flags.UseLocalMCP, "use-local-mcp", false, "Use local MCP server instead of default")
	cmd.Flags().StringVar(&flags.MCPServerURL, "mcp-server", "", "Custom MCP server URL")

	// Flag completions
	cmd.RegisterFlagCompletionFunc("region", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("mermaid", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterFileExt
	})

	return cmd
}

func runPulumi(flags *PulumiFlags) {
	// Validate inputs
	if flags.InputText == "" && flags.MermaidFile == "" {
		if !collectInputInteractively(flags) {
			pterm.Error.Println("No input provided. Operation cancelled.")
			return
		}
	}

	// Initialize MCP client
	mcpClient, err := initializeMCPClient(flags)
	if err != nil {
		pterm.Error.Printf("Failed to initialize MCP client: %v\n", err)
		return
	}
	defer mcpClient.Close()

	// Generate infrastructure
	if err := generateInfrastructure(flags, mcpClient); err != nil {
		pterm.Error.Printf("Failed to generate infrastructure: %v\n", err)
		return
	}

	pterm.Success.Printf("Pulumi project generated successfully in: %s\n", flags.OutputDir)
}

func initializeMCPClient(flags *PulumiFlags) (*llm.MCPClient, error) {
	serverURL := flags.MCPServerURL
	if serverURL == "" {
		if flags.UseLocalMCP {
			serverURL = "http://localhost:3000" // Default local
		} else if envURL := os.Getenv("MCP_SERVER_URL"); envURL != "" {
			serverURL = envURL
		} else {
			// Use official Pulumi MCP server
			serverURL = "https://mcp.ai.pulumi.com/mcp"
		}
	}

	client := llm.NewMCPClient(serverURL)
	if err := client.Connect(); err != nil {
		// Log warning but proceed, as agents should handle missing tools gracefully
		pterm.Warning.Printf("Could not connect to MCP server at %s: %v. Continuing without MCP tools.\n", serverURL, err)
	}

	return client, nil
}

func generateInfrastructure(flags *PulumiFlags, mcpClient *llm.MCPClient) error {
	// Build RuntimeContext
	runtime, err := shared.BuildRuntimeContext()
	if err != nil {
		return fmt.Errorf("failed to build runtime context: %w", err)
	}

	// 1. Analyze Architecture
	analyzer := agents.NewArchitectureAnalyzer(mcpClient, runtime)

	input := map[string]string{
		"aws_region": flags.AwsRegion,
	}

	if flags.InputText != "" {
		textParser := parsers.NewTextParser()
		processedText, err := textParser.Process(flags.InputText)
		if err != nil {
			return fmt.Errorf("invalid text input: %w", err)
		}
		input["text"] = processedText
	}

	if flags.MermaidFile != "" {
		mermaidParser := parsers.NewMermaidParser()
		content, err := mermaidParser.ParseFile(flags.MermaidFile)
		if err != nil {
			return fmt.Errorf("failed to parse mermaid file: %w", err)
		}
		input["mermaid_content"] = content
	}

	pterm.Info.Println("Analyzing architecture requirements...")
	analysis, err := analyzer.Analyze(input)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Show summary of analysis
	pterm.Info.Printf("Identified %d services, %d databases\n", len(analysis.Services), len(analysis.Storage.Databases))

	// 2. Generate Code
	generator := agents.NewPulumiGenerator(mcpClient, runtime)

	projectConfig := map[string]string{
		"project_name":     flags.ProjectName,
		"aws_region":       flags.AwsRegion,
		"output_directory": flags.OutputDir,
	}

	if projectConfig["project_name"] == "" {
		projectConfig["project_name"] = "pulumi-generated-project"
	}

	pterm.Info.Println("Generating Pulumi project code...")
	project, err := generator.Generate(analysis, projectConfig)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// 3. Validate (if not skipped)
	if !flags.SkipValidation {
		pterm.Info.Println("Validating generated infrastructure...")
		validator := agents.NewInfrastructureValidator(mcpClient, runtime)
		validationResult, err := validator.Validate(project)
		if err != nil {
			pterm.Warning.Printf("Validation failed: %v\n", err)
		} else {
			if !validationResult.IsValid {
				pterm.Warning.Println("Validation found issues:")
				for _, issue := range validationResult.Issues {
					pterm.Warning.Printf("- %s\n", issue)
				}
				for _, risk := range validationResult.SecurityRisks {
					pterm.Error.Printf("- Security Risk: %s\n", risk)
				}

				if !flags.AutoConfirm {
					confirm, _ := pterm.DefaultInteractiveConfirm.Show("Do you want to proceed with writing files despite these issues?")
					if !confirm {
						return fmt.Errorf("operation cancelled by user due to validation issues")
					}
				}
			} else {
				pterm.Success.Println("Validation passed successfully.")
			}
		}
	}

	// 4. Write Files
	if err := writeProjectFiles(project, flags.OutputDir); err != nil {
		return fmt.Errorf("failed to write project files: %w", err)
	}

	return nil
}

func writeProjectFiles(project *agents.GeneratedProject, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	for filename, content := range project.ProjectFiles {
		path := filepath.Join(outputDir, filename)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func collectInputInteractively(flags *PulumiFlags) bool {
	pterm.DefaultHeader.WithFullWidth().Println("Pulumi Infrastructure Generator")

	text, _ := pterm.DefaultInteractiveTextInput.Show("Describe your infrastructure")
	if text == "" {
		return false
	}
	flags.InputText = text

	if flags.ProjectName == "" {
		name, _ := pterm.DefaultInteractiveTextInput.Show("Project Name (default: pulumi-generated-project)")
		if name != "" {
			flags.ProjectName = name
		}
	}

	if flags.AwsRegion == "us-east-1" { // Default value
		region, _ := pterm.DefaultInteractiveTextInput.Show("AWS Region (default: us-east-1)")
		if region != "" {
			flags.AwsRegion = region
		}
	}

	return true
}
