package fuzz

import (
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"testing"
)

// A featureful program must ignore an incompatible featureless supplied archive.
func TestFuzzerChoosesProgramRuntimeFeatures(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	checkout, err := PrepareBuilt("../..", buildcache.Adamic(t), filepath.Join(directory, "checkout"))
	if err != nil {
		t.Fatal(err)
	}
	checkout.runtime, err = native.RuntimeLibrary("", native.Options{})
	if err != nil {
		t.Fatal(err)
	}
	outcome := checkout.Try(`function count(...items:number[]):number {return arguments.length;}
const call=count;
console.log(call(1,2).toString());
`, filepath.Join(directory, "program"))
	if outcome.Verdict != Agreed || string(outcome.Node.Stdout) != "2\n" || string(outcome.Native.Stdout) != "2\n" || string(outcome.Backend.Stdout) != "2\n" {
		t.Fatalf("program runtime features did not agree: %#v", outcome)
	}
}
