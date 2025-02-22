# magi-cli

## Usage
> This cli template shows the date and time in the terminal

magi-cli

## Description

```
This is a template CLI application, which can be used as a boilerplate for awesome CLI tools written in Go.
This template prints the date or time to the terminal.
```
## Examples

```bash
magi-cli date
magi-cli date --format 20060102
magi-cli time
magi-cli time --live
```

## Flags
|Flag|Usage|
|----|-----|
|`--debug`|enable debug messages|
|`--disable-update-checks`|disables update checks|
|`--raw`|print unstyled raw output (set it if output is written to a file)|

## Commands
|Command|Usage|
|-------|-----|
|`magi-cli completion`|Generate the autocompletion script for the specified shell|
|`magi-cli date`|Prints the current date.|
|`magi-cli help`|Help about any command|
|`magi-cli time`|Prints the current time|
# ... completion
`magi-cli completion`

## Usage
> Generate the autocompletion script for the specified shell

magi-cli completion

## Description

```
Generate the autocompletion script for magi-cli for the specified shell.
See each sub-command's help for details on how to use the generated script.

```

## Commands
|Command|Usage|
|-------|-----|
|`magi-cli completion bash`|Generate the autocompletion script for bash|
|`magi-cli completion fish`|Generate the autocompletion script for fish|
|`magi-cli completion powershell`|Generate the autocompletion script for powershell|
|`magi-cli completion zsh`|Generate the autocompletion script for zsh|
# ... completion bash
`magi-cli completion bash`

## Usage
> Generate the autocompletion script for bash

magi-cli completion bash

## Description

```
Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(magi-cli completion bash)

To load completions for every new session, execute once:

#### Linux:

	magi-cli completion bash > /etc/bash_completion.d/magi-cli

#### macOS:

	magi-cli completion bash > $(brew --prefix)/etc/bash_completion.d/magi-cli

You will need to start a new shell for this setup to take effect.

```

## Flags
|Flag|Usage|
|----|-----|
|`--no-descriptions`|disable completion descriptions|
# ... completion fish
`magi-cli completion fish`

## Usage
> Generate the autocompletion script for fish

magi-cli completion fish

## Description

```
Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	magi-cli completion fish | source

To load completions for every new session, execute once:

	magi-cli completion fish > ~/.config/fish/completions/magi-cli.fish

You will need to start a new shell for this setup to take effect.

```

## Flags
|Flag|Usage|
|----|-----|
|`--no-descriptions`|disable completion descriptions|
# ... completion powershell
`magi-cli completion powershell`

## Usage
> Generate the autocompletion script for powershell

magi-cli completion powershell

## Description

```
Generate the autocompletion script for powershell.

To load completions in your current shell session:

	magi-cli completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.

```

## Flags
|Flag|Usage|
|----|-----|
|`--no-descriptions`|disable completion descriptions|
# ... completion zsh
`magi-cli completion zsh`

## Usage
> Generate the autocompletion script for zsh

magi-cli completion zsh

## Description

```
Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(magi-cli completion zsh)

To load completions for every new session, execute once:

#### Linux:

	magi-cli completion zsh > "${fpath[1]}/_magi-cli"

#### macOS:

	magi-cli completion zsh > $(brew --prefix)/share/zsh/site-functions/_magi-cli

You will need to start a new shell for this setup to take effect.

```

## Flags
|Flag|Usage|
|----|-----|
|`--no-descriptions`|disable completion descriptions|
# ... date
`magi-cli date`

## Usage
> Prints the current date.

magi-cli date

## Flags
|Flag|Usage|
|----|-----|
|`-f, --format string`|specify a custom date format (default "02 Jan 06")|
# ... help
`magi-cli help`

## Usage
> Help about any command

magi-cli help [command]

## Description

```
Help provides help for any command in the application.
Simply type magi-cli help [path to command] for full details.
```
# ... time
`magi-cli time`

## Usage
> Prints the current time

magi-cli time

## Description

```
You can print a live clock with the '--live' flag!
```

## Flags
|Flag|Usage|
|----|-----|
|`-l, --live`|live output|


---
> **Documentation automatically generated with [PTerm](https://github.com/pterm/cli-template) on 22 February 2025**
