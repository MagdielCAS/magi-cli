---
trigger: always_on
---
# Security Rules

## 1. Secrets & Data

- **Least Privilege**: Request minimal scopes/permissions.
- **No Hardcoding**: API keys/tokens must come from Viper config or secure storage.
- **Redaction**: ANY logging of requests/responses MUST be redacted.
  - Use `RuntimeContext.RedactedCopy()` if available.

## 2. Boundaries

- **Documentation**: Command help text (`Long` description) MUST state if data leaves the machine.
- **HTTPS**: All external calls must use TLS 1.2+ (enforced via shared HTTP client).
