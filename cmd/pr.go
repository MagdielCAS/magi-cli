package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/MagdielCAS/magi-cli/pkg/pr"
	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Review local commits with AI agents and open a GitHub pull request",
	Long: `pr scans your local commits that have not been pushed to origin/<branch>, runs an AgenticGoKit
workflow to review the diff, fills the repository pull request template, and creates the PR via GitHub CLI.

Data handling:
  • Sends the git diff between HEAD and origin/<branch>, AGENTS.md contents, and optional user context
    to your configured AI provider.
  • No other files are uploaded.

Usage:
  magi pr

Security note: The review agents run with the hardened shared HTTP client, redact API keys, and never
persist model responses. The command shells out to 'git' and 'gh' with explicit arguments.`,
	RunE: runPR,
}

func init() {
	rootCmd.AddCommand(prCmd)
}

func runPR(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	if err := ensureGitRepo(ctx); err != nil {
		return err
	}

	runtimeCtx, err := shared.BuildRuntimeContext()
	if err != nil {
		return err
	}

	branch, err := currentBranchName(ctx)
	if err != nil {
		return err
	}

	repoRoot, err := repoRootPath(ctx)
	if err != nil {
		return err
	}

	diff, baseRef, baseBranch, err := diffAgainstBaseBranch(ctx, branch)
	if err != nil {
		return err
	}

	templatePath := filepath.Join(repoRoot, ".github", "pull_request_template.md")
	templateBody, err := pr.LoadPullRequestTemplate(templatePath)
	if err != nil {
		return err
	}

	guidelines, err := pr.CollectAgentGuidelines(repoRoot)
	if err != nil {
		return err
	}

	additionalContext, err := promptAdditionalContext()
	if err != nil {
		return err
	}

	reviewer := pr.NewAgenticReviewer(runtimeCtx)
	artifacts, err := reviewer.Review(ctx, pr.ReviewInput{
		Diff:              diff,
		Branch:            branch,
		RemoteRef:         baseRef,
		Guidelines:        guidelines,
		AdditionalContext: additionalContext,
		Template:          templateBody,
	})
	if err != nil {
		return err
	}

	logFindings(*artifacts)

	confirmed, err := pterm.DefaultInteractiveConfirm.WithDefaultValue(true).
		Show("Create the pull request with the filled template above?")
	if err != nil {
		return fmt.Errorf("confirmation prompt failed: %w", err)
	}
	if !confirmed {
		pterm.Warning.Println("Pull request creation cancelled by user.")
		return nil
	}

	pterm.Info.Println("Ensuring the branch is pushed before creating the pull request...")
	if err := runPush(cmd, nil); err != nil {
		return fmt.Errorf("failed to push branch prior to PR creation: %w", err)
	}

	prURL, err := createPullRequest(ctx, branch, baseBranch, artifacts.Plan)
	if err != nil {
		return err
	}

	comment := pr.FormatFindingsComment(artifacts.Analysis)
	if err := commentOnPullRequest(ctx, comment); err != nil {
		return err
	}

	pterm.Success.Printf("Pull request created: %s\n", prURL)
	return nil
}

func promptAdditionalContext() (string, error) {
	wantContext, err := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).
		Show("Add optional context for the AI reviewers?")
	if err != nil {
		return "", fmt.Errorf("context confirmation failed: %w", err)
	}
	if !wantContext {
		return "", nil
	}

	content, err := pterm.DefaultInteractiveTextInput.
		WithMultiLine().
		WithDefaultText("Enter any risk, testing, or deployment notes").
		Show()
	if err != nil {
		return "", fmt.Errorf("context input failed: %w", err)
	}
	return strings.TrimSpace(content), nil
}

func logFindings(artifacts pr.ReviewArtifacts) {
	pterm.DefaultSection.Println("Agent Findings")
	printList("Summary", []string{artifacts.Analysis.Summary})
	printList("Code Smells", artifacts.Analysis.CodeSmells)
	printList("Security Concerns", artifacts.Analysis.SecurityConcerns)
	printList("AGENTS Alerts", artifacts.Analysis.AgentsGuidelineAlerts)
	printList("Test Recommendations", artifacts.Analysis.TestRecommendations)
	printList("Documentation Updates", artifacts.Analysis.DocumentationUpdates)
	printList("Risk Callouts", artifacts.Analysis.RiskCallouts)

	pterm.DefaultSection.Println("Filled Pull Request Template")
	fmt.Println(strings.TrimSpace(artifacts.Plan.Body))
}

func printList(title string, entries []string) {
	pterm.Println(pterm.ThemeDefault.SectionStyle.Sprint(title))
	if len(entries) == 0 {
		pterm.Println("  • None")
		return
	}
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		pterm.Println("  • " + entry)
	}
}

func diffAgainstBaseBranch(ctx context.Context, branch string) (string, string, string, error) {
	baseRef, baseBranch, err := resolveBaseBranch(ctx, branch)
	if err != nil {
		return "", "", "", err
	}

	diff, err := runGit(ctx, "diff", fmt.Sprintf("%s..HEAD", baseRef))
	if err != nil {
		return "", "", "", err
	}

	if strings.TrimSpace(diff) == "" {
		return "", "", "", fmt.Errorf("no differences detected between HEAD and %s", baseRef)
	}

	return diff, baseRef, baseBranch, nil
}

func resolveBaseBranch(ctx context.Context, branch string) (string, string, error) {
	remote := branchRemote(ctx, branch)

	mergeRef, err := runGit(ctx, "config", fmt.Sprintf("branch.%s.merge", branch))
	baseBranch := ""
	if err == nil {
		baseBranch = strings.TrimPrefix(strings.TrimSpace(mergeRef), "refs/heads/")
	}

	if baseBranch == "" {
		headRef, headErr := runGit(ctx, "symbolic-ref", fmt.Sprintf("refs/remotes/%s/HEAD", remote))
		if headErr == nil {
			prefix := fmt.Sprintf("refs/remotes/%s/", remote)
			baseBranch = strings.TrimPrefix(strings.TrimSpace(headRef), prefix)
		}
	}

	if baseBranch == "" {
		baseBranch = "main"
	}

	baseRef := fmt.Sprintf("%s/%s", remote, baseBranch)
	if _, err := runGit(ctx, "rev-parse", "--verify", baseRef); err != nil {
		return "", "", fmt.Errorf("unable to resolve %s: %w", baseRef, err)
	}

	return baseRef, baseBranch, nil
}

func repoRootPath(ctx context.Context) (string, error) {
	output, err := runGit(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to determine repository root: %w", err)
	}
	return strings.TrimSpace(output), nil
}

func createPullRequest(ctx context.Context, branch, base string, plan pr.PullRequestPlan) (string, error) {
	bodyFile, err := writeTempFile("magi-pr-body-*.md", plan.Body)
	if err != nil {
		return "", err
	}
	defer os.Remove(bodyFile)

	args := []string{
		"pr", "create",
		"--title", strings.TrimSpace(plan.Title),
		"--body-file", bodyFile,
		"--head", branch,
	}
	if base != "" {
		args = append(args, "--base", base)
	}
	if _, err := runGH(ctx, args...); err != nil {
		return "", err
	}

	info, err := runGH(ctx, "pr", "view", "--json", "number,url")
	if err != nil {
		return "", err
	}

	type prInfo struct {
		Number int    `json:"number"`
		URL    string `json:"url"`
	}
	var parsed prInfo
	if err := json.Unmarshal([]byte(info), &parsed); err != nil {
		return "", fmt.Errorf("failed to parse gh pr view response: %w", err)
	}
	if strings.TrimSpace(parsed.URL) == "" {
		return "", fmt.Errorf("gh did not return a pull request URL")
	}

	return parsed.URL, nil
}

func commentOnPullRequest(ctx context.Context, body string) error {
	if strings.TrimSpace(body) == "" {
		return nil
	}

	commentFile, err := writeTempFile("magi-pr-comment-*.md", body)
	if err != nil {
		return err
	}
	defer os.Remove(commentFile)

	if _, err := runGH(ctx, "pr", "comment", "--body-file", commentFile); err != nil {
		return err
	}

	pterm.Success.Println("Posted agent findings as a PR comment.")
	return nil
}

func writeTempFile(pattern, content string) (string, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return file.Name(), nil
}

func runGH(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	var stdout strings.Builder
	var stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh %s failed: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}

	return stdout.String(), nil
}
