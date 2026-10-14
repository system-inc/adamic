// Package testsupport builds isolated runtime mutants through native's toolchain.
package testsupport

import (
	"fmt"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
	"os"
	"path/filepath"
)

// BuildMapMutant replaces only map.o in the matching shared runtime archive.
// Mutants are temporary test builds, never products. Compilation uses relative
// source names and maps both temporary and cached header paths out of debug info.
func BuildMapMutant(source, changed, output string, options native.Options) error {
	library, err := native.RuntimeLibraryForSource("", source, options)
	if err != nil {
		return err
	}
	directory := filepath.Dir(output)
	flags := native.Flags(options)
	for _, feature := range native.Features(source) {
		flags = append(flags, "-D"+feature+"=1")
	}
	flags = append(flags, "-ffile-prefix-map="+directory+"=.", "-fdebug-prefix-map="+directory+"=.", "-ffile-prefix-map="+filepath.Dir(library)+"=/runtime", "-fdebug-prefix-map="+filepath.Dir(library)+"=/runtime")
	compiler, err := toolchain.Path("clang")
	if err != nil {
		return err
	}
	run := func(program string, arguments ...string) error {
		command := toolchain.Command(program, arguments...)
		command.Dir = directory
		if result, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("mutant build: %w\n%s", err, result)
		}
		return nil
	}
	if err := os.WriteFile(filepath.Join(directory, "map.c"), []byte(changed), 0644); err != nil {
		return err
	}
	args := append(append([]string{}, flags...), "-I", filepath.Dir(library), "-c", "map.c", "-o", "map.o")
	if err := run(compiler, args...); err != nil {
		return err
	}
	contents, err := os.ReadFile(library)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "mutant.a"), contents, 0644); err != nil {
		return err
	}
	archiver, err := toolchain.Path("ar")
	if err != nil {
		return err
	}
	if err := run(archiver, "rcs", "mutant.a", "map.o"); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0644); err != nil {
		return err
	}
	args = append(append([]string{}, flags...), "-I", filepath.Dir(library), "-o", filepath.Base(output), "main.c")
	args = append(args, native.RuntimeLinkFlags("mutant.a")...)
	args = append(args, "-lm", "-pthread")
	return run(compiler, args...)
}
