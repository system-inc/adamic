package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestCheckedViewDictionaryKeys(t *testing.T) {
	for _, shape := range []string{"fixed", "producer"} {
		for _, name := range []string{"good", "empty", "read-wrong"} {
			t.Run(shape+"-"+name, func(t *testing.T) {
				program, path := interfaceFixture(t, "dictionaries/source/keys-"+shape+"-"+name)
				truth := onNode(t, path)
				want := "2\n10\nz\na\n"
				if name == "empty" {
					want = ""
				}
				if name == "read-wrong" {
					want += "42\n"
				}
				if truth.exitCode != 0 || string(truth.stdout) != want {
					t.Fatalf("Node: %#v", truth)
				}
				sanitized, binary := nativelyUncached(t, program)
				for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if name == "read-wrong" {
						diagnostic := "adamic: panic: field read failed: view.table['z']; expected string | undefined, found number\n"
						if got.exitCode != 70 || viewReadDiagnosticMismatch(diagnostic, got.stderr, program) || string(got.stdout) != "2\n10\nz\na\n" {
							t.Fatalf("lazy enumeration/lookup: %#v", got)
						}
					} else if d := disagreement(truth, got); d != "" {
						t.Fatalf("%s: got %#v truth %#v", d, got, truth)
					}
				}
				if name != "read-wrong" {
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}

func TestCheckedViewDictionaryKeysEagerMutant(t *testing.T) {
	program, path := interfaceFixture(t, "dictionaries/source/keys-producer-good")
	truth := onNode(t, path)
	c := native.C(program)
	call := "adamic_view_dictionary_source_keys("
	if !strings.Contains(c, call) {
		t.Fatal("no keys hook")
	}
	wrapper := `static adamic_array *keys_mutant(const adamic_object *object) {
 adamic_array *keys=adamic_view_dictionary_source_keys(object);
 for(size_t i=0;i<keys->length;i++) (void)adamic_view_dictionary_source_read(object,keys->elements[i].reference,1u << adamic_view_union_string,0,"Object.keys[key]","string");
 return keys;
 }`
	c = "#include \"view_dictionaries.h\"\n" + wrapper + "\n" + strings.ReplaceAll(c, call, "keys_mutant(")
	binary := filepath.Join(t.TempDir(), "keys-mutant")
	if err := native.Build(c, binary, native.Options{}); err != nil {
		t.Fatal("semantic mutant must compile: ", err)
	}
	code := javascript.JavaScript(program)
	if !strings.Contains(code, "Object.keys(") {
		t.Fatal("no JS keys hook")
	}
	js := filepath.Join(t.TempDir(), "keys-mutant.mjs")
	code = `const keysMutant=object=>{const keys=Object.keys(object);for(const key of keys)adamicViewDictionaryRead(object,key,['string'],0,'Object.keys[key]','string');return keys;};` + "\n" + strings.ReplaceAll(code, "Object.keys(", "keysMutant(")
	if err := os.WriteFile(js, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	for _, got := range []run{execute(t, binary), onNode(t, js)} {
		if got.exitCode != 70 || disagreement(truth, got) == "" {
			t.Fatalf("eager element check escaped: %#v", got)
		}
		t.Logf("key-only control catches eager element mutant: exit=%d stderr=%q", got.exitCode, got.stderr)
	}
}

func TestCheckedViewDictionaryValuesEntries(t *testing.T) {
	for _, method := range []string{"values", "entries"} {
		for _, storage := range []string{"fixed", "producer"} {
			for _, kind := range []string{"string", "number", "object", "array"} {
				for _, name := range []string{"good", "wrong", "empty"} {
					t.Run(method+"-"+storage+"-"+kind+"-"+name, func(t *testing.T) {
						if method == "entries" && kind == "array" {
							path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/dictionaries/source/"+method+"-"+storage+"-"+kind+"-"+name+".a"))
							if err != nil {
								t.Fatal(err)
							}
							truth := onNode(t, path)
							if truth.exitCode != 0 {
								t.Fatalf("Node: %#v", truth)
							}
							_, err = lowered(t, path)
							var refusal *lower.Refused
							if !errors.As(err, &refusal) || !strings.Contains(refusal.What, "unsupported tuple contract") || !strings.Contains(refusal.What, "[element]") {
								t.Fatalf("named demanded tuple frontier: %v", err)
							}
							t.Logf("uncredited array-entry consumer remains a named tuple refusal: %v", err)
							return
						}
						program, path := interfaceFixture(t, "dictionaries/source/"+method+"-"+storage+"-"+kind+"-"+name)
						truth := onNode(t, path)
						output := "1\nname\n"
						if kind == "number" {
							output = "1\n42\n"
						}
						if name == "empty" {
							output = "0\n"
						}
						if name == "wrong" {
							output = "1\n42\n"
							if kind == "number" {
								output = "1\nwrong\n"
							}
						}
						if truth.exitCode != 0 || string(truth.stdout) != output {
							t.Fatalf("Node: %#v", truth)
						}
						sanitized, binary := nativelyUncached(t, program)
						for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
							if name == "wrong" {
								expected, found := "string", "number"
								if kind == "number" {
									expected, found = "number", "string"
								}
								diagnostic := "adamic: panic: field read failed: Object." + method + "(view.table)[key]; expected " + expected + ", found " + found + "\n"
								prefix := ""
								if kind == "object" {
									diagnostic = "adamic: panic: field read failed: value.name is not a string; expected string, found number\n"
									prefix = "1\n"
								}
								if kind == "array" {
									diagnostic = "adamic: panic: element read failed: value[0] expected string, found number\n"
									prefix = "1\n"
								}
								if got.exitCode != 70 || viewReadDiagnosticMismatch(diagnostic, got.stderr, program) || string(got.stdout) != prefix {
									t.Fatalf("named enumeration refusal: %#v, want %q stdout %q", got, diagnostic, prefix)
								}
							} else if d := disagreement(truth, got); d != "" {
								t.Fatalf("%s: got %#v truth %#v", d, got, truth)
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
	}
}

func TestCheckedViewDictionaryEnumerationMutants(t *testing.T) {
	for _, mutant := range []struct{ name, fixture, call, wrapper, jsCall, jsWrapper string }{
		{"skip-element-check", "values-check-wrong-number", "adamic_view_dictionary_source_read(", `static adamic_view_dictionary_result enumeration_mutant(const adamic_object *object,const adamic_string *key,unsigned int kinds,size_t child,const char *expression,const char *declared) { (void)kinds;adamic_view_dictionary_result result=adamic_view_dictionary_source_read(object,key,~0u,child,expression,declared);if(result.value.kind==adamic_view_union_number){result.value.payload.reference=adamic_string_from_number(result.value.payload.number);result.value.kind=adamic_view_union_string;}return result;}`, "adamicViewDictionaryRead(", `const enumerationMutant=(object,key)=>object[key];`},
		{"accept-wrong-shape", "values-check-wrong-array", "adamic_view_dictionary_source_read(", `static adamic_view_dictionary_result enumeration_mutant(const adamic_object *object,const adamic_string *key,unsigned int kinds,size_t child,const char *expression,const char *declared) {return adamic_view_dictionary_source_read(object,key,kinds|(1u<<adamic_view_union_array),child,expression,declared);}`, "adamicViewDictionaryRead(", `const enumerationMutant=(object,key,kinds,child,expression,declared)=>adamicViewDictionaryRead(object,key,[...kinds,'array'],child,expression,declared);`},
		{"drop-transitive-check", "values-fixed-object-wrong", "adamic_object_view(", `static adamic_value enumeration_mutant(const adamic_object *object,const char *name,adamic_slot_cache *cache,unsigned char wanted,const char *type,const char *expression) {adamic_value *slot=adamic_object_read(object,name,cache,expression);if(wanted==3 && adamic_object_field_types(object)[cache->index]==1)return (adamic_value){.reference=adamic_string_from_number(slot->number)};return adamic_object_view(object,name,cache,wanted,type,expression);}`, "adamicViewField(", `const enumerationMutant=(object,key)=>object[key];`},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "dictionaries/source/"+mutant.fixture)
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node: %#v", truth)
			}
			want := "adamic: panic: field read failed: Object.values(view.table)[key]; expected string, found number\n"
			prefix := ""
			if mutant.name == "accept-wrong-shape" {
				want = "adamic: panic: field read failed: Object.values(view.table)[key]; expected Entry, found array\n"
			}
			if mutant.name == "drop-transitive-check" {
				want = "adamic: panic: field read failed: value.name is not a string; expected string, found number\n"
				prefix = "1\n"
			}
			for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if got.exitCode != 70 || viewReadDiagnosticMismatch(want, got.stderr, program) || string(got.stdout) != prefix {
					t.Fatalf("pinned control: %#v; want %q", got, want)
				}
			}
			c := native.C(program)
			if !strings.Contains(c, mutant.call) {
				t.Fatal("no C mutation target")
			}
			c = "#include \"view_dictionaries.h\"\n" + mutant.wrapper + "\n" + strings.ReplaceAll(c, mutant.call, "enumeration_mutant(")
			binary := filepath.Join(t.TempDir(), "enumeration-mutant")
			if err := native.Build(c, binary, native.Options{}); err != nil {
				t.Fatal("semantic mutant must compile: ", err)
			}
			code := javascript.JavaScript(program)
			if !strings.Contains(code, mutant.jsCall) {
				t.Fatal("no JS mutation target")
			}
			code = mutant.jsWrapper + "\n" + strings.ReplaceAll(code, mutant.jsCall, "enumerationMutant(")
			js := filepath.Join(t.TempDir(), "enumeration-mutant.mjs")
			if err := os.WriteFile(js, []byte(code), 0600); err != nil {
				t.Fatal(err)
			}
			for _, got := range []run{execute(t, binary), onNode(t, js)} {
				if got.exitCode != 0 || disagreement(truth, got) != "" {
					t.Fatalf("semantic mutant must execute wrong checked program: %#v truth %#v", got, truth)
				}
				t.Logf("pinned enumeration control catches %s: exit=%d stdout=%q", mutant.name, got.exitCode, got.stdout)
			}
		})
	}
}

func TestCheckedViewDictionaryArrayEntryCreationIsLazy(t *testing.T) {
	for _, storage := range []string{"fixed", "producer"} {
		for _, name := range []string{"good", "wrong", "empty"} {
			t.Run(storage+"-"+name, func(t *testing.T) {
				program, path := interfaceFixture(t, "dictionaries/source/entries-"+storage+"-array-unread-"+name)
				truth := onNode(t, path)
				expected := "1\n"
				if name == "empty" {
					expected = "0\n"
				}
				if truth.exitCode != 0 || string(truth.stdout) != expected {
					t.Fatalf("Node: %#v", truth)
				}
				sanitized, binary := nativelyUncached(t, program)
				for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if d := disagreement(truth, got); d != "" {
						t.Fatalf("%s: %#v", d, got)
					}
				}
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			})
		}
	}
}

func TestCheckedViewDictionaryEnumerationRecheck(t *testing.T) {
	for _, method := range []string{"values", "entries"} {
		t.Run(method, func(t *testing.T) {
			program, path := interfaceFixture(t, "dictionaries/recheck/"+method)
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != "1\n" {
				t.Fatalf("Node: %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if d := disagreement(truth, got); d != "" {
					t.Fatalf("%s: %#v", d, got)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}
