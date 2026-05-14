## 2024-10-24 - Avoid strings.Split for simple line parsing when performance matters
**Learning:** Found several places where `strings.Split` is used on strings (e.g., `strings.Split(a.diff, "\n")` in `KeyExtractor`). When allocating many strings, `strings.Index` and manual slicing is faster and creates fewer allocations. The memory context mentions this specifically: "Performance convention: When parsing command output for specific markers (e.g., 'HEAD branch:'), prefer using 'strings.Index' and manual slicing over 'strings.Split' or 'strings.Scanner' to minimize allocations and processing time."
**Action:** Replace `strings.Split` with manual slicing via `strings.Index` when processing potentially large strings like diffs.

## 2026-02-25 - [Go Regex Performance]
**Learning:** Compiling regex inside a function that is called frequently causes significant performance degradation (re-compilation). In `internal/cli/i18n/agents.go`, moving regexes to package-level variables reduced allocations by ~75% and latency by ~37% for small diffs.
**Action:** Ensure all static regex patterns are compiled using `regexp.MustCompile` at package level or in `init()`/`sync.Once`.
