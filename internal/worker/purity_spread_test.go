package worker

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// A spread runs the getters lowering says it may (ir.ObjectLiteral's SpreadGetters), and an object
// literal's own type is fresh, which fails a narrower view by excess properties. spreadGetters relates the
// literal by its regular type (#z6h9s7k), so compute's spread of a { readonly label: string } may run the
// literal's getter, which writes the module global state, and compute is refused before any Wasm build.
// Related raw, the getter was left out and compute was judged pure (spread_getters.mutants.md in
// internal/lower).
func TestWasmPurityInspectsANarrowSpreadsLiteralGetter(t *testing.T) {
	t.Parallel()
	builds := 0
	err := BuildWith("testdata/wasm/spread_getter_narrow_view.a", filepath.Join(t.TempDir(), "worker"), Options{
		Wasm:      []string{"compute"},
		BuildWasm: func(string, string, string, []string) error { builds++; return fmt.Errorf("unexpected Wasm build") },
	})
	for _, want := range []string{"compute", "accesses module global state", "fix:"} {
		if err == nil || !strings.Contains(err.Error(), want) || builds != 0 {
			t.Fatalf("want compute refused as impure (%s) before the Wasm build, builds=%d: %v", want, builds, err)
		}
	}
}
