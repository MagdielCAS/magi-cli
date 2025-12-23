---
trigger: always_on
---
# Coding Standards & Patterns

## 1. Output & Logging (**STRICT**)

- **User Output**: MUST use `github.com/pterm/pterm`.
  - ✅ `pterm.Success.Println(...)`
  - ✅ `pterm.Error.Printf(...)`
  - ❌ `fmt.Println`, `log.Println` (Forbidden for user-facing output).
- **Interactivity**:
  - Check `utils.IsInteractive()` before prompting.
  - Use `pterm.InteractiveSelectPrinter` etc.

## 2. Configuration (**STRICT**)

- **Source**: MUST use `github.com/spf13/viper`.
- **Pattern**: `viper.GetString("key")`.
- **Forbidden**: `os.Getenv` for app config.

## 3. Error Handling

- **Cobra**: Use `RunE` signatures.
- **Return**: Bubbling errors up is preferred. Wrap with context: `fmt.Errorf("context: %w", err)`.
- **Exit**: Do NOT call `os.Exit` in command logic (only `main.go`).

## 4. Testing

- **Table-Driven**: Use table-driven tests for all logic.
- **Naming**: `Test<Function>_<Scenario>`.
- **Golden Files**: Use `cmp.Diff` for complex output comparison.
