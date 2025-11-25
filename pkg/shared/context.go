package shared

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// RuntimeContext exposes sanitized configuration and shared dependencies that can be reused
// across commands without reaching into unrelated domains directly.
type RuntimeContext struct {
	Provider         string
	BaseURL          string
	APIKey           string
	LightModel       string
	HeavyModel       string
	Fallback         string
	LightEndpoint    ModelEndpoint
	HeavyEndpoint    ModelEndpoint
	FallbackEndpoint ModelEndpoint
	HTTPClient       *http.Client
	AnalysisTimeout  time.Duration
	WriterTimeout    time.Duration
}

// ModelEndpoint describes the credentials and endpoint overrides for a specific model class.
type ModelEndpoint struct {
	APIKey   string
	BaseURL  string
	Provider string
}

var (
	defaultHTTPClient     *http.Client
	defaultHTTPClientOnce sync.Once
)

// DefaultHTTPClient returns a hardened HTTP client shared by commands that need to
// communicate with AI providers.
func DefaultHTTPClient() *http.Client {
	defaultHTTPClientOnce.Do(func() {
		defaultHTTPClient = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
			},
		}
	})

	return defaultHTTPClient
}

// BuildRuntimeContext constructs a RuntimeContext from viper configuration and returns
// actionable pointers commands can share without duplicating sensitive logic.
func BuildRuntimeContext() (*RuntimeContext, error) {
	apiKey := strings.TrimSpace(viper.GetString("api.key"))
	if apiKey == "" {
		return nil, fmt.Errorf("missing api.key in configuration")
	}

	provider := strings.TrimSpace(viper.GetString("api.provider"))
	if provider == "" {
		provider = "openai"
	}

	globalBaseURL := strings.TrimSpace(viper.GetString("api.base_url"))

	ctx := &RuntimeContext{
		Provider:   provider,
		BaseURL:    globalBaseURL,
		APIKey:     apiKey,
		LightModel: strings.TrimSpace(viper.GetString("api.light_model")),
		HeavyModel: strings.TrimSpace(viper.GetString("api.heavy_model")),
		Fallback:   strings.TrimSpace(viper.GetString("api.fallback_model")),
		LightEndpoint: ModelEndpoint{
			APIKey:   fallbackString(strings.TrimSpace(viper.GetString("api.light.api_key")), apiKey),
			BaseURL:  fallbackString(strings.TrimSpace(viper.GetString("api.light.base_url")), globalBaseURL),
			Provider: fallbackString(strings.TrimSpace(viper.GetString("api.light.provider")), provider),
		},
		HeavyEndpoint: ModelEndpoint{
			APIKey:   fallbackString(strings.TrimSpace(viper.GetString("api.heavy.api_key")), apiKey),
			BaseURL:  fallbackString(strings.TrimSpace(viper.GetString("api.heavy.base_url")), globalBaseURL),
			Provider: fallbackString(strings.TrimSpace(viper.GetString("api.heavy.provider")), provider),
		},
		FallbackEndpoint: ModelEndpoint{
			APIKey:   fallbackString(strings.TrimSpace(viper.GetString("api.fallback.api_key")), apiKey),
			BaseURL:  fallbackString(strings.TrimSpace(viper.GetString("api.fallback.base_url")), globalBaseURL),
			Provider: fallbackString(strings.TrimSpace(viper.GetString("api.fallback.provider")), provider),
		},
		HTTPClient:      DefaultHTTPClient(),
		AnalysisTimeout: getDurationOrDefault("agent.analysis.timeout", 5*time.Minute),
		WriterTimeout:   getDurationOrDefault("agent.writer.timeout", 5*time.Minute),
	}

	return ctx, nil
}

// RedactedCopy provides a safe snapshot that can be logged or passed to prompts
// without leaking the API key.
func (rc *RuntimeContext) RedactedCopy() RuntimeContext {
	clone := *rc
	if clone.APIKey != "" {
		clone.APIKey = "***REDACTED***"
	}
	if clone.LightEndpoint.APIKey != "" {
		clone.LightEndpoint.APIKey = "***REDACTED***"
	}
	if clone.HeavyEndpoint.APIKey != "" {
		clone.HeavyEndpoint.APIKey = "***REDACTED***"
	}
	if clone.FallbackEndpoint.APIKey != "" {
		clone.FallbackEndpoint.APIKey = "***REDACTED***"
	}

	return clone
}

func fallbackString(primary, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}

func getDurationOrDefault(key string, defaultVal time.Duration) time.Duration {
	val := viper.GetDuration(key)
	if val == 0 {
		return defaultVal
	}
	return val
}
