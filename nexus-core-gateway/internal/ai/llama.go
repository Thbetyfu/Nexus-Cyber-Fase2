// Package ai mengimplementasikan integrasi kecerdasan buatan untuk analisis forensik siber tingkat lanjut.
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

// LlamaClient mengimplementasikan Reasoning Layer (Pemberi Keputusan Forensik Fase 2).
// Menggunakan API OpenRouter untuk mengakses model Qwen3 235B-A22B — model open-weight dengan penalaran tertinggi
// untuk menganalisis Advanced Persistent Threat (APT) secara mendalam dan mendeteksi False Positives.
//
// Alasan Arsitektural (Why):
// Kunci API dimuat secara dinamis dari environment (`OPENROUTER_API_KEY`) dan TIDAK PERNAH di-hardcode
// demi mematuhi prinsip Zero Trust (ISO 27001 - Kontrol A.10 Cryptography & A.18 Compliance).
type LlamaClient struct {
	APIKey   string
	Model    string
	Endpoint string
}

// MitigationAction merepresentasikan rekomendasi penanganan otomatis yang dihasilkan secara cerdas oleh AI.
type MitigationAction struct {
	ActionType string                 `json:"action_type"` // BLOCK_IP | ISOLATE | PATCH | REDIRECT_HONEYPOT | SHUFFLE_MTD
	Priority   string                 `json:"priority"`    // CRITICAL | HIGH | MEDIUM
	Parameters map[string]interface{} `json:"parameters"`  // Parameter dinamis tambahan (seperti durasi IP block)
}

// LlamaForensicResult menyimpan respon analisis forensik terstruktur dari Llama/Qwen3-235B.
type LlamaForensicResult struct {
	ThreatVerdict     string             `json:"threat_verdict"` // Verdict: CONFIRMED_MALICIOUS | FALSE_POSITIVE | ADVANCED_PERSISTENT
	AttackerIntent    string             `json:"attacker_intent"`
	AttackVector      string             `json:"attack_vector"`
	Confidence        float64            `json:"confidence"`
	MitigationActions []MitigationAction `json:"mitigation_actions"`
	ForensicSummary   string             `json:"forensic_summary"` // Laporan forensik formal untuk regulator
}

// AttackContext membungkus data historis dan intelijen ancaman dinamis yang dikirim ke AI.
type AttackContext struct {
	AttackHistory []map[string]interface{} `json:"attack_history"` // Riwayat serangan IP yang sama dalam 24 jam terakhir
	ThreatIntel   map[string]interface{}   `json:"threat_intel"`   // Umpan ancaman STIX/TAXII dari BSSN atau ID-CERT
	SystemState   SystemState              `json:"system_state"`
}

// SystemState mencerminkan status kesehatan infrastruktur saat ini.
type SystemState struct {
	ActiveIncidents   int    `json:"active_incidents"`
	LastMTDShuffle    string `json:"last_mtd_shuffle"`
	CurrentAlertLevel string `json:"current_alert_level"`
}

// openRouterRequest mewakili skema payload chat completions OpenAI-compatible untuk OpenRouter.
type openRouterRequest struct {
	Model    string          `json:"model"`
	Messages []openRouterMsg `json:"messages"`
}

type openRouterMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// LLAMA_SYSTEM_PROMPT mendikte kepribadian, misi, dan kontrak output dari Reasoning Layer.
// Insting AI diarahkan untuk melindungi infrastruktur vital Indonesia (OJK, Bank Indonesia)
// serta mematuhi UU Pelindungan Data Pribadi (UU PDP) No. 27/2022.
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

// NewLlamaClient mengkonstruksi client LlamaClient secara dinamis.
//
// Alasan Arsitektural (Why):
// Dirancang fleksibel untuk mendukung endpoint on-premise lokal (seperti vLLM/Ollama) demi kerahasiaan data kedaulatan negara,
// dengan fallback otomatis ke OpenRouter jika server on-premise tidak dikonfigurasi.
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

// AnalyzeEscalatedThreat melakukan analisis forensik mendalam pada payload yang dicurigai.
//
// Alasan Arsitektural (Why):
// Fungsi ini didesain agar dijalankan secara asinkron (goroutine) dengan context budget 30 detik.
// Proses ini sengaja tidak memblokir (non-blocking) arus request trafik utama pengguna demi menjamin
// tingkat kelancaran respons aplikasi (low latency proxy) tetap optimal.
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
	req.Header.Set("HTTP-Referer", "https://nexus-cyber.go.id") // Atribusi domain ke OpenRouter
	req.Header.Set("X-Title", "Nexus Cyber Gateway")

	// Anggaran waktu 30 detik agar model LLM raksasa memiliki waktu berpikir yang cukup di server inference.
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

// AnalyzeIntent adalah antarmuka pembungkus ramah warisan (legacy) untuk proxy_core.go.
func (l *LlamaClient) AnalyzeIntent(payload string) (*LlamaForensicResult, error) {
	ctx := AttackContext{
		AttackHistory: []map[string]interface{}{},
		ThreatIntel:   map[string]interface{}{},
		SystemState:   SystemState{ActiveIncidents: 0, LastMTDShuffle: "unknown", CurrentAlertLevel: "NORMAL"},
	}
	return l.AnalyzeEscalatedThreat(nil, payload, ctx)
}

// ParseLlamaResponse mengekstrak respon JSON Llama secara tangguh menggunakan 3-Stage Parser.
func ParseLlamaResponse(raw string) (*LlamaForensicResult, error) {
	raw = strings.TrimSpace(raw)
	var result LlamaForensicResult

	// Stage 1: Unmarshal Langsung
	if err := json.Unmarshal([]byte(raw), &result); err == nil {
		return &result, nil
	}

	// Stage 2: Ekstraksi dari blok kode markdown ```json ... ```
	if idx := strings.Index(raw, "```json"); idx != -1 {
		end := strings.Index(raw[idx+7:], "```")
		if end != -1 {
			if err := json.Unmarshal([]byte(strings.TrimSpace(raw[idx+7:idx+7+end])), &result); err == nil {
				return &result, nil
			}
		}
	}

	// Stage 3: Bracket Search (Kurung kurawal pertama ke penutup terakhir)
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
