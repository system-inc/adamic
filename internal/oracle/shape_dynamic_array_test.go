package oracle

import (
	"strings"
	"testing"
)

func TestShapeDynamicArrayExecution(t *testing.T) {
	for _, name := range []string{"proven-dynamic", "nonconforming-dynamic"} {
		t.Run(name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane3/"+name)
			reference := onNode(t, path)
			if reference.exitCode != 0 {
				t.Fatalf("Node: %+v", reference)
			}
			expected := reference
			if name == "nonconforming-dynamic" {
				expected = run{stdout: []byte("true\n"), stderr: []byte("adamic: panic: field read failed: (held as Identifier).ready is not a boolean; expected boolean, found number\n"), exitCode: 70}
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, actual := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if why := disagreement(expected, actual); why != "" {
					t.Fatalf("%s: %+v", why, actual)
				}
			}
			if name == "proven-dynamic" {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			} else {
				eraseShapeChecksMutant(program)
				for _, actual := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if actual.exitCode != 0 || disagreement(expected, actual) == "" || strings.Contains(string(actual.stderr), "compiler bug") {
						t.Fatalf("dynamic unsafe-erasure mutant not caught by a successful wrong execution: %+v", actual)
					}
					t.Logf("unsafe dynamic read-erasure mutant caught: exit=%d stdout=%q", actual.exitCode, actual.stdout)
				}
			}
		})
	}
}
