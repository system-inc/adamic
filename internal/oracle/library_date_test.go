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
	for _, name := range []string{"epoch", "utc", "getters", "format", "parse", "clock", "own", "invalid_iso"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/library_date_" + name + ".a", true, false})
	}
}

func TestDateNodeOnlyMutants(t *testing.T) {
	for _, probe := range []struct{ fixture, operation, replacement string }{
		{"epoch", "date_time", "date_getUTCSeconds"},
		{"utc", "date_UTC", ""},
		{"getters", "date_getUTCMonth", "date_getUTCDate"},
		{"format", "date_toISOString", "date_toUTCString"},
		{"parse", "date_parse", ""},
		{"clock", "date_now", ""},
		{"own", "date_constructor_own", "date_prototype_own"},
		{"invalid_iso", "date_toISOString", "date_toUTCString"},
	} {
		t.Run(probe.fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_date_"+probe.fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := 0
			mutate := func(value ir.Expression) ir.Expression {
				call, ok := value.(ir.NodeFSFile)
				if !ok || call.Operation != probe.operation {
					return value
				}
				changed++
				if probe.replacement != "" {
					call.Operation = probe.replacement
					return call
				}
				return ir.Binary{Operator: ir.Add, Left: call, Right: ir.NumberConstant{Value: 0.25}}
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
			for index := range program.Functions {
				mutateStringExpressions(reflect.ValueOf(&program.Functions[index].Body).Elem(), mutate)
			}
			if changed == 0 {
				t.Fatal("no mutation")
			}
			truth := onNode(t, path)
			actual, binary := natively(t, program)
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			results := map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				results["WASI"] = onWASI(t, native.C(program))
			}
			for backend, result := range results {
				if result.exitCode != 0 || len(result.stderr) != 0 || disagreement(truth, result) == "" {
					t.Fatalf("%s: mutant must finish cleanly and differ only in the Node comparison: %+v", backend, result)
				}
				t.Logf("%s: clean mutant caught only by Node (%s)", backend, disagreement(truth, result))
			}
		})
	}
}

func TestDateWASIAgreesWithNode(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	for _, name := range []string{"epoch", "utc", "getters", "format", "parse", "clock", "own", "invalid_iso"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_date_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected, actual := onNode(t, path), onWASI(t, native.C(program))
			if difference := disagreement(expected, actual); difference != "" {
				t.Fatalf("WASI: %s: %+v", difference, actual)
			}
		})
	}
}

func TestDateFormsIgnoreHostTimezone(t *testing.T) {
	t.Setenv("TZ", "Asia/Kathmandu")
	for _, name := range []string{"format", "parse", "getters"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_date_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
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
			for backend, result := range results {
				if diff := disagreement(expected, result); diff != "" {
					t.Fatalf("%s under TZ=Asia/Kathmandu: %s", backend, diff)
				}
			}
		})
	}
}
