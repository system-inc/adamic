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
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/library_array_tail_reduce_right.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/library_array_tail_sort.a", true, false})
}

func TestArrayTailSortAgreesWithNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_array_tail_sort.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
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
		if difference := disagreement(truth, result); difference != "" {
			t.Fatalf("%s: %s: %+v", backend, difference, result)
		}
	}
}

func TestArrayTailSortNodeOnlyMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_array_tail_sort.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name != "array_default_compare" {
			continue
		}
		mutateStringExpressions(reflect.ValueOf(&function.Body).Elem(), func(value ir.Expression) ir.Expression {
			comparison, ok := value.(ir.Binary)
			if !ok {
				return value
			}
			if comparison.Operator == ir.Less {
				comparison.Operator = ir.Greater
				changed++
			} else if comparison.Operator == ir.Greater {
				comparison.Operator = ir.Less
				changed++
			}
			return comparison
		})
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
		if result.exitCode != 0 || len(result.stderr) != 0 || disagreement(truth, result) != "stdout differs" {
			t.Fatalf("%s: mutant must finish cleanly and differ only from Node stdout: %+v", backend, result)
		}
		t.Logf("%s: clean reversed comparator caught only by Node stdout", backend)
	}
}

func TestArrayTailReduceRightAgreesWithNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_array_tail_reduce_right.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
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
		if difference := disagreement(truth, result); difference != "" {
			t.Fatalf("%s: %s: %+v", backend, difference, result)
		}
	}
}

func TestArrayTailReduceRightNodeOnlyMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_array_tail_reduce_right.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name != "array_reduce_right" {
			continue
		}
		mutateStringExpressions(reflect.ValueOf(&function.Body).Elem(), func(value ir.Expression) ir.Expression {
			comparison, ok := value.(ir.Binary)
			if !ok {
				return value
			}
			if comparison.Operator == ir.GreaterOrEqual {
				comparison.Operator = ir.Greater
				changed++
			}
			return comparison
		})
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
		if result.exitCode != 0 || len(result.stderr) != 0 || disagreement(truth, result) != "stdout differs" {
			t.Fatalf("%s: mutant must finish cleanly and differ only from Node stdout: %+v", backend, result)
		}
		t.Logf("%s: clean skipped-zero reduction caught only by Node stdout", backend)
	}
}
