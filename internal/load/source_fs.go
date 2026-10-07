package load

import (
	_ "embed"
	"strings"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// The prelude is served from inside the binary, at a path no disk has, so a program checks the same
// wherever adamic runs.
const (
	preludeDirectory = "/adamic-prelude"
	preludePath      = preludeDirectory + "/adamic.d.ts"
)

//go:embed prelude.d.ts
var prelude string

// sourceFS is the disk as the checker sees an Adamic program.
//
// TypeScript knows nothing of `.a`, so every X.a is shown to it as X.a.ts. That is the whole trick:
// the resolver already turns `./geometry` into geometry.ts by trying extensions, and in the same way
// it turns `./geometry.a` into geometry.a.ts, so imports between .a files need no change to
// TypeScript at all. displayName turns the name back on the way out.
//
// A real X.a.ts always wins over the alias, and Load refuses a program where both exist, so no file
// is ever shadowed silently.
type sourceFS struct {
	vfs.FS
	overlay map[string]string
}

// adamicFile is the .a file behind a path the checker asked for, when there is one and no real .ts
// file of that name.
func (s *sourceFS) adamicFile(path string) (string, bool) {
	if !strings.HasSuffix(path, ".a.ts") || s.FS.FileExists(path) {
		return "", false
	}
	adamicPath := strings.TrimSuffix(path, ".ts")
	return adamicPath, s.FS.FileExists(adamicPath)
}

// displayName is the name a person wrote, for a file name the checker uses.
func (s *sourceFS) displayName(path string) string {
	if adamicPath, isAdamic := s.adamicFile(path); isAdamic {
		return adamicPath
	}
	return path
}

func (s *sourceFS) FileExists(path string) bool {
	if path == preludePath {
		return true
	}
	if _, isAdamic := s.adamicFile(path); isAdamic {
		return true
	}
	if _, exists := s.overlay[path]; exists {
		return true
	}
	return s.FS.FileExists(path)
}

func (s *sourceFS) ReadFile(path string) (string, bool) {
	if path == preludePath {
		return prelude, true
	}
	if source, exists := s.overlay[path]; exists {
		return source, true
	}
	if adamicPath, isAdamic := s.adamicFile(path); isAdamic {
		if source, exists := s.overlay[adamicPath]; exists {
			return source, true
		}
		return s.FS.ReadFile(adamicPath)
	}
	return s.FS.ReadFile(path)
}

func (s *sourceFS) DirectoryExists(path string) bool {
	return path == preludeDirectory || s.FS.DirectoryExists(path)
}

func (s *sourceFS) Stat(path string) vfs.FileInfo {
	if adamicPath, isAdamic := s.adamicFile(path); isAdamic {
		return s.FS.Stat(adamicPath)
	}
	return s.FS.Stat(path)
}

func (s *sourceFS) Realpath(path string) string {
	if path == preludePath {
		return path
	}
	if adamicPath, isAdamic := s.adamicFile(path); isAdamic {
		return s.FS.Realpath(adamicPath) + ".ts"
	}
	return s.FS.Realpath(path)
}

// The checker never writes. These refuse rather than pass through, so a write can't land on disk by
// accident through a file system that exists to read.

func (s *sourceFS) WriteFile(path string, data string) error {
	return errReadOnly
}

func (s *sourceFS) AppendFile(path string, data string) error {
	return errReadOnly
}

func (s *sourceFS) Remove(path string) error {
	return errReadOnly
}

func (s *sourceFS) Chtimes(path string, aTime time.Time, mTime time.Time) error {
	return errReadOnly
}
