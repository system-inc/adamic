package oracle

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewDictionaryLookupBatch(t *testing.T) {
	for _, c := range []struct{ name, output, failure string }{
		{"map-string-good", "name\n", ""},
		{"map-string-missing", "undefined\n", ""},
		{"map-string-wrong", "42\n", "adamic: panic: field read failed: view['item']; expected string | undefined, found number\n"},
		{"map-generic-good", "name\n", ""},
		{"map-generic-missing", "undefined\n", ""},
		{"map-generic-wrong", "42\n", "adamic: panic: field read failed: table['item']; expected string | undefined, found number\n"},
		{"package-paths-good", "name\n", ""},
		{"package-paths-missing", "missing\n", ""},
		{"package-paths-wrong", "number\n", "adamic: panic: field read failed: view['imports']; expected MapLike<string> | undefined, found number\n"},
		{"package-paths-nested-wrong", "42\n", "adamic: panic: field read failed: table['item']; expected string | undefined, found number\n"},
	} {
		for _, producer := range []bool{false, true} {
			c := c
			if producer {
				if strings.HasPrefix(c.name, "map-string-") {
					c.failure = strings.ReplaceAll(c.failure, "view[", "view.table[")
				}
				if strings.HasPrefix(c.name, "package-paths-") {
					c.failure = strings.ReplaceAll(c.failure, "view[", "paths[")
				}
				for _, prefix := range []string{"map-string-", "map-generic-", "package-paths-"} {
					if strings.HasPrefix(c.name, prefix) {
						c.name = prefix + "producer-" + strings.TrimPrefix(c.name, prefix)
						break
					}
				}
			}
			t.Run(c.name, func(t *testing.T) {
				program, path := interfaceFixture(t, "dictionaries/source/"+c.name)
				truth := onNode(t, path)
				if truth.exitCode != 0 || string(truth.stdout) != c.output {
					t.Fatalf("Node: %#v", truth)
				}
				sanitized, binary := nativelyUncached(t, program)
				for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if c.failure == "" {
						if d := disagreement(truth, got); d != "" {
							t.Fatal(d)
						}
					} else if got.exitCode != 70 || viewReadDiagnosticMismatch(c.failure, got.stderr, program) {
						t.Fatalf("lookup contract: %#v; want %q", got, c.failure)
					}
				}
				if c.failure == "" {
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}

// Removing the erased-source view origin must lose the named read contract.
func TestCheckedViewDictionaryLookupOriginMutant(t *testing.T) {
	program, _ := interfaceFixture(t, "dictionaries/source/map-string-wrong")
	if len(program.ViewOrigins) == 0 {
		t.Fatal("erased source was admitted without a checked view origin")
	}
	program.ViewOrigins = nil
	binary := filepath.Join(t.TempDir(), "origin-mutant")
	if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
		t.Fatal("semantic mutant must compile: ", err)
	}
	got := execute(t, binary)
	want := "adamic: panic: record operation failed: source storage does not match the declared element representation\n"
	if got.exitCode != 70 || viewReadDiagnosticMismatch(want, got.stderr, program) {
		t.Fatalf("native mutant must lose the named field refusal: %#v", got)
	}
	t.Logf("origin mutant caught by named control: native loses field contract and reaches ordinary storage refusal")
	js := filepath.Join(t.TempDir(), "origin-mutant.mjs")
	if err := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	got = onNode(t, js)
	if got.exitCode != 0 || string(got.stdout) != "42\n" {
		t.Fatalf("JS mutant must lose the field refusal: %#v", got)
	}
	t.Logf("origin mutant caught by named control: JavaScript exit=%d stdout=%q", got.exitCode, got.stdout)
}
