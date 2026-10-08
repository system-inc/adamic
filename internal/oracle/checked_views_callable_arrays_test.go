package oracle

import (
	"strings"
	"testing"
)

func TestCheckedViewCallableArrays(t *testing.T) {
	for _, family := range []struct{ name, field, expected, good, arity, nodeArity, element string }{
		{"call-expression", "factory.createCallExpression", "(expression: Expression, typeArguments: readonly TypeNode[] | undefined, argumentsArray: readonly Expression[] | undefined) => CallExpression", "1:5\n", "1", "1:9\n", "argumentsArray[0]"},
		{"inline-expressions", "context.factory.inlineExpressions", "(expressions: readonly Expression[]) => Expression", "5\n", "0", "9\n", "expressions[0]"},
	} {
		variants := []string{"good", "wrong-value", "wrong-arity", "wrong-result", "wrong-element"}
		if family.name == "call-expression" {
			variants = append(variants, "undefined-array")
		}
		for _, variant := range variants {
			t.Run(family.name+"/"+variant, func(t *testing.T) {
				program, path := interfaceFixture(t, "lane5/array-callables/"+family.name+"/"+variant)
				truth := onNode(t, path)
				nodeOut := family.good
				switch variant {
				case "wrong-arity":
					nodeOut = family.nodeArity
				case "undefined-array":
					nodeOut = "1:-1\n"
				case "wrong-element":
					nodeOut = "[object Object]\n"
					if family.name == "call-expression" {
						nodeOut = "1:[object Object]\n"
					}
				case "wrong-result":
					nodeOut = "undefined\n"
				}
				if variant == "wrong-value" || variant == "wrong-result" && family.name == "call-expression" {
					if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "TypeError:") {
						t.Fatalf("Node %#v", truth)
					}
				} else if truth.exitCode != 0 || string(truth.stdout) != nodeOut {
					t.Fatalf("Node %#v want %q", truth, nodeOut)
				}
				sanitized, binary := nativelyUncached(t, program)
				for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
					if variant == "good" || variant == "undefined-array" {
						if diff := disagreement(truth, got); diff != "" {
							t.Fatalf("%s backend%d %#v", diff, index, got)
						}
						continue
					}
					found := "number"
					if variant == "wrong-arity" {
						found = "function with arity " + family.arity
					}
					if variant == "wrong-result" {
						found = "function with incompatible result representation"
					}
					message := "adamic: panic: field read failed: " + family.field + " expected " + family.expected + ", found " + found + "\n"
					if variant == "wrong-value" && index < 2 {
						message = "adamic: panic: field read failed: " + family.field + " is not a " + family.expected + "; expected " + family.expected + ", found number\n"
					}
					if variant == "wrong-element" {
						message = "adamic: panic: element read failed: " + family.element + " expected number, found object\n"
					}
					if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
						t.Fatalf("backend%d %#v want %q", index, got, message)
					}
				}
				if variant == "good" || variant == "undefined-array" {
					if report := leaksUncached(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}
