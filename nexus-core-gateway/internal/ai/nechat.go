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
	"github.com/nexus-cyber/nexus-core-gateway/pkg/logger"
)

type NechatClient struct {
	APIKey   string
	Model    string
	Endpoint string
}

func NewNechatClient() *NechatClient {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	endpoint := os.Getenv("AI_PROVIDER_URL")
	model := os.Getenv("AI_MODEL_REASONING")
	if model == "" {
		model = "nexus-brain"
	}

	return &NechatClient{
		APIKey:   apiKey,
		Model:    model,
		Endpoint: endpoint,
	}
}

func (n *NechatClient) Chat(logs []logger.TelemetryLog, query string) (string, error) {
	// Hybrid System: Menggunakan nexus-brain dengan Fallback Expert System
	
	logsBytes, _ := json.MarshalIndent(logs, "", "  ")
	logsContext := string(logsBytes)
	
	// Truncate logs to save context window (Optimization)
	if len(logsContext) > 2000 {
		logsContext = logsContext[len(logsContext)-2000:]
	}

	systemPrompt := `Anda adalah NEXUS-SOC-BRAIN v2.5, Inti Kecerdasan Pertahanan Siber Otonom. 
Lingkungan: Nexus Command Center (NCC) | Teknologi: Moving Target Defense (MTD), PQC, Digital Hallucination.
Tugas: 
1. Analisis Log Telemetri secara taktis (Cari pola SQLi, XSS, Brute Force, Scanning).
2. Berikan laporan keamanan yang sangat detail, profesional, dan berfokus pada data.
3. Berikan saran taktis (misal: "Aktifkan Shield", "Lakukan Shuffle", "Investigasi IP").
4. Jelaskan bagaimana MTD (port rotation) melindungi Target (Portfolio OJK).
Bahasa: Indonesia Profesional (Bahasa Intelijen SOC).`

	userPrompt := fmt.Sprintf("STATUS TELEMETRI SAAT INI:\n%s\n\nPERTANYAAN ADMIN: %s\n\nBerikan analisis mendalam:", logsContext, query)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model":    n.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"stream": false,
		"options": map[string]interface{}{
			"num_ctx":     4096, // Upgraded Context Window for Long-term Memory
			"temperature": 0.2,  // Focused & Precise Reasoning
			"top_p":       0.9,
			"repeat_penalty": 1.2,
		},
	})

	req, err := http.NewRequest("POST", n.Endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return n.generateExpertFallback(logs, query), nil
	}

	req.Header.Set("Content-Type", "application/json")
	
	// Timeout 45 detik untuk memberikan ruang load model di RAM
	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[AI-BUSY] Mengaktifkan Expert Fallback: %v\n", err)
		return n.generateExpertFallback(logs, query), nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return n.generateExpertFallback(logs, query), nil
	}

	var orResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &orResp); err != nil || len(orResp.Choices) == 0 {
		return n.generateExpertFallback(logs, query), nil
	}

	return orResp.Choices[0].Message.Content, nil
}

func (n *NechatClient) generateExpertFallback(logs []logger.TelemetryLog, query string) string {
	q := strings.ToLower(query)
	
	// Analyze Logs for Heuristics
	sqli := 0
	brute := 0
	blocked := 0
	for _, l := range logs {
		status := strings.ToUpper(l.Status)
		if strings.Contains(status, "SQL") || strings.Contains(strings.ToUpper(l.Endpoint), "SELECT") {
			sqli++
		}
		if strings.Contains(status, "BLOCKED") || strings.Contains(status, "DROPPED") {
			blocked++
		}
		if strings.Contains(status, "AUTH") || strings.Contains(status, "LOGIN") {
			brute++
		}
	}

	res := "🛡️ **NEXUS EXPERT ANALYST (Operational Mode)**\n\n"
	
	// Smart Keyword Dispatcher
	if strings.Contains(q, "apa") || strings.Contains(q, "siapa") || strings.Contains(q, "tahu") {
		res += "Saya adalah modul analisis otonom Nexus. Saya memantau trafik secara real-time dan menggunakan MTD (Moving Target Defense) untuk melindungi aset Anda.\n\n"
	}
	
	if strings.Contains(q, "website") || strings.Contains(q, "lindung") {
		res += "Saat ini saya sedang melindungi portal portfolio yang terhubung ke data center simulasi OJK. Sistem MTD aktif merotasi port setiap 60 detik untuk membingungkan penyerang.\n\n"
	}

	if sqli > 0 {
		res += fmt.Sprintf("⚠️ **Peringatan:** Terdeteksi %d upaya SQL Injection. Sistem telah memblokir IP tersebut secara otomatis.\n", sqli)
	} else if blocked > 0 {
		res += fmt.Sprintf("ℹ️ **Status:** Menangani %d paket mencurigakan dalam 5 menit terakhir. Kondisi stabil.\n", blocked)
	} else {
		res += "✅ **Parameter Optimal:** Tidak ada ancaman aktif yang terdeteksi dalam telemetri saat ini.\n"
	}

	if strings.Contains(q, "help") || strings.Contains(q, "bantu") {
		res += "\nAnda dapat menanyakan tentang 'status trafik', 'serangan aktif', atau 'metode perlindungan' kami."
	}

	res += "\n\n*Catatan: Menggunakan analisis heuristik lokal karena AI Core (Ollama) sedang offline atau sinkronisasi.*"
	return res
}
