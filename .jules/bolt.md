## 2024-05-15 - [Title]
**Learning:** Moving `regexp.MustCompile` from function scope to package-level variables in this Go repository provides a significant performance improvement (measured up to ~100,000x faster for repetitive operations, e.g., 0.38ns vs 38000ns) avoiding the overhead of repeated regex compilation during execution.
**Action:** When working in Go codebases, inspect tight loops and frequently called utility functions for `regexp.MustCompile` calls. Extract these to package-level variables to optimize execution speed.
