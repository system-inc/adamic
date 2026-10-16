package runtimeproducts

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// modulePath is the repository's module, the prefix of every package of it.
const modulePath = "github.com/system-inc/adamic"

// leanSlice are the packages of the tree a lean build's recipe keys whole, buildcache's own and what it links
// (buildcache's leanRecipeOf): a build may reach them as it likes.
var leanSlice = []string{modulePath + "/internal/buildcache", modulePath + "/internal/buildcache/location", modulePath + "/internal/toolchain"}

// pureConstructors are the standard library calls an initializer may make and still only set its own variable: no
// effect on the process, the environment, the working directory or a file (Node's rail 1, (c)). The one list of them.
var pureConstructors = []string{"regexp.MustCompile", "regexp.MustCompilePOSIX", "errors.New", "strings.NewReplacer"}

// leanCensus is every way a lean build's code reaches past the files its key declares (Node's rails on #earlycut), from
// entry ("Type.Method" or "Function") in the package of the tree in directory, importPath, for linux/amd64, as the line
// builds: each function, method, variable, constant or type it reaches, or an init or initializer that counts as
// reached, defined in a file declared doesn't name (a path relative to root, or under a directory it names); each
// package of the tree it reaches outside leanSlice; and each thing the census can't follow, which fails closed: a call
// through a function value held in a package variable or a field, or its use as a value, an interface call or one
// through a type parameter, reflection, a go:linkname, a cgo or assembly file. Empty
// when the build is complete by its declared files and leanSlice.
func leanCensus(root, directory, importPath, entry string, declared []string) ([]string, error) {
	context := build.Default
	context.GOOS, context.GOARCH, context.CgoEnabled = "linux", "amd64", true
	listed, err := context.ImportDir(directory, 0)
	if err != nil {
		return nil, err
	}
	findings := map[string]bool{}
	for _, name := range append(slices.Clone(listed.CgoFiles), listed.SFiles...) {
		findings[fmt.Sprintf("%s is a cgo or assembly file, which the census can't read: review it", name)] = true
	}
	fileSet := token.NewFileSet()
	var files []*ast.File
	for _, name := range append(slices.Clone(listed.GoFiles), listed.CgoFiles...) {
		file, err := parser.ParseFile(fileSet, filepath.Join(directory, name), nil, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	// As the make census does, with every import an empty package: the census needs only which of this package's own
	// objects a use names, and what package a selector reaches, which resolve with no GOROOT, as on a runner.
	info := &types.Info{Uses: map[*ast.Ident]types.Object{}, Defs: map[*ast.Ident]types.Object{}}
	checked, _ := (&types.Config{Importer: emptyImports{}, Error: func(error) {}}).Check(importPath, fileSet, files, info)
	if checked == nil {
		return nil, fmt.Errorf("%s didn't type-check at all", directory)
	}
	bodies := map[types.Object]ast.Node{}
	var inits []*ast.FuncDecl
	initializers := map[types.Object]ast.Expr{}
	var entryObject types.Object
	for _, file := range files {
		for _, declaration := range file.Decls {
			switch declaration := declaration.(type) {
			case *ast.FuncDecl:
				if declaration.Recv == nil && declaration.Name.Name == "init" {
					inits = append(inits, declaration)
					continue
				}
				object := info.Defs[declaration.Name]
				bodies[object] = declaration
				if receiverName(declaration)+declaration.Name.Name == entry {
					entryObject = object
				}
			case *ast.GenDecl:
				if declaration.Tok != token.VAR {
					continue
				}
				for _, spec := range declaration.Specs {
					value := spec.(*ast.ValueSpec)
					for index, name := range value.Names {
						if index < len(value.Values) {
							initializers[info.Defs[name]] = value.Values[index]
						} else if len(value.Values) == 1 {
							initializers[info.Defs[name]] = value.Values[0]
						}
					}
				}
			}
		}
	}
	if entryObject == nil {
		return nil, fmt.Errorf("%s defines no %s", directory, entry)
	}
	reached := map[types.Object]bool{}
	var queue []ast.Node
	mark := func(object types.Object) {
		if object == nil || object.Pkg() != checked || reached[object] {
			return
		}
		switch object := object.(type) {
		case *types.Var:
			if object.IsField() || object.Parent() != checked.Scope() {
				return
			}
		case *types.Const, *types.TypeName:
			if object.Parent() != checked.Scope() {
				return
			}
		}
		reached[object] = true
		if body, ok := bodies[object]; ok {
			queue = append(queue, body)
		}
		if value, ok := initializers[object]; ok {
			queue = append(queue, value)
		}
		// A reached type's methods are all reached: a value of it handed out as an interface is called back through them.
		if named, ok := object.Type().(*types.Named); ok {
			if _, isType := object.(*types.TypeName); isType {
				for index := range named.NumMethods() {
					reachedMethod := named.Method(index)
					if !reached[reachedMethod] {
						reached[reachedMethod] = true
						if body, ok := bodies[reachedMethod]; ok {
							queue = append(queue, body)
						}
					}
				}
			}
		}
	}
	at := func(node ast.Node) string {
		position := fileSet.Position(node.Pos())
		return fmt.Sprintf("%s:%d", filepath.Base(position.Filename), position.Line)
	}
	walk := func(node ast.Node) {
		ast.Inspect(node, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.Ident:
				switch object := info.Uses[node].(type) {
				case *types.PkgName:
					path := object.Imported().Path()
					switch {
					case path == "reflect":
						findings[fmt.Sprintf("%s: reflection (reflect), which the census can't follow", at(node))] = true
					case strings.HasPrefix(path, modulePath+"/") && !slices.Contains(leanSlice, path):
						findings[fmt.Sprintf("%s: reaches %s, a package of the tree no lean recipe keys", at(node), path)] = true
					}
				case *types.Var:
					// A function value in a package variable or a field may be set by code outside the reach, so its use
					// fails closed, called or handed on. One in a parameter or a local came from the reached code that set it.
					if _, isFunction := object.Type().Underlying().(*types.Signature); isFunction && (object.IsField() || object.Parent() == checked.Scope()) {
						findings[fmt.Sprintf("%s: the function value %s, held in a package variable or a field, which the census can't follow", at(node), node.Name)] = true
					}
					mark(object)
				case nil:
					// The census's one limit (Tests' and Node's on 4de7029d15): with every import empty, a selector whose
					// receiver's type comes from an import doesn't resolve and is skipped here, never a finding. It's safe
					// today because a type of this package reaches the walk through a declaration that names it, which
					// takes all its methods. Closing it means type-checking with real imports, then failing closed on any
					// selector still unresolved, on a box with a go that can load them (#wrr2dwd).
				default:
					mark(object)
				}
			case *ast.CallExpr:
				function := ast.Unparen(node.Fun)
				if index, ok := function.(*ast.IndexExpr); ok {
					function = index.X
				}
				var name *ast.Ident
				switch function := function.(type) {
				case *ast.Ident:
					name = function
				case *ast.SelectorExpr:
					name = function.Sel
				}
				if name == nil {
					return true
				}
				switch object := info.Uses[name].(type) {
				case *types.Func:
					if signature, ok := object.Type().(*types.Signature); ok && signature.Recv() != nil {
						switch signature.Recv().Type().Underlying().(type) {
						case *types.Interface:
							findings[fmt.Sprintf("%s: an interface call of %s, which the census can't follow", at(node), name.Name)] = true
						}
						if _, isParameter := signature.Recv().Type().(*types.TypeParam); isParameter {
							findings[fmt.Sprintf("%s: a call of %s through a type parameter, which the census can't follow", at(node), name.Name)] = true
						}
					}
				}
			}
			return true
		})
	}
	mark(entryObject)
	for {
		for len(queue) > 0 {
			node := queue[0]
			queue = queue[1:]
			walk(node)
		}
		// Rail 1, (b) and (c): an init or an initializer the reach didn't take counts as reached when it writes a variable
		// the reached code reads or calls anything but a pure constructor, so what runs at load can't set what the build
		// reads unseen. One that only sets its own variable from literals reaches the build only through that variable,
		// which (a) takes with the variable.
		grown := false
		for _, declaration := range inits {
			if !reached[info.Defs[declaration.Name]] && effective(declaration.Body, info, checked, reached) {
				reached[info.Defs[declaration.Name]] = true
				findings[fmt.Sprintf("init in %s counts as reached", filepath.Base(fileSet.Position(declaration.Pos()).Filename))] = !declares(root, declared, fileSet.Position(declaration.Pos()).Filename)
				queue = append(queue, declaration.Body)
				grown = true
			}
		}
		for object, value := range initializers {
			if !reached[object] && effective(value, info, checked, reached) {
				mark(object)
				grown = true
			}
		}
		if !grown {
			break
		}
	}
	for object := range reached {
		if object == nil || !object.Pos().IsValid() {
			continue
		}
		file := fileSet.Position(object.Pos()).Filename
		if !declares(root, declared, file) {
			findings[fmt.Sprintf("%s, defined in %s, which the build's key doesn't declare", object.Name(), filepath.Base(file))] = true
		}
		if content, err := os.ReadFile(file); err == nil && strings.Contains(string(content), "//go:linkname") {
			findings[fmt.Sprintf("%s holds a go:linkname, which the census can't follow", filepath.Base(file))] = true
		}
	}
	var census []string
	for finding, holds := range findings {
		if holds {
			census = append(census, finding)
		}
	}
	sort.Strings(census)
	return census, nil
}

// effective says whether code, an init's body or an initializer, writes a variable reached holds or calls anything but a
// builtin, a conversion or a pure constructor (pureConstructors).
func effective(code ast.Node, info *types.Info, checked *types.Package, reached map[types.Object]bool) bool {
	found := false
	ast.Inspect(code, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.FuncLit:
			// A function literal is a value until it's called, and a call of it is a call through a value.
			return false
		case *ast.AssignStmt:
			for _, target := range node.Lhs {
				if name, ok := ast.Unparen(target).(*ast.Ident); ok && reached[info.Uses[name]] {
					found = true
				}
			}
		case *ast.CallExpr:
			function := ast.Unparen(node.Fun)
			var name *ast.Ident
			qualified := ""
			switch function := function.(type) {
			case *ast.Ident:
				name = function
			case *ast.SelectorExpr:
				name = function.Sel
				if pkg, ok := ast.Unparen(function.X).(*ast.Ident); ok {
					if imported, ok := info.Uses[pkg].(*types.PkgName); ok {
						qualified = imported.Imported().Path() + "." + function.Sel.Name
					}
				}
			}
			switch info.Uses[name].(type) {
			case *types.Builtin, *types.TypeName:
				return true
			case nil:
				if !slices.Contains(pureConstructors, qualified) {
					found = true
				}
			default:
				found = true
			}
		}
		return !found
	})
	return found
}

// declares says whether file, an absolute path, is one declared names, relative to root, or under a directory it names.
func declares(root string, declared []string, file string) bool {
	relative, err := filepath.Rel(root, file)
	if err != nil {
		return false
	}
	relative = filepath.ToSlash(relative)
	for _, name := range declared {
		if relative == name || strings.HasPrefix(relative, strings.TrimSuffix(name, "/")+"/") {
			return true
		}
	}
	return false
}

// receiverName is a method declaration's receiver type's name with a dot, or "" for a function.
func receiverName(declaration *ast.FuncDecl) string {
	if declaration.Recv == nil || len(declaration.Recv.List) == 0 {
		return ""
	}
	receiver := declaration.Recv.List[0].Type
	if star, ok := receiver.(*ast.StarExpr); ok {
		receiver = star.X
	}
	if index, ok := receiver.(*ast.IndexExpr); ok {
		receiver = index.X
	}
	if name, ok := receiver.(*ast.Ident); ok {
		return name.Name + "."
	}
	return ""
}

// emptyImports answers every import with an empty, complete package named for its path's last element, so the census
// type-checks with no GOROOT and no module cache, as buildcache's make census does.
type emptyImports struct{}

func (emptyImports) Import(path string) (*types.Package, error) {
	elements := strings.Split(path, "/")
	name := elements[len(elements)-1]
	if len(elements) > 1 && len(name) > 1 && name[0] == 'v' && strings.Trim(name[1:], "0123456789") == "" {
		name = elements[len(elements)-2]
	}
	imported := types.NewPackage(path, name)
	imported.MarkComplete()
	return imported, nil
}

// repositoryRoot is the tree this package sits in, three directories up.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// An archive's build reaches nothing of internal/native but the files its key declares (runtimeFiles), and no package
// of the tree but buildcache's slice, so its lean recipe keys all the code that makes it and no compiler file (Node's
// rails on #earlycut): what native.RuntimeRequest.Build reaches, every init and initializer that writes what it reads or
// calls anything, and nothing the census can't follow. Mutant: jobserver.go left out of runtimeFiles.
func TestTheRuntimeBuildReachesOnlyItsDeclaredFiles(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	census, err := leanCensus(root, filepath.Join(root, "internal", "native"), modulePath+"/internal/native", "RuntimeRequest.Build", runtimeFiles)
	if err != nil {
		t.Fatal(err)
	}
	if len(census) > 0 {
		t.Fatalf("the runtime archive's build reaches past what its key declares:\n%s", strings.Join(census, "\n"))
	}
}

// The census follows a build's reach and fails closed on what it can't (Node's rails): a function moved into an
// undeclared file, an init calling os.Setenv in one, an init writing a variable the build reads, a call through a
// function value, an interface call, reflection, a package of the tree outside buildcache's slice, and a type whose
// method is reached each name a finding, and so does a function value held in a package variable, called or handed
// on. A call through a parameter given a literal names none. An initializer that only sets its own variable from a literal or a pure
// constructor, in an undeclared file, names none. Mutants: rail (c) off (the os.Setenv init passes); rail (b) off; a pure
// constructor taken as an effect; package variables' function values followed silently.
func TestTheLeanCensusFollowsTheBuildAndFailsClosed(t *testing.T) {
	t.Parallel()
	base := map[string]string{
		"build.go":  "package fixture\n\ntype Request struct{}\n\nvar mode = \"fast\"\n\nfunc (r Request) Build() string { return helper() + mode }\n",
		"helper.go": "package fixture\n\nfunc helper() string { return \"h\" }\n",
		"tables.go": "package fixture\n\nimport \"regexp\"\n\nvar name = regexp.MustCompile(`x`)\nvar table = map[string]int{\"a\": 1}\n\nfunc unused() {}\n",
	}
	declared := []string{"fixture/build.go", "fixture/helper.go"}
	for name, test := range map[string]struct {
		files map[string]string
		want  string
	}{
		"complete by its declared files":    {nil, ""},
		"a function moved to an undeclared": {map[string]string{"helper.go": "", "moved.go": "package fixture\n\nfunc helper() string { return \"h\" }\n"}, "helper, defined in moved.go"},
		"an init calling os.Setenv":         {map[string]string{"env.go": "package fixture\n\nimport \"os\"\n\nfunc init() { os.Setenv(\"CC\", \"gcc\") }\n"}, "init in env.go counts as reached"},
		"an init writing what it reads":     {map[string]string{"set.go": "package fixture\n\nfunc init() { mode = \"slow\" }\n"}, "init in set.go counts as reached"},
		"a call through a function value":   {map[string]string{"helper.go": "package fixture\n\nvar hook func() string\n\nfunc helper() string { return hook() }\n"}, "the function value hook"},
		"a function value handed on":        {map[string]string{"helper.go": "package fixture\n\nvar hook func() string\n\nfunc helper() string { return call(hook) }\n\nfunc call(f func() string) string { return f() }\n"}, "the function value hook"},
		"a call through a parameter":        {map[string]string{"helper.go": "package fixture\n\nfunc helper() string { return call(func() string { return \"h\" }) }\n\nfunc call(f func() string) string { return f() }\n"}, ""},
		"an interface call":                 {map[string]string{"helper.go": "package fixture\n\ntype namer interface{ name() string }\n\nfunc helper() string { var n namer; return n.name() }\n"}, "an interface call of name"},
		"reflection":                        {map[string]string{"helper.go": "package fixture\n\nimport \"reflect\"\n\nfunc helper() string { return reflect.TypeOf(1).Name() }\n"}, "reflection (reflect)"},
		"a package of the tree":             {map[string]string{"helper.go": "package fixture\n\nimport \"github.com/system-inc/adamic/internal/ir\"\n\nfunc helper() string { return ir.Name }\n"}, "reaches github.com/system-inc/adamic/internal/ir"},
		"a reached type's other methods":    {map[string]string{"methods.go": "package fixture\n\nfunc (r Request) Other() string { return \"o\" }\n"}, "Other, defined in methods.go"},
	} {
		root := t.TempDir()
		directory := filepath.Join(root, "fixture")
		os.MkdirAll(directory, 0o755)
		files := map[string]string{}
		for file, content := range base {
			files[file] = content
		}
		for file, content := range test.files {
			files[file] = content
		}
		for file, content := range files {
			if content != "" {
				os.WriteFile(filepath.Join(directory, file), []byte(content), 0o644)
			}
		}
		census, err := leanCensus(root, directory, "example.com/fixture", "Request.Build", declared)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		switch {
		case test.want == "" && len(census) > 0:
			t.Errorf("%s: %q, want nothing", name, census)
		case test.want != "" && !slices.ContainsFunc(census, func(finding string) bool { return strings.Contains(finding, test.want) }):
			t.Errorf("%s: %q, want a finding naming %q", name, census, test.want)
		}
	}
}

// An archive's lean build names runtimeproducts as its caller in the trace's journal, so Settle keys it by this
// package's code and buildcache's (leanRecipeOf), never by the binary that settled it. Mutant: the caller read a frame
// too near, buildcache's own.
// Not parallel: it sets ADAMIC_BUILD_CACHE and ADAMIC_BUILD_TRACE, which are process-wide.
func TestAnArchivesBuildIsJournaledAsRuntimeproducts(t *testing.T) {
	trace := t.TempDir()
	t.Setenv("ADAMIC_BUILD_CACHE", "off")
	t.Setenv("ADAMIC_BUILD_TRACE", trace)
	if _, err := buildcache.LeanGet(buildcache.Inputs{Name: "runtimeproducts journal probe"}, func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "probe"), []byte("probe\n"), 0o644)
	}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(trace, "journal.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
		var entry struct {
			Event, Name, LeanCaller string
			Lean                    bool
		}
		if json.Unmarshal([]byte(line), &entry) == nil && entry.Event == "built" {
			if !entry.Lean || entry.LeanCaller != modulePath+"/internal/native/runtimeproducts" {
				t.Fatalf("the built entry is lean %v, its caller %q; want runtimeproducts' lean build", entry.Lean, entry.LeanCaller)
			}
			return
		}
	}
	t.Fatalf("no built entry in the journal: %s", content)
}
