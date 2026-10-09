package load

import (
	_ "embed"
	"strings"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// The prelude is served from inside the binary, at a path no disk has, so a program checks the same
// wherever adamic runs.
const (
	preludeDirectory = "/adamic-prelude"
	preludePath      = preludeDirectory + "/adamic.d.ts"
	setPreludePath   = preludeDirectory + "/set.d.ts"
)

//go:embed prelude.d.ts
var prelude string

//go:embed prelude_set.d.ts
var setPrelude string

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
	overlay        map[tspath.RootedFilePath]string
	projectConsole bool
	nodeTypes      bool
}

// adamicFile is the .a file behind a path the checker asked for, when there is one and no real .ts
// file of that name.
func (s *sourceFS) adamicFile(path tspath.RootedFilePath) (tspath.RootedFilePath, bool) {
	if !strings.HasSuffix(path.AsString(), ".a.ts") || s.FS.FileExists(path) {
		return "", false
	}
	adamicPath := path.RemoveExtension(".ts").AppendSuffix("")
	return adamicPath, s.FS.FileExists(adamicPath)
}

// displayName is the name a person wrote, for a file name the checker uses.
func (s *sourceFS) displayName(path tspath.RootedFilePath) string {
	if adamicPath, isAdamic := s.adamicFile(path); isAdamic {
		return adamicPath.AsString()
	}
	return path.AsString()
}

func (s *sourceFS) FileExists(path tspath.RootedFilePath) bool {
	if path == preludePath || path == setPreludePath {
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

func (s *sourceFS) ReadFile(path tspath.RootedFilePath) (string, bool) {
	if path == preludePath {
		if s.nodeTypes {
			return nodePrelude(), true
		}
		if s.projectConsole {
			// Project console declarations can accept arbitrary data. This true
			// runtime signature must not shadow an inherited Node console method
			// with the narrower standalone Adamic signature.
			text := strings.ReplaceAll(prelude, "log(message: string | null | undefined): void;", "log(...data: any[]): void;")
			text = strings.ReplaceAll(text, "error(message: string | null | undefined): void;", "error(...data: any[]): void;")
			return text, true
		}
		return prelude, true
	}
	if path == setPreludePath {
		return setPrelude, true
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

func (s *sourceFS) DirectoryExists(path tspath.RootedDirectoryPath) bool {
	return path == preludeDirectory || s.FS.DirectoryExists(path)
}

func (s *sourceFS) Stat(path tspath.RootedPath) vfs.FileInfo {
	if adamicPath, isAdamic := s.adamicFile(tspath.RootedFilePathFromPath(path)); isAdamic {
		return s.FS.Stat(adamicPath.AsPath())
	}
	return s.FS.Stat(path)
}

func (s *sourceFS) Realpath(path tspath.RootedPath) tspath.RootedPath {
	if path == preludePath || path == setPreludePath {
		return path
	}
	if adamicPath, isAdamic := s.adamicFile(tspath.RootedFilePathFromPath(path)); isAdamic {
		return tspath.RootedFilePathFromPath(s.FS.Realpath(adamicPath.AsPath())).AppendSuffix(".ts").AsPath()
	}
	return s.FS.Realpath(path)
}

// The checker never writes. These refuse rather than pass through, so a write can't land on disk by
// accident through a file system that exists to read.

func (s *sourceFS) WriteFile(path tspath.RootedFilePath, data string) error {
	return errReadOnly
}

func (s *sourceFS) AppendFile(path tspath.RootedFilePath, data string) error {
	return errReadOnly
}

func (s *sourceFS) Remove(path tspath.RootedPath) error {
	return errReadOnly
}

func (s *sourceFS) Chtimes(path tspath.RootedPath, aTime time.Time, mTime time.Time) error {
	return errReadOnly
}
