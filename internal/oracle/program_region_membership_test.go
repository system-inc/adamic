package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Port of d23a4fdc's ruling mutants. The landing already has the actual-storage
// reuse mutant; keep that proof rather than duplicate it here.
func TestProgramRegionExtraMemberMutant(test *testing.T) {
	test.Parallel()
	programRegionInferenceMutant(test, "extra-member")
}
func TestProgramRegionMissingMemberMutant(test *testing.T) {
	test.Parallel()
	programRegionInferenceMutant(test, "missing-member")
}

// The required inference mutants preserve behavior; they must be leak-clean and
// sanitizer-clean rather than deliberately corrupting lifetime machinery.
func programRegionInferenceMutant(test *testing.T, mode string) {
	test.Helper()
	path, _ := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/program_region/program_region_ownership.a"),
	)
	baseline := checkProgramRegion(test, programRegionLowered(test, path, true), path)
	if baseline[5] != 2 {
		test.Fatalf("baseline membership census: regions %d, want 2", baseline[5])
	}

	program := programRegionLowered(test, path, true)
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
						test.Fatal("baseline membership did not match mutation")
					}
					value.FieldByName("ProgramRegion").SetBool(mode == "extra-member")
					changed = true
				}
			}
			for index := 0; index < value.NumField(); index++ {
				mutate(value.Field(index))
			}
		case reflect.Slice:
			for index := 0; index < value.Len(); index++ {
				mutate(value.Index(index))
			}
		}
	}
	mutate(reflect.ValueOf(&program.Main).Elem())
	for index := range program.Functions {
		mutate(reflect.ValueOf(&program.Functions[index].Body).Elem())
	}
	if !changed {
		test.Fatal("mutant did not affect an allocation")
	}
	counts := checkProgramRegion(test, program, path)
	if counts[5] == 2 {
		test.Fatal("mutant escaped the independent two-member census")
	}
	test.Logf(
		"%s mutant caught by membership census: regions %d, baseline 2; Node, sanitizers and leaks clean",
		mode,
		counts[5],
	)
	if mode == "extra-member" && counts[5] != baseline[5]+1 {
		test.Fatal("over-inclusion did not retain one extra member until teardown")
	}
	if mode == "missing-member" && counts[5] != baseline[5]-1 {
		test.Fatal("under-inclusion did not lose one member in the independent count census")
	}
}
