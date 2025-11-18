/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

func init() {
	rootCmd.AddCommand(commitCmd)
}

func runCommit(cmd *cobra.Command, _ []string) error {
	if err := ensureGitRepo(cmd.Context()); err != nil {
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

	pterm.Info.Println("Generating commit message with the configured AI provider...")
	message, err := llm.GenerateCommitMessage(cmd.Context(), runtimeCtx, diff)
	if err != nil {
		return err
	}

	message = utils.RemoveCodeBlock(message)
	message = normalizeCommitMessage(message)
	if err := validateCommitFormat(message); err != nil {
		pterm.Warning.Printf("Generated commit message failed validation: %v. Retrying with guidance...\n", err)
		message, err = retryCommitMessage(cmd.Context(), runtimeCtx, diff, message, err)
		if err != nil {
			pterm.Error.PrintOnError(err)
			message = utils.RemoveCodeBlock(message)
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

func ensureGitRepo(ctx context.Context) error {
	_, err := runGit(ctx, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return fmt.Errorf("this command must be run inside a git repository: %w", err)
	}
	return nil
}

func listGitFiles(ctx context.Context, staged bool) ([]string, error) {
	args := []string{"diff", "--name-only"}
	if staged {
		args = append(args, "--cached")
	}

	output, err := runGit(ctx, args...)
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
	output, err := runGit(ctx, "status", "--short", "--untracked-files")
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
	_, err := runGit(ctx, args...)
	return err
}

func diffAgainstOrigin(ctx context.Context, files []string) (string, error) {
	currentBranch, err := currentBranchName(ctx)
	if err != nil {
		return "", err
	}

	remoteRef := fmt.Sprintf("origin/%s", currentBranch)
	if _, err := runGit(ctx, "rev-parse", "--verify", remoteRef); err != nil {
		return "", fmt.Errorf("unable to find %s. Fetch the branch from origin and try again: %w", remoteRef, err)
	}

	args := []string{"diff", "--cached", remoteRef, "--"}
	args = append(args, files...)

	diff, err := runGit(ctx, args...)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(diff) == "" {
		return "", errors.New("diff against origin is empty")
	}

	return diff, nil
}

func currentBranchName(ctx context.Context) (string, error) {
	output, err := runGit(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to determine current branch: %w", err)
	}

	branch := strings.TrimSpace(output)
	if branch == "" {
		return "", errors.New("current branch name is empty")
	}

	return branch, nil
}

func branchRemote(ctx context.Context, branch string) string {
	output, err := runGit(ctx, "config", fmt.Sprintf("branch.%s.remote", branch))
	if err != nil {
		return "origin"
	}

	remote := strings.TrimSpace(output)
	if remote == "" {
		return "origin"
	}

	return remote
}

func gitHooksDir(ctx context.Context) (string, error) {
	output, err := runGit(ctx, "rev-parse", "--path-format=absolute", "--git-path", "hooks")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

func hasGitHook(ctx context.Context, hookName string) (bool, string, error) {
	hooksDir, err := gitHooksDir(ctx)
	if err != nil {
		return false, "", err
	}

	path := filepath.Join(hooksDir, hookName)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, path, nil
		}
		return false, "", err
	}

	if info.IsDir() {
		return false, path, nil
	}

	if info.Mode().Perm()&0111 == 0 {
		return false, path, nil
	}

	return true, path, nil
}

func gitCommit(ctx context.Context, message string) error {
	hasHook, hookPath, hookErr := hasGitHook(ctx, "pre-commit")
	if hookErr != nil {
		pterm.Warning.Printf("Unable to determine pre-commit hooks: %v\n", hookErr)
	} else if hasHook {
		pterm.Warning.Printf("Detected pre-commit hook at %s. Hook output will be shown if it fails.\n", hookPath)
	}

	result, err := runGitRaw(ctx, "commit", "-m", message)
	if err != nil {
		logGitFailure(err)
		if hasHook && hookErr == nil {
			pterm.Warning.Println("Pre-commit hook blocked the commit. Fix the issues it reported and retry.")
		}
		return err
	}

	if trimmed := strings.TrimSpace(result.Stdout); trimmed != "" {
		pterm.Info.Println(trimmed)
	}

	return nil
}

func runGit(ctx context.Context, args ...string) (string, error) {
	result, err := runGitRaw(ctx, args...)
	if err != nil {
		return "", err
	}
	return result.Stdout, nil
}

type gitExecResult struct {
	Stdout string
	Stderr string
}

type gitCmdError struct {
	args   []string
	result gitExecResult
}

func (e *gitCmdError) Error() string {
	joined := strings.Join(e.args, " ")
	message := strings.TrimSpace(e.result.Stderr)
	if message == "" {
		message = "unknown error"
	}
	return fmt.Sprintf("git %s failed: %s", joined, message)
}

func (e *gitCmdError) Result() gitExecResult {
	return e.result
}

func runGitRaw(ctx context.Context, args ...string) (gitExecResult, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := gitExecResult{Stdout: stdout.String(), Stderr: stderr.String()}
	if err != nil {
		return result, &gitCmdError{args: args, result: result}
	}

	return result, nil
}

func logGitFailure(err error) {
	var gitErr *gitCmdError
	if errors.As(err, &gitErr) {
		stdout := strings.TrimSpace(gitErr.result.Stdout)
		stderr := strings.TrimSpace(gitErr.result.Stderr)
		var combined []string
		if stdout != "" {
			combined = append(combined, stdout)
		}
		if stderr != "" {
			combined = append(combined, stderr)
		}
		if len(combined) > 0 {
			pterm.Error.Println(strings.Join(combined, "\n"))
			return
		}
	}

	if err != nil {
		pterm.Error.Println(err.Error())
	}
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
