package oracle

import "testing"

// Admission of a dictionary container must not demand its unread rich elements.
func TestCheckedViewDictionaryCommandOptions(t *testing.T) {
	for _, name := range []string{"good", "wrong", "missing"} {
		t.Run(name, func(t *testing.T) {
			program, path := interfaceFixture(t, "dictionaries/source/command-options-"+name)
			truth := onNode(t, path)
			expected := map[string]string{"good": "object\n", "wrong": "number\n", "missing": "undefined\n"}[name]
			if truth.exitCode != 0 || string(truth.stdout) != expected {
				t.Fatalf("Node: %#v", truth)
			}
			sanitized, _ := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if name == "good" {
					if d := disagreement(truth, got); d != "" {
						t.Fatal(d)
					}
					continue
				}
				found := map[string]string{"wrong": "number", "missing": "missing"}[name]
				want := "adamic: panic: field read failed: view.options is not a CompilerOptions; expected CompilerOptions, found " + found + "\n"
				if name == "missing" {
					want = "adamic: panic: field read failed: view.options is not initialized; expected CompilerOptions, found missing\n"
				}
				if got.exitCode != 70 || string(got.stderr) != want {
					t.Fatalf("container check: %#v; want %q", got, want)
				}
			}
		})
	}
}
