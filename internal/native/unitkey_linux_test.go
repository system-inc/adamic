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
func TestTheSearchProbeListsNothingUnderOpt(t *testing.T) {
	t.Parallel()
	tracer, err := exec.LookPath("strace")
	if err != nil {
		t.Skip("no strace on this box to watch clang's reads")
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	probe := t.TempDir()
	arguments := slices.Concat([]string{"-f", "-qq", "-e", "trace=openat", compiler}, Flags(Options{Debug: true, Sanitize: true}), unitSearchProbe(probe))
	command := exec.Command(tracer, arguments...)
	command.Dir = probe
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("strace %v: %v\n%s", arguments, err, output)
	}
	if opened := regexp.MustCompile(`(?m)^.*openat\([^,]*, "/opt(/[^"]*)?".*$`).FindAll(output, -1); len(opened) > 0 {
		t.Errorf("the search probe opened under /opt:\n%s", opened)
	}
	if !regexp.MustCompile(`(?m)^#include <\.\.\.> search starts here:$`).Match(output) {
		t.Errorf("the probe printed no search path:\n%s", output)
	}
}
