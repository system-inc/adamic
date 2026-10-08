package oracle

import "testing"

// The ninth ranked group of lane 2 array contracts: parameters of ArrowFunction, GetAccessor and
// IndexSignature declarations as HasType members; modifiers of Function, Method, Property and Import
// declarations; NamedImports.elements, ModuleBlock.statements, NewExpression.arguments,
// TemplateExpression.templateSpans, NodeBuilderContext.typeStack (pushed and popped through the
// view, as tsc's node builder does), ResolvedType.indexInfos and ResolvedType.properties, each modeled
// on tsc's declaration with a synthetic tagged root. A refused probe pins what Adamic printed before
// its lazy read failed.
func TestCheckedViewRanked9ArrayContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, printed, diagnostic string }{
		{"ranked9-arrow-parameters", "x,y\n", "", ""},
		{"ranked9-arrow-parameters-wrong-array", "undefined\n", "", "cast failed: field read failed: arrow(raw).parameters is not a NodeArray<ParameterDeclaration>; expected NodeArray<ParameterDeclaration>, found boolean"},
		{"ranked9-function-modifiers", "export async\n", "", ""},
		{"ranked9-function-modifiers-wrong-decorator", "@undefined\n", "", "cast failed: field read failed: modifier.expression is not a Identifier; expected Identifier, found string"},
		{"ranked9-get-parameters", "size:0:\n", "", ""},
		{"ranked9-import-modifiers", "export:./a\n", "", ""},
		{"ranked9-import-modifiers-wrong-array", "6\n", "", "cast failed: field read failed: importDeclaration(raw).modifiers matches no member of NodeArray<ModifierLike> | undefined; expected NodeArray<ModifierLike> | undefined, found string"},
		{"ranked9-index-infos", "4r:key,8w\n", "", ""},
		{"ranked9-index-infos-wrong-array", "undefined\n", "", "cast failed: field read failed: resolved(raw).indexInfos is not a readonly IndexInfo[]; expected readonly IndexInfo[], found number"},
		{"ranked9-index-infos-wrong-readonly", "4\n1\n", "4\n", "cast failed: field read failed: infos[0]!.isReadonly is not a boolean; expected boolean, found number"},
		{"ranked9-index-parameters", "key:string\n", "", ""},
		{"ranked9-index-parameters-wrong-name", "1\nundefined\n", "1\n", "cast failed: field read failed: parameters[0]!.name is not a Identifier; expected Identifier, found string"},
		{"ranked9-method-modifiers", "@bound abstract\n", "", ""},
		{"ranked9-module-block", "x,./b\n", "", ""},
		{"ranked9-module-block-filter", "2\n", "", ""},
		{"ranked9-module-block-wrong-tag", "1\nnew\n", "1\n", "cast failed: field read failed: statements[0].kind expected \"import\" | \"variable\", found string new"},
		{"ranked9-named-imports", "a,default as b!\n", "", ""},
		{"ranked9-named-imports-wrong-array", "1\n", "", "cast failed: field read failed: importDeclaration(raw).namedBindings!.elements is not a NodeArray<ImportSpecifier>; expected NodeArray<ImportSpecifier>, found string"},
		{"ranked9-named-imports-wrong-element", "1\nundefined\n", "1\n", "cast failed: element read failed: elements[0] expected ImportSpecifier, found string"},
		{"ranked9-named-imports-wrong-type-only", "a\nno\n", "a\n", "cast failed: field read failed: elements[0]!.isTypeOnly is not a boolean; expected boolean, found string"},
		{"ranked9-new-arguments", "Map(entries)\n", "", ""},
		{"ranked9-new-arguments-absent", "no arguments\n", "", ""},
		{"ranked9-new-arguments-wrong-element", "1\nundefined\n", "1\n", "cast failed: element read failed: items[0] expected Identifier, found number"},
		{"ranked9-properties", "2:ab:b\n", "", ""},
		{"ranked9-properties-push", "2:c\n", "", ""},
		{"ranked9-properties-push-wrong", "2\n", "", "element write failed: <array write> expected { escapedName: string; flags: string; }, found { escapedName: string; flags: number; }"},
		{"ranked9-properties-wrong-name", "4\n1\n", "4\n", "cast failed: field read failed: symbols[0]!.escapedName is not a string; expected string, found number"},
		{"ranked9-property-modifiers-absent", "none\n", "", ""},
		{"ranked9-property-modifiers-wrong-tag", "readonly\n", "", "cast failed: field read failed: node.modifiers[element].kind expected \"abstract\" | \"async\" | \"decorator\" | \"export\", found string readonly"},
		{"ranked9-template-spans", "a${x}b${y}c\n", "", ""},
		{"ranked9-template-spans-wrong-array", "undefined\n", "", "cast failed: field read failed: templateExpression(raw).templateSpans is not a NodeArray<TemplateSpan>; expected NodeArray<TemplateSpan>, found object"},
		{"ranked9-template-spans-wrong-literal", "b\nstring\n", "", "cast failed: field read failed: spans[0]!.literal.kind expected \"template-middle\" | \"template-tail\", found string string"},
		{"ranked9-type-stack", "-1:true:1:1,2,7\n", "", ""},
		{"ranked9-type-stack-assign", "1:3\n", "", ""},
		{"ranked9-type-stack-includes-wrong", "false\n", "", "cast failed: element read failed: context(raw).typeStack[element] expected number, found string"},
		{"ranked9-type-stack-wrong-element", "1\na1\n", "1\n", "cast failed: element read failed: stack[0] expected number, found string"},
		{"ranked9-type-stack-wrong-push", "a,7\n", "", "cast failed: element read failed: <array write> expected number, found heap pointers"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "lane2/"+probe.name)
			truth := onNode(t, path)
			want := run{stdout: []byte(probe.node)}
			if difference := disagreement(want, truth); difference != "" {
				t.Fatal("Node: " + difference)
			}
			if probe.diagnostic != "" {
				want = run{exitCode: 70, stdout: []byte(probe.printed), stderr: []byte("adamic: panic: " + probe.diagnostic + "\n")}
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
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
