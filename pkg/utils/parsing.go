package utils

import "regexp"

// codeBlockRegex is used to extract content from markdown code blocks.
// It is non-greedy to extract only the first block.
var codeBlockRegex = regexp.MustCompile(`(\` + "`" + "`" + "`" + `[\w-]*)\n([\s\S]*?)(\` + "`" + "`" + "`" + `)`)

// RemoveCodeBlock removes code block tags from a string.
// If no code block tags are found, it returns the original string.
func RemoveCodeBlock(input string) string {
	matches := codeBlockRegex.FindStringSubmatch(input)
	if len(matches) == 0 {
		return input
	}

	return matches[2]
}
