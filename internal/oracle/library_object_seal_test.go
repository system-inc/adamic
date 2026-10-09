package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"shapes", "collections", "regexp"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/library_object_seal_" + name + ".a", true, false})
	}
}

func TestObjectSealNodeOnlyMutants(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, method, replacement string }{
		{"shapes", "preventExtensions", "seal"},
		{"shapes", "isFrozen", "isExtensible"},
		{"collections", "seal", ""},
		{"regexp", "isSealed", "isFrozen"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_object_seal_"+probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := 0
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
				if call, ok := value.(ir.ObjectCall); ok && call.Method == probe.method && (probe.method != "isFrozen" || call.IntegrityShape == 1) {
					changed++
					if probe.replacement == "" {
						return call.Arguments[0]
					}
					call.Method = probe.replacement
					return call
				}
				return value
			})
			if changed == 0 {
				t.Fatal("no mutation")
			}
			expected := onNode(t, path)
			actual, binary := natively(t, program)
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			results := map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				results["WASI"] = onWASI(t, native.C(program))
			}
			for name, result := range results {
				if result.exitCode != 0 || len(result.stderr) != 0 || disagreement(expected, result) != "stdout differs" {
					t.Fatalf("%s mutant must be caught only by Node stdout: %+v, %s", name, result, disagreement(expected, result))
				}
				t.Log(name + ": clean mutant caught only by Node stdout")
			}
		})
	}
}
