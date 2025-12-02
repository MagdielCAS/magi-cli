package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/MagdielCAS/magi-cli/internal/cli/docker/compose"
	"github.com/MagdielCAS/magi-cli/internal/cli/docker/dockerfile"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var projectFiles = map[string]string{
	"go.mod":         "go",
	"nuxt.config.js": "nuxt",
	"nuxt.config.ts": "nuxt",
	"next.config.js": "next",
	"next.config.ts": "next",
	"package.json":   "node",
}

func NewDockerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docker",
		Short: "Manage Docker resources",
		Long: `Manage Docker resources including Dockerfiles and Docker Compose.

Available subcommands:
  compose     Set up Docker Compose for the project

Usage:
  magi docker [command]

Examples:
  # Detect project type, generate Dockerfile, build and run
  magi docker

  # Set up Docker Compose
  magi docker compose`,
		Run: func(cmd *cobra.Command, args []string) {
			runDocker()
		},
	}

	cmd.AddCommand(compose.NewComposeCommand())

	return cmd
}

func runDocker() {
	if !isDockerRunning() {
		pterm.Error.Println("Docker is not running. Please start Docker and try again.")
		return
	}

	projectType := identifyProjectType()
	pterm.Info.Printf("Detected project type: %s\n", projectType)

	if projectType == "unknown" {
		pterm.Warning.Println("Could not identify project type. Skipping Dockerfile generation.")
	} else {
		if _, err := os.Stat("Dockerfile"); os.IsNotExist(err) {
			generateDockerfile(projectType)
		} else {
			pterm.Info.Println("Dockerfile already exists. Skipping generation.")
		}
	}

	buildAndRunDocker()
}

func identifyProjectType() string {
	for file, projectType := range projectFiles {
		if _, err := os.Stat(file); err == nil {
			return projectType
		}
	}
	return "unknown"
}

func isDockerRunning() bool {
	cmd := exec.Command("docker", "info")
	return cmd.Run() == nil
}

func generateDockerfile(projectType string) {
	spinner, _ := pterm.DefaultSpinner.Start("Generating Dockerfile...")
	var content string

	// Ask user for port
	port, _ := pterm.DefaultInteractiveTextInput.WithDefaultValue("8080").Show("Enter the port your application listens on")
	if port == "" {
		port = "8080"
	}

	switch projectType {
	case "go":
		content = dockerfile.GenerateGoDockerfile(port)
	case "next":
		content = dockerfile.GenerateNextNuxtDockerfile("next", port, ".next")
	case "nuxt":
		content = dockerfile.GenerateNextNuxtDockerfile("nuxt", port, ".output")
	case "node":
		content = dockerfile.GenerateNodeDockerfile(port)
	}

	if content != "" {
		err := os.WriteFile("Dockerfile", []byte(content), 0644)
		if err != nil {
			spinner.Fail(fmt.Sprintf("Failed to create Dockerfile: %v", err))
			return
		}
		spinner.Success("Dockerfile generated successfully")
	} else {
		spinner.Fail("Failed to generate Dockerfile content")
	}
}

func buildAndRunDocker() {
	// Simple build and run implementation
	// In the future, this could be more interactive or configurable

	// Check if Dockerfile exists
	if _, err := os.Stat("Dockerfile"); os.IsNotExist(err) {
		pterm.Error.Println("No Dockerfile found. Cannot build.")
		return
	}

	pterm.Info.Println("Building Docker image...")

	wd, err := os.Getwd()
	if err != nil {
		pterm.Error.Printf("Failed to get working directory: %v\n", err)
		return
	}
	imageName := strings.ToLower(filepath.Base(wd))

	buildCmd := exec.Command("docker", "build", "-t", imageName, ".")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		pterm.Error.Printf("Docker build failed: %v\n", err)
		return
	}
	pterm.Success.Println("Docker image built successfully")

	pterm.Info.Println("Running Docker container...")

	portMapping, _ := pterm.DefaultInteractiveTextInput.WithDefaultValue("8080:8080").Show("Enter port mapping (host:container)")
	if portMapping == "" {
		portMapping = "8080:8080"
	}

	runCmd := exec.Command("docker", "run", "-p", portMapping, imageName)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	if err := runCmd.Run(); err != nil {
		pterm.Error.Printf("Docker run failed: %v\n", err)
		return
	}
}
