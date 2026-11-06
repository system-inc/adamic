package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func TestViewArrays(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"tagged-null-read", "tagged-null-lie", "guard-forwarded", "guard-ref-push", "guard-index-write", "guard-mutable-param", "guard-slice", "guard-spread", "guard-from", "unknown-null", "legacy-null", "generic-object-lie", "identity", "cast-unknown", "cast-object", "cast-union", "storage-lie", "renamed-lie", "unknown-scalar", "find-result-lie", "findLast-result-lie", "join", "json", "string", "index-good", "index-lie", "try-lie", "optional-element", "at-lie", "per-read", "bounds", "unknown-wide", "object-wide", "generic", "no-element", "for-of-good", "for-of-lie", "values-lie", "forEach-good", "forEach-lie", "map-good", "map-lie", "filter-good", "filter-lie", "filter-predicate-good", "filter-predicate-lie", "narrowing-cast", "some-good", "some-lie", "every-good", "every-lie", "find-good", "find-lie", "findIndex-good", "findIndex-lie", "findLast-good", "findLast-lie", "findLastIndex-good", "findLastIndex-lie", "reduce-good", "reduce-lie", "mutable", "slice", "spread", "concat"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs("testdata/views_v3/" + name + ".ts")
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 {
				t.Fatalf("source Node failed: %#v", node)
			}
			program, err := lowered(t, path)
			// A narrowing cast of a viewed array is no road to an unchecked Identifier read. To a
			// mutable array it's the writable cast NotYet (cast-object and its twins), and to a
			// readonly one no run-time check can prove it, so it's refused.
			if name == "narrowing-cast" {
				var refused *lower.Refused
				if !errors.As(err, &refused) || refused.What != "a cast the runtime can't check" {
					t.Fatalf("want the uncheckable cast refusal, got %v", err)
				}
				return
			}
			if name == "guard-mutable-param" {
				var refused *lower.Refused
				if !errors.As(err, &refused) || !strings.Contains(refused.Error(), "adamic/invariant-mutable") {
					t.Fatalf("want existing mutable-storage invariant, got %v", err)
				}
				return
			}
			if strings.HasPrefix(name, "guard-") || name == "unknown-null" || name == "legacy-null" || strings.HasPrefix(name, "cast-") || name == "join" || name == "json" || name == "string" || name == "mutable" || name == "slice" || name == "spread" || name == "concat" {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) {
					t.Fatalf("want NotYet at reader, got %v", err)
				}
				if name == "unknown-null" && !strings.Contains(notYet.What, "legacy nullable-pointer storage") {
					t.Fatal(notYet)
				}
				if strings.HasPrefix(name, "cast-") && notYet.What != "a cast that makes a viewed readonly array writable" {
					t.Fatal(notYet)
				}
				if name == "mutable" && !strings.Contains(notYet.What, "a checked view of mutable array field items") {
					t.Fatal(notYet)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			lying := strings.HasSuffix(name, "-lie") || name == "per-read" || name == "object-wide" || name == "generic"
			sanitized, binary := nativelyUncached(t, program)
			backends := map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)}
			// A failed check is a panic, never a throw: the program's own catch can't take it on any
			// backend, so each one exits 70 with the same words and none prints 'caught'.
			if name == "try-lie" {
				for backend, got := range backends {
					if got.exitCode != 70 || strings.Contains(string(got.stdout), "caught") || string(got.stderr) != string(backends["native"].stderr) {
						t.Fatalf("%s: a failed element check inside a try with a catch: %#v; native stderr %q", backend, got, backends["native"].stderr)
					}
				}
			}
			for backend, got := range backends {
				if !lying {
					if diff := disagreement(node, got); diff != "" {
						t.Fatalf("%s: %s; %#v", backend, diff, got)
					}
				} else {
					prefix := "before\n"
					if name == "per-read" {
						prefix = "first\n"
					}
					found := "; found number"
					if name == "tagged-null-lie" {
						found = "; found null"
					}
					if strings.Contains(name, "result-lie") || name == "generic-object-lie" {
						found = "; found object"
					}
					if got.exitCode != 70 || string(got.stdout) != prefix || !strings.HasPrefix(string(got.stderr), "adamic: panic: element read failed: ") || !strings.Contains(string(got.stderr), found) {
						t.Fatalf("%s: wanted first-read failure before handing out the lie; got %#v", backend, got)
					}
					expected := ""
					if name == "generic" || name == "generic-object-lie" {
						expected = "Node"
					}
					if name == "renamed-lie" {
						expected = "Token"
					}
					if name == "object-wide" {
						expected = "object"
					}
					if expected != "" && !strings.Contains(string(got.stderr), " is not a "+expected+";") {
						t.Fatalf("%s: wrong instantiated element type in %q", backend, got.stderr)
					}
					if !strings.HasPrefix(string(node.stdout), prefix) {
						t.Fatal("panic stdout is not Node's prefix")
					}
				}
			}
			if !lying {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

// The language prelude takes strings only, so this whole-array reader stops
// before IR exists. Keep Node's success and the earlier checker stop explicit.
func TestViewArrayConsole(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("testdata/views_v3/console.ts")
	if err != nil {
		t.Fatal(err)
	}
	if got := executeWith(t, nil, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path); got.exitCode != 0 {
		t.Fatalf("source Node: %#v", got)
	}
	_, err = load.Load([]string{path})
	var rejected *load.CheckError
	if !errors.As(err, &rejected) || !strings.Contains(err.Error(), "TS2345") {
		t.Fatalf("want prelude's string-only console check, got %v", err)
	}
}
