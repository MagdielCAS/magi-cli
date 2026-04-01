## 2024-05-24 - [Avoid Regex Compilation in Functions]
**Learning:** Re-compiling regular expressions using `regexp.MustCompile` inside frequently called functions degrades performance unnecessarily, as the compilation cost is paid on every function call.
**Action:** Always move `regexp.MustCompile` out to package-level variables so that the regular expressions are compiled only once when the package is initialized.
