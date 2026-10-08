package oracle

import (
	"strings"
	"testing"
)

func TestCheckedViewNullableSelection(t *testing.T) {
	for _, shape := range []string{"scalar", "tagged", "object"} {
		for _, variant := range []string{"null", "undefined", "both"} {
			for _, mutation := range []string{"", "-wrong", "-nested"} {
				if mutation == "-nested" && shape != "tagged" {
					continue
				}
				name := shape + "-" + variant + mutation
				t.Run(name, func(t *testing.T) {
					program, path := interfaceFixture(t, "nullish/selection/"+name)
					truth := onNode(t, path)
					native, binary := nativelyUncached(t, program)
					for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if mutation == "" {
							if diff := disagreement(truth, got); diff != "" {
								t.Fatalf("%s: %#v", diff, got)
							}
						} else if got.exitCode != 70 || !strings.Contains(string(got.stderr), "value") {
							t.Fatalf("nullable selection mutant ran on: %#v", got)
						}
					}
					if mutation == "" {
						if report := leaks(t, program, binary); report != "" {
							t.Fatal(report)
						}
					}
					t.Logf("Node exit=%d stdout=%q; mutation=%q", truth.exitCode, truth.stdout, mutation)
				})
			}
		}
	}
}
