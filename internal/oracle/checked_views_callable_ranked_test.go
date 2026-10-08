package oracle

import (
	"strings"
	"testing"
)

func TestCheckedViewCallableRankedFamilies(t *testing.T) {
	for _, family := range []struct{ directory, field, good, optional, arity, payload, payloadOut string }{
		{"block", "createBlock", "4\n", "3\n3\n4\n", "1", "statements[0]!.value", "3\n"},
		{"array-literal", "createArrayLiteralExpression", "3\n", "0\n3\n3\n4\n", "1", "elements[0]!.value", "3\n"},
		{"object-literal", "createObjectLiteralExpression", "3\n", "0\n3\n3\n4\n", "1", "properties[0]!.value", "3\n"},
		{"function-expression", "createFunctionExpression", "3\n", "3\n8\n8\n", "1", "body.value", "3\n"},
		{"update-block", "updateBlock", "8\n", "", "1", "statements[0]!.value", "5\n"},
		{"emit-notification", "enableEmitNotification", "5\n", "", "2", "", ""},
		{"substitution", "enableSubstitution", "5\n", "", "2", "", ""},
		{"return-statement", "createReturnStatement", "3\n", "7\n7\n3\n", "2", "expression.value", "3\n"},
	} {
		for _, variant := range []string{"good", "optional-values", "wrong-value", "wrong-arity", "wrong-result", "wrong-members", "wrong-parameter-payload"} {
			if variant == "optional-values" && family.optional == "" || (variant == "wrong-result" || variant == "wrong-parameter-payload") && family.payload == "" {
				continue
			}
			t.Run(family.directory+"/"+variant, func(t *testing.T) {
				program, path := interfaceFixture(t, "lane5/ranked-callables/"+family.directory+"/"+variant)
				truth := onNode(t, path)
				out := family.good
				switch variant {
				case "optional-values":
					out = family.optional
				case "wrong-arity":
					out = "9\n"
				case "wrong-result":
					out = "undefined\n"
				case "wrong-members":
					if family.directory == "block" {
						out = "3\n"
					}
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
						found = "found function with arity " + family.arity
					case "wrong-result":
						found = "found function with incompatible result representation"
					case "wrong-members":
						found = "found function with incompatible parameter representations"
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
