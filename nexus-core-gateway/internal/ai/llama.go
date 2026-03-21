package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// LlamaClient implements @skill-dual-brain Reasoning Layer via OpenRouter API.
// Model: Qwen3 235B-A22B — highest-reasoning open-weight model for APT forensics.
// API Key loaded from OPENROUTER_API_KEY env variable. NEVER hardcoded.
type LlamaClient struct {
	APIKey   string
	Model    string
	Endpoint string
}

// MitigationAction represents an autonomous remediation step recommended by the AI.
type MitigationAction struct {
	ActionType string                 `json:"action_type"` // BLOCK_IP|ISOLATE|PATCH|REDIRECT_HONEYPOT|SHUFFLE_MTD
	Priority   string                 `json:"priority"`    // CRITICAL|HIGH|MEDIUM
	Parameters map[string]interface{} `json:"parameters"`
}

// LlamaForensicResult is the structured forensic output from the Reasoning Layer.
type LlamaForensicResult struct {
	ThreatVerdict     string             `json:"threat_verdict"` // CONFIRMED_MALICIOUS|FALSE_POSITIVE|ADVANCED_PERSISTENT
	AttackerIntent    string             `json:"attacker_intent"`
	AttackVector      string             `json:"attack_vector"`
	Confidence        float64            `json:"confidence"`
	MitigationActions []MitigationAction `json:"mitigation_actions"`
	ForensicSummary   string             `json:"forensic_summary"`
}

// AttackContext holds the rich dynamic context passed to the Reasoning Layer.
type AttackContext struct {
	AttackHistory []map[string]interface{} `json:"attack_history"`
	ThreatIntel   map[string]interface{}   `json:"threat_intel"`
	SystemState   SystemState              `json:"system_state"`
}

// SystemState holds current infrastructure telemetry.
type SystemState struct {
	ActiveIncidents   int    `json:"active_incidents"`
	LastMTDShuffle    string `json:"last_mtd_shuffle"`
	CurrentAlertLevel string `json:"current_alert_level"`
}

// openRouterRequest is the OpenRouter API chat completions request schema.
type openRouterRequest struct {
	Model    string          `json:"model"`
	Messages []openRouterMsg `json:"messages"`
}

type openRouterMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openRouterResponse mirrors the OpenAI-compatible response format.
type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// LLAMA_SYSTEM_PROMPT defines the Reasoning Layer's mission and output contract.
// Think step-by-step internally — final output MUST be JSON only.
const LLAMA_SYSTEM_PROMPT = `You are an expert cybersecurity analyst for Indonesia's critical 
digital infrastructure. You receive escalated threats that have been pre-screened 
by a rapid classifier (Qwen3 32B).

Your job:
1. Analyze the attacker's INTENT based on full context
2. Determine if this is a confirmed threat or false positive
3. Recommend specific autonomous mitigation actions
4. Write a forensic summary for regulators (BSSN, BI, OJK)

Think step by step internally, but your FINAL response must be ONLY valid JSON.
Indonesian context: Protect PDNS, BI, OJK infrastructure. Comply with UU PDP No. 27/2022.`

// NewLlamaClient constructs the Reasoning Layer client dynamically.
// Designed to connect to any OpenAI-compatible provider, such as local vLLM for data privacy.
func NewLlamaClient(model string) *LlamaClient {
	apiKey := os.Getenv("AI_PROVIDER_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	}

	endpoint := os.Getenv("AI_PROVIDER_URL")
	if endpoint == "" {
		endpoint = "https://openrouter.ai/api/v1/chat/completions"
	}

	envModel := os.Getenv("AI_MODEL_REASONING")
	if envModel != "" {
		model = envModel
	} else if model == "" {
		model = "qwen/qwen3-235b-a22b"
	}

	return &LlamaClient{
		APIKey:   apiKey,
		Model:    model,
		Endpoint: endpoint,
	}
}

// AnalyzeEscalatedThreat performs deep forensic analysis on a suspicious payload.
// Async-safe: call this inside a goroutine with a 30s timeout context.
func (l *LlamaClient) AnalyzeEscalatedThreat(qwenResult *QwenResult, payload string, ctx AttackContext) (*LlamaForensicResult, error) {
	qwenJSON, _ := json.MarshalIndent(qwenResult, "", "  ")
	historyJSON, _ := json.MarshalIndent(ctx.AttackHistory, "", "  ")
	intelJSON, _ := json.MarshalIndent(ctx.ThreatIntel, "", "  ")
	stateJSON, _ := json.MarshalIndent(ctx.SystemState, "", "  ")

	truncPayload := payload
	if len(truncPayload) > 300 {
		truncPayload = truncPayload[:300] + "...[TRUNCATED]"
	}

	userPrompt := fmt.Sprintf(`ESCALATED THREAT — Requires deep analysis.

=== Qwen3-32B Pre-screening Result ===
%s

=== Attack History (same source IP, last 24h) ===
%s

=== Threat Intelligence (BSSN/ID-CERT STIX feeds) ===
%s

=== Current System State ===
%s

=== Raw Payload (sanitized) ===
%s

Analyze the attacker's intent. Consider APT patterns.
Respond ONLY with this JSON structure:
{
  "threat_verdict": "CONFIRMED_MALICIOUS|FALSE_POSITIVE|ADVANCED_PERSISTENT",
  "attacker_intent": "...",
  "attack_vector": "...",
  "confidence": 0.0,
  "mitigation_actions": [
    {"action_type": "BLOCK_IP|ISOLATE|PATCH|REDIRECT_HONEYPOT|SHUFFLE_MTD",
     "priority": "CRITICAL|HIGH|MEDIUM", "parameters": {}}
  ],
  "forensic_summary": "..."
}`,
		string(qwenJSON), string(historyJSON), string(intelJSON), string(stateJSON), truncPayload)

	reqBody, _ := json.Marshal(openRouterRequest{
		Model: l.Model,
		Messages: []openRouterMsg{
			{Role: "system", Content: LLAMA_SYSTEM_PROMPT},
			{Role: "user", Content: userPrompt},
		},
	})

	req, err := http.NewRequest("POST", l.Endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("openrouter_req_build: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+l.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://nexus-cyber.go.id") // OpenRouter attribution
	req.Header.Set("X-Title", "Nexus Cyber Gateway")

	// 30s budget for deep reasoning — runs async (non-blocking to main traffic)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openrouter_timeout: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var orResp openRouterResponse
	if err := json.Unmarshal(body, &orResp); err != nil || len(orResp.Choices) == 0 {
		return nil, fmt.Errorf("openrouter_parse_response: %s", string(body)[:min2(len(body), 200)])
	}

	rawContent := orResp.Choices[0].Message.Content
	return ParseLlamaResponse(rawContent)
}

// AnalyzeIntent is the legacy-compatible interface used by proxy_core.go.
// It wraps AnalyzeEscalatedThreat with minimal context for backward compatibility.
func (l *LlamaClient) AnalyzeIntent(payload string) (isMalicious bool, err error) {
	ctx := AttackContext{
		AttackHistory: []map[string]interface{}{},
		ThreatIntel:   map[string]interface{}{},
		SystemState:   SystemState{ActiveIncidents: 0, LastMTDShuffle: "unknown", CurrentAlertLevel: "NORMAL"},
	}
	result, err := l.AnalyzeEscalatedThreat(nil, payload, ctx)
	if err != nil {
		return false, err
	}
	return result.ThreatVerdict == "CONFIRMED_MALICIOUS" || result.ThreatVerdict == "ADVANCED_PERSISTENT", nil
}

// ParseLlamaResponse parses OpenRouter output using 3-stage robust JSON parsing.
func ParseLlamaResponse(raw string) (*LlamaForensicResult, error) {
	raw = strings.TrimSpace(raw)
	var result LlamaForensicResult

	// Stage 1: Direct parse
	if err := json.Unmarshal([]byte(raw), &result); err == nil {
		return &result, nil
	}

	// Stage 2: JSON code block extraction
	if idx := strings.Index(raw, "```json"); idx != -1 {
		end := strings.Index(raw[idx+7:], "```")
		if end != -1 {
			if err := json.Unmarshal([]byte(strings.TrimSpace(raw[idx+7:idx+7+end])), &result); err == nil {
				return &result, nil
			}
		}
	}

	// Stage 3: Bracket search
	start := strings.Index(raw, "{")
	last := strings.LastIndex(raw, "}")
	if start != -1 && last != -1 && last > start {
		if err := json.Unmarshal([]byte(raw[start:last+1]), &result); err == nil {
			return &result, nil
		}
	}

	return nil, fmt.Errorf("llama_parse_error: %s", raw[:min2(len(raw), 200)])
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}
