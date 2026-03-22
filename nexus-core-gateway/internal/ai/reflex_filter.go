package ai

import (
	"regexp"
	"strings"
)

// ReflexFilter implements @skill-dual-brain Reflex Layer (Phase 1).
// Fokus: Filtrasi heuristik cepat untuk SQLi dan XSS dasar.
type ReflexFilter struct {
	sqliPatterns []*regexp.Regexp
	xssPatterns  []*regexp.Regexp
}

func NewReflexFilter() *ReflexFilter {
	// Heuristics for common SQLi patterns
	sqliRegex := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(UNION|SELECT|INSERT|UPDATE|DELETE|DROP|ALTER).*FROM`),
		regexp.MustCompile(`(?i)' OR '.*'='`),
		regexp.MustCompile(`(?i)--`),
		regexp.MustCompile(`(?i);`),
	}

	// Heuristics for common XSS patterns
	xssRegex := []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script.*?>.*?</script.*?>`),
		regexp.MustCompile(`(?i)onclick=`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)alert\(`),
	}

	// Heuristics for System Override / Defacement attempts
	overrideRegex := []*regexp.Regexp{
		regexp.MustCompile(`(?i)"command"\s*:\s*"deface"`),
		regexp.MustCompile(`(?i)system/override`),
	}

	return &ReflexFilter{
		sqliPatterns: append(sqliRegex, overrideRegex...),
		xssPatterns:  xssRegex,
	}
}

// InspectRequest inspects query parameters and body for anomalies.
// Target: Latensi < 10ms (Reflex Layer Baseline).
func (f *ReflexFilter) InspectRequest(data string) (isThreat bool, threatType string) {
	// Check for SQLi
	for _, p := range f.sqliPatterns {
		if p.MatchString(data) {
			return true, "SQL_INJECTION_DETECTED"
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

// Sanitize cleans up basic characters (Optional logic for next phase).
func (f *ReflexFilter) Sanitize(data string) string {
	return strings.ReplaceAll(data, "<", "&lt;")
}
