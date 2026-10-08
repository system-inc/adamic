package oracle

import (
	"strings"
	"testing"
)

func TestCheckedViewCallablePropertyWitnesses(t *testing.T) {
	for _, family := range []struct{ name, field, signature string }{
		{"property-access", "createPropertyAccessExpression", "(expression: Expression, name: string | Expression) => PropertyAccessExpression"},
		{"property-assignment", "createPropertyAssignment", "(name: string | Expression, initializer: Expression) => PropertyAssignment"},
	} {
		for _, probe := range []struct{ name, out, found string }{
			{"good", "8\n", ""}, {"boxed-argument", "9\n", ""},
			{"wrong-value", "", "number"}, {"wrong-arity", "9\n", "function with arity 1"},
			{"wrong-result", "undefined\n", "function with incompatible result representation"},
			{"wrong-members", "8\n", "function with incompatible parameter representations"},
		} {
			t.Run(family.name+"/"+probe.name, func(t *testing.T) {
				program, path := interfaceFixture(t, "lane5/"+family.name+"/"+probe.name)
				truth := onNode(t, path)
				if probe.name == "wrong-value" {
					if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "TypeError:") {
						t.Fatalf("Node %#v", truth)
					}
				} else if truth.exitCode != 0 || string(truth.stdout) != probe.out {
					t.Fatalf("Node %#v want %q", truth, probe.out)
				}
				sanitized, binary := nativelyUncached(t, program)
				for i, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
					if probe.found == "" {
						if diff := disagreement(truth, got); diff != "" {
							t.Fatal(diff)
						}
						continue
					}
					expected := family.signature
					field := "factory." + family.field
					message := "adamic: panic: cast failed: field read failed: " + field + " expected " + expected + ", found " + probe.found + "\n"
					if i < 2 && probe.name == "wrong-value" {
						message = "adamic: panic: cast failed: field read failed: " + field + " is not a " + expected + "; expected " + expected + ", found number\n"
					}
					if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
						t.Fatalf("backend%d %#v want %q", i, got, message)
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
