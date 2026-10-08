package oracle

import "testing"

func TestCheckedViewObjectUnions(t *testing.T) {
	for _, probe := range []struct{ name, stdout, diagnostic string }{
		{"unions-objects-good", "true\ntrue\n", ""},
		{"unions-objects-wrong-payload", "", "field read failed: child.ready is not a boolean; expected boolean, found number"},
		{"unions-objects-wrong-tag", "", "field read failed: box.child.tag expected \"left\" | \"right\", found string wrong"},
		{"unions-objects-missing-tag", "", "field read failed: view.child.tag is not initialized; expected \"left\" | \"right\", found missing"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane1/"+probe.name)
			want := run{stdout: []byte(probe.stdout)}
			if probe.diagnostic == "" {
				if difference := disagreement(want, onNode(t, path)); difference != "" {
					t.Fatal("Node: " + difference)
				}
			} else {
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: " + probe.diagnostic + "\n")
			}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
		})
	}
}
