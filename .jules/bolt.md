## 2024-05-19 - Regex Compilation Overhead
**Learning:** Recompiling regexes using `regexp.MustCompile` inside frequently called methods (like agent `Execute` methods or string parsers) adds severe overhead. In Go, compiling a regex is expensive.
**Action:** Always move `regexp.MustCompile` calls to package-level variables so they are compiled only once on initialization. This can provide up to a 40x speedup in parsing tasks.
