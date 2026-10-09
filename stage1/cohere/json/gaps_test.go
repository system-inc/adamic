package json

import (
	"path/filepath"
	"testing"
)

func TestDocumentedStageZeroGaps(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join("gaps", "multiplePush.ts"))
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
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "a,b\n" {
			t.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
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

// Step 21 admits the unchanged dynamic repeat under its source catch.
func TestClosedRepeatInTryGap(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join("gaps", "repeatInTry.ts"))
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, path)
	nativeRun, binary := natively(t, program, "a", "b", "c")
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native", nativeRun}, {"Node", onNode(t, path, "a", "b", "c")}, {"JavaScript backend", onJavaScriptBackend(t, program, "a", "b", "c")},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "xxx\n" {
			t.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(t, program, binary, "a", "b", "c"); report != "" {
		t.Fatal(report)
	}
}
