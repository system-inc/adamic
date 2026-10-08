package oracle

import (
	"path/filepath"
	"testing"
)

// The fourteenth ranked group of lane 2 array contracts: ConstructorTypeNode modifiers, parameters and
// type parameters, modifiers of MethodSignature, PropertySignature, NamespaceExportDeclaration,
// ExportAssignment and EnumDeclaration, JSDocSignature.parameters, JsxFragment.children,
// UnionTypeNode.types, DefaultClause.statements, EmitNode helpers and tokenSourceMapRanges (reached
// through an optional emitNode), InterfaceType localTypeParameters and resolvedBaseTypes,
// declaredProperties, TransientSymbol.declarations and CircularBuildOrder.circularDiagnostics, each
// modeled on tsc's declaration with a synthetic tagged root. A refused probe pins what Adamic
// printed before its read failed.
func TestCheckedViewRanked14ArrayContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, printed, diagnostic string }{
		{"ranked14-circular-diagnostics", "2:6202\n", "", ""},
		{"ranked14-circular-diagnostics-wrong-message", "6202\n3\n", "6202\n", "field read failed: diagnostics[0]!.messageText is not a string; expected string, found number"},
		{"ranked14-constructor-type", "abstract new <T>(a, b) => object\n", "", ""},
		{"ranked14-constructor-type-absent", "none:none:0\n", "", ""},
		{"ranked14-constructor-type-wrong-parameters", "undefined\n", "", "field read failed: constructorType(raw).parameters is not a NodeArray<ParameterDeclaration>; expected NodeArray<ParameterDeclaration>, found object"},
		{"ranked14-default-clause", "x;return\n", "", ""},
		{"ranked14-default-clause-wrong-tag", "1\nenum\n", "1\n", "field read failed: statements[0].kind expected \"return\" | \"variable\", found string enum"},
		{"ranked14-emit-helpers", "typescript:extends:false:0,typescript:awaiter:false:-\n", "", ""},
		{"ranked14-emit-helpers-absent", "none\n", "", ""},
		{"ranked14-emit-helpers-wrong-scoped", "x\nno\n", "", "field read failed: helpers[0].scoped is not a boolean; expected boolean, found string"},
		{"ranked14-enum-modifiers-absent", "none\n", "", ""},
		{"ranked14-enum-modifiers-wrong-array", "5\n", "", "field read failed: enumDeclaration(raw).modifiers matches no member of NodeArray<ModifierLike> | undefined; expected NodeArray<ModifierLike> | undefined, found string"},
		{"ranked14-export-assignment-modifiers", "export:true:value\n", "", ""},
		{"ranked14-interface-type", "1:524288,2097152:size\n", "", ""},
		{"ranked14-interface-type-base-push", "2:1\n", "", ""},
		{"ranked14-interface-type-local-undefined", "none:0\n", "", ""},
		{"ranked14-interface-type-wrong-bases", "undefined\n", "", "field read failed: interfaceType(raw).resolvedBaseTypes is not a Type[]; expected Type[], found number"},
		{"ranked14-interface-type-wrong-property", "size\ntrue\n", "size\n", "field read failed: properties[0]!.flags is not a number; expected number, found boolean"},
		{"ranked14-jsdoc-signature-parameters", "a [b]\n", "", ""},
		{"ranked14-jsdoc-signature-parameters-wrong-bracketed", "a\nno\n", "a\n", "field read failed: parameters[0]!.isBracketed is not a boolean; expected boolean, found string"},
		{"ranked14-jsx-fragment-children", "a<>b</>c\n", "", ""},
		{"ranked14-jsx-fragment-children-wrong-nested", "a<>2</>\n", "", "field read failed: child.text is not a string; expected string, found number"},
		{"ranked14-method-signature-modifiers", "readonly\n", "", ""},
		{"ranked14-namespace-export-modifiers", "@grammar\n", "", ""},
		{"ranked14-property-signature-modifiers-wrong-tag", "decorator\n", "", "field read failed: item.kind expected \"abstract\" | \"export\" | \"readonly\", found string decorator"},
		{"ranked14-token-source-map-ranges", "1-2,_,5-9\n", "", ""},
		{"ranked14-token-source-map-ranges-wrong-element", "1\n2\n", "1\n", "field read failed: first.end is not a number; expected number, found string"},
		{"ranked14-transient-declarations", "1:variable\n", "", ""},
		{"ranked14-transient-declarations-absent", "none\n", "", ""},
		{"ranked14-union-type-types", "string | number\n", "", ""},
		{"ranked14-union-type-types-wrong-element", "T\nidentifier\n", "T\n", "field read failed: types[0]!.kind expected \"keyword\", found string identifier"},
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

// view_lazy.go keeps a field-name fallback: once any interface in the program has a field whose
// contract is unsupported, a viewed read of a field with that name and no contract of its own is
// refused. Here an unused interface's text, string | ((name: string) => string), refuses a
// KeywordTypeNode.text read. tsc's EmitHelper.text has that type.
func TestCheckedViewRanked14Frontiers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, refusal string }{
		{"ranked14-text-name-fallback", "string\n", ":61:13: Adamic 0.1 refuses checked view read of field text with unsupported callable contract; prove or implement the callable contract before reading this field"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, pathErr := filepath.Abs(checkedViewFixturePath(filepath.Join(repository, "stage3/interface-downcasts/lane2", probe.name+".a")))
			if pathErr != nil {
				t.Fatal(pathErr)
			}
			if difference := disagreement(run{stdout: []byte(probe.node)}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			_, err := lowered(t, path)
			if err == nil {
				t.Fatal("frontier program was admitted")
			}
			if got := err.Error(); got != path+probe.refusal {
				t.Fatalf("refusal: %s", got)
			}
		})
	}
}
