// Package location is where the build cache lives on this machine. It is a package of its own so internal/native,
// which every Adamic compiler links, keeps its caches under the same directory buildcache keeps products in without
// linking buildcache itself (testing, net/http and the tracer come with it).
package location

import (
	"os"
	"path/filepath"
)

// BuildCache is ADAMIC_BUILD_CACHE_DIR, or the user cache directory's adamic-build. That a sandboxed build must name
// its cache is the toolchain resolver's rule (#eehjrv6, Builder's), for when it lands.
func BuildCache() (string, error) {
	if directory := os.Getenv("ADAMIC_BUILD_CACHE_DIR"); directory != "" {
		return filepath.Abs(directory)
	}
	user, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(user, "adamic-build"), nil
}
