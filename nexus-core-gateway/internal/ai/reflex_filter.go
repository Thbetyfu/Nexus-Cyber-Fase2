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
		regexp.MustCompile(`(?i)(UNION|SELECT|INSERT|UPDATE|DELETE|DROP|ALTER).*FROM`),
		regexp.MustCompile(`(?i)' OR '.*'='`),
		regexp.MustCompile(`(?i)--`),
		regexp.MustCompile(`(?i);`),
	}

	xssRegex := []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script.*?>.*?</script.*?>`),
		regexp.MustCompile(`(?i)onclick=`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)alert\(`),
	}

	traversalRegex := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\.{2,}[/\\]`), // Catch ../, ..\, .../
		regexp.MustCompile(`(?i)%2e%2e%2f`),   // URL Encoded ../
		regexp.MustCompile(`(?i)/etc/`),       // System config access
		regexp.MustCompile(`(?i)C:\\`),        // Windows system access
		regexp.MustCompile(`(?i)\x00`),        // Null byte injection
	}

	return &ReflexFilter{
		sqliPatterns:      sqliRegex,
		xssPatterns:       xssRegex,
		traversalPatterns: traversalRegex,
	}
}

func (f *ReflexFilter) InspectRequest(data string) (isThreat bool, threatType string) {
	for _, p := range f.sqliPatterns {
		if p.MatchString(data) {
			return true, "SQL_INJECTION_DETECTED"
		}
	}

	for _, p := range f.traversalPatterns {
		if p.MatchString(data) {
			return true, "PATH_TRAVERSAL_DETECTED"
		}
	}

	for _, p := range f.xssPatterns {
		if p.MatchString(data) {
			return true, "XSS_DETECTED"
		}
	}

	return false, ""
}

// Sanitize cleans up basic characters (Optional logic for next phase).
func (f *ReflexFilter) Sanitize(data string) string {
	return strings.ReplaceAll(data, "<", "&lt;")
}
