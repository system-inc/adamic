package oracle

import "testing"

// The thirteenth ranked group of lane 2 array contracts: type parameters of ClassExpression,
// FunctionTypeNode, MethodSignature and ConstructorDeclaration, FunctionExpression.modifiers,
// MethodSignature.parameters, MappedTypeNode.members, type arguments of ImportTypeNode, TypeQueryNode
// and TaggedTemplateExpression, JSDocImportTag and JSDocTypedefTag comments, ConditionalRoot infer and
// outer type parameters (reached through ConditionalType.root), AnonymousType.aliasTypeArguments,
// SourceFile.commentDirectives and SourceFile.parseDiagnostics, each modeled on tsc's declaration with
// a synthetic tagged root. A whole numeric enum field admits numbers outside its members, as in
// TypeScript. A refused probe pins what Adamic printed before its read failed.
func TestCheckedViewRanked13ArrayContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, printed, diagnostic string }{
		{"ranked13-anonymous-alias-arguments", "4,8\n", "", ""},
		{"ranked13-anonymous-alias-arguments-wrong-array", "1\nundefined\n", "1\n", "element read failed: types[0] expected Type, found array"},
		{"ranked13-class-expression-type-parameters", "T\n", "", ""},
		{"ranked13-class-expression-type-parameters-wrong-element", "1\nkeyword\n", "1\n", "field read failed: parameters[0]!.kind expected \"type-parameter\", found string keyword"},
		{"ranked13-comment-directives", "0-20:expect-error,30-45:ignore\n", "", ""},
		{"ranked13-comment-directives-open-enum", "20\n7\n", "", ""},
		{"ranked13-conditional-root", "true:1:2\n", "", ""},
		{"ranked13-conditional-root-absent", "none:none\n", "", ""},
		{"ranked13-conditional-root-push", "2:262145\n", "", ""},
		{"ranked13-conditional-root-wrong-element", "1\nT\n", "1\n", "field read failed: outer[0]!.flags is not a number; expected number, found string"},
		{"ranked13-constructor-type-parameters-absent", "none\n", "", ""},
		{"ranked13-function-expression-modifiers", "async:x\n", "", ""},
		{"ranked13-function-type-type-parameters", "<T,U>(value) => void\n", "", ""},
		{"ranked13-import-type-arguments", "import(./a)<string>\n", "", ""},
		{"ranked13-jsdoc-import-comment", "see Foo\n", "", ""},
		{"ranked13-jsdoc-typedef-comment", "Pair:a pair\n", "", ""},
		{"ranked13-jsdoc-typedef-comment-wrong", "boolean\n", "", "field read failed: jsdocTypedef(raw).comment matches no member of string | NodeArray<JSDocComment> | undefined; expected string | NodeArray<JSDocComment> | undefined, found boolean"},
		{"ranked13-mapped-members", "K:x\n", "", ""},
		{"ranked13-mapped-members-absent", "none\n", "", ""},
		{"ranked13-mapped-members-wrong-array", "1\n", "", "field read failed: mapped(raw).members matches no member of NodeArray<TypeElement> | undefined; expected NodeArray<TypeElement> | undefined, found string"},
		{"ranked13-method-signature", "map<U>(callback,thisArg)\n", "", ""},
		{"ranked13-method-signature-wrong-parameter", "\n", "", "field read failed: item.name is not a Identifier; expected Identifier, found number"},
		{"ranked13-parse-diagnostics", "2:1005,1128:3\n", "", ""},
		{"ranked13-parse-diagnostics-push", "2:1128\n", "", ""},
		{"ranked13-parse-diagnostics-wrong-start", "1005\nthree\n", "1005\n", "field read failed: diagnostics[0]!.start is not a number; expected number, found string"},
		{"ranked13-tagged-template-arguments-absent", "html`x`\n", "", ""},
		{"ranked13-tagged-template-arguments-wrong-element", "<>\n", "", "element read failed: node.typeArguments[element] expected KeywordTypeNode, found number"},
		{"ranked13-type-query-arguments", "typeof f<number, string>\n", "", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "lane2/"+probe.name)
			truth := onNode(t, path)
			want := run{stdout: []byte(probe.node)}
			if difference := viewReadDisagreement(want, truth, program); difference != "" {
				t.Fatal("Node: " + difference)
			}
			if probe.diagnostic != "" {
				want = run{exitCode: 70, stdout: []byte(probe.printed), stderr: []byte("adamic: panic: " + probe.diagnostic + "\n")}
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := viewReadDisagreement(want, got, program); difference != "" {
					t.Fatalf("%s; stderr %q; stdout %q; exit %d", difference, got.stderr, got.stdout, got.exitCode)
				}
			}
			if probe.diagnostic == "" {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}
