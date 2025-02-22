# Configuration Guide

magi-cli uses Viper for configuration management. This guide explains all available configuration options and how to set them.

## Configuration File

The default configuration file is located at `$HOME/.magi-cli/config.yaml`. You can specify a different location using the `--config` flag.

### Example Configuration

```yaml
api:
  key: "your-api-key"
  endpoint: "https://api.openai.com/v1"
  model: "gpt-4"

defaults:
  temperature: 0.7
  max_tokens: 1000
  context_lines: 5

output:
  format: "text"
  color: true
  verbose: false

cache:
  enabled: true
  ttl: 3600
  max_size: "100MB"
```

## Configuration Options

### API Settings

- `api.key`: Your OpenAI API key or compatible API key
- `api.endpoint`: API endpoint URL
- `api.model`: Default model to use

### Default Parameters

- `defaults.temperature`: AI response creativity (0.0-1.0)
- `defaults.max_tokens`: Maximum tokens per request
- `defaults.context_lines`: Default context lines for code analysis

### Output Settings

- `output.format`: Default output format (text|json|yaml)
- `output.color`: Enable/disable colored output
- `output.verbose`: Enable/disable verbose logging

### Cache Settings

- `cache.enabled`: Enable/disable response caching
- `cache.ttl`: Cache time-to-live in seconds
- `cache.max_size`: Maximum cache size

## Managing Configuration

### Command Line

```bash
# Set a configuration value
magi-cli config set api.key your-api-key

# Get current configuration
magi-cli config get

# Reset to defaults
magi-cli config reset
```

### Environment Variables

You can override configuration using environment variables:

```bash
export MAGI_API_KEY=your-api-key
export MAGI_OUTPUT_FORMAT=json
```