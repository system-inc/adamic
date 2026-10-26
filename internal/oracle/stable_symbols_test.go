package oracle

import (
	"regexp"
	"testing"
)

// Proof mutations locate printed declarations by readable spelling, never IR IDs.
// Require one symbol so a renamed or ambiguous fixture cannot silently lose its control.
func emittedLocalSymbol(t *testing.T, code, prefix, name string) string {
	t.Helper()
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(prefix) + `[A-Za-z0-9_]+_` + regexp.QuoteMeta(name) + `\b`)
	unique := map[string]bool{}
	for _, match := range pattern.FindAllString(code, -1) {
		unique[match] = true
	}
	if len(unique) != 1 {
		t.Fatalf("want one emitted %s%s declaration, found %d", prefix, name, len(unique))
	}
	for match := range unique {
		return match
	}
	panic("unreachable")
}

func emittedFunctionSymbol(t *testing.T, code, declaration string) string {
	t.Helper()
	pattern := regexp.MustCompile(`\badamic_function_(?:[A-Za-z0-9_]*_)?` + regexp.QuoteMeta(declaration) + `_[a-f0-9]{32}\b`)
	unique := map[string]bool{}
	for _, match := range pattern.FindAllString(code, -1) {
		unique[match] = true
	}
	if len(unique) != 1 {
		t.Fatalf("want one emitted function %s, found %d", declaration, len(unique))
	}
	for match := range unique {
		return match
	}
	panic("unreachable")
}
