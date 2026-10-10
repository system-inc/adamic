package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// contractionExplains says whether a program's disagreement with Node is only that Node's V8 was
// compiled with multiply-adds contracted, as Node's on macOS arm64 is (internal/native/fused_test.go):
// the same program, its runtime built that way and its own arithmetic not, must then print exactly
// what Node did. Adamic itself never fuses, so it prints one answer everywhere. On a machine that
// can't fuse the build is the unfused one, and nothing is explained.
func contractionExplains(t *testing.T, program *ir.Program, oracle run) bool {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "fused")
	if err := native.Build(native.C(program), binary, native.Options{FusedRuntime: true}); err != nil {
		t.Fatal(err)
	}
	if disagreement(oracle, execute(t, binary)) != "" {
		return false
	}
	t.Logf("differs from this Node only as its V8's contracted multiply-adds do")
	return true
}
