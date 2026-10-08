package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Keep original valid programs visible while their shared admission blockers remain.
// These source controls and named compile refusals do not certify runtime pairs.
func TestCheckedViewBrandsOriginalBlockedPairs(t *testing.T) {
	declarations := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if declarations == "" {
		t.Skip("set ADAMIC_BRAND_ORIGINAL_DECLS to complete pinned declarations")
	}
	for _, pair := range []struct{ typ, family, field string }{{"LeftHandSideExpression & Identifier", "recursive intersection payload", "child"}} {
		for _, variant := range []string{"good", "internal", "undefined", "wrong", "null", "missing"} {
			t.Run(pair.typ+"/"+variant, func(t *testing.T) {
				input, err := os.ReadFile("../../stage3/interface-downcasts/lane4/original/" + strings.ToLower(pair.typ) + "-" + variant + ".a")
				if err != nil {
					t.Fatal(err)
				}
				bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
				bound = strings.Replace(bound, "'original-tsc-private'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "brand-private.d.ts"))), 1)
				file := filepath.Join(t.TempDir(), "original-gap.a")
				if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
					t.Fatal(err)
				}
				text := map[string]string{"good": "word-built", "internal": "__importAttributes", "undefined": "undefined", "wrong": "42", "null": "null", "missing": "undefined"}[variant] + "\n"
				if diff := disagreement(run{stdout: []byte(text)}, onNode(t, file)); diff != "" {
					t.Fatal("Node: " + diff)
				}
				_, err = lowered(t, file)
				suffix := "Adamic 0.1 refuses checked view read of field " + pair.field + " with unsupported " + pair.family + " contract; prove or implement the " + pair.family + " contract before reading this field"
				if err == nil || !strings.HasSuffix(err.Error(), suffix) {
					t.Fatalf("expected named original admission gap, got %v", err)
				}
				t.Log("original pair remains uncredited: " + pair.family)
			})
		}
	}
}
