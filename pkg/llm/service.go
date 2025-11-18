/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
**/
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

	client := b.httpClient
	if client == nil {
		client = b.runtime.HTTPClient
	}
	if client == nil {
		client = shared.DefaultHTTPClient()
	}

	return &Service{
		provider:   b.runtime.Provider,
		model:      model,
		apiKey:     apiKey,
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: client,
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
	provider   string
	model      string
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// ChatMessage represents a message in a chat completion request.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents a chat completion call.
type ChatCompletionRequest struct {
	Messages    []ChatMessage
	Temperature float32
}

// ChatCompletion sends a chat completion request and returns the assistant response text.
func (s *Service) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (string, error) {
	if len(req.Messages) == 0 {
		return "", fmt.Errorf("at least one message is required")
	}

	payload := openAIChatRequest{
		Model:       s.model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to encode completion request: %w", err)
	}

	url := s.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create completion request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("chat completion request failed: %w", err)
	}
	defer resp.Body.Close()

	var parsed openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error.Message != "" {
			return "", fmt.Errorf("provider error: %s", parsed.Error.Message)
		}
		return "", fmt.Errorf("provider returned %s", resp.Status)
	}

	if len(parsed.Choices) == 0 || parsed.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("provider response did not contain a message")
	}

	return parsed.Choices[0].Message.Content, nil
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float32       `json:"temperature"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
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
