package cmd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBranchRemoteUsesConfiguredRemote(t *testing.T) {
	repo := initTestRepo(t)
	remote := initBareRemote(t)
	runGitCmd(t, repo, "remote", "add", "origin", remote)
	runGitCmd(t, repo, "push", "-u", "origin", "main")

	withGitEnv(t, repo)
	got, err := branchRemote(context.Background(), "main")
	if err != nil {
		t.Fatalf("branchRemote returned error: %v", err)
	}
	if got != "origin" {
		t.Fatalf("expected remote origin, got %s", got)
	}
}

func TestBranchRemoteFallsBackToUpstream(t *testing.T) {
	repo := initTestRepo(t)
	remote := initBareRemote(t)
	runGitCmd(t, repo, "remote", "add", "origin", remote)
	runGitCmd(t, repo, "push", "-u", "origin", "main")
	runGitCmd(t, repo, "config", "--unset", "branch.main.remote")

	withGitEnv(t, repo)
	got, err := branchRemote(context.Background(), "main")
	if err != nil {
		t.Fatalf("branchRemote returned error: %v", err)
	}
	if got != "origin" {
		t.Fatalf("expected remote origin via upstream, got %s", got)
	}
}

func TestBranchRemoteErrorsForMissingRemote(t *testing.T) {
	repo := initTestRepo(t)
	runGitCmd(t, repo, "config", "branch.main.remote", "missing")

	withGitEnv(t, repo)
	_, err := branchRemote(context.Background(), "main")
	if err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("expected error referencing missing remote, got %v", err)
	}
}

func TestGitHooksDirReturnsCleanPath(t *testing.T) {
	repo := initTestRepo(t)
	withGitEnv(t, repo)

	path, err := gitHooksDir(context.Background())
	if err != nil {
		t.Fatalf("gitHooksDir returned error: %v", err)
	}

	expected := filepath.Join(repo, ".git", "hooks")
	if !pathsEqual(path, expected) {
		t.Fatalf("expected hooks dir %s, got %s", expected, path)
	}
}

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	runGitCmd(t, dir, "init", "--initial-branch=main")
	runGitCmd(t, dir, "config", "user.name", "magi-tests")
	runGitCmd(t, dir, "config", "user.email", "magi@example.com")

	contentPath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(contentPath, []byte("# test repo\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	runGitCmd(t, dir, "add", "README.md")
	runGitCmd(t, dir, "commit", "-m", "init")

	return dir
}

func initBareRemote(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "remote.git")
	cmd := exec.Command("git", "init", "--bare", dir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init bare failed: %v\n%s", err, string(output))
	}
	return dir
}

func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
}

func withGitEnv(t *testing.T, repo string) {
	t.Helper()
	t.Setenv("GIT_DIR", filepath.Join(repo, ".git"))
	t.Setenv("GIT_WORK_TREE", repo)
}

func pathsEqual(a, b string) bool {
	aClean, err := filepath.EvalSymlinks(a)
	if err != nil {
		aClean = filepath.Clean(a)
	}
	bClean, err := filepath.EvalSymlinks(b)
	if err != nil {
		bClean = filepath.Clean(b)
	}
	return aClean == bClean
}
