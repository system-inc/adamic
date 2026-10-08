package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// Original binding reads remain refused; replacing them with direct reads would
// certify a different source expression and hide the method-detachment boundary.
func TestCheckedViewCallableOriginalBindingRefusals(t *testing.T) {
	for _, family := range []struct{ directory, out string }{{"variable-declaration", "9\n"}, {"variable-list", "9\n"}} {
		t.Run(family.directory, func(t *testing.T) {
			path, err := filepath.Abs(checkedViewFixturePath(repository + "/stage3/interface-downcasts/lane5/optional-aggregates/" + family.directory + "/good.a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != family.out {
				t.Fatalf("Node %#v", truth)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), "unsupported destructuring representation conversion") {
				t.Fatalf("original binding refusal: %v", err)
			}
		})
	}
}

func TestCheckedViewCallableOptionalAggregates(t *testing.T) {
	for _, family := range []struct{ directory, field, good, payload, payloadOut string }{
		{"parameter-declaration", "createParameterDeclaration", "3\n", "name.value", "3\n"},
		{"variable-statement", "createVariableStatement", "5\n", "modifiers[0]!.value", "9\n"},
	} {
		for _, variant := range []string{"good", "optional-values", "wrong-value", "wrong-arity", "wrong-result", "wrong-members", "wrong-parameter-payload"} {
			t.Run(family.directory+"/"+variant, func(t *testing.T) {
				program, path := interfaceFixture(t, "lane5/optional-aggregates/"+family.directory+"/"+variant)
				truth := onNode(t, path)
				out := family.good
				switch variant {
				case "optional-values":
					out = "3\n4\n6\n"
					if family.directory == "variable-statement" {
						out = "5\n7\n"
					}
				case "wrong-arity":
					out = "9\n"
				case "wrong-result":
					out = "undefined\n"
				case "wrong-parameter-payload":
					out = family.payloadOut
				}
				if variant == "wrong-value" {
					if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "TypeError:") {
						t.Fatalf("Node %#v", truth)
					}
				} else if truth.exitCode != 0 || string(truth.stdout) != out {
					t.Fatalf("Node %#v want %q", truth, out)
				}
				sanitized, binary := nativelyUncached(t, program)
				for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
					if variant == "good" || variant == "optional-values" {
						if difference := disagreement(truth, got); difference != "" {
							t.Fatal(difference)
						}
						continue
					}
					field, found := family.field, "found number"
					switch variant {
					case "wrong-arity":
						found = "found function with arity 1"
					case "wrong-members":
						found = "found function with incompatible parameter representations"
					case "wrong-result":
						found = "found function with incompatible result representation"
					case "wrong-parameter-payload":
						field = family.payload
						found = "expected string, found number"
					}
					if got.exitCode != 70 || len(got.stdout) != 0 || !strings.Contains(string(got.stderr), field) || !strings.Contains(string(got.stderr), found) {
						t.Fatalf("backend%d exit %d stdout %q stderr %q want %q and %q; sanitized exit %d stderr %q", index, got.exitCode, got.stdout, got.stderr, field, found, sanitized.exitCode, sanitized.stderr)
					}
				}
				if variant == "good" || variant == "optional-values" {
					if report := leaksUncached(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}

func TestCheckedViewCallableDestructuredSiblingRefusal(t *testing.T) {
	path, err := filepath.Abs(checkedViewFixturePath(repository + "/stage3/interface-downcasts/lane5/gaps/destructured-sibling.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "5\n" {
		t.Fatalf("Node %#v", truth)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "a function viewed as unknown or object (dynamic function descriptors)") {
		t.Fatalf("rest callable refusal: %v", err)
	}
}
