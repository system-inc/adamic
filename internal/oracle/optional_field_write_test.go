package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
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
static adamic_value *mutant_layout_write(adamic_object *object, const char *name, adamic_slot_cache *cache) {
    adamic_value *slot = adamic_object_write_field(object, name, cache);
    adamic_object_orders(object)[adamic_slot_index(object, slot)] = adamic_slot_index(object, slot) + 1;
    return slot;
}
` + strings.ReplaceAll(source, "adamic_object_write_field(", "mutant_layout_write(")
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

// This restores the old plain-object branch of class key enumeration. It must
// compile and finish cleanly; only the source Node observation catches its list.
func TestOptionalFieldAliasCatchesStaticEnumeration(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_field_alias.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "adamic_class_object_keys(", "mutant_static_keys(")
	if mutant == source {
		t.Fatal("static enumeration mutant target absent")
	}
	helper := `
#include "adamic.h"
#include <string.h>
static adamic_array *mutant_static_keys(const adamic_object *object) {
 const adamic_shape *shape = object->shape;
 adamic_array *keys = adamic_array_new(shape->count, true);
 for (size_t index = 0; index < shape->count; index++) {
  const char *name = shape->names[adamic_public_index(shape, index)];
  adamic_string *key = adamic_string_allocate(strlen(name));
  memcpy((char *)key->bytes, name, key->length);
  adamic_array_push(keys, (adamic_value){.reference = key});
 }
 return keys;
}
`
	expected := onNode(t, path)
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "static-keys")
		if err := native.Build(helper+mutant, binary, native.Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		got := execute(t, binary)
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("mutant must finish cleanly: %d %s", got.exitCode, got.stderr)
		}
		if disagreement(expected, got) == "" {
			t.Fatal("static key list escaped Node comparison")
		}
		t.Logf("Node caught static enumeration, sanitize=%t: %q", sanitize, got.stdout)
	}
}

// This is pending admission, not a passing backend parity fixture.
func TestOptionalFieldCheckedViewPending(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_field_checked_view_pending.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "0\n" {
		t.Fatalf("Node: %+v", truth)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !(strings.Contains(err.Error(), "checked view field slot") || strings.Contains(err.Error(), "a checked field alias requiring an optional, accessor, or representation conversion")) {
		t.Fatalf("pending boundary changed: %v", err)
	}
	t.Skip("acceptance dependency: compiler/views-rehearsal 2b6c032a; checked-view optional-write integration remains NotYet")
}
