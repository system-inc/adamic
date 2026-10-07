package load

import (
	"embed"
	"io/fs"
	"path"
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// Apple's frameworks reach Adamic through binding files: declarations of what an apple/ module
// exports, each tagged with the Objective-C it calls (docs/apple.md). They're served like the
// prelude, from inside the binary at a path no disk has, and a program loads only the ones it
// imports.
//
// These are the seed for the bridge's first proof, written by hand. The generator (#qxe07rq) writes
// them from the SDK's headers on the Mac that builds, and replaces this directory.
const appleDirectory = "/adamic-apple"

//go:embed apple
var appleBindings embed.FS

// appleImport is an import of an apple/ module, as written in a source or a binding file. A match
// inside a comment or a string only loads a binding file the program didn't need, which costs a
// parse and changes nothing the checker proves.
var appleImport = regexp.MustCompile(`(?:from|import)\s*['"](apple/[a-z0-9/-]+)['"]`)

// relativeImport is an import of one of the program's own modules.
var relativeImport = regexp.MustCompile(`(?:from|import)\s*['"](\.{1,2}/[^'"]+)['"]`)

// appleRoots is every binding file the program reaches, as a root for the checker: the apple/
// modules its sources import, followed through its own imports and the binding files' imports of
// each other. It adds each one to overlay, which is how the source file system serves it. A module
// with no binding file is left out, and the checker says it can't find it.
func appleRoots(source *sourceFS, roots []string, overlay map[string]string) []string {
	added := []string{}
	seen := map[string]bool{}
	var visit func(fileName string, text string)
	visit = func(fileName string, text string) {
		for _, match := range appleImport.FindAllStringSubmatch(text, -1) {
			// apple/appkit/window is embedded as apple/appkit/window.d.ts and served at
			// /adamic-apple/appkit/window.d.ts.
			embedded := match[1] + ".d.ts"
			bindingPath := "/adamic-" + embedded
			if seen[bindingPath] {
				continue
			}
			seen[bindingPath] = true
			binding, err := appleBindings.ReadFile(embedded)
			if err != nil {
				continue
			}
			overlay[bindingPath] = string(binding)
			added = append(added, bindingPath)
			visit(bindingPath, string(binding))
		}
		if strings.HasPrefix(fileName, appleDirectory+"/") {
			return
		}
		for _, match := range relativeImport.FindAllStringSubmatch(text, -1) {
			imported := tspath.CombinePaths(path.Dir(fileName), match[1])
			if seen[imported] {
				continue
			}
			seen[imported] = true
			if text, exists := source.ReadFile(imported); exists {
				visit(imported, text)
			}
		}
	}
	for _, root := range roots {
		seen[root] = true
		// The checker knows a .a file as .a.ts; its text is the .a's.
		if text, exists := source.ReadFile(root); exists {
			visit(root, text)
		}
	}
	return added
}

// Some apple/ modules are written in Adamic rather than declared: State, which re-renders a view
// when it's set, or a decoder over what Apple hands back. They're embedded as apple/<path>.a and
// found by the resolver where it looks for any package, under node_modules: appleFS answers for
// node_modules/apple/<path>.ts in every directory, and gives each one real path, so a module two
// directories import is one module.
const appleImplementations = "/node_modules/apple/"

type appleFS struct{ vfs.FS }

// implementation is the embedded Adamic source behind a path the resolver tries, when there is one.
func (s *appleFS) implementation(path string) (string, bool) {
	at := strings.LastIndex(path, appleImplementations)
	if at < 0 || !strings.HasSuffix(path, ".ts") {
		return "", false
	}
	module := strings.TrimSuffix(path[at+len(appleImplementations):], ".ts")
	source, err := appleBindings.ReadFile("apple/" + module + ".a")
	return string(source), err == nil
}

func (s *appleFS) FileExists(path string) bool {
	if _, found := s.implementation(path); found {
		return true
	}
	return s.FS.FileExists(path)
}

func (s *appleFS) ReadFile(path string) (string, bool) {
	if source, found := s.implementation(path); found {
		return source, true
	}
	return s.FS.ReadFile(path)
}

func (s *appleFS) DirectoryExists(path string) bool {
	if at := strings.LastIndex(path+"/", appleImplementations); at >= 0 {
		return true
	}
	if strings.HasSuffix(path, "/node_modules") {
		return true
	}
	return s.FS.DirectoryExists(path)
}

func (s *appleFS) Realpath(path string) string {
	if _, found := s.implementation(path); found {
		return path[strings.LastIndex(path, appleImplementations):]
	}
	return s.FS.Realpath(path)
}

// IsApple reports whether a declaration comes from one of Apple's binding files, so its calls are
// foreign: an Objective-C message or a C function, never an Adamic body.
func IsApple(sourceFile *ast.SourceFile) bool {
	return sourceFile != nil && strings.HasPrefix(sourceFile.FileName(), appleDirectory+"/")
}

// AppleModules is every apple/ module there's a binding file for, by its import path.
func AppleModules() []string {
	modules := []string{}
	_ = fs.WalkDir(appleBindings, "apple", func(name string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() && strings.HasSuffix(name, ".d.ts") {
			modules = append(modules, strings.TrimSuffix(name, ".d.ts"))
		}
		return nil
	})
	return modules
}
