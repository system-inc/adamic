package fuzz

import (
	"path/filepath"
	"testing"
)

// Feature-bearing programs select their runtime from the emitted C.
func TestFuzzerUsesProgramsFeatures(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	checkout, err := Prepare("../..", filepath.Join(directory, "checkout"))
	if err != nil {
		t.Fatal(err)
	}
	outcome := checkout.Try(`function count(...items:number[]):number {return arguments.length;}
const call=count;
console.log(call(1,2).toString());
`, filepath.Join(directory, "program"))
	if outcome.Verdict != Agreed || string(outcome.Node.Stdout) != "2\n" ||
		string(outcome.Native.Stdout) != "2\n" || string(outcome.Backend.Stdout) != "2\n" ||
		outcome.Node.ExitCode != 0 || outcome.Native.ExitCode != 0 || outcome.Backend.ExitCode != 0 {
		t.Fatalf("want matching feature runtimes held to Node, got %#v", outcome)
	}
}
