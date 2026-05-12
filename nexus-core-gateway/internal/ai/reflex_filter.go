package ai

import (
	"regexp"
	"strings"
)

// ReflexFilter implements @skill-dual-brain Reflex Layer (Phase 1).
// Fokus: Filtrasi heuristik cepat untuk SQLi dan XSS dasar.
type ReflexFilter struct {
	sqliPatterns      []*regexp.Regexp
	xssPatterns       []*regexp.Regexp
	traversalPatterns []*regexp.Regexp
}

func NewReflexFilter() *ReflexFilter {
	sqliRegex := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(UNION|SELECT|INSERT|UPDATE|DELETE|DROP|ALTER|CREATE|TRUNCATE).*`),
		regexp.MustCompile(`(?i)' OR '.*'='`),
		regexp.MustCompile(`(?i)" OR ".*"="`),
		regexp.MustCompile(`(?i)--`),
		regexp.MustCompile(`(?i);`),
		regexp.MustCompile(`(?i)0x[0-9a-fA-F]+`),
		regexp.MustCompile(`(?i)SLEEP\s*\(`),
	}

	xssRegex := []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script.*?>.*?</script.*?>`),
		regexp.MustCompile(`(?i)on\w+\s*=\s*".*?"`),
		regexp.MustCompile(`(?i)on\w+\s*=\s*'.*?'`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)alert\s*\(`),
		regexp.MustCompile(`(?i)document\.cookie`),
	}

	traversalRegex := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\.{2,}[/\\]`), 
		regexp.MustCompile(`(?i)%2e%2e%2f`),   
		regexp.MustCompile(`(?i)/etc/passwd`), 
		regexp.MustCompile(`(?i)/etc/shadow`), 
		regexp.MustCompile(`(?i)C:\\`),        
		regexp.MustCompile(`(?i)win\.ini`),     
	}

	return &ReflexFilter{
		sqliPatterns:      sqliRegex,
		xssPatterns:       xssRegex,
		traversalPatterns: traversalRegex,
	}
}

// InspectRequest checks the payload for malicious patterns.
func (f *ReflexFilter) InspectRequest(data string) (isThreat bool, threatType string) {
	data = strings.ToLower(data)

	// Check for SQLi
	for _, p := range f.sqliPatterns {
		if p.MatchString(data) {
			return true, "SQL_INJECTION_DETECTED"
		}
	}

	// Check for Path Traversal
	for _, p := range f.traversalPatterns {
		if p.MatchString(data) {
			return true, "PATH_TRAVERSAL_DETECTED"
		}
	}

	// Check for XSS
	for _, p := range f.xssPatterns {
		if p.MatchString(data) {
			return true, "XSS_DETECTED"
		}
	}

	return false, ""
}

// InspectAdvanced checks both the payload and the User-Agent for scanners/tools.
func (f *ReflexFilter) InspectAdvanced(data string, ua string) (isThreat bool, threatType string) {
	// 1. Tool/Scanner Detection
	ua = strings.ToLower(ua)
	scanners := []string{"sqlmap", "gobuster", "dirb", "nmap", "nikto", "burp", "zap", "acunetix"}
	for _, s := range scanners {
		if strings.Contains(ua, s) {
			return true, "MALICIOUS_SCANNER_TOOL_DETECTED"
		}
	}

	// 2. Standard Payload Check
	return f.InspectRequest(data)
}

// Sanitize cleans up basic characters (Optional logic for next phase).
func (f *ReflexFilter) Sanitize(data string) string {
	return strings.ReplaceAll(data, "<", "&lt;")
}
