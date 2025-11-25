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
    provider: ""       # Optional provider override (e.g., openrouter)
  heavy:
    api_key: ""        # Optional override for heavy calls
    base_url: ""       # Optional override for heavy calls
    provider: ""       # Optional provider override
  fallback:
    api_key: ""        # Optional override for fallback calls
    base_url: ""       # Optional override for fallback calls
    provider: ""       # Optional provider override

output:
  format: "text"
  color: true

cache:
  enabled: true
  ttl: 3600
```

## Configuration Options

### API Settings _(Since v0.3.0)_

- `api.provider`: Primary AI provider slug (defaults to `openai`).
- `api.key`: Default API key used for all calls unless overridden.
- `api.base_url`: Default base URL for the provider.
- `api.light_model`: Model used for "light" requests such as PR template writing.
- `api.heavy_model`: Model used for "heavy" analysis (diff reviews, commit generation).
- `api.fallback_model`: Optional fallback when a primary tier is missing.
- `api.light.api_key`, `api.heavy.api_key`, `api.fallback.api_key`: Optional API key overrides per tier so you can scope credentials to least-privilege roles.
- `api.light.base_url`, `api.heavy.base_url`, `api.fallback.base_url`: Optional endpoint overrides (e.g., Azure OpenAI, OpenRouter) per tier.
- `api.light.provider`, `api.heavy.provider`, `api.fallback.provider`: Optional provider overrides per tier when different vendor slugs are required.

### Output Settings

- `output.format`: Default output format (text|json|yaml)
- `output.color`: Enable/disable colored output

### Cache Settings

- `cache.enabled`: Enable/disable response caching
- `cache.ttl`: Cache time-to-live in seconds

### Agent Settings _(Since v0.4.0)_

- `agent.analysis.timeout`: Timeout for the analysis agent (default `3m`).
- `agent.writer.timeout`: Timeout for the writer agent (default `2m`).

## Pull Request Command Settings _(Since v0.3.0)_

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

# Set agent timeouts
magi config set agent.analysis.timeout 5m
magi config set agent.writer.timeout 3m
```
