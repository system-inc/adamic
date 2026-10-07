package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Register here so the relation unit does not change the oracle's central harness.
func init() {
	for _, name := range []string{"class", "fresh", "declared", "subclass", "generic_subclass", "boolean_as_number", "boolean_as_number_array", "boolean_as_number_ops", "boolean_as_number_parameter", "boolean_as_number_return", "false_as_number"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "internal/oracle/testdata/optional_widening_" + name + ".a", lowers: true})
	}
}

// Pin the independent observations of the programs that Adamic now refuses.
func TestOptionalWideningObservedNode(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob(filepath.Join(repository, "internal/lower/testdata/optional_widening/*.a"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		name := filepath.Base(path)
		if !strings.HasPrefix(name, "widening_") && !(strings.HasPrefix(name, "class_") && name != "class_field.a") {
			continue
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			want := "string\n"
			if strings.HasPrefix(name, "widening_") {
				observed, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".expected")
				if err != nil {
					t.Fatal(err)
				}
				want = string(observed)
			}
			absolute, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			got := onNode(t, absolute)
			if got.exitCode != 0 || len(got.stderr) != 0 || string(got.stdout) != want {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q; want %q", got.exitCode, got.stdout, got.stderr, want)
			}
		})
	}
}
