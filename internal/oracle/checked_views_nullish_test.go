package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewNullish(t *testing.T) {
	for _, representation := range []string{"number", "boolean", "string", "object", "array", "callable"} {
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
	for _, representation := range []string{"number", "boolean", "string", "object", "array", "callable"} {
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

func TestCheckedViewNullishCallableSignatureMutant(t *testing.T) {
	path := filepath.Join(repository, "stage3/interface-downcasts/nullish/fixtures/callable-signature-mutant.a")
	path, pathErr := filepath.Abs(checkedViewFixturePath(path))
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	program, err := lowered(t, path)
	truth := onNode(t, path)
	if truth.exitCode != 0 || !strings.Contains(string(truth.stdout), "42\n") {
		t.Fatalf("Node control: %#v", truth)
	}
	if err != nil {
		t.Fatal(err)
	}
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
		expected := "adamic: panic: cast failed: field read failed: node.value expected () => string, found function with incompatible result representation\n"
		if got.exitCode != 70 || string(got.stdout) != "function:false:false\nordinary\n" || string(got.stderr) != expected {
			t.Fatalf("nullable signature: %#v want %q", got, expected)
		}
	}

}

func TestCheckedViewNullishMapUnread(t *testing.T) {
	for _, variant := range []string{"null", "undefined", "both"} {
		t.Run(variant, func(t *testing.T) {
			program, path := interfaceFixture(t, "nullish/fixtures/map-"+variant+"-unread")
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
			t.Logf("unread map admitted, Node and both backends: %q", truth.stdout)
		})
	}
}
func TestCheckedViewNullishMapReads(t *testing.T) {
	for _, variant := range []string{"null", "undefined", "both"} {
		for _, mutation := range []string{"", "-wrong", "-opposite"} {
			if variant == "both" && mutation == "-opposite" {
				continue
			}
			t.Run(variant+mutation, func(t *testing.T) {
				program, path := interfaceFixture(t, "nullish/fixtures/map-"+variant+mutation)
				truth := onNode(t, path)
				native, binary := nativelyUncached(t, program)
				for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if mutation == "" {
						if diff := disagreement(truth, got); diff != "" {
							t.Fatal(diff)
						}
					} else if got.exitCode != 70 || !strings.Contains(string(got.stderr), "node.value") {
						t.Fatalf("map read mutant ran on: %#v", got)
					}
				}
				if mutation == "" {
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
				t.Logf("Node exit=%d stdout=%q, mutation=%q", truth.exitCode, truth.stdout, mutation)
			})
		}
	}
}

func TestCheckedViewNullishLiterals(t *testing.T) {
	for _, mutation := range []string{"", "-wrong"} {
		t.Run("literal"+mutation, func(t *testing.T) {
			program, path := interfaceFixture(t, "nullish/fixtures/literal-both"+mutation)
			truth := onNode(t, path)
			checked, binary := nativelyUncached(t, program)
			for _, got := range []run{checked, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if mutation == "" {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatal(difference)
					}
				} else if got.exitCode != 70 || !strings.Contains(string(got.stderr), "node.value") || !strings.Contains(string(got.stderr), "made-here") {
					t.Fatalf("finite nullish literal mutant ran on: %#v", got)
				}
			}
			if mutation == "" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			t.Logf("Node: exit=%d stdout=%q; checked literal variant %q", truth.exitCode, truth.stdout, mutation)
		})
	}
}
