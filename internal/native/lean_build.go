package native

import (
	"fmt"
	"os"
	"path/filepath"
)

// BuildLean compiles the C at sourcePath in one translation unit and links it with the runtime archive in
// archiveDirectory, runtime.a beside the runtime's headers as RuntimeRequest.Build fills a directory, into a native
// binary at output, each clang holding a token of jobserver (nil bounds none). Handed the archive Build would link, it
// writes Build's bytes (internal/lower's TestNativeAgreementLeanBuildIsBuildsBytes).
//
// It's Build for a lean product (buildcache.LeanProductOf), whose build is a function of what it's handed alone. Build
// reads the process's runtime supplier or the runtime cache for its archive, the process's jobserver and
// ADAMIC_NATIVE_SPLIT; BuildLean reads none of them, nor any other environment, so it refuses what would: a target
// (WASI_SYSROOT), a split build (the unit cache), and a program linking the regexp compiler's archive, which it isn't
// handed. runtimeproducts' TestTheLeanBuildReachesOnlyItsDeclaredFiles holds its reach to agreementproducts.NativeFiles.
func BuildLean(sourcePath string, archiveDirectory string, output string, options Options, jobserver *Jobserver) error {
	if options.Target != "" || options.Request || options.Split {
		return fmt.Errorf("native: a lean build is one native translation unit, with no target, request or split: %+v", options)
	}
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("native: %w", err)
	}
	if regexpCompilerSource(string(source)) {
		return fmt.Errorf("native: %s links the regexp compiler's archive, which a lean build isn't handed", sourcePath)
	}
	return linkOne(string(source), output, options, Flags(options), filepath.Join(archiveDirectory, "runtime.a"), jobserver, nil)
}
