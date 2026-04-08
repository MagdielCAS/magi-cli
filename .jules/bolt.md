## 2024-06-25 - Prevent regex recompilation
**Learning:** Calling `regexp.MustCompile` inside a function body forces the regular expression to be compiled on every single invocation. This is a massive performance bottleneck for functions called frequently.
**Action:** Always move `regexp.MustCompile` calls out of function scopes and into package-level global variables, so the regex is compiled only once at initialization. This provides a roughly 10x performance improvement (~8000 ns/op to ~880 ns/op).
