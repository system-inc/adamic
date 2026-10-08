package oracle

import (
	"testing"
)

func TestCheckedViewCallableBoxing(t *testing.T) {
	for _, probe := range []struct{ name, stdout string }{
		{"number-result", "5\n"}, {"boolean-result", "true\n"}, {"parameter", "5\n"}, {"union-producer", "5\nyes\n"}, {"parameter-throw", "5\n"}, {"stored", "5\n"}, {"view-result", "5\n"}, {"view-parameter", "5\n"}, {"array-predicate-parameter", "2,3\n"}, {"array-map-parameter", "1,2,3\n"}, {"array-map-parameter-throw", "1\n"}, {"array-sort-parameter", "1,2,3\n"}, {"named-producer", "5\n"}, {"named-union-parameter", "5\n"}, {"array-forEach-parameter", "1\n2\n"}, {"array-reduce-parameter", "6\n"}, {"map-parameter", "one:1\n"}, {"set-parameter", "1\n"}, {"undefined-parameter", "5\nundefined\n"}, {"heap-string-parameter", "some-text\n"}, {"predicate-parameter", "1,2,3\n"}, {"optional-absent", "undefined\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/boxing/"+probe.name)
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != probe.stdout {
				t.Fatalf("Node: %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatal(difference)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

func TestCheckedViewCallableBoxingRefusals(t *testing.T) {
	for _, probe := range []struct{ name, expected, found, stdout string }{
		{"view-result-wrong", "() => string | number", "function with incompatible result representation", "true\n"},
		{"view-result-wrong-members", "() => string | number", "function with incompatible result representation", "true\n"},
		{"view-parameter-wrong", "(value: string | number) => string", "function with incompatible parameter representations", "5\n"},
		{"view-parameter-wrong-members", "(value: string | number) => string", "function with incompatible parameter representations", "5\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/boxing/"+probe.name)
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != probe.stdout {
				t.Fatalf("Node: %#v", truth)
			}
			sanitized, _ := nativelyUncached(t, program)
			for _, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
				expected := "adamic: panic: cast failed: field read failed: viewed.run expected " + probe.expected + ", found " + probe.found + "\n"
				if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != expected {
					t.Fatalf("refusal %#v want %q", got, expected)
				}
			}
		})
	}
}
