package oracle

import "testing"

// The eighth ranked group of lane 2 array contracts: JsxElement.children (recursive JsxChild),
// MethodDeclaration.parameters, FunctionDeclaration.parameters, ParameterDeclaration.modifiers,
// EnumDeclaration.members, Bundle.sourceFiles, SourceFile.imports, Signature.compositeSignatures and
// GenericType.typeParameters, each modeled on tsc's declaration with a synthetic tagged root. A push
// of a flat TypeParameter record is certified against the original element contract; a push of a Signature, which holds an array, stays a
// named runtime refusal. A required typeParameters field that is missing refuses, as declared.
// A refused probe pins what Adamic printed before its lazy read failed.
func TestCheckedViewRanked8ArrayContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, printed, diagnostic string }{
		{"ranked8-bundle", "2:a.ts=1,b.ts=0\n", "", ""},
		{"ranked8-bundle-wrong-array", "undefined\n", "", "cast failed: field read failed: bundle(raw).sourceFiles is not a readonly SourceFile[]; expected readonly SourceFile[], found number"},
		{"ranked8-bundle-wrong-element", "1\nundefined\n", "1\n", "cast failed: element read failed: files[0] expected SourceFile, found string"},
		{"ranked8-composite", "1,3\n", "", ""},
		{"ranked8-composite-absent", "absent\n", "", ""},
		{"ranked8-composite-nested", "5\n", "", ""},
		{"ranked8-composite-push", "1,4\n", "", "element write failed: <array write> expected object, found uncertified source element contract"},
		{"ranked8-composite-wrong-array", "4\n", "", "cast failed: field read failed: resolved(raw).callSignatures[0]!.compositeSignatures matches no member of Signature[] | undefined; expected Signature[] | undefined, found string"},
		{"ranked8-composite-wrong-element", "1\none\n", "1\n", "cast failed: field read failed: composite[0]!.minArgumentCount is not a number; expected number, found string"},
		{"ranked8-enum", "identifier:Red:-,string:Blue:Red\n", "", ""},
		{"ranked8-enum-wrong-array", "1\n", "", ""},
		{"ranked8-enum-wrong-initializer", "Red\nundefined\n", "Red\n", "cast failed: field read failed: member.initializer matches no member of Identifier | undefined; expected Identifier | undefined, found number"},
		{"ranked8-enum-wrong-name", "1\nnumber\n", "1\n", "cast failed: field read failed: members[0]!.name.kind expected \"identifier\" | \"string\", found string number"},
		{"ranked8-function-parameters", "1:only\n", "", ""},
		{"ranked8-function-parameters-wrong-element", "1\nundefined\n", "1\n", "cast failed: element read failed: parameters[0] expected ParameterDeclaration, found string"},
		{"ranked8-generic", "2,4t\n", "", ""},
		{"ranked8-generic-missing", "none\n", "", "cast failed: field read failed: generic(raw).typeParameters is not initialized; expected TypeParameter[] | undefined, found missing"},
		{"ranked8-generic-push", "2:8\n", "", ""},
		{"ranked8-generic-push-wrong", "2:8\n", "", "element write failed: <array write> expected { flags: string; }, found { flags: number; }"},
		{"ranked8-generic-undefined", "none\n", "", ""},
		{"ranked8-generic-wrong-element", "2\nfalse\n", "2\n", "cast failed: field read failed: parameters[0]!.isThisType matches no member of boolean | undefined; expected boolean | undefined, found string"},
		{"ranked8-imports", "q./b,t./c\n", "", ""},
		{"ranked8-imports-through-bundle", "./b+./c\n", "", ""},
		{"ranked8-imports-wrong-tag", "b\nidentifier\n", "", "cast failed: field read failed: imports[0].kind expected \"string\" | \"template-literal\", found string identifier"},
		{"ranked8-imports-wrong-text", "string\n7\n", "string\n", "cast failed: field read failed: imports[0]!.text is not a string; expected string, found number"},
		{"ranked8-jsx", "3:hi {name}<b>bold\n", "", ""},
		{"ranked8-jsx-fragment", "<div><>x{}\n", "", ""},
		{"ranked8-jsx-lazy", "ok\n", "", ""},
		{"ranked8-jsx-wrong-array", "2\n", "", "cast failed: field read failed: jsxElement(raw).children is not a NodeArray<JsxChild>; expected NodeArray<JsxChild>, found string"},
		{"ranked8-jsx-wrong-nested", "<p>hi <b>5\n", "", "cast failed: field read failed: child.text is not a string; expected string, found number"},
		{"ranked8-jsx-wrong-tag", "1\nstring\n", "1\n", "cast failed: field read failed: children[0].kind expected \"jsx-element\" | \"jsx-expression\" | \"jsx-fragment\" | \"jsx-text\", found string string"},
		{"ranked8-method-parameters", "2:b\n", "", ""},
		{"ranked8-method-parameters-wrong-array", "undefined\n", "", "cast failed: field read failed: method(raw).parameters is not a NodeArray<ParameterDeclaration>; expected NodeArray<ParameterDeclaration>, found number"},
		{"ranked8-parameter-modifiers", "export @inject\n", "", ""},
		{"ranked8-parameter-modifiers-absent", "absent\n", "", ""},
		{"ranked8-parameter-modifiers-wrong-array", "undefined\n", "", "cast failed: field read failed: parameter(raw).modifiers matches no member of NodeArray<ModifierLike> | undefined; expected NodeArray<ModifierLike> | undefined, found number"},
		{"ranked8-parameter-modifiers-wrong-tag", "1\nstatic\n", "1\n", "cast failed: field read failed: modifiers[0].kind expected \"abstract\" | \"decorator\" | \"export\", found string static"},
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
