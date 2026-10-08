package oracle

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// Skip only the direct constructor's numeric boxing. Store the raw double bits
// in a reference slot, as an unfitted entry would, while keeping the mutant C
// type-correct. It must compile and then fail at runtime, not at the C compiler.
func TestMapUnionConstructorBoxingMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/map_union_value_paths.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	constructors := 0
	for _, statement := range program.Main {
		declaration, ok := statement.(ir.Declare)
		if !ok {
			continue
		}
		constructor, ok := declaration.Value.(ir.MapNew)
		if !ok {
			continue
		}
		for _, entry := range constructor.Entries {
			box, ok := entry[1].(ir.Box)
			if !ok {
				continue
			}
			number, ok := box.Value.(ir.NumberConstant)
			if ok && number.Value == 18 {
				constructors++
			}
		}
	}
	if constructors != 1 {
		t.Fatalf("want one boxed numeric constructor entry, got %d", constructors)
	}
	source := native.C(program)
	number := "(" + strconv.FormatFloat(18, 'x', -1, 64) + ")"
	needle := "adamic_box_number(" + number + ")"
	if strings.Count(source, needle) != 1 {
		t.Fatalf("want one constructor boxing call %q", needle)
	}
	mutant := `#include "adamic.h"
static adamic_heap *map_constructor_unboxed_number(double number) {
 adamic_value unboxed = {.number = number};
 return unboxed.reference;
}
` + strings.Replace(source, needle, "map_constructor_unboxed_number("+number+")", 1)
	want := onNode(t, path)
	for _, mode := range []struct {
		name    string
		options native.Options
	}{
		{"release", native.Options{Release: true}},
		{"sanitized", native.Options{Sanitize: true}},
	} {
		t.Run(mode.name, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(mutant, binary, mode.options); err != nil {
				t.Fatalf("mutant must compile: %v", err)
			}
			got := executeWith(t, []string{leakSanitizer()}, binary)
			if got.exitCode == 0 || disagreement(want, got) == "" {
				t.Fatal("constructor boxing mutant survived")
			}
			if mode.options.Sanitize && !strings.Contains(string(got.stderr), "Sanitizer") && !strings.Contains(string(got.stderr), "runtime error") {
				t.Fatalf("want sanitizer to catch the unboxed reference, got exit %d stderr %q", got.exitCode, got.stderr)
			}
			t.Logf("caught missing constructor boxing at runtime: exit %d", got.exitCode)
		})
	}
}
