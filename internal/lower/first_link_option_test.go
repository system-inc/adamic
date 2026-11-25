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

// firstLinkOptionDoors are the directories, from the repository's root, whose Go files may set any field called
// FirstLinkTraps, lower's Options' and the two backends' of the same name: cmd/adamic-firstlink, the one command
// that sets them, and internal/lower, whose own tests set lower's and whose other files may only read it. A
// trap-mode program is measurement and never a product, so a setter anywhere else is how one would leak into a
// build, an oracle run or a gate.
var firstLinkOptionDoors = []string{"internal/lower", "cmd/adamic-firstlink"}

// firstLinkOptionBackends are the directories whose Go files may read a FirstLinkTraps but never set one: the
// backends, each refusing a trap-mode program unless its own field is set (native_runtime.Options and
// javascript.Options), and native_runtime, which declares native's.
var firstLinkOptionBackends = []string{"internal/native", "internal/javascript", "internal/native_runtime"}

// firstLinkLowerImport is internal/lower's import path, as a file outside it names lower's Options through.
const firstLinkLowerImport = `"github.com/system-inc/adamic/internal/lower"`

// firstLinkOptionLeaks is every Go file under root that could set a FirstLinkTraps but the door's, by its path from
// root. Outside cmd/adamic-firstlink and internal/lower's tests, that's a setter: a composite literal's key, an
// assignment to the field, the field's address taken, which a write could go through, or a literal of lower's
// Options with its fields unkeyed, which sets every one. Outside internal/lower and the backends, it's a read as
// well, and, in a file that isn't a test, a string literal holding the name, which reflection could set the field
// by. override, by path from root, is read in place of a file's text, so a mutant needs no tree of its own. Test
// data, the TypeScript checkout under cohere and node_modules hold no Go of ours, so they aren't walked. A file
// that doesn't parse is a leak if its text names FirstLinkTraps at all, so it fails closed.
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
		test := strings.HasSuffix(relative, "_test.go")
		lowerItself := directory == "internal/lower"
		if !strings.HasSuffix(relative, ".go") || directory == "cmd/adamic-firstlink" || lowerItself && test {
			return nil
		}
		reads := lowerItself || slices.Contains(firstLinkOptionBackends, directory)
		text, overridden := override[relative]
		if !overridden {
			contents, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			text = string(contents)
		}
		if !strings.Contains(text, "FirstLinkTraps") && !strings.Contains(text, firstLinkLowerImport) && !lowerItself {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), relative, text, parser.SkipObjectResolution)
		if err != nil {
			if strings.Contains(text, "FirstLinkTraps") {
				leaks = append(leaks, relative+" (doesn't parse, and names FirstLinkTraps)")
			}
			return nil
		}
		// lowerName is how this file names lower's package: "." inside it or by a dot import, its import's name
		// otherwise, and "" in a file that doesn't import it.
		lowerName := ""
		if lowerItself {
			lowerName = "."
		}
		for _, imported := range file.Imports {
			if imported.Path.Value != firstLinkLowerImport {
				continue
			}
			lowerName = "lower"
			if imported.Name != nil {
				lowerName = imported.Name.Name
			}
		}
		options := func(expression ast.Expr) bool {
			switch expression := ast.Unparen(expression).(type) {
			case *ast.Ident:
				return lowerName == "." && expression.Name == "Options"
			case *ast.SelectorExpr:
				name, isName := expression.X.(*ast.Ident)
				return isName && lowerName != "" && name.Name == lowerName && expression.Sel.Name == "Options"
			}
			return false
		}
		unkeyed := func(literal *ast.CompositeLit) bool {
			if len(literal.Elts) == 0 {
				return false
			}
			_, keyed := literal.Elts[0].(*ast.KeyValueExpr)
			return !keyed
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
			case *ast.BasicLit:
				leaked = leaked || !reads && !test && node.Kind == token.STRING && strings.Contains(node.Value, "FirstLinkTraps")
			case *ast.CompositeLit:
				leaked = leaked || options(node.Type) && unkeyed(node)
				// A slice's, an array's or a map's elements may leave their type out: []lower.Options{{true}}.
				var element ast.Expr
				switch collection := node.Type.(type) {
				case *ast.ArrayType:
					element = collection.Elt
				case *ast.MapType:
					element = collection.Value
				}
				if star, isPointer := element.(*ast.StarExpr); isPointer {
					element = star.X
				}
				if element == nil || !options(element) {
					break
				}
				for _, value := range node.Elts {
					if pair, isPair := value.(*ast.KeyValueExpr); isPair {
						value = pair.Value
					}
					if inner, isLiteral := value.(*ast.CompositeLit); isLiteral && inner.Type == nil && unkeyed(inner) {
						leaked = true
					}
				}
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

// Options.FirstLinkTraps has one door: no Go file in the repository but cmd/adamic-firstlink and internal/lower's
// tests sets it or either backend's field of the same name, none outside internal/lower and the backends reads
// one, and none of theirs that isn't a test holds the name in a string, so no environment variable, flag or gate
// of adamic's can lower or build a program in trap mode. Mutants, each named: cmd/adamic's loweringOptions sets
// FirstLinkTraps: true; cmd/adamic finds the field by its name through reflection; cmd/adamic writes lower's
// Options unkeyed; LowerWithOptions sets it from the environment.
func TestFirstLinkTrapsOptionHasOneDoor(t *testing.T) {
	t.Parallel()
	// Reads: every Go file in the repository
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if leaks := firstLinkOptionLeaks(t, root, nil); len(leaks) != 0 {
		t.Fatalf("FirstLinkTraps can be set outside %v's door: %v", firstLinkOptionDoors, leaks)
	}
	for _, mutant := range []struct{ name, path, before, after string }{
		{"loweringOptions setting it", "cmd/adamic/main.go", "Debug: debug, ModuleFormat: format}", "Debug: debug, ModuleFormat: format, FirstLinkTraps: true}"},
		{"the field found by its name", "cmd/adamic/main.go", "\nfunc compileLibrary(", "\nvar firstLinkField, _ = reflect.TypeFor[lower.Options]().FieldByName(\"FirstLinkTraps\")\n\nfunc compileLibrary("},
		{"lower's Options unkeyed", "cmd/adamic/main.go", "\nfunc compileLibrary(", "\nvar unkeyed = lower.Options{true}\n\nfunc compileLibrary("},
		{"LowerWithOptions setting it from the environment", "internal/lower/lower.go", "options Options) (*ir.Program, error) {\n", "options Options) (*ir.Program, error) {\n\toptions.FirstLinkTraps = os.Getenv(\"ADAMIC_FIRST_LINK_TRAPS\") == \"1\"\n"},
	} {
		text, err := os.ReadFile(filepath.Join(root, mutant.path))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(text), mutant.before) != 1 {
			t.Fatalf("%s: %s no longer holds %q once; the mutant can't be planted", mutant.name, mutant.path, mutant.before)
		}
		planted := strings.Replace(string(text), mutant.before, mutant.after, 1)
		if leaks := firstLinkOptionLeaks(t, root, map[string]string{mutant.path: planted}); !slices.Equal(leaks, []string{mutant.path}) {
			t.Errorf("with %s: got %v, want %s named", mutant.name, leaks, mutant.path)
		}
	}
}

// The door's directory may be there or not: on a tree without cmd/adamic-firstlink the guard still finds no
// leak, and one there that sets the option is let through, while the same line in cmd/adamic is named, as an
// assignment or a key. internal/lower's tests may set lower's, and its other files may only read it. A backend
// may read its own field but never set it, by a key, an assignment or the field's address, and anywhere else a
// plain read is named too, as is the name in a string outside a test, and lower's Options unkeyed, through an
// import's own name or with its type left out.
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
	write("internal/lower/lower_test.go", "package lower\n\nvar trapped = Options{FirstLinkTraps: true}\n\nvar unkeyed = Options{true}\n")
	write("internal/native/emit.go", "package native\n\nfunc refuses(options EmitOptions) bool { return !options.Options.FirstLinkTraps }\n")
	write("internal/javascript/javascript.go", "package javascript\n\nfunc refuses(options Options) bool { return !(options).FirstLinkTraps }\n")
	write("cmd/adamic/keyed.go", "package main\n\nimport \"github.com/system-inc/adamic/internal/lower\"\n\nvar options = lower.Options{Debug: true}\n")
	write("cmd/adamic/reflect_test.go", "package main\n\nconst field = \"FirstLinkTraps\"\n")
	if leaks := firstLinkOptionLeaks(t, root, nil); len(leaks) != 0 {
		t.Fatalf("without cmd/adamic-firstlink, lower and the backends reading theirs: got %v, want no leak", leaks)
	}
	write("cmd/adamic-firstlink/main.go", "package main\n\nvar options = lower.Options{FirstLinkTraps: true}\n")
	if leaks := firstLinkOptionLeaks(t, root, nil); len(leaks) != 0 {
		t.Fatalf("with cmd/adamic-firstlink setting it: got %v, want no leak", leaks)
	}
	write("cmd/adamic/main.go", "package main\n\nfunc set(options *lower.Options) { options.FirstLinkTraps = true }\n")
	write("cmd/adamic/options.go", "package main\n\nvar options = lower.Options{FirstLinkTraps: true}\n")
	write("cmd/adamic/read.go", "package main\n\nfunc read(options lower.Options) bool { return options.FirstLinkTraps }\n")
	write("cmd/adamic/reflect.go", "package main\n\nvar field, _ = reflect.TypeFor[lower.Options]().FieldByName(\"FirstLinkTraps\")\n")
	write("cmd/adamic/unkeyed.go", "package main\n\nimport lowering \"github.com/system-inc/adamic/internal/lower\"\n\nvar options = &lowering.Options{true}\n")
	write("internal/javascript/set.go", "package javascript\n\nfunc set(options *Options) { (options.FirstLinkTraps), _ = true, 0 }\n")
	write("internal/lower/door.go", "package lower\n\nfunc set(options *Options) { options.FirstLinkTraps = true }\n")
	write("internal/lower/positional.go", "package lower\n\nvar options = Options{true}\n")
	write("internal/native/key.go", "package native\n\nvar options = EmitOptions{Options: native_runtime.Options{FirstLinkTraps: true}}\n")
	write("internal/native_runtime/address.go", "package native_runtime\n\nfunc set(options *Options) { flag := &options.FirstLinkTraps; *flag = true }\n")
	write("internal/oracle/elided.go", "package oracle\n\nimport \"github.com/system-inc/adamic/internal/lower\"\n\nvar options = map[string]*lower.Options{\"trapped\": {true}}\n")
	want := []string{"cmd/adamic/main.go", "cmd/adamic/options.go", "cmd/adamic/read.go", "cmd/adamic/reflect.go", "cmd/adamic/unkeyed.go", "internal/javascript/set.go", "internal/lower/door.go", "internal/lower/positional.go", "internal/native/key.go", "internal/native_runtime/address.go", "internal/oracle/elided.go"}
	if leaks := firstLinkOptionLeaks(t, root, nil); !slices.Equal(leaks, want) {
		t.Fatalf("setters, and a read, a string or an unkeyed Options outside the backends: got %v, want %v", leaks, want)
	}
}
