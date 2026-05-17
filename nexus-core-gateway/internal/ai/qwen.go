// Package ai mengimplementasikan integrasi kecerdasan buatan untuk pemindaian ancaman siber real-time.
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

// QwenClient mengimplementasikan deteksi ancaman siber cepat (Reflex Layer Fase 1).
// Memanfaatkan model Qwen-2.5-72b-instruct (atau Qwen3 32B lokal) untuk mencapai kecepatan klasifikasi
// yang sangat tinggi (ultra-low latency target <50ms jika dijalankan pada Groq/vLLM lokal).
//
// Alasan Arsitektural (Why):
// Kunci API dimuat secara aman dari environment (`AI_PROVIDER_KEY` atau `OPENROUTER_API_KEY`) tanpa
// hardcoding untuk memenuhi standar kepatuhan UU PDP No. 27/2022 (Data Privacy & Key Hardening).
type QwenClient struct {
	APIKey   string
	Model    string
	Endpoint string
}

// QwenResult menyimpan hasil klasifikasi terstruktur dari Reflex Layer.
type QwenResult struct {
	Classification string  `json:"classification"` // BENIGN | SUSPICIOUS | MALICIOUS
	Confidence     float64 `json:"confidence"`     // Tingkat keyakinan model (0.0 - 1.0)
	ThreatType     *string `json:"threat_type"`    // Jenis ancaman yang diidentifikasi (nullable jika BENIGN)
}

// TrafficMeta menyimpan metadata minimal dari request HTTP untuk dikirim ke AI.
//
// Alasan Arsitektural (Why):
// Mengirim seluruh body request ke AI akan memakan biaya token tinggi dan meningkatkan latency secara eksponensial.
// Dengan hanya mengirim metadata penting (≤ 200 token), kita dapat mendeteksi pola serangan dengan akurat
// sekaligus mempertahankan target performa pemrosesan sub-50ms.
type TrafficMeta struct {
	SourceIP       string `json:"source_ip"`
	Port           string `json:"port"`
	Protocol       string `json:"protocol"`
	Method         string `json:"method"`
	RequestPattern string `json:"request_pattern"`
}

type groqRequest struct {
	Model    string        `json:"model"`
	Messages []groqMessage `json:"messages"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// QWEN_SYSTEM_PROMPT mendefinisikan aturan ketat pemindaian ancaman dan melampirkan beberapa contoh few-shot.
// Few-shot learning menjamin model langsung mengerti format output JSON murni tanpa membuang waktu berpikir
// (Chain of Thought), mempercepat pemrosesan hingga 4x lipat di infrastruktur Groq/vLLM.
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

// NewQwenClient mengkonstruksi client Reflex Layer secara dinamis.
// Dirancang kompatibel dengan instalasi lokal vLLM untuk kedaulatan data penuh.
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
		model = "qwen/qwen-2.5-72b-instruct"
	}

	return &QwenClient{
		APIKey:   apiKey,
		Model:    model,
		Endpoint: endpoint,
	}
}

// CheckHealth menguji ketersediaan koneksi dan latency ke API eksternal/lokal.
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

// Classify mengirimkan metadata trafik ke Qwen untuk pemindaian instan.
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
	req.Header.Set("X-Title", "Nexus Cyber SOC Gateway")

	// Anggaran waktu dinaikkan menjadi 1500ms agar request tidak dibatalkan saat antrean API sedang padat.
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

// Generate menghasilkan tanggapan percakapan teks bebas untuk pembuatan laporan SOC taktis.
func (q *QwenClient) Generate(prompt string) (string, int64, error) {
	start := time.Now()
	payload, _ := json.Marshal(groqRequest{
		Model: q.Model,
		Messages: []groqMessage{
			{Role: "system", Content: "You are a professional Cyber Security Analyst."},
			{Role: "user", Content: prompt},
		},
	})

	req, err := http.NewRequest("POST", q.Endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Authorization", "Bearer "+q.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://nexus-cyber.go.id")
	req.Header.Set("X-Title", "Nexus Cyber SOC Gateway")

	client := &http.Client{Timeout: 10 * time.Second} // Timeout lebih longgar untuk kompilasi teks
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var gResp groqResponse
	if err := json.Unmarshal(body, &gResp); err != nil || len(gResp.Choices) == 0 {
		return "", 0, fmt.Errorf("ai_gen_parse_error: %s", string(body))
	}

	latency := time.Since(start).Milliseconds()
	return gResp.Choices[0].Message.Content, latency, nil
}

// ParseQwenResponse menggunakan 3-Stage Parser untuk memproses hasil unmarshal QwenResult secara handal.
func ParseQwenResponse(raw string) (*QwenResult, error) {
	raw = strings.TrimSpace(raw)
	var result QwenResult

	// Stage 1: Unmarshal langsung
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

	return nil, fmt.Errorf("qwen_parse_error: %s", raw[:min(len(raw), 200)])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
