/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package llm

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	openaiShared "github.com/openai/openai-go/shared"

	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

// ModelVariant defines the logical model buckets (light/heavy/fallback) available to commands.
type ModelVariant int

const (
	ModelVariantHeavy ModelVariant = iota
	ModelVariantLight
	ModelVariantFallback
)

// ServiceBuilder lets commands configure which model tier, API key, and endpoint should be used.
type ServiceBuilder struct {
	runtime         *shared.RuntimeContext
	variant         ModelVariant
	customModel     string
	apiKeyOverride  string
	baseURLOverride string
	httpClient      *http.Client
}

// NewServiceBuilder creates a builder tied to the given runtime context.
func NewServiceBuilder(runtime *shared.RuntimeContext) *ServiceBuilder {
	return &ServiceBuilder{
		runtime: runtime,
		variant: ModelVariantHeavy,
	}
}

// UseLightModel selects the light model tier configured in the runtime context.
func (b *ServiceBuilder) UseLightModel() *ServiceBuilder {
	b.variant = ModelVariantLight
	return b
}

// UseHeavyModel selects the heavy model tier (default).
func (b *ServiceBuilder) UseHeavyModel() *ServiceBuilder {
	b.variant = ModelVariantHeavy
	return b
}

// UseFallbackModel selects the configured fallback model tier.
func (b *ServiceBuilder) UseFallbackModel() *ServiceBuilder {
	b.variant = ModelVariantFallback
	return b
}

// WithModel overrides the model identifier entirely.
func (b *ServiceBuilder) WithModel(model string) *ServiceBuilder {
	b.customModel = strings.TrimSpace(model)
	return b
}

// WithAPIKey overrides the API key used for the service.
func (b *ServiceBuilder) WithAPIKey(key string) *ServiceBuilder {
	b.apiKeyOverride = strings.TrimSpace(key)
	return b
}

// WithBaseURL overrides the base URL used for the service.
func (b *ServiceBuilder) WithBaseURL(baseURL string) *ServiceBuilder {
	b.baseURLOverride = strings.TrimSpace(baseURL)
	return b
}

// WithHTTPClient overrides the HTTP client (otherwise RuntimeContext HTTP client is reused).
func (b *ServiceBuilder) WithHTTPClient(client *http.Client) *ServiceBuilder {
	b.httpClient = client
	return b
}

// Build resolves the requested configuration and returns a ready-to-use LLM service.
func (b *ServiceBuilder) Build() (*Service, error) {
	if b.runtime == nil {
		return nil, fmt.Errorf("runtime context is required")
	}

	model, endpoint := b.resolveVariantConfig()
	if b.customModel != "" {
		model = b.customModel
	}
	if model == "" {
		return nil, fmt.Errorf("model is not configured for the selected variant")
	}

	apiKey := firstNonEmpty(b.apiKeyOverride, endpoint.APIKey, b.runtime.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("api key is not configured")
	}

	baseURL := firstNonEmpty(b.baseURLOverride, endpoint.BaseURL, b.runtime.BaseURL, providerDefaultBaseURL(b.runtime.Provider))
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is not configured")
	}

	httpClient := b.httpClient
	if httpClient == nil {
		httpClient = b.runtime.HTTPClient
	}
	if httpClient == nil {
		httpClient = shared.DefaultHTTPClient()
	}

	trimmedBaseURL := strings.TrimRight(baseURL, "/")
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(trimmedBaseURL),
		option.WithHTTPClient(httpClient),
	)

	return &Service{
		provider: b.runtime.Provider,
		model:    model,
		apiKey:   apiKey,
		baseURL:  trimmedBaseURL,
		client:   client,
	}, nil
}

func (b *ServiceBuilder) resolveVariantConfig() (string, shared.ModelEndpoint) {
	switch b.variant {
	case ModelVariantLight:
		return b.runtime.LightModel, b.runtime.LightEndpoint
	case ModelVariantFallback:
		return b.runtime.Fallback, b.runtime.FallbackEndpoint
	default:
		return b.runtime.HeavyModel, b.runtime.HeavyEndpoint
	}
}

// Service executes LLM requests using the resolved configuration.
type Service struct {
	provider string
	model    string
	apiKey   string
	baseURL  string
	client   openai.Client
}

// ChatMessage represents a message in a chat completion request.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents a chat completion call.
type ChatCompletionRequest struct {
	Messages         []ChatMessage
	Temperature      float64
	MaxTokens        float64
	TopP             float64
	FrequencyPenalty float64
	PresencePenalty  float64
}

// ChatCompletion sends a chat completion request and returns the assistant response text.
func (s *Service) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (string, error) {
	if len(req.Messages) == 0 {
		return "", fmt.Errorf("at least one message is required")
	}

	messages, err := buildMessageParams(req.Messages)
	if err != nil {
		return "", err
	}

	temperature := req.Temperature
	if req.Temperature == 0 {
		temperature = 0.2
	}

	params := openai.ChatCompletionNewParams{
		Model:    openaiShared.ChatModel(s.model),
		Messages: messages,
	}

	if req.MaxTokens > 0 {
		params.MaxCompletionTokens = openai.Int(int64(req.MaxTokens))
		params.MaxTokens = openai.Int(int64(req.MaxTokens))
	}
	params.Temperature = openai.Float(temperature)
	if req.TopP != 0 {
		params.TopP = openai.Float(req.TopP)
	}
	if req.FrequencyPenalty != 0 {
		params.FrequencyPenalty = openai.Float(req.FrequencyPenalty)
	}
	if req.PresencePenalty != 0 {
		params.PresencePenalty = openai.Float(req.PresencePenalty)
	}

	resp, err := s.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("chat completion request failed: %w", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("provider response did not contain a message")
	}

	return resp.Choices[0].Message.Content, nil
}

func buildMessageParams(messages []ChatMessage) ([]openai.ChatCompletionMessageParamUnion, error) {
	results := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))
	for _, msg := range messages {
		role := strings.ToLower(strings.TrimSpace(msg.Role))
		var param openai.ChatCompletionMessageParamUnion
		switch role {
		case "system":
			param = openai.SystemMessage(msg.Content)
		case "user":
			param = openai.UserMessage(msg.Content)
		case "assistant":
			param = openai.AssistantMessage(msg.Content)
		case "developer":
			// Developer role is not supported by every OpenAI-compatible API.
			// Treat it as a system instruction to remain compatible.
			param = openai.SystemMessage(msg.Content)
		default:
			return nil, fmt.Errorf("unsupported message role %q", msg.Role)
		}
		results = append(results, param)
	}
	return results, nil
}

func providerDefaultBaseURL(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai":
		return "https://api.openai.com/v1"
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
