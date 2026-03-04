package main

import (
	"regexp"
	"strings"
	"testing"
)

var diff2 = strings.Repeat("+ t('hello')\n- ignored\n+ i18n.t(\"world\")\n", 100)

func BenchmarkRegexCurrent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		patterns := []*regexp.Regexp{
			regexp.MustCompile(`(?:^|[^a-zA-Z0-9_])t\((?:'([^']+)'|"([^"]+)")\)`),
			regexp.MustCompile(`i18n\.t\((?:'([^']+)'|"([^"]+)")\)`),
			regexp.MustCompile(`\$t\((?:'([^']+)'|"([^"]+)")\)`),
			regexp.MustCompile(`<T[^>]+key=(?:'([^']+)'|"([^"]+)")`),
			regexp.MustCompile(`<T[^>]+keyName=(?:'([^']+)'|"([^"]+)")`),
		}
		for _, p := range patterns {
			p.MatchString(diff2)
		}
	}
}

var globalPatterns2 = []*regexp.Regexp{
	regexp.MustCompile(`(?:^|[^a-zA-Z0-9_])t\((?:'([^']+)'|"([^"]+)")\)`),
	regexp.MustCompile(`i18n\.t\((?:'([^']+)'|"([^"]+)")\)`),
	regexp.MustCompile(`\$t\((?:'([^']+)'|"([^"]+)")\)`),
	regexp.MustCompile(`<T[^>]+key=(?:'([^']+)'|"([^"]+)")`),
	regexp.MustCompile(`<T[^>]+keyName=(?:'([^']+)'|"([^"]+)")`),
}

func BenchmarkRegexOptimized(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, p := range globalPatterns2 {
			p.MatchString(diff2)
		}
	}
}
