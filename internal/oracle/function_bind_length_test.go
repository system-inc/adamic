package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestFunctionBindLengthFunctionSlot(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/bind_refused/length_function_slot.ts"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "1\n" {
		t.Fatalf("source Node: %+v", truth)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "a dynamic-shape object in a graph program") || !strings.Contains(err.Error(), "length_function_slot.ts:1:") {
		t.Fatalf("want existing graph/dynamic stop at Function.length, got %v", err)
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
