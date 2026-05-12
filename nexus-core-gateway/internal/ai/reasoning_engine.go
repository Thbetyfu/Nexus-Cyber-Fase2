package ai

import (
	"fmt"
	"os"
)

// ReasoningEngine is the backward-compatible facade for the Reasoning Layer.
// Internally it now delegates to LlamaClient (OpenRouter / Qwen3-235B-A22B).
// This preserves the interface expected by proxy_core.go without changes.
type ReasoningEngine struct {
	client  *LlamaClient
	Enabled bool
}

// NewReasoningEngine constructs a ReasoningEngine backed by OpenRouter.
// The `url` and `model` parameters are kept for API compatibility but
// the URL is now sourced from OpenRouter and model from the argument.
func NewReasoningEngine(url, model string) *ReasoningEngine {
	// If model is still the old "llama3" default, upgrade it silently.
	if model == "llama3" || model == "" {
		model = "qwen/qwen3-235b-a22b"
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Println("[WARN] OPENROUTER_API_KEY not set — Reasoning Layer degraded (Fail-Open).")
	}

	return &ReasoningEngine{
		client:  NewLlamaClient(model),
		Enabled: true,
	}
}

// AnalyzeIntent is the main interface used by proxy_core.go.
// Returns the full forensic result from the AI.
func (re *ReasoningEngine) AnalyzeIntent(payload string) (*LlamaForensicResult, error) {
	if !re.Enabled {
		return nil, fmt.Errorf("reasoning engine disabled")
	}
	return re.client.AnalyzeIntent(payload)
}
