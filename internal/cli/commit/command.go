/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package commit

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MagdielCAS/magi-cli/pkg/git"
	"github.com/MagdielCAS/magi-cli/pkg/llm"
	"github.com/MagdielCAS/magi-cli/pkg/shared"
	"github.com/MagdielCAS/magi-cli/pkg/utils"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Generate a conventional commit message with AI and create the commit",
	Long: `commit analyzes your local changes, lets you pick which files to include,
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
the contextual diff needed to craft the message.`,
	RunE: runCommit,
}

func CommitCmd() *cobra.Command {
	return commitCmd
}

func runCommit(cmd *cobra.Command, _ []string) error {
	if err := git.EnsureGitRepo(cmd.Context()); err != nil {
		return err
	}

	runtimeCtx, err := shared.BuildRuntimeContext()
	if err != nil {
		return err
	}

	staged, err := listGitFiles(cmd.Context(), true)
	if err != nil {
		return err
	}

	var targetFiles []string
	switch {
	case len(staged) > 0:
		pterm.Info.Printf("Detected %d staged file(s); skipping selection UI.\n", len(staged))
		targetFiles = staged
	default:
		targetFiles, err = promptForUnstagedFiles(cmd.Context())
		if err != nil {
			return err
		}
	}

	if len(targetFiles) == 0 {
		return errors.New("no files selected for commit")
	}

	diff, err := diffAgainstOrigin(cmd.Context(), targetFiles)
	if err != nil {
		return err
	}

	pterm.Info.Printf("Using light model: %s\n", runtimeCtx.LightModel)
	pterm.Info.Println("Generating commit message with the configured AI provider...")

	message, err := llm.GenerateCommitMessage(cmd.Context(), runtimeCtx, diff)
	if err != nil {
		return err
	}

	message = utils.RemoveCodeBlock(message)
	message = normalizeCommitMessage(message)
	if validationErr := validateCommitFormat(message); validationErr != nil {
		pterm.Warning.Printf("Generated commit message failed validation: %v. Retrying with guidance...\n", validationErr)
		originalMessage := message
		if fixedMessage, err := retryCommitMessage(cmd.Context(), runtimeCtx, diff, message, validationErr); err == nil {
			message = fixedMessage
		} else {
			pterm.Error.PrintOnError(err)
			message = originalMessage
		}
	}

	pterm.DefaultBox.WithTitle("Suggested Commit Message").Println(message)

	confirmed, err := pterm.DefaultInteractiveConfirm.WithDefaultValue(true).
		Show("Use this commit message?")
	if err != nil {
		return fmt.Errorf("confirmation prompt failed: %w", err)
	}
	if !confirmed {
		pterm.Warning.Println("Commit aborted by user.")
		return nil
	}

	if err := gitCommit(cmd.Context(), message); err != nil {
		return err
	}

	pterm.Success.Println("Commit created successfully.")
	return nil
}

func listGitFiles(ctx context.Context, staged bool) ([]string, error) {
	args := []string{"diff", "--name-only"}
	if staged {
		args = append(args, "--cached")
	}

	output, err := git.RunGit(ctx, args...)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func promptForUnstagedFiles(ctx context.Context) ([]string, error) {
	statusEntries, err := gitStatusEntries(ctx)
	if err != nil {
		return nil, err
	}

	if len(statusEntries) == 0 {
		return nil, errors.New("no unstaged files detected")
	}

	options := make([]string, 0, len(statusEntries))
	displayToPath := make(map[string]string, len(statusEntries))
	for _, entry := range statusEntries {
		display := fmt.Sprintf("%s %s", entry.Status, entry.Path)
		options = append(options, display)
		displayToPath[display] = entry.Path
	}

	selected, err := pterm.DefaultInteractiveMultiselect.
		WithOptions(options).
		WithDefaultText("Select files to include in the commit").
		Show()
	if err != nil {
		return nil, fmt.Errorf("file selection failed: %w", err)
	}

	if len(selected) == 0 {
		return nil, errors.New("no files were selected")
	}

	var paths []string
	for _, label := range selected {
		if path := displayToPath[label]; path != "" {
			paths = append(paths, path)
		}
	}

	if err := gitAdd(ctx, paths); err != nil {
		return nil, err
	}

	return paths, nil
}

type statusEntry struct {
	Status string
	Path   string
}

func gitStatusEntries(ctx context.Context) ([]statusEntry, error) {
	output, err := git.RunGit(ctx, "status", "--short", "--untracked-files")
	if err != nil {
		return nil, err
	}

	var entries []statusEntry
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, " \r")
		if strings.TrimSpace(line) == "" {
			continue
		}

		if len(line) < 3 {
			continue
		}

		status := line[:2]
		path := strings.TrimSpace(line[3:])

		if len(status) == 2 && (status == "??" || status[1] != ' ') {
			entries = append(entries, statusEntry{
				Status: status,
				Path:   path,
			})
		}
	}

	return entries, nil
}

func gitAdd(ctx context.Context, files []string) error {
	args := append([]string{"add", "--"}, files...)
	_, err := git.RunGit(ctx, args...)
	return err
}

func diffAgainstOrigin(ctx context.Context, files []string) (string, error) {
	currentBranch, err := git.CurrentBranchName(ctx)
	if err != nil {
		return "", err
	}

	remoteRef := fmt.Sprintf("origin/%s", currentBranch)
	if _, err := git.RunGit(ctx, "rev-parse", "--verify", remoteRef); err != nil {
		return "", fmt.Errorf("unable to find %s. Fetch the branch from origin and try again: %w", remoteRef, err)
	}

	args := []string{"diff", "--cached", remoteRef, "--"}
	args = append(args, files...)

	diff, err := git.RunGit(ctx, args...)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(diff) == "" {
		return "", errors.New("diff against origin is empty")
	}

	return diff, nil
}

func gitCommit(ctx context.Context, message string) error {
	hasHook, hookPath, hookErr := git.HasGitHook(ctx, "pre-commit")
	if hookErr != nil {
		pterm.Warning.Printf("Unable to determine pre-commit hooks: %v\n", hookErr)
	} else if hasHook {
		pterm.Warning.Printf("Detected pre-commit hook at %s. Hook output will be shown if it fails.\n", hookPath)
	}

	result, err := git.RunGitRaw(ctx, "commit", "-m", message)
	if err != nil {
		git.LogGitFailure(err)
		if hasHook && hookErr == nil {
			pterm.Warning.Println("Pre-commit hook blocked the commit. Fix the issues it reported and retry.")
		}
		return err
	}

	if strings.TrimSpace(result.Stdout) != "" {
		pterm.Info.Println("git commit completed with additional output omitted to protect sensitive data.")
	}
	pterm.Success.Println("Commit recorded.")

	return nil
}

func normalizeCommitMessage(message string) string {
	if idx := strings.Index(message, "\n"); idx != -1 {
		message = message[:idx]
	}
	return strings.TrimSpace(message)
}

func validateCommitFormat(message string) error {
	if message == "" {
		return errors.New("commit message is empty")
	}

	typeEnd := strings.Index(message, ":")
	if typeEnd == -1 {
		return errors.New("missing ':' separator")
	}

	meta := message[:typeEnd]
	if !strings.Contains(meta, "(") || !strings.HasSuffix(meta, ")") {
		return errors.New("missing scope in commit message")
	}

	scopeStart := strings.Index(meta, "(")
	commitType := meta[:scopeStart]
	if !isAllowedCommitType(commitType) {
		return fmt.Errorf("unsupported commit type %q", commitType)
	}

	payload := strings.TrimSpace(message[typeEnd+1:])
	if payload == "" {
		return errors.New("missing commit description")
	}

	fields := strings.Fields(payload)
	if len(fields) < 2 {
		return errors.New("commit description must include an emoji and summary text")
	}
	if !isAllowedGitmoji(fields[0]) {
		return fmt.Errorf("unsupported gitmoji %q", fields[0])
	}

	return nil
}

func isAllowedCommitType(commitType string) bool {
	switch commitType {
	case "feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert":
		return true
	default:
		return false
	}
}

var allowedGitmoji = map[string]struct{}{
	"✨":  {},
	"🐛":  {},
	"📚":  {},
	"🎨":  {},
	"♻️": {},
	"⚡️": {},
	"✅":  {},
	"🔧":  {},
	"👷":  {},
	"🔨":  {},
	"⏪️": {},
}

func isAllowedGitmoji(emoji string) bool {
	_, ok := allowedGitmoji[emoji]
	return ok
}

func retryCommitMessage(ctx context.Context, runtimeCtx *shared.RuntimeContext, diff, previous string, validationErr error) (string, error) {
	fixed, err := llm.FixCommitMessage(ctx, runtimeCtx, diff, previous, validationErr)
	if err != nil {
		return "", fmt.Errorf("unable to refine commit message after validation failure: %w", err)
	}

	fixed = utils.RemoveCodeBlock(fixed)
	fixed = normalizeCommitMessage(fixed)
	if err := validateCommitFormat(fixed); err != nil {
		return "", fmt.Errorf("ai failed to produce a valid commit message after refinement: %w", err)
	}

	return fixed, nil
}
