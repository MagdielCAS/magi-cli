package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/MagdielCAS/magi-cli/pkg/agent"
	"github.com/MagdielCAS/magi-cli/pkg/llm"
)

// Data Models

type TranslationData struct {
	Keys []I18nKey `json:"keys"`
}

type I18nKey struct {
	Key          string            `json:"key"`
	Context      string            `json:"context"`
	Translations map[string]string `json:"translations"`
}

// Agent Implementations

// KeyExtractor Agent
type KeyExtractor struct {
	diff string
}

func NewKeyExtractor(diff string) *KeyExtractor {
	return &KeyExtractor{diff: diff}
}

func (a *KeyExtractor) Name() string {
	return "key_extractor"
}

func (a *KeyExtractor) WaitForResults() []string {
	return []string{} // No dependencies, runs first
}

func (a *KeyExtractor) Execute(input map[string]string) (string, error) {
	var keys []I18nKey
	lines := strings.Split(a.diff, "\n")

	// Regex patterns for different i18n usage
	// We use two capturing groups: one for single quotes, one for double quotes
	patterns := []*regexp.Regexp{
		// t('key') or t("key")
		regexp.MustCompile(`(?:^|[^a-zA-Z0-9_])t\((?:'([^']+)'|"([^"]+)")\)`),
		// i18n.t('key') or i18n.t("key")
		regexp.MustCompile(`i18n\.t\((?:'([^']+)'|"([^"]+)")\)`),
		// $t('key') or $t("key")
		regexp.MustCompile(`\$t\((?:'([^']+)'|"([^"]+)")\)`),
		// <T key="key" />
		regexp.MustCompile(`<T[^>]+key=(?:'([^']+)'|"([^"]+)")`),
		// <T keyName="key" />
		regexp.MustCompile(`<T[^>]+keyName=(?:'([^']+)'|"([^"]+)")`),
	}

	for _, line := range lines {
		// We only care about added lines
		if !strings.HasPrefix(line, "+") {
			continue
		}

		// Remove the "+" prefix
		content := line[1:]

		for _, pattern := range patterns {
			matches := pattern.FindAllStringSubmatch(content, -1)
			for _, match := range matches {
				// match[0] is full match
				// match[1] is single quote group
				// match[2] is double quote group
				var key string
				if len(match) > 1 && match[1] != "" {
					key = match[1]
				} else if len(match) > 2 && match[2] != "" {
					key = match[2]
				}

				if key != "" {
					// Basic context extraction (just the line content for now)
					context := strings.TrimSpace(content)
					if len(context) > 100 {
						context = context[:100] + "..."
					}

					keys = append(keys, I18nKey{
						Key:     key,
						Context: context,
					})
				}
			}
		}
	}

	// Remove duplicates
	uniqueKeys := make(map[string]I18nKey)
	for _, k := range keys {
		if _, exists := uniqueKeys[k.Key]; !exists {
			uniqueKeys[k.Key] = k
		}
	}

	finalKeys := make([]I18nKey, 0, len(uniqueKeys))
	for _, k := range uniqueKeys {
		finalKeys = append(finalKeys, k)
	}

	jsonData, err := json.Marshal(finalKeys)
	if err != nil {
		return "", fmt.Errorf("failed to marshal keys: %w", err)
	}

	return string(jsonData), nil
}

// TranslationGenerator Agent
type TranslationGenerator struct {
	llmService *llm.Service
}

func NewTranslationGenerator(service *llm.Service) *TranslationGenerator {
	return &TranslationGenerator{
		llmService: service,
	}
}

func (a *TranslationGenerator) Name() string {
	return "translation_generator"
}

func (a *TranslationGenerator) WaitForResults() []string {
	return []string{"key_extractor"}
}

func (a *TranslationGenerator) Execute(input map[string]string) (string, error) {
	keysJSON := input["key_extractor"]
	var keys []I18nKey
	if err := json.Unmarshal([]byte(keysJSON), &keys); err != nil {
		return "", fmt.Errorf("failed to parse extracted keys: %w", err)
	}

	if len(keys) == 0 {
		return `{"keys": []}`, nil
	}

	langs := strings.Join(languages, ", ")
	batchSize := 15 // Process 15 keys at a time to avoid timeouts
	var allTranslatedKeys []I18nKey

	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}
		batch := keys[i:end]
		batchJSON, _ := json.Marshal(batch)

		prompt := fmt.Sprintf(`You are a professional translator.
Translate the following i18n keys to %s.
The input is a JSON array of keys.
Return a JSON object with the following structure:
{
  "keys": [
    {
      "key": "original_key",
      "context": "context if available",
      "translations": {
        "en": "English translation",
        "de": "German translation",
        ... (one key for each target language code)
      }
    }
  ]
}

Input Keys:
%s
`, langs, string(batchJSON))

		req := llm.ChatCompletionRequest{
			Messages: []llm.ChatMessage{
				{Role: "system", Content: "You are a helpful assistant that generates i18n translations."},
				{Role: "user", Content: prompt},
			},
			Temperature: 0.3,
		}

		// Retry logic
		var response string
		var err error
		maxRetries := 3
		for attempt := 0; attempt < maxRetries; attempt++ {
			response, err = a.llmService.ChatCompletion(context.Background(), req)
			if err == nil {
				break
			}
			// Exponential backoff: 2s, 4s, 8s
			sleepTime := time.Duration(1<<attempt) * 2 * time.Second
			time.Sleep(sleepTime)
		}

		if err != nil {
			return "", fmt.Errorf("failed to translate batch %d-%d after %d retries: %w", i, end, maxRetries, err)
		}

		// Clean response
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)

		var batchResult TranslationData
		if err := json.Unmarshal([]byte(response), &batchResult); err != nil {
			// Try parsing as raw array if the model forgot the wrapper
			var rawKeys []I18nKey
			if err2 := json.Unmarshal([]byte(response), &rawKeys); err2 == nil {
				allTranslatedKeys = append(allTranslatedKeys, rawKeys...)
				continue
			}
			return "", fmt.Errorf("failed to parse batch response: %w", err)
		}
		allTranslatedKeys = append(allTranslatedKeys, batchResult.Keys...)
	}

	finalResult := TranslationData{Keys: allTranslatedKeys}
	finalJSON, err := json.Marshal(finalResult)
	if err != nil {
		return "", fmt.Errorf("failed to marshal final result: %w", err)
	}

	return string(finalJSON), nil
}

// TranslationEnhancer Agent
type TranslationEnhancer struct {
	llmService *llm.Service
}

func NewTranslationEnhancer(service *llm.Service) *TranslationEnhancer {
	return &TranslationEnhancer{
		llmService: service,
	}
}

func (a *TranslationEnhancer) Name() string {
	return "translation_enhancer"
}

func (a *TranslationEnhancer) WaitForResults() []string {
	return []string{"translation_generator"}
}

func (a *TranslationEnhancer) Execute(input map[string]string) (string, error) {
	translationsJSON := input["translation_generator"]

	// Construct prompt for enhancement
	prompt := fmt.Sprintf(`You are a professional localization expert.
Review the following translations and enhance them for clarity, consistency, and professional tone.
If a context is provided, use it to ensure the translation fits the usage.
Do not change the keys.
Return the result in the exact same JSON structure.

Input Translations:
%s
`, translationsJSON)

	// Call LLM
	req := llm.ChatCompletionRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: "You are a helpful assistant that enhances i18n translations."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.2, // Lower temperature for consistency
	}

	response, err := a.llmService.ChatCompletion(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("LLM enhancement failed: %w", err)
	}

	// Clean response
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	return response, nil
}

// SQLGenerator Agent
type SQLGenerator struct{}

func NewSQLGenerator() *SQLGenerator {
	return &SQLGenerator{}
}

func (a *SQLGenerator) Name() string {
	return "sql_generator"
}

func (a *SQLGenerator) WaitForResults() []string {
	return []string{"translation_enhancer"}
}

func (a *SQLGenerator) Execute(input map[string]string) (string, error) {
	translationsJSON := input["translation_enhancer"]

	var translationData TranslationData
	// Try to unmarshal as TranslationData first
	if err := json.Unmarshal([]byte(translationsJSON), &translationData); err != nil {
		// Try as array of keys
		var keys []I18nKey
		if err2 := json.Unmarshal([]byte(translationsJSON), &keys); err2 != nil {
			return "", fmt.Errorf("failed to parse translations for SQL generation: %w", err)
		}
		translationData = TranslationData{Keys: keys}
	}

	var sb strings.Builder
	sb.WriteString("-- Auto-generated i18n SQL script\n")
	sb.WriteString("BEGIN;\n\n")

	for _, k := range translationData.Keys {
		sb.WriteString(fmt.Sprintf("-- Key: %s\n", k.Key))
		for lang, content := range k.Translations {
			escapedContent := strings.ReplaceAll(content, "'", "''")
			sb.WriteString(fmt.Sprintf("INSERT INTO i18n_translations (key, locale, content) VALUES ('%s', '%s', '%s') ON CONFLICT (key, locale) DO UPDATE SET content = EXCLUDED.content;\n", k.Key, lang, escapedContent))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("COMMIT;\n")
	return sb.String(), nil
}

// Helper to ensure agents implement the interface
var _ agent.AgentInstance = &KeyExtractor{}
var _ agent.AgentInstance = &TranslationGenerator{}
var _ agent.AgentInstance = &TranslationEnhancer{}
var _ agent.AgentInstance = &SQLGenerator{}
