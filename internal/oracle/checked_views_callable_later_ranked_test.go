package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewCallableLaterRankedFamilies(t *testing.T) {
	for _, family := range []struct{ directory, field, good, optional, arity, payload, payloadOut, resultOut, variants string }{
		{"canonical-file-name", "getCanonicalFileName", "abc/3\n", "", "0", "", "", "3\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members"},
		{"emit-host-options", "getCompilerOptions", "3\n", "", "1", "value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-parameter-payload"},
		{"package-cache", "getPackageJsonInfoCache", "3\n", "", "1", "value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-parameter-payload"},
		{"module-block", "createModuleBlock", "3\n", "0\n3\n", "0", "statement.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"reflect-get", "createReflectGetCall", "6\n", "3\n3\n6\n", "0", "target.value", "1\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"type-operator", "createTypeOperatorNode", "146\n", "146\n151\n161\n", "0", "type.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"type-parameter", "createTypeParameterDeclaration", "13\n", "3\n3\n7\n13\n", "0", "modifier.value", "1\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"function-update", "updateFunctionDeclaration", "36\n", "1\n7\n36\n", "0", "typeParameter.value", "5\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"import-update", "updateImportDeclaration", "15\n", "5\n5\n15\n", "0", "moduleSpecifier.value", "4\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"program-directory", "getCurrentDirectory", "abc\n", "", "1", "", "", "3\n", "good,wrong-value,wrong-arity,wrong-result"},
		{"system-directory", "getCurrentDirectory", "abc\n", "", "1", "", "", "3\n", "good,wrong-value,wrong-arity,wrong-result"},
		{"cancellation-check", "throwIfCancellationRequested", "3\n", "", "1", "", "", "", "good,wrong-value,wrong-arity"},
		{"writer-text", "getText", "abc\n", "", "1", "", "", "3\n", "good,wrong-value,wrong-arity,wrong-result"},
		{"redirect-path", "toPath", "root/3\n", "", "0", "", "", "9\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members"},
		{"iteration-type", "resolveIterationType", "7\n", "0\n3\n7\n", "0", "type.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"specifier-directory", "getCurrentDirectory", "abc\n", "", "1", "", "", "3\n", "good,wrong-value,wrong-arity,wrong-result"},
		{"expression-type-arguments", "createExpressionWithTypeArguments", "7\n", "3\n3\n7\n", "0", "expression.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"indexed-access-type", "createIndexedAccessTypeNode", "7\n", "", "0", "objectType.value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"logical-not", "createLogicalNot", "3\n", "", "0", "operand.value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"outer-expressions", "restoreOuterExpressions", "7\n", "4\n4\n9\n", "0", "innerExpression.value", "4\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"get-accessor-update", "updateGetAccessorDeclaration", "21\n", "4\n4\n4\n4\n4\n4\n4\n21\n", "0", "parameter.value", "4\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"parenthesized-update", "updateParenthesizedExpression", "7\n", "", "0", "node.value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"set-accessor-update", "updateSetAccessorDeclaration", "16\n", "4\n4\n4\n4\n4\n4\n4\n16\n", "0", "parameter.value", "4\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"host-source-files", "getSourceFiles", "3\n", "", "1", "sourceFile.value", "3\n", "0\n", "good,wrong-value,wrong-arity,wrong-result,wrong-parameter-payload"},
		{"property-declaration", "createPropertyDeclaration", "24\n", "4\n4\n4\n4\n4\n4\n4\n3\n8\n", "0", "modifier.value", "2\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"qualified-name", "createQualifiedName", "7\n", "7\n7\n7\n7\n", "0", "left.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"import-declaration", "createImportDeclaration", "10\n", "3\n3\n10\n", "0", "moduleSpecifier.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"function-call-call", "createFunctionCallCall", "6\n", "3\n6\n", "0", "target.value", "1\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"partial-expression", "createPartiallyEmittedExpression", "7\n", "3\n3\n7\n", "0", "expression.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"strict-inequality", "createStrictInequality", "7\n", "", "0", "left.value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"parameter-update", "updateParameterDeclaration", "28\n", "5\n5\n5\n4\n", "0", "node.value", "1\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"property-update", "updatePropertyDeclaration", "25\n", "5\n5\n5\n5\n5\n5\n5\n4\n9\n", "0", "node.value", "1\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"resolution-settings", "getCompilationSettings", "3\n", "", "1", "value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-parameter-payload"},
		{"scanner-scan", "scan", "3\n", "", "1", "", "", "bad\n", "good,wrong-value,wrong-arity,wrong-result"},
		{"lexical-start", "startLexicalEnvironment", "3\n", "", "1", "", "", "", "good,wrong-value,wrong-arity"},
		{"writer-space", "writeSpace", "6\n", "", "0", "", "", "", "good,wrong-value,wrong-arity,wrong-members"},
		{"substitution-hook", "onSubstituteNode", "4\n", "", "0", "node.value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"internal-name", "getInternalName", "3\n", "3\n3\n4\n5\n6\n", "0", "node.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"call-update", "updateCallExpression", "10\n", "7\n3\n10\n", "0", "argument.value", "4\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"class-update", "updateClassDeclaration", "21\n", "1\n7\n21\n", "0", "member.value", "6\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"prefix-operand", "parenthesizeOperandOfPrefixUnary", "3\n", "", "0", "operand.value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"syntax-kind", "formatSyntaxKind", "Identifier\n", "undefined\nIdentifier\nUnknown\n", "0", "", "", "3\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members"},
		{"static-block", "createClassStaticBlockDeclaration", "3\n", "", "0", "body.value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"prefix-unary", "createPrefixUnaryExpression", "43\n", "44\n", "0", "operand.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"export-specifier", "createExportSpecifier", "7\n", "4\n8\n7\n", "0", "propertyName.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"property-signature", "createPropertySignature", "4\n", "3\n10\n10\n10\n10\n10\n10\n10\n", "0", "modifier.value", "1\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"type-check", "createTypeCheck", "9\n", "7\n12\n9\n9\n10\n9\n9\n9\n11\n", "0", "value.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"lexical-end", "endLexicalEnvironment", "3\n", "0\n", "1", "statement.value", "3\n", "0\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-parameter-payload"},
		{"module-format", "getEmitModuleFormatOfFile", "3\n", "", "0", "sourceFile.value", "3\n3\n", "bad\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"false", "createFalse", "3\n", "", "1", "", "", "undefined\n", "good,wrong-value,wrong-arity,wrong-result"},
		{"parenthesized-type", "createParenthesizedType", "3\n", "", "0", "type.value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"variable-update-tagged", "updateVariableDeclaration", "7\n", "10\n5\n", "0", "name.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"conditional-expression", "createConditionalExpression", "12\n", "15\n", "0", "condition.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"binary-update", "updateBinaryExpression", "10\n", "11\n", "0", "node.value", "2\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"config-diagnostic", "addConfigDiagnostic", "3\n", "", "0", "diag.value", "3\n", "", "good,wrong-value,wrong-arity,wrong-members,wrong-parameter-payload"},
		{"scanner-text", "setText", "3\n", "0\n3\n9\n3\n", "0", "", "", "", "good,optional-values,wrong-value,wrong-arity,wrong-members"},
		{"read-helpers", "readEmitHelpers", "3\n", "0\n", "1", "helper.value", "3\n", "0\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-parameter-payload"},
		{"case-sensitive", "useCaseSensitiveFileNames", "true\n", "false\n", "1", "", "", "bad\n", "good,optional-values,wrong-value,wrong-arity,wrong-result"},
		{"diagnostic-newline", "getNewLine", "abc\n", "", "1", "", "", "3\n", "good,wrong-value,wrong-arity,wrong-result"},
		{"logical-and", "createLogicalAnd", "7\n", "", "0", "left.value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"named-exports", "createNamedExports", "0\n", "1\n", "0", "element.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"token-text", "getTokenText", "abc\n", "", "1", "", "", "3\n", "good,wrong-value,wrong-arity,wrong-result"},
		{"token-value", "getTokenValue", "abc\n", "", "1", "", "", "3\n", "good,wrong-value,wrong-arity,wrong-result"},
		{"writer-line", "writeLine", "3\n", "3\n3\n3\n7\n", "0", "", "", "", "good,optional-values,wrong-value,wrong-arity,wrong-members"},
		{"watcher-close", "close", "3\n", "", "1", "", "", "", "good,wrong-value,wrong-arity"},
		{"source-file-path", "getSourceFileByPath", "3\n", "0\n", "0", "sourceFile.value", "3\n", "undefined\n", "good,optional-values,wrong-value,wrong-arity,wrong-result,wrong-members,wrong-parameter-payload"},
		{"emit-resolver", "getEmitResolver", "3\n", "", "1", "value", "3\n", "undefined\n", "good,wrong-value,wrong-arity,wrong-result,wrong-parameter-payload"},
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

// variable-update is the earlier untagged reduction. The original BindingName
// tags are certified separately by variable-update-tagged.
func TestCheckedViewCallableLaterRankedOriginalReadRefusals(t *testing.T) {
	for _, witness := range []struct{ directory, output, refusal string }{
		{"for-update", "15\n", "unsupported untagged object union contract"},
		{"comma-reducer", "7\n", "unbound-method"},
		{"variable-update", "7\n", "unsupported untagged object union contract"},
		{"arrow-function", "3\n", "unsupported untagged object union contract"},
		{"literal-type", "3\n", "unsupported untagged object union contract"},
		{"environment-variable", "abc\n", "unbound-method"},
	} {
		t.Run(witness.directory, func(t *testing.T) {
			path, err := filepath.Abs(repository + "/stage3/interface-downcasts/lane5/later-ranked-callables/" + witness.directory + "/good.a")
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != witness.output {
				t.Fatalf("Node %#v", truth)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), witness.refusal) {
				t.Fatalf("original read refusal: %v", err)
			}
		})
	}
}
