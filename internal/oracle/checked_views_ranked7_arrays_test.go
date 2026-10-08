package oracle

import (
	"path/filepath"
	"testing"
)

// The seventh ranked group of lane 2 array contracts: ClassDeclaration.heritageClauses,
// HeritageClause.types, SetAccessorDeclaration.parameters, ConstructorDeclaration.parameters,
// Type.aliasTypeArguments, TemplateLiteralType.types, SourceFile.bindDiagnostics,
// Diagnostic.relatedInformation and ParsedCommandLine.projectReferences, each modeled on tsc's
// declaration with a synthetic tagged root. Pushing a DiagnosticRelatedInformation record (its
// original elements hold undefined-typed fields) or a DiagnosticWithLocation (it holds a SourceFile)
// stays a named runtime refusal outside the certified flat write subset. A refused probe pins what Adamic printed before its lazy read failed.
func TestCheckedViewRanked7ArrayContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, printed, diagnostic string }{
		{"ranked7-alias", "2:4+8\n", "", ""},
		{"ranked7-alias-absent", "absent\n", "", ""},
		{"ranked7-alias-nested", "32\n", "", ""},
		{"ranked7-alias-wrong-array", "undefined\n", "", "field read failed: indexed(raw).objectType.aliasTypeArguments matches no member of readonly Type[] | undefined; expected readonly Type[] | undefined, found object"},
		{"ranked7-alias-wrong-element", "1\nfour\n", "1\n", "field read failed: types[0]!.flags is not a number; expected number, found string"},
		{"ranked7-bind", "2:1002:other.ts:6\n", "", ""},
		{"ranked7-bind-assign", "0:0\n", "", ""},
		{"ranked7-bind-filter", "2\n", "", ""},
		{"ranked7-bind-push", "2:1002\n", "", "element write failed: <array write> expected object, found uncertified source element contract"},
		{"ranked7-bind-wrong-array", "undefined\n", "", "field read failed: sourceFile(raw).bindDiagnostics is not a DiagnosticWithLocation[]; expected DiagnosticWithLocation[], found number"},
		{"ranked7-bind-wrong-file", "1001\nundefined\n", "1001\n", "field read failed: diagnostics[0]!.file is not a SourceFile; expected SourceFile, found string"},
		{"ranked7-constructor-parameters", "2:a,b\n", "", ""},
		{"ranked7-constructor-parameters-wrong-name", "a,false\n", "", "field read failed: parameter.name.text is not a string; expected string, found boolean"},
		{"ranked7-constructor-parameters-wrong-tag", "a\nproperty\n", "a\n", "field read failed: parameters[0]!.kind expected \"parameter\", found string property"},
		{"ranked7-heritage", "2:96:First,Second\n", "", ""},
		{"ranked7-heritage-absent", "absent\n", "", ""},
		{"ranked7-heritage-some", "true\n", "", ""},
		{"ranked7-heritage-types-lazy", "2:First\n", "", ""},
		{"ranked7-heritage-types-wrong-array", "undefined\n", "", "field read failed: classDeclaration(raw).heritageClauses![0]!.types is not a NodeArray<ExpressionWithTypeArguments>; expected NodeArray<ExpressionWithTypeArguments>, found number"},
		{"ranked7-heritage-types-wrong-element", "1\nundefined\n", "1\n", "element read failed: types[0] expected ExpressionWithTypeArguments, found string"},
		{"ranked7-heritage-wrong-array", "7\n", "", "field read failed: classDeclaration(raw).heritageClauses matches no member of NodeArray<HeritageClause> | undefined; expected NodeArray<HeritageClause> | undefined, found string"},
		{"ranked7-heritage-wrong-token", "1\n5\n", "1\n", "field read failed: clauses[0]!.token expected 96 | 119, found number 5"},
		{"ranked7-references", "/core:none:false,/util:./util:true\n", "", ""},
		{"ranked7-references-absent", "absent\n", "", ""},
		{"ranked7-references-wrong-array", "5\n", "", "field read failed: commandLine(raw).projectReferences matches no member of readonly ProjectReference[] | undefined; expected readonly ProjectReference[] | undefined, found string"},
		{"ranked7-references-wrong-circular", "/core\nyes\n", "/core\n", "field read failed: references[0]!.circular matches no member of boolean | undefined; expected boolean | undefined, found string"},
		{"ranked7-references-wrong-path", "1\n3\n", "1\n", "field read failed: references[0]!.path is not a string; expected string, found number"},
		{"ranked7-related", "7,8:none\n", "", ""},
		{"ranked7-related-absent", "absent\n", "", ""},
		{"ranked7-related-push", "7,9\n", "", "element write failed: <array write> expected object, found uncertified source element contract"},
		{"ranked7-related-wrong-array", "4\n", "", "field read failed: commandLine(raw).errors[0]!.relatedInformation matches no member of DiagnosticRelatedInformation[] | undefined; expected DiagnosticRelatedInformation[] | undefined, found string"},
		{"ranked7-related-wrong-code", "seven\n", "", "field read failed: related.code is not a number; expected number, found string"},
		{"ranked7-set-parameters", "next\n", "", ""},
		{"ranked7-set-parameters-wrong-array", "1\n", "", "field read failed: setAccessor(raw).parameters is not a NodeArray<ParameterDeclaration>; expected NodeArray<ParameterDeclaration>, found object"},
		{"ranked7-template-types", "2:12:1\n", "", ""},
		{"ranked7-template-types-wrong-element", "1\nundefined\n", "1\n", "element read failed: types[0] expected Type, found string"},
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

// tsc's addRelatedInfo assigns a fresh array to relatedInformation; here the Diagnostic is read
// from a viewed array element. That write has no reifiable source-slot certificate yet,
// so it is refused before lowering, never compiled.
func TestCheckedViewRanked7Frontiers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, refusal string }{
		{"ranked7-related-assign", "empty:0\n", ":34:1: stage 0 can't lower a checked write without a reifiable source-slot type certificate yet"},
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
