// Command projects writes the repository root's TypeScript projects, the ones cohere checks every Adamic program in:
// tsconfig.json, a solution root that references the other two, tsconfig.standalone.json and tsconfig.node.json. Run it
// from the repository's root, after adding, moving or removing a program under the projects' include directories, or
// after changing what one imports:
//
//	go run ./internal/load/projects
//
// A program checks under Node's declarations exactly when Load gives its program Node's declarations, and Load decides
// that with one predicate, load.UsesNodeModules, over the program it builds first (internal/load/node_library.go). So
// tsconfig.node.json's files are each source under the include directories whose own first program, load.Unchecked,
// the predicate answers yes for, asked file by file, never a pattern that could drift from it, and
// tsconfig.standalone.json leaves those same files out. TestTheRootProjectsAreLoadsOwn holds all three files to what this
// writes. Each project carries the options Load sets (compilerOptions in internal/load/load.go) and the prelude Load
// roots: prelude.d.ts and its console, the standalone one or Node's (internal/load/source_fs.go). The Node project adds
// the pinned @types/node by its type root, the same bytes Load serves from the binary.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/system-inc/adamic/internal/load"
)

// includes are the directories the standalone project includes, relative to the repository's root.
var includes = []string{"dedication", "internal/load/testdata", "internal/oracle/testdata", "stage1", "bench"}

// The prelude files, as the projects name them from the repository's root.
const (
	preludeFile        = "internal/load/prelude.d.ts"
	preludeConsoleFile = "internal/load/prelude_console.d.ts"
	preludeNodeFile    = "internal/load/prelude_node.d.ts"
)

// The three files this writes, relative to the repository's root.
const (
	solutionFile   = "tsconfig.json"
	standaloneFile = "tsconfig.standalone.json"
	nodeFile       = "tsconfig.node.json"
)

// compilerOptions is load.go's compilerOptions as a tsconfig spells it, field for field and in its order.
type compilerOptions struct {
	Strict                     bool     `json:"strict"`
	NoUncheckedIndexedAccess   bool     `json:"noUncheckedIndexedAccess"`
	ExactOptionalPropertyTypes bool     `json:"exactOptionalPropertyTypes"`
	ErasableSyntaxOnly         bool     `json:"erasableSyntaxOnly"`
	VerbatimModuleSyntax       bool     `json:"verbatimModuleSyntax"`
	AllowImportingTsExtensions bool     `json:"allowImportingTsExtensions"`
	NoEmit                     bool     `json:"noEmit"`
	Module                     string   `json:"module"`
	ModuleDetection            string   `json:"moduleDetection"`
	ModuleResolution           string   `json:"moduleResolution"`
	Target                     string   `json:"target"`
	Lib                        []string `json:"lib"`
	TypeRoots                  []string `json:"typeRoots,omitempty"`
	Types                      []string `json:"types"`
}

// project is one tsconfig: its options, `.a` read as TypeScript (cohere's TypeScript reads sourceExtensions, stock
// tools ignore it), and what it checks.
type project struct {
	CompilerOptions  compilerOptions `json:"compilerOptions"`
	SourceExtensions []string        `json:"sourceExtensions"`
	Files            []string        `json:"files"`
	Include          []string        `json:"include,omitempty"`
	Exclude          []string        `json:"exclude,omitempty"`
}

// solution is a tsconfig that checks nothing itself and references the projects that do, which cohere runs in its
// place (cohere's command/cohere/ownership.go, resolveOwnership's solution roots).
type solution struct {
	Files      []string    `json:"files"`
	References []reference `json:"references"`
}

type reference struct {
	Path string `json:"path"`
}

// options is the options every Adamic program is checked under, standalone.
func options() compilerOptions {
	return compilerOptions{
		Strict:                     true,
		NoUncheckedIndexedAccess:   true,
		ExactOptionalPropertyTypes: true,
		ErasableSyntaxOnly:         false,
		VerbatimModuleSyntax:       true,
		AllowImportingTsExtensions: true,
		NoEmit:                     true,
		Module:                     "esnext",
		ModuleDetection:            "force",
		ModuleResolution:           "bundler",
		Target:                     "es2024",
		Lib:                        []string{"es2024"},
		Types:                      []string{},
	}
}

// sources is every Adamic source under the include directories, the files Load takes as a program's entry (a .a file
// or a .ts file that isn't a declaration file), node_modules left out as tsconfig's default exclude leaves it, by path
// from the repository's root with forward slashes, sorted.
func sources(repository string) ([]string, error) {
	found := []string{}
	for _, include := range includes {
		err := filepath.WalkDir(filepath.Join(repository, include), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".a") && (!strings.HasSuffix(name, ".ts") || strings.HasSuffix(name, ".d.ts")) {
				return nil
			}
			relative, err := filepath.Rel(repository, path)
			if err != nil {
				return err
			}
			found = append(found, filepath.ToSlash(relative))
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	slices.Sort(found)
	return found, nil
}

// placement is where the include directories' sources check: under Node's declarations, the files whose own first
// program Load gives them, and the files Load refuses before it builds a program at all (an import of an Apple module
// it has no binding for, say), which have no Node view to match and so stay standalone, as every file was before the
// split. Each is sorted, by path from the repository's root.
type placement struct {
	node    []string
	refused []string
}

// place asks load.UsesNodeModules of load.Unchecked for each source alone, its imports and the prelude, as Load asks
// it of the program it builds first.
func place(repository string) (placement, error) {
	all, err := sources(repository)
	if err != nil {
		return placement{}, err
	}
	absolute, err := filepath.Abs(repository)
	if err != nil {
		return placement{}, err
	}
	node := make([]bool, len(all))
	failures := make([]error, len(all))
	next := make(chan int)
	var workers sync.WaitGroup
	for range runtime.GOMAXPROCS(0) {
		workers.Go(func() {
			for index := range next {
				program, err := load.Unchecked([]string{filepath.Join(absolute, filepath.FromSlash(all[index]))})
				if err != nil {
					failures[index] = fmt.Errorf("%s: %w", all[index], err)
					continue
				}
				node[index] = load.UsesNodeModules(program)
			}
		})
	}
	for index := range all {
		next <- index
	}
	close(next)
	workers.Wait()
	placed := placement{node: []string{}, refused: []string{}}
	for index, path := range all {
		if failures[index] != nil {
			placed.refused = append(placed.refused, failures[index].Error())
		} else if node[index] {
			placed.node = append(placed.node, path)
		}
	}
	return placed, nil
}

// render is the three files' bytes, by their names, for these Node sources.
func render(nodeSources []string) (map[string][]byte, error) {
	standalone := project{
		CompilerOptions:  options(),
		SourceExtensions: []string{".a"},
		Files:            []string{preludeConsoleFile, preludeFile},
		Include:          includes,
		Exclude:          append([]string{"**/node_modules"}, nodeSources...),
	}
	nodeOptions := options()
	nodeOptions.TypeRoots = []string{"./internal/load/node_types/node_modules/@types"}
	nodeOptions.Types = []string{"node"}
	node := project{
		CompilerOptions:  nodeOptions,
		SourceExtensions: []string{".a"},
		Files:            append([]string{preludeNodeFile, preludeFile}, nodeSources...),
	}
	root := solution{
		Files:      []string{},
		References: []reference{{Path: "./" + standaloneFile}, {Path: "./" + nodeFile}},
	}
	rendered := map[string][]byte{}
	for name, value := range map[string]any{solutionFile: root, standaloneFile: standalone, nodeFile: node} {
		var buffer bytes.Buffer
		encoder := json.NewEncoder(&buffer)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "\t")
		if err := encoder.Encode(value); err != nil {
			return nil, err
		}
		rendered[name] = buffer.Bytes()
	}
	return rendered, nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "projects:", err)
		os.Exit(1)
	}
}

func run() error {
	placed, err := place(".")
	if err != nil {
		return err
	}
	rendered, err := render(placed.node)
	if err != nil {
		return err
	}
	for _, name := range []string{solutionFile, standaloneFile, nodeFile} {
		if err := os.WriteFile(name, rendered[name], 0o644); err != nil {
			return err
		}
	}
	for _, refusal := range placed.refused {
		fmt.Println("projects: standalone, since Load builds it no program:", refusal)
	}
	fmt.Printf("projects: %d sources check under Node's declarations, in %s\n", len(placed.node), nodeFile)
	return nil
}
