package oracle

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// Removing the declared absent slots restores the original stop at the later write.
func TestOptionalFieldWriteCatchesDroppedSlot(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_field_write.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name != "maybeRecord" {
			continue
		}
		declaration := function.Body[0].(ir.Declare)
		literal := declaration.Value.(ir.ObjectLiteral)
		if len(literal.Missing) != 1 || literal.Missing[0].Name != "slot" {
			t.Fatal("mutant target absent")
		}
		literal.Missing = nil
		declaration.Value = literal
		function.Body[0] = declaration
		changed = true
	}
	if !changed {
		t.Fatal("mutant target absent")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := execute(t, binary)
	if result.exitCode != 70 || !strings.Contains(string(result.stderr), "a field the checker proved is there is missing") {
		t.Fatalf("want restored missing-field stop: exit %d, stderr %q", result.exitCode, result.stderr)
	}
	if disagreement(onNode(t, path), result) == "" {
		t.Fatal("Node did not catch dropped absent slot")
	}
	t.Log("Node caught dropped absent slot: missing-field stop, exit 70")
}

// Each mutation changes generated C without bypassing compilation or the Node observation.
func TestOptionalFieldPresenceCatchesMutants(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_field_presence.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	expected := onNode(t, path)
	absent := regexp.MustCompile(`adamic_object_absent\(([^,]+), [0-9]+\);`)
	layoutOrder := `#include "adamic.h"
static void mutant_layout_write(adamic_object *object, const char *name, bool initialized) {
    adamic_object_set_initialized(object, name, initialized);
    adamic_slot_cache cache = {NULL, 0};
    (void)adamic_object_field(object, name, &cache);
    adamic_object_orders(object)[cache.index] = cache.index + 1;
}
` + strings.ReplaceAll(source, "adamic_object_set_initialized(", "mutant_layout_write(")
	for _, mutant := range []struct {
		name string
		code string
	}{
		{"initially present", absent.ReplaceAllString(source, `(void)$1;`)},
		{"layout order instead of write order", layoutOrder},
		{"delete does not remove", strings.ReplaceAll(source, "adamic_object_delete(", "adamic_object_has_own(")},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			if mutant.code == source {
				t.Fatal("mutant target absent")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(mutant.code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := execute(t, binary)
			if result.exitCode != 0 {
				t.Fatalf("mutant stopped instead of producing wrong presence: %d %s", result.exitCode, result.stderr)
			}
			if disagreement(expected, result) == "" {
				t.Fatal("Node did not catch presence mutant")
			}
			t.Log("Node caught incorrect field presence or insertion order by stdout")
		})
	}
}

func TestOptionalFieldConstructionCatchesDroppedReservation(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_field_construction.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	reserved := regexp.MustCompile(`adamic_object_copy_reserving_checked\(([^,]+), &adamic_shape_[0-9]+, [^\n]*?\)`)
	mutant := reserved.ReplaceAllString(source, `adamic_object_copy($1)`)
	if mutant == source {
		t.Fatal("mutant target absent")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := execute(t, binary)
	if result.exitCode != 70 || !strings.Contains(string(result.stderr), "a field the checker proved is there is missing") {
		t.Fatalf("want missing-field stop, got %d %s", result.exitCode, result.stderr)
	}
	if disagreement(onNode(t, path), result) == "" {
		t.Fatal("Node did not catch dropped spread reservation")
	}
	t.Log("Node caught dropped spread reservation: missing-field stop, exit 70")
}

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/optional_field_checked_copy.a", true, true})
}
