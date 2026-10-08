package oracle

import "testing"

// The eleventh ranked group of lane 2 array contracts: ClassDeclaration.typeParameters, modifiers of
// Export, Variable, ImportEquals, ClassExpression, Constructor, Arrow and FunctionExpression
// declarations, FunctionExpression.parameters, ImportAttributes.elements, IndexInfo.components (an
// array of intersections, reached through ResolvedType.indexInfos), InterfaceDeclaration
// heritageClauses and members, SourceFile.libReferenceDirectives and
// NodeBuilderContext.reverseMappedStack, each modeled on tsc's declaration with a synthetic tagged
// root. Pushing a ReverseMappedSymbol, which holds an object field, stays a named runtime refusal.
// A refused probe pins what Adamic printed before its read failed.
func TestCheckedViewRanked11ArrayContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, printed, diagnostic string }{
		{"ranked11-arrow-modifiers", "async:x\n", "", ""},
		{"ranked11-arrow-modifiers-decorator", "1\ndecorator\n", "1\n", "cast failed: field read failed: modifiers[0]!.kind expected \"async\" | \"declare\" | \"export\" | \"static\", found string decorator"},
		{"ranked11-class-expression-modifiers", "@sealed\n", "", ""},
		{"ranked11-class-type-parameters", "T,U\n", "", ""},
		{"ranked11-class-type-parameters-wrong-array", "1\n", "", "cast failed: field read failed: classDeclaration(raw).typeParameters matches no member of NodeArray<TypeParameterDeclaration> | undefined; expected NodeArray<TypeParameterDeclaration> | undefined, found string"},
		{"ranked11-constructor-modifiers-wrong-tag", "override\n", "", "cast failed: field read failed: modifiers[element].kind expected \"async\" | \"declare\" | \"decorator\" | \"export\" | \"static\", found string override"},
		{"ranked11-export-modifiers", "@internal:true\n", "", ""},
		{"ranked11-function-expression-parameters", "a,b\n", "", ""},
		{"ranked11-function-expression-parameters-unread-kind", "a\n", "", ""},
		{"ranked11-import-attributes", "118:type: json, \"mode\": strict:false\n", "", ""},
		{"ranked11-import-attributes-wrong-token", "0\n1\n", "0\n", "cast failed: field read failed: node.token expected 118 | 132, found number 1"},
		{"ranked11-import-attributes-wrong-value", "type\nundefined\n", "type\n", "cast failed: field read failed: elements[0]!.value is not a StringLiteral; expected StringLiteral, found string"},
		{"ranked11-import-equals-modifiers-absent", "none\n", "", ""},
		{"ranked11-index-components", "3[key];none\n", "", ""},
		{"ranked11-index-components-wrong-array", "undefined\n", "", "cast failed: field read failed: resolved(raw).indexInfos[0]!.components matches no member of ElementWithComputedPropertyName[] | undefined; expected ElementWithComputedPropertyName[] | undefined, found object"},
		{"ranked11-index-components-wrong-name", "3\nidentifier\n", "", "cast failed: field read failed: components[0].name.kind expected \"computed\", found string identifier"},
		{"ranked11-interface-heritage", "A&B\n", "", ""},
		{"ranked11-interface-heritage-wrong-types", "1\n1\n", "1\n", "cast failed: field read failed: clauses[0]!.types is not a NodeArray<ExpressionWithTypeArguments>; expected NodeArray<ExpressionWithTypeArguments>, found string"},
		{"ranked11-interface-members", "none:size: number;area(0)\n", "", ""},
		{"ranked11-interface-members-wrong-tag", "size\nproperty\n", "", "cast failed: field read failed: members[0].kind expected \"method-signature\" | \"property-signature\", found string property"},
		{"ranked11-lib-references", "es2020,dom!\n", "", ""},
		{"ranked11-lib-references-wrong-end", "es2020\n31\n", "es2020\n", "cast failed: field read failed: directives[0]!.end is not a number; expected number, found string"},
		{"ranked11-reverse-mapped", "1:a:false\n", "", ""},
		{"ranked11-reverse-mapped-pop", "b:1:1\n", "", ""},
		{"ranked11-reverse-mapped-push", "2\n", "", "element write failed: <array write> expected object, found uncertified source element contract"},
		{"ranked11-reverse-mapped-undefined", "none\n", "", ""},
		{"ranked11-reverse-mapped-wrong-links", "a\ntrue\n", "a\n", "cast failed: field read failed: stack[0]!.links is not a { propertyType?: Type; }; expected { propertyType?: Type; }, found number"},
		{"ranked11-variable-modifiers", "export declare\n", "", ""},
		{"ranked11-variable-modifiers-wrong-array", "undefined\n", "", "cast failed: field read failed: variableStatement(raw).modifiers matches no member of NodeArray<ModifierLike> | undefined; expected NodeArray<ModifierLike> | undefined, found number"},
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
