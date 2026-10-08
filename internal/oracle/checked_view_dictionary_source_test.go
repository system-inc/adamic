package oracle

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These controls run compiler-generated C and JavaScript, not component adapters.
func TestCheckedViewDictionarySource(t *testing.T) {
	for _, test := range []struct{ name, truth, stderr string }{
		{"options-good", "42\nundefined\n", ""},
		{"options-read-wrong", "object\n", "adamic: panic: field read failed: view.options['value']; expected string | number | boolean | undefined, found object\n"},
		{"array-read-wrong", "object\n", "adamic: panic: field read failed: view.symbols['item']; expected Entry | undefined, found array\n"},
		{"options-wrong", "[object Object]\nundefined\n", "adamic: panic: field read failed: view.options['value']; expected string | number | boolean | undefined, found object\n"},
		{"nested-good", "name\n", ""},
		{"nested-wrong", "42\n", "adamic: panic: field read failed: entry.name is not a string; expected string, found number\n"},
		{"array-wrong", "undefined\n", "adamic: panic: field read failed: view.symbols['item']; expected Entry | undefined, found array\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "dictionaries/components/"+test.name)
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != test.truth {
				t.Fatalf("source Node: %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if test.stderr == "" {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatal(difference)
					}
				} else if got.exitCode != 70 || string(got.stderr) != test.stderr {
					t.Fatalf("want pinned exit 70: %#v", got)
				}
			}
			if test.stderr == "" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

func TestCheckedViewDictionarySourceMutants(t *testing.T) {
	for _, mutant := range []struct{ name, fixture, call, wrapper, jsCall, jsWrapper string }{
		{"skip-check", "options-read-wrong", "adamic_view_dictionary_source_read(", `static adamic_view_dictionary_result dictionary_mutant(const adamic_object *object, const adamic_string *key, unsigned int kinds, size_t child, const char *expression, const char *declared) { (void)kinds; return adamic_view_dictionary_source_read(object,key,~0u,child,expression,declared); }`, "adamicViewDictionaryRead(", `const dictionaryMutant = (object,key,kinds,child,expression,declared) => adamicViewDictionaryRead(object,key,['number','boolean','string','object','array','undefined'],child,expression,declared);`},
		{"accept-wrong-shape", "array-read-wrong", "adamic_view_dictionary_source_read(", `static adamic_view_dictionary_result dictionary_mutant(const adamic_object *object, const adamic_string *key, unsigned int kinds, size_t child, const char *expression, const char *declared) { return adamic_view_dictionary_source_read(object,key,kinds | (1u << adamic_view_union_array),child,expression,declared); }`, "adamicViewDictionaryRead(", `const dictionaryMutant = (object,key,kinds,child,expression,declared) => adamicViewDictionaryRead(object,key,[...kinds,'array'],child,expression,declared);`},
		// A native unchecked scalar read needs a safe physical adapter. Conversion
		// avoids pointer reinterpretation while removing the declared string check.
		{"drop-transitive-check", "nested-wrong", "adamic_object_view(", `static adamic_value dictionary_mutant(const adamic_object *object, const char *name, adamic_slot_cache *cache, unsigned char wanted, const char *type, const char *expression) { (void)type; adamic_value *slot=adamic_object_read(object,name,cache,expression); if (wanted == 3 && adamic_object_field_types(object)[cache->index] == 1) return (adamic_value){.reference=adamic_string_from_number(slot->number)}; return *slot; }`, "adamicViewField(", `const dictionaryMutant = (object,key) => object[key];`},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "dictionaries/components/"+mutant.fixture)
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatal("source Node failed")
			}
			c := native.C(program)
			if !strings.Contains(c, mutant.call) {
				t.Fatal("source mutation changed no emitted call")
			}
			c = "#include \"view_dictionaries.h\"\n" + mutant.wrapper + "\n" + strings.ReplaceAll(c, mutant.call, "dictionary_mutant(")
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(c, binary, native.Options{}); err != nil {
				t.Fatal("mutant must compile: ", err)
			}
			js := filepath.Join(t.TempDir(), "mutant.mjs")
			code := javascript.JavaScript(program)
			if !strings.Contains(code, mutant.jsCall) {
				t.Fatal("source mutation changed no JavaScript call")
			}
			code = mutant.jsWrapper + "\n" + strings.ReplaceAll(code, mutant.jsCall, "dictionaryMutant(")
			if err := os.WriteFile(js, []byte(code), 0600); err != nil {
				t.Fatal(err)
			}
			for _, got := range []run{execute(t, binary), onNode(t, js)} {
				if got.exitCode != 0 || len(got.stdout) == 0 {
					t.Fatalf("mutant must lose refusal: %#v", got)
				}
				t.Logf("pinned source read catches %s: exit=%d stdout=%q stderr=%q", mutant.name, got.exitCode, got.stdout, got.stderr)
			}
		})
	}
}

func TestCheckedViewDictionaryStorageMutants(t *testing.T) {
	for _, mutant := range []struct{ name, fixture, call, wrapper, jsCall, jsWrapper string }{
		{"skip-check", "storage-union-shape-wrong", "adamic_view_dictionary_source_read(", `static adamic_view_dictionary_result dictionary_mutant(const adamic_object *object, const adamic_string *key, unsigned int kinds, size_t child, const char *expression, const char *declared) { (void)kinds; return adamic_view_dictionary_source_read(object,key,~0u,child,expression,declared); }`, "adamicViewDictionaryRead(", `const dictionaryMutant = (object,key,kinds,child,expression,declared) => adamicViewDictionaryRead(object,key,['number','boolean','string','object','array','undefined'],child,expression,declared);`},
		{"accept-wrong-shape", "storage-array-shape-wrong", "adamic_view_dictionary_source_read(", `static adamic_view_dictionary_result dictionary_mutant(const adamic_object *object, const adamic_string *key, unsigned int kinds, size_t child, const char *expression, const char *declared) { return adamic_view_dictionary_source_read(object,key,kinds | (1u << adamic_view_union_array),child,expression,declared); }`, "adamicViewDictionaryRead(", `const dictionaryMutant = (object,key,kinds,child,expression,declared) => adamicViewDictionaryRead(object,key,[...kinds,'array'],child,expression,declared);`},
		// A native unchecked scalar read needs a safe physical adapter. Conversion
		// avoids pointer reinterpretation while removing the declared string check.
		{"drop-transitive-check", "storage-object-wrong", "adamic_object_view(", `static adamic_value dictionary_mutant(const adamic_object *object, const char *name, adamic_slot_cache *cache, unsigned char wanted, const char *type, const char *expression) { (void)type; adamic_value *slot=adamic_object_read(object,name,cache,expression); if (wanted == 3 && adamic_object_field_types(object)[cache->index] == 1) return (adamic_value){.reference=adamic_string_from_number(slot->number)}; return *slot; }`, "adamicViewField(", `const dictionaryMutant = (object,key) => object[key];`},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "dictionaries/source/"+mutant.fixture)
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatal("source Node failed")
			}
			c := native.C(program)
			if !strings.Contains(c, mutant.call) {
				t.Fatal("source mutation changed no emitted call")
			}
			c = "#include \"view_dictionaries.h\"\n" + mutant.wrapper + "\n" + strings.ReplaceAll(c, mutant.call, "dictionary_mutant(")
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(c, binary, native.Options{}); err != nil {
				t.Fatal("mutant must compile: ", err)
			}
			js := filepath.Join(t.TempDir(), "mutant.mjs")
			code := javascript.JavaScript(program)
			if !strings.Contains(code, mutant.jsCall) {
				t.Fatal("source mutation changed no JavaScript call")
			}
			code = mutant.jsWrapper + "\n" + strings.ReplaceAll(code, mutant.jsCall, "dictionaryMutant(")
			if err := os.WriteFile(js, []byte(code), 0600); err != nil {
				t.Fatal(err)
			}
			for _, got := range []run{execute(t, binary), onNode(t, js)} {
				if got.exitCode != 0 || len(got.stdout) == 0 {
					t.Fatalf("mutant must lose refusal: %#v", got)
				}
				t.Logf("pinned source read catches %s: exit=%d stdout=%q stderr=%q", mutant.name, got.exitCode, got.stdout, got.stderr)
			}
		})
	}
}
