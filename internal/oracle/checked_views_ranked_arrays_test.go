package oracle

import "testing"

func TestCheckedViewRankedArrayContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, diagnostic string }{
		{"ranked-union-types", "2\n1\n", ""},
		{"ranked-intersection-types", "1\n3\n", ""},
		{"ranked-signature-type-parameters", "1\n4:T\n", ""},
		{"ranked-signature-parameters", "1\nargument\n", ""},
		{"ranked-source-statements", "1:0:9:false\nstatement\n", ""},
		{"ranked-signature-absent", "absent\n", ""},
		{"ranked-union-lazy", "2\n1\n", ""},
		{"ranked-union-wrong-array", "undefined\n", "cast failed: field read failed: union(raw).types is not a Type[]; expected Type[], found number"},
		{"ranked-intersection-wrong-field", "wrong\n", "cast failed: field read failed: first.id is not a number; expected number, found string"},
		{"ranked-signature-wrong-array", "undefined\n", "cast failed: field read failed: signature(raw).typeParameters matches no member of readonly TypeParameter[] | undefined; expected readonly TypeParameter[] | undefined, found number"},
		{"ranked-signature-wrong-element", "undefined\n", "cast failed: element read failed: parameters[0] expected Symbol, found number"},
		{"ranked-source-wrong-array", "undefined\n", "cast failed: field read failed: source(raw).statements is not a NodeArray<Statement>; expected NodeArray<Statement>, found object"},
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
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.diagnostic + "\n")}
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
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
