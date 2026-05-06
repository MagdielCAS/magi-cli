## 2024-05-06 - [Hoist regexp.MustCompile to package level variables]
**Learning:** In Go, the `regexp` package parses regular expressions and builds execution machines at compile time. Repeating `regexp.MustCompile` inside frequently executed functions introduces significant CPU and memory overhead. Moving them to global variables ensures they are compiled exactly once at application startup.
**Action:** When finding `regexp.MustCompile` inside functions, hoist them out to package-level variables, especially within loops, hot paths, or frequently called agent methods like `Execute`.
