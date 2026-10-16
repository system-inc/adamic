package native

import (
	bridge "github.com/system-inc/adamic/bridge/tsgo"
	"github.com/system-inc/adamic/internal/toolchain"
)

// BuildSplitTSGo builds TSGoC output as stable units with its explicitly selected
// checker archive. BuildTSGo dispatches here under the generated-source policy.
func BuildSplitTSGo(source, output, archive string, options Options) error {
	if err := ValidateOptions(options); err != nil {
		return err
	}
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
	compiler, err := toolchain.Path("clang")
	if err != nil {
		return "", err
	}
	version, err := compilerVersion(compiler)
	if err != nil {
		return "", err
	}
	cache, err := cacheDirectory("native-runtime")
	if err != nil {
		return "", err
	}
	flags := append(sourceFlags(source, options), "-DADAMIC_TSGO")
	jobs, err := splitJobs(options)
	if err != nil {
		return "", err
	}
	return cachedRuntime(files, flags, compiler, string(version), cache, jobs)
}
