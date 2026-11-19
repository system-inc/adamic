package native

import (
	bridge "github.com/system-inc/adamic/bridge/tsgo"
	"github.com/system-inc/adamic/internal/native_runtime"
	"github.com/system-inc/adamic/internal/toolchain"
)

// BuildSplitTSGo is an explicit prototype entrypoint for TSGoC output. BuildTSGo is unchanged;
// its owner can decide whether to dispatch here when Options.Split is requested.
func BuildSplitTSGo(source, output, archive string, options native_runtime.Options) error {
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
func splitTSGoRuntime(source string, options native_runtime.Options) (string, error) {
	files, err := native_runtime.ReadRuntime(native_runtime.Sources, "c")
	if err != nil {
		return "", err
	}
	files = append(files, native_runtime.RuntimeFile{Name: "tsgo.h", Contents: bridge.Header})
	compiler, err := toolchain.Path("clang")
	if err != nil {
		return "", err
	}
	// Under the build's triple, which picks clang's config files (#5xsd2ha), as the runtime's compiles pass it.
	resolved, err := native_runtime.ResolveTarget(options)
	if err != nil {
		return "", err
	}
	version, err := toolchain.Command(compiler, "--target="+resolved.Target, "--version").CombinedOutput()
	if err != nil {
		return "", err
	}
	cache, err := native_runtime.CacheDirectory("native-runtime")
	if err != nil {
		return "", err
	}
	flags := append(native_runtime.SourceFlags(source, options), "-DADAMIC_TSGO")
	return native_runtime.CachedRuntime(files, flags, compiler, string(version), cache)
}
