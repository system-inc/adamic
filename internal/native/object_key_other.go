//go:build !linux && !darwin

package native

import "io/fs"

// fileChange is unknown off Linux and macOS, so main.c gets no object key there (objectReadOf).
func fileChange(info fs.FileInfo) (uint64, int64, bool) {
	return 0, 0, false
}
