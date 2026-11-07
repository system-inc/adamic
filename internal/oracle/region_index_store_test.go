package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// A statement that hands a fresh function's object to a parameter that flows nowhere gets a region,
// and the object is freed when the statement ends (internal/native/region.go, #22h1bv2). A parameter
// stored into an outer array, by index or by push, outlives the call, so escapes must mark it and the
// object stays on the heap: in a region, the array would point at freed memory by the next statement.
// control.a pins that the region path is live for this shape, so the stores' missing regions are
// escapes' doing and not a shape that never gets one. `return p.x` would mark the escape on its own, a
// Property followed to the parameter, so each keep returns a Binary, which escaping consumes, or
// nothing. Each program agrees with Node on the JavaScript backend and natively under ASan with leak
// detection. The mutant, escapes' SetIndex arm removed, is in testdata/region_index_store/mutants.md.
const regionIndexStoreDirectory = "internal/oracle/testdata/region_index_store"

// regionIndexStoreAgrees runs one program Node's way, the JavaScript backend's and natively sanitized,
// holds all three to the pinned text, and returns the program's C for its region assertion.
func regionIndexStoreAgrees(t *testing.T, name, want string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, regionIndexStoreDirectory, name))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != want || len(truth.stderr) != 0 {
		t.Fatalf("Node: %d %q %s", truth.exitCode, truth.stdout, truth.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatalf("JavaScript backend: %s", difference)
	}
	code := native.C(program)
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if difference := disagreement(truth, result); difference != "" {
		t.Fatalf("native: %s %d %q %s", difference, result.exitCode, result.stdout, result.stderr)
	}
	return code
}

func TestRegionIndexStoreControl(t *testing.T) {
	t.Parallel()
	code := regionIndexStoreAgrees(t, "control.a", "9\n")
	if !strings.Contains(code, "ADAMIC_REGION") {
		t.Fatal("control.a's statement holds no region: the region path isn't live for this shape, so the stores' missing regions prove nothing")
	}
}

func TestRegionIndexStoreKeepsObjectOffRegion(t *testing.T) {
	t.Parallel()
	code := regionIndexStoreAgrees(t, "index_store.a", "9 p7 7\n")
	if strings.Contains(code, "ADAMIC_REGION") {
		t.Fatal("index_store.a's C holds a region: the object stored into outer[0] is freed when its statement ends")
	}
}

// keep returns nothing here, so no return can mark the escape: the SetIndex arm is all there is.
func TestRegionVoidIndexStoreKeepsObjectOffRegion(t *testing.T) {
	t.Parallel()
	code := regionIndexStoreAgrees(t, "void_index_store.a", "p7 7\n")
	if strings.Contains(code, "ADAMIC_REGION") {
		t.Fatal("void_index_store.a's C holds a region: the object stored into outer[0] is freed when its statement ends")
	}
}

// The push isn't the SetIndex arm's to mark. outer.push(p) lowers to an ir.ArrayPush, which isn't
// pure and so isn't a consumer (borrow.go's consumes), and escaping's general rule, the operand check
// at the end of derived, marks the parameter it holds as escaping.
func TestRegionPushStoreKeepsObjectOffRegion(t *testing.T) {
	t.Parallel()
	code := regionIndexStoreAgrees(t, "push_store.a", "9 p7 2\n")
	if strings.Contains(code, "ADAMIC_REGION") {
		t.Fatal("push_store.a's C holds a region: the object pushed onto outer is freed when its statement ends")
	}
}
