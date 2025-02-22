# Commands Reference

This document provides detailed information about all available magi-cli commands.

## Global Flags

- `--config`: Path to config file (default is $HOME/.magi-cli/config.yaml)
- `--verbose`: Enable verbose output
- `--help`: Help for any command
- `--version`: Display version information

## Core Commands

### config

Manage magi-cli configuration.

```bash
# Set API key
magi-cli config set api-key <your-api-key>

# Get current configuration
magi-cli config get

# Reset configuration
magi-cli config reset
```

### analyze

Analyze code and provide insights.

```bash
magi-cli analyze [file/directory] [flags]

Flags:
  -d, --depth int     Analysis depth level (default 1)
  -f, --format string Output format (json|yaml|text) (default "text")
```

### suggest

Get AI-powered code suggestions.

```bash
magi-cli suggest [file] [flags]

Flags:
  -c, --context int   Lines of context to include (default 5)
  -t, --type string   Suggestion type (refactor|optimize|secure)
```

### doc

Generate or modify documentation.

```bash
magi-cli doc [command] [flags]

Available Commands:
  generate    Generate documentation for code
  update      Update existing documentation
```

## Additional Commands

[More commands will be added as they are implemented]