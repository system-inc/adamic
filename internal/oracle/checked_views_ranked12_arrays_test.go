package oracle

import "testing"

// The twelfth ranked group of lane 2 array contracts: JSDocFunctionType.parameters, type parameters
// of Method, Function, TypeAlias and Interface declarations, ModuleDeclaration.modifiers,
// TypeParameterDeclaration.modifiers, ParsedCommandLine.errors, SourceFile.moduleAugmentations and
// packageJsonLocations, JsonSourceFile.statements, TemplateLiteralTypeNode.templateSpans,
// TypeLiteralNode.members, CaseClause.statements, NodeWithTypeArguments.typeArguments and
// JSDoc.comment (string | NodeArray<JSDocComment>), each modeled on tsc's declaration with a
// synthetic tagged root. Pushing a Diagnostic, whose original elements hold undefined-typed fields,
// stays a named runtime refusal. A refused probe pins what Adamic printed before its read failed.
func TestCheckedViewRanked12ArrayContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, printed, diagnostic string }{
		{"ranked12-case-statements", "case x: let y; return\n", "", ""},
		{"ranked12-case-statements-wrong-tag", "1\nmodule\n", "1\n", "field read failed: statements[0].kind expected \"return\" | \"variable\", found string module"},
		{"ranked12-command-errors", "2:5023,6046:none\n", "", ""},
		{"ranked12-command-errors-push", "2:6046\n", "", "element write failed: <array write> expected object, found uncertified source element contract"},
		{"ranked12-command-errors-wrong-code", "bad\n50231\n", "bad\n", "field read failed: errors[0]!.code is not a number; expected number, found string"},
		{"ranked12-function-type-parameters-absent", "none\n", "", ""},
		{"ranked12-interface-type-parameters", "T:pvalue\n", "", ""},
		{"ranked12-interface-type-parameters-wrong-name", "\n", "", "field read failed: item.name is not a Identifier; expected Identifier, found string"},
		{"ranked12-jsdoc-comment-array", "see Foo\n", "", ""},
		{"ranked12-jsdoc-comment-string", "plain\n", "", ""},
		{"ranked12-jsdoc-comment-wrong", "undefined\n", "", "field read failed: jsDoc(raw).comment matches no member of string | NodeArray<JSDocComment> | undefined; expected string | NodeArray<JSDocComment> | undefined, found number"},
		{"ranked12-jsdoc-comment-wrong-element", "see 7\n", "", "field read failed: item.text is not a string; expected string, found number"},
		{"ranked12-jsdoc-function-parameters", "none:a,b\n", "", ""},
		{"ranked12-jsdoc-function-parameters-wrong-array", "1\n", "", "field read failed: jsdocFunction(raw).parameters is not a NodeArray<ParameterDeclaration>; expected NodeArray<ParameterDeclaration>, found string"},
		{"ranked12-json-statements", "1:object-literal:1\n", "", ""},
		{"ranked12-json-statements-wrong-expression", "1\nstring\n", "1\n", "field read failed: statements[0]!.expression.kind expected \"object-literal\", found string string"},
		{"ranked12-method-type-parameters", "K,V\n", "", ""},
		{"ranked12-module-augmentations", "\"express\",global\n", "", ""},
		{"ranked12-module-augmentations-wrong-tag", "express\nkeyword\n", "", "field read failed: augmentations[0].kind expected \"identifier\" | \"string\", found string keyword"},
		{"ranked12-module-modifiers", "export declare:string:fs\n", "", ""},
		{"ranked12-module-modifiers-wrong-array", "undefined\n", "", "field read failed: moduleDeclaration(raw).modifiers matches no member of NodeArray<ModifierLike> | undefined; expected NodeArray<ModifierLike> | undefined, found boolean"},
		{"ranked12-package-json-locations", "2:/a/package.json;/package.json\n", "", ""},
		{"ranked12-package-json-locations-absent", "none\n", "", ""},
		{"ranked12-package-json-locations-wrong-element", "1\n1\n", "1\n", "element read failed: locations[0] expected string, found number"},
		{"ranked12-template-literal-type", "a<string>-<number>!\n", "", ""},
		{"ranked12-template-literal-type-wrong-type", "!\nundefined\n", "!\n", "field read failed: spans[0]!.type is not a KeywordTypeNode; expected KeywordTypeNode, found string"},
		{"ranked12-type-alias-type-parameters", "T:object\n", "", ""},
		{"ranked12-type-arguments", "Map<string, number>\n", "", ""},
		{"ranked12-type-arguments-wrong-element", "K\nidentifier\n", "K\n", "field read failed: typeArguments[0]!.kind expected \"keyword\", found string identifier"},
		{"ranked12-type-literal-members", "px,mf\n", "", ""},
		{"ranked12-type-literal-members-wrong-tag", "px\n", "", "field read failed: node.members[element].kind expected \"method-signature\" | \"property-signature\", found string parameter"},
		{"ranked12-type-parameter-modifiers", "const in\n", "", ""},
		{"ranked12-type-parameter-modifiers-decorator", "1\ndecorator\n", "1\n", "field read failed: modifiers[0]!.kind expected \"const\" | \"declare\" | \"export\" | \"in\" | \"out\", found string decorator"},
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
