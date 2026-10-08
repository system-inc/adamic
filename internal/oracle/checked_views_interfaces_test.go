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
			want := run{stdout: []byte(probe.stdout)}
			if probe.diagnostic == "" {
				if difference := viewReadDisagreement(want, onNode(t, path), program); difference != "" {
					t.Fatal("Node: " + difference)
				}
			} else {
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: " + probe.diagnostic + "\n")
			}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := viewReadDisagreement(want, got, program); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
			if probe.name == "interfaces-wrong-inherited" {
				dropNestedView(program)
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if viewReadDisagreement(want, got, program) == "" {
						t.Fatal("skip inherited transitive view escaped assertion")
					}
				}
			}
		})
	}
}
