package lower

import (
	"strings"
	"testing"
)

// Throwaway measurement for #5qx6qv4, never committed: each source's lowering outcome, printed.
func TestSourceSlotPendingProbe(t *testing.T) {
	plain := sourceSlotPendingPredicate
	mapped := strings.Replace(plain, "symbol?: Sym | undefined }", "symbol?: Sym | undefined; locals?: Map<string, Sym> | undefined }", 1)
	generic := func(source string) string {
		return strings.Replace(source, "function setValueDeclaration(symbol: Sym, node: Decl)", "function setValueDeclaration<D extends Decl>(symbol: Sym, node: D)", 1)
	}
	for _, row := range []struct{ name, source string }{
		{"163 plain declared", plain},
		{"163 plain generic", generic(plain)},
		{"163 map declared", mapped},
		{"163 map generic", generic(mapped)},
	} {
		_, err := lowerSourceAs(t, row.source, ".ts")
		outcome := "lowers"
		if err != nil {
			outcome = err.Error()
			if at := strings.Index(outcome, "main.ts:"); at >= 0 {
				outcome = outcome[at:]
			}
		}
		t.Logf("PROBE %s: %s", row.name, outcome)
	}
}
