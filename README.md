<h1 align="center">magi</h1>
<p align="center">A powerful AI-assisted CLI for developers that enhances productivity</p>

<p align="center">

<a style="text-decoration: none" href="https://github.com/MagdielCAS/magi-cli/releases">
<img src="https://img.shields.io/github/v/release/MagdielCAS/magi-cli?style=flat-square" alt="Latest Release">
</a>

<a style="text-decoration: none" href="https://github.com/MagdielCAS/magi-cli/releases">
<img src="https://img.shields.io/github/downloads/MagdielCAS/magi-cli/total.svg?style=flat-square" alt="Downloads">
</a>

<a style="text-decoration: none" href="https://github.com/MagdielCAS/magi-cli/stargazers">
<img src="https://img.shields.io/github/stars/MagdielCAS/magi-cli.svg?style=flat-square" alt="Stars">
</a>

<a style="text-decoration: none" href="https://github.com/MagdielCAS/magi-cli/fork">
<img src="https://img.shields.io/github/forks/MagdielCAS/magi-cli.svg?style=flat-square" alt="Forks">
</a>

<a style="text-decoration: none" href="https://github.com/MagdielCAS/magi-cli/issues">
<img src="https://img.shields.io/github/issues/MagdielCAS/magi-cli.svg?style=flat-square" alt="Issues">
</a>

<a style="text-decoration: none" href="https://opensource.org/licenses/BSD-2-Clause">
<img src="https://img.shields.io/badge/License-BSD_2--Clause-orange.svg?style=flat-square" alt="License: BSD-2">
</a>

<br/>

<a style="text-decoration: none" href="https://github.com/MagdielCAS/magi-cli/releases">
<img src="https://img.shields.io/badge/platform-windows%20%7C%20macos%20%7C%20linux-informational?style=for-the-badge" alt="Downloads">
</a>

<br/>

</p>

----

<p align="center">
<strong><a href="https://MagdielCAS.github.io/magi-cli/#/installation">Installation</a></strong>
|
<strong><a href="https://MagdielCAS.github.io/magi-cli/#/docs">Documentation</a></strong>
|
<strong><a href="https://MagdielCAS.github.io/magi-cli/#/CONTRIBUTING">Contributing</a></strong>
</p>

----

## Demo

![magi-cli animation](docs/_assets/magi-cli-animation.svg)

magi-cli is a command-line interface tool designed to enhance programmer productivity by leveraging AI capabilities. Built with Go, it provides a suite of commands that help automate and streamline common programming tasks.

## Key Features

- AI-assisted code generation
- Smart code analysis
- Project scaffolding
- Documentation assistance
- Code review suggestions
- Cryptographic utilities
- SSH connection management
- AI-powered i18n translation
- And more...

## Installation

**Windows**
```powershell
iwr instl.sh/MagdielCAS/magi-cli/windows | iex
```

**macOS**
```bash
curl -sSL instl.sh/MagdielCAS/magi-cli/macos | bash
```

**Linux**
```bash
curl -sSL instl.sh/MagdielCAS/magi-cli/linux | bash
```

## Quick Start

1. Set up your OpenAI API key:
```bash
magi config set api-key your-api-key
```

2. Verify installation:
```bash
magi --version
```

## Usage

```bash
magi [command] [flags]
```

For detailed documentation of all available commands:
```bash
magi --help
```

## Configuration

magi uses a configuration file located at `$HOME/.magi/config.yaml`. You can modify settings using:

```bash
magi config [key] [value]
```

You can also create a local configuration file for project-specific settings:

```bash
magi config init
```

## Support

- 📫 Report issues on [GitHub Issues](https://github.com/MagdielCAS/magi-cli/issues)
- 💬 Join discussions in [GitHub Discussions](https://github.com/MagdielCAS/magi-cli/discussions)

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra)
- Configuration managed by [Viper](https://github.com/spf13/viper)
- Powered by OpenAI compatible APIs