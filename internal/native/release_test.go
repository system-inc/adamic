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

const historicalCommonFlags = "-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls"

// These literals pin order and bytes to the pre-LTO policy, independently of Flags.
func TestNonShippingFlagsStayIdentical(t *testing.T) {
	t.Parallel()
	for _, release := range []bool{false, true} {
		for _, row := range []struct {
			name    string
			options Options
			suffix  string
		}{
			{"counted", Options{Release: release, Count: true}, " -DADAMIC_COUNT -O2"},
			{"sanitized", Options{Release: release, Sanitize: true}, " -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"},
			{"oracle counts", Options{Release: release, Sanitize: true, Count: true}, " -DADAMIC_COUNT -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"},
			{"slab probe", Options{Release: release, Sanitize: true, Count: true, slabs: true, cpu: "haswell"}, " -DADAMIC_COUNT -DADAMIC_SLABS -march=haswell -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"},
		} {
			want := strings.ReplaceAll(historicalCommonFlags+row.suffix, " ", "\x00")
			for _, got := range [][]string{Flags(row.options), LinkFlags(row.options)} {
				if strings.Join(got, "\x00") != want {
					t.Errorf("%s release=%v: got %q, want %q", row.name, release, got, want)
				}
			}
		}
	}
	want := strings.ReplaceAll(historicalCommonFlags+" -O2", " ", "\x00")
	for _, got := range [][]string{Flags(Options{}), LinkFlags(Options{})} {
		if strings.Join(got, "\x00") != want {
			t.Errorf("ordinary oracle/test flags changed: %q", got)
		}
	}
}

func TestReleaseFlagsReachLink(t *testing.T) {
	t.Parallel()
	options := Options{Release: true}
	compile, link := Flags(options), LinkFlags(options)
	if !slices.Contains(compile, "-flto=thin") {
		t.Fatal("shipped release is not ThinLTO")
	}
	if len(link) < len(compile) || !slices.Equal(compile, link[:len(compile)]) {
		t.Fatal("link lost or reordered compilation flags")
	}
	if goruntime.GOOS != "darwin" && !slices.Contains(link, "-fuse-ld=lld") {
		t.Fatal("release link is not using lld")
	}
}

func TestReleaseArithmeticIsNeverFused(t *testing.T) {
	t.Parallel()
	cpu, why, ok := fusingProcessor()
	if !ok {
		t.Skip(why)
	}
	binary := filepath.Join(t.TempDir(), "release")
	if err := Build(fusedHarness, binary, Options{Release: true, cpu: cpu}); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", binary); got != "0 0 0\n" {
		t.Fatalf("shipped arithmetic differs: %q", got)
	}
	// This changes only the actual program compilation/link command, not the runtime.
	// With the semantic option omitted, clang's default contraction must be observable.
	source := filepath.Join(t.TempDir(), "fused.c")
	if err := os.WriteFile(source, []byte(fusedHarness), 0o644); err != nil {
		t.Fatal(err)
	}
	library, err := RuntimeLibrary("", Options{Release: true, cpu: cpu})
	if err != nil {
		t.Fatal(err)
	}
	flags := slices.DeleteFunc(LinkFlags(Options{Release: true, cpu: cpu}), func(s string) bool { return s == "-ffp-contract=off" })
	flags = append(flags, "-o", binary, source)
	flags = append(flags, RuntimeLinkFlags(library)...)
	flags = append(flags, "-lm")
	if output, err := exec.Command("clang", flags...).CombinedOutput(); err != nil {
		t.Fatalf("mutant compile: %v\n%s", err, output)
	}
	if got := runWithInput(t, "", binary); got == "0 0 0\n" {
		t.Fatal("link contraction-option mutant survived")
	}
	t.Log("link-only -ffp-contract=off omission caught by arithmetic output")
}
