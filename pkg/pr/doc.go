// Package pr hosts reusable helpers for the pull request workflows that power
// the `magi pr` command. It includes prompt builders, AGENTS guideline loaders,
// and the AgenticReviewer used in cmd/pr.go so commands can share hardened logic
// without reimplementing multi-agent orchestration or sanitization steps.
package pr
