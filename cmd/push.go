package cmd

import (
	"context"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push the current branch and auto-configure the upstream if needed",
	Long: `push wraps git push and automatically adds --set-upstream when the current branch
lacks an upstream. The command never sends source data anywhere—it shells out to your
local git binary and surfaces any hook output so you only run push once.

Data handling:
  • The command invokes git locally and does not upload project data on its own.

Usage:
  magi push

Security note: The command respects your git hooks and displays hook output when a push fails.`,
	RunE: runPush,
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	if err := ensureGitRepo(ctx); err != nil {
		return err
	}

	branch, err := currentBranchName(ctx)
	if err != nil {
		return err
	}

	hasUpstream := branchHasUpstream(ctx)
	hookDetected := warnOnHook(ctx, "pre-push")
	remote := branchRemote(ctx, branch)
	args := buildPushArgs(branch, remote, hasUpstream)

	if !hasUpstream {
		pterm.Info.Printf("No upstream configured for %s; running 'git %s'.\n", branch, strings.Join(args, " "))
	}

	if _, err := runGitRaw(ctx, args...); err != nil {
		logGitFailure(err)
		if hookDetected {
			pterm.Warning.Println("A pre-push hook is configured and may have blocked the push. Review the hook output above.")
		}
		return err
	}

	pterm.Success.Println("Branch pushed successfully.")
	return nil
}

func branchHasUpstream(ctx context.Context) bool {
	if _, err := runGit(ctx, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"); err != nil {
		return false
	}
	return true
}

func warnOnHook(ctx context.Context, hookName string) bool {
	hasHook, hookPath, err := hasGitHook(ctx, hookName)
	if err != nil {
		pterm.Warning.Printf("Unable to determine %s hooks: %v\n", hookName, err)
		return false
	}
	if hasHook {
		pterm.Warning.Printf("Detected %s hook at %s. Hook output will be shown if it fails.\n", hookName, hookPath)
		return true
	}
	return false
}

func buildPushArgs(branch, remote string, hasUpstream bool) []string {
	if hasUpstream {
		return []string{"push"}
	}
	return []string{"push", "--set-upstream", remote, branch}
}
