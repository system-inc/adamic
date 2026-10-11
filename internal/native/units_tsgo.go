package native

import (
	"os/exec"

	bridge "github.com/system-inc/adamic/bridge/tsgo"
)

// BuildSplitTSGo is an explicit prototype entrypoint for TSGoC output. BuildTSGo is unchanged;
// its owner can decide whether to dispatch here when Options.Split is requested.
func BuildSplitTSGo(source, output, archive string, options Options) error {
	archive, err := checkerArchivePath(archive)
	if err != nil {
		return err
	}
	library, err := splitTSGoRuntime(source, options)
	if err != nil {
		return err
	}
	return buildUnitsWithLibrary(source, output, options, library, archive)
}

// Reuse the existing runtime cache with the checker ABI header in its snapshot, and ADAMIC_TSGO
// plus the program's features in its flags. The ordinary runtime's empty tsgo.c object cannot
// satisfy these calls, and a runtime built without the features has other layouts than the units.
func splitTSGoRuntime(source string, options Options) (string, error) {
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		return "", err
	}
	files = append(files, runtimeFile{"tsgo.h", bridge.Header})
	compiler, err := exec.LookPath("clang")
	if err != nil {
		return "", err
	}
	version, err := exec.Command(compiler, "--version").CombinedOutput()
	if err != nil {
		return "", err
	}
	cache, err := cacheDirectory("native-runtime")
	if err != nil {
		return "", err
	}
	flags := append(sourceFlags(source, options), "-DADAMIC_TSGO")
	return cachedRuntime(files, flags, compiler, string(version), cache)
}
