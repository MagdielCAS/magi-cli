# Getting Started with magi-cli

This guide will walk you through the installation and basic setup of magi-cli.

## Installation

### Prerequisites

- Go 1.19 or higher
- An OpenAI API key or compatible LLM API key

### Install via Go

```bash
go install github.com/MagdielCAS/magi-cli@latest
```

### Manual Installation

1. Clone the repository:
```bash
git clone https://github.com/MagdielCAS/magi-cli.git
```

2. Build the binary:
```bash
cd magi-cli
go build
```

## Initial Setup

1. Configure your API key:
```bash
magi-cli config set api-key your-api-key
```

2. Verify installation:
```bash
magi-cli --version
```

## Basic Usage

Here are some common commands to get you started:

```bash
# Get help
magi-cli --help

# Generate code documentation
magi-cli doc generate ./path/to/file

# Analyze code
magi-cli analyze ./path/to/file

# Get code suggestions
magi-cli suggest ./path/to/file
```

## Next Steps

- Read the [Commands Reference](./commands.md) for detailed information about available commands
- Learn about [Configuration](./configuration.md) options

