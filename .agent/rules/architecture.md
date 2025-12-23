---
trigger: always_on
---

# Architecture & Project Structure Rules

## 1. Directory Structure

- **cmd/**: Wiring only. Contains `root.go` and subdirectories strictly for command registration. **NO BUSINESS LOGIC**.
- **internal/cli/**: Implementation center. Each command group gets a package (e.g., `internal/cli/config/`).
- **pkg/**: Shared libraries. `pkg/shared` for cross-cutting concerns.
  - **Constraint**: `pkg/` packages must **NEVER** import `cmd/` or `internal/cli/`.

## 2. Command Package Structure

When adding a new command (e.g., `myfeature`), create `internal/cli/myfeature/`:

- **command.go**: MANDATORY entry point. Defines `func MyFeatureCmd() *cobra.Command`.
- **<subcommand>.go**: Implementation files (e.g., `list.go`, `create.go`).
- **<subcommand>_test.go**: Companion test files mandatory for every implementation file.

## 3. Forbidden Patterns

- ❌ Placing logic directly in `cmd/<file>.go`.
- ❌ Creating "utils" packages without domain scope (use `pkg/shared` or specific `pkg/<domain>`).

## 4. Command Lifecycle Rules

1. **Design doc first**: Capture inputs/outputs/security in an issue before coding.
2. **Single responsibility**: Each command lives in its own file under `internal/cli/<pkg>/`.
3. **Shared utilities**: Extract reusable logic into `pkg/shared` or `pkg/<domain>`.
4. **Flag discipline**: Declare flags in `init()`; validate in `RunE`.
5. **Rollout notes**: Update changelog and docs.

## 5. Command Pattern & Definition

When implementing commands, you **MUST** follow this pattern for `cobra.Command` definition:

```go
var cmd = &cobra.Command{
    // Use: one-line usage message. [ ] = optional, ... = multiple, | = exclusive
    Use:   "my-command [arg]",

    // Short: concise summary for "help" output
    Short: "Concise summary",

    // Long: detailed description -> subcommands -> usage -> examples
    Long: `Detailed description of what the command does.

Available subcommands:
  sub1    Description of sub1
  sub2    Description of sub2

Usage:
  magi my-command [arg]

Examples:
  # Example 1
  magi my-command value

  # Example 2
  magi my-command --flag`,

    // Run: function called on execution. USE RunE in strict mode, but Run is shown here.
    Run: func(cmd *cobra.Command, args []string) {
        // Implementation
    },
}
```

### Checklist for New Commands
- [ ] `Use` is correct and descriptive.
- [ ] `Short` provides a quick summary.
- [ ] `Long` follows: Description -> Subcommands -> Usage -> Examples.
- [ ] Examples cover common use cases.