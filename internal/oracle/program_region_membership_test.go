package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Port of d23a4fdc's ruling mutants. The landing already has the actual-storage
// reuse mutant; keep that proof rather than duplicate it here.
func TestProgramRegionExtraMemberMutant(t *testing.T) {
	t.Parallel()
	programRegionInferenceMutant(t, "extra-member")
}
func TestProgramRegionMissingMemberMutant(t *testing.T) {
	t.Parallel()
	programRegionInferenceMutant(t, "missing-member")
}

// The required inference mutants preserve behavior; they must be leak-clean and
// sanitizer-clean rather than deliberately corrupting lifetime machinery.
func programRegionInferenceMutant(t *testing.T, mode string) {
	t.Helper()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/program_region/program_region_ownership.a"))
	baseline := checkProgramRegion(t, programRegionLowered(t, path, true), path)
	if baseline[5] != 2 {
		t.Fatalf("baseline membership census: regions %d, want 2", baseline[5])
	}

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
	if counts[5] == 2 {
		t.Fatal("mutant escaped the independent two-member census")
	}
	t.Logf("%s mutant caught by membership census: regions %d, baseline 2; Node, sanitizers and leaks clean", mode, counts[5])
	if mode == "extra-member" && counts[5] != baseline[5]+1 {
		t.Fatal("over-inclusion did not retain one extra member until teardown")
	}
	if mode == "missing-member" && counts[5] != baseline[5]-1 {
		t.Fatal("under-inclusion did not lose one member in the independent count census")
	}
}
