package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Local bound lengths use runtime metadata, while Function-typed graph views
// retain their independent dynamic-property boundary. The typed arrow prevents
// replacing runtime metadata with its wider declared signature's arity.
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

// Removing the local length stop exposes the separate unknown-function boundary
// in this helper. Class method entries retain the metadata stop.
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
			what := "a function viewed as unknown or object (dynamic function descriptors)"
			if probe.file == "keyed_length_method.a" {
				what = "a bound function's length (bound length metadata)"
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), what) {
				t.Fatalf("want %s, got %v", what, err)
			}
			compiler := buildcache.AdamicBuiltElsewhere(t)
			for _, backend := range []string{"c", "js"} {
				stopped := execute(t, compiler, backend, path)
				if stopped.exitCode == 0 || !strings.Contains(string(stopped.stderr), what) {
					t.Fatalf("%s must stop before emission: %+v", backend, stopped)
				}
			}
		})
	}
}
