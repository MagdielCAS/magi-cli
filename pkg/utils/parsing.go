package utils

import "regexp"

// ⚡ BOLT OPTIMIZATION: Move regex to package level to avoid recompiling on every function call.
var codeBlockRegex = regexp.MustCompile(`(\` + "`" + "`" + "`" + `[\w-]*)\n([\s\S]*)(\` + "`" + "`" + "`" + `)`)

// RemoveCodeBlock removes code block tags from a string.
// If no code block tags are found, it returns the original string.
func RemoveCodeBlock(input string) string {
	matches := codeBlockRegex.FindStringSubmatch(input)
	if len(matches) == 0 {
		return input
	}

	return matches[2]
}
