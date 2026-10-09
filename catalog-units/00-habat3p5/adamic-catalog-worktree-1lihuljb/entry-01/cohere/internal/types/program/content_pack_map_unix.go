//go:build darwin || linux

package program

import (
	"os"
	"syscall"
)

// mapFile maps a file read-only for the life of the process, or returns nil. Nothing unmaps it: the pack
// serves strings over the mapping, and the compiler and the rules keep substrings of those as long as the
// process runs. See the content pack's package comment.
func mapFile(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	information, err := file.Stat()
	if err != nil || information.Size() == 0 {
		return nil
	}
	mapped, err := syscall.Mmap(int(file.Fd()), 0, int(information.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil
	}
	return mapped
}
