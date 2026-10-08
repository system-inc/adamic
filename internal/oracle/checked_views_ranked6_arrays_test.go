package oracle

import (
	"path/filepath"
	"testing"
)

// The sixth ranked group of lane 2 array contracts: ResolvedType.callSignatures,
// TemplateLiteralType.texts, TupleType.labeledElementDeclarations, ClassDeclaration.members,
// JsxAttributes.properties, HasJSDoc.jsDoc and the ClassDeclaration | ClassExpression members read,
// each modeled on tsc's declaration with a synthetic tagged root. A push of a JSDoc record, whose
// parent is an object union, stays a named runtime refusal outside the certified flat write subset.
// A refused probe pins what Adamic printed before its lazy read failed.
func TestCheckedViewRanked6ArrayContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, printed, diagnostic string }{
		{"ranked6-class-like-union-cast", "1\n", "", ""},
		{"ranked6-calls", "2:value:0\n", "", ""},
		{"ranked6-calls-filter", "2\n", "", ""},
		{"ranked6-calls-lazy", "2:1\n", "", ""},
		{"ranked6-calls-wrong-array", "4\n", "", "field read failed: resolved(raw).callSignatures is not a readonly Signature[]; expected readonly Signature[], found string"},
		{"ranked6-calls-wrong-element", "1:undefined\n", "", "element read failed: calls[0] expected Signature, found number"},
		{"ranked6-class-like", "a;b,c\n", "", ""},
		{"ranked6-class-like-wrong-name", "9\n", "", "field read failed: member.name.text is not a string; expected string, found number"},
		{"ranked6-class-members", "2:1:size\n", "", ""},
		{"ranked6-class-members-wrong-array", "undefined\n", "", "field read failed: classDeclaration(raw).members is not a NodeArray<ClassElement>; expected NodeArray<ClassElement>, found object"},
		{"ranked6-class-members-wrong-name", "undefined\n", "", "field read failed: members[0]!.name is not a Identifier; expected Identifier, found string"},
		{"ranked6-class-members-wrong-tag", "1:size\nproperty\n", "1:size\n", "field read failed: members[0]!.kind expected \"method\" | \"property\", found string jsdoc"},
		{"ranked6-element-flags", "3:13\n", "", ""},
		{"ranked6-jsdoc-parent", "1:first\n", "", ""},
		{"ranked6-jsdoc-parent-absent", "absent\n", "", ""},
		{"ranked6-jsdoc-parent-push", "2:second\n", "", "element write failed: <array write> expected { kind: \"jsdoc\"; comment: string; }, found uncertified incoming record contract"},
		{"ranked6-jsdoc-parent-wrong-array", "4\n", "", "field read failed: jsDoc(raw).parent.jsDoc matches no member of JSDocArray | undefined; expected JSDocArray | undefined, found string"},
		{"ranked6-jsdoc-parent-wrong-doc", "1\nundefined\n", "1\n", "element read failed: docs[0] expected JSDoc, found number"},
		{"ranked6-jsdoc-parent-wrong-element", "1:first\ntag\n", "1:first\n", "field read failed: docs[0]!.kind expected \"jsdoc\", found string tag"},
		{"ranked6-jsdoc-parent-wrong-tag", "0\n", "", "field read failed: jsDoc(raw).parent.kind expected \"class\" | \"class-expression\" | \"function\", found string tag"},
		{"ranked6-jsx", "2:rest\n", "", ""},
		{"ranked6-jsx-hole", "2:jsx-attribute\ntrue\n", "2:jsx-attribute\n", "element read failed: properties[1] expected JsxAttributeLike, found undefined"},
		{"ranked6-jsx-some", "false\n", "", ""},
		{"ranked6-jsx-wrong-array", "undefined\n", "", "field read failed: jsxAttributes(raw).properties is not a NodeArray<JsxAttributeLike>; expected NodeArray<JsxAttributeLike>, found boolean"},
		{"ranked6-jsx-wrong-spread", "undefined\n", "", "field read failed: first.expression is not a Identifier; expected Identifier, found string"},
		{"ranked6-jsx-wrong-tag", "identifier\n", "", "field read failed: properties[0].kind expected \"jsx-attribute\" | \"jsx-spread\", found string identifier"},
		{"ranked6-labels", "named:head,parameter:tail\n", "", ""},
		{"ranked6-labels-absent", "absent\n", "", ""},
		{"ranked6-labels-hole", "2:head:none\n", "", ""},
		{"ranked6-labels-wrong-array", "undefined\n", "", "field read failed: tupleType(raw).labeledElementDeclarations matches no member of readonly (NamedTupleMember | ParameterDeclaration | undefined)[] | undefined; expected readonly (NamedTupleMember | ParameterDeclaration | undefined)[] | undefined, found number"},
		{"ranked6-labels-wrong-tag", "head\n", "", "field read failed: labels[0].kind expected \"named\" | \"parameter\", found string property"},
		{"ranked6-template", "3:a|b|c:c\n", "", ""},
		{"ranked6-template-assign", "3:xyz\n", "", ""},
		{"ranked6-template-assign-wrong", "x1\n", "", "element read failed: raw.texts[0] expected number, found string"},
		{"ranked6-template-join-wrong", "1|2\n", "", "element read failed: template(raw).texts[element] expected string, found number"},
		{"ranked6-template-map", "0,2,3\n", "", ""},
		{"ranked6-template-wrong-array", "3\n", "", "field read failed: template(raw).texts is not a readonly string[]; expected readonly string[], found string"},
		{"ranked6-template-wrong-element", "2\n1\n", "2\n", "element read failed: texts[0] expected string, found number"},
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

// A JSDocArray built with its own jsDocCache field still needs array own-field production.
// Its Node control remains a named refusal before either backend executes.
func TestCheckedViewRanked6Frontiers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, refusal string }{
		{"ranked6-jsdoc-parent-cache", "1:param\n", ":36:50: stage 0 can't lower array own-field production with unsupported field jsDocCache yet"},
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
