package native

import "os/exec"

// ClangIn is clangIn for a product builder outside this package that runs clang itself: compiler (a path, or a name
// toolchain resolves) in directory, with arguments and then the maps that take the compiler's installation and any
// sysroot out of what the compile records of a path (toolMaps, #8wgk65j). The builder's own prefix maps take its own
// directories out; this takes the tools out, so a product's bytes don't depend on where the tools sit (#dc483fc).
func ClangIn(directory, compiler string, arguments ...string) *exec.Cmd {
	return clangIn(directory, compiler, arguments...)
}
