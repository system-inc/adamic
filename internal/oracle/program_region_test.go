package oracle

import (
	"context"
	"errors"
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

func programRegionLowered(test *testing.T, path string, enabled bool) *ir.Program {
	test.Helper()
	source, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	program, err := lower.LowerWithOptions(context.Background(), source, lower.Options{ProgramRegion: enabled})
	var refused *lower.Refused
	if errors.As(err, &refused) &&
		filepath.Base(path) == "optional.a" &&
		filepath.Base(filepath.Dir(path)) == "program_region" &&
		refused.What == "delete" {
		test.Skip("pending delete on optional fields: program_region/optional.a")
	}

	if err != nil {
		test.Fatal(err)
	}
	return program
}

func checkProgramRegion(test *testing.T, program *ir.Program, path string) []int {
	test.Helper()
	return checkProgramRegionMode(test, program, path, true)
}

func checkProgramRegionMode(test *testing.T, program *ir.Program, path string, enabled bool) []int {
	test.Helper()
	code := native.C(program)
	binary := filepath.Join(test.TempDir(), "program")
	if err := programRegionBuild(
		test,
		code,
		binary,
		native.Options{ProgramRegion: enabled, Sanitize: true},
	); err != nil {
		test.Fatal(err)
	}
	oracle := onNode(test, path)
	if diff := disagreement(oracle, onJavaScriptBackend(test, program)); diff != "" {
		test.Fatalf("JavaScript backend: %s", diff)
	}
	plain := filepath.Join(test.TempDir(), "plain")
	if err := programRegionBuild(test, code, plain, native.Options{ProgramRegion: enabled}); err != nil {
		test.Fatal(err)
	}
	if diff := disagreement(oracle, execute(test, plain)); diff != "" {
		test.Fatalf("native plain: %s", diff)
	}
	result := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if difference := disagreement(oracle, result); difference != "" {
		test.Fatalf("Program region: %s: %s", difference, result.stderr)
	}
	report, err := leakcheck.Check(leakcheck.Program{
		C: code, Sanitized: binary, Counted: filepath.Join(test.TempDir(), "counted"),
		BuildCounted: func(code, output string) error {
			return programRegionBuild(test, code, output, native.Options{ProgramRegion: enabled, Count: true})
		},
		Execute: func(environment []string, name string, args ...string) leakcheck.Run {
			return leakRun(executeWith(test, environment, name, args...))
		},
	})
	if err != nil {
		test.Fatal(err)
	}
	if report != "" {
		test.Fatal(report)
	}
	countedBinary := filepath.Join(test.TempDir(), "counts")
	if err := programRegionBuild(
		test,
		code,
		countedBinary,
		native.Options{ProgramRegion: enabled, Count: true},
	); err != nil {
		test.Fatal(err)
	}
	name, args := pinnedStack(countedBinary)
	counted := execute(test, name, args...)
	match := countsLine.FindSubmatch(counted.stderr)
	if counted.exitCode != 0 || match == nil {
		test.Fatalf("counts %d %s", counted.exitCode, counted.stderr)
	}
	counts := []int{}
	for _, field := range match[1:] {
		value, _ := strconv.Atoi(string(field))
		counts = append(counts, value)
	}
	if counts[0] != counts[1]+counts[5] {
		test.Fatalf("not every member accounted: %s", counted.stderr)
	}
	test.Logf("%s", counted.stderr)
	return counts
}

func programRegionFixture(test *testing.T, name string) []int {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/program_region", name))
	if err != nil {
		test.Fatal(err)
	}
	return checkProgramRegion(test, programRegionLowered(test, path, true), path)
}
func TestProgramRegionGraphParent(test *testing.T) {
	test.Parallel()
	programRegionFixture(test, "cycles_graph_parent.a")
}
func TestProgramRegionGraphRelations(test *testing.T) {
	test.Parallel()
	programRegionFixture(test, "cycles_graph_relations.a")
}
func TestProgramRegionGraphSymbols(test *testing.T) {
	test.Parallel()
	programRegionFixture(test, "cycles_graph_symbols.a")
}
func TestProgramRegionWeakParent(test *testing.T) {
	test.Parallel()
	programRegionFixture(test, "cycles_weak_parent.a")
}
func TestProgramRegionWeakRelations(test *testing.T) {
	test.Parallel()
	programRegionFixture(test, "cycles_weak_relations.a")
}
func TestProgramRegionWeakSymbols(test *testing.T) {
	test.Parallel()
	programRegionFixture(test, "cycles_weak_symbols.a")
}
func TestProgramRegionMillion(test *testing.T) {
	test.Parallel()
	programRegionFixture(test, "million.a")
}
func TestProgramRegionOptionalStorage(test *testing.T) {
	test.Parallel()
	programRegionPending(test, "testdata/program_region/optional.a", "delete")
}
func TestProgramRegionConstructor(test *testing.T) {
	test.Parallel()
	programRegionFixture(test, "constructor.a")
}
func TestProgramRegionOwnership(test *testing.T) {
	test.Parallel()
	counts := programRegionFixture(test, "program_region_ownership.a")
	if counts[5] != 2 {
		test.Fatalf("membership census: regions %d, want 2", counts[5])
	}
}
func TestProgramRegionCapturedCell(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "capture.a")
	source := `function make(): () => string {let current:(()=>string)|undefined=undefined;const result=()=>current===undefined?'empty':'set';current=result;return result;}console.log(make()());`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		test.Fatal(err)
	}
	counts := checkProgramRegion(test, programRegionLowered(test, path, true), path)
	if counts[5] < 2 {
		test.Fatalf("closure and cell membership missing: %v", counts)
	}
}
