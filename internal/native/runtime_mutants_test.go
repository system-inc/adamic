package native

import (
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// runtimeMutantClang is the clang a runtime mutant product (primitiveMutants, recordMutants) compiles and links with:
// clangIn, so after the builder's own maps, which take its scratch directories out, come the maps that take the tools
// root out. A bare toolchain.CommandIn here mapped the build directory and not where clang is installed, so two boxes
// whose tools sit apart could build two sets of bytes for one key (#dc483fc).
func runtimeMutantClang(directory string, arguments ...string) *exec.Cmd {
	return clangIn(directory, "clang", arguments...)
}

// A runtime mutant product's clang maps where clang is installed to /adamic-compiler, after the builder's own map of its
// directory. These products build without -g, and a sanitized probe built bare from two tools roots (the same clang
// reached through a second installation directory) matched byte for byte on this box: their bytes don't name the tools
// root today, so a two-root byte pin could not fail here. This pins the command instead, read independently of
// toolMaps: the installation is clang's resolved path two directories up. Mutant: runtimeMutantClang returning
// toolchain.CommandIn(directory, "clang", arguments...), red.
func TestRuntimeMutantClangMapsTheToolsRoot(t *testing.T) {
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
	directory := t.TempDir()
	own := prefixMaps(directory, "/adamic-runtime")
	arguments := append(append(Flags(Options{Sanitize: true, Count: true}), own...), "-c", "exceptions.c", "-o", "exceptions.o")
	command := runtimeMutantClang(directory, arguments...)
	tools := "-ffile-prefix-map=" + installation + "=/adamic-compiler"
	at := slices.Index(command.Args, tools)
	if at < 0 {
		t.Fatalf("a runtime mutant's clang maps no tools root: %q lacks %s", command.Args, tools)
	}
	for _, flag := range own {
		if index := slices.Index(command.Args, flag); index < 0 || index > at {
			t.Fatalf("the builder's own map %s must stay, before the tools root's: %q", flag, command.Args)
		}
	}
	for _, flag := range command.Args {
		if strings.HasPrefix(flag, "-ffile-prefix-map=") && flag != tools && strings.HasPrefix(flag, "-ffile-prefix-map="+installation+"=") {
			t.Fatalf("the tools root is mapped twice, to different names: %q", command.Args)
		}
	}
}
