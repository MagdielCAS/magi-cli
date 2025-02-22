magi-cli
------

A simple CLI tool built with Go that leverages OpenAI compatible LLM APIs to help me with my day-to-day tasks as a programmer

------

## Features

- 🚀 Quick setup with simple configuration
- 🤖 AI-powered assistance for common programming tasks
- ⚡ Fast and lightweight
- 🛠️ Extensible command structure
- 🔒 Secure API key management

## Installation

```bash
go install github.com/MagdielCAS/magi-cli@latest
```

## Quick Start

1. Set up your OpenAI API key:
```bash
magi-cli config set api-key your-api-key
```

2. Verify installation:
```bash
magi-cli --version
```

## Usage

```bash
magi-cli [command] [flags]
```

For detailed documentation of all available commands:
```bash
magi-cli --help
```

## Configuration

magi-cli uses a configuration file located at `$HOME/.magi-cli/config.yaml`. You can modify settings using:

```bash
magi-cli config [key] [value]
```

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on how to submit pull requests, report issues, and contribute to the project.

## License

This project is licensed under the BSD 2-Clause License - see the [LICENSE](LICENSE) file for details.

## Support

- 📫 Report issues on [GitHub Issues](https://github.com/MagdielCAS/magi-cli/issues)
- 💬 Join discussions in [GitHub Discussions](https://github.com/MagdielCAS/magi-cli/discussions)

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra)
- Configuration managed by [Viper](https://github.com/spf13/viper)
- Powered by OpenAI compatible APIs
