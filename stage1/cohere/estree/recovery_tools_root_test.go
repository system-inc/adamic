package estree

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

// The sanitized port build's clang maps where clang is installed to /adamic-compiler, after the build's own maps of its
// work directory. The port builds without -g, and a sanitized probe built bare from two tools roots matched byte for
// byte on this box: its bytes don't name the tools root today, so a two-root byte pin could not fail here. This pins
// the command instead, read independently of native's toolMaps: the installation is clang's resolved path two
// directories up. Mutant: recoveryClang returning toolchain.CommandIn(work, "clang", args...), red.
func TestRecoveryClangMapsTheToolsRoot(t *testing.T) {
	t.Parallel()
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	real, err := filepath.EvalSymlinks(compiler)
	if err != nil {
		t.Fatal(err)
	}
	installation := filepath.Dir(filepath.Dir(real))
	if installation == filepath.Dir(installation) {
		t.Skipf("clang at %s has no installation directory to map", real)
	}
	work := t.TempDir()
	own := []string{"-ffile-prefix-map=" + work + "=/adamic-estree", "-fdebug-prefix-map=" + work + "=/adamic-estree"}
	args := append(append(native.Flags(native.Options{Sanitize: true}), own...), "-I", "runtime", "port.c", "-o", "port", "-lm")
	command := recoveryClang(work, args...)
	tools := "-ffile-prefix-map=" + installation + "=/adamic-compiler"
	at := slices.Index(command.Args, tools)
	if at < 0 {
		t.Fatalf("the port build's clang maps no tools root: %q lacks %s", command.Args, tools)
	}
	for _, flag := range own {
		if index := slices.Index(command.Args, flag); index < 0 || index > at {
			t.Fatalf("the build's own map %s must stay, before the tools root's: %q", flag, command.Args)
		}
	}
}
