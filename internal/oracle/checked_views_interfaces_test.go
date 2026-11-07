package oracle

import "testing"

func TestCheckedViewInterfaces(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, stdout, diagnostic string }{
		{"interfaces-good", "true:okok\ntrue\n", ""},
		{
			"interfaces-missing-inherited",
			"",
			"field read failed: view.child.ready is not initialized; expected boolean, found missing",
		},
		{
			"interfaces-wrong-inherited",
			"",
			"field read failed: view.child.ready is not a boolean; expected boolean, found number",
		},
		{
			"interfaces-uninitialized-object",
			"",
			"field read failed: view.child.next is not initialized; expected Child, found uninitialized",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			program, path := interfaceFixture(test, "lane1/"+probe.name)
			want := run{stdout: []byte(probe.stdout)}
			if probe.diagnostic == "" {
				if difference := disagreement(want, onNode(test, path)); difference != "" {
					test.Fatal("Node: " + difference)
				}
			} else {
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: " + probe.diagnostic + "\n")
			}
			actual, _ := nativelyUncached(test, program)
			for _, got := range []run{actual, releasedUncached(test, program), onJavaScriptBackend(test, program)} {
				if difference := disagreement(want, got); difference != "" {
					test.Fatalf("%s; got %#v", difference, got)
				}
			}
			if probe.name == "interfaces-wrong-inherited" {
				dropNestedView(program)
				for _, got := range []run{releasedUncached(test, program), onJavaScriptBackend(test, program)} {
					if disagreement(want, got) == "" {
						test.Fatal("skip inherited transitive view escaped assertion")
					}
				}
			}
		})
	}
}
