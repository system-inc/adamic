package program

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// overlayFS answers for a few files from memory and passes everything else to the disk.
//
// It exists for an editor's save: the buffer being saved is newer than the file on disk, and the
// fixes and formatting that land must be computed from the buffer, against a type graph that holds
// the buffer, or they would be offsets into text that is about to be replaced. Only reading and
// existence are answered here. Nothing in a build writes, and a directory listing still comes from the
// disk, which is right for a file that exists there and only misses a buffer that was never saved.
type overlayFS struct {
	vfs.FS
	files map[string]string
}

// newOverlayFS keys the overlay by the normalized path the compiler asks with.
func newOverlayFS(inner vfs.FS, overlay map[string]string) vfs.FS {
	files := make(map[string]string, len(overlay))
	for path, contents := range overlay {
		files[tspath.NormalizePath(path)] = contents
	}
	return &overlayFS{FS: inner, files: files}
}

func (overlay *overlayFS) FileExists(path tspath.RootedFilePath) bool {
	if _, present := overlay.files[tspath.NormalizePath(path.AsString())]; present {
		return true
	}
	return overlay.FS.FileExists(path)
}

func (overlay *overlayFS) ReadFile(path tspath.RootedFilePath) (string, bool) {
	if contents, present := overlay.files[tspath.NormalizePath(path.AsString())]; present {
		return contents, true
	}
	return overlay.FS.ReadFile(path)
}
