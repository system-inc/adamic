package oracle

import "testing"

// Census calls these dynamic reads dictionaries; their receiver is a string.
// Keep the candidate row until the complete extraction/indexing witness passes.
func TestCheckedViewDictionaryStringIndexing(t *testing.T) {
	for _, shape := range []string{"string-index", "path-index"} {
		for _, test := range []struct{ name, output string }{
			{"good", "65\n"}, {"surrogate", "55357\n"}, {"missing", "missing\n"}, {"negative", "missing\n"}, {"wrong", "missing\n"},
		} {
			t.Run(shape+"-"+test.name, func(t *testing.T) {
				program, path := interfaceFixture(t, "dictionaries/source/"+shape+"-"+test.name)
				truth := onNode(t, path)
				if truth.exitCode != 0 || string(truth.stdout) != test.output {
					t.Fatalf("Node: %#v", truth)
				}
				sanitized, binary := nativelyUncached(t, program)
				for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if test.name == "wrong" {
						want := "adamic: panic: cast failed: field read failed: view.text is not a string; expected string, found number\n"
						if shape == "path-index" {
							want = "adamic: panic: cast failed: field read failed: view.text is not a Path; expected Path, found number\n"
						}
						if got.exitCode != 70 || string(got.stderr) != want {
							t.Fatalf("named refusal: %#v, want %q", got, want)
						}
					} else if d := disagreement(truth, got); d != "" {
						t.Fatal(d)
					}
				}
				if test.name != "wrong" {
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}
