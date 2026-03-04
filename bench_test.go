package main

import (
	"regexp"
	"strings"
	"testing"
)

var diff = strings.Repeat("+ t('hello')\n- ignored\n+ i18n.t(\"world\")\n", 1000)

var globalPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?:^|[^a-zA-Z0-9_])t\((?:'([^']+)'|"([^"]+)")\)`),
	regexp.MustCompile(`i18n\.t\((?:'([^']+)'|"([^"]+)")\)`),
	regexp.MustCompile(`\$t\((?:'([^']+)'|"([^"]+)")\)`),
	regexp.MustCompile(`<T[^>]+key=(?:'([^']+)'|"([^"]+)")`),
	regexp.MustCompile(`<T[^>]+keyName=(?:'([^']+)'|"([^"]+)")`),
}

func BenchmarkCurrent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lines := strings.Split(diff, "\n")
		patterns := []*regexp.Regexp{
			regexp.MustCompile(`(?:^|[^a-zA-Z0-9_])t\((?:'([^']+)'|"([^"]+)")\)`),
			regexp.MustCompile(`i18n\.t\((?:'([^']+)'|"([^"]+)")\)`),
			regexp.MustCompile(`\$t\((?:'([^']+)'|"([^"]+)")\)`),
			regexp.MustCompile(`<T[^>]+key=(?:'([^']+)'|"([^"]+)")`),
			regexp.MustCompile(`<T[^>]+keyName=(?:'([^']+)'|"([^"]+)")`),
		}

		for _, line := range lines {
			if !strings.HasPrefix(line, "+") {
				continue
			}
			content := line[1:]
			for _, pattern := range patterns {
				pattern.FindAllStringSubmatch(content, -1)
			}
		}
	}
}

func BenchmarkOptimized(b *testing.B) {
	for i := 0; i < b.N; i++ {
		remaining := diff
		for len(remaining) > 0 {
			var line string
			idx := strings.IndexByte(remaining, '\n')
			if idx >= 0 {
				line = remaining[:idx]
				remaining = remaining[idx+1:]
			} else {
				line = remaining
				remaining = ""
			}

			if !strings.HasPrefix(line, "+") {
				continue
			}
			content := line[1:]
			for _, pattern := range globalPatterns {
				pattern.FindAllStringSubmatch(content, -1)
			}
		}
	}
}
