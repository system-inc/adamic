package oracle

import (
	"path/filepath"
	"testing"
)

// The fifth ranked group of lane 2 array contracts: ResolvedType.constructSignatures,
// ClassDeclaration.modifiers, TupleTypeNode.elements, ExpressionWithTypeArguments.typeArguments,
// JSDoc.tags and FunctionLikeDeclaration.parameters, each modeled on tsc's declaration with a
// synthetic tagged root. FunctionLikeDeclaration is reached through member casts and a proven
// upcast to the union; a direct cast to the union is lane 4's union-target admission and stays a
// named frontend refusal. A refused probe pins what Adamic printed before its lazy read failed.
func TestCheckedViewRanked5ArrayContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, printed, diagnostic string }{
		{"ranked5-function", "2:first,second\n", "", ""},
		{"ranked5-function-arrow", "arrow:0\n", "", ""},
		{"ranked5-function-mixed", "3\n", "", ""},
		{"ranked5-function-wrong-array", "undefined\n", "", "field read failed: declaration.parameters is not a NodeArray<ParameterDeclaration>; expected NodeArray<ParameterDeclaration>, found number"},
		{"ranked5-function-wrong-kind", "1\n", "", "cast failed: this Base is not a MethodDeclaration"},
		{"ranked5-function-wrong-name", "first,\n", "", "field read failed: parameter.name is not a Identifier; expected Identifier, found number"},
		{"ranked5-heritage", "2:8\n", "", ""},
		{"ranked5-heritage-absent", "absent\n", "", ""},
		{"ranked5-heritage-lazy", "5\n", "", ""},
		{"ranked5-heritage-wrong-array", "undefined\n", "", "field read failed: heritage(raw).typeArguments matches no member of NodeArray<TypeNode> | undefined; expected NodeArray<TypeNode> | undefined, found number"},
		{"ranked5-heritage-wrong-late", "5\nlate\n", "5\n", "field read failed: typeArguments[1]!.pos is not a number; expected number, found string"},
		{"ranked5-jsdoc", "2:param:none\n", "", ""},
		{"ranked5-jsdoc-absent", "absent\n", "", ""},
		{"ranked5-jsdoc-filter", "2\n", "", ""},
		{"ranked5-jsdoc-filter-wrong", "1\n", "", "field read failed: tag.tagName.text is not a string; expected string, found number"},
		{"ranked5-jsdoc-wrong-comment", "see\n12\n", "see\n", "field read failed: tags[0]!.comment matches no member of string | undefined; expected string | undefined, found number"},
		{"ranked5-jsdoc-wrong-element", "1\nundefined\n", "1\n", "element read failed: tags[0] expected JSDocTag, found number"},
		{"ranked5-modifiers", "2:sealed\n", "", ""},
		{"ranked5-modifiers-absent", "absent\n", "", ""},
		{"ranked5-modifiers-some", "true\n", "", ""},
		{"ranked5-modifiers-wrong-array", "6\n", "", "field read failed: classDeclaration(raw).modifiers matches no member of NodeArray<ModifierLike> | undefined; expected NodeArray<ModifierLike> | undefined, found string"},
		{"ranked5-modifiers-wrong-decorator", "5\n", "", "field read failed: first.expression.text is not a string; expected string, found number"},
		{"ranked5-modifiers-wrong-tag", "1\n", "", "field read failed: modifiers[0].kind expected \"abstract\" | \"decorator\" | \"export\", found string static"},
		{"ranked5-signatures", "2:1:value\n", "", ""},
		{"ranked5-signatures-assign", "0:0\n", "", ""},
		{"ranked5-signatures-assign-elements-wrong", "x1\n", "", "field read failed: raw.constructSignatures[0]!.parameters[0]!.escapedName is not a number; expected number, found string"},
		{"ranked5-signatures-assign-records", "1:3:next\n", "", ""},
		{"ranked5-signatures-assign-wrong", "undefined\n", "", "field read failed: raw.constructSignatures[0]!.minArgumentCount is not a string; expected string, found number"},
		{"ranked5-signatures-declaration", "method2:absent\n", "", ""},
		{"ranked5-signatures-lazy", "2:1\n", "", ""},
		{"ranked5-signatures-map", "1,0\n", "", ""},
		{"ranked5-signatures-map-wrong", "1,zero\n", "", "field read failed: signature.minArgumentCount is not a number; expected number, found string"},
		{"ranked5-signatures-wrong-array", "undefined\n", "", "field read failed: resolved(raw).constructSignatures is not a readonly Signature[]; expected readonly Signature[], found number"},
		{"ranked5-signatures-wrong-element", "one\n", "", "field read failed: signatures[0]!.minArgumentCount is not a number; expected number, found string"},
		{"ranked5-signatures-wrong-parameters", "1:4\n", "", "field read failed: signatures[0]!.parameters[0]!.escapedName is not a string; expected string, found number"},
		{"ranked5-tuple", "2:1:rest\n", "", ""},
		{"ranked5-tuple-for-of", "7\n", "", ""},
		{"ranked5-tuple-wrong-array", "1\n", "", "field read failed: tuple(raw).elements is not a NodeArray<NamedTupleMember | TypeNode>; expected NodeArray<NamedTupleMember | TypeNode>, found object"},
		{"ranked5-tuple-wrong-element", "2:undefined\n", "", "element read failed: elements[0] expected NamedTupleMember | TypeNode, found string"},
		{"ranked5-tuple-wrong-name", "3\n", "", "field read failed: first.name.text is not a string; expected string, found number"},
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

// A direct cast to the FunctionLikeDeclaration union needs union-target view admission, which the
// plan assigns to lane 4's cast routing. Until then it is refused before lowering, never compiled.
func TestCheckedViewRanked5UnionCastFrontier(t *testing.T) {
	t.Parallel()
	path, pathErr := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/lane2/ranked5-function-union-cast.a"))
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	if got := onNode(t, path); got.exitCode != 0 || string(got.stdout) != "1\n" {
		t.Fatalf("Node: %#v", got)
	}
	_, err := lowered(t, path)
	if err == nil {
		t.Fatal("cast to the FunctionLikeDeclaration union was admitted")
	}
	if got := err.Error(); got != path+":33:69: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)" {
		t.Fatalf("refusal: %s", got)
	}
}
