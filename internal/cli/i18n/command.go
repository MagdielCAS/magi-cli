package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/MagdielCAS/magi-cli/pkg/agent"
	"github.com/MagdielCAS/magi-cli/pkg/git"
	"github.com/MagdielCAS/magi-cli/pkg/llm"
	"github.com/MagdielCAS/magi-cli/pkg/shared"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	originBranch string
	maxTokens    int
	textFormat   bool
	autoConfirm  bool
	tolgeeOutput bool
	languages    []string
	outputFile   string
)

var i18nCmd = &cobra.Command{
	Use:   "i18n",
	Short: "AI-powered i18n translation management",
	Long: `Automates the extraction and translation of i18n keys from code changes.
It compares the current branch with an origin branch to find new keys,
then uses AI agents to generate translations in specified languages.`,
	RunE: runI18n,
}

func I18nCmd() *cobra.Command {
	i18nCmd.Flags().StringVar(&originBranch, "origin", "main", "Origin branch to compare against")
	i18nCmd.Flags().IntVar(&maxTokens, "max-tokens", 1000, "Max tokens for AI response")
	i18nCmd.Flags().BoolVar(&textFormat, "text-format", false, "Use text format instead of JSON schema")
	i18nCmd.Flags().BoolVar(&autoConfirm, "yes", false, "Auto-confirm all prompts")
	i18nCmd.Flags().BoolVar(&tolgeeOutput, "tolgee", false, "Generate Tolgee-compatible output files")
	i18nCmd.Flags().StringSliceVar(&languages, "languages", []string{"en", "de"}, "Target languages for translation")
	i18nCmd.Flags().StringVarP(&outputFile, "output", "o", "i18n_translations.json", "Output file for translations")

	return i18nCmd
}

func runI18n(cmd *cobra.Command, args []string) error {
	pterm.DefaultSection.Println("Running AI-Powered I18n Extraction")

	// 1. Git Integration
	ctx := cmd.Context()
	if err := git.EnsureGitRepo(ctx); err != nil {
		return err
	}

	currentBranch, err := git.CurrentBranchName(ctx)
	if err != nil {
		return err
	}

	pterm.Info.Printf("Comparing branch '%s' with origin '%s'...\n", currentBranch, originBranch)

	// Using 3 dots ... finds the merge base.
	diffOutput, err := git.RunGit(ctx, "diff", "-U0", "--no-color", fmt.Sprintf("origin/%s...%s", originBranch, currentBranch))
	if err != nil {
		// Try without origin prefix if it fails, maybe it's a local branch comparison
		diffOutput, err = git.RunGit(ctx, "diff", "-U0", "--no-color", fmt.Sprintf("%s...%s", originBranch, currentBranch))
		if err != nil {
			return fmt.Errorf("failed to get git diff: %w", err)
		}
	}

	if diffOutput == "" {
		pterm.Warning.Println("No changes detected between branches.")
		return nil
	}

	// 2. Initialize Agents
	pool := agent.NewAgentPool()

	// Key Extractor
	keyExtractor := NewKeyExtractor(diffOutput)
	pool.WithAgent(keyExtractor)

	// Translation Generator
	runtimeCtx, err := shared.BuildRuntimeContext()
	if err != nil {
		return fmt.Errorf("failed to build runtime context: %w", err)
	}

	llmService, err := llm.NewServiceBuilder(runtimeCtx).UseHeavyModel().Build()
	if err != nil {
		return fmt.Errorf("failed to build LLM service: %w", err)
	}

	translationGenerator := NewTranslationGenerator(llmService)
	pool.WithAgent(translationGenerator)

	// Translation Enhancer
	pool.WithAgent(NewTranslationEnhancer(llmService))

	// SQL Generator
	pool.WithAgent(NewSQLGenerator())

	// Execute Agents
	spinner, _ := pterm.DefaultSpinner.Start("Analyzing code, extracting keys, and generating translations...")
	results, err := pool.ExecuteAgents(nil)
	if err != nil {
		spinner.Fail("Agent execution failed: " + err.Error())
		return err
	}
	spinner.Success("Analysis complete!")

	// 3. Process Results
	// We are interested in the final output from TranslationEnhancer (for JSON/Tolgee) and SQLGenerator (for SQL)

	// Check if keys were found
	keysJSON := results["key_extractor"]
	var keys []I18nKey
	if err := json.Unmarshal([]byte(keysJSON), &keys); err != nil {
		return fmt.Errorf("failed to parse extracted keys: %w", err)
	}

	if len(keys) == 0 {
		pterm.Info.Println("No new i18n keys found.")
		return nil
	}

	pterm.Success.Printf("Found %d new keys.\n", len(keys))

	// Get Enhanced Translations
	translationsJSON := results["translation_enhancer"]
	var translationData TranslationData
	if err := json.Unmarshal([]byte(translationsJSON), &translationData); err != nil {
		// Try to unmarshal as array of keys directly
		var keysWithTrans []I18nKey
		if err2 := json.Unmarshal([]byte(translationsJSON), &keysWithTrans); err2 == nil {
			translationData = TranslationData{Keys: keysWithTrans}
		} else {
			pterm.Warning.Printf("Failed to parse translations JSON: %v\n", err)
			return nil
		}
	}

	// Display Results
	pterm.DefaultSection.Println("Generated Translations")
	for _, k := range translationData.Keys {
		pterm.Println(pterm.Cyan(k.Key))
		for lang, trans := range k.Translations {
			pterm.Printf("  %s: %s\n", strings.ToUpper(lang), trans)
		}
		pterm.Println()
	}

	// Output Handling
	if !autoConfirm {
		confirmed, _ := pterm.DefaultInteractiveConfirm.Show("Save these translations?")
		if !confirmed {
			pterm.Info.Println("Aborted.")
			return nil
		}
	}

	// Save JSON
	if err := createTranslationFile(&translationData); err != nil {
		pterm.Error.Println("Failed to save JSON file:", err)
	}

	// Save SQL
	sqlScript := results["sql_generator"]
	if err := createSQLFile(sqlScript); err != nil {
		pterm.Error.Println("Failed to save SQL file:", err)
	}

	// Save Tolgee
	if tolgeeOutput {
		if err := createTolgeeFiles(&translationData); err != nil {
			pterm.Error.Println("Failed to save Tolgee files:", err)
		}
	}

	return nil
}

func createTranslationFile(data *TranslationData) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	filename := outputFile
	if filename == "" {
		filename = "i18n_translations.json"
	}
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return err
	}
	pterm.Success.Println("Saved translations to " + filename)
	return nil
}

func createSQLFile(content string) error {
	filename := "i18n_insert.sql"
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return err
	}
	pterm.Success.Println("Saved SQL script to " + filename)
	return nil
}

func createTolgeeFiles(data *TranslationData) error {
	// Convert to map[lang]map[key]value
	langMaps := make(map[string]map[string]string)

	for _, k := range data.Keys {
		for lang, trans := range k.Translations {
			if _, ok := langMaps[lang]; !ok {
				langMaps[lang] = make(map[string]string)
			}
			langMaps[lang][k.Key] = trans
		}
	}

	var savedFiles []string
	for lang, content := range langMaps {
		filename := fmt.Sprintf("%s.json", lang)
		if err := writeJSONFile(filename, content); err != nil {
			return err
		}
		savedFiles = append(savedFiles, filename)
	}

	pterm.Success.Printf("Saved Tolgee files (%s)\n", strings.Join(savedFiles, ", "))
	return nil
}

func writeJSONFile(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}
