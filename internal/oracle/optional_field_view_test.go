package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, fixture := range []struct {
		path            string
		lowers, checked bool
	}{
		{"internal/oracle/testdata/optional_field_checked_view_pending.a", true, false},
		{"internal/oracle/testdata/optional_field_checked_view.a", true, false},
		{"internal/oracle/testdata/optional_field_checked_view_enumeration.a", true, false},
		{"internal/oracle/testdata/optional_field_checked_view_misfit.a", true, true},
		{"internal/oracle/testdata/optional_field_checked_view_literal.a", true, true},
	} {
		fixtures = append(fixtures, fixture)
	}
}

func TestOptionalFieldCheckedViewWrites(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_field_checked_view.a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 {
		t.Fatalf("Node: %+v", truth)
	}
	got, binary := nativelyUncached(t, program)
	for name, result := range map[string]run{"sanitized": got, "release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, result); difference != "" {
			t.Fatalf("%s %s: exit %d stdout %q stderr %s", name, difference, result.exitCode, result.stdout, result.stderr)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func TestOptionalFieldCheckedViewMisfit(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_field_checked_view_misfit.a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "writing\nstored\n0\n" {
		t.Fatalf("Node: %+v", truth)
	}
	want := run{exitCode: 70, stdout: []byte("writing\n"), stderr: []byte("adamic: panic: field write failed: slot expected string, found number\n")}
	got, _ := nativelyUncached(t, program)
	for name, result := range map[string]run{"sanitized": got, "release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Fatalf("%s %s: exit %d stdout %q stderr %s", name, difference, result.exitCode, result.stdout, result.stderr)
		}
	}
}

func TestOptionalFieldCheckedViewCatchesDroppedStore(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_field_checked_view.a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "adamic_object_view_store(", "mutant_dropped_store(")
	if mutant == source {
		t.Fatal("mutant target absent")
	}
	helper := `
#include "adamic.h"
#include <string.h>
static void mutant_dropped_store(adamic_object *object, const char *name, adamic_slot_cache *cache, adamic_value value, unsigned char wanted) {
 adamic_slot_cache storage = {NULL, 0};
 (void)adamic_object_optional_find(object, name, &storage);
 if (strcmp(name, "slot") == 0 && storage.index < object->shape->count && adamic_object_field_types(object)[storage.index] == 7) { (void)cache; (void)value; (void)wanted; }
 else adamic_object_view_store(object,name,cache,value,wanted);
}
`
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(helper+mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := execute(t, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("dropped store must produce clean wrong output: %+v", got)
	}
	if disagreement(onNode(t, path), got) == "" {
		t.Fatal("dropped checked store escaped Node")
	}
	t.Log("Node catches omitted checked optional store and presence")
}

// Each mutation changes the production runtime, leaving lowering and source intact.
func optionalViewRuntimeMutant(t *testing.T, fixture, nativeFrom, nativeTo, jsFrom, jsTo string, check bool) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+fixture+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if check {
		want = run{exitCode: 70, stdout: []byte("writing\n"), stderr: []byte("adamic: panic: field write failed: slot expected string, found number\n")}
	}
	bytes, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/object.c"))
	if err != nil {
		t.Fatal(err)
	}
	original := string(bytes)
	if strings.Count(original, nativeFrom) != 1 {
		t.Fatal("native mutation must change one production site")
	}
	definitions := regexp.MustCompile(`(?m)^(?:adamic_object \*|adamic_value \*|adamic_closure \*|adamic_maybe_number |adamic_maybe_boolean |adamic_view_union_value |adamic_value |void |bool )((?:adamic_object_|adamic_view_)[a-z_]+)\(`)
	prefix := ""
	for _, match := range definitions.FindAllStringSubmatch(original, -1) {
		prefix += "#define " + match[1] + " optional_mutant_" + match[1] + "\n"
	}
	code := prefix + strings.Replace(original, nativeFrom, nativeTo, 1) + native.C(program)
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(code, binary, native.Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		got := execute(t, binary)
		if strings.Contains(string(got.stderr), "Sanitizer") {
			t.Fatalf("mutant must be caught by semantic assertions: %+v", got)
		}
		if disagreement(want, got) == "" {
			t.Fatal("native mutant escaped")
		}
		t.Logf("native mutant caught, sanitize=%t: exit %d stdout %q stderr %q", sanitize, got.exitCode, got.stdout, got.stderr)
	}
	js := javascript.JavaScript(program)
	if strings.Count(js, jsFrom) != 1 {
		t.Fatal("JS mutation must change one production site")
	}
	file := filepath.Join(t.TempDir(), "mutant.mjs")
	if err := os.WriteFile(file, []byte(strings.Replace(js, jsFrom, jsTo, 1)), 0600); err != nil {
		t.Fatal(err)
	}
	got := onNode(t, file)
	if disagreement(want, got) == "" {
		t.Fatal("JavaScript mutant escaped")
	}
	t.Logf("JavaScript mutant caught: exit %d stdout %q stderr %q", got.exitCode, got.stdout, got.stderr)
}

func TestOptionalFieldCheckedViewCatchesMissingPublication(t *testing.T) {
	t.Parallel()
	optionalViewRuntimeMutant(t, "optional_field_checked_view",
		"    adamic_object_publish(object, cache->index);", "    (void)object;",
		"    adamicWriteField(object, name, value);\n};", "    adamicWriteField(object, name, value); delete object[name];\n};", false)
}

func TestOptionalFieldCheckedViewCatchesMissingWriteGuard(t *testing.T) {
	t.Parallel()
	optionalViewRuntimeMutant(t, "optional_field_checked_view_misfit",
		"    if (!fits) {", "    if (false && !fits) {", "    if (!fits) {", "    if (false && !fits) {", true)
}

func TestOptionalFieldCheckedViewCatchesAbsentReadAsZero(t *testing.T) {
	t.Parallel()
	optionalViewRuntimeMutant(t, "optional_field_checked_view",
		"if (wanted == 7) return (adamic_value){.number = adamic_maybe_number_pack((adamic_maybe_number){false, 0})};",
		"if (wanted == 7) return (adamic_value){.number = adamic_maybe_number_pack((adamic_maybe_number){true, 0})};",
		"    if (value === undefined) return value;", "    if (value === undefined) return 0;", false)
}

func TestOptionalFieldCheckedViewLiteral(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_field_checked_view_literal.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "1\n" {
		t.Fatalf("Node: %+v", truth)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.slot expected 0 | undefined, found number 1\n")}
	got, _ := nativelyUncached(t, program)
	for _, result := range []run{got, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Fatalf("literal optional check %s: %+v", difference, result)
		}
	}
	changed := false
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
		if property, ok := value.(ir.Property); ok && property.Name == "slot" && len(property.ViewAllowed) != 0 {
			property.ViewAllowed = nil
			changed = true
			return property
		}
		return value
	})
	if !changed {
		t.Fatal("literal mutation target absent")
	}
	mutant, _ := nativelyUncached(t, program)
	for _, result := range []run{mutant, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, result); difference != "" {
			t.Fatalf("omitted literal check must reproduce Node: %s %+v", difference, result)
		}
	}
	t.Log("omitted optional literal contract caught by pinned diagnostic in both backends")
}

func TestOptionalFieldCheckedViewEnumeration(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_field_checked_view_enumeration.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 {
		t.Fatalf("Node: %+v", truth)
	}
	got, binary := nativelyUncached(t, program)
	for _, result := range []run{got, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, result); difference != "" {
			t.Fatalf("enumeration %s: %+v", difference, result)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func TestOptionalFieldCheckedViewCatchesBooleanRead(t *testing.T) {
	t.Parallel()
	optionalViewRuntimeMutant(t, "optional_field_checked_view",
		"adamic_maybe_boolean_pack((adamic_maybe_boolean){true, snapshot.payload.boolean})",
		"adamic_maybe_boolean_pack((adamic_maybe_boolean){true, true})",
		"return adamicViewField(object, name, expression, type === 7 ? 1 : type === 9 ? 2 : type, expected, allowed);",
		"return type === 9 ? true : adamicViewField(object, name, expression, type === 7 ? 1 : type === 9 ? 2 : type, expected, allowed);", false)
}

func TestOptionalFieldCheckedViewCatchesMissingBoxing(t *testing.T) {
	t.Parallel()
	optionalViewRuntimeMutant(t, "optional_field_checked_view",
		"if (incoming == 1) value.reference = adamic_box_number(value.number);",
		"if (incoming == 1) value.reference = NULL;",
		"    adamicWriteField(object, name, value);\n};",
		"    adamicWriteField(object, name, actual === 10 ? undefined : value);\n};", false)
}
