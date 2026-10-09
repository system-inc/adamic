package program

import (
	"io"
	"os"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// readIdentities is the stat each file's bytes were read under, for a program built without a content pack,
// which keeps its own (#q6dey77). It lets a reader holding the program's copy of a file ask whether that copy
// is still the file's without reading it again. In memory only: a `--no-cache` run reads and writes no cache.
type readIdentities struct {
	mutex  sync.Mutex
	byPath map[string]fileIdentity
}

// identityFS stats each file before reading it, as the content pack does, so the identity it records can only
// be older than the bytes read after it, and records it when both succeed. Everything else passes through.
type identityFS struct {
	vfs.FS
	identities *readIdentities
}

func (f *identityFS) ReadFile(path tspath.RootedFilePath) (string, bool) {
	identity, statted := statIdentity(path.AsString())
	contents, ok := f.FS.ReadFile(path)
	if statted && ok {
		f.identities.mutex.Lock()
		f.identities.byPath[path.AsString()] = identity
		f.identities.mutex.Unlock()
	}
	return contents, ok
}

// identity is the stat path was read under, when one was recorded.
func (r *readIdentities) identity(path string) (fileIdentity, bool) {
	if r == nil {
		return fileIdentity{}, false
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	identity, known := r.byPath[path]
	return identity, known
}

/*
 * startsWithByteOrderMark reports whether the file's first bytes are one of the marks typescript-go's decodeBytes
 * acts on (internal/vfs/internal/internal.go, copied as decodeLikeDisk): FF FE or FE FF, whose UTF-16 it decodes,
 * or EF BB BF, which it drops. Any other file it returns byte for byte. A file that can't be opened or read
 * counts as marked, so the caller reads the disk.
 */
func startsWithByteOrderMark(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return true
	}
	defer file.Close()
	var head [3]byte
	count, err := io.ReadFull(file, head[:])
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return true
	}
	switch {
	case count >= 2 && (head[0] == 0xFF && head[1] == 0xFE || head[0] == 0xFE && head[1] == 0xFF):
		return true
	case count == 3 && head[0] == 0xEF && head[1] == 0xBB && head[2] == 0xBF:
		return true
	}
	return false
}
