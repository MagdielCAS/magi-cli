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

## Additional Commands

[More commands will be added as they are implemented]