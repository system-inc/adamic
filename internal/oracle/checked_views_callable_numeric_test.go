package oracle

import (
	"strings"
	"testing"
)

func TestCheckedViewCallableNumericLiteral(t *testing.T) {
	for _, probe := range []struct{ name, out, found string }{
		{"good", "5\n", ""}, {"optional-values", "5:none\n5:none\n7:1\n8:none\n", ""},
		{"wrong-value", "", "number"}, {"wrong-arity", "", "function with arity 1"},
		{"wrong-members", "", "function with incompatible parameter representations"},
		{"wrong-required", "", "function with incompatible parameter representations"},
		{"wrong-result", "", "function with incompatible result representation"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/numeric-literal/"+probe.name)
			truth := onNode(t, path)
			if probe.found == "" && (truth.exitCode != 0 || string(truth.stdout) != probe.out) {
				t.Fatalf("Node %#v", truth)
			}
			if probe.found != "" {
				if probe.name == "wrong-value" {
					if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "TypeError:") {
						t.Fatalf("Node wrong callable %#v", truth)
					}
				} else {
					want := "5\n"
					if probe.name == "wrong-result" {
						want = "undefined\n"
					}
					if truth.exitCode != 0 || string(truth.stdout) != want {
						t.Fatalf("Node erased contract %#v want %q", truth, want)
					}
				}
			}
			sanitized, binary := nativelyUncached(t, program)
			for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
				if probe.found == "" {
					if diff := disagreement(truth, got); diff != "" {
						t.Fatal(diff)
					}
					continue
				}
				expected := "(value: string | number, numericLiteralFlags?: number | undefined) => NumericLiteral"
				message := "adamic: panic: field read failed: factory.createNumericLiteral expected " + expected + ", found " + probe.found + "\n"
				if probe.name == "wrong-value" && index < 2 {
					message = "adamic: panic: field read failed: factory.createNumericLiteral is not a " + expected + "; expected " + expected + ", found number\n"
				}
				if got.exitCode != 70 || len(got.stdout) != 0 || viewReadDiagnosticMismatch(message, got.stderr, program) {
					t.Fatalf("backend%d %#v want %q", index, got, message)
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
