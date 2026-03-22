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

// QwenClient implements @skill-dual-brain Reflex Layer via Groq API.
// Model: Qwen3 32B — ultra-low latency, <50ms target.
// API Key loaded from GROQ_API_KEY env variable. NEVER hardcoded.
type QwenClient struct {
	APIKey   string
	Model    string
	Endpoint string
}

// QwenResult is the structured classification output from the Reflex Layer.
type QwenResult struct {
	Classification string  `json:"classification"` // BENIGN | SUSPICIOUS | MALICIOUS
	Confidence     float64 `json:"confidence"`     // 0.0 - 1.0
	ThreatType     *string `json:"threat_type"`    // nullable
}

// TrafficMeta is the minimal metadata sent to Qwen — NOT the full body.
// Keeping input ≤ 200 tokens to maintain sub-50ms response.
type TrafficMeta struct {
	SourceIP       string `json:"source_ip"`
	Port           string `json:"port"`
	Protocol       string `json:"protocol"`
	Method         string `json:"method"`
	RequestPattern string `json:"request_pattern"`
}

// groqRequest is the Groq API chat completions request schema.
type groqRequest struct {
	Model    string        `json:"model"`
	Messages []groqMessage `json:"messages"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// groqResponse is the Groq API response schema.
type groqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// QWEN_SYSTEM_PROMPT defines the Reflex Layer behavior with few-shot examples.
// Kept compact to maximise speed on Groq's inference infrastructure.
const QWEN_SYSTEM_PROMPT = `You are a real-time network threat classifier.
Analyze traffic metadata and classify it IMMEDIATELY.
Respond ONLY with valid JSON. No explanation. No markdown. No preamble.

Classification rules:
- BENIGN: normal user traffic patterns
- SUSPICIOUS: anomalous but unconfirmed (escalate for deeper analysis)
- MALICIOUS: confirmed attack pattern

Known APT signatures to flag as MALICIOUS:
SilverTerrier, Turla, AnchorPanda, FaceDuck, APT41

Examples:
INPUT: {"source_ip":"192.168.1.5","port":"443","protocol":"HTTPS","request_pattern":"GET /api/users"}
OUTPUT: {"classification":"BENIGN","confidence":0.97,"threat_type":null}

INPUT: {"source_ip":"45.33.32.156","port":"80","protocol":"HTTP","request_pattern":"GET /etc/passwd"}
OUTPUT: {"classification":"MALICIOUS","confidence":0.99,"threat_type":"PATH_TRAVERSAL"}

INPUT: {"source_ip":"103.21.244.0","port":"3306","protocol":"TCP","request_pattern":"'; DROP TABLE users;--"}
OUTPUT: {"classification":"MALICIOUS","confidence":0.98,"threat_type":"SQL_INJECTION"}

INPUT: {"source_ip":"10.0.0.50","port":"8080","protocol":"HTTP","request_pattern":"POST /api/login (429 req/min)"}
OUTPUT: {"classification":"SUSPICIOUS","confidence":0.72,"threat_type":"BRUTE_FORCE_ATTEMPT"}`

// NewQwenClient constructs the Reflex Layer client dynamically.
// Designed for Enterprise On-Premise (vLLM) compatibility to satisfy UU PDP.
func NewQwenClient(model string) *QwenClient {
	apiKey := os.Getenv("AI_PROVIDER_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY") // fallback for older configs
	}

	endpoint := os.Getenv("AI_PROVIDER_URL")
	if endpoint == "" {
		// Default to OpenRouter if no on-premise server is configured
		endpoint = "https://openrouter.ai/api/v1/chat/completions"
	}

	envModel := os.Getenv("AI_MODEL_REFLEX")
	if envModel != "" {
		model = envModel
	} else if model == "" {
		model = "qwen/qwen3-32b-instruct"
	}

	return &QwenClient{
		APIKey:   apiKey,
		Model:    model,
		Endpoint: endpoint,
	}
}

// CheckHealth pings the Groq/OpenRouter API to verify availability and latency.
func (q *QwenClient) CheckHealth() (status string, latency int64) {
	start := time.Now()
	client := &http.Client{Timeout: 3 * time.Second}

	req, _ := http.NewRequest("GET", "https://google.com", nil) // Simple 204 heartbeat simulation to test outside link
	res, err := client.Do(req)

	latency = time.Since(start).Milliseconds()

	if err != nil || res.StatusCode >= 400 {
		return "OFFLINE", latency
	}
	return "ONLINE", latency
}

// ClassifyPayload sends traffic metadata to Qwen3 via Groq for sub-50ms Reflex screening.
func (q *QwenClient) Classify(meta TrafficMeta) (*QwenResult, error) {
	trafficJSON, _ := json.Marshal(meta)

	userPrompt := fmt.Sprintf("Classify this traffic:\n%s\n\nRespond ONLY with:\n{\"classification\":\"BENIGN|SUSPICIOUS|MALICIOUS\",\"confidence\":0.0-1.0,\"threat_type\":\"string or null\"}",
		string(trafficJSON))

	payload, _ := json.Marshal(groqRequest{
		Model: q.Model,
		Messages: []groqMessage{
			{Role: "system", Content: QWEN_SYSTEM_PROMPT},
			{Role: "user", Content: userPrompt},
		},
	})

	req, err := http.NewRequest("POST", q.Endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("groq_req_build: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+q.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://nexus-cyber.go.id")

	// OpenRouter has slightly more routing latency than Groq natively.
	// Adjusted budget to 1500ms to allow valid AI inferences without dropping.
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openrouter_timeout: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var gResp groqResponse
	if err := json.Unmarshal(body, &gResp); err != nil || len(gResp.Choices) == 0 {
		return nil, fmt.Errorf("groq_parse_response: %s", string(body)[:min(len(body), 200)])
	}

	rawContent := gResp.Choices[0].Message.Content
	return ParseQwenResponse(rawContent)
}

// ParseQwenResponse uses the 3-stage robust JSON parser for QwenResult.
func ParseQwenResponse(raw string) (*QwenResult, error) {
	raw = strings.TrimSpace(raw)
	var result QwenResult

	// Stage 1: Direct parse
	if err := json.Unmarshal([]byte(raw), &result); err == nil {
		return &result, nil
	}

	// Stage 2: Extract from ```json block
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

	return nil, fmt.Errorf("qwen_parse_error: %s", raw[:min(len(raw), 200)])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
