package oracle

import (
	"context"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
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

// Register counted witnesses with the ordinary flag-off Node oracle and counts.
func init() {
	for _, path := range []string{"cycles_weak_parent.a", "cycles_weak_relations.a", "cycles_weak_symbols.a", "program_region_map_storage.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + path, true, false})
	}
}

func TestProgramRegionCyclesGraphParent(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/cycles_graph_parent.a"))
	checkProgramRegion(t, programRegionLowered(t, path, true), path)
}
func TestProgramRegionCyclesGraphRelations(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/cycles_graph_relations.a"))
	checkProgramRegion(t, programRegionLowered(t, path, true), path)
}
func TestProgramRegionCyclesGraphSymbols(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/cycles_graph_symbols.a"))
	checkProgramRegion(t, programRegionLowered(t, path, true), path)
}
func TestProgramRegionCyclesWeakParent(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/cycles_weak_parent.a"))
	checkProgramRegion(t, programRegionLowered(t, path, true), path)
}
func TestProgramRegionCyclesWeakRelations(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/cycles_weak_relations.a"))
	checkProgramRegion(t, programRegionLowered(t, path, true), path)
}
func TestProgramRegionCyclesWeakSymbols(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/cycles_weak_symbols.a"))
	checkProgramRegion(t, programRegionLowered(t, path, true), path)
}
func TestProgramRegionGraphRegionsMillion(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/graph_regions/million.a"))
	checkProgramRegion(t, programRegionLowered(t, path, true), path)
}
func TestProgramRegionProgramRegionOwnership(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/program_region_ownership.a"))
	checkProgramRegion(t, programRegionLowered(t, path, true), path)
}
func checkProgramRegionMapperStorage(t *testing.T, enabled bool) {
	t.Helper()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/program_region_map_storage.a"))
	oracle := onNode(t, path)

	code := native.C(programRegionLowered(t, path, enabled))
	allocation := regexp.MustCompile(`adamic_array \*\s*(\w+) = \((\w+) \? adamic_retain\((\w+)\) : adamic_array_new\([^\n]+;`)
	matches := allocation.FindAllStringSubmatchIndex(code, -1)
	if len(matches) != 1 {
		t.Fatalf("expected exactly one reusable mapper allocation, got %d", len(matches))
	}
	m := matches[0]
	result, unique, source := code[m[2]:m[3]], code[m[4]:m[5]], code[m[6]:m[7]]
	condition, message := result+" != "+source, "counted mapper control did not reuse"
	if enabled {
		condition, message = result+" == "+source, "Program member array storage was reused"
	}
	assertion := fmt.Sprintf("\n if (adamic_program_is(%s) != %t) adamic_panic(\"mapper fixture has wrong membership\", sizeof \"mapper fixture has wrong membership\" - 1);\n if (%s) adamic_panic(\"%s\", sizeof \"%s\" - 1);\n", source, enabled, condition, message, message)
	checked := code[:m[1]] + assertion + code[m[1]:]
	binary := filepath.Join(t.TempDir(), "storage")
	if err := native.Build(checked, binary, native.Options{ProgramRegion: enabled, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	resultRun := executeWith(t, []string{leakSanitizer()}, binary)
	if difference := disagreement(oracle, resultRun); difference != "" {
		t.Fatalf("storage invariant: %s %s", difference, resultRun.stderr)
	}
	report, err := leakcheck.Check(leakcheck.Program{
		C: checked, Sanitized: binary, Counted: filepath.Join(t.TempDir(), "counted"),
		BuildCounted: func(code, output string) error {
			return native.Build(code, output, native.Options{ProgramRegion: enabled, Count: true})
		},
		Execute: func(environment []string, name string, args ...string) leakcheck.Run {
			return leakRun(executeWith(t, environment, name, args...))
		},
	})
	if err != nil || report != "" {
		t.Fatalf("leak check: %v %s", err, report)
	}
	t.Logf("flag=%t: actual storage assertion, unchanged Node output, ASan/UBSan and leak check passed", enabled)
	if !enabled {
		return
	}
	// Runtime uniqueness must reject marked storage. Admit it deliberately at
	// the selected mapper branch and require the independent address invariant
	// to fail under the same sanitizer build and unchanged Node oracle.
	guard := regexp.MustCompile(`bool ` + regexp.QuoteMeta(unique) + ` = [^\n]+;`)
	if len(guard.FindAllString(checked, -1)) != 1 {
		t.Fatal("unique guard missing/ambiguous")
	}
	mutant := guard.ReplaceAllString(checked, "bool "+unique+" = true;")
	bad := filepath.Join(t.TempDir(), "reuse-member-mutant")
	if err := native.Build(mutant, bad, native.Options{ProgramRegion: true, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	rejected := executeWith(t, []string{leakSanitizer()}, bad)
	if rejected.exitCode != 70 || !strings.Contains(string(rejected.stderr), "Program member array storage was reused") || disagreement(oracle, rejected) == "" {
		t.Fatalf("member reuse mutant escaped: %d %s", rejected.exitCode, rejected.stderr)
	}
	t.Log("reuse-admitted mutant caught by actual storage invariant under ASan/UBSan")
}

func TestProgramRegionMapperStorage(t *testing.T) {
	t.Parallel()
	checkProgramRegionMapperStorage(t, true)
}
func TestProgramRegionMapperCountedStorage(t *testing.T) {
	t.Parallel()
	checkProgramRegionMapperStorage(t, false)
}

func checkProgramRegionMembershipMutant(t *testing.T, extra bool) {
	t.Helper()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/program_region_ownership.a"))
	baseline := checkProgramRegion(t, programRegionLowered(t, path, true), path)
	program := programRegionLowered(t, path, true)
	changed := false
	var mutate func(reflect.Value)
	mutate = func(value reflect.Value) {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return
			}
			copy := reflect.New(value.Elem().Type()).Elem()
			copy.Set(value.Elem())
			mutate(copy)
			value.Set(copy)
		case reflect.Ptr:
			if !value.IsNil() {
				mutate(value.Elem())
			}
		case reflect.Struct:
			if literal, ok := value.Interface().(ir.ObjectLiteral); ok && !changed {
				for _, field := range literal.Fields {
					if extra && field.Name == "text" || !extra && field.Name == "id" {
						if literal.ProgramRegion == extra {
							t.Fatal("baseline mark does not match ruling mutation")
						}
						value.FieldByName("ProgramRegion").SetBool(extra)
						changed = true
						break
					}
				}
			}
			for i := 0; i < value.NumField(); i++ {
				mutate(value.Field(i))
			}
		case reflect.Slice:
			for i := 0; i < value.Len(); i++ {
				mutate(value.Index(i))
			}
		}
	}
	mutate(reflect.ValueOf(program).Elem())
	if !changed {
		t.Fatal("ruling mutant did not change an allocation")
	}
	counts := checkProgramRegion(t, program, path)
	delta := -1
	if extra {
		delta = 1
	}
	if counts[5] != baseline[5]+delta {
		t.Fatal("ruling membership mutation was not accounted")
	}
	t.Logf("ruling mutant extra=%t: Node unchanged, sanitized and leak clean, member delta %d", extra, delta)
}
func TestProgramRegionExtraLeafMutant(t *testing.T) {
	t.Parallel()
	checkProgramRegionMembershipMutant(t, true)
}
func TestProgramRegionMissingCyclicMemberMutant(t *testing.T) {
	t.Parallel()
	checkProgramRegionMembershipMutant(t, false)
}

func TestProgramRegionConstruction(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/program_region_construction.a"))
	checkProgramRegion(t, programRegionLowered(t, path, true), path)
}
func TestProgramRegionShortObjectMutant(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/program_region_construction.a"))
	code := native.C(programRegionLowered(t, path, true))
	sizes := regexp.MustCompile(`adamic_object_size\((\w+)->shape->count\)`)
	if len(sizes.FindAllString(code, -1)) == 0 {
		t.Fatal("no object adoption to mutate")
	}
	mutant := sizes.ReplaceAllString(code, "sizeof(adamic_object) + $1->shape->count * sizeof(adamic_value)")
	binary := filepath.Join(t.TempDir(), "short-object")
	if err := native.Build(mutant, binary, native.Options{ProgramRegion: true, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{leakSanitizer()}, binary)
	if result.exitCode == 0 || !strings.Contains(string(result.stderr), "heap-buffer-overflow") {
		t.Fatalf("short object mutant escaped: %d %s", result.exitCode, result.stderr)
	}
	t.Log("field-only adoption size caught by ASan heap-buffer-overflow")
}

// Program rows use an explicit switch, so ordinary counted rows stay comparable.
func init() {
	additionalFixtureCounts = append(additionalFixtureCounts, func(t *testing.T) []string {
		paths := []string{"cycles_graph_parent.a", "cycles_graph_relations.a", "cycles_graph_symbols.a", "cycles_weak_parent.a", "cycles_weak_relations.a", "cycles_weak_symbols.a", "graph_regions/million.a", "program_region_ownership.a", "program_region_map_storage.a", "program_region_construction.a", "program_region_closure.a"}
		rows := make([]string, 0, len(paths))
		for _, path := range paths {
			source := "internal/oracle/testdata/" + path
			absolute, err := filepath.Abs(filepath.Join(repository, source))
			if err != nil {
				t.Fatal(err)
			}
			code := native.C(programRegionLowered(t, absolute, true))
			binary := filepath.Join(t.TempDir(), "program-counts")
			if err := native.Build(code, binary, native.Options{ProgramRegion: true, Count: true}); err != nil {
				t.Fatal(err)
			}
			name, args := pinnedStack(binary)
			result := execute(t, name, args...)
			match := countsLine.FindSubmatch(result.stderr)
			if result.exitCode != 0 || match == nil {
				t.Fatalf("Program counts %s: %d %s", path, result.exitCode, result.stderr)
			}
			rows = append(rows, fmt.Sprintf("| %s (Program region) | %s | %s | %s | %s | %s | %s |", source, match[1], match[2], match[3], match[4], match[5], match[6]))
		}
		return rows
	})
}

func TestProgramRegionClosure(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/program_region_closure.a"))
	checkProgramRegion(t, programRegionLowered(t, path, true), path)
}

func checkProgramRegionDisabled(t *testing.T, fixture string) {
	t.Helper()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+fixture))
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.LowerWithOptions(context.Background(), loaded, lower.Options{})
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "adamic/cycle-capable") {
		t.Fatalf("flag off must refuse strong cycles: %v", err)
	}
}
func TestProgramRegionDisabledParent(t *testing.T) {
	t.Parallel()
	checkProgramRegionDisabled(t, "cycles_graph_parent.a")
}
func TestProgramRegionDisabledRelations(t *testing.T) {
	t.Parallel()
	checkProgramRegionDisabled(t, "cycles_graph_relations.a")
}
func TestProgramRegionDisabledSymbols(t *testing.T) {
	t.Parallel()
	checkProgramRegionDisabled(t, "cycles_graph_symbols.a")
}
func TestProgramRegionDisabledMillion(t *testing.T) {
	t.Parallel()
	checkProgramRegionDisabled(t, "graph_regions/million.a")
}
func TestProgramRegionDisabledOwnership(t *testing.T) {
	t.Parallel()
	checkProgramRegionDisabled(t, "program_region_ownership.a")
}
func TestProgramRegionDisabledConstruction(t *testing.T) {
	t.Parallel()
	checkProgramRegionDisabled(t, "program_region_construction.a")
}
func TestProgramRegionDisabledClosure(t *testing.T) {
	t.Parallel()
	checkProgramRegionDisabled(t, "program_region_closure.a")
}
