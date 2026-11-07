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
func TestReleaseFixtureKeepsArithmeticUnfused(test *testing.T) {
	test.Parallel()
	cpu, why, ok := fusingProcessor()
	if !ok {
		test.Skip(why)
	}
	path, err := filepath.Abs("../oracle/testdata/release_fma.a")
	if err != nil {
		test.Fatal(err)
	}
	node := exec.Command("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", path)
	want, err := node.CombinedOutput()
	if err != nil {
		test.Fatalf("Node: %v\n%s", err, want)
	}
	if string(want) != "0\n0\n0\n" {
		test.Fatalf("fixture's Node rounding boundaries are wrong: %q", want)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		test.Fatal(err)
	}
	generated := C(program)
	directory := test.TempDir()
	source := filepath.Join(directory, "main.c")
	if err := os.WriteFile(source, []byte(generated), 0o644); err != nil {
		test.Fatal(err)
	}
	options := Options{Release: true, cpu: cpu}
	binary := filepath.Join(directory, "good")
	if err := Build(generated, binary, options); err != nil {
		test.Fatal(err)
	}
	if got := runWithInput(test, "", binary); got != string(want) {
		test.Fatalf("ThinLTO fixture differs from Node: %q vs %q", got, want)
	}
	library, err := RuntimeLibrary("", options)
	if err != nil {
		test.Fatal(err)
	}
	mutant := filepath.Join(directory, "mutant")
	flags := slices.DeleteFunc(LinkFlags(options), func(s string) bool { return s == "-ffp-contract=off" })
	flags = append(flags, "-I", filepath.Dir(library), "-o", mutant, source)
	flags = append(flags, RuntimeLinkFlags(library)...)
	flags = append(flags, "-lm")
	if output, err := exec.Command("clang", flags...).CombinedOutput(); err != nil {
		test.Fatalf("mutant compile: %v\n%s", err, output)
	}
	got := runWithInput(test, "", mutant)
	if got == string(want) {
		test.Fatal("fixture cannot see a link-only -ffp-contract=off omission")
	}
	test.Logf("FMA target -march=%q; Node and ThinLTO %q; link-option omission %q", cpu, want, got)
}

func TestReleaseSplitFixtureAgreesWithNode(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs("../oracle/testdata/release_fma.a")
	if err != nil {
		test.Fatal(err)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		test.Fatal(err)
	}
	binary := filepath.Join(test.TempDir(), "split")
	if err := Build(C(program), binary, Options{Release: true, Split: true}); err != nil {
		test.Fatal(err)
	}
	if got := runWithInput(test, "", binary); got != "0\n0\n0\n" {
		test.Fatalf("split release rounding: %q", got)
	}
}
