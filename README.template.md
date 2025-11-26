<h1 align="center">{{ .Name }}</h1>
<p align="center">{{ .Short }}</p>

<p align="center">

<a style="text-decoration: none" href="https://github.com/{{ .ProjectPath }}/releases">
<img src="https://img.shields.io/github/v/release/{{ .ProjectPath }}?style=flat-square" alt="Latest Release">
</a>

<a style="text-decoration: none" href="https://github.com/{{ .ProjectPath }}/releases">
<img src="https://img.shields.io/github/downloads/{{ .ProjectPath }}/total.svg?style=flat-square" alt="Downloads">
</a>

<a style="text-decoration: none" href="https://github.com/{{ .ProjectPath }}/stargazers">
<img src="https://img.shields.io/github/stars/{{ .ProjectPath }}.svg?style=flat-square" alt="Stars">
</a>

<a style="text-decoration: none" href="https://github.com/{{ .ProjectPath }}/fork">
<img src="https://img.shields.io/github/forks/{{ .ProjectPath }}.svg?style=flat-square" alt="Forks">
</a>

<a style="text-decoration: none" href="https://github.com/{{ .ProjectPath }}/issues">
<img src="https://img.shields.io/github/issues/{{ .ProjectPath }}.svg?style=flat-square" alt="Issues">
</a>

<a style="text-decoration: none" href="https://opensource.org/licenses/BSD-2-Clause">
<img src="https://img.shields.io/badge/License-BSD_2--Clause-orange.svg?style=flat-square" alt="License: BSD-2">
</a>

<br/>

<a style="text-decoration: none" href="https://github.com/{{ .ProjectPath }}/releases">
<img src="https://img.shields.io/badge/platform-windows%20%7C%20macos%20%7C%20linux-informational?style=for-the-badge" alt="Downloads">
</a>

<br/>

</p>

----

<p align="center">
<strong><a href="{{ .GitHubPagesURL }}/#/installation">Installation</a></strong>
|
<strong><a href="{{ .GitHubPagesURL }}/#/docs">Documentation</a></strong>
|
<strong><a href="{{ .GitHubPagesURL }}/#/CONTRIBUTING">Contributing</a></strong>
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
- And more...

## Installation

**Windows**
```powershell
{{ .InstallCommandWindows }}
```

**macOS**
```bash
{{ .InstallCommandMacOS }}
```

**Linux**
```bash
{{ .InstallCommandLinux }}
```

## Quick Start

1. Set up your OpenAI API key:
```bash
{{ .Name }} config set api-key your-api-key
```

2. Verify installation:
```bash
{{ .Name }} --version
```

## Usage

```bash
{{ .Name }} [command] [flags]
```

For detailed documentation of all available commands:
```bash
{{ .Name }} --help
```

## Configuration

{{ .Name }} uses a configuration file located at `$HOME/.{{ .Name }}/config.yaml`. You can modify settings using:

```bash
{{ .Name }} config [key] [value]
```

You can also create a local configuration file for project-specific settings:

```bash
{{ .Name }} config init
```

## Support

- 📫 Report issues on [GitHub Issues](https://github.com/{{ .ProjectPath }}/issues)
- 💬 Join discussions in [GitHub Discussions](https://github.com/{{ .ProjectPath }}/discussions)

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra)
- Configuration managed by [Viper](https://github.com/spf13/viper)
- Powered by OpenAI compatible APIs