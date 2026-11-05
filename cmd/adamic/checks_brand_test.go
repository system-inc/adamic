package main

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Each brand mint has its own line in the checked report, and a count (rulings.md:29, #yth65gz).
// Mutant: the mint lines not printed; the line and the count go missing.
func TestExplainChecksListsBrandMints(t *testing.T) {
	var output strings.Builder
	explainPredicateChecks(&output, &ir.Program{BrandMints: []ir.BrandMint{
		{Where: "/b.ts:3:12", Target: "SortedArray<number>", Members: []string{" __sortedArrayBrand"}},
		{Where: "/a.ts:9:4", Target: "SortedReadonlyArray<string>", Members: []string{" __sortedArrayBrand"}},
	}})
	text := output.String()
	for _, want := range []string{
		`a.ts:9:4: brand mint SortedReadonlyArray<string> adds  __sortedArrayBrand: erased, unchecked`,
		`b.ts:3:12: brand mint SortedArray<number> adds  __sortedArrayBrand: erased, unchecked`,
		"adamic: brand mints: 2\n",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("report lacks %q:\n%s", want, text)
		}
	}
	if strings.Index(text, "a.ts:9:4") > strings.Index(text, "b.ts:3:12") {
		t.Errorf("mints aren't in site order:\n%s", text)
	}
}
