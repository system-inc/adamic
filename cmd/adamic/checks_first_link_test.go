package main

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// The checked report of a program lowered in first-link trap mode says so with its count of traps, and a
// program lowered any other way has no such line. Mutant: the line not printed; the trapped program's report
// goes without it.
func TestExplainChecksNamesFirstLinkTraps(t *testing.T) {
	t.Parallel()
	var trapped strings.Builder
	explainPredicateChecks(&trapped, &ir.Program{Traps: []ir.Trap{{Kind: "NotYet", Phase: "Body"}, {Kind: "Refused", Phase: "Refusal"}}})
	if !strings.HasPrefix(trapped.String(), "adamic: first-link traps: 2\n") {
		t.Errorf("a trap-mode program's report doesn't start with its traps:\n%s", trapped.String())
	}
	var ordinary strings.Builder
	explainPredicateChecks(&ordinary, &ir.Program{})
	if strings.Contains(ordinary.String(), "first-link") {
		t.Errorf("an ordinary program's report names first-link traps:\n%s", ordinary.String())
	}
}
