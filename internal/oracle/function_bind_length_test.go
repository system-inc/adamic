package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// A bound closure's length isn't built: the bound function lowering makes it an ir.Function with no
// SourceLength, so native would read 0 where Node gives the target's length less the fixed arguments.
// Every road a bound function takes to a Function-typed .length read stops today at the graph
// boundary, because a bound closure captures a closure and so makes the program a graph program,
// where a dynamic property read is refused. That stop is held for another reason, so each road is
// pinned here: whoever lifts it meets every one of them (#0ty4971). The typed arrow is the row a
// static length would get wrong, since its declared type names two parameters and the arrow has one.
func TestFunctionBindLengthFunctionSlot(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, stdout, line string }{
		{"length_function_slot.ts", "1\n", ":1:"},
		{"length_returned.ts", "1\n", ":4:"},
		{"length_field.ts", "2\n", ":3:"},
		{"length_array.ts", "2\n3\n", ":3:"},
		{"length_method.ts", "2 1\n", ":5:"},
		{"length_typed_arrow.ts", "1 0\n", ":2:"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/bind_refused", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != probe.stdout {
				t.Fatalf("source Node: %+v", truth)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), "a dynamic-shape object in a graph program") || !strings.Contains(err.Error(), probe.file+probe.line) {
				t.Fatalf("want existing graph/dynamic stop at Function.length, got %v", err)
			}
		})
	}
}

// The shared lower stop protects both emitted backends before a keyed read can
// observe the synthetic adapter's length. Node supplies each source answer.
func TestFunctionBindKeyedLength(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, stdout string }{
		{"keyed_length_two.a", "1\n"},
		{"keyed_length_zero.a", "0\n"},
		{"keyed_length_method.a", "2\n"},
		{"keyed_length_default.a", "1\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/bind_refused", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != probe.stdout || len(truth.stderr) != 0 {
				t.Fatalf("source Node: %+v", truth)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), "a bound function's length (bound length metadata)") {
				t.Fatalf("want bound length stop protecting native and JavaScript, got %v", err)
			}
			compiler := buildcache.AdamicBuiltElsewhere(t)
			for _, backend := range []string{"c", "js"} {
				stopped := execute(t, compiler, backend, path)
				if stopped.exitCode == 0 || !strings.Contains(string(stopped.stderr), "a bound function's length (bound length metadata)") {
					t.Fatalf("%s must stop before emission: %+v", backend, stopped)
				}
			}
		})
	}
}
