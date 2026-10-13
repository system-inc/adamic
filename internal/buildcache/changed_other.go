//go:build !linux && !darwin

package buildcache

import (
	"io/fs"
	"time"
)

// changeTime is unknown here, so cleanObject hashes every file rather than trust a modification time alone.
func changeTime(info fs.FileInfo) (time.Time, bool) {
	return time.Time{}, false
}
