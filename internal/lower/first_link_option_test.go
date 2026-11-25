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

// firstLinkOptionDoors are the directories, from the repository's root, whose Go files may name any field called
// FirstLinkTraps, lower's Options' and the two backends' of the same name: internal/lower, which reads its own,
// and cmd/adamic-firstlink, the one command that sets them. A trap-mode program is measurement and never a
// product, so a setter anywhere else is how one would leak into a build, an oracle run or a gate.
var firstLinkOptionDoors = []string{"internal/lower", "cmd/adamic-firstlink"}

// firstLinkOptionBackends are the directories whose Go files may read a FirstLinkTraps but never set one: the
// backends, each refusing a trap-mode program unless its own field is set (native_runtime.Options and
// javascript.Options), and native_runtime, which declares native's.
var firstLinkOptionBackends = []string{"internal/native", "internal/javascript", "internal/native_runtime"}

// firstLinkOptionLeaks is every Go file under root outside firstLinkOptionDoors that sets FirstLinkTraps, by its
// path from root, or, outside firstLinkOptionBackends too, names it at all. A setter is a composite literal's key,
// an assignment to the selector, or the selector's address taken, which a write could go through. override, by
// path from root, is read in place of a file's text, so a mutant needs no tree of its own. Test data, the
// TypeScript checkout under cohere and node_modules hold no Go of ours, so they aren't walked. A file that doesn't
// parse is a leak if its text names FirstLinkTraps at all, so it fails closed.
func firstLinkOptionLeaks(t *testing.T, root string, override map[string]string) []string {
	t.Helper()
	var leaks []string
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
		directory := filepath.ToSlash(filepath.Dir(relative))
		if !strings.HasSuffix(relative, ".go") || slices.Contains(firstLinkOptionDoors, directory) {
			return nil
		}
		reads := slices.Contains(firstLinkOptionBackends, directory)
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
			leaks = append(leaks, relative+" (doesn't parse, and names FirstLinkTraps)")
			return nil
		}
		named := func(expression ast.Expr) bool {
			selector, isSelector := ast.Unparen(expression).(*ast.SelectorExpr)
			return isSelector && selector.Sel.Name == "FirstLinkTraps"
		}
		leaked := false
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.KeyValueExpr:
				if key, isName := node.Key.(*ast.Ident); isName && key.Name == "FirstLinkTraps" {
					leaked = true
				}
			case *ast.AssignStmt:
				leaked = leaked || slices.ContainsFunc(node.Lhs, named)
			case *ast.UnaryExpr:
				leaked = leaked || node.Op == token.AND && named(node.X)
			case *ast.SelectorExpr:
				leaked = leaked || !reads && named(node)
			}
			return !leaked
		})
		if leaked {
			leaks = append(leaks, relative)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return leaks
}

// Options.FirstLinkTraps has one door: no Go file in the repository outside internal/lower and
// cmd/adamic-firstlink sets it or either backend's field of the same name, and none outside them and the backends
// reads one, so no environment variable, flag or gate of adamic's can lower or build a program in trap mode.
// Mutant: cmd/adamic's loweringOptions sets FirstLinkTraps: true, and the test names cmd/adamic/main.go.
func TestFirstLinkTrapsOptionHasOneDoor(t *testing.T) {
	t.Parallel()
	// Reads: every Go file in the repository
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if leaks := firstLinkOptionLeaks(t, root, nil); len(leaks) != 0 {
		t.Fatalf("FirstLinkTraps is set outside %v, or read outside them and %v: %v", firstLinkOptionDoors, firstLinkOptionBackends, leaks)
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
	if leaks := firstLinkOptionLeaks(t, root, map[string]string{"cmd/adamic/main.go": mutant}); !slices.Equal(leaks, []string{"cmd/adamic/main.go"}) {
		t.Fatalf("with loweringOptions setting FirstLinkTraps: got %v, want cmd/adamic/main.go named", leaks)
	}
}

// The door's directory may be there or not: on a tree without cmd/adamic-firstlink the guard still finds no
// leak, and one there that sets the option is let through, while the same line in cmd/adamic is named, as an
// assignment or a key. A backend may read its own field but never set it, by a key, an assignment or the field's
// address, and anywhere else a plain read is named too.
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
	write("internal/native/emit.go", "package native\n\nfunc refuses(options EmitOptions) bool { return !options.Options.FirstLinkTraps }\n")
	write("internal/javascript/javascript.go", "package javascript\n\nfunc refuses(options Options) bool { return !(options).FirstLinkTraps }\n")
	if leaks := firstLinkOptionLeaks(t, root, nil); len(leaks) != 0 {
		t.Fatalf("without cmd/adamic-firstlink, the backends reading theirs: got %v, want no leak", leaks)
	}
	write("cmd/adamic-firstlink/main.go", "package main\n\nvar options = lower.Options{FirstLinkTraps: true}\n")
	if leaks := firstLinkOptionLeaks(t, root, nil); len(leaks) != 0 {
		t.Fatalf("with cmd/adamic-firstlink setting it: got %v, want no leak", leaks)
	}
	write("cmd/adamic/main.go", "package main\n\nfunc set(options *lower.Options) { options.FirstLinkTraps = true }\n")
	write("cmd/adamic/options.go", "package main\n\nvar options = lower.Options{FirstLinkTraps: true}\n")
	write("cmd/adamic/read.go", "package main\n\nfunc read(options lower.Options) bool { return options.FirstLinkTraps }\n")
	write("internal/javascript/set.go", "package javascript\n\nfunc set(options *Options) { (options.FirstLinkTraps), _ = true, 0 }\n")
	write("internal/native/key.go", "package native\n\nvar options = EmitOptions{Options: native_runtime.Options{FirstLinkTraps: true}}\n")
	write("internal/native_runtime/address.go", "package native_runtime\n\nfunc set(options *Options) { flag := &options.FirstLinkTraps; *flag = true }\n")
	want := []string{"cmd/adamic/main.go", "cmd/adamic/options.go", "cmd/adamic/read.go", "internal/javascript/set.go", "internal/native/key.go", "internal/native_runtime/address.go"}
	if leaks := firstLinkOptionLeaks(t, root, nil); !slices.Equal(leaks, want) {
		t.Fatalf("setters, and a read outside the backends: got %v, want %v", leaks, want)
	}
}
