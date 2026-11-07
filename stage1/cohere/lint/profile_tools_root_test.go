package lint

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

// secondToolsRoot is clang's installation reached at a second place: a directory of another name and length whose bin
// holds clang's own binary (a hard link, or a copy across devices) and whose every other entry, bin's siblings
// included, links to the installation's. clang finds its headers and libraries from where its binary runs, so what it
// records of them names this root instead (#8wgk65j's two-root pair). It returns the first root's clang and the
// second's.
func secondToolsRoot(test *testing.T) (string, string) {
	test.Helper()
	compiler, pathError := toolchain.Path("clang")
	if pathError != nil {
		test.Fatal(pathError)
	}
	real, resolveSymlinksError := filepath.EvalSymlinks(compiler)
	if resolveSymlinksError != nil {
		test.Fatal(resolveSymlinksError)
	}
	installation := filepath.Dir(filepath.Dir(real))
	if installation == filepath.Dir(installation) {
		test.Skipf("clang at %s has no installation directory to reach a second way", real)
	}
	root := filepath.Join(test.TempDir(), "tools-reached-at-a-second-place")
	link := func(from, to string) {
		entries, readDirectoryError := os.ReadDir(from)
		if readDirectoryError != nil {
			test.Fatal(readDirectoryError)
		}
		if createDirectoryError := os.MkdirAll(to, 0o755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		for _, entry := range entries {
			if from == filepath.Dir(real) && entry.Name() == filepath.Base(real) {
				continue
			}
			if symlinkError := os.Symlink(
				filepath.Join(from, entry.Name()),
				filepath.Join(to, entry.Name()),
			); symlinkError != nil {
				test.Fatal(symlinkError)
			}
		}
	}
	bin := filepath.Join(root, filepath.Base(filepath.Dir(real)))
	link(installation, root)
	if removeError := os.Remove(bin); removeError != nil {
		test.Fatal(removeError)
	}
	link(filepath.Dir(real), bin)
	second := filepath.Join(bin, filepath.Base(real))
	if linkError := os.Link(real, second); linkError != nil {
		contents, readError := os.ReadFile(real)
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(second, contents, 0o755); writeError != nil {
			test.Fatal(writeError)
		}
	}
	return real, second
}

// The profiled build's bytes don't depend on where the tools sit: one probe, built the profiled way (with -g, its
// directory mapped, its compilation directory named) through profiledClang with clang reached through two tools roots,
// each in its own directory, is one set of bytes. The control is the same two builds with a bare clang: they must
// differ, or this box's clang records nothing of where it sits and the test says so and skips. Mutant: profiledClang
// returning toolchain.Command(compiler, flags...), red.
func TestProfiledBuildIsTheSameBytesFromTwoToolsRoots(test *testing.T) {
	test.Parallel()
	first, second := secondToolsRoot(test)
	// stddef.h is one of clang's own headers, under its installation; an exported function's parameter types keep its
	// ptrdiff_t and size_t in the debug info at -O2, as the runtime's do.
	source := "#include <stddef.h>\n#include <stdio.h>\nptrdiff_t adamic_probe_difference(ptrdiff_t left, size_t right) { return left - (ptrdiff_t)right; }\nint main(int argc, char **argv) { (void)argv; printf(\"%td\\n\", adamic_probe_difference(argc, sizeof(max_align_t))); return 0; }\n"
	options := native.Options{Jobs: 1}
	build := func(compiler string, bare bool) []byte {
		test.Helper()
		directory := test.TempDir()
		if writeError := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
		flags := append(
			native.SourceFlags(source, options),
			"-g",
			"-ffile-prefix-map="+directory+"="+profiledCompilationDirectory,
			"-ffile-compilation-dir="+profiledCompilationDirectory,
			"-o",
			filepath.Join(directory, "profiled"),
			filepath.Join(directory, "main.c"),
			"-lm",
		)
		var command *exec.Cmd
		if bare {
			command = toolchain.Command(compiler, flags...)
		} else {
			command = profiledClang(compiler, flags...)
		}
		if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
			test.Fatalf("clang %s: %v\n%s", compiler, combinedOutputError, output)
		}
		contents, readError := os.ReadFile(filepath.Join(directory, "profiled"))
		if readError != nil {
			test.Fatal(readError)
		}
		return contents
	}
	if bytes.Equal(build(first, true), build(second, true)) {
		test.Skipf(
			"a bare clang at %s and at %s builds one set of bytes, so there is no tools root in them to map",
			first,
			second,
		)
	}
	if !bytes.Equal(build(first, false), build(second, false)) {
		test.Fatalf(
			"the profiled build from clang at %s and at %s differs: where the tools sit reaches its bytes",
			first,
			second,
		)
	}
}
