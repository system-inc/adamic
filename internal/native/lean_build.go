package native

import (
	"fmt"
	"os"
	"path/filepath"
)

// BuildLean compiles the C at sourcePath in one translation unit with compiler, an absolute path, and links it with the
// runtime archive in archiveDirectory, runtime.a beside the runtime's headers as RuntimeRequest.Build fills a directory,
// into a native binary at output, each clang holding a token of jobserver (nil bounds none). Handed the archive Build
// would link, it writes Build's bytes (internal/lower's TestNativeAgreementLeanBuildIsBuildsBytes), wherever output is
// and whatever jobserver it's handed.
//
// It's Build for a lean product (buildcache.LeanProductOf), whose build is a function of what it's handed alone. Build
// reads the process's runtime supplier or the runtime cache for its archive, the process's jobserver,
// ADAMIC_NATIVE_SPLIT, the temporary directory `TMPDIR` names for its build directory, and the clang `PATH` finds;
// BuildLean reads none of them, nor any other environment, so it refuses what would: a target (WASI_SYSROOT), a split
// build (the unit cache), a compiler named by anything but an absolute path, and a program linking the regexp
// compiler's archive, which it isn't handed. Its build directory is build, beside output, made at the start and removed
// at the end, so a trace of it reads one spelling every run. runtimeproducts' TestTheLeanBuildReachesOnlyItsDeclaredFiles
// holds its reach to agreementproducts.NativeFiles, and every program it starts to a path it's handed.
func BuildLean(compiler string, sourcePath string, archiveDirectory string, output string, options Options, jobserver *Jobserver) error {
	if options.Target != "" || options.Request || options.Split {
		return fmt.Errorf("native: a lean build is one native translation unit, with no target, request or split: %+v", options)
	}
	if !filepath.IsAbs(compiler) {
		return fmt.Errorf("native: a lean build runs the compiler it's handed by its absolute path, not %q", compiler)
	}
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("native: %w", err)
	}
	if regexpCompilerSource(string(source)) {
		return fmt.Errorf("native: %s links the regexp compiler's archive, which a lean build isn't handed", sourcePath)
	}
	directory := filepath.Join(filepath.Dir(output), "build")
	if err := os.Mkdir(directory, 0o755); err != nil {
		return fmt.Errorf("native: the lean build's directory: %w", err)
	}
	defer os.RemoveAll(directory)
	return linkOne(directory, compiler, string(source), output, options, Flags(options), filepath.Join(archiveDirectory, "runtime.a"), jobserver, nil)
}
