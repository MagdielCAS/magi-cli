package utils

import "regexp"

// RemoveCodeBlock removes code block tags from a string.
// If no code block tags are found, it returns the original string.
func RemoveCodeBlock(input string) string {
	re := regexp.MustCompile(`(\` + "`" + "`" + "`" + `[\w-]*)\n([\s\S]*)(\` + "`" + "`" + "`" + `)`)

	matches := re.FindStringSubmatch(input)
	if len(matches) == 0 {
		return input
	}

	return matches[2]
}
