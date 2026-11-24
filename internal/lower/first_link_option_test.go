package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// firstLinkOptionDoors are the directories, from the repository's root, whose Go files may name
// Options.FirstLinkTraps: internal/lower, which reads it, and cmd/adamic-firstlink, the one command that sets
// it. A trap-mode program is measurement and never a product, so a reader anywhere else is how one would leak
// into a build, an oracle run or a gate.
var firstLinkOptionDoors = []string{"internal/lower", "cmd/adamic-firstlink"}

// firstLinkOptionReaders is every Go file under root outside firstLinkOptionDoors that names FirstLinkTraps, as
// a composite literal's key or a selector, by its path from root. override, by path from root, is read in place
// of a file's text, so a mutant needs no tree of its own. Test data, the TypeScript checkout under cohere and
// node_modules hold no reader of ours, so they aren't walked. A file that doesn't parse is a reader if its text
// names FirstLinkTraps at all, so it fails closed.
func firstLinkOptionReaders(t *testing.T, root string, override map[string]string) []string {
	t.Helper()
	var readers []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.IsDir() {
			if name := entry.Name(); name == ".git" || name == "testdata" || name == "node_modules" || relative == "cohere" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(relative, ".go") || slices.Contains(firstLinkOptionDoors, filepath.ToSlash(filepath.Dir(relative))) {
			return nil
		}
		text, overridden := override[relative]
		if !overridden {
			contents, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			text = string(contents)
		}
		if !strings.Contains(text, "FirstLinkTraps") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), relative, text, parser.SkipObjectResolution)
		if err != nil {
			readers = append(readers, relative+" (doesn't parse, and names FirstLinkTraps)")
			return nil
		}
		named := false
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.KeyValueExpr:
				if key, isName := node.Key.(*ast.Ident); isName && key.Name == "FirstLinkTraps" {
					named = true
				}
			case *ast.SelectorExpr:
				if node.Sel.Name == "FirstLinkTraps" {
					named = true
				}
			}
			return !named
		})
		if named {
			readers = append(readers, relative)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return readers
}

// Options.FirstLinkTraps has one door: no Go file in the repository outside internal/lower and
// cmd/adamic-firstlink names it, so no environment variable, flag or gate of adamic's can lower a program in
// trap mode. Mutant: cmd/adamic's loweringOptions sets FirstLinkTraps: true, and the test names cmd/adamic/main.go.
func TestFirstLinkTrapsOptionHasOneDoor(t *testing.T) {
	t.Parallel()
	// Reads: every Go file in the repository
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if readers := firstLinkOptionReaders(t, root, nil); len(readers) != 0 {
		t.Fatalf("FirstLinkTraps is named outside %v: %v", firstLinkOptionDoors, readers)
	}
	main, err := os.ReadFile(filepath.Join(root, "cmd/adamic/main.go"))
	if err != nil {
		t.Fatal(err)
	}
	const options = "Debug: debug, ModuleFormat: format}"
	if strings.Count(string(main), options) != 1 {
		t.Fatalf("cmd/adamic/main.go's loweringOptions no longer ends %q; the mutant can't be planted", options)
	}
	mutant := strings.Replace(string(main), options, "Debug: debug, ModuleFormat: format, FirstLinkTraps: true}", 1)
	if readers := firstLinkOptionReaders(t, root, map[string]string{"cmd/adamic/main.go": mutant}); !slices.Equal(readers, []string{"cmd/adamic/main.go"}) {
		t.Fatalf("with loweringOptions setting FirstLinkTraps: got %v, want cmd/adamic/main.go named", readers)
	}
}

// The door's directory may be there or not: on a tree without cmd/adamic-firstlink the guard still finds no
// reader, and one there that sets the option is let through, while the same line in cmd/adamic is named, as a
// selector or a key.
func TestFirstLinkTrapsOptionDoorMayBeAbsent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write := func(path, text string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, path)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, path), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("internal/lower/lower.go", "package lower\n\ntype Options struct{ FirstLinkTraps bool }\n\nfunc read(options Options) bool { return options.FirstLinkTraps }\n")
	if readers := firstLinkOptionReaders(t, root, nil); len(readers) != 0 {
		t.Fatalf("without cmd/adamic-firstlink: got %v, want no reader", readers)
	}
	write("cmd/adamic-firstlink/main.go", "package main\n\nvar options = lower.Options{FirstLinkTraps: true}\n")
	if readers := firstLinkOptionReaders(t, root, nil); len(readers) != 0 {
		t.Fatalf("with cmd/adamic-firstlink setting it: got %v, want no reader", readers)
	}
	write("cmd/adamic/main.go", "package main\n\nfunc set(options *lower.Options) { options.FirstLinkTraps = true }\n")
	write("cmd/adamic/options.go", "package main\n\nvar options = lower.Options{FirstLinkTraps: true}\n")
	if readers := firstLinkOptionReaders(t, root, nil); !slices.Equal(readers, []string{"cmd/adamic/main.go", "cmd/adamic/options.go"}) {
		t.Fatalf("cmd/adamic naming it: got %v, want both of its files", readers)
	}
}
