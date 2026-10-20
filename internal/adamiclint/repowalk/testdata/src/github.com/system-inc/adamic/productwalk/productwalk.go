// A product directory from internal/buildcache is outside every package, but the rule can't trace it, so it is not
// flagged; this package sits under adamic's path because only adamic may import its internal packages.
package productwalk

import (
	"io/fs"
	"path/filepath"

	"github.com/system-inc/adamic/internal/buildcache"
)

func productDirectory() {
	_ = filepath.WalkDir(buildcache.Product("native-runtime"), func(path string, entry fs.DirEntry, err error) error {
		return err
	})
}
