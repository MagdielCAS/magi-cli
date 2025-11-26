# magi

## Usage
> A powerful AI-assisted CLI for developers that enhances productivity

magi

## Description

```
magi is a command-line interface tool that leverages AI capabilities
to enhance developer productivity. It provides various commands for code analysis,
documentation, suggestions, and more.

Available Commands:
  setup         Initial setup of magi
  config        Manage magi configuration
  completion    Generate completion script

Usage:
  magi [command]

Examples:
  # Run the setup wizard
  magi setup

  # Configure your API key
  magi config set api.key your-api-key

Run 'magi [command] --help' for more information on a specific command.
```
## Examples

```bash
  magi config set api.key your-api-key
```

## Flags
|Flag|Usage|
|----|-----|
|`--author string`|author name for copyright attribution (default "Magdiel Campelo <github.com/MagdielCAS>")|
|`--config string`|config file (default is $HOME/.magi/config.yaml)|

## Commands
|Command|Usage|
|-------|-----|
|`magi commit`|Generate a conventional commit message with AI and create the commit|
|`magi completion`|Generate completion script|
|`magi config`|Manages the magi configuration|
|`magi help`|Help about any command|
|`magi pr`|Review local commits with AI agents and open a GitHub pull request|
|`magi push`|Push the current branch and auto-configure the upstream if needed|
|`magi setup`|Starts an interactive setup wizard for magi|
|`magi version`|Shows the version of magi|
# ... commit
`magi commit`

## Usage
> Generate a conventional commit message with AI and create the commit

magi commit

## Description

```
commit analyzes your local changes, lets you pick which files to include,
and generates a conventional commit message using the configured AI provider.

Data handling:
  • The command sends the git diff for the selected files to your configured AI provider.
  • No other file contents or metadata leave your machine.

Usage:
  magi commit

Examples:
  # Stage files manually and let magi craft the commit message
  git add pkg/foo && magi commit

  # Select unstaged files interactively and commit them with an AI message
  magi commit

Security note: Requests are performed with the shared hardened HTTP client and only include
the contextual diff needed to craft the message.
```
# ... completion
`magi completion`

## Usage
> Generate completion script

magi completion [bash|zsh|fish]

## Description

```
To load completions:

Bash:
  $ source <(magi completion bash)

Zsh:
  # If shell completion is not already enabled in your environment:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session:
  $ magi completion zsh > "${fpath[1]}/_magi"

Fish:
  $ magi completion fish | source

  # To load completions for each session:
  $ magi completion fish > ~/.config/fish/completions/magi.fish

```
# ... config
`magi config`

## Usage
> Manages the magi configuration

magi config

## Description

```
Manages the magi configuration. You can get, set, list, and reset configuration values.

Available subcommands:
  get     Gets a configuration value
  set     Sets a configuration value
  list    Lists all configuration values
  reset   Resets the configuration
  init    Initialize a local configuration file

Usage:
  magi config [command]

Examples:
  # Get a configuration value
  magi config get api.key

  # Set a configuration value
  magi config set api.heavy_model gpt-4

  # Set the API provider
  magi config set api.provider custom

  # Set the base URL for the custom provider
  magi config set api.base_url http://localhost:8080

  # List all configuration values
  magi config list

  # Reset the configuration
  magi config reset

  # Initialize a local configuration file
  magi config init

Run 'magi config [command] --help' for more information on a specific command.
```

## Commands
|Command|Usage|
|-------|-----|
|`magi config get`|Gets a configuration value|
|`magi config init`|Initialize a local configuration file|
|`magi config list`|Lists all configuration values|
|`magi config reset`|Resets the configuration|
|`magi config set`|Sets a configuration value|
# ... config get
`magi config get`

## Usage
> Gets a configuration value

magi config get [key]

## Description

```
Gets a configuration value.

Usage:
  magi config get [key]

Examples:
  # Get the value of a key
  magi config get api.model

Run 'magi config get --help' for more information on a specific command.
```
# ... config init
`magi config init`

## Usage
> Initialize a local configuration file

magi config init

## Description

```
Initialize a local configuration file (.magi.yaml) in the current directory.
This file will override the global configuration (but only the ones with the same key).
The goal is to have custom envs like models or keys for different workspaces.

Usage:
  magi config init

Examples:
  # Initialize a local configuration file
  magi config init

```
# ... config list
`magi config list`

## Usage
> Lists all configuration values

magi config list

## Description

```
Lists all configuration values.

Usage:
  magi config list

Examples:
  # List all values
  magi config list

Run 'magi config list --help' for more information on a specific command.
```
# ... config reset
`magi config reset`

## Usage
> Resets the configuration

magi config reset

## Description

```
Resets the configuration to its default values.

Usage:
  magi config reset

Examples:
  # Reset the configuration
  magi config reset

Run 'magi config reset --help' for more information on a specific command.
```
# ... config set
`magi config set`

## Usage
> Sets a configuration value

magi config set [key] [value]

## Description

```
Sets a configuration value.

Usage:
  magi config set [key] [value]

Examples:
  # Set the value of a key
  magi config set api.model gpt-4

Run 'magi config set --help' for more information on a specific command.
```
# ... help
`magi help`

## Usage
> Help about any command

magi help [command]

## Description

```
Help provides help for any command in the application.
Simply type magi help [path to command] for full details.
```
# ... pr
`magi pr`

## Usage
> Review local commits with AI agents and open a GitHub pull request

magi pr

## Description

```
Review local commits and create a GitHub pull request.

This command scans your commits that differ from the upstream branch (default: origin/<branch>),
runs an AI-powered review workflow to analyze the diff, fills the repository's pull request template,
and creates the PR using the GitHub CLI ('gh').

Data handling:
  • Sends the git diff between HEAD and origin/<branch>, AGENTS.md contents, and optional user context
    to your configured AI provider.
  • No other files are uploaded.

Security note:
  • The review agents run with a hardened HTTP client.
  • API keys are redacted.
  • Model responses are not persisted unless --output-file is used.
  • Shells out to 'git' and 'gh' with explicit arguments.
```
## Examples

```bash
  # Interactive mode (default)
  magi pr

  # Dry run and save report to a file
  magi pr --dry-run --output-file review.md

  # Target a specific branch
  magi pr --target-branch develop

  # Create PR without commenting findings
  magi pr --no-comment
```

## Flags
|Flag|Usage|
|----|-----|
|`--dry-run`|Run the agents and output results, but do not create a PR|
|`--no-comment`|Do not add the agent findings as a comment to the PR|
|`--only-create`|Create the PR but do not add any comments|
|`--output-file string`|Write the agent results to a markdown file|
|`--target-branch string`|Specify the target branch for the Pull Request|
# ... push
`magi push`

## Usage
> Push the current branch and auto-configure the upstream if needed

magi push

## Description

```
push wraps git push and automatically adds --set-upstream when the current branch
lacks an upstream. The command never sends source data anywhere—it shells out to your
local git binary and surfaces any hook output so you only run push once.

Data handling:
  • The command invokes git locally and does not upload project data on its own.

Usage:
  magi push

Security note: The command respects your git hooks and displays hook output when a push fails.
```
# ... setup
`magi setup`

## Usage
> Starts an interactive setup wizard for magi

magi setup

## Description

```
The setup command starts an interactive wizard to help you configure magi for first use.
It will guide you through setting up your API key and other preferences.
This command can also be run non-interactively by providing the required flags.

Usage:
  magi setup [flags]

Examples:
  # Run the interactive setup wizard
  magi setup

  # Run setup non-interactively with OpenAI
  magi setup --api-provider openai --api-key YOUR_API_KEY --heavy-model gpt-4

  # Run setup non-interactively with a custom provider
  magi setup --api-provider custom --base-url http://localhost:8080 --api-key YOUR_API_KEY --heavy-model custom-model
```

## Flags
|Flag|Usage|
|----|-----|
|`--api-key string`|Your OpenAI API key|
|`--api-provider string`|API provider (e.g., openai, custom)|
|`--base-url string`|Base URL for custom OpenAI compatible API|
|`--fallback-model string`|Fallback model (e.g., gpt-3.5-turbo)|
|`--format string`|Default output format (e.g., text, json, yaml)|
|`--heavy-model string`|Model for heavy tasks (e.g., gpt-4)|
|`--light-model string`|Model for light tasks (e.g., gpt-3.5-turbo)|
# ... version
`magi version`

## Usage
> Shows the version of magi

magi version

## Description

```
Shows the current version of magi, build date, and commit hash.

Usage:
  magi version

Examples:
  # Default behavior
  magi version

  # Show version in JSON format
  magi version --json

Run 'magi version --help' for more information on a specific command.
```


---
> **Documentation automatically generated with [PTerm](https://github.com/pterm/cli-template) on 26 November 2025**
