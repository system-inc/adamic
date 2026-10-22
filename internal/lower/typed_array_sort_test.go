package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func typedArraySortComparatorRefusal(t *testing.T, kind string) {
	t.Helper()
	for _, comparator := range []string{"(a, b) => a - b", "compare", "undefined"} {
		source := "function compare(a: number, b: number): number { return a - b; }\nconst values = new " + kind + "([10, 2]);\nconsole.log(`${values.sort(" + comparator + ")[0]},${values[1]}`);\n"
		path := filepath.Join(t.TempDir(), "comparator.a")
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		node := runAgreementNode(t, path)
		if node.code != 0 || string(node.stdout) != "2,10\n" || len(node.stderr) != 0 {
			t.Fatalf("Node: %+v", node)
		}
		_, err := lowerSource(t, source)
		var gap *NotYet
		if !errors.As(err, &gap) || !strings.Contains(err.Error(), "typed array sort with a comparator") {
			t.Errorf("%s %s: expected named comparator gap, got %v", kind, comparator, err)
		}
	}
}

func TestTypedArraySortComparatorUint8(t *testing.T) {
	t.Parallel()
	typedArraySortComparatorRefusal(t, "Uint8Array")
}
func TestTypedArraySortComparatorUint16(t *testing.T) {
	t.Parallel()
	typedArraySortComparatorRefusal(t, "Uint16Array")
}
func TestTypedArraySortComparatorInt32(t *testing.T) {
	t.Parallel()
	typedArraySortComparatorRefusal(t, "Int32Array")
}
func TestTypedArraySortComparatorFloat64(t *testing.T) {
	t.Parallel()
	typedArraySortComparatorRefusal(t, "Float64Array")
}
