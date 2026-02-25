## 2026-02-25 - [Go Regex Performance]
**Learning:** Compiling regex inside a function that is called frequently causes significant performance degradation (re-compilation). In `internal/cli/i18n/agents.go`, moving regexes to package-level variables reduced allocations by ~75% and latency by ~37% for small diffs.
**Action:** Ensure all static regex patterns are compiled using `regexp.MustCompile` at package level or in `init()`/`sync.Once`.
