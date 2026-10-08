package oracle

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Admission of a dictionary container must not demand its unread rich elements.
func TestCheckedViewDictionaryCommandOptions(t *testing.T) {
	for _, name := range []string{"good", "wrong", "missing"} {
		t.Run(name, func(t *testing.T) {
			program, path := interfaceFixture(t, "dictionaries/source/command-options-"+name)
			truth := onNode(t, path)
			expected := map[string]string{"good": "object\n", "wrong": "number\n", "missing": "undefined\n"}[name]
			if truth.exitCode != 0 || string(truth.stdout) != expected {
				t.Fatalf("Node: %#v", truth)
			}
			sanitized, _ := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if name == "good" {
					if d := disagreement(truth, got); d != "" {
						t.Fatal(d)
					}
					continue
				}
				found := map[string]string{"wrong": "number", "missing": "missing"}[name]
				want := "adamic: panic: field read failed: view.options is not a CompilerOptions; expected CompilerOptions, found " + found + "\n"
				if name == "missing" {
					want = "adamic: panic: field read failed: view.options is not initialized; expected CompilerOptions, found missing\n"
				}
				if got.exitCode != 70 || string(got.stderr) != want {
					t.Fatalf("container check: %#v; want %q", got, want)
				}
			}
		})
	}
}

func TestCheckedViewDictionaryContainerBatch(t *testing.T) {
	for _, shape := range []struct {
		fixture, field, declared string
		optional                 bool
	}{
		{"wildcard-directories", "wildcardDirectories", "MapLike<WatchDirectoryFlags> | undefined", true},
		{"watch-options", "watchOptions", "WatchOptions" + " | undefined", true},
		{"version-paths", "paths", "MapLike<string[]>", false},
		{"incremental-multi-options", "options", "CompilerOptions" + " | undefined", true},
		{"incremental-bundle-options", "options", "CompilerOptions" + " | undefined", true},
		{"incremental-options", "options", "CompilerOptions" + " | undefined", true},
		{"reusable-state-options", "compilerOptions", "CompilerOptions", false},
		{"builder-state-options", "compilerOptions", "CompilerOptions", false},
	} {
		for _, name := range []string{"good", "wrong", "missing"} {
			t.Run(shape.fixture+"-"+name, func(t *testing.T) {
				program, path := interfaceFixture(t, "dictionaries/source/"+shape.fixture+"-"+name)
				truth := onNode(t, path)
				expected := map[string]string{"good": "object\n", "wrong": "number\n", "missing": "undefined\n"}[name]
				if truth.exitCode != 0 || string(truth.stdout) != expected {
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
					want := "adamic: panic: field read failed: view." + shape.field + " is not a " + shape.declared + "; expected " + shape.declared + ", found number\n"
					if name == "missing" {
						want = "adamic: panic: field read failed: view." + shape.field + " is not initialized; expected " + shape.declared + ", found missing\n"
					}
					if got.exitCode != 70 || string(got.stderr) != want {
						t.Fatalf("pinned container check: %#v; want %q", got, want)
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

// The parent-container refusal must itself detect a semantic bypass.
func TestCheckedViewDictionaryContainerMutant(t *testing.T) {
	program, _ := interfaceFixture(t, "dictionaries/source/command-options-wrong")
	c := native.C(program)
	if !strings.Contains(c, "adamic_object_view(") {
		t.Fatal("no native container check to mutate")
	}
	wrapper := `#include "adamic.h"
#include <string.h>
static adamic_value dictionary_container_mutant(const adamic_object *object, const char *name, adamic_slot_cache *cache, unsigned char wanted, const char *type, const char *expression) {
 if (strcmp(name,"options") == 0) return (adamic_value){.reference=(void *)object};
 return adamic_object_view(object,name,cache,wanted,type,expression);
}
`
	c = wrapper + strings.ReplaceAll(c, "adamic_object_view(", "dictionary_container_mutant(")
	binary := filepath.Join(t.TempDir(), "container-mutant")
	if err := native.Build(c, binary, native.Options{}); err != nil {
		t.Fatal("semantic mutant must compile: ", err)
	}
	code := javascript.JavaScript(program)
	if !strings.Contains(code, "adamicViewField(") {
		t.Fatal("no JavaScript container check to mutate")
	}
	code = "const dictionaryContainerMutant = (object,key) => object[key];\n" + strings.ReplaceAll(code, "adamicViewField(", "dictionaryContainerMutant(")
	js := filepath.Join(t.TempDir(), "container-mutant.mjs")
	if err := os.WriteFile(js, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	for _, got := range []run{execute(t, binary), onNode(t, js)} {
		if got.exitCode != 0 || len(got.stdout) == 0 {
			t.Fatalf("mutant must lose named container refusal: %#v", got)
		}
		t.Logf("container bypass caught by pinned wrong-value control: exit=%d stdout=%q", got.exitCode, got.stdout)
	}
}
