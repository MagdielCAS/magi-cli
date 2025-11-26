package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
)

func RunGit(ctx context.Context, args ...string) (string, error) {
	result, err := RunGitRaw(ctx, args...)
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

func RunGitRaw(ctx context.Context, args ...string) (gitExecResult, error) {
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

func LogGitFailure(err error) {
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

func EnsureGitRepo(ctx context.Context) error {
	_, err := RunGit(ctx, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return fmt.Errorf("this command must be run inside a git repository: %w", err)
	}
	return nil
}

func CurrentBranchName(ctx context.Context) (string, error) {
	output, err := RunGit(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to determine current branch: %w", err)
	}

	branch := strings.TrimSpace(output)
	if branch == "" {
		return "", errors.New("current branch name is empty")
	}

	return branch, nil
}

func BranchRemote(ctx context.Context, branch string) (string, error) {
	if branch == "" {
		return "", errors.New("branch name is required to determine its remote")
	}

	remote := ""
	if output, err := RunGit(ctx, "config", fmt.Sprintf("branch.%s.remote", branch)); err == nil {
		remote = strings.TrimSpace(output)
	}

	if remote == "" {
		if upstream, err := RunGit(ctx, "rev-parse", "--abbrev-ref", fmt.Sprintf("%s@{u}", branch)); err == nil {
			if parts := strings.SplitN(strings.TrimSpace(upstream), "/", 2); len(parts) == 2 {
				remote = parts[0]
			}
		}
	}

	if remote == "" {
		remote = "origin"
	}

	if err := verifyRemoteExists(ctx, remote); err != nil {
		return "", fmt.Errorf("unable to find remote %s for branch %s: %w", remote, branch, err)
	}

	return remote, nil
}

func verifyRemoteExists(ctx context.Context, remote string) error {
	if strings.TrimSpace(remote) == "" {
		return errors.New("remote name cannot be empty")
	}
	if _, err := RunGit(ctx, "remote", "get-url", remote); err != nil {
		return err
	}
	return nil
}

func HasGitHook(ctx context.Context, hookName string) (bool, string, error) {
	hooksDir, err := GitHooksDir(ctx)
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

func GitHooksDir(ctx context.Context) (string, error) {
	output, err := RunGit(ctx, "rev-parse", "--path-format=absolute", "--git-path", "hooks")
	if err != nil {
		return "", err
	}

	path := strings.TrimSpace(output)
	if path == "" {
		return "", fmt.Errorf("git did not return a hooks directory")
	}

	return filepath.Clean(path), nil
}
