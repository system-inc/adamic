package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewCallableScalarWitnesses(t *testing.T) {
	for _, family := range []struct{ name, field, expected, good, arity, wrongResult string }{
		{"scanner", "scanner.hasPrecedingLineBreak", "() => boolean", "true\n", "1", "7\n"},
		{"performance", "performance.mark", "(markName: string) => void", "beginTracing\n", "0", ""},
	} {
		for _, name := range []string{"good", "wrong-value", "wrong-arity", "wrong-result"} {
			t.Run(family.name+"/"+name, func(t *testing.T) {
				program, path := interfaceFixture(t, "lane5/"+family.name+"/"+name)
				truth := onNode(t, path)
				stdout := family.good
				if name == "wrong-arity" {
					stdout = ""
					if family.name == "scanner" {
						stdout = "false\n"
					}
				}
				if name == "wrong-result" {
					stdout = family.wrongResult
				}
				if name == "wrong-value" {
					if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "TypeError:") {
						t.Fatalf("Node: %#v", truth)
					}
				} else if truth.exitCode != 0 || string(truth.stdout) != stdout {
					t.Fatalf("Node: %#v", truth)
				}
				sanitized, binary := nativelyUncached(t, program)
				for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
					if name == "good" {
						if difference := disagreement(truth, got); difference != "" {
							t.Fatal(difference)
						}
						continue
					}
					found := "number"
					if name == "wrong-arity" {
						found = "function with arity " + family.arity
					}
					if name == "wrong-result" {
						found = "function with incompatible result representation"
					}
					expected := "adamic: panic: field read failed: " + family.field + " expected " + family.expected + ", found " + found + "\n"
					if name == "wrong-value" && index < 2 {
						expected = "adamic: panic: field read failed: " + family.field + " is not a " + family.expected + "; expected " + family.expected + ", found number\n"
					}
					if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != expected {
						t.Fatalf("backend %d: %#v want %q", index, got, expected)
					}
				}
				if name == "good" {
					if report := leaksUncached(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}

func TestCheckedViewCallableMixedResultBoundary(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/lane5/gaps/mixed-scalar-boxed-result.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "5\n" {
		t.Fatalf("Node: %#v", truth)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "a function value returning union of differently held members") {
		t.Fatalf("mixed result boundary: %v", err)
	}
}
