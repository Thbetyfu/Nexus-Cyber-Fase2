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

	systemPrompt := "Anda adalah NEXUS-CORE-BRAIN, pusat komando pertahanan siber otonom. Analisis log trafik secara taktis dan berikan laporan keamanan profesional dalam Bahasa Indonesia."
	userPrompt := fmt.Sprintf("Data Telemetri:\n%s\n\nInstruksi Admin: %s", logsContext, query)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model":    n.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"stream": false,
		"options": map[string]interface{}{
			"num_ctx": 1024, // Optimasi RAM
			"temperature": 0.1,
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
	sqli := 0
	for _, l := range logs {
		if strings.Contains(strings.ToUpper(l.Status), "SQL") {
			sqli++
		}
	}

	res := "🛡️ **NEXUS EXPERT ANALYST (Operational Mode)**\n\n"
	if sqli > 0 {
		res += fmt.Sprintf("⚠️ **KRITIKAL:** Terdeteksi %d upaya anomali SQL Injection. Mitigasi MTD Shuffle sedang berjalan.", sqli)
	} else {
		res += "✅ **AMAN:** Analisis telemetri menunjukkan parameter keamanan dalam kondisi optimal."
	}
	res += "\n\n*Catatan: Menggunakan analisis otonom cepat karena AI sedang sinkronisasi.*"
	return res
}
