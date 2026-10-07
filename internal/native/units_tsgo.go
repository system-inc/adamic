package native

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	bridge "github.com/system-inc/adamic/bridge/tsgo"
)

// BuildSplitTSGo is an explicit prototype entrypoint for TSGoC output. BuildTSGo is unchanged;
// its owner can decide whether to dispatch here when Options.Split is requested.
func BuildSplitTSGo(source, output, archive string, options Options) error {
	archive, err := filepath.Abs(archive)
	if err != nil {
		return err
	}
	if _, err := os.Stat(archive); err != nil {
		return fmt.Errorf("native: checker archive: %w", err)
	}
	library, err := splitTSGoRuntime(options)
	if err != nil {
		return err
	}
	return buildUnitsWithLibrary(source, output, options, library, []string{archive, "-lpthread", "-ldl"})
}

// Reuse the existing runtime cache with the checker ABI header in its snapshot and ADAMIC_TSGO
// in its flags. The ordinary runtime's empty tsgo.c object cannot satisfy these calls.
func splitTSGoRuntime(options Options) (string, error) {
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
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	flags := append(Flags(options), "-DADAMIC_TSGO")
	return cachedRuntime(files, flags, compiler, string(version), filepath.Join(cache, "adamic", "runtime"))
}
