package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewDictionaryFiniteKeys(t *testing.T) {
	for _, shape := range []struct {
		fixture, declared, leaf string
		optional                bool
	}{
		{"indexed-cache", "Type | undefined", "value.flags", true},
		{"iteration-cache", "Type", "value.flags", false},
		{"iterable-cache", "IterationTypes | undefined", "value.nextType.flags", true},
		{"signature-cache", "Signature | undefined", "value.minArgumentCount", true},
	} {
		for _, name := range []string{"good", "wrong", "missing", "nested-wrong", "wrong-array"} {
			t.Run(shape.fixture+"-"+name, func(t *testing.T) {
				program, path := interfaceFixture(t, "dictionaries/source/"+shape.fixture+"-"+name)
				truth := onNode(t, path)
				output := map[string]string{"good": "object\n", "wrong": "number\n", "missing": "undefined\n", "nested-wrong": "wrong\n", "wrong-array": "object\n"}[name]
				if truth.exitCode != 0 || string(truth.stdout) != output {
					t.Fatalf("Node: %#v", truth)
				}
				sanitized, binary := nativelyUncached(t, program)
				valid := name == "good" || name == "missing" && shape.optional
				for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if valid {
						if d := disagreement(truth, got); d != "" {
							t.Fatal(d)
						}
						continue
					}
					found := map[string]string{"wrong": "number", "missing": "undefined", "wrong-array": "array"}[name]
					want := "adamic: panic: field read failed: view[key]; expected " + shape.declared + ", found " + found + "\n"
					if name == "nested-wrong" {
						want = "adamic: panic: field read failed: " + shape.leaf + " is not a number; expected number, found string\n"
					}
					if got.exitCode != 70 || viewReadDiagnosticMismatch(want, got.stderr, program) {
						t.Fatalf("finite-key check: %#v; want %q", got, want)
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

func TestCheckedViewDictionaryFiniteMutants(t *testing.T) {
	for _, mutant := range []struct{ name, fixture, call, wrapper, jsCall, jsWrapper string }{
		{"skip-check", "indexed-cache-wrong-array", "adamic_view_dictionary_source_read(", `static adamic_view_dictionary_result finite_mutant(const adamic_object *object, const adamic_string *key, unsigned int kinds, size_t child, const char *expression, const char *declared) { (void)kinds; return adamic_view_dictionary_source_read(object,key,~0u,child,expression,declared); }`, "adamicViewDictionaryRead(", `const finiteMutant=(object,key,kinds,child,expression,declared)=>adamicViewDictionaryRead(object,key,['number','boolean','string','object','array','undefined'],child,expression,declared);`},
		{"accept-wrong-shape", "indexed-cache-wrong-array", "adamic_view_dictionary_source_read(", `static adamic_view_dictionary_result finite_mutant(const adamic_object *object, const adamic_string *key, unsigned int kinds, size_t child, const char *expression, const char *declared) { return adamic_view_dictionary_source_read(object,key,kinds | (1u << adamic_view_union_array),child,expression,declared); }`, "adamicViewDictionaryRead(", `const finiteMutant=(object,key,kinds,child,expression,declared)=>adamicViewDictionaryRead(object,key,[...kinds,'array'],child,expression,declared);`},
		{"drop-transitive-check", "indexed-cache-nested-wrong", "adamic_object_view(", `static adamic_value finite_mutant(const adamic_object *object, const char *name, adamic_slot_cache *cache, unsigned char wanted, const char *type, const char *expression) { if (wanted==1) return (adamic_value){.number=42}; return adamic_object_view(object,name,cache,wanted,type,expression); }`, "adamicViewField(", `const finiteMutant=(object,key)=>object[key];`},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			program, _ := interfaceFixture(t, "dictionaries/source/"+mutant.fixture)
			c := native.C(program)
			if !strings.Contains(c, mutant.call) {
				t.Fatal("no native check to mutate")
			}
			c = "#include \"view_dictionaries.h\"\n" + mutant.wrapper + "\n" + strings.ReplaceAll(c, mutant.call, "finite_mutant(")
			binary := filepath.Join(t.TempDir(), "finite-mutant")
			if err := native.Build(c, binary, native.Options{}); err != nil {
				t.Fatal("semantic mutant must compile: ", err)
			}
			code := javascript.JavaScript(program)
			if !strings.Contains(code, mutant.jsCall) {
				t.Fatal("no JS check to mutate")
			}
			code = mutant.jsWrapper + "\n" + strings.ReplaceAll(code, mutant.jsCall, "finiteMutant(")
			js := filepath.Join(t.TempDir(), "finite-mutant.mjs")
			if err := os.WriteFile(js, []byte(code), 0600); err != nil {
				t.Fatal(err)
			}
			for _, got := range []run{execute(t, binary), onNode(t, js)} {
				if got.exitCode != 0 || len(got.stdout) == 0 {
					t.Fatalf("mutant must lose named refusal: %#v", got)
				}
				t.Logf("pinned finite-key control catches %s: exit=%d stdout=%q", mutant.name, got.exitCode, got.stdout)
			}
		})
	}
}

func TestCheckedViewDictionaryOptionalShortCircuitMutant(t *testing.T) {
	program, path := interfaceFixture(t, "dictionaries/source/optional-paths-receiver-missing")
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "missing\n" {
		t.Fatalf("Node: %#v", truth)
	}
	changed := false
	for i := range program.Functions {
		function := &program.Functions[i]
		if function.Name != "optional_dictionary_field" {
			continue
		}
		for j, statement := range function.Body {
			ret, ok := statement.(ir.Return)
			if !ok {
				continue
			}
			conditional, ok := ret.Value.(ir.Conditional)
			if !ok {
				continue
			}
			conditional.Condition = ir.BooleanConstant{Value: true}
			ret.Value = conditional
			function.Body[j] = ret
			changed = true
		}
	}
	if !changed {
		t.Fatal("no optional receiver branch to mutate")
	}
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{sanitized, onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || !strings.Contains(string(got.stderr), "view?.paths") {
			t.Fatalf("mutant must execute the forbidden absent-receiver read: %#v", got)
		}
		t.Logf("Node absent-receiver control catches forced read: exit=%d stderr=%q", got.exitCode, got.stderr)
	}
}

func TestCheckedViewDictionaryOptionalRequiredField(t *testing.T) {
	for _, name := range []string{"good", "missing", "receiver-missing", "undefined"} {
		t.Run(name, func(t *testing.T) {
			program, path := interfaceFixture(t, "dictionaries/source/optional-required-paths-"+name)
			truth := onNode(t, path)
			expected := "undefined\n"
			if name == "good" {
				expected = "object\n"
			}
			if truth.exitCode != 0 || string(truth.stdout) != expected {
				t.Fatalf("Node: %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			valid := name == "good" || name == "receiver-missing"
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if valid {
					if d := disagreement(truth, got); d != "" {
						t.Fatal(d)
					}
				} else if got.exitCode != 70 || viewReadDiagnosticMismatch("adamic: panic: field read failed: view?.paths; expected MapLike<string[]>, found undefined\n", got.stderr, program) {
					t.Fatalf("optional receiver must preserve required field contract: %#v", got)
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

func TestCheckedViewDictionaryEnumMap(t *testing.T) {
	for _, prefix := range []string{"enum-map-", "enum-map-producer-"} {
		for _, name := range []string{"good", "open-number", "wrong", "missing"} {
			t.Run(prefix+name, func(t *testing.T) {
				program, path := interfaceFixture(t, "dictionaries/source/"+prefix+name)
				truth := onNode(t, path)
				output := map[string]string{"good": "1\n", "open-number": "42\n", "wrong": "wrong\n", "missing": "undefined\n"}[name]
				if truth.exitCode != 0 || string(truth.stdout) != output {
					t.Fatalf("Node: %#v", truth)
				}
				sanitized, binary := nativelyUncached(t, program)
				for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if name != "wrong" {
						if d := disagreement(truth, got); d != "" {
							t.Fatal(d)
						}
					} else if got.exitCode != 70 || viewReadDiagnosticMismatch("adamic: panic: field read failed: view.table['item']; expected WatchDirectoryFlags | undefined, found string\n", got.stderr, program) {
						t.Fatalf("enum dictionary check: %#v", got)
					}
				}
				if name != "wrong" {
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}
