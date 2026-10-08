package oracle

import (
	"strings"
	"testing"
)

func TestCheckedViewCallableAggregateWitnesses(t *testing.T) {
	for _, family := range []struct{ name, field, expected, good, arity, nodeArity, payload string }{
		{"expression-statement", "factory.createExpressionStatement", "(expression: Expression) => ExpressionStatement", "5\n", "0", "9\n", "statement.expression.value"},
		{"diagnostic-add", "diagnostics.add", "(diagnostic: Diagnostic) => void", "out.js:denied\n", "0", "wrong\n", ""},
		{"void-zero", "factory.createVoidZero", "() => VoidExpression", "0\n", "1", "9\n", "node.expression.value"},
		{"this", "factory.createThis", "() => ThisExpression", "1\n", "1", "9\n", "node.value"},
		{"emit-helper", "context.requestEmitHelper", "(helper: EmitHelper) => void", "decorate\n", "0", "wrong\n", ""},
		{"node-check-flag", "resolver.hasNodeCheckFlag", "(node: Node, flags: number) => boolean", "true\n", "1", "false\n", ""},
	} {
		variants := []string{"good", "wrong-value", "wrong-arity", "wrong-result"}
		if family.payload != "" {
			variants = append(variants, "wrong-payload")
		}
		for _, variant := range variants {
			t.Run(family.name+"/"+variant, func(t *testing.T) {
				program, path := interfaceFixture(t, "lane5/aggregate/"+family.name+"/"+variant)
				truth := onNode(t, path)
				nodeOut := family.good
				nodeThrows := variant == "wrong-value" || variant == "wrong-result" && (family.name == "expression-statement" || family.name == "void-zero")
				if variant == "wrong-arity" {
					nodeOut = family.nodeArity
				}
				if variant == "wrong-payload" {
					nodeOut = "bad\n"
				}
				if variant == "wrong-result" {
					nodeOut = ""
					if family.name == "this" {
						nodeOut = "undefined\n"
					}
					if family.name == "node-check-flag" {
						nodeOut = "7\n"
					}
				}
				if nodeThrows {
					if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "TypeError:") {
						t.Fatalf("Node %#v", truth)
					}
				} else if truth.exitCode != 0 || string(truth.stdout) != nodeOut {
					t.Fatalf("Node %#v want %q", truth, nodeOut)
				}
				sanitized, binary := nativelyUncached(t, program)
				for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
					if variant == "good" {
						if difference := disagreement(truth, got); difference != "" {
							t.Fatal(difference)
						}
						continue
					}
					field, expected, found := family.field, family.expected, "number"
					if variant == "wrong-arity" {
						found = "function with arity " + family.arity
					}
					if variant == "wrong-result" {
						found = "function with incompatible result representation"
					}
					if variant == "wrong-payload" {
						field, expected, found = family.payload, "number", "string"
					}
					message := "adamic: panic: field read failed: " + field + " expected " + expected + ", found " + found + "\n"
					if variant == "wrong-payload" || index < 2 && variant == "wrong-value" {
						message = "adamic: panic: field read failed: " + field + " is not a " + expected + "; expected " + expected + ", found " + found + "\n"
					}
					if got.exitCode != 70 || len(got.stdout) != 0 || viewReadDiagnosticMismatch(message, got.stderr, program) {
						t.Fatalf("backend%d %#v want %q", index, got, message)
					}
				}
				if variant == "good" {
					if report := leaksUncached(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}

func TestCheckedViewCallableAggregateParameterReads(t *testing.T) {
	for _, probe := range []struct{ name, field, expected, found, stdout string }{
		{"diagnostic-add", "diagnostic.messageText", "number", "string", "out.js:denied\n"},
		{"emit-helper", "helper.name", "number", "string", "decorate\n"},
		{"node-check-flag", "node.flags", "string", "number", "false\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/aggregate/"+probe.name+"/wrong-parameter-payload")
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != probe.stdout {
				t.Fatalf("Node %#v", truth)
			}
			sanitized, _ := nativelyUncached(t, program)
			for _, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
				message := "adamic: panic: field read failed: " + probe.field + " is not a " + probe.expected + "; expected " + probe.expected + ", found " + probe.found + "\n"
				if got.exitCode != 70 || len(got.stdout) != 0 || viewReadDiagnosticMismatch(message, got.stderr, program) {
					t.Fatalf("parameter read %#v want %q", got, message)
				}
			}
		})
	}
}
