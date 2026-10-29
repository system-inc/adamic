package oracle

import "testing"

// A ?. read through a checked view (#1yr7rdh) reads undefined from an absent receiver and checks a
// present one, so the checked source agrees with Node and with its ordinary control on every backend,
// and leaks nothing.
func TestCheckedViewOptionalReadAgreesWithNode(t *testing.T) {
	t.Parallel()
	want := run{stdout: []byte("name\nmissing\n")}
	for _, name := range []string{"lane1/optional-read", "lane1/optional-read-control"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, name)
			if difference := disagreement(want, onNode(t, path)); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			sanitized, binary := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native sanitized": sanitized, "native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s", backend, difference)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}
