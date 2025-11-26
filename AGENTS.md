# AGENTS Guidelines for This Repository

This repository contains a Golang CLI application located in the root of this repository. When
working on the project interactively with an agent (e.g. the Codex CLI) please follow
the guidelines below so that the development experience continues to work smoothly.

Following these practices ensures that the agent-assisted development workflow stays
fast and dependable.

## 1. Secure CLI With AI Capabilities

The following rules describe how `magi` keeps user data secure while exposing AI-backed commands. They apply to every new feature, test, or refactor.

### Baseline Principles

1. **Least privilege.** Only request API keys, scopes, or filesystem access the command truly needs. Commands must fail fast when secrets are missing instead of guessing defaults.
2. **Explicit configuration.** All user-facing settings live in Viper configuration files and are discoverable via `magi config`. Never load implicit environment-specific defaults.
3. **Clear data boundaries.** Commands must document what user information leaves the machine, for what purpose, and whether it is persisted.
4. **Deterministic prompts.** All AI prompts are reviewable strings checked into the repository. No runtime string building that mixes unescaped user input with template logic.
5. **Reproducible builds.** Go modules are pinned in `go.sum` and binaries are reproducible using `make build`.

### Runtime Safeguards

- Reuse the shared dependencies described in `pkg/shared/context.go` to ensure HTTP clients enforce TLS 1.2+, timeouts, and redaction helpers.
- Never log raw API responses when they may include user content or secrets. Use `RuntimeContext.RedactedCopy()` before logging diagnostic data.
- When shelling out, pass explicit argument arrays and sanitize user inputs.
- Avoid storing AI responses on disk unless the command states why and where the information is persisted.

### Review Checklist

Before merging a change that touches CLI execution paths:

- [ ] Tests demonstrate the user-visible behavior and cover security-sensitive branches.
- [ ] Command documentation reflects configuration, prompts, and network behavior.
- [ ] Command shares code via `pkg/shared` instead of copying helpers across domains.
- [ ] Any new dependencies are vetted for license compatibility and updated in `go.mod`.

## 2. Secure CLI Agent Rules

You are maintaining `magi`, a Go-based CLI that calls AI models. Follow these steps before changing runtime behavior:

1. Load configuration via Viper and construct a `shared.RuntimeContext` (see `pkg/shared/context.go`) instead of duplicating config parsing.
2. Harden external calls: reuse `shared.DefaultHTTPClient()` and define per-command allowlists for file access, HTTP domains, and shell arguments.
3. Log only redacted data and document every outbound request in the command help text.
4. Provide a regression test or recorded fixture whenever you touch prompts, serialization, or encryption.
5. Record any new configuration knobs in `docs/commands.md` and `docs/configuration.md` during the same pull request.

## 3. Shared Code Agent Rules

Use the `/pkg` namespace to host reusable logic. Follow these constraints when contributing shared functionality:

1. **Domain separation.** `/pkg/shared` houses cross-cutting helpers (e.g., configuration loaders, HTTP clients). Domain-specific packages live under `/pkg/<domain>` and must not import command packages.
2. **Stability.** Expose small, intention-revealing APIs. Once exported, treat functions as public commitments—document them and cover them with tests.
3. **Security hardening.** Centralize TLS policies, retry strategies, and redaction helpers inside shared packages. Commands should call these utilities rather than duplicating security checks.
4. **Documentation.** Every shared package gets a package comment plus README snippet describing consumers. Reference the relevant agent files to avoid circular dependencies.
5. **Version awareness.** When shared APIs change, bump the CLI minor version and update all dependents in the same pull request.
6. **Example usage.** Provide small examples (via `_test.go` or doc comments) demonstrating how commands should use the shared helper.

## 4. Command Lifecycle Agent Rules

1. **Design doc first.** Capture the user goal, input/output schema, and security considerations in an issue or doc before writing code.
2. **Single responsibility.** Each command must live in its own file under `cmd/` and may register subcommands via dedicated packages (e.g., `cmd/config`). Avoid mixed concerns.
3. **Shared utilities.** If logic is reused across commands, extract it into `pkg/shared` (see `pkg/shared/context.go`) or another package under `pkg/` before merging.
4. **Flag discipline.** Declare flags in `init()` with clear defaults and `cmd.MarkFlagRequired` where applicable. Validate them in `RunE` handlers.
5. **Telemetry hooks.** Capture minimal analytics (if enabled) via a shared interface so analytics code never pollutes command logic.
6. **Rollout notes.** Update `CHANGELOG` (when available) and docs, highlighting breaking changes and migration paths.

## 5. Command Documentation Agent Rules

Whenever you add or modify a command (see `cmd/setup.go` or `cmd/config.go` for structure), update documentation as follows:

1. **Help text parity.** The `cobra.Command` `Use`, `Short`, and `Long` fields must match the CLI examples recorded in docs.
2. **Docs targets.** Update both `docs/commands.md` and `docs/getting-started.md` (if onboarding flow changes). Mention new flags, configuration keys, prompts, and any security considerations.
3. **Examples.** Provide at least one interactive example and one non-interactive example mirroring `setup` and `config` docs. Use realistic flag values and show expected output or side effects.
4. **Security callouts.** Every command description must state what data leaves the machine (API calls, file uploads) and how secrets are handled.
5. **Versioning.** Append a `Since vX.Y.Z` note for new commands/flags to help downstream package maintainers.
6. **Validation.** Run `go test ./...` and `make docs` (if available) to ensure code samples compile and docs build.

## 6. Command Documentation Agent Rules

Whenever you add or modify a command (see `cmd/setup.go` or `cmd/config.go` for structure), update documentation as follows:

1. **Help text parity.** The `cobra.Command` `Use`, `Short`, and `Long` fields must match the CLI examples recorded in docs.
2. **Docs targets.** Update both `docs/commands.md` and `docs/getting-started.md` (if onboarding flow changes). Mention new flags, configuration keys, prompts, and any security considerations.
3. **Examples.** Provide at least one interactive example and one non-interactive example mirroring `setup` and `config` docs. Use realistic flag values and show expected output or side effects.
4. **Security callouts.** Every command description must state what data leaves the machine (API calls, file uploads) and how secrets are handled.
5. **Versioning.** Append a `Since vX.Y.Z` note for new commands/flags to help downstream package maintainers.
6. **Validation.** Run `go test ./...` and `make docs` (if available) to ensure code samples compile and docs build.

## 7. Testing Agent Rules

1. **Unit-first.** Cover each package with table-driven tests that assert both success and failure paths. Follow the examples in `cmd/version_test.go` for structure.
2. **Security-sensitive tests.** For commands that touch secrets or remote AI providers, include tests that verify secrets are not logged and that prompts redact sensitive fields.
3. **Shared context coverage.** Add tests under `pkg/shared` to capture regression cases (e.g., missing config fields, HTTP client reuse).
4. **Golden files.** When testing AI prompt builders, store prompt templates under `testdata/` and compare outputs using `cmp.Diff`.
5. **CI hooks.** Every PR must pass `go test ./...`, `golangci-lint run` (when enabled), and `make vet`. Document new required tooling in `README.md`.
6. **Fixtures hygiene.** Never hardcode live API keys. Use env var placeholders (`MAGI_TEST_API_KEY`) and skip tests when prerequisites are missing.

## 8. Command Pattern Agent Rules

When implementing or modifying CLI commands, you **MUST** follow the repository's established pattern as defined in `docs/CONTRIBUTING.md`.

### Command Structure Guidelines

Every `cobra.Command` definition must include:

1.  **Use**: A clear usage string (e.g., `command [args]`).
2.  **Short**: A concise summary of what the command does.
3.  **Long**: A detailed description that includes:
    *   Explanation of the command's purpose and behavior.
    *   **Available subcommands** (if applicable), listed with descriptions.
    *   **Usage** section showing the command syntax.
    *   **Examples** section showing common usage scenarios.

### Example Pattern

```go
var cmd = &cobra.Command{
    // Use is the one-line usage message.
    // Recommended syntax is as follows:
    //   [ ] identifies an optional argument. Arguments that are not enclosed in brackets are required.
    //   ... indicates that you can specify multiple values for the previous argument.
    //   |   indicates mutually exclusive information. You can use the argument to the left of the separator or the
    //       argument to the right of the separator. You cannot use both arguments in a single use of the command.
    //   { } delimits a set of mutually exclusive arguments when one of the arguments is required. If the arguments are
    //       optional, they are enclosed in brackets ([ ]).
    // Example: add [-F file | -D dir]... [-f format] profile
    Use:   "my-command [arg]",

    // Short is the short description shown in the 'help' output.
    Short: "Concise summary",

    // Long is the long message shown in the 'help <this-command>' output.
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

    // Run is the function called when the command is executed.
    Run: func(cmd *cobra.Command, args []string) {
        // Default behavior implementation
    },
}
```

### Checklist for New Commands

- [ ] `Use` field is correct and descriptive.
- [ ] `Short` field provides a quick summary.
- [ ] `Long` field follows the structure: Description -> Subcommands -> Usage -> Examples.
- [ ] Examples cover common use cases.
- [ ] Help text is user-friendly and comprehensive.
