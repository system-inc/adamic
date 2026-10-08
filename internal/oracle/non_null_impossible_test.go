package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// Historical TypeScript literal assertions carry inserted checks.
func init() {
	for _, operand := range []string{"undefined", "null"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/non_null_refuse_" + operand + ".ts", true, true})
	}
}

func TestLiteralNonNullTypeScriptChecks(t *testing.T) {
	t.Parallel()
	paths := []string{"non_null_refuse_undefined.ts", "non_null_refuse_null.ts"}
	for _, name := range paths {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if filepath.Ext(path) == ".ts" {
				if err != nil {
					t.Fatal(err)
				}
				if program.NonNullChecks.Proven != 0 || program.NonNullChecks.Checked != 1 {
					t.Fatalf("want proven 0 checked 1, got %#v", program.NonNullChecks)
				}
				expression := "undefined!"
				if strings.Contains(name, "null.ts") {
					expression = "null!"
				}
				if difference := disagreement(run{stdout: []byte(strings.TrimSuffix(expression, "!") + "\n")}, onNode(t, path)); difference != "" {
					t.Fatal("source Node: " + difference)
				}
				want := run{stderr: []byte("adamic: panic: non-null assertion failed at " + path + ":1:13: " + expression + " is null or undefined\n"), exitCode: 70}
				native, _ := nativelyUncached(t, program)
				for _, got := range []run{native, onJavaScriptBackend(t, program)} {
					if difference := disagreement(want, got); difference != "" {
						t.Fatalf("%s: %#v", difference, got)
					}
				}
				return
			}
		})
	}
}
