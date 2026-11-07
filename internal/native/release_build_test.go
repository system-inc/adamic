package native

import (
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"slices"
	"strings"
	"testing"
)

// -Wcast-function-type-strict joined with the closure convention (area-next): a closure called
// through the wrong code type fails to compile instead of running.
const historicalCommonFlags = "-std=c11 -Wall -Wextra -Werror -Wcast-function-type-strict -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -pthread"

// These literals pin order and bytes to the pre-LTO policy, independently of Flags.
func TestNonShippingFlagsStayIdentical(test *testing.T) {
	test.Parallel()
	for _, release := range []bool{false, true} {
		for _, row := range []struct {
			name    string
			options Options
			suffix  string
		}{
			{"counted", Options{Release: release, Count: true}, " -DADAMIC_COUNT -O2"},
			{
				"thread sanitized",
				Options{Release: release, ThreadSanitize: true},
				" -O1 -g -fsanitize=thread -DADAMIC_TSAN_TEST",
			},
			{
				"sanitized",
				Options{Release: release, Sanitize: true},
				" -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all -fsanitize-undefined-strip-path-components=-1",
			},
			{
				"oracle counts",
				Options{Release: release, Sanitize: true, Count: true},
				" -DADAMIC_COUNT -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all -fsanitize-undefined-strip-path-components=-1",
			},
			{
				"slab probe",
				Options{Release: release, Sanitize: true, Count: true, slabs: true, cpu: "haswell"},
				" -DADAMIC_COUNT -DADAMIC_SLABS -march=haswell -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all -fsanitize-undefined-strip-path-components=-1",
			},
		} {
			want := strings.ReplaceAll(historicalCommonFlags+row.suffix, " ", "\x00")
			for _, got := range [][]string{Flags(row.options), LinkFlags(row.options)} {
				if strings.Join(got, "\x00") != want {
					test.Errorf("%s release=%v: got %q, want %q", row.name, release, got, want)
				}
			}
		}
	}
	want := strings.ReplaceAll(historicalCommonFlags+" -O2", " ", "\x00")
	for _, got := range [][]string{Flags(Options{}), LinkFlags(Options{})} {
		if strings.Join(got, "\x00") != want {
			test.Errorf("ordinary oracle/test flags changed: %q", got)
		}
	}
}

func TestReleaseFlagsReachLink(test *testing.T) {
	test.Parallel()
	options := Options{Release: true}
	compile, link := Flags(options), LinkFlags(options)
	if !slices.Contains(compile, "-flto=thin") {
		test.Fatal("shipped release is not ThinLTO")
	}
	if len(link) < len(compile) || !slices.Equal(compile, link[:len(compile)]) {
		test.Fatal("link lost or reordered compilation flags")
	}
	if goruntime.GOOS != "darwin" && !slices.Contains(link, "-fuse-ld=lld") {
		test.Fatal("release link is not using lld")
	}
}

func TestReleaseArithmeticIsNeverFused(test *testing.T) {
	test.Parallel()
	cpu, why, ok := fusingProcessor()
	if !ok {
		test.Skip(why)
	}
	binary := filepath.Join(test.TempDir(), "release")
	if buildError := Build(fusedHarness, binary, Options{Release: true, cpu: cpu}); buildError != nil {
		test.Fatal(buildError)
	}
	if got := runWithInput(test, "", binary); got != "0 0 0\n" {
		test.Fatalf("shipped arithmetic differs: %q", got)
	}
	// This changes only the actual program compilation/link command, not the runtime.
	// With the semantic option omitted, clang's default contraction must be observable.
	source := filepath.Join(test.TempDir(), "fused.c")
	if writeError := os.WriteFile(source, []byte(fusedHarness), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	library, runtimeLibraryError := RuntimeLibrary("", Options{Release: true, cpu: cpu})
	if runtimeLibraryError != nil {
		test.Fatal(runtimeLibraryError)
	}
	flags := slices.DeleteFunc(
		LinkFlags(Options{Release: true, cpu: cpu}),
		func(flag string) bool { return flag == "-ffp-contract=off" },
	)
	flags = append(flags, "-o", binary, source)
	flags = append(flags, RuntimeLinkFlags(library)...)
	flags = append(flags, "-lm")
	if output, combinedOutputError := exec.Command("clang", flags...).CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("mutant compile: %v\n%s", combinedOutputError, output)
	}
	if got := runWithInput(test, "", binary); got == "0 0 0\n" {
		test.Fatal("link contraction-option mutant survived")
	}
	test.Log("link-only -ffp-contract=off omission caught by arithmetic output")
}
