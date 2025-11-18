package llm

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

func TestServiceBuilderUsesVariantConfig(t *testing.T) {
	rt := &shared.RuntimeContext{
		Provider:   "openai",
		APIKey:     "global",
		BaseURL:    "https://api.example.com/v1",
		LightModel: "gpt-3.5",
		HeavyModel: "gpt-4",
		Fallback:   "gpt-3.5",
		LightEndpoint: shared.ModelEndpoint{
			APIKey:  "light-key",
			BaseURL: "https://light.example.com/v1/",
		},
		HeavyEndpoint: shared.ModelEndpoint{
			APIKey:  "heavy-key",
			BaseURL: "",
		},
		HTTPClient: shared.DefaultHTTPClient(),
	}

	service, err := NewServiceBuilder(rt).UseLightModel().Build()
	if err != nil {
		t.Fatalf("unexpected error building service: %v", err)
	}

	if service.model != "gpt-3.5" {
		t.Fatalf("expected light model, got %s", service.model)
	}
	if service.apiKey != "light-key" {
		t.Fatalf("expected light API key, got %s", service.apiKey)
	}
	if service.baseURL != "https://light.example.com/v1" {
		t.Fatalf("expected trimmed base URL, got %s", service.baseURL)
	}
}

func TestServiceChatCompletion(t *testing.T) {
	var path string
	testClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			path = req.URL.Path
			if got := req.Header.Get("Authorization"); got != "Bearer heavy-key" {
				t.Fatalf("unexpected authorization header %s", got)
			}
			body := io.NopCloser(strings.NewReader(successfulChatCompletionResponse))
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       body,
				Header:     make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		}),
	}
	rt := &shared.RuntimeContext{
		Provider:   "openai",
		APIKey:     "global",
		HeavyModel: "gpt-4",
		HeavyEndpoint: shared.ModelEndpoint{
			APIKey:  "heavy-key",
			BaseURL: "https://example.com",
		},
		HTTPClient: testClient,
	}

	service, err := NewServiceBuilder(rt).UseHeavyModel().Build()
	if err != nil {
		t.Fatalf("unexpected error building service: %v", err)
	}

	resp, err := service.ChatCompletion(context.Background(), ChatCompletionRequest{
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("completion failed: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("unexpected response %s", resp)
	}
	if path != "/chat/completions" {
		t.Fatalf("unexpected request path %s", path)
	}
}

const successfulChatCompletionResponse = `{
  "id": "chatcmpl-test",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "gpt-4",
  "service_tier": "default",
  "system_fingerprint": "fp",
  "choices": [
    {
      "index": 0,
      "finish_reason": "stop",
      "logprobs": {
        "content": [],
        "refusal": []
      },
      "message": {
        "role": "assistant",
        "content": "ok",
        "refusal": "",
        "annotations": [],
        "tool_calls": [],
        "audio": null,
        "function_call": {
          "arguments": "",
          "name": ""
        }
      }
    }
  ],
  "usage": {
    "prompt_tokens": 1,
    "completion_tokens": 1,
    "total_tokens": 2,
    "completion_tokens_details": {
      "accepted_prediction_tokens": 0,
      "audio_tokens": 0,
      "reasoning_tokens": 0,
      "rejected_prediction_tokens": 0
    },
    "prompt_tokens_details": {
      "audio_tokens": 0,
      "cached_tokens": 0
    }
  }
}`

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
