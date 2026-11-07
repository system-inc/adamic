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
	path, absolutePathError := filepath.Abs("../oracle/testdata/release_fma.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	node := exec.Command("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", path)
	want, combinedOutputError := node.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("Node: %v\n%s", combinedOutputError, want)
	}
	if string(want) != "0\n0\n0\n" {
		test.Fatalf("fixture's Node rounding boundaries are wrong: %q", want)
	}
	checked, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	generated := C(program)
	directory := test.TempDir()
	source := filepath.Join(directory, "main.c")
	if writeError := os.WriteFile(source, []byte(generated), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	options := Options{Release: true, cpu: cpu}
	binary := filepath.Join(directory, "good")
	if buildError := Build(generated, binary, options); buildError != nil {
		test.Fatal(buildError)
	}
	if got := runWithInput(test, "", binary); got != string(want) {
		test.Fatalf("ThinLTO fixture differs from Node: %q vs %q", got, want)
	}
	library, runtimeLibraryError := RuntimeLibrary("", options)
	if runtimeLibraryError != nil {
		test.Fatal(runtimeLibraryError)
	}
	mutant := filepath.Join(directory, "mutant")
	flags := slices.DeleteFunc(LinkFlags(options), func(s string) bool { return s == "-ffp-contract=off" })
	flags = append(flags, "-I", filepath.Dir(library), "-o", mutant, source)
	flags = append(flags, RuntimeLinkFlags(library)...)
	flags = append(flags, "-lm")
	if output, combinedOutputError := exec.Command("clang", flags...).CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("mutant compile: %v\n%s", combinedOutputError, output)
	}
	got := runWithInput(test, "", mutant)
	if got == string(want) {
		test.Fatal("fixture cannot see a link-only -ffp-contract=off omission")
	}
	test.Logf("FMA target -march=%q; Node and ThinLTO %q; link-option omission %q", cpu, want, got)
}

func TestReleaseSplitFixtureAgreesWithNode(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../oracle/testdata/release_fma.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	checked, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	binary := filepath.Join(test.TempDir(), "split")
	if buildError := Build(C(program), binary, Options{Release: true, Split: true}); buildError != nil {
		test.Fatal(buildError)
	}
	if got := runWithInput(test, "", binary); got != "0\n0\n0\n" {
		test.Fatalf("split release rounding: %q", got)
	}
}
