package oracle

import "testing"

func TestCheckedViewOptionalDeclarations(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, stdout, diagnostic string }{
		{"optional-declarations-present", "2\nfirst\n", ""},
		{"optional-declarations-absent", "missing\n", ""},
		{"optional-declarations-undefined", "missing\n", ""},
		{"optional-declarations-lazy", "2\nfirst\n", ""},
		{"optional-declarations-wrong-array", "", "cast failed: field read failed: symbol(raw).declarations matches no member of Declaration[] | undefined; expected Declaration[] | undefined, found number"},
		{"optional-declarations-wrong-element", "", "cast failed: element read failed: declarations[0] expected Declaration, found number"},
		{"optional-declarations-wrong-field", "", "cast failed: field read failed: declaration.name is not a string; expected string, found number"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "lane2/"+probe.name)
			want := run{stdout: []byte(probe.stdout)}
			// Unsafe source casts execute on Node too; their checked failure is the
			// language's additional contract, pinned below for both backends.
			source := onNode(t, path)
			if probe.diagnostic == "" {
				if difference := disagreement(want, source); difference != "" {
					t.Fatal("Node: " + difference)
				}
			} else {
				if source.exitCode != 0 {
					t.Fatalf("Node control: %#v", source)
				}
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
