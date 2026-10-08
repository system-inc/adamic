package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewCallableCanonicalWitness(t *testing.T) {
	for _, probe := range []struct{ name, found, node string }{
		{"good", "", "canonical/FILE\n"},
		{"wrong-value", "number", ""},
		{"wrong-arity", "function with arity 0", "/bad\n"},
		{"wrong-result", "function with incompatible result representation", "7\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/canonical/"+probe.name)
			truth := onNode(t, path)
			if probe.name == "wrong-value" {
				if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "TypeError:") {
					t.Fatalf("Node non-callable: %#v", truth)
				}
			} else if truth.exitCode != 0 || string(truth.stdout) != probe.node {
				t.Fatalf("Node: %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
				if probe.found == "" {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatal(difference)
					}
					continue
				}
				expected := "adamic: panic: cast failed: field read failed: newProgram.getCanonicalFileName expected GetCanonicalFileName, found " + probe.found + "\n"
				if probe.found == "number" && index < 2 {
					expected = "adamic: panic: cast failed: field read failed: newProgram.getCanonicalFileName is not a GetCanonicalFileName; expected GetCanonicalFileName, found number\n"
				}
				if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != expected {
					t.Fatalf("backend %d: %#v want %q", index, got, expected)
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

func TestCheckedViewCallableWatcherWitness(t *testing.T) {
	for _, probe := range []struct{ name, found string }{
		{"good", ""}, {"wrong-value", "number"}, {"wrong-arity", "function with arity 1"}, {"wrong-result", "function with incompatible result representation"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/watcher/"+probe.name)
			truth := onNode(t, path)
			stdout := "closed\n"
			if probe.name == "wrong-result" {
				stdout = ""
			}
			if probe.name == "wrong-value" {
				if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "TypeError:") {
					t.Fatalf("Node: %#v", truth)
				}
			} else if truth.exitCode != 0 || string(truth.stdout) != stdout {
				t.Fatalf("Node: %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
				if probe.found == "" {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatal(difference)
					}
					continue
				}
				expected := "adamic: panic: cast failed: field read failed: watcher.close expected () => void, found " + probe.found + "\n"
				if probe.found == "number" && index < 2 {
					expected = "adamic: panic: cast failed: field read failed: watcher.close is not a () => void; expected () => void, found number\n"
				}
				if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != expected {
					t.Fatalf("backend %d %#v want %q", index, got, expected)
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

func TestCheckedViewCallablePendingBoundaries(t *testing.T) {
	for _, probe := range []struct{ name, refusal string }{
		{"debug-assert-contract", "adamic/no-type-predicate"},
		{"intrinsic-member-read", "unbound-method"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(checkedViewFixturePath(filepath.Join(repository, "stage3/interface-downcasts/lane5/gaps", probe.name+".a")))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != "function\n" {
				t.Fatalf("Node: %#v", truth)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), probe.refusal) {
				t.Fatalf("pending boundary: %v", err)
			}
		})
	}
}

func TestCheckedViewCallableNullableArity(t *testing.T) {
	program, path := interfaceFixture(t, "lane5/probes/nullable-wrong-arity")
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "function\n/bad\n" {
		t.Fatalf("Node: %#v", truth)
	}
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
		expected := "adamic: panic: cast failed: field read failed: node.value expected () => string, found function with arity 1\n"
		if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != expected {
			t.Fatalf("nullable callable: %#v want %q", got, expected)
		}
	}
}
