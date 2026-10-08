package oracle

import (
	"fmt"
	"testing"
)

// These valid source controls remain excluded from certification and counts.
func TestCheckedViewCallableShareCMapProducerBoundaries(t *testing.T) {
	for _, row := range []struct {
		rank     int
		declared string
	}{
		{1706, "Map<number, { type: Type; declarations: IndexSignatureDeclaration[]; }>"},
		{1769, "Map<string, RegExp>"},
	} {
		t.Run(fmt.Sprintf("rank-%d", row.rank), func(t *testing.T) {
			program, path := interfaceFixture(t, fmt.Sprintf("lane5/share-c/families/rank-%d/good", row.rank))
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != "completed\n" || len(truth.stderr) != 0 {
				t.Fatalf("Node control: %#v", truth)
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: Map contract failed: (value as Target).items; expected " + row.declared + ", found " + row.declared + "\n")}
			sanitized, _ := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Fatalf("recorded producer boundary changed: %s; stderr %q", diff, got.stderr)
				}
			}
		})
	}
}
