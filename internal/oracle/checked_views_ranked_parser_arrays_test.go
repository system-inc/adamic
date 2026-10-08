package oracle

import "testing"

func TestCheckedViewRankedParserArrayContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, diagnostic string }{
		{"ranked2-variables", "2:1\n", ""},
		{"ranked2-variables-wrong-array", "undefined\n", "field read failed: variables(raw).declarations is not a NodeArray<VariableDeclaration>; expected NodeArray<VariableDeclaration>, found number"},
		{"ranked2-call", "2:1\n", ""},
		{"ranked2-call-wrong-array", "undefined\n", "field read failed: call(raw).arguments is not a NodeArray<Expression>; expected NodeArray<Expression>, found number"},
		{"ranked2-object", "2:1\n", ""},
		{"ranked2-object-wrong-array", "undefined\n", "field read failed: object(raw).properties is not a NodeArray<ObjectLiteralElementLike>; expected NodeArray<ObjectLiteralElementLike>, found number"},
		{"ranked2-tuple", "2:1\n", ""},
		{"ranked2-tuple-wrong-array", "undefined\n", "field read failed: tuple(raw).elementFlags is not a readonly ElementFlags[]; expected readonly ElementFlags[], found number"},
		{"ranked2-array", "2:1\n", ""},
		{"ranked2-array-wrong-array", "undefined\n", "field read failed: array(raw).elements is not a NodeArray<Expression>; expected NodeArray<Expression>, found number"},
		{"ranked2-tuple-number", "7\n", ""},
		{"ranked2-tuple-wrong-member", "2\n", "field read failed: values[0] expected ElementFlags.Required, found number 2"},
		{"ranked2-tuple-wrong-flag", "wrong\n", "element read failed: values[0] expected ElementFlags, found string"},
		{"ranked2-call-lazy", "2:1\n", ""},
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
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.diagnostic + "\n")}
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
