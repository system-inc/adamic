// Package executable publishes executable copies without leaking writable descriptors into sibling forks.
// Keeping this helper and its fork-isolation proof separate lets Loom select a small declared compiler consumer.
package executable

import (
	"os"
	"path/filepath"
	"syscall"
)

// Write excludes forks until the writable descriptor is closed; rename alone preserves its inode.
// A sibling fork can inherit the descriptor and make exec fail with ETXTBSY (golang/go#22315).
// This follows main's 80640d7a executable writer.
func Write(path string, data []byte) error {
	var temporary string
	defer func() {
		if temporary != "" {
			os.Remove(temporary)
		}
	}()
	err := func() error {
		syscall.ForkLock.Lock()
		defer syscall.ForkLock.Unlock()
		file, err := os.CreateTemp(filepath.Dir(path), ".executable-*")
		if err != nil {
			return err
		}
		temporary = file.Name()
		if _, err := file.Write(data); err != nil {
			file.Close()
			return err
		}
		if err := file.Chmod(0755); err != nil {
			file.Close()
			return err
		}
		return file.Close()
	}()
	if err != nil {
		return err
	}
	return os.Rename(temporary, path)
}
