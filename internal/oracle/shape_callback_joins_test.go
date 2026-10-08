package oracle

import (
	"strings"
	"testing"
)

func TestShapeCallbackJoinExecution(t *testing.T) {
	for _, name := range []string{"proven-callback-join", "nonconforming-callback-join"} {
		t.Run(name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane3/"+name)
			reference := onNode(t, path)
			if reference.exitCode != 0 {
				t.Fatalf("Node: %+v", reference)
			}
			expected := reference
			if name == "nonconforming-callback-join" {
				if string(reference.stdout) != "true\n0\n" {
					t.Fatalf("Node: %+v", reference)
				}
				expected = run{stdout: []byte("true\n"), stderr: []byte("adamic: panic: field read failed: (value as Identifier).ready is not a boolean; expected boolean, found number\n"), exitCode: 70}
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, actual := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if why := disagreement(expected, actual); why != "" {
					t.Fatalf("%s: %+v", why, actual)
				}
			}
			if name == "proven-callback-join" {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			} else {
				eraseShapeChecksMutant(program)
				for _, actual := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if actual.exitCode != 0 || disagreement(expected, actual) == "" || strings.Contains(string(actual.stderr), "compiler bug") {
						t.Fatalf("unsafe join-erasure mutant escaped: %+v", actual)
					}
					t.Logf("unsafe join-erasure mutant caught: exit=%d stdout=%q", actual.exitCode, actual.stdout)
				}
			}
		})
	}
}

func TestShapeCallbackJoinCountRows(t *testing.T) {
	for _, name := range []string{"proven-callback-join", "nonconforming-callback-join"} {
		t.Log(counted(t, "stage3/interface-downcasts/lane3/"+name+".a", false, nil, false, false))
	}
}
