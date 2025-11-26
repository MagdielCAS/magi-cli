# Contributing to magi-cli

We love your input! We want to make contributing to magi-cli as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## Development Process

We use GitHub to host code, to track issues and feature requests, as well as accept pull requests.

1. Fork the repo and create your branch from `main`
2. If you've added code that should be tested, add tests
3. If you've changed APIs, update the documentation
4. Ensure the test suite passes
5. Make sure your code lints
6. Issue that pull request!

## Follow the Code Pattern

Follow the established pattern used in commands, adhering to [CLI Guidelines](https://clig.dev/#guidelines). Here's what to include:

**Command Structure:**

- Use descriptive `Use`, `Short`, and `Long` fields
- Provide clear usage examples
- Include comprehensive help text
- Follow the pattern of having subcommands when appropriate
- Implement a default behavior when no subcommand is specified
- **Output:** Use stdout for primary output and stderr for diagnostic/error messages
- **Configuration:** Support configuration via flags, environment variables, and config files (in that order of precedence)
- **Interactivity:** Detect if the output is a TTY before prompting for input or using colors

**Example from crypto command:**
```go
var newCmd = &cobra.Command{
    Use:   "new-command",
    Short: "Short description",
    Long: `Detailed command description.

Available subcommands:
  subcommand1     Description of subcommand1
  subcommand2     Description of subcommand2

Usage:
  magi new-command [command]

Examples:
  # Default behavior
  magi new-command

  # Using some subcommand
  magi new-command subcommand1

Run 'magi new-command [command] --help' for more information on a specific command.`,
    Run: func(cmd *cobra.Command, args []string) {
        // Default behavior implementation
    },
}
```

### Command Documentation

Ensure the command includes:

- Clear and concise descriptions
- Information about subcommands
- Default behavior explanation
- Usage examples

## Pull Request Process

1. Update the README.md with details of changes to the interface, if applicable
2. Update the docs/ directory with any new documentation
3. The PR will be merged once you have the sign-off of at least one other developer

## Any contributions you make will be under the BSD 2-Clause License

In short, when you submit code changes, your submissions are understood to be under the same [BSD 2-Clause License](LICENSE) that covers the project. Feel free to contact the maintainers if that's a concern.

## Report bugs using GitHub's [issue tracker](https://github.com/MagdielCAS/magi-cli/issues)

We use GitHub issues to track public bugs. Report a bug by [opening a new issue](https://github.com/MagdielCAS/magi-cli/issues/new); it's that easy!

## Write bug reports with detail, background, and sample code

**Great Bug Reports** tend to have:

- A quick summary and/or background
- Steps to reproduce
  - Be specific!
  - Give sample code if you can
- What you expected would happen
- What actually happens
- Notes (possibly including why you think this might be happening, or stuff you tried that didn't work)

## Developer Certificate of Origin

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.

## License

By contributing, you agree that your contributions will be licensed under its BSD 2-Clause License.

## References

This document was adapted from the open-source contribution guidelines for [Facebook's Draft](https://github.com/facebook/draft-js/blob/master/CONTRIBUTING.md).