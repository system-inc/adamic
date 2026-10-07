package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Runtime-loaded operands and an FMA-capable target expose lost contraction protection.
func TestReleaseFixtureKeepsArithmeticUnfused(t *testing.T) {
	t.Parallel()
	cpu, why, ok := fusingProcessor()
	if !ok {
		t.Skip(why)
	}
	path, err := filepath.Abs("../oracle/testdata/release_fma.a")
	if err != nil {
		t.Fatal(err)
	}
	node := exec.Command("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", path)
	want, err := node.CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v\n%s", err, want)
	}
	if string(want) != "0\n0\n0\n" {
		t.Fatalf("fixture's Node rounding boundaries are wrong: %q", want)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	generated := C(program)
	directory := t.TempDir()
	source := filepath.Join(directory, "main.c")
	if err := os.WriteFile(source, []byte(generated), 0o644); err != nil {
		t.Fatal(err)
	}
	options := Options{Release: true, cpu: cpu}
	binary := filepath.Join(directory, "good")
	if err := Build(generated, binary, options); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", binary); got != string(want) {
		t.Fatalf("ThinLTO fixture differs from Node: %q vs %q", got, want)
	}
	library, err := RuntimeLibrary("", options)
	if err != nil {
		t.Fatal(err)
	}
	mutant := filepath.Join(directory, "mutant")
	flags := slices.DeleteFunc(LinkFlags(options), func(s string) bool { return s == "-ffp-contract=off" })
	flags = append(flags, "-I", filepath.Dir(library), "-o", mutant, source)
	flags = append(flags, RuntimeLinkFlags(library)...)
	flags = append(flags, "-lm")
	if output, err := exec.Command("clang", flags...).CombinedOutput(); err != nil {
		t.Fatalf("mutant compile: %v\n%s", err, output)
	}
	got := runWithInput(t, "", mutant)
	if got == string(want) {
		t.Fatal("fixture cannot see a link-only -ffp-contract=off omission")
	}
	t.Logf("FMA target -march=%q; Node and ThinLTO %q; link-option omission %q", cpu, want, got)
}
