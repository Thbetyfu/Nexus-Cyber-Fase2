package utils

import "strings"

// INJECTION_PATTERNS adalah daftar string berbahaya yang bisa menyebabkan
// "Prompt Injection" atau "Jailbreak" pada model AI.
// Reference: OWASP LLM01 - Prompt Injection
var INJECTION_PATTERNS = []string{
	"ignore previous instructions",
	"forget everything above",
	"you are now",
	"system prompt",
	"assistant:",
	"<|im_start|>", "<|im_end|>", // Qwen special tokens
	"<|begin_of_text|>", "<|end_of_text|>", // Llama special tokens
	"[INST]", "[/INST]", // Llama instruction tokens
	"<<SYS>>", "<</SYS>>", // Llama system tags
	"jailbreak", "DAN mode",
	"pretend you are", "act as if",
	"disregard all",
}

// SanitizeTrafficForPrompt is a mandatory shield before any data touches AI models.
// It prevents Prompt Injection (OWASP LLM01) and Jailbreak attempts.
func SanitizeTrafficForPrompt(rawInput string, maxChars int) string {
	sanitized := rawInput

	for _, pattern := range INJECTION_PATTERNS {
		sanitized = strings.ReplaceAll(sanitized, pattern, "[FILTERED]")
		// Also check for case-insensitive variants
		sanitized = strings.ReplaceAll(strings.ToLower(sanitized), strings.ToLower(pattern), "[FILTERED]")
	}

	// Escape curly braces to not break f-string templates
	sanitized = strings.ReplaceAll(sanitized, "{", "{{")
	sanitized = strings.ReplaceAll(sanitized, "}", "}}")

	// Truncate to prevent token budget overflow
	if len(sanitized) > maxChars {
		return sanitized[:maxChars] + "...[TRUNCATED]"
	}

	return sanitized
}
