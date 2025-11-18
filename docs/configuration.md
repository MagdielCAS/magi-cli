# Configuration Guide

magi-cli uses Viper for configuration management. This guide explains all available configuration options and how to set them.

## Configuration File

The default configuration file is located at `$HOME/.magi-cli/config.yaml`. You can specify a different location using the `--config` flag.

### Example Configuration

```yaml
api:
  provider: "openai"
  key: "your-api-key"
  base_url: "https://api.openai.com/v1"
  light_model: "gpt-3.5-turbo"
  heavy_model: "gpt-4"
  fallback_model: "gpt-3.5-turbo"
  light:
    api_key: ""        # Optional override for light calls
    base_url: ""       # Optional override for light calls
  heavy:
    api_key: ""        # Optional override for heavy calls
    base_url: ""       # Optional override for heavy calls
  fallback:
    api_key: ""        # Optional override for fallback calls
    base_url: ""       # Optional override for fallback calls

output:
  format: "text"
  color: true

cache:
  enabled: true
  ttl: 3600
```

## Configuration Options

### API Settings

- `api.provider`: AI provider slug (defaults to `openai`)
- `api.key`: Default API key used for all calls unless overridden
- `api.base_url`: Default base URL for the provider
- `api.light_model`: Model used for "light" requests
- `api.heavy_model`: Model used for "heavy" requests
- `api.fallback_model`: Optional fallback model
- `api.light.api_key`, `api.heavy.api_key`, `api.fallback.api_key`: Optional API key overrides
- `api.light.base_url`, `api.heavy.base_url`, `api.fallback.base_url`: Optional endpoint overrides

### Output Settings

- `output.format`: Default output format (text|json|yaml)
- `output.color`: Enable/disable colored output

### Cache Settings

- `cache.enabled`: Enable/disable response caching
- `cache.ttl`: Cache time-to-live in seconds

## Pull Request Command Settings

The `magi pr` command reuses the API configuration above and additionally expects:

- `.github/pull_request_template.md` to exist so the agent can fill it.
- At least one `AGENTS.md` file if you want repository-specific guardrails enforced during the review.
- The GitHub CLI (`gh`) must be installed and authenticated because it creates the pull request and posts the review comment on your behalf.

No additional configuration keys are required; `magi pr` automatically uses the heavy model for deep review and the light model (when configured) for writing the template. If only one model tier is configured, it is reused for every step.

## Managing Configuration

### Command Line

```bash
# Set a global API key and base URL
magi config set api.key sk-xxx
magi config set api.base_url https://api.openai.com/v1

# Override only the heavy model endpoint
magi config set api.heavy.api_key sk-heavy-only
magi config set api.heavy.base_url https://enterprise-gateway.example.com/v1
```
