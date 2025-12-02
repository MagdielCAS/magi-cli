package compose

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MagdielCAS/magi-cli/pkg/llm"
	"github.com/MagdielCAS/magi-cli/pkg/shared"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func NewComposeCommand() *cobra.Command {
	var autoAccept bool

	cmd := &cobra.Command{
		Use:   "compose [flags]",
		Short: "Set up Docker Compose for the project",
		Long: `Generates and manages a docker-compose.yml file for your project.

This command analyzes your project structure, detects Dockerfiles, and helps you create a comprehensive Docker Compose configuration. It supports adding common services (databases, caches, etc.) and uses AI to validate and refine the generated configuration.

Usage:
  magi docker compose [flags]

Examples:
  # Interactive mode (default)
  magi docker compose

  # Auto-accept all prompts with defaults
  magi docker compose --yes

Security:
  This command sends the generated configuration and any custom service descriptions to the configured LLM provider for validation and generation. Ensure no secrets are hardcoded in your service descriptions.`,
		Run: func(cmd *cobra.Command, args []string) {
			runCompose(cmd.Context(), autoAccept)
		},
	}

	cmd.Flags().BoolVarP(&autoAccept, "yes", "y", false, "Auto-accept prompts")
	return cmd
}

func runCompose(ctx context.Context, autoAccept bool) {
	// 1. Dockerfile Discovery
	dockerfiles := findDockerfiles()
	if len(dockerfiles) == 0 {
		pterm.Warning.Println("No Dockerfiles found. Proceeding with service selection only.")
	} else {
		pterm.Info.Printf("Found %d Dockerfiles: %v\n", len(dockerfiles), dockerfiles)
	}

	// 2. Service Selection
	serviceNames := make([]string, 0, len(ServiceConfigs))
	for name := range ServiceConfigs {
		serviceNames = append(serviceNames, name)
	}
	serviceNames = append(serviceNames, "Custom Service")
	sort.Strings(serviceNames)

	selectedServices, _ := pterm.DefaultInteractiveMultiselect.
		WithOptions(serviceNames).
		WithDefaultText("Select services to include").
		Show()

	if len(selectedServices) == 0 && len(dockerfiles) == 0 {
		pterm.Warning.Println("No services selected and no Dockerfiles found. Exiting.")
		return
	}

	// 3. Content Generation
	composeContent, err := generateComposeContent(ctx, selectedServices, dockerfiles, autoAccept)
	if err != nil {
		pterm.Error.Printf("Failed to generate compose content: %v\n", err)
		return
	}

	// 4. AI Validation
	validatedContent, err := validateDockerCompose(ctx, composeContent)
	if err != nil {
		pterm.Warning.Printf("AI validation failed (using original content): %v\n", err)
		validatedContent = composeContent
	}

	// 5. File Creation
	err = os.WriteFile("docker-compose.yml", []byte(validatedContent), 0644)
	if err != nil {
		pterm.Error.Printf("Failed to write docker-compose.yml: %v\n", err)
		return
	}
	pterm.Success.Println("docker-compose.yml created successfully")

	// 6. Post-Creation Actions
	for _, serviceName := range selectedServices {
		config := ServiceConfigs[serviceName]
		if config.AfterComposeCreated != nil {
			if err := config.AfterComposeCreated(ctx, validatedContent, autoAccept); err != nil {
				pterm.Warning.Printf("Post-creation action for %s failed: %v\n", serviceName, err)
			}
		}
	}
}

func findDockerfiles() []string {
	var dockerfiles []string
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "Dockerfile") {
			// Use relative path
			relPath, _ := filepath.Rel(".", path)
			dockerfiles = append(dockerfiles, relPath)
		}
		return nil
	})
	return dockerfiles
}

func generateComposeContent(ctx context.Context, services []string, dockerfiles []string, autoAccept bool) (string, error) {
	var sb strings.Builder
	sb.WriteString("version: '3.8'\n\n")

	// Add top-level configurations (templates/anchors)
	for _, serviceName := range services {
		if serviceName == "Custom Service" {
			continue
		}
		config := ServiceConfigs[serviceName]
		if config.BeforeServices != nil {
			before := config.BeforeServices()
			if before != "" {
				sb.WriteString(before + "\n\n")
			}
		}
	}

	sb.WriteString("services:\n")

	// Add app services from Dockerfiles
	for i, df := range dockerfiles {
		dir := filepath.Dir(df)
		serviceName := "app"
		if i > 0 {
			serviceName = fmt.Sprintf("app-%d", i+1)
		}
		// If directory name is meaningful, use it
		if dir != "." {
			serviceName = filepath.Base(dir)
		}

		sb.WriteString(fmt.Sprintf("  %s:\n", serviceName))
		sb.WriteString(fmt.Sprintf("    build: %s\n", dir))
		port := getExposedPort(df)
		sb.WriteString(fmt.Sprintf("    ports:\n      - \"%s:%s\"\n", port, port))
		sb.WriteString("    restart: unless-stopped\n\n")
	}

	// Process dependencies and add services
	finalServices := make([]string, len(services))
	copy(finalServices, services)

	// Check dependencies
	for _, serviceName := range services {
		config := ServiceConfigs[serviceName]
		if config.CheckOtherServices != nil {
			additionalConfig, err := config.CheckOtherServices(finalServices, autoAccept)
			if err != nil {
				return "", err
			}
			if additionalConfig != "" {
				sb.WriteString(additionalConfig + "\n\n")
			}
		}
	}

	// Add selected services
	for _, serviceName := range services {
		if serviceName == "Custom Service" {
			customDesc, _ := pterm.DefaultInteractiveTextInput.Show("Describe the custom service you want to add")
			if customDesc != "" {
				customConfig, err := generateCustomServiceConfig(ctx, customDesc)
				if err != nil {
					pterm.Warning.Printf("Failed to generate custom service config: %v\n", err)
				} else {
					sb.WriteString(customConfig + "\n\n")
				}
			}
			continue
		}
		config := ServiceConfigs[serviceName]
		sb.WriteString(config.ConfigFunc() + "\n\n")
	}

	return sb.String(), nil
}

func generateCustomServiceConfig(ctx context.Context, description string) (string, error) {
	spinner, _ := pterm.DefaultSpinner.Start("Generating custom service configuration...")

	runtime, err := shared.BuildRuntimeContext()
	if err != nil {
		return "", err
	}

	builder := llm.NewServiceBuilder(runtime)
	service, err := builder.Build()
	if err != nil {
		return "", err
	}

	prompt := []llm.ChatMessage{
		{
			Role: "system",
			Content: `You are a Docker Compose expert. Generate a Docker Compose service configuration based on the user's description.
            Return ONLY the YAML configuration for the service, indented with 2 spaces. Do not include 'services:' or 'version:'.
            Do not include markdown code blocks.`,
		},
		{
			Role:    "user",
			Content: description,
		},
	}

	req := llm.ChatCompletionRequest{
		Messages:  prompt,
		MaxTokens: 1024,
	}

	result, err := service.ChatCompletion(ctx, req)
	if err != nil {
		spinner.Fail("Failed to generate custom service")
		return "", err
	}

	// Remove code blocks if present
	result = strings.TrimPrefix(result, "```yaml")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSuffix(result, "```")

	spinner.Success("Custom service generated")
	return result, nil

}

func validateDockerCompose(ctx context.Context, content string) (string, error) {
	spinner, _ := pterm.DefaultSpinner.Start("GPETE analyzing compose file...")

	runtime, err := shared.BuildRuntimeContext()
	if err != nil {
		spinner.Warning("Could not create runtime context for AI validation")
		return content, nil
	}

	builder := llm.NewServiceBuilder(runtime)
	service, err := builder.Build()
	if err != nil {
		spinner.Warning("Could not build LLM service for AI validation")
		return content, nil
	}

	prompt := []llm.ChatMessage{
		{
			Role: "system",
			Content: `You are a Senior Software Engineer and Docker Compose expert.
            Analyze and fix any indentation errors, misconfigurations, typos,
            missing volume setups, and network configurations. Ensure the file can run.
            Return ONLY the fixed docker-compose.yml content. Do not include markdown code blocks.`,
		},
		{
			Role:    "user",
			Content: content,
		},
	}

	req := llm.ChatCompletionRequest{
		Messages:  prompt,
		MaxTokens: 2048,
	}

	result, err := service.ChatCompletion(ctx, req)
	if err != nil {
		spinner.Fail("AI validation failed")
		return content, err
	}

	// Remove code blocks if present (simple cleanup)
	result = strings.TrimPrefix(result, "```yaml")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSuffix(result, "```")

	spinner.Success("AI validation complete")
	return result, nil
}

func getExposedPort(dockerfilePath string) string {
	file, err := os.Open(dockerfilePath)
	if err != nil {
		return "8080"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "EXPOSE") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				return parts[1]
			}
		}
	}
	return "8080"
}
