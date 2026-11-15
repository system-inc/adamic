package native

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// The search path probe opens no directory under /opt. Without --rocm-path, clang's Linux driver lists /opt under -v
// to look for a ROCm installation, a read a traced product build refuses as one no key can name (Adam's tree 4efff9dd,
// 414 product tests fell back untraced). Traced by strace where the box has it, under the flags a sanitized Debug
// build hands the probe, -Werror among them. An open under clang's own install root is the compiler loading itself, not
// a listing: the rented image installs its toolchain under /opt/adamic-block, where the loader opens clang's libraries
// (#9vcrbpg). The exemption covers clang's install-relative ROCm candidates too (its install directory and the one
// above), which sit inside /opt/adamic-block/tools there, so on the rented image this test no longer guards those
// reads: harmless while traced product builds run only on Workshop's tree builder, where clang isn't under /opt, and
// worth revisiting before cmd/traced runs on a rented machine. Mutant, red on a rented machine and a house box alike:
// the probe without --rocm-path, which lists /opt itself.
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
	// clang's install root is the directory above its bin, as found and with links resolved, since the loader may name
	// either.
	resolved, err := filepath.EvalSymlinks(compiler)
	if err != nil {
		t.Fatal(err)
	}
	installRoots := []string{filepath.Dir(filepath.Dir(compiler)), filepath.Dir(filepath.Dir(resolved))}
	var opened []string
	for _, match := range regexp.MustCompile(`(?m)^.*openat\([^,]*, "(/opt(?:/[^"]*)?)".*$`).FindAllSubmatch(output, -1) {
		path := filepath.Clean(string(match[1]))
		if slices.ContainsFunc(installRoots, func(root string) bool {
			return root != "/" && root != "/opt" && (path == root || strings.HasPrefix(path, root+"/"))
		}) {
			continue
		}
		opened = append(opened, string(match[0]))
	}
	if len(opened) > 0 {
		t.Errorf("the search probe opened under /opt, outside clang's install root %s:\n%s", installRoots, opened)
	}
	if !regexp.MustCompile(`(?m)^#include <\.\.\.> search starts here:$`).Match(output) {
		t.Errorf("the probe printed no search path:\n%s", output)
	}
}
