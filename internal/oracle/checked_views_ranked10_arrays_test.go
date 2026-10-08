package oracle

import (
	"path/filepath"
	"testing"
)

// The tenth ranked group of lane 2 array contracts: ArrowFunction.typeParameters, Get and Set
// accessor modifiers as HasType members, NamedExports.elements, SourceFile.referencedFiles and
// typeReferenceDirectives, ArrayBindingPattern and ObjectBindingPattern elements (recursive through
// BindingName), CallExpression.typeArguments, ClassExpression.heritageClauses and members,
// JSDocTemplateTag.typeParameters, InterfaceType.typeParameters, TypeReference.resolvedTypeArguments
// and Symbol.declarations read through Symbol | undefined, each modeled on tsc's declaration with a
// synthetic tagged root. A refused probe pins what Adamic printed before its lazy read failed.
func TestCheckedViewRanked10ArrayContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, printed, diagnostic string }{
		{"ranked10-array-binding", "[a,,{x,y}]\n", "", ""},
		{"ranked10-array-binding-wrong-tag", "1\nidentifier\n", "1\n", "field read failed: elements[0].kind expected \"binding-element\" | \"omitted\", found string identifier"},
		{"ranked10-arrow-type-parameters", "T extends string, U = number\n", "", ""},
		{"ranked10-arrow-type-parameters-absent", "none\n", "", ""},
		{"ranked10-arrow-type-parameters-wrong-constraint", "T\nundefined\n", "T\n", "field read failed: first.constraint matches no member of TypeNode | undefined; expected TypeNode | undefined, found string"},
		{"ranked10-call-type-arguments", "f<string, number>\n", "", ""},
		{"ranked10-call-type-arguments-absent", "none\n", "", ""},
		{"ranked10-call-type-arguments-wrong-element", "T\nidentifier\n", "T\n", "field read failed: node.typeArguments![0]!.kind expected \"keyword\", found string identifier"},
		{"ranked10-class-expression", "Base:pa,mb\n", "", ""},
		{"ranked10-class-expression-absent", "none:0\n", "", ""},
		{"ranked10-class-expression-wrong-heritage", "undefined\n", "", "field read failed: classExpression(raw).heritageClauses matches no member of NodeArray<HeritageClause> | undefined; expected NodeArray<HeritageClause> | undefined, found number"},
		{"ranked10-class-expression-wrong-member", "1\nundefined\n", "1\n", "field read failed: members[0]!.name is not a Identifier; expected Identifier, found string"},
		{"ranked10-get-modifiers", "static @memo\n", "", ""},
		{"ranked10-interface-type", "2:true\n", "", ""},
		{"ranked10-interface-type-undefined", "none\n", "", ""},
		{"ranked10-interface-type-wrong-element", "1\nT\n", "1\n", "field read failed: parameters[0]!.flags is not a number; expected number, found string"},
		{"ranked10-named-exports", "a,c as b\n", "", ""},
		{"ranked10-named-exports-wrong-array", "undefined\n", "", "field read failed: namedExports(raw).elements is not a NodeArray<ExportSpecifier>; expected NodeArray<ExportSpecifier>, found number"},
		{"ranked10-named-exports-wrong-name", "a\nkeyword\n", "", "field read failed: elements[0]!.name.kind expected \"identifier\" | \"string\", found string keyword"},
		{"ranked10-object-binding", "p:q=d\n", "", ""},
		{"ranked10-object-binding-wrong-array", "1\n", "", "field read failed: objectBinding(raw).elements is not a NodeArray<BindingElement>; expected NodeArray<BindingElement>, found string"},
		{"ranked10-object-binding-wrong-nested", "{a,[3]}\n", "", "field read failed: name.text is not a string; expected string, found number"},
		{"ranked10-references", "b.ts,c.ts!:node@99\n", "", ""},
		{"ranked10-references-wrong-array", "4\n", "", "field read failed: sourceFile(raw).typeReferenceDirectives is not a readonly FileReference[]; expected readonly FileReference[], found string"},
		{"ranked10-references-wrong-file-name", "4\n4\n", "4\n", "field read failed: files[0]!.fileName is not a string; expected string, found number"},
		{"ranked10-references-wrong-mode", "node\n2\n", "node\n", "field read failed: directives[0]!.resolutionMode matches no member of 1 | 99 | undefined; expected 1 | 99 | undefined, found number"},
		{"ranked10-resolved-arguments", "8,16\n", "", ""},
		{"ranked10-resolved-arguments-wrong-array", "undefined\n", "", "field read failed: typeReference(raw).resolvedTypeArguments matches no member of readonly Type[] | undefined; expected readonly Type[] | undefined, found number"},
		{"ranked10-set-modifiers-absent", "none\n", "", ""},
		{"ranked10-set-modifiers-wrong-tag", "abstract\n", "", "field read failed: node.modifiers[element].kind expected \"decorator\" | \"export\" | \"static\", found string abstract"},
		{"ranked10-symbol-declarations", "1:variable;none\n", "", ""},
		{"ranked10-symbol-declarations-absent", "none\n", "", ""},
		{"ranked10-template-tag", "@template K,V\n", "", ""},
		{"ranked10-template-tag-wrong-element", "1\nkeyword\n", "1\n", "field read failed: parameters[0]!.kind expected \"type-parameter\", found string keyword"},
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

// tsc caches resolvedTypeArguments by assigning an array to the optional field. slotContract has no
// certificate for an array representation, so setProperty refuses the checked write before
// lowering; it is never compiled.
func TestCheckedViewRanked10Frontiers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, refusal string }{
		{"ranked10-resolved-arguments-assign", "2:64\n", ":54:1: stage 0 can't lower a checked write without a reifiable source-slot type certificate yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, pathErr := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/lane2", probe.name+".a"))
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
