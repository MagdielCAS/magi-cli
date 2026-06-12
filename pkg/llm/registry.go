/**
 * Copyright © 2025 Magdiel Campelo <github.com/MagdielCAS/magi-cli>
 * This file is part of the magi-cli
 **/
package llm

import (
	"fmt"
	"strings"
	"sync"

	"github.com/MagdielCAS/magi-cli/pkg/shared"
)

// Registry holds the registered capabilities and providers.
type Registry struct {
	mu           sync.RWMutex
	providers    map[string]IntelligenceProvider
	capabilities map[string]Capability
}

var globalRegistry = &Registry{
	providers:    make(map[string]IntelligenceProvider),
	capabilities: make(map[string]Capability),
}

func init() {
	RegisterCapability(&GenerateCommitMessageCapability{})
	RegisterCapability(&FixCommitMessageCapability{})
}

// RegisterProvider registers a static provider.
func RegisterProvider(name string, p IntelligenceProvider) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.providers[strings.ToLower(name)] = p
}

// RegisterCapability registers a task/intent capability.
func RegisterCapability(c Capability) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.capabilities[c.Name()] = c
}

// GetCapability retrieves a registered capability by name.
func GetCapability(name string) (Capability, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	c, ok := globalRegistry.capabilities[name]
	if !ok {
		return nil, fmt.Errorf("capability %q not registered", name)
	}
	return c, nil
}

// ResolveProvider dynamically resolves the active provider based on runtime context.
func ResolveProvider(runtime *shared.RuntimeContext) (IntelligenceProvider, error) {
	if runtime == nil {
		return nil, fmt.Errorf("runtime context is required to resolve provider")
	}

	providerName := strings.ToLower(strings.TrimSpace(runtime.Provider))
	switch providerName {
	case "openai", "custom":
		return NewOpenAIProvider(runtime), nil
	case "copilot_cli":
		return NewCopilotCLIProvider(runtime), nil
	case "claude_code":
		return NewClaudeCodeProvider(runtime), nil
	default:
		// Check static registrations as a fallback
		globalRegistry.mu.RLock()
		p, ok := globalRegistry.providers[providerName]
		globalRegistry.mu.RUnlock()
		if ok {
			return p, nil
		}
		return nil, fmt.Errorf("unsupported intelligence provider: %s", runtime.Provider)
	}
}
