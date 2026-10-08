package oracle

import (
	"strings"
	"testing"
)

func TestCheckedViewCallableElementAccess(t *testing.T) {
	for _, probe := range []struct{ name, out, found string }{
		{"good", "5:3\n", ""}, {"boxed-argument", "5:4\n", ""},
		{"wrong-value", "", "number"}, {"wrong-arity", "5:9\n", "function with arity 1"},
		{"wrong-result", "", "function with incompatible result representation"},
		{"wrong-members", "5:3\n", "function with incompatible parameter representations"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/element-access/"+probe.name)
			truth := onNode(t, path)
			if probe.name == "wrong-value" || probe.name == "wrong-result" {
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
						t.Fatalf("%s: Node %#v backend%d %#v", diff, truth, i, got)
					}
					continue
				}
				expected := "(expression: Expression, index: number | Expression) => ElementAccessExpression"
				message := "adamic: panic: field read failed: factory.createElementAccessExpression expected " + expected + ", found " + probe.found + "\n"
				if i < 2 && probe.name == "wrong-value" {
					message = "adamic: panic: field read failed: factory.createElementAccessExpression is not a " + expected + "; expected " + expected + ", found number\n"
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
