package utils

import "regexp"

// codeBlockRegex caches the compiled regex for code block extraction.
// Performance optimization: Compiling regexes at package initialization
// avoids expensive recompilation on every RemoveCodeBlock() call.
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
