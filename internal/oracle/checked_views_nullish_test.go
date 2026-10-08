package oracle

import (
	"strings"
	"testing"
)

func TestCheckedViewNullish(t *testing.T) {
	for _, representation := range []string{"number", "boolean", "string", "object", "array"} {
		for _, variant := range []string{"null", "undefined", "both", "coalesce"} {
			name := representation + "-" + variant
			t.Run(name, func(t *testing.T) {
				program, path := interfaceFixture(t, "nullish/fixtures/"+name)
				truth := onNode(t, path)
				if truth.exitCode != 0 {
					t.Fatalf("Node source: %#v", truth)
				}
				checked, binary := nativelyUncached(t, program)
				for _, got := range []run{checked, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatal(difference)
					}
				}
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
				t.Logf("Node and both backends: %q", truth.stdout)
			})
		}
	}
}

func TestCheckedViewNullishMutants(t *testing.T) {
	for _, representation := range []string{"number", "boolean", "string", "object", "array"} {
		for _, variant := range []string{"null", "undefined", "both"} {
			mutations := []string{"wrong", "missing"}
			if variant != "both" {
				mutations = append(mutations, "opposite")
			}
			for _, mutation := range mutations {
				name := representation + "-" + variant + "-" + mutation
				t.Run(name, func(t *testing.T) {
					program, path := interfaceFixture(t, "nullish/fixtures/"+name)
					truth := onNode(t, path)
					t.Logf("source Node: exit=%d stdout=%q stderr=%q", truth.exitCode, truth.stdout, truth.stderr)
					checked, _ := nativelyUncached(t, program)
					for _, got := range []run{checked, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if got.exitCode != 70 || !strings.Contains(string(got.stderr), "node.value") || !strings.Contains(string(got.stderr), "expected") || !strings.Contains(string(got.stderr), "found") {
							t.Fatalf("nullish mutant ran on: %#v", got)
						}
						t.Logf("caught %s: %s", mutation, got.stderr)
					}
				})
			}
		}
	}
}

func TestCheckedViewNullishTransitiveMutants(t *testing.T) {
	for _, representation := range []string{"object", "array"} {
		t.Run(representation, func(t *testing.T) {
			program, path := interfaceFixture(t, "nullish/fixtures/"+representation+"-both-nested")
			truth := onNode(t, path)
			t.Logf("Node: exit=%d stdout=%q", truth.exitCode, truth.stdout)
			sanitized, _ := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if got.exitCode != 70 || !strings.Contains(string(got.stderr), "string") {
					t.Fatalf("nested read ran on: %#v", got)
				}
				t.Logf("caught nested read: %s", got.stderr)
			}
		})
	}
}

func TestCheckedViewNullishRegexIdentity(t *testing.T) {
	program, path := interfaceFixture(t, "nullish/fixtures/regex-identity")
	truth := onNode(t, path)
	checked, binary := nativelyUncached(t, program)
	for _, got := range []run{checked, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatal(difference)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Logf("Node and both backends: %q", truth.stdout)
}
