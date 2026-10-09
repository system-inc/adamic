package json

import (
	"path/filepath"
	"testing"
)

func TestDocumentedStageZeroGaps(t *testing.T) {
	t.Parallel()
	for _, gap := range []struct {
		file, output string
		args         []string
	}{
		{"multiplePush.ts", "a,b\n", nil},
		{"repeatInTry.ts", "xxx\n", []string{"a", "b", "c"}},
	} {
		t.Run(gap.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("gaps", gap.file))
			if err != nil {
				t.Fatal(err)
			}
			result := onNode(t, path, gap.args...)
			if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != gap.output {
				t.Fatalf("Node: %+v", result)
			}
			program := lowered(t, path)
			actual, binary := natively(t, program, gap.args...)
			for _, side := range []run{actual, onJavaScriptBackend(t, program, gap.args...)} {
				if side.exitCode != 0 || len(side.stderr) != 0 || string(side.stdout) != gap.output {
					t.Fatalf("closed gap differs from Node: %+v", side)
				}
			}
			if report := leaks(t, program, binary, gap.args...); report != "" {
				t.Fatal(report)
			}
		})
	}
}

// emptyFallback.ts lowers on compiler/area-stack (views slice 1): the untyped [] fallback is no
// longer an array of never. It is held to Node on native ASan/UBSan, the JavaScript backend and
// the leak check. The port's checked panic fallback still stands; retiring it is cohere's change.
func TestClosedEmptyFallbackGap(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join("gaps", "emptyFallback.ts"))
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, path)
	nativeRun, binary := natively(t, program)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native", nativeRun}, {"Node", onNode(t, path)}, {"JavaScript backend", onJavaScriptBackend(t, program)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "0\n" {
			t.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
