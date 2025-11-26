# Commands Reference

This document provides detailed information about all available magi commands.

## Global Flags

- `--config`: Path to config file (default is $HOME/.magi/config.yaml)
- `--author`: Author name for copyright attribution
- `--debug`: Enable debug messages
- `--raw`: Print unstyled raw output
- `--disable-update-checks`: Disables update checks
- `--help`: Help for any command
- `--version`: Display version information

## Core Commands

### setup

Starts an interactive setup wizard for magi.

The setup command starts an interactive wizard to help you configure magi for first use.
It will guide you through setting up your API key and other preferences.

```bash
# Run the interactive setup wizard
magi setup
```

### completion

Generate completion script for your shell.

To load completions:

**Bash:**
```bash
source <(magi completion bash)
```

**Zsh:**
```bash
# If shell completion is not already enabled in your environment:
echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session:
magi completion zsh > "${fpath[1]}/_magi"
```

**Fish:**
```bash
magi completion fish | source

# To load completions for each session:
magi completion fish > ~/.config/fish/completions/magi.fish
```

### analyze

Analyze code and provide insights.

```bash
magi analyze [file/directory] [flags]

Flags:
  -d, --depth int     Analysis depth level (default 1)
  -f, --format string Output format (json|yaml|text) (default "text")
```

### commit _(Since v0.3.0)_

Generate an AI-assisted conventional commit message for staged or selected files and create the commit with a single command. The wizard validates the final summary, shows the hook output if a pre-commit hook blocks the commit, and never persists AI responses on disk.

**Interactive example**
```bash
# Let magi ask which unstaged files should be included
magi commit
```

**Non-interactive example**
```bash
# Stage files manually so magi skips the selection prompt
git add pkg/foo pkg/bar && magi commit
```

Security callout:
- Sends only the git diff for the selected files to your configured AI provider to generate the commit summary; no other file contents or metadata leave the machine.
- Shells out to `git` with explicit arguments and surfaces hook output without logging the full git stdout, protecting secrets printed by hooks.
- Requires a configured AI provider/API key via `magi config` so secrets are never requested ad hoc.

### push _(Since v0.3.0)_

Push the current branch to its upstream remote. magi automatically detects when the branch has no upstream configured and re-runs the push with `--set-upstream` so you only invoke the command once.

**Interactive example**
```bash
# Warns if a pre-push hook is present before delegating to git push
magi push
```

**Non-interactive example**
```bash
# Use in scripts/CI to guarantee --set-upstream is supplied when needed
magi push >/tmp/push.log
```

Security callout:
- Relies entirely on your local git installation; no new data is sent to remote services beyond what git already transmits for a push.
- Warns when a pre-push hook exists and prints the hook output if the hook blocks the push.

### pr _(Since v0.3.0)_

Run an AgenticGoKit review of the local commits that differ from `origin/<branch>`, fill `.github/pull_request_template.md`, and open a GitHub pull request with the `gh` CLI. The command asks for extra context before invoking the agents, prints the generated review and template, and only creates the PR after you confirm.

**Flags:** _(Since v0.4.1)_
- `--dry-run`: Run the agents and output results to the terminal, but do not create a PR.
- `--output-file <path>`: Write the agent results (plan and findings) to a markdown file.
- `--no-comment`: Create the PR but do not add the agent findings as a comment.
- `--only-create`: Create the PR with the filled template but do not add any comments (alias for `--no-comment`).
- `--target-branch <branch>`: Specify the target branch for the Pull Request (defaults to the detected base branch).

**Interactive example**
```bash
# Answer prompts for extra context and confirmation before the PR is created
magi pr
```

**Non-interactive example**
```bash
# Pipe confirmation to run inside scripts while still reviewing diffs securely
yes | magi pr
```

**Dry Run with Output File**
```bash
# Generate the review and save it to a file without creating a PR
magi pr --dry-run --output-file review.md
```

**Target Specific Branch**
```bash
# Create a PR targeting the 'develop' branch
magi pr --target-branch develop
```

Security callout:
- Sends only the diff between `HEAD` and `origin/<branch>` (or target branch), AGENTS.md contents, and any optional user-provided notes to the configured AI provider.
- Uses the hardened HTTP client, enforces TLS 1.2+, and never logs raw model responses that might contain secrets (redacted copies are stored when needed).
- Shells out to `git` and `gh` with explicit argument arrays after confirming the local branch is pushed and sanitized hook output is surfaced.
- Documents outbound data (diff + AGENTS guidelines) in the command help text so users know exactly what leaves their machine.
- Respects configured timeouts for analysis and writing phases (see `magi config`).

## Additional Commands

[More commands will be added as they are implemented]
