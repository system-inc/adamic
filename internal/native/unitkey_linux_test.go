package native

import (
	"os/exec"
	"regexp"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// The search path probe opens no directory under /opt. Without --rocm-path, clang's Linux driver lists /opt under -v
// to look for a ROCm installation, a read a traced product build refuses as one no key can name (Adam's tree 4efff9dd,
// 414 product tests fell back untraced). Traced by strace where the box has it, under the flags a sanitized Debug
// build hands the probe, -Werror among them. Mutant, red: the probe without --rocm-path.
func TestTheSearchProbeListsNothingUnderOpt(test *testing.T) {
	test.Parallel()
	tracer, lookPathError := exec.LookPath("strace")
	if lookPathError != nil {
		test.Skip("no strace on this box to watch clang's reads")
	}
	compiler, pathError := toolchain.Path("clang")
	if pathError != nil {
		test.Fatal(pathError)
	}
	probe := test.TempDir()
	arguments := slices.Concat(
		[]string{"-f", "-qq", "-e", "trace=openat", compiler},
		Flags(Options{Debug: true, Sanitize: true}),
		unitSearchProbe(probe),
	)
	command := exec.Command(tracer, arguments...)
	command.Dir = probe
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("strace %v: %v\n%s", arguments, combinedOutputError, output)
	}
	if opened := regexp.MustCompile(`(?m)^.*openat\([^,]*, "/opt(/[^"]*)?".*$`).FindAll(output, -1); len(opened) > 0 {
		test.Errorf("the search probe opened under /opt:\n%s", opened)
	}
	if !regexp.MustCompile(`(?m)^#include <\.\.\.> search starts here:$`).Match(output) {
		test.Errorf("the probe printed no search path:\n%s", output)
	}
}
