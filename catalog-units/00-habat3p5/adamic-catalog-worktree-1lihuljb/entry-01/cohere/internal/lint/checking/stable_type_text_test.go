package type_checking

import "testing"

// TestStablePropertyNamePrintsASymbolKeyByItsNameAlone: a member keyed by a symbol prints `[name]` without the
// symbol's creation id, from the raw 0xFE byte typescript-go names it with (#z9jcxp1). The pattern this replaced
// matched the rune U+00FE and printed `"\xfe@brand@1532"`, with an id that moved between engines.
func TestStablePropertyNamePrintsASymbolKeyByItsNameAlone(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]string{
		"\xfe@BaseRequestBrand@1532": "[BaseRequestBrand]",
		"\xfe@BaseRequestBrand@1525": "[BaseRequestBrand]",
		"\xfe@iterator":              "[iterator]",
		"þ@brand@12":                 `"þ@brand@12"`,
		"label":                      "label",
		"data-label":                 `"data-label"`,
	} {
		if got := StablePropertyName(name); got != want {
			t.Errorf("StablePropertyName(%q) = %q, want %q", name, got, want)
		}
	}
}
