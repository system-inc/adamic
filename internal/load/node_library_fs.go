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

// nodeLibraryFileSystem extends the bundled library view with the pinned Node package.
// Embedded paths take precedence over disk and caller-provided source overlays.
type nodeLibraryFileSystem struct{ vfs.FS }

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

func (fileSystem *nodeLibraryFileSystem) FileExists(path tspath.RootedFilePath) bool {
	if name, bundled := nodeBundleName(path.AsString()); bundled {
		info, statError := fs.Stat(embeddedNodeTypes, name)
		return statError == nil && !info.IsDir()
	}
	return !callerNodeTypes(path.AsString()) && fileSystem.FS.FileExists(path)
}

func (fileSystem *nodeLibraryFileSystem) ReadFile(path tspath.RootedFilePath) (string, bool) {
	if name, bundled := nodeBundleName(path.AsString()); bundled {
		data, readFileError := embeddedNodeTypes.ReadFile(name)
		if readFileError != nil {
			return "", false
		}
		file := strings.TrimPrefix(name, "node_types/node_modules/@types/node/")
		text, correctNodeHostAxiomsError := CorrectNodeHostAxioms(file, string(data))
		return text, correctNodeHostAxiomsError == nil
	}
	if callerNodeTypes(path.AsString()) {
		return "", false
	}
	return fileSystem.FS.ReadFile(path)
}

func (fileSystem *nodeLibraryFileSystem) DirectoryExists(path tspath.RootedDirectoryPath) bool {
	if name, bundled := nodeBundleName(path.AsString()); bundled {
		info, statError := fs.Stat(embeddedNodeTypes, name)
		return statError == nil && info.IsDir()
	}
	return !callerNodeTypes(path.AsString()) && fileSystem.FS.DirectoryExists(path)
}

func (fileSystem *nodeLibraryFileSystem) GetAccessibleEntries(path tspath.RootedDirectoryPath) vfs.Entries {
	if name, bundled := nodeBundleName(path.AsString()); bundled {
		var result vfs.Entries
		entries, readDirectoryError := embeddedNodeTypes.ReadDir(name)
		if readDirectoryError != nil {
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
	entries := fileSystem.FS.GetAccessibleEntries(path)
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

func (fileSystem *nodeLibraryFileSystem) Stat(path tspath.RootedPath) vfs.FileInfo {
	if name, bundled := nodeBundleName(path.AsString()); bundled {
		info, _ := fs.Stat(embeddedNodeTypes, name)
		return info
	}
	if callerNodeTypes(path.AsString()) {
		return nil
	}
	return fileSystem.FS.Stat(path)
}

func (fileSystem *nodeLibraryFileSystem) Realpath(path tspath.RootedPath) tspath.RootedPath {
	if _, bundled := nodeBundleName(path.AsString()); bundled || callerNodeTypes(path.AsString()) {
		return path
	}
	return fileSystem.FS.Realpath(path)
}
