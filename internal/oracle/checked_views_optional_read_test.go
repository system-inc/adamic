package oracle

import "testing"

// A ?. read through a checked view (#1yr7rdh) reads undefined from an absent receiver and checks a
// present one, so the checked source agrees with Node and with its ordinary control on every backend,
// and leaks nothing.
func TestCheckedViewOptionalReadAgreesWithNode(test *testing.T) {
	test.Parallel()
	want := run{stdout: []byte("name\nmissing\n")}
	for _, name := range []string{"lane1/optional-read", "lane1/optional-read-control"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			program, path := interfaceFixture(test, name)
			if difference := disagreement(want, onNode(test, path)); difference != "" {
				test.Fatal("source Node: " + difference)
			}
			sanitized, binary := nativelyUncached(test, program)
			for backend, got := range map[string]run{
				"native sanitized": sanitized,
				"native release":   releasedUncached(test, program),
				"JavaScript":       onJavaScriptBackend(test, program),
			} {
				if difference := disagreement(want, got); difference != "" {
					test.Errorf("%s: %s", backend, difference)
				}
			}
			if report := leaks(test, program, binary); report != "" {
				test.Fatal(report)
			}
		})
	}
}
