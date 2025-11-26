package i18n

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestKeyExtractor_Execute(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		expected []string
	}{
		{
			name: "Basic t function",
			diff: `
+ console.log(t('hello.world'));
+ const x = t("another.key");
`,
			expected: []string{"hello.world", "another.key"},
		},
		{
			name: "React components",
			diff: `
+ <T key="component.key" />
+ <T keyName="trans.key" />
`,
			expected: []string{"component.key", "trans.key"},
		},
		{
			name: "Vue/Other patterns",
			diff: `
+ $t('vue.key')
+ i18n.t('lib.key')
`,
			expected: []string{"vue.key", "lib.key"},
		},
		{
			name: "Ignore unchanged lines",
			diff: `
  t('ignored.key')
+ t('included.key')
- t('deleted.key')
`,
			expected: []string{"included.key"},
		},
		{
			name: "Complex context",
			diff: `
+ if (true) { return t('nested.key'); }
+ <div title={t('attr.key')}></div>
`,
			expected: []string{"nested.key", "attr.key"},
		},
		{
			name: "Duplicate keys",
			diff: `
+ t('dup.key')
+ t('dup.key')
`,
			expected: []string{"dup.key"},
		},
		{
			name: "Quotes in keys",
			diff: `
+ t("It's a key")
+ t('Say "Hello"')
`,
			expected: []string{"It's a key", `Say "Hello"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewKeyExtractor(tt.diff)
			result, err := extractor.Execute(nil)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			var keys []I18nKey
			if err := json.Unmarshal([]byte(result), &keys); err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}

			// Create a map for easier checking
			keyMap := make(map[string]bool)
			for _, k := range keys {
				keyMap[k.Key] = true
			}

			for _, expectedKey := range tt.expected {
				if !keyMap[expectedKey] {
					t.Errorf("Expected key %q not found", expectedKey)
				}
			}

			if len(keys) != len(tt.expected) {
				t.Errorf("Expected %d keys, got %d", len(tt.expected), len(keys))
			}
		})
	}
}

func TestSQLGenerator_Execute(t *testing.T) {
	// Mock input data from TranslationEnhancer
	inputData := TranslationData{
		Keys: []I18nKey{
			{
				Key: "test.key",
				Translations: map[string]string{
					"en": "Hello World",
					"de": "Hallo Welt",
				},
			},
			{
				Key: "escape.key",
				Translations: map[string]string{
					"en": "It's a test",
					"fr": "C'est un test",
				},
			},
		},
	}

	jsonData, err := json.Marshal(inputData)
	if err != nil {
		t.Fatalf("Failed to marshal input data: %v", err)
	}

	generator := NewSQLGenerator()
	input := map[string]string{
		"translation_enhancer": string(jsonData),
	}

	result, err := generator.Execute(input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify SQL content
	if !strings.Contains(result, "BEGIN;") {
		t.Error("SQL should start transaction")
	}
	if !strings.Contains(result, "COMMIT;") {
		t.Error("SQL should commit transaction")
	}

	// Check first key
	if !strings.Contains(result, "VALUES ('test.key', 'en', 'Hello World')") {
		t.Error("Missing English translation for test.key")
	}
	if !strings.Contains(result, "VALUES ('test.key', 'de', 'Hallo Welt')") {
		t.Error("Missing German translation for test.key")
	}

	// Check escaping
	if !strings.Contains(result, "VALUES ('escape.key', 'en', 'It''s a test')") {
		t.Error("Single quote not escaped correctly in English")
	}
	if !strings.Contains(result, "VALUES ('escape.key', 'fr', 'C''est un test')") {
		t.Error("Single quote not escaped correctly in French")
	}
}
