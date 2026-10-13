// Package location is where the build cache lives on this machine. It is a package of its own so internal/native,
// which every Adamic compiler links, keeps its caches under the same directory buildcache keeps products in without
// linking buildcache itself (testing, net/http and the tracer come with it).
package location

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/system-inc/adamic/internal/toolchain"
)

// BuildCache is ADAMIC_BUILD_CACHE_DIR, or the user cache directory's adamic-build. In a sandbox (ADAMIC_SANDBOX=1) it
// is ADAMIC_BUILD_CACHE_DIR or an error: products are never written under a home the sandbox didn't declare.
func BuildCache() (string, error) {
	if directory := os.Getenv("ADAMIC_BUILD_CACHE_DIR"); directory != "" {
		return filepath.Abs(directory)
	}
	if toolchain.Sandboxed() {
		return "", errors.New("ADAMIC_SANDBOX=1 keeps products only where ADAMIC_BUILD_CACHE_DIR says, and ADAMIC_BUILD_CACHE_DIR is unset")
	}
	user, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(user, "adamic-build"), nil
}
