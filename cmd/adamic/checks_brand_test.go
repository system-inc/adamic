package main

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Each brand mint has its own line in the checked report, and a count (rulings.md:29, #yth65gz). An
// overload's mint says the overload whose result makes it (#z1qkw7d). Mutants: the mint lines not printed;
// the lines and the count go missing. The overload's line printed as a cast's; its line goes missing.
func TestExplainChecksListsBrandMints(t *testing.T) {
	t.Parallel()
	var output strings.Builder
	explainPredicateChecks(&output, &ir.Program{BrandMints: []ir.BrandMint{
		{Where: "/b.ts:3:12", Target: "SortedArray<number>", Members: []string{" __sortedArrayBrand"}},
		{Where: "/a.ts:9:4", Target: "SortedReadonlyArray<string>", Members: []string{" __sortedArrayBrand"}},
		{Where: "/c.ts:2:1", Target: "Sorted<number>", Members: []string{"__sorted"}, Overload: "overload 1 of sorted"},
	}})
	text := output.String()
	for _, want := range []string{
		`a.ts:9:4: brand mint SortedReadonlyArray<string> adds  __sortedArrayBrand: erased, unchecked`,
		`b.ts:3:12: brand mint SortedArray<number> adds  __sortedArrayBrand: erased, unchecked`,
		`c.ts:2:1: brand mint Sorted<number> adds __sorted as the result of overload 1 of sorted: erased, unchecked`,
		"adamic: brand mints: 3\n",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("report lacks %q:\n%s", want, text)
		}
	}
	if strings.Index(text, "a.ts:9:4") > strings.Index(text, "b.ts:3:12") {
		t.Errorf("mints aren't in site order:\n%s", text)
	}
}
