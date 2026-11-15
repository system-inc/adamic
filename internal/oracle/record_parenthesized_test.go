package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// A parenthesized record read is judged at the access inside its parentheses, never cast as one
// itself: recordLiteralRead's walk once turned (r["k"]) into a lowering panic (#grc95d6).
func TestRecordParenthesizedReads(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"p1_value", "p2_optional_length", "p3_written_template", "p4_property"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/record_parenthesized/"+name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			backend := onJavaScriptBackend(t, program)
			if diff := disagreement(truth, backend); diff != "" {
				t.Fatalf("javascript: %s; Node stdout %q exit %d; backend stdout %q exit %d", diff, truth.stdout, truth.exitCode, backend.stdout, backend.exitCode)
			}
			actual, leaked, comparable := nativelyCheckingLeaks(t, program, true)
			if leaked != "" {
				t.Fatalf("leaks: %s", leaked)
			}
			if comparable {
				if diff := disagreement(truth, actual); diff != "" {
					t.Fatalf("native: %s", diff)
				}
			}
		})
	}
	t.Run("p5_prototype_name", func(t *testing.T) {
		t.Parallel()
		path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/record_parenthesized/p5_prototype_name.ts"))
		if err != nil {
			t.Fatal(err)
		}
		_, err = lowered(t, path)
		if err == nil || !strings.Contains(err.Error(), "names an Object.prototype member") {
			t.Fatalf("want the prototype-name refusal at the access inside the parentheses, got %v", err)
		}
	})
}
