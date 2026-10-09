package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

const genericValuesDirectory = "stage3/fixtures/generic-values/"

func init() {
	for _, name := range []string{"01_comparer", "02_index_default", "03_utility_default", "04_alias"} {
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

func TestGenericValueAlias(t *testing.T) { t.Parallel(); genericValueOracle(t, "04_alias", "") }

// Plant an escape after the direct calls: no single type may be guessed from
// the first call. The slot remains polymorphic and must be refused.
func TestGenericValueAliasEscapeMutant(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "escaped.a")
	source := `function identity<T>(value: T): T { return value; }
const generic = identity;
console.log(generic(4) + " " + generic("x"));
const escaped: { readonly run: <U>(value: U) => U } = {run: generic};
console.log(escaped.run(4) + " " + escaped.run("x"));`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	if expected.exitCode != 0 || string(expected.stdout) != "4 x\n4 x\n" || len(expected.stderr) != 0 {
		t.Fatalf("Node control: %+v", expected)
	}
	_, err := lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.Error(), "generic-function-values") {
		t.Fatalf("escape was not checked/refused: %v", err)
	}
	t.Logf("planted polymorphic escape refused: %v", err)
}

func TestGenericValueAliasEarlyRead(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "early.a")
	source := `function identity<T>(value: T): T { return value; }
try { run(); } catch { console.log("early"); }
const generic = identity;
function run(): void { console.log("" + generic(4)); }
run();`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	actual, binary := nativelyUncached(t, program)
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	for name, got := range map[string]run{"native": actual, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(expected, got); difference != "" {
			t.Errorf("%s TDZ: %s; expected %+v got %+v", name, difference, expected, got)
		}
	}
}

func TestGenericValueHigherRankNodeControl(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, genericValuesDirectory, "05_higher_rank.a"))
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	if expected.exitCode != 0 || string(expected.stdout) != "4 x\n" || len(expected.stderr) != 0 {
		t.Fatalf("Node control: %+v", expected)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "box.run (U, V)") {
		t.Fatalf("higher-rank program escaped its ruled refusal: %v", err)
	}
}

func TestGenericValueDeclaredDefaultWrongResult(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "default.a")
	source := `function constant<T = number>(): number { return 4; }
const run: () => number = constant;
console.log("" + run());`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
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
		if strings.HasPrefix(function.Name, "constant_") && !function.Closure {
			function.Body = []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: 17}}}
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("default mutant changed %d instances", changed)
	}
	actual, binary := nativelyUncached(t, program)
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	for name, got := range map[string]run{"native": actual, "javascript": onJavaScriptBackend(t, program)} {
		if disagreement(expected, got) != "stdout differs" || got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("%s default mutant escaped Node: %+v", name, got)
		}
	}
}
