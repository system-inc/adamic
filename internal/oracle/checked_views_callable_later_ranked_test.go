package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewCallableLaterRankedFamilies(t *testing.T) {
	for _, family := range []struct{ directory, field, good, optional, arity, payload, payloadOut, resultOut, variants string }{
		{"export-declaration", "createExportDeclaration", "3\n", "0\n14\n", "0", "exportClause.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"computed-name", "updateComputedPropertyName", "7\n", "", "0", "node.value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"source-file-update", "updateSourceFile", "7\n", "7\n13\n", "0", "statement.value", "4\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"context-diagnostic", "addDiagnostic", "3\n", "", "0", "diag.value", "3\n", "", "good,wrong-value,wrong-arity,wrong-members,wrong-parameter-payload"},
		{"token-end", "getTokenEnd", "3\n", "", "1", "", "", "bad\n", "good,wrong-value,wrong-arity,wrong-result"},
		{"token-full-start", "getTokenFullStart", "3\n", "", "1", "", "", "bad\n", "good,wrong-value,wrong-arity,wrong-result"},
		{"helper-factory", "getEmitHelperFactory", "3\n", "", "1", "value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-parameter-payload"},
		{"hoist-variable", "hoistVariableDeclaration", "3\n", "", "0", "node.value", "3\n", "", "good,wrong-value,wrong-arity,wrong-members,wrong-parameter-payload"},
		{"modifier-flags", "createModifiersFromModifierFlags", "3\n", "", "0", "modifier.value", "3\n", "", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"resolution-path", "toPath", "root/3\n", "", "0", "", "", "9\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members"},
		{"performance-measure", "measure", "29\n", "7\n7\n19\n29\n", "0", "", "", "", "good,optional-values,wrong-value,wrong-arity,wrong-members"},
		{"binary", "createBinaryExpression", "8\n", "8\n9\n", "0", "left.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"declaration-name", "getDeclarationName", "3\n", "0\n0\n3\n4\n6\n", "0", "node.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"left-access", "parenthesizeLeftSideOfAccess", "3\n", "3\n3\n3\n4\n", "0", "expression.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"source-files", "getSourceFiles", "3\n", "", "1", "sourceFile.value", "3\n", "0\n", "good,wrong-value,wrong-arity,wrong-result,wrong-parameter-payload"},
		{"type-reference", "createTypeReferenceNode", "8\n", "3\n3\n8\n8\n", "0", "typeNode.value", "5\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"program-options", "getCompilerOptions", "3\n", "", "1", "value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-parameter-payload"},
		{"context-options", "getCompilerOptions", "3\n", "", "1", "value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-parameter-payload"},
		{"system-write", "write", "6\n", "", "0", "", "", "", "good,wrong-value,wrong-arity,wrong-members"},
		{"system-exit", "exit", "3\n", "3\n3\n7\n", "0", "", "", "", "good,optional-values,wrong-value,wrong-arity,wrong-members"},
		{"null", "createNull", "3\n", "", "1", "", "", "undefined\n", "good,wrong-value,wrong-arity,wrong-result"},
		{"parenthesized", "createParenthesizedExpression", "3\n", "", "0", "expression.value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"disallowed-comma", "parenthesizeExpressionForDisallowedComma", "3\n", "", "2", "expression.value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"token-start", "getTokenStart", "3\n", "", "1", "", "", "bad\n", "good,wrong-value,wrong-arity,wrong-result"},
		{"true", "createTrue", "3\n", "", "1", "", "", "undefined\n", "good,wrong-value,wrong-arity,wrong-result"},
		{"local-name", "getLocalName", "3\n", "3\n3\n3\n10\n5\n", "1", "node.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
	} {
		for _, variant := range strings.Split(family.variants, ",") {
			t.Run(family.directory+"/"+variant, func(t *testing.T) {
				program, path := interfaceFixture(t, "lane5/later-ranked-callables/"+family.directory+"/"+variant)
				truth := onNode(t, path)
				out := family.good
				switch variant {
				case "optional-values":
					out = family.optional
				case "wrong-arity":
					out = "9\n"
				case "wrong-result":
					out = family.resultOut
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

func TestCheckedViewCallableLaterRankedUnionRefusal(t *testing.T) {
	path, err := filepath.Abs(repository + "/stage3/interface-downcasts/lane5/later-ranked-callables/string-from-node/good.a")
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "3\n" {
		t.Fatalf("Node %#v", truth)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "unsupported untagged object union contract") {
		t.Fatalf("union payload refusal: %v", err)
	}
}

func TestCheckedViewCallableLaterRankedBindingRefusal(t *testing.T) {
	path, err := filepath.Abs(repository + "/stage3/interface-downcasts/lane5/later-ranked-callables/if-statement/good.a")
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "7\n" {
		t.Fatalf("Node %#v", truth)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "unsupported destructuring representation conversion contract") {
		t.Fatalf("binding refusal: %v", err)
	}
}

func TestCheckedViewCallableLaterRankedMethodRefusal(t *testing.T) {
	path, err := filepath.Abs(repository + "/stage3/interface-downcasts/lane5/later-ranked-callables/lift-block/good.a")
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "3\n" {
		t.Fatalf("Node %#v", truth)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "unbound-method") {
		t.Fatalf("method refusal: %v", err)
	}
}

func TestCheckedViewCallableLaterRankedIntrinsicSetRefusal(t *testing.T) {
	for _, variant := range []string{"add", "has"} {
		t.Run(variant, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/later-ranked-callables/set-intrinsic/"+variant)
			truth := onNode(t, path)
			want := "1\n"
			if variant == "has" {
				want = "true\n"
			}
			if truth.exitCode != 0 || string(truth.stdout) != want {
				t.Fatalf("Node %#v", truth)
			}
			got := onJavaScriptBackend(t, program)
			if got.exitCode != 70 || len(got.stdout) != 0 || !strings.Contains(string(got.stderr), "found function with unknown signature") {
				t.Fatalf("intrinsic signature refusal: %#v", got)
			}
			// Native compilation is the separately logged representation-conversion
			// frontier. This test pins the existing JavaScript refusal only.
		})
	}
}
