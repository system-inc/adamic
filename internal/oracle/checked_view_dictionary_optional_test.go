package oracle

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewDictionaryOptionalContainers(t *testing.T) {
	for _, shape := range []struct {
		fixture, field, declared string
		required                 bool
	}{
		{"optional-watch", "watchOptions", "WatchOptions | undefined", false},
		{"optional-options", "options", "CompilerOptions", true},
		{"optional-wildcard", "wildcardDirectories", "MapLike<WatchDirectoryFlags> | undefined", false},
	} {
		for _, name := range []string{"good", "wrong", "missing", "receiver-missing", "undefined"} {
			t.Run(shape.fixture+"-"+name, func(t *testing.T) {
				program, path := interfaceFixture(t, "dictionaries/source/"+shape.fixture+"-"+name)
				truth := onNode(t, path)
				expected := map[string]string{"good": "object\n", "wrong": "number\n", "missing": "undefined\n", "receiver-missing": "undefined\n", "undefined": "undefined\n"}[name]
				if truth.exitCode != 0 || string(truth.stdout) != expected {
					t.Fatalf("Node: %#v", truth)
				}
				sanitized, binary := nativelyUncached(t, program)
				valid := name != "wrong" && !((name == "missing" || name == "undefined") && shape.required)
				for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if valid {
						if d := disagreement(truth, got); d != "" {
							t.Fatal(d)
						}
						continue
					}
					want := "adamic: panic: cast failed: field read failed: view?." + shape.field + " is not a " + shape.declared + "; expected " + shape.declared + ", found number\n"
					if !shape.required {
						want = "adamic: panic: cast failed: field read failed: view?." + shape.field + " matches no member of " + shape.declared + "; expected " + shape.declared + ", found number\n"
					}
					if name == "missing" {
						want = "adamic: panic: cast failed: field read failed: view?." + shape.field + " is not initialized; expected " + shape.declared + ", found missing\n"
					}
					if name == "undefined" {
						want = "adamic: panic: cast failed: field read failed: view?." + shape.field + " is not a " + shape.declared + "; expected " + shape.declared + ", found nullish\n"
					}
					if got.exitCode != 70 || string(got.stderr) != want {
						t.Fatalf("optional receiver check: %#v; want %q", got, want)
					}
				}
				if valid {
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}

// Restore the receiver/field absence conflation without failing compilation.
func TestCheckedViewDictionaryOptionalAbsenceMutant(t *testing.T) {
	program, _ := interfaceFixture(t, "dictionaries/source/optional-options-missing")
	c := native.C(program)
	if !strings.Contains(c, "adamic_object_optional_view_undefined(") {
		t.Fatal("no optional field check to mutate")
	}
	wrapper := `#include "adamic.h"
static adamic_value optional_absence_mutant(const adamic_object *object, const char *name, adamic_slot_cache *cache, unsigned char wanted, const char *type, const char *expression, bool absent, bool optional, bool undefined_member) {
 return adamic_object_optional_view_undefined(object,name,cache,wanted,type,expression,absent || optional,optional,undefined_member);
}
`
	c = wrapper + strings.ReplaceAll(c, "adamic_object_optional_view_undefined(", "optional_absence_mutant(")
	binary := filepath.Join(t.TempDir(), "optional-mutant")
	if err := native.Build(c, binary, native.Options{}); err != nil {
		t.Fatal("semantic mutant must compile: ", err)
	}
	code := javascript.JavaScript(program)
	if !strings.Contains(code, "adamicViewField(") {
		t.Fatal("no JavaScript optional field check to mutate")
	}
	code = "const optionalAbsenceMutant = (object,name,expression,type,expected,allowed,absent,optional) => adamicViewField(object,name,expression,type,expected,allowed,absent || optional,optional);\n" + strings.ReplaceAll(code, "adamicViewField(", "optionalAbsenceMutant(")
	js := filepath.Join(t.TempDir(), "optional-mutant.mjs")
	if err := os.WriteFile(js, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	for _, got := range []run{execute(t, binary), onNode(t, js)} {
		if got.exitCode != 0 || string(got.stdout) != "undefined\n" {
			t.Fatalf("mutant must lose required presence refusal: %#v", got)
		}
		t.Logf("optional absence mutant caught by required-field control: exit=%d stdout=%q", got.exitCode, got.stdout)
	}
}
