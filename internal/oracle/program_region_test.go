package oracle

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func programRegionLowered(t *testing.T, path string, enabled bool) *ir.Program {
	t.Helper()
	source, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.LowerWithOptions(context.Background(), source, lower.Options{ProgramRegion: enabled})
	if err != nil {
		t.Fatal(err)
	}
	return program
}

func checkProgramRegion(t *testing.T, program *ir.Program, path string) []int {
	t.Helper()
	code := native.C(program)
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(code, binary, native.Options{ProgramRegion: true, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	oracle := onNode(t, path)
	if diff := disagreement(oracle, onJavaScriptBackend(t, program)); diff != "" {
		t.Fatalf("JavaScript backend: %s", diff)
	}
	plain := filepath.Join(t.TempDir(), "plain")
	if err := native.Build(code, plain, native.Options{ProgramRegion: true}); err != nil {
		t.Fatal(err)
	}
	if diff := disagreement(oracle, execute(t, plain)); diff != "" {
		t.Fatalf("native plain: %s", diff)
	}
	result := executeWith(t, []string{leakSanitizer()}, binary)
	if difference := disagreement(oracle, result); difference != "" {
		t.Fatalf("Program region: %s: %s", difference, result.stderr)
	}
	report, err := leakcheck.Check(leakcheck.Program{
		C: code, Sanitized: binary, Counted: filepath.Join(t.TempDir(), "counted"),
		BuildCounted: func(code, output string) error {
			return native.Build(code, output, native.Options{ProgramRegion: true, Count: true})
		},
		Execute: func(environment []string, name string, args ...string) leakcheck.Run {
			return leakRun(executeWith(t, environment, name, args...))
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if report != "" {
		t.Fatal(report)
	}
	countedBinary := filepath.Join(t.TempDir(), "counts")
	if err := native.Build(code, countedBinary, native.Options{ProgramRegion: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	name, args := pinnedStack(countedBinary)
	counted := execute(t, name, args...)
	match := countsLine.FindSubmatch(counted.stderr)
	if counted.exitCode != 0 || match == nil {
		t.Fatalf("counts %d %s", counted.exitCode, counted.stderr)
	}
	counts := []int{}
	for _, field := range match[1:] {
		value, _ := strconv.Atoi(string(field))
		counts = append(counts, value)
	}
	if counts[0] != counts[1]+counts[5] {
		t.Fatalf("not every member accounted: %s", counted.stderr)
	}
	t.Logf("%s", counted.stderr)
	return counts
}

func programRegionFixture(t *testing.T, name string) []int {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/program_region", name))
	if err != nil {
		t.Fatal(err)
	}
	return checkProgramRegion(t, programRegionLowered(t, path, true), path)
}
func TestProgramRegionGraphParent(t *testing.T) {
	t.Parallel()
	programRegionFixture(t, "cycles_graph_parent.a")
}
func TestProgramRegionGraphRelations(t *testing.T) {
	t.Parallel()
	programRegionFixture(t, "cycles_graph_relations.a")
}
func TestProgramRegionGraphSymbols(t *testing.T) {
	t.Parallel()
	programRegionFixture(t, "cycles_graph_symbols.a")
}
func TestProgramRegionWeakParent(t *testing.T) {
	t.Parallel()
	programRegionFixture(t, "cycles_weak_parent.a")
}
func TestProgramRegionWeakRelations(t *testing.T) {
	t.Parallel()
	programRegionFixture(t, "cycles_weak_relations.a")
}
func TestProgramRegionWeakSymbols(t *testing.T) {
	t.Parallel()
	programRegionFixture(t, "cycles_weak_symbols.a")
}
func TestProgramRegionMillion(t *testing.T) { t.Parallel(); programRegionFixture(t, "million.a") }
func TestProgramRegionOptionalStorage(t *testing.T) {
	t.Parallel()
	programRegionFixture(t, "optional.a")
}
func TestProgramRegionConstructor(t *testing.T) {
	t.Parallel()
	programRegionFixture(t, "constructor.a")
}
func TestProgramRegionOwnership(t *testing.T) {
	t.Parallel()
	counts := programRegionFixture(t, "program_region_ownership.a")
	if counts[5] != 2 {
		t.Fatalf("membership census: regions %d, want 2", counts[5])
	}
}
func TestProgramRegionCapturedCell(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "capture.a")
	source := `function make(): () => string {let current:(()=>string)|undefined=undefined;const result=()=>current===undefined?'empty':'set';current=result;return result;}console.log(make()());`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	counts := checkProgramRegion(t, programRegionLowered(t, path, true), path)
	if counts[5] < 2 {
		t.Fatalf("closure and cell membership missing: %v", counts)
	}
}
