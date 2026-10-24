// Package testtools locates setup.sh's WASI SDK for Worker integration tests.
package testtools

import (
	"os"
	"path/filepath"
)

// ConfigureWASI runs before testing starts, never in a parallel test. It only
// selects tools, and builds nothing. An explicit sysroot always wins; otherwise
// use setup.sh's layout under ADAMIC_TOOLS or next to a tools directory on PATH.
// This also works under env -i, where setup's PATH survives but env.sh does not.
func ConfigureWASI() {
	if os.Getenv("WASI_SYSROOT") != "" {
		return
	}
	var roots []string
	if tools := os.Getenv("ADAMIC_TOOLS"); tools != "" {
		roots = append(roots, filepath.Join(tools, "wasi-sdk/share/wasi-sysroot"))
	} else {
		for _, directory := range filepath.SplitList(os.Getenv("PATH")) {
			roots = append(roots, filepath.Join(directory, "../wasi-sdk/share/wasi-sysroot"), filepath.Join(directory, "../share/wasi-sysroot"))
		}
	}
	for _, root := range roots {
		root = filepath.Clean(root)
		if info, err := os.Stat(root); err == nil && info.IsDir() {
			if err := os.Setenv("WASI_SYSROOT", root); err != nil {
				panic(err)
			}
			return
		}
	}
}
