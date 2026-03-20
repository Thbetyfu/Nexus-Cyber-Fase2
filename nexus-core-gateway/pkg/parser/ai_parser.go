package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AIParseError is a typed error for failed AI response parsing.
type AIParseError struct {
	RawResponse string
}

func (e *AIParseError) Error() string {
	preview := e.RawResponse
	if len(preview) > 200 {
		preview = preview[:200]
	}
	return fmt.Sprintf("ai_parse_error: cannot parse AI response: %s", preview)
}

// ParseAIJSON implements a 3-stage robust JSON parsing strategy.
// This handles AI outputs that may contain narration outside JSON.
// Stage 1: Direct parse
// Stage 2: Extract from ```json ... ``` code block
// Stage 3: Bracket search (first { to last })
func ParseAIJSON(raw string, target interface{}) error {
	raw = strings.TrimSpace(raw)

	// Stage 1: Direct parse
	if err := json.Unmarshal([]byte(raw), target); err == nil {
		return nil
	}

	// Stage 2: JSON codeblock extraction
	if idx := strings.Index(raw, "```json"); idx != -1 {
		end := strings.Index(raw[idx+7:], "```")
		if end != -1 {
			block := strings.TrimSpace(raw[idx+7 : idx+7+end])
			if err := json.Unmarshal([]byte(block), target); err == nil {
				return nil
			}
		}
	}

	// Stage 3: Bracket search — most tolerant
	start := strings.Index(raw, "{")
	last := strings.LastIndex(raw, "}")
	if start != -1 && last != -1 && last > start {
		if err := json.Unmarshal([]byte(raw[start:last+1]), target); err == nil {
			return nil
		}
	}

	return &AIParseError{RawResponse: raw}
}
