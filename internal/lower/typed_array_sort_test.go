package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func typedArraySortComparatorRefusal(test *testing.T, kind string) {
	test.Helper()
	for _, comparator := range []string{"(a, b) => a - b", "compare", "undefined"} {
		source := "function compare(a: number, b: number): number { return a - b; }\nconst values = new " +
			kind +
			"([10, 2]);\nconsole.log(`${values.sort(" +
			comparator +
			")[0]},${values[1]}`);\n"
		path := filepath.Join(test.TempDir(), "comparator.a")
		if writeError := os.WriteFile(path, []byte(source), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
		node := runAgreementNode(test, path)
		if node.code != 0 || string(node.stdout) != "2,10\n" || len(node.stderr) != 0 {
			test.Fatalf("Node: %+v", node)
		}
		_, lowerSourceError := lowerSource(test, source)
		var gap *NotYet
		if !errors.As(lowerSourceError, &gap) ||
			!strings.Contains(lowerSourceError.Error(), "typed array sort with a comparator") {
			test.Errorf("%s %s: expected named comparator gap, got %v", kind, comparator, lowerSourceError)
		}
	}
}

func TestTypedArraySortComparatorUint8(test *testing.T) {
	test.Parallel()
	typedArraySortComparatorRefusal(test, "Uint8Array")
}
func TestTypedArraySortComparatorUint16(test *testing.T) {
	test.Parallel()
	typedArraySortComparatorRefusal(test, "Uint16Array")
}
func TestTypedArraySortComparatorInt32(test *testing.T) {
	test.Parallel()
	typedArraySortComparatorRefusal(test, "Int32Array")
}
func TestTypedArraySortComparatorFloat64(test *testing.T) {
	test.Parallel()
	typedArraySortComparatorRefusal(test, "Float64Array")
}
