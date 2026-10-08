package oracle

import (
	"errors"

	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewCallableSourceContracts(t *testing.T) {
	for _, probe := range []struct{ name, stdout, found string }{
		{"number-good", "8\n", ""},
		{"number-wrong-value", "", "number"},
		{"number-wrong-arity", "", "function with arity 0"},
		{"optional-absent", "undefined\n", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/probes/"+probe.name)
			truth := onNode(t, path)
			if probe.found == "" && (truth.exitCode != 0 || string(truth.stdout) != probe.stdout) {
				t.Fatalf("Node %#v", truth)
			}
			if probe.name == "number-wrong-arity" && (truth.exitCode != 0 || string(truth.stdout) != "7\n") {
				t.Fatalf("Node arity control %#v", truth)
			}
			if probe.name == "number-wrong-value" && (truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "TypeError:") || !strings.Contains(string(truth.stderr), "is not a function")) {
				t.Fatalf("Node non-callable control %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
				if probe.found == "" {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatal(difference)
					}
					continue
				}
				message := "adamic: panic: field read failed: read(input).run expected (value: number) => number, found " + probe.found + "\n"
				if probe.found == "number" && index < 2 {
					message = "adamic: panic: field read failed: read(input).run is not a (value: number) => number; expected (value: number) => number, found number\n"
				}
				if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
					t.Fatalf("backend%d got %#v want %q", index, got, message)
				}
			}
			if probe.found == "" {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

func TestCheckedViewCallableDirectoryPairs(t *testing.T) {
	for _, field := range []string{"getCurrentDirectory", "getCommonSourceDirectory"} {
		for _, probe := range []struct{ name, found string }{
			{"good", ""}, {"wrong-value", "number"}, {"wrong-arity", "function with arity 1"},
			{"wrong-result", "function with incompatible result representation"},
			{"generic", ""}, {"callback", ""}, {"stored", ""}, {"direct-good", ""}, {"direct-arity", "function with arity 1"},
		} {
			t.Run(field+"/"+probe.name, func(t *testing.T) {
				program, path := interfaceFixture(t, "lane5/group1/"+field+"-"+probe.name)
				truth := onNode(t, path)
				if probe.found == "" && (truth.exitCode != 0 || string(truth.stdout) != "/here\n") {
					t.Fatalf("Node %#v", truth)
				}
				if probe.name == "wrong-value" && (truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "TypeError:") || !strings.Contains(string(truth.stderr), "is not a function")) {
					t.Fatalf("Node non-callable control %#v", truth)
				}
				if probe.name == "wrong-result" && (truth.exitCode != 0 || string(truth.stdout) != "7\n") {
					t.Fatalf("Node result control %#v", truth)
				}
				// Node permits arity/result mismatches; the checked-view contract is stronger.
				if strings.Contains(probe.name, "arity") && (truth.exitCode != 0 || string(truth.stdout) != "/bad\n") {
					t.Fatalf("Node arity control %#v", truth)
				}
				sanitized, binary := nativelyUncached(t, program)
				for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
					if probe.found == "" {
						if difference := disagreement(truth, got); difference != "" {
							t.Fatal(difference)
						}
						continue
					}
					message := "adamic: panic: field read failed: p." + field + " expected () => string, found " + probe.found + "\n"
					if probe.found == "number" && index < 2 {
						message = "adamic: panic: field read failed: p." + field + " is not a () => string; expected () => string, found number\n"
					}
					if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
						t.Fatalf("backend%d got %#v want %q", index, got, message)
					}
				}
				if probe.found == "" {
					if report := leaksUncached(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}

func TestCheckedViewCallableMarker(t *testing.T) {
	for _, name := range []string{"parser-cache", "discarded", "discarded-wrong-arity", "write-back", "discarded-required", "stored-call"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(checkedViewFixturePath(filepath.Join(repository, "stage3/interface-downcasts/lane5/marker", name+".a")))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node %#v", truth)
			}
			program, err := lowered(t, path)
			if name == "write-back" {
				if string(truth.stdout) != "undefined\n" {
					t.Fatalf("Node write-back %#v", truth)
				}
				if err == nil || !strings.Contains(err.Error(), "adamic/invariant-mutable") || !strings.Contains(err.Error(), "which can write AnyFunction where () => boolean is read") {
					t.Fatalf("write-back refusal: %v", err)
				}
				return
			}
			if name == "parser-cache" || name == "discarded-required" || name == "stored-call" {
				var refused *lower.Refused
				if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/invariant-mutable") || !strings.Contains(err.Error(), "write void where boolean is read") {
					t.Fatalf("want the ruled marker result refusal, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{releasedUncached(t, program), actual, onJavaScriptBackend(t, program)} {
				if name == "discarded-wrong-arity" {
					expected := "adamic: panic: field read failed: debug.shouldLog expected () => boolean, found function with arity 1\n"
					if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != expected {
						t.Fatalf("wrong arity %#v expected %q", got, expected)
					}
				} else if difference := disagreement(truth, got); difference != "" {
					t.Fatal(difference)
				}
			}
			if name != "discarded-wrong-arity" && name != "discarded-required" && name != "stored-call" {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

func TestCheckedViewCallableMethods(t *testing.T) {
	for _, probe := range []struct{ name, stdout, found string }{
		{"method-good", "10\n", ""},
		{"method-wrong-arity", "7\n", "function with arity 0"},
		{"method-wrong-result", "/bad\n", "function with incompatible result representation"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/group1/"+probe.name)
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != probe.stdout {
				t.Fatalf("Node %#v", truth)
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{releasedUncached(t, program), actual, onJavaScriptBackend(t, program)} {
				if probe.found == "" {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatal(difference)
					}
					continue
				}
				expected := "adamic: panic: field read failed: debug.run expected (value: number) => number, found " + probe.found + "\n"
				if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != expected {
					t.Fatalf("method %#v expected %q", got, expected)
				}
			}
			if probe.found == "" {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

func TestCheckedViewStoredMarkerCalls(t *testing.T) {
	for _, name := range []string{"boolean", "number", "string", "object", "array", "record", "void", "wrong-arity", "field-good", "observed"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(checkedViewFixturePath(filepath.Join(repository, "stage3/interface-downcasts/lane5/stored-marker", name+".a")))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node %#v", truth)
			}
			_, err = lowered(t, path)
			if name == "void" || name == "field-good" {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "a call through an erased never-rest callable marker") {
					t.Fatalf("want the ruled stored-marker call refusal, got %v", err)
				}
				return
			}
			var refused *lower.Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/invariant-mutable") || !strings.Contains(err.Error(), "write void where") {
				t.Fatalf("want the ruled marker result refusal, got %v", err)
			}
		})
	}
}
