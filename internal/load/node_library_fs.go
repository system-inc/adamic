package load

import (
	"embed"
	"io/fs"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// Both packages are unchanged npm publications. undici-types is the locked
// dependency referenced by Node's globals; it must also work without npm.
//
//go:embed node_types/node_modules
var embeddedNodeTypes embed.FS

const nodeBundleDirectory = "bundled:///node"

// nodeLibraryFS extends the bundled library view with the pinned Node package.
// Embedded paths take precedence over disk and caller-provided source overlays.
type nodeLibraryFS struct{ vfs.FS }

func nodeBundleName(path string) (string, bool) {
	if path == nodeBundleDirectory {
		return "node_types", true
	}
	suffix, ok := strings.CutPrefix(path, nodeBundleDirectory+"/")
	return "node_types/" + suffix, ok
}

func callerNodeTypes(path string) bool {
	return !strings.HasPrefix(path, nodeBundleDirectory+"/") &&
		(strings.HasSuffix(path, "/node_modules/@types/node") || strings.Contains(path, "/node_modules/@types/node/"))
}

func (s *nodeLibraryFS) FileExists(path tspath.RootedFilePath) bool {
	if name, bundled := nodeBundleName(path.AsString()); bundled {
		info, err := fs.Stat(embeddedNodeTypes, name)
		return err == nil && !info.IsDir()
	}
	return !callerNodeTypes(path.AsString()) && s.FS.FileExists(path)
}

func (s *nodeLibraryFS) ReadFile(path tspath.RootedFilePath) (string, bool) {
	if name, bundled := nodeBundleName(path.AsString()); bundled {
		data, err := embeddedNodeTypes.ReadFile(name)
		return string(data), err == nil
	}
	if callerNodeTypes(path.AsString()) {
		return "", false
	}
	return s.FS.ReadFile(path)
}

func (s *nodeLibraryFS) DirectoryExists(path tspath.RootedDirectoryPath) bool {
	if name, bundled := nodeBundleName(path.AsString()); bundled {
		info, err := fs.Stat(embeddedNodeTypes, name)
		return err == nil && info.IsDir()
	}
	return !callerNodeTypes(path.AsString()) && s.FS.DirectoryExists(path)
}

func (s *nodeLibraryFS) GetAccessibleEntries(path tspath.RootedDirectoryPath) vfs.Entries {
	if name, bundled := nodeBundleName(path.AsString()); bundled {
		var result vfs.Entries
		entries, err := embeddedNodeTypes.ReadDir(name)
		if err != nil {
			return result
		}
		for _, entry := range entries {
			if entry.IsDir() {
				result.Directories = append(result.Directories, entry.Name())
			} else {
				result.Files = append(result.Files, entry.Name())
			}
		}
		return result
	}
	if callerNodeTypes(path.AsString()) {
		return vfs.Entries{}
	}
	entries := s.FS.GetAccessibleEntries(path)
	if strings.HasSuffix(path.AsString(), "/node_modules/@types") {
		var directories []string
		for _, name := range entries.Directories {
			if name != "node" {
				directories = append(directories, name)
			}
		}
		entries.Directories = directories
	}
	return entries
}

func (s *nodeLibraryFS) Stat(path tspath.RootedPath) vfs.FileInfo {
	if name, bundled := nodeBundleName(path.AsString()); bundled {
		info, _ := fs.Stat(embeddedNodeTypes, name)
		return info
	}
	if callerNodeTypes(path.AsString()) {
		return nil
	}
	return s.FS.Stat(path)
}

func (s *nodeLibraryFS) Realpath(path tspath.RootedPath) tspath.RootedPath {
	if _, bundled := nodeBundleName(path.AsString()); bundled || callerNodeTypes(path.AsString()) {
		return path
	}
	return s.FS.Realpath(path)
}
