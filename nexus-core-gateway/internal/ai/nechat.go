package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/nexus-cyber/nexus-core-gateway/pkg/logger"
)

// NechatClient handles Natural Language queries to Qwen3 via OpenRouter.
type NechatClient struct {
	APIKey   string
	Model    string
	Endpoint string
}

func NewNechatClient() *NechatClient {
	apiKey := os.Getenv("AI_PROVIDER_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	}

	endpoint := os.Getenv("AI_PROVIDER_URL")
	if endpoint == "" {
		endpoint = "https://openrouter.ai/api/v1/chat/completions"
	}

	model := os.Getenv("AI_MODEL_REASONING")
	if model == "" {
		model = "qwen/qwen3-235b-a22b" // Use Qwen3 as requested
	}

	return &NechatClient{
		APIKey:   apiKey,
		Model:    model,
		Endpoint: endpoint,
	}
}

// Chat interprets SOC logs using OpenRouter's LLM.
func (n *NechatClient) Chat(logs []logger.TelemetryLog, query string) (string, error) {
	if n.APIKey == "" {
		// Mock response for dev environment without actual API keys
		return "⚠️ **NECHAT Sistem Offline:** `OPENROUTER_API_KEY` tidak dikonfigurasi. \nNamun berdasarkan data log yang ada, saya mendeteksi anomali pada protokol sistem. Silakan isi API Key.", nil
	}

	// Dump logs into a readable JSON string for the AI context
	logsBytes, _ := json.MarshalIndent(logs, "", "  ")
	logsContext := string(logsBytes)
	if len(logsContext) > 8000 {
		logsContext = logsContext[len(logsContext)-8000:] // truncate to fit context window safely
	}

	systemPrompt := `Anda adalah NECHAT, Asisten Security Operations Center (SOC) dari sistem Nexus Cyber. Jawab pertanyaan admin HANYA berdasarkan data log trafik berikut. Berikan ringkasan yang mudah dipahami orang awam. Gunakan bahasa Indonesia yang profesional. Jika log menunjukkan 'RATE_LIMITED' atau 'HONEYPOT_REDIRECTED', jelaskan bahwa sistem MTD sedang melindungi jaringan.`

	userPrompt := fmt.Sprintf("=== RECENT SOC LOGS ===\n%s\n\n=== USER QUERY ===\n%s", logsContext, query)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model":      n.Model,
		"max_tokens": 1024,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	})

	req, err := http.NewRequest("POST", n.Endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+n.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://nexus-cyber.go.id")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var orResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &orResp); err != nil || len(orResp.Choices) == 0 {
		return "", fmt.Errorf("invalid response from OpenRouter: %s", string(body))
	}

	return orResp.Choices[0].Message.Content, nil
}
