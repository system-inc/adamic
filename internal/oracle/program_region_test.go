package oracle

import (
	"context"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
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

func TestProgramRegionFixtures(t *testing.T) {
	additional := map[string]bool{
		"graph_regions_closure.a": true, "graph_regions_generic_capture.a": true,
		"graph_regions_static_private.a": true, "graph_regions_structural_literal.a": true,
		"graph_regions_throw.a": true, "graph_regions_flow.a": true,
		"graph_regions_counted_container.a": true, "graph_regions_weak_mixed.a": true,
	}
	for _, fixture := range fixtures {
		if !strings.Contains(fixture.path, "/cycles_") && !strings.HasSuffix(fixture.path, "graph_regions/million.a") && !strings.HasSuffix(fixture.path, "program_region_ownership.a") && !additional[filepath.Base(fixture.path)] {
			continue
		}
		t.Run(fixture.path, func(t *testing.T) {
			path, _ := filepath.Abs(filepath.Join(repository, fixture.path))
			program := programRegionLowered(t, path, true)
			counts := checkProgramRegion(t, program, path)
			if counts[5] == 0 && !strings.HasSuffix(path, "cycles_weak_symbols.a") && !strings.HasSuffix(path, "cycles_weak_relations.a") {
				t.Fatal("fixture acquired no Program members")
			}
		})
	}
}

// The required inference mutants preserve behavior; they must be leak-clean and
// sanitizer-clean rather than deliberately corrupting lifetime machinery.
func TestProgramRegionInferenceMutants(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/program_region_ownership.a"))
	baseline := checkProgramRegion(t, programRegionLowered(t, path, true), path)
	for _, mode := range []string{"extra-member", "missing-member"} {
		t.Run(mode, func(t *testing.T) {
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
						names := map[string]bool{}
						for _, field := range literal.Fields {
							names[field.Name] = true
						}
						if (mode == "extra-member" && names["text"]) || (mode == "missing-member" && names["id"]) {
							if literal.ProgramRegion != (mode == "missing-member") {
								t.Fatal("baseline membership did not match mutation")
							}
							value.FieldByName("ProgramRegion").SetBool(mode == "extra-member")
							changed = true
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
				t.Fatal("mutant did not affect an allocation")
			}
			counts := checkProgramRegion(t, program, path)
			if mode == "extra-member" && counts[5] != baseline[5]+1 {
				t.Fatal("over-inclusion did not retain one extra member until teardown")
			}
			if mode == "missing-member" && counts[5] != baseline[5]-1 {
				t.Fatal("under-inclusion did not fall back to graph allocation")
			}
		})
	}
}
