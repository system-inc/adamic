package runtimeproducts

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
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
// effect on the process, the environment, the working directory or a file (Node's rail 1, (c)). The one list of them,
// each as go/types names a function.
var pureConstructors = []string{"regexp.MustCompile", "regexp.MustCompilePOSIX", "errors.New", "strings.NewReplacer"}

// censusExports is the export data of internal/native and every package it and its tests import, a product built ahead
// (TestProduct_LeanCensusExports), so the census type-checks with real types and runs no go (Tests' and Node's must 1
// on 4de7029d15, as internal/ir's call-target guard does).
func censusExports(t testing.TB) string {
	t.Helper()
	return buildcache.GoExports(t, "./internal/native")
}

func TestProduct_LeanCensusExports(t *testing.T) {
	t.Parallel()
	censusExports(t)
}

// censusImports is what product, a GoExports product's directory, holds: each package's export data by its import path
// (never a variant compiled for a test), the packages of the tree among them, and every package it lists, variants by
// their own path.
func censusImports(product string) (exports map[string]string, tree []string, listed map[string]bool, err error) {
	encoded, err := os.ReadFile(filepath.Join(product, "listing.json"))
	if err != nil {
		return nil, nil, nil, err
	}
	var listing []buildcache.ListedPackage
	if err := json.Unmarshal(encoded, &listing); err != nil {
		return nil, nil, nil, err
	}
	exports, listed = map[string]string{}, map[string]bool{}
	for _, pkg := range listing {
		path, _, _ := strings.Cut(pkg.ImportPath, " [")
		listed[path] = true
		if pkg.Export == "" || path != pkg.ImportPath {
			continue
		}
		exports[path] = filepath.Join(product, pkg.Export)
		if strings.HasPrefix(path, modulePath+"/") && !strings.HasSuffix(path, ".test") {
			tree = append(tree, path)
		}
	}
	return exports, tree, listed, nil
}

// leanCensus is every way a lean build's code reaches past the files its key declares (Node's rails and musts on
// #earlycut), from entry ("Type.Method" or "Function") in the package of the tree in directory, importPath, type-checked
// with its test files against exports' real types (censusImports), for linux/amd64, as the line builds:
//   - each function, method, variable, constant or type it reaches, or an init or initializer that counts as reached,
//     defined in a file declared doesn't name (a path relative to root, or under a directory it names). A type is
//     reached with every method of it whenever a reached declaration names it: a variable's type, a field's, a
//     parameter's or a result's (M4). An interface call reaches every type of the package's own files that implements
//     it (M3);
//   - each write to a reached package variable, through an index, a selector, a star or its address handed on, from a
//     file declared doesn't name, whatever reaches it, or from code of a declared one the reach doesn't take; any use
//     in an undeclared file of one a copy shares storage with; and each reached package variable another package could
//     write, an exported one (M1, M2);
//   - each package of the tree it reaches outside leanSlice, and each interface call some type of such a package
//     implements;
//   - and each thing the census can't follow, which fails closed: a type error, a call it couldn't type, a call through a
//     function value held in a package variable or a field or its use as a value, a call through a type parameter,
//     reflection, a go:linkname, a cgo or assembly file.
//
// The package's test files are read with its own (rail 1, M2) when its tests build lean, testsBuildLean: a write or an
// init in one runs in the package's test binary alone, so it can reach a lean build only when that binary links a lean
// caller.
//
// Empty when the build is complete by its declared files and leanSlice.
func leanCensus(root, directory, importPath, entry string, declared []string, exports map[string]string, tree []string, testsBuildLean bool) ([]string, error) {
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
	production := map[*ast.File]bool{}
	for index, name := range slices.Concat(listed.GoFiles, listed.CgoFiles, listed.TestGoFiles) {
		file, err := parser.ParseFile(fileSet, filepath.Join(directory, name), nil, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
		production[file] = index < len(listed.GoFiles)+len(listed.CgoFiles)
	}
	// read says whether what file holds runs in a process that can build lean: the package's own files always, its test
	// files when its test binary links a lean caller.
	read := func(file *ast.File) bool { return production[file] || testsBuildLean }
	// Real types, from export data built ahead: a type error, which no build of the package has, is the census reading
	// what the build doesn't, so it fails closed.
	imports := importer.ForCompiler(fileSet, "gc", func(path string) (io.ReadCloser, error) {
		if exported, ok := exports[path]; ok {
			return os.Open(exported)
		}
		return nil, fmt.Errorf("the census's export data holds no %s", path)
	})
	info := &types.Info{
		Uses: map[*ast.Ident]types.Object{}, Defs: map[*ast.Ident]types.Object{},
		Types: map[ast.Expr]types.TypeAndValue{}, Selections: map[*ast.SelectorExpr]*types.Selection{},
	}
	checkError := func(err error) {
		var typeError types.Error
		if errors.As(err, &typeError) {
			position := typeError.Fset.Position(typeError.Pos)
			err = fmt.Errorf("%s:%d: %s", filepath.Base(position.Filename), position.Line, typeError.Msg)
		}
		findings[fmt.Sprintf("a type error, which fails the census closed: %v", err)] = true
	}
	checked, _ := (&types.Config{Importer: imports, Error: checkError}).Check(importPath, fileSet, files, info)
	if checked == nil {
		return nil, fmt.Errorf("%s didn't type-check at all", directory)
	}
	at := func(node ast.Node) string {
		position := fileSet.Position(node.Pos())
		return fmt.Sprintf("%s:%d", filepath.Base(position.Filename), position.Line)
	}
	fileOf := func(node ast.Node) string { return fileSet.Position(node.Pos()).Filename }
	bodies := map[types.Object]ast.Node{}
	var inits []*ast.FuncDecl
	initializers := map[types.Object]ast.Expr{}
	var entryObject types.Object
	for _, file := range files {
		for _, declaration := range file.Decls {
			switch declaration := declaration.(type) {
			case *ast.FuncDecl:
				if declaration.Recv == nil && declaration.Name.Name == "init" {
					if read(file) {
						inits = append(inits, declaration)
					}
					continue
				}
				object := info.Defs[declaration.Name]
				bodies[object] = declaration
				if production[file] && receiverName(declaration)+declaration.Name.Name == entry {
					entryObject = object
				}
			case *ast.GenDecl:
				for _, spec := range declaration.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						// A reached type's declaration is read, so the types its fields, elements and methods' signatures
						// name are reached with their methods, and the constants its array lengths name (M4).
						bodies[info.Defs[spec.Name]] = spec
					case *ast.ValueSpec:
						for index, name := range spec.Names {
							switch {
							case !read(file):
							case index < len(spec.Values):
								initializers[info.Defs[name]] = spec.Values[index]
							case len(spec.Values) == 1:
								initializers[info.Defs[name]] = spec.Values[0]
							}
							// A reached variable's declared type is reached with its methods (M4), as a function's
							// signature is with its declaration.
							if spec.Type != nil {
								bodies[info.Defs[name]] = spec.Type
							}
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
	var mark func(object types.Object)
	mark = func(object types.Object) {
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
		// A reached type's methods are all reached: a value of it handed out as an interface is called back through them,
		// as fmt calls a field's String when it prints the value holding it (M4).
		if _, isType := object.(*types.TypeName); isType {
			if named, ok := object.Type().(*types.Named); ok {
				for method := range named.Methods() {
					mark(method)
				}
			}
		}
	}
	// implementers reaches every type of the package's own files that implements iface (M3): an interface call in reached
	// code dispatches to any value of it, wherever that value was made. A type of the tree outside leanSlice that
	// implements it is a finding, since no lean recipe keys its methods.
	implementers := func(iface *types.Interface, method string, node ast.Node) {
		implements := func(typ types.Type) bool {
			return types.Implements(typ, iface) || types.Implements(types.NewPointer(typ), iface)
		}
		for _, name := range checked.Scope().Names() {
			object, isType := checked.Scope().Lookup(name).(*types.TypeName)
			if !isType || object.IsAlias() || types.IsInterface(object.Type()) || !production[fileAt(files, fileSet, object.Pos())] {
				continue
			}
			if named, ok := object.Type().(*types.Named); ok && named.TypeParams().Len() > 0 {
				// A generic type implements iface only once instantiated, so one with a method of the name counts.
				if found, _, _ := types.LookupFieldOrMethod(types.NewPointer(object.Type()), true, checked, method); found != nil {
					mark(object)
				}
				continue
			}
			if implements(object.Type()) {
				mark(object)
			}
		}
		for _, path := range tree {
			if path == importPath || slices.Contains(leanSlice, path) {
				continue
			}
			other, err := imports.Import(path)
			if err != nil {
				findings[fmt.Sprintf("%s: an interface call of %s, and the census can't read %s's types: %v", at(node), method, path, err)] = true
				continue
			}
			for _, name := range other.Scope().Names() {
				if object, isType := other.Scope().Lookup(name).(*types.TypeName); isType && !object.IsAlias() && !types.IsInterface(object.Type()) && implements(object.Type()) {
					findings[fmt.Sprintf("%s: an interface call of %s, which %s.%s implements, a package of the tree no lean recipe keys", at(node), method, path, name)] = true
				}
			}
		}
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
					// A function value in a package variable or a field of this package may be set by code outside the
					// reach, so its use fails closed, called or handed on. One in a parameter or a local came from the
					// reached code that set it.
					if _, isFunction := object.Type().Underlying().(*types.Signature); isFunction && object.Pkg() == checked && (object.IsField() || object.Parent() == checked.Scope()) {
						findings[fmt.Sprintf("%s: the function value %s, held in a package variable or a field, which the census can't follow", at(node), node.Name)] = true
					}
					mark(object)
				default:
					mark(object)
				}
			case *ast.SelectorExpr:
				selection := info.Selections[node]
				if selection == nil || selection.Kind() == types.FieldVal {
					return true
				}
				method, _ := selection.Obj().(*types.Func)
				if method == nil {
					return true
				}
				receiver := method.Signature().Recv()
				if receiver == nil {
					return true
				}
				if _, isParameter := types.Unalias(selection.Recv()).(*types.TypeParam); isParameter {
					findings[fmt.Sprintf("%s: a call of %s through a type parameter, which the census can't follow", at(node), method.Name())] = true
					return true
				}
				if iface, ok := receiver.Type().Underlying().(*types.Interface); ok {
					implementers(iface, method.Name(), node)
				}
			case *ast.CallExpr:
				if _, typed := info.Types[node.Fun]; !typed {
					findings[fmt.Sprintf("%s: a call the census couldn't type, which it can't follow", at(node))] = true
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
			if !reached[info.Defs[declaration.Name]] && effective(declaration.Body, info, reached) {
				reached[info.Defs[declaration.Name]] = true
				findings[fmt.Sprintf("init in %s counts as reached", filepath.Base(fileOf(declaration)))] = !declares(root, declared, fileOf(declaration))
				queue = append(queue, declaration.Body)
				grown = true
			}
		}
		for object, value := range initializers {
			if !reached[object] && effective(value, info, reached) {
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
		// M2: a reached package variable another package could write is beyond the census.
		if variable, ok := object.(*types.Var); ok && variable.Exported() {
			findings[fmt.Sprintf("%s, which the build reads, is exported, so a package the census doesn't read can write it", object.Name())] = true
		}
	}
	// M1 and M2: a write to a reached package variable, anywhere in the package's files and its test files, whatever
	// reaches it, from a file the build's key doesn't declare. So is any use there of one whose storage a copy shares (a
	// slice, a map, a pointer), since a local holding it writes what the build reads where no write names the variable.
	// In a declared file, whose content is keyed, a write is the build's own inside a reached function, an init or an
	// initializer the reach took, which run as it does; anywhere else the value it stores isn't keyed, an unreached
	// setter's that a test calls before the window (Tests' must on 929af5812d).
	for _, file := range files {
		if !read(file) {
			continue
		}
		if declares(root, declared, fileOf(file)) {
			for _, declaration := range file.Decls {
				if !takenByTheReach(declaration, info, reached) {
					eachWrite(declaration, info, true, func(root types.Object, node ast.Node) {
						if variable, ok := root.(*types.Var); ok && reached[variable] {
							findings[fmt.Sprintf("%s: writes %s, which the build reads, in code the reach doesn't take", at(node), variable.Name())] = true
						}
					})
				}
			}
			continue
		}
		eachWrite(file, info, true, func(root types.Object, node ast.Node) {
			if variable, ok := root.(*types.Var); ok && reached[variable] {
				findings[fmt.Sprintf("%s: writes %s, which the build reads, in a file its key doesn't declare", at(node), variable.Name())] = true
			}
		})
		ast.Inspect(file, func(node ast.Node) bool {
			if name, ok := node.(*ast.Ident); ok {
				if variable, ok := info.Uses[name].(*types.Var); ok && reached[variable] && sharesStorage(variable.Type(), map[types.Type]bool{}) {
					findings[fmt.Sprintf("%s: uses %s, which the build reads and a copy of which shares its storage, in a file its key doesn't declare", at(node), variable.Name())] = true
				}
			}
			return true
		})
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

// takenByTheReach says whether declaration runs as the reach does: a reached function or method, an init, or package
// variables the reach took, every one of them.
func takenByTheReach(declaration ast.Decl, info *types.Info, reached map[types.Object]bool) bool {
	switch declaration := declaration.(type) {
	case *ast.FuncDecl:
		return declaration.Recv == nil && declaration.Name.Name == "init" || reached[info.Defs[declaration.Name]]
	case *ast.GenDecl:
		if declaration.Tok != token.VAR {
			return true
		}
		for _, spec := range declaration.Specs {
			for _, name := range spec.(*ast.ValueSpec).Names {
				if !reached[info.Defs[name]] {
					return false
				}
			}
		}
		return true
	}
	return false
}

// sharesStorage says whether a copy of a value of typ shares storage with it, so writing the copy writes the original:
// a pointer, slice, map, channel, function or interface, or a struct or array holding one.
func sharesStorage(typ types.Type, seen map[types.Type]bool) bool {
	if seen[typ] {
		return false
	}
	seen[typ] = true
	switch typ := typ.Underlying().(type) {
	case *types.Pointer, *types.Slice, *types.Map, *types.Chan, *types.Signature, *types.Interface:
		return true
	case *types.Array:
		return sharesStorage(typ.Elem(), seen)
	case *types.Struct:
		for field := range typ.Fields() {
			if sharesStorage(field.Type(), seen) {
				return true
			}
		}
	}
	return false
}

// fileAt is the file of files holding position.
func fileAt(files []*ast.File, fileSet *token.FileSet, position token.Pos) *ast.File {
	for _, file := range files {
		if fileSet.File(file.Pos()) == fileSet.File(position) {
			return file
		}
	}
	return nil
}

// effective says whether code, an init's body or an initializer, writes a variable reached holds (M1: through an index,
// a selector, a star or its address) or calls anything but a builtin, a conversion or a pure constructor
// (pureConstructors).
func effective(code ast.Node, info *types.Info, reached map[types.Object]bool) bool {
	found := false
	eachWrite(code, info, false, func(root types.Object, _ ast.Node) {
		if reached[root] {
			found = true
		}
	})
	ast.Inspect(code, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.FuncLit:
			// A function literal is a value until it's called, and a call of it is a call through a value.
			return false
		case *ast.CallExpr:
			if typed, ok := info.Types[node.Fun]; ok && (typed.IsType() || typed.IsBuiltin()) {
				return true
			}
			if function := calledFunction(node, info); function == nil || !slices.Contains(pureConstructors, function.FullName()) {
				found = true
			}
		}
		return !found
	})
	return found
}

// calledFunction is the function or method call names, or nil for a call through a value.
func calledFunction(call *ast.CallExpr, info *types.Info) *types.Func {
	switch function := ast.Unparen(call.Fun).(type) {
	case *ast.Ident:
		called, _ := info.Uses[function].(*types.Func)
		return called
	case *ast.SelectorExpr:
		called, _ := info.Uses[function.Sel].(*types.Func)
		return called
	case *ast.IndexExpr:
		return calledFunction(&ast.CallExpr{Fun: function.X}, info)
	case *ast.IndexListExpr:
		return calledFunction(&ast.CallExpr{Fun: function.X}, info)
	}
	return nil
}

// eachWrite calls visit with the root object of each place code writes, and the node writing it (M1): an assignment's
// or an increment's target, a range's assigned key and value, a value whose address is taken, the first argument of
// copy, clear or delete, and the receiver a pointer method takes the address of. A root is the variable a target is
// reached through, by indexes, slices, field selectors and stars: runtimeFeatures for runtimeFeatures[0] = x, r for
// r.L.x = y. Function literals are entered when literals is set.
func eachWrite(code ast.Node, info *types.Info, literals bool, visit func(root types.Object, node ast.Node)) {
	target := func(expression ast.Expr, node ast.Node) {
		if root := writeRoot(expression, info); root != nil {
			visit(root, node)
		}
	}
	ast.Inspect(code, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.FuncLit:
			return literals
		case *ast.AssignStmt:
			if node.Tok != token.DEFINE {
				for _, left := range node.Lhs {
					target(left, node)
				}
			}
		case *ast.IncDecStmt:
			target(node.X, node)
		case *ast.RangeStmt:
			if node.Tok == token.ASSIGN {
				for _, assigned := range []ast.Expr{node.Key, node.Value} {
					if assigned != nil {
						target(assigned, node)
					}
				}
			}
		case *ast.UnaryExpr:
			if node.Op == token.AND {
				target(node.X, node)
			}
		case *ast.CallExpr:
			if builtin, ok := info.Uses[identOf(node.Fun)].(*types.Builtin); ok && len(node.Args) > 0 {
				switch builtin.Name() {
				case "copy", "clear", "delete":
					target(node.Args[0], node)
				}
			}
		case *ast.SelectorExpr:
			// A pointer method called on an addressable value takes its address.
			if selection := info.Selections[node]; selection != nil && selection.Kind() == types.MethodVal {
				if method, ok := selection.Obj().(*types.Func); ok {
					if _, pointer := method.Signature().Recv().Type().(*types.Pointer); pointer {
						if _, already := selection.Recv().Underlying().(*types.Pointer); !already {
							target(node.X, node)
						}
					}
				}
			}
		}
		return true
	})
}

// identOf is the identifier expression names, through parentheses, or nil.
func identOf(expression ast.Expr) *ast.Ident {
	name, _ := ast.Unparen(expression).(*ast.Ident)
	return name
}

// writeRoot is the variable a write to expression writes through, or nil when it isn't one (a call's result).
func writeRoot(expression ast.Expr, info *types.Info) types.Object {
	for {
		switch place := ast.Unparen(expression).(type) {
		case *ast.Ident:
			if object := info.Uses[place]; object != nil {
				return object
			}
			return info.Defs[place]
		case *ast.IndexExpr:
			expression = place.X
		case *ast.IndexListExpr:
			expression = place.X
		case *ast.SliceExpr:
			expression = place.X
		case *ast.StarExpr:
			expression = place.X
		case *ast.SelectorExpr:
			if selection := info.Selections[place]; selection != nil {
				if selection.Kind() != types.FieldVal {
					return nil
				}
				expression = place.X
				continue
			}
			// A qualified name: another package's variable.
			return info.Uses[place.Sel]
		default:
			return nil
		}
	}
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

// censusTypes is the export data the census type-checks against (censusExports, censusImports), and whether
// internal/native's test binary links this package, the one lean caller of native's code.
func censusTypes(t *testing.T) (exports map[string]string, tree []string, testsBuildLean bool) {
	t.Helper()
	exports, tree, listed, err := censusImports(censusExports(t))
	if err != nil {
		t.Fatal(err)
	}
	if !listed[modulePath+"/internal/native"] {
		t.Fatal("the census's export data doesn't list internal/native")
	}
	return exports, tree, listed[modulePath+"/internal/native/runtimeproducts"]
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
	exports, tree, testsBuildLean := censusTypes(t)
	census, err := leanCensus(root, filepath.Join(root, "internal", "native"), modulePath+"/internal/native", "RuntimeRequest.Build", runtimeFiles, exports, tree, testsBuildLean)
	if err != nil {
		t.Fatal(err)
	}
	if len(census) > 0 {
		t.Fatalf("the runtime archive's build reaches past what its key declares:\n%s", strings.Join(census, "\n"))
	}
}

// The census follows a build's reach with real types and fails closed on what it can't (Node's rails, Mentat's M1 to M4
// on 4de7029d15, Tests' must on 929af5812d). Each of these names a finding: a function moved into an undeclared file;
// an init there calling os.Setenv, writing a variable the build reads, through an index too, or taking its address; a
// write to one from code nothing reaches, through a selector, a pointer method or clear, or through a local copy
// sharing its storage; an unreached setter writing one in a declared file; an exported variable the build reads; a test
// file's write when the package's tests build lean; a call through a function value held in a package variable, called
// or handed on; an interface call, reaching a type of an undeclared file that implements it, or one a package of the
// tree implements; reflection; a package of the tree outside buildcache's slice; a reached type's other methods, and
// the methods of a type reached only as a field's or a variable's; and a type error. A call through a parameter given a
// literal names none, nor does a reached function's write in a declared file, a test file's write when the tests don't
// build lean, or an initializer that only sets its own variable from a literal or a pure constructor in an undeclared
// file. Mutants: rail (c) off (the os.Setenv init passes); rail (b) off; a pure constructor taken as an effect; package
// variables' function values followed silently; a write's root read through bare names alone; an address taken, a
// pointer method's receiver or a builtin's argument not a write; writes from code nothing reaches unread; declared
// files' writes unread, and read in reached functions too; uses of a variable sharing its storage unread, and a slice
// taken as not sharing it; the exported rule off; test files read whether or not the tests build lean, and never; an
// interface's implementers not reached; the tree's implementers unread; a reached type's declaration unread; a reached
// variable's declared type unread; type errors dropped.
func TestTheLeanCensusFollowsTheBuildAndFailsClosed(t *testing.T) {
	t.Parallel()
	base := map[string]string{
		"build.go": "package fixture\n\nimport \"fmt\"\n\ntype Request struct{ L level }\n\ntype level int\n\ntype config struct{ depth int }\n\n" +
			"func (c *config) reset() { c.depth = 0 }\n\nvar mode = \"fast\"\n\nvar features = []string{\"a\"}\n\nvar settings config\n\n" +
			"func (r Request) Build() string { return helper() + mode + features[0] + fmt.Sprint(r, settings.depth) }\n",
		"helper.go": "package fixture\n\nfunc helper() string { return \"h\" }\n",
		"tables.go": "package fixture\n\nimport \"regexp\"\n\nvar name = regexp.MustCompile(`x`)\nvar table = map[string]int{\"a\": 1}\n\nfunc unused() {}\n",
	}
	declared := []string{"fixture/build.go", "fixture/helper.go"}
	exports, tree, _ := censusTypes(t)
	for name, test := range map[string]struct {
		files map[string]string
		lean  bool
		want  string
	}{
		"complete by its declared files":         {nil, true, ""},
		"a function moved to an undeclared":      {map[string]string{"helper.go": "", "moved.go": "package fixture\n\nfunc helper() string { return \"h\" }\n"}, true, "helper, defined in moved.go"},
		"an init calling os.Setenv":              {map[string]string{"env.go": "package fixture\n\nimport \"os\"\n\nfunc init() { os.Setenv(\"CC\", \"gcc\") }\n"}, true, "init in env.go counts as reached"},
		"an init writing what it reads":          {map[string]string{"set.go": "package fixture\n\nfunc init() { mode = \"slow\" }\n"}, true, "init in set.go counts as reached"},
		"an init writing through an index":       {map[string]string{"set.go": "package fixture\n\nfunc init() { features[0] = \"b\" }\n"}, true, "init in set.go counts as reached"},
		"an init taking an address":              {map[string]string{"set.go": "package fixture\n\nfunc init() { p := &mode; _ = p }\n"}, true, "init in set.go counts as reached"},
		"a write from code nothing reaches":      {map[string]string{"other.go": "package fixture\n\nfunc SetMode(m string) { mode = m }\n"}, true, "writes mode"},
		"a write through a selector":             {map[string]string{"other.go": "package fixture\n\nfunc tune() { settings.depth = 2 }\n"}, true, "writes settings"},
		"a pointer method's call":                {map[string]string{"other.go": "package fixture\n\nfunc tidy() { settings.reset() }\n"}, true, "writes settings"},
		"a builtin clearing it":                  {map[string]string{"other.go": "package fixture\n\nfunc wipe() { clear(features) }\n"}, true, "writes features"},
		"a copy written in a local":              {map[string]string{"other.go": "package fixture\n\nfunc alias() { local := features; local[0] = \"b\" }\n"}, true, "uses features"},
		"an unreached setter in a declared file": {map[string]string{"helper.go": "package fixture\n\nfunc helper() string { return \"h\" }\n\nfunc SetMode(m string) { mode = m }\n"}, true, "writes mode, which the build reads, in code the reach doesn't take"},
		"a reached write in a declared file":     {map[string]string{"helper.go": "package fixture\n\nfunc helper() string { mode = \"m\"; return \"h\" }\n"}, true, ""},
		"an exported variable the build reads":   {map[string]string{"helper.go": "package fixture\n\nvar Prefix = \"p\"\n\nfunc helper() string { return Prefix }\n"}, true, "Prefix, which the build reads, is exported"},
		"a test file's write, tests built lean":  {map[string]string{"set_test.go": "package fixture\n\nimport \"testing\"\n\nfunc TestSet(t *testing.T) { mode = \"slow\" }\n"}, true, "writes mode"},
		"a test file's write, tests never lean":  {map[string]string{"set_test.go": "package fixture\n\nimport \"testing\"\n\nfunc TestSet(t *testing.T) { mode = \"slow\" }\n"}, false, ""},
		"a call through a function value":        {map[string]string{"helper.go": "package fixture\n\nvar hook func() string\n\nfunc helper() string { return hook() }\n"}, true, "the function value hook"},
		"a function value handed on":             {map[string]string{"helper.go": "package fixture\n\nvar hook func() string\n\nfunc helper() string { return call(hook) }\n\nfunc call(f func() string) string { return f() }\n"}, true, "the function value hook"},
		"a call through a parameter":             {map[string]string{"helper.go": "package fixture\n\nfunc helper() string { return call(func() string { return \"h\" }) }\n\nfunc call(f func() string) string { return f() }\n"}, true, ""},
		"an interface call":                      {map[string]string{"helper.go": "package fixture\n\ntype namer interface{ name() string }\n\nvar current namer\n\nfunc helper() string { return current.name() }\n", "loud.go": "package fixture\n\ntype loud struct{}\n\nfunc (loud) name() string { return \"l\" }\n"}, true, "loud, defined in loud.go"},
		"an interface call the tree implements":  {map[string]string{"helper.go": "package fixture\n\nvar failure error\n\nfunc helper() string { return failure.Error() }\n"}, true, "a package of the tree no lean recipe keys"},
		"reflection":                             {map[string]string{"helper.go": "package fixture\n\nimport \"reflect\"\n\nfunc helper() string { return reflect.TypeOf(1).Name() }\n"}, true, "reflection (reflect)"},
		"a package of the tree":                  {map[string]string{"helper.go": "package fixture\n\nimport \"github.com/system-inc/adamic/internal/ir\"\n\nfunc helper() string { var f ir.Function; _ = f; return \"h\" }\n"}, true, "reaches github.com/system-inc/adamic/internal/ir"},
		"a reached type's other methods":         {map[string]string{"methods.go": "package fixture\n\nfunc (r Request) Other() string { return \"o\" }\n"}, true, "Other, defined in methods.go"},
		"a type reached as a field's":            {map[string]string{"methods.go": "package fixture\n\nfunc (l level) String() string { return \"l\" }\n"}, true, "String, defined in methods.go"},
		"a type reached as a variable's":         {map[string]string{"helper.go": "package fixture\n\ntype counter int\n\nvar current counter\n\nfunc helper() string { _ = current; return \"h\" }\n", "methods.go": "package fixture\n\nfunc (c counter) String() string { return \"c\" }\n"}, true, "String, defined in methods.go"},
		"a type error":                           {map[string]string{"helper.go": "package fixture\n\nfunc helper() string { return missing() }\n"}, true, "a type error, which fails the census closed"},
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
		census, err := leanCensus(root, directory, "example.com/fixture", "Request.Build", declared, exports, tree, test.lean)
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
