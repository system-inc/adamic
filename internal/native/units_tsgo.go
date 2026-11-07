package native

import (
	bridge "github.com/system-inc/adamic/bridge/tsgo"
	"github.com/system-inc/adamic/internal/toolchain"
)

// BuildSplitTSGo is an explicit prototype entrypoint for TSGoC output. BuildTSGo is unchanged;
// its owner can decide whether to dispatch here when Options.Split is requested.
func BuildSplitTSGo(source, output, archive string, options Options) error {
	archive, checkerArchivePathError := checkerArchivePath(archive)
	if checkerArchivePathError != nil {
		return checkerArchivePathError
	}
	library, splitTsGoRuntimeError := splitTSGoRuntime(source, options)
	if splitTsGoRuntimeError != nil {
		return splitTsGoRuntimeError
	}
	return buildUnitsWithLibrary(source, output, options, library, archive)
}

// Reuse the existing runtime cache with the checker ABI header in its snapshot, and ADAMIC_TSGO
// plus the program's features in its flags. The ordinary runtime's empty tsgo.c object cannot
// satisfy these calls, and a runtime built without the features has other layouts than the units.
func splitTSGoRuntime(source string, options Options) (string, error) {
	files, readRuntimeError := readRuntime(runtime, "runtime")
	if readRuntimeError != nil {
		return "", readRuntimeError
	}
	files = append(files, runtimeFile{"tsgo.h", bridge.Header})
	compiler, pathError := toolchain.Path("clang")
	if pathError != nil {
		return "", pathError
	}
	version, combinedOutputError := toolchain.Command(compiler, "--version").CombinedOutput()
	if combinedOutputError != nil {
		return "", combinedOutputError
	}
	cache, cacheDirectoryError := cacheDirectory("native-runtime")
	if cacheDirectoryError != nil {
		return "", cacheDirectoryError
	}
	flags := append(sourceFlags(source, options), "-DADAMIC_TSGO")
	return cachedRuntime(files, flags, compiler, string(version), cache)
}
