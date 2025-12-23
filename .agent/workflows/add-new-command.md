---
description: Add a new command to the CLI following strict project guidelines
---

# Add New Command Workflow

This workflow automates the process of adding a new command to the `magi-cli` project. It ensures that all new commands adhere to the strict architectural and coding standards defined in `AGENTS.md`.

## Prerequisites
- You must have the command name (e.g., `audit`).
- You must have a clear description of what the command does.
- You must know the input flags/arguments and expected output.

## Workflow Steps

### 1. Analyze Requirements
1.  **Ask the user** for the command name, description, and key features if not already provided.
2.  **Plan the package name**: It should be `internal/cli/<command_name>`.
3.  **Plan the command structure**:
    - `Use`: `command [flags]`
    - `Short`: One-line summary.
    - `Long`: Detailed description including `Available subcommands`, `Usage`, and `Examples`.

### 2. Create Command Structure
1.  **Create Directory**:
    ```bash
    mkdir -p internal/cli/<command_name>
    ```
2.  **Create `command.go`**:
    -   Create `internal/cli/<command_name>/command.go`.
    -   **Package**: `package <command_name>`
    -   **Imports**: Include `github.com/spf13/cobra`, `github.com/pterm/pterm`, etc.
    -   **Factory Function**: `func New<CommandName>Cmd() *cobra.Command` (or `func <CommandName>Cmd()`).
    -   **Command Definition**: Follow the `AGENTS.md` pattern strictly.
        -   Start with `Use`, `Short`, `Long`.
        -   Define `RunE` (not `Run`).
        -   Use `pterm` for output.
        -   Use `viper` for config (if needed).
    -   **Validation**: Add strict input validation at the start of `RunE`.

    ```go
    package <command_name>

    import (
        "github.com/spf13/cobra"
        "github.com/pterm/pterm"
    )

    func New<CommandName>Cmd() *cobra.Command {
        cmd := &cobra.Command{
            Use:   "<command-name> [flags]",
            Short: "<short description>",
            Long: `<long description>

    Usage:
      magi <command-name> [flags]

    Examples:
      magi <command-name> --example`,
            RunE: func(cmd *cobra.Command, args []string) error {
                // Implementation here
                pterm.Success.Println("<CommandName> executed successfully")
                return nil
            },
        }
        return cmd
    }
    ```

### 3. Implement Tests
1.  **Create `command_test.go`**:
    -   Create `internal/cli/<command_name>/command_test.go`.
    -   Implement table-driven tests for the command.
    -   Verify basic execution and flag parsing.

### 4. Register Command
1.  **Edit `cmd/root.go`**:
    -   Import the new package: `github.com/MagdielCAS/magi-cli/internal/cli/<command_name>`.
    -   In `init()`, add the command: `rootCmd.AddCommand(<command_name>.New<CommandName>Cmd())`.

### 5. Update Documentation
1.  **Edit `docs/commands.md`**:
    -   Add a new section for the command.
    -   Include the same `Use`, `Short`, `Long` (description + examples) content.
    -   Note any configuration keys used.

### 6. Verify
1.  **Compile**: Run `go build ./...` to ensure no syntax errors.
2.  **Test**: Run `go test ./internal/cli/<command_name>/...` to pass local tests.
3.  **Lint**: Run `go vet ./internal/cli/<command_name>/...`.

## Final Check
- [ ] Does `Use`, `Short`, `Long` match between code and docs?
- [ ] Are forbidden functions (`fmt.Println`, `log.Println`) avoided?
- [ ] Is `pterm` used for output?
- [ ] Is the command registered in `root.go`?
