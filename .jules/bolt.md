## 2024-05-19 - Regex Compilation Overhead
**Learning:** Recompiling regexes using `regexp.MustCompile` inside frequently called methods (like agent `Execute` methods or string parsers) adds severe overhead. In Go, compiling a regex is expensive.
**Action:** Always move `regexp.MustCompile` calls to package-level variables so they are compiled only once on initialization. This can provide up to a 40x speedup in parsing tasks.

## 2024-05-19 - GitHub Actions CI Push on PR
**Learning:** `GITHUB_REF` in GitHub Actions for a `pull_request` event evaluates to `refs/pull/PR_NUMBER/merge`. Pushing to this ref directly via `git push origin HEAD:${GITHUB_REF}` fails with `deny updating a hidden ref`.
**Action:** Always use conditional target refs. E.g. `HEAD:${{ github.head_ref }}` for PR events and `HEAD:${GITHUB_REF}` for other events (like push).
