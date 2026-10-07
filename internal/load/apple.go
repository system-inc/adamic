package load

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/system-inc/adamic/internal/apple/generate"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// Apple's frameworks reach Adamic through binding files: declarations of what an apple/ module
// exports, each tagged with the Objective-C it calls (docs/apple.md). They're served like the
// prelude, at a path no disk has, and a program loads only the ones it imports.
//
// Foundation's, AppKit's and CoreGraphics' are generated from this Mac's SDK the first time a
// program imports one (internal/apple/generate), into a cache keyed by the SDK's build and the
// generator's version, and held to the SDK's headers by compiling their check file before any
// program sees them. The few written by hand are embedded: whole modules where nothing is
// generated (SwiftUI's, over the Swift shim; Core Foundation's run loop), and additions to a
// generated module, merged into it (Data's utf8Text, two messages to another class).
const appleDirectory = "/adamic-apple"

//go:embed apple
var appleBindings embed.FS

// generatedFrameworks are the frameworks generated from the SDK, by their import path's segment.
var generatedFrameworks = map[string]string{"appkit": "AppKit", "coregraphics": "CoreGraphics", "foundation": "Foundation"}

// AppleBindingsVariable names a directory of generated binding files to read instead of the
// cache, for a test's generated fixtures or a machine with no SDK.
const AppleBindingsVariable = "ADAMIC_APPLE_BINDINGS"

var generatedBindings struct {
	once      sync.Once
	directory string
	err       error
}

// appleBindingsDirectory is where the generated binding files are: AppleBindingsVariable's
// directory when it's set, or this Mac's cache, generated the first time it's needed.
func appleBindingsDirectory() (string, error) {
	if directory := os.Getenv(AppleBindingsVariable); directory != "" {
		return directory, nil
	}
	generatedBindings.once.Do(func() {
		generatedBindings.directory, generatedBindings.err = cachedAppleBindings()
	})
	return generatedBindings.directory, generatedBindings.err
}

var appleLeaves struct {
	once    sync.Once
	classes map[string]bool
	err     error
}

// AppleLeaves are the Objective-C classes the cycle finder doesn't walk through, read from the leaf
// table generated beside the bindings (internal/apple/generate, leaves.go): so the table always
// answers for the SDK the program is compiled against.
func AppleLeaves() (map[string]bool, error) {
	appleLeaves.once.Do(func() {
		directory, err := appleBindingsDirectory()
		if err != nil {
			appleLeaves.err = err
			return
		}
		text, err := os.ReadFile(filepath.Join(directory, generate.LeavesFile))
		if err != nil {
			appleLeaves.err = fmt.Errorf("the leaf table beside Apple's bindings: %w", err)
			return
		}
		appleLeaves.classes = map[string]bool{}
		for _, line := range strings.Split(string(text), "\n") {
			if class, isLeaf := strings.CutPrefix(line, "leaf "); isLeaf {
				appleLeaves.classes[strings.TrimSpace(class)] = true
			}
		}
	})
	return appleLeaves.classes, appleLeaves.err
}

// cachedAppleBindings generates the SDK's bindings once per SDK build and generator version.
// A directory is complete only once renamed into place, so a run cut short leaves nothing a later
// run would trust.
func cachedAppleBindings() (string, error) {
	if runtime.GOOS != "darwin" {
		return "", fmt.Errorf("Apple's frameworks are generated from Xcode's SDK, which only a Mac has; set %s to a directory of generated bindings", AppleBindingsVariable)
	}
	build, err := generate.SDKBuild("macosx")
	if err != nil {
		return "", err
	}
	caches, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	directory := filepath.Join(caches, "adamic", "apple", "macosx-"+build+"-"+generate.Version())
	if _, err := os.Stat(directory); err == nil {
		return directory, nil
	}
	if err := os.MkdirAll(filepath.Dir(directory), 0o755); err != nil {
		return "", err
	}
	partial, err := os.MkdirTemp(filepath.Dir(directory), filepath.Base(directory)+".partial-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(partial)
	frameworks := []string{}
	for _, name := range generatedFrameworks {
		frameworks = append(frameworks, name)
	}
	sort.Strings(frameworks)
	output, err := generate.FromSDK(context.Background(), "macosx", "macos", frameworks)
	if err != nil {
		return "", fmt.Errorf("generating Apple's bindings: %w", err)
	}
	if err := output.Write(partial); err != nil {
		return "", err
	}
	if err := generate.CheckSDK(context.Background(), "macosx", partial); err != nil {
		return "", err
	}
	if err := os.Rename(partial, directory); err != nil {
		// Another compile finished first; its directory is the same bytes.
		if _, statError := os.Stat(directory); statError == nil {
			return directory, nil
		}
		return "", err
	}
	return directory, nil
}

// appleBinding is the binding files for an apple/ module (apple/appkit/window), by the path each
// is served at: the generated one, and an embedded addition beside it, or the embedded module
// where its framework isn't generated.
func appleBinding(module string) (map[string]string, error) {
	files := map[string]string{}
	relative := strings.TrimPrefix(module, "apple/")
	framework, _, _ := strings.Cut(relative, "/")
	_, generated := generatedFrameworks[framework]
	if embedded, err := appleBindings.ReadFile(module + ".d.ts"); err == nil {
		if !generated {
			files[appleDirectory+"/"+relative+".d.ts"] = string(embedded)
			return files, nil
		}
		files[appleDirectory+"/"+relative+".addition.d.ts"] = string(embedded)
	}
	if !generated && os.Getenv(AppleBindingsVariable) == "" {
		return files, nil
	}
	directory, err := appleBindingsDirectory()
	if err != nil {
		return files, err
	}
	if binding, err := os.ReadFile(filepath.Join(directory, filepath.FromSlash(relative)+".d.ts")); err == nil {
		files[appleDirectory+"/"+relative+".d.ts"] = string(binding)
	}
	return files, nil
}

// appleImport is an import of an apple/ module, as written in a source or a binding file. A match
// inside a comment or a string only loads a binding file the program didn't need, which costs a
// parse and changes nothing the checker proves.
var appleImport = regexp.MustCompile(`(?:from|import)\s*['"](apple/[a-z0-9/-]+)['"]`)

// relativeImport is an import of one of the program's own modules.
var relativeImport = regexp.MustCompile(`(?:from|import)\s*['"](\.{1,2}/[^'"]+)['"]`)

// appleRoots is every binding file the program reaches, as a root for the checker: the apple/
// modules its sources import, followed through its own imports and the binding files' imports of
// each other. It adds each one to overlay, which is how the source file system serves it. A module
// with no binding file is left out, and the checker says it can't find it; bindings that can't be
// generated are an error.
func appleRoots(source *sourceFS, roots []string, overlay map[string]string) ([]string, error) {
	added := []string{}
	seen := map[string]bool{}
	var failure error
	var visit func(fileName string, text string)
	visit = func(fileName string, text string) {
		for _, match := range appleImport.FindAllStringSubmatch(text, -1) {
			// apple/appkit/window is served at /adamic-apple/appkit/window.d.ts.
			if seen[match[1]] {
				continue
			}
			seen[match[1]] = true
			files, err := appleBinding(match[1])
			if err != nil && failure == nil {
				failure = err
			}
			for _, bindingPath := range sortedPaths(files) {
				overlay[bindingPath] = files[bindingPath]
				added = append(added, bindingPath)
				visit(bindingPath, files[bindingPath])
			}
			// An apple/ module written in Adamic is found by the resolver, but the binding files it
			// imports (apple/swiftui/views imports apple/swiftui/native) are roots like any others.
			if implementation, err := appleBindings.ReadFile(match[1] + ".a"); err == nil {
				visit(appleImplementations+strings.TrimPrefix(match[1], "apple/")+".ts", string(implementation))
			}
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
	return added, failure
}

func sortedPaths(files map[string]string) []string {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
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
