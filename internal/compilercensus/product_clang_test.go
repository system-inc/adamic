package compilercensus

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// productClangRoots are the trees whose Go files build products.
var productClangRoots = []string{"internal", "cmd", "stage1", "stage2", "stage3"}

// bareProductClang is each function, in a file that builds a product, that runs clang by name itself rather than through
// native's clangIn: how many such calls it holds, and why it may. clangIn maps the compiler's installation and any sysroot
// out of what a compile records of a path (toolMaps, #8wgk65j); a bare clang maps only the directories its caller names,
// so the same key builds different bytes on two boxes whose tools sit in different places. Every entry here but the timer
// mutant's is one compiler/products-through-clangin moves onto clangIn, and takes off this list in the same commit.
var bareProductClang = map[string]struct {
	calls  int
	reason string
}{
	"internal/native/to_primitive_test.go:primitiveMutants":                       {2, "moves to clangIn in compiler/products-through-clangin"},
	"internal/native/record_test.go:recordMutants":                                {2, "moves to clangIn in compiler/products-through-clangin"},
	"stage1/cohere/lint/profile_compilation_main_test.go:compilationProductKinds": {1, "moves to clangIn in compiler/products-through-clangin"},
	"stage1/cohere/estree/recovery_cache_test.go:recoveryStableNative":            {1, "moves to clangIn in compiler/products-through-clangin"},
	"internal/native/timers_test.go:timerRuntime.buildBare":                       {1, "the reproducibility pin's mutant: it must stay bare so TestTimerMutantProductPinCatchesBareClang can show the pin fails"},
}

// productCommandCalls are the calls that run a program by name, each with the index of its program argument.
var productCommandCalls = map[string]int{
	"exec.Command":               0,
	"exec.CommandContext":        1,
	"toolchain.Command":          0,
	"toolchain.CommandIn":        1,
	"toolchain.CommandContextIn": 2,
}

// In a Go file that builds a product (it names buildcache.Product or buildcache.Inputs), no function runs clang or clang++ by name itself
// unless bareProductClang gives it a reason: a product's clang runs through native's clangIn, which maps the tools root
// out of the bytes as every native product's compile does. TestEveryCCompilerRunsThroughNative reads every name of a
// compiler but skips internal/native, and lists product builders that map their own directories; this reads the one shape
// that escapes the tools-root maps, in internal/native too, so a new bare call reds at its own change. A listed function
// whose count no longer matches reds too, so the list says what is. Mutant: a bare toolchain.CommandIn(..., "clang", ...)
// added to a product file.
func TestProductClangRunsThroughClangIn(t *testing.T) {
	t.Parallel()
	// Reads: *.go
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	found := map[string][]string{}
	for _, tree := range productClangRoots {
		start := filepath.Join(root, tree)
		err := filepath.WalkDir(start, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				if path == start && errors.Is(err, fs.ErrNotExist) {
					return filepath.SkipDir
				}
				return err
			}
			if entry.IsDir() {
				if entry.Name() == "testdata" || entry.Name() == "node_modules" || strings.HasPrefix(entry.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			relative, _ := filepath.Rel(root, path)
			relative = filepath.ToSlash(relative)
			files := token.NewFileSet()
			parsed, err := parser.ParseFile(files, path, nil, parser.SkipObjectResolution)
			if err != nil {
				return err
			}
			if !buildsProduct(parsed) {
				return nil
			}
			for _, declaration := range parsed.Decls {
				function, ok := declaration.(*ast.FuncDecl)
				if !ok || function.Body == nil {
					continue
				}
				site := relative + ":" + functionName(function)
				ast.Inspect(function.Body, func(node ast.Node) bool {
					if call, ok := node.(*ast.CallExpr); ok && bareClang(call) {
						found[site] = append(found[site], strings.TrimPrefix(files.Position(call.Pos()).String(), root+"/"))
					}
					return true
				})
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for site, positions := range found {
		listed, ok := bareProductClang[site]
		if !ok {
			for _, position := range positions {
				t.Errorf("%s runs clang by name in a file that builds a product: run it through native's clangIn (or a seam on it, outside package native), so the tools root is mapped out of the bytes, or list %s in bareProductClang with why it may", position, site)
			}
			continue
		}
		if len(positions) != listed.calls {
			t.Errorf("%s runs clang by name %d times, bareProductClang lists %d: %s", site, len(positions), listed.calls, strings.Join(positions, ", "))
		}
	}
	for site := range bareProductClang {
		if _, ok := found[site]; !ok {
			t.Errorf("bareProductClang lists %s, which runs no bare clang in a product file now: take it off", site)
		}
	}
}

// buildsProduct reports whether a file names buildcache.Product or buildcache.Inputs: a file that writes a product's
// inputs and hands them to a wrapper elsewhere in its package (estree's recoveryProduct) builds a product as surely as
// one that calls Product itself.
func buildsProduct(file *ast.File) bool {
	builds := false
	ast.Inspect(file, func(node ast.Node) bool {
		if name := selectorName(node); name == "buildcache.Product" || name == "buildcache.Inputs" {
			builds = true
		}
		return !builds
	})
	return builds
}

// bareClang reports whether call runs clang or clang++ by name through exec or toolchain.
func bareClang(call *ast.CallExpr) bool {
	index, ok := productCommandCalls[selectorName(call.Fun)]
	if !ok || len(call.Args) <= index {
		return false
	}
	literal, ok := call.Args[index].(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return false
	}
	value, err := strconv.Unquote(literal.Value)
	return err == nil && (value == "clang" || value == "clang++")
}

// selectorName is package.Name for a selector on a package's name, or "".
func selectorName(expression ast.Node) string {
	selector, ok := expression.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return ""
	}
	return pkg.Name + "." + selector.Sel.Name
}

// functionName is a function's name, or Receiver.name for a method.
func functionName(function *ast.FuncDecl) string {
	if function.Recv == nil || len(function.Recv.List) == 0 {
		return function.Name.Name
	}
	receiver := function.Recv.List[0].Type
	if star, ok := receiver.(*ast.StarExpr); ok {
		receiver = star.X
	}
	if index, ok := receiver.(*ast.IndexExpr); ok {
		receiver = index.X
	}
	if ident, ok := receiver.(*ast.Ident); ok {
		return ident.Name + "." + function.Name.Name
	}
	return function.Name.Name
}
