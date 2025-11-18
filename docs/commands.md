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

### commit

Generate an AI-assisted conventional commit message for staged or selected files and create the commit.

```bash
# Use currently staged files
magi commit

# Select unstaged files interactively and let magi commit them
magi commit
```

Security callout:
- Sends only the git diff for the selected files to your configured AI provider in order to craft the commit message.
- No other project metadata or secrets leave your machine.
- Warns when a local pre-commit hook is detected and surfaces the hook output if it blocks the commit so you can fix the reported issues.

### push

Push the current branch to its upstream remote. magi automatically detects when the branch has no upstream configured and re-runs the push with `--set-upstream` so you only invoke the command once.

```bash
# Push normally; magi will add --set-upstream the first time
magi push
```

Security callout:
- Relies entirely on your local git installation; no new data is sent to remote services beyond what git already transmits for a push.
- Warns when a pre-push hook exists and prints the hook output if the hook blocks the push.

## Additional Commands

[More commands will be added as they are implemented]
