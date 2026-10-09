package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

const genericValuesDirectory = "stage3/fixtures/generic-values/"

func init() {
	for _, name := range []string{"01_comparer", "02_index_default", "03_utility_default"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{genericValuesDirectory + name + ".a", true, false})
	}
}

func genericValueOracle(t *testing.T, name string, mutation string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, genericValuesDirectory, name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	for i := range program.Functions {
		function := &program.Functions[i]
		if mutation == "result" && strings.HasPrefix(function.Name, "equateValues_") && !function.Closure {
			function.Body = []ir.Statement{ir.Return{Value: ir.BooleanConstant{Value: false}}}
			changed++
		}
		if mutation == "identity" && function.SourceIdentity != 0 && function.Closure {
			function.SourceIdentity = i + 1000
			changed++
		}
	}
	if mutation != "" && changed == 0 {
		t.Fatal("mutant did not change any instance")
	}
	actual, binary := nativelyUncached(t, program)
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	for name, got := range map[string]run{"native": actual, "javascript": onJavaScriptBackend(t, program)} {
		difference := disagreement(expected, got)
		if mutation == "" && difference != "" {
			t.Fatalf("%s: %s; expected %+v got %+v", name, difference, expected, got)
		}
		if mutation != "" && (difference != "stdout differs" || got.exitCode != 0 || len(got.stderr) != 0) {
			t.Fatalf("%s mutant: %s; %+v", name, difference, got)
		}
	}
}

func TestGenericValueComparer(t *testing.T) { t.Parallel(); genericValueOracle(t, "01_comparer", "") }
func TestGenericValueIndexDefault(t *testing.T) {
	t.Parallel()
	genericValueOracle(t, "02_index_default", "")
}
func TestGenericValueUtilityDefault(t *testing.T) {
	t.Parallel()
	genericValueOracle(t, "03_utility_default", "")
}
func TestGenericValueWrongResult(t *testing.T) {
	t.Parallel()
	genericValueOracle(t, "01_comparer", "result")
}
func TestGenericValueWrongIdentity(t *testing.T) {
	t.Parallel()
	genericValueOracle(t, "01_comparer", "identity")
}

func TestGenericValueIndexWrongResult(t *testing.T) {
	t.Parallel()
	genericValueOracle(t, "02_index_default", "result")
}
func TestGenericValueUtilityWrongResult(t *testing.T) {
	t.Parallel()
	genericValueOracle(t, "03_utility_default", "result")
}
