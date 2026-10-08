package oracle

import "testing"

func TestCheckedViewInterfaces(t *testing.T) {
	for _, probe := range []struct{ name, stdout, diagnostic string }{
		{"interfaces-good", "true:okok\ntrue\n", ""},
		{"interfaces-missing-inherited", "", "field read failed: view.child.ready is not initialized; expected boolean, found missing"},
		{"interfaces-wrong-inherited", "", "field read failed: view.child.ready is not a boolean; expected boolean, found number"},
		{"interfaces-uninitialized-object", "", "field read failed: view.child.next is not initialized; expected Child, found uninitialized"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane1/"+probe.name)
			checkV1SourceNode(t, path, probe.name)
			want := run{stdout: []byte(probe.stdout)}
			if probe.diagnostic == "" {
				if difference := disagreement(want, onNode(t, path)); difference != "" {
					t.Fatal("Node: " + difference)
				}
			} else {
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: " + probe.diagnostic + "\n")
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
			if probe.diagnostic == "" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			if probe.name == "interfaces-wrong-inherited" {
				dropNestedView(program)
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if disagreement(want, got) == "" {
						t.Fatal("skip inherited transitive view escaped assertion")
					}
				}
			}
		})
	}
}
