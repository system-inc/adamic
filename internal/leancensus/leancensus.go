// Package leancensus holds a lean build's code to the files its key declares (#earlycut lever 3), for the tests of a
// package whose builds are lean. From the build's entries, everything it reaches must be defined in a declared file or
// in a package its recipe keys whole, every package variable it reads must be written only by its own initializer,
// anywhere in its package and the tests that share a process with a lean build, and anything the census can't follow
// fails closed. Three rounds of listing the shapes a write could take each opened the next, so the census holds that
// one claim instead (Mentat's reframe, Node's and Tests' ruling on adcc8fd771). It reads real types from export data a
// GoExports product built ahead, so it runs no go.
package leancensus

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
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
)

// ModulePath is the repository's module, the prefix of every package of it.
const ModulePath = "github.com/system-inc/adamic"

// PureConstructors are the standard library calls an initializer of a package variable the build reads may make and
// still only set its own variable: no effect on the process, the environment, the working directory or a file (Node's
// rail 1, (c)). The one list of them, each as go/types names a function.
var PureConstructors = []string{"regexp.MustCompile", "regexp.MustCompilePOSIX", "errors.New", "strings.NewReplacer"}

// A Census is one lean build's code to hold: Entries ("Type.Method" or "Function") in the package of the tree in
// Directory, ImportPath, under Root; Declared, the files and directories (relative to Root) its key reads as files;
// Slice, the packages of the tree its recipe keys whole, which the build may reach as it likes; and Exports and Tree,
// the export data it type-checks against, by import path, and the packages of the tree among them (Imports). The
// package's test files are read with its own when TestsBuildLean: a write or an init in one runs in the package's test
// binary alone, so it can reach a lean build only when that binary builds one.
type Census struct {
	Root, Directory, ImportPath string
	Entries                     []string
	Declared, Slice             []string
	Exports                     map[string]string
	Tree                        []string
	TestsBuildLean              bool
}

// A listedPackage is the part of a GoExports listing a census reads (buildcache.ListedPackage).
type listedPackage struct {
	ImportPath, Export string
}

// Imports is what product, a GoExports product's directory, holds: each package's export data by its import path
// (never a variant compiled for a test), the packages of the tree among them, and every package it lists, variants by
// their own path.
func Imports(product string) (exports map[string]string, tree []string, listed map[string]bool, err error) {
	encoded, err := os.ReadFile(filepath.Join(product, "listing.json"))
	if err != nil {
		return nil, nil, nil, err
	}
	var listing []listedPackage
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
		if strings.HasPrefix(path, ModulePath+"/") && !strings.HasSuffix(path, ".test") {
			tree = append(tree, path)
		}
	}
	return exports, tree, listed, nil
}

// Findings is every way the build's code reaches past what its key declares, sorted, empty when it's complete:
//   - each function, method, variable, constant or type it reaches, or an init or initializer that counts as reached,
//     defined in a file Declared doesn't name. A type is reached with every method of it whenever a reached
//     declaration names it (M4), and an interface call reaches every type of the package's own files that implements it
//     (M3);
//   - each package variable it reads that isn't written by its own initializer alone: a write anywhere in the files
//     read, reached code's included, through an index, a selector, a star, an address or a pointer method's receiver;
//     a use that shares its storage (its address, a slice of it, or, for a value that shares storage, the value itself
//     copied, passed, returned or sent); an init that names it; an initializer of anything but literals and pure
//     constructors; its being exported; or a go:linkname anywhere in the tree naming a reached package (the invariant);
//   - each value of an interface type the entries are handed, a parameter's or a field of the receiver's or a
//     parameter's, that a type of the tree outside Slice implements, or that names no method to follow (any);
//   - each package of the tree it reaches outside Slice, and each interface call a type of such a package implements;
//   - and each thing the census can't follow, which fails closed: a type error, a call it couldn't type, a call through a
//     function value held in a package variable or a field or its use as a value, a call through a type parameter,
//     reflection, a cgo or assembly file.
func (census Census) Findings() ([]string, error) {
	// The files of the platform the export data was built on, the box running the census: linux/amd64 at the gate, as
	// the line builds, so the files it reads and the types it reads them with are one platform's.
	context := build.Default
	context.GOOS, context.GOARCH, context.CgoEnabled = runtime.GOOS, runtime.GOARCH, true
	listed, err := context.ImportDir(census.Directory, 0)
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
		file, err := parser.ParseFile(fileSet, filepath.Join(census.Directory, name), nil, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
		production[file] = index < len(listed.GoFiles)+len(listed.CgoFiles)
	}
	// read says whether what file holds runs in a process that can build lean: the package's own files always, its test
	// files when its test binary builds lean.
	read := func(file *ast.File) bool { return production[file] || census.TestsBuildLean }
	// Real types, from export data built ahead: a type error, which no build of the package has, is the census reading
	// what the build doesn't, so it fails closed.
	imports := importer.ForCompiler(fileSet, "gc", func(path string) (io.ReadCloser, error) {
		if exported, ok := census.Exports[path]; ok {
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
	checked, _ := (&types.Config{Importer: imports, Error: checkError}).Check(census.ImportPath, fileSet, files, info)
	if checked == nil {
		return nil, fmt.Errorf("%s didn't type-check at all", census.Directory)
	}
	at := func(node ast.Node) string {
		position := fileSet.Position(node.Pos())
		return fmt.Sprintf("%s:%d", filepath.Base(position.Filename), position.Line)
	}
	fileOf := func(node ast.Node) string { return fileSet.Position(node.Pos()).Filename }
	declares := func(file string) bool { return declares(census.Root, census.Declared, file) }
	bodies := map[types.Object]ast.Node{}
	var inits []*ast.FuncDecl
	initializers := map[types.Object]ast.Expr{}
	entries := map[types.Object]bool{}
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
				if production[file] && slices.Contains(census.Entries, receiverName(declaration)+declaration.Name.Name) {
					entries[object] = true
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
	if len(entries) != len(census.Entries) {
		return nil, fmt.Errorf("%s defines %d of the entries %v", census.Directory, len(entries), census.Entries)
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
	// code dispatches to any value of it, wherever that value was made. A type of the tree outside Slice that implements
	// it is a finding, since no lean recipe keys its methods.
	implementers := func(iface *types.Interface, what string, node ast.Node) {
		implements := func(typ types.Type) bool {
			return types.Implements(typ, iface) || types.Implements(types.NewPointer(typ), iface)
		}
		for _, name := range checked.Scope().Names() {
			object, isType := checked.Scope().Lookup(name).(*types.TypeName)
			if !isType || object.IsAlias() || types.IsInterface(object.Type()) || !production[fileAt(files, fileSet, object.Pos())] {
				continue
			}
			if named, ok := object.Type().(*types.Named); ok && named.TypeParams().Len() > 0 {
				// A generic type implements iface only once instantiated, so one with any of its methods counts.
				for method := range iface.Methods() {
					if found, _, _ := types.LookupFieldOrMethod(types.NewPointer(object.Type()), true, checked, method.Name()); found != nil {
						mark(object)
						break
					}
				}
				continue
			}
			if implements(object.Type()) {
				mark(object)
			}
		}
		for _, path := range census.Tree {
			if path == census.ImportPath || slices.Contains(census.Slice, path) {
				continue
			}
			other, err := imports.Import(path)
			if err != nil {
				findings[fmt.Sprintf("%s: %s, and the census can't read %s's types: %v", at(node), what, path, err)] = true
				continue
			}
			for _, name := range other.Scope().Names() {
				if object, isType := other.Scope().Lookup(name).(*types.TypeName); isType && !object.IsAlias() && !types.IsInterface(object.Type()) && implements(object.Type()) {
					findings[fmt.Sprintf("%s: %s, which %s.%s implements, a package of the tree no lean recipe keys", at(node), what, path, name)] = true
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
					case strings.HasPrefix(path, ModulePath+"/") && !slices.Contains(census.Slice, path):
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
				if method == nil || method.Signature().Recv() == nil {
					return true
				}
				if _, isParameter := types.Unalias(selection.Recv()).(*types.TypeParam); isParameter {
					findings[fmt.Sprintf("%s: a call of %s through a type parameter, which the census can't follow", at(node), method.Name())] = true
					return true
				}
				if iface, ok := method.Signature().Recv().Type().Underlying().(*types.Interface); ok {
					implementers(iface, "an interface call of "+method.Name(), node)
				}
			case *ast.CallExpr:
				if _, typed := info.Types[node.Fun]; !typed {
					findings[fmt.Sprintf("%s: a call the census couldn't type, which it can't follow", at(node))] = true
				}
			}
			return true
		})
	}
	for entry := range entries {
		mark(entry)
		// What the entries are handed: each interface value among their parameters and the fields of their receiver and
		// parameters, followed through their types, reaches every implementer, as a call of it would (Node's ruling 3).
		signature := entry.(*types.Func).Signature()
		var handed []*types.Var
		if signature.Recv() != nil {
			handed = append(handed, signature.Recv())
		}
		for parameter := range signature.Params().Variables() {
			handed = append(handed, parameter)
		}
		for _, value := range handed {
			for _, iface := range interfacesIn(value.Type(), map[types.Type]bool{}) {
				what := fmt.Sprintf("%s is handed a value of %s", entry.Name(), iface)
				if iface.(*types.Interface).Empty() {
					findings[what+", which names no method the census can follow"] = true
					continue
				}
				implementers(iface.(*types.Interface), what, bodies[entry])
			}
		}
	}
	// Rail 1, (c): an init or an initializer the reach didn't take counts as reached when it calls anything but a pure
	// constructor, so what runs at load can't set what the build reads unseen. Each is weighed once, so a blank
	// variable's initializer, which no use reaches, can't keep the loop turning (Mentat's should on adcc8fd771).
	weighed := map[ast.Node]bool{}
	for {
		for len(queue) > 0 {
			node := queue[0]
			queue = queue[1:]
			walk(node)
		}
		grown := false
		for _, declaration := range inits {
			if !weighed[declaration] && !pure(declaration.Body, info) {
				weighed[declaration] = true
				findings[fmt.Sprintf("init in %s counts as reached", filepath.Base(fileOf(declaration)))] = !declares(fileOf(declaration))
				queue = append(queue, declaration.Body)
				grown = true
			}
		}
		for object, value := range initializers {
			if !weighed[value] && !reached[object] && !pure(value, info) {
				weighed[value] = true
				if !declares(fileOf(value)) {
					findings[fmt.Sprintf("%s: an initializer that calls, which counts as reached, in a file the build's key doesn't declare", at(value))] = true
				}
				mark(object)
				queue = append(queue, value)
				grown = true
			}
		}
		if !grown {
			break
		}
	}
	reachedPackages := []string{census.ImportPath}
	for object := range reached {
		if object == nil || !object.Pos().IsValid() {
			continue
		}
		if file := fileSet.Position(object.Pos()).Filename; !declares(file) {
			findings[fmt.Sprintf("%s, defined in %s, which the build's key doesn't declare", object.Name(), filepath.Base(file))] = true
		}
		variable, isVariable := object.(*types.Var)
		if !isVariable {
			continue
		}
		// The invariant: a package variable the build reads is written only by its own initializer, of literals and pure
		// constructors, and no package but its own can reach it.
		if variable.Exported() {
			findings[fmt.Sprintf("%s, which the build reads, is exported, so a package the census doesn't read can write it", variable.Name())] = true
		}
		if value, ok := initializers[variable]; ok && !pure(value, info) {
			findings[fmt.Sprintf("%s: %s, which the build reads, is initialized by a call", at(value), variable.Name())] = true
		}
	}
	for _, file := range files {
		if !read(file) {
			continue
		}
		own := map[ast.Node]bool{}
		for _, declaration := range file.Decls {
			if declaration, ok := declaration.(*ast.GenDecl); ok {
				for _, spec := range declaration.Specs {
					if spec, ok := spec.(*ast.ValueSpec); ok {
						for _, value := range spec.Values {
							own[value] = true
						}
					}
				}
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if own[node] {
				// A variable's own initializer: what the invariant allows, held to pure above.
				return false
			}
			switch node := node.(type) {
			case *ast.FuncDecl:
				if node.Recv == nil && node.Name.Name == "init" {
					ast.Inspect(node.Body, func(inner ast.Node) bool {
						if name, ok := inner.(*ast.Ident); ok {
							if variable, ok := info.Uses[name].(*types.Var); ok && reached[variable] {
								findings[fmt.Sprintf("%s: an init names %s, which the build reads", at(inner), variable.Name())] = true
							}
						}
						return true
					})
				}
			}
			return true
		})
		eachWrite(file, info, own, func(root types.Object, node ast.Node) {
			if variable, ok := root.(*types.Var); ok && reached[variable] {
				findings[fmt.Sprintf("%s: writes %s, which the build reads and only its initializer may write", at(node), variable.Name())] = true
			}
		})
		eachShare(file, info, own, func(variable *types.Var, node ast.Node) {
			if reached[variable] {
				findings[fmt.Sprintf("%s: shares the storage of %s, which the build reads, so a copy can write it", at(node), variable.Name())] = true
			}
		})
	}
	reachedPackages = append(reachedPackages, census.Slice...)
	for _, named := range linknames(census.Root, reachedPackages) {
		findings[named] = true
	}
	var result []string
	for finding, holds := range findings {
		if holds {
			result = append(result, finding)
		}
	}
	sort.Strings(result)
	return result, nil
}

// interfacesIn is each interface type typ holds: itself, or through its pointers, elements, keys and the fields of its
// structs, the named ones' underlying types included.
func interfacesIn(typ types.Type, seen map[types.Type]bool) []types.Type {
	if seen[typ] {
		return nil
	}
	seen[typ] = true
	switch underlying := typ.Underlying().(type) {
	case *types.Interface:
		return []types.Type{underlying}
	case *types.Pointer:
		return interfacesIn(underlying.Elem(), seen)
	case *types.Slice:
		return interfacesIn(underlying.Elem(), seen)
	case *types.Array:
		return interfacesIn(underlying.Elem(), seen)
	case *types.Chan:
		return interfacesIn(underlying.Elem(), seen)
	case *types.Map:
		return append(interfacesIn(underlying.Key(), seen), interfacesIn(underlying.Elem(), seen)...)
	case *types.Struct:
		var all []types.Type
		for field := range underlying.Fields() {
			all = append(all, interfacesIn(field.Type(), seen)...)
		}
		return all
	}
	return nil
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

// pure says whether code, an init's body or an initializer, calls nothing but builtins, conversions and pure
// constructors (PureConstructors), and holds no function literal, whose call would be a call through a value.
func pure(code ast.Node, info *types.Info) bool {
	clean := true
	ast.Inspect(code, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.FuncLit:
			clean = false
		case *ast.CallExpr:
			if typed, ok := info.Types[node.Fun]; ok && (typed.IsType() || typed.IsBuiltin()) {
				return true
			}
			if function := calledFunction(node, info); function == nil || !slices.Contains(PureConstructors, function.FullName()) {
				clean = false
			}
		}
		return clean
	})
	return clean
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

// eachWrite calls visit with the root object of each place code writes, and the node writing it (M1), outside the
// nodes skip names: an assignment's or an increment's target, a range's assigned key and value, a value whose address
// is taken, the first argument of copy, clear or delete, and the receiver a pointer method takes the address of. A
// root is the variable a target is reached through, by indexes, slices, field selectors and stars: features for
// features[0] = x, r for r.L.x = y.
func eachWrite(code ast.Node, info *types.Info, skip map[ast.Node]bool, visit func(root types.Object, node ast.Node)) {
	target := func(expression ast.Expr, node ast.Node) {
		if root := writeRoot(expression, info); root != nil {
			visit(root, node)
		}
	}
	ast.Inspect(code, func(node ast.Node) bool {
		if skip[node] {
			return false
		}
		switch node := node.(type) {
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

// eachShare calls visit with each package variable code uses in a way that shares its storage, and the node using it,
// outside the nodes skip names: sliced, or, when it holds what a copy shares (SharesStorage), its value copied into
// another variable or a literal, passed (a method's receiver included), returned, sent or converted. Reading it through
// an index or a field, ranging over it, comparing it and asking len or cap share nothing. Taking its address is a
// write (eachWrite).
func eachShare(code ast.Node, info *types.Info, skip map[ast.Node]bool, visit func(variable *types.Var, node ast.Node)) {
	var parents []ast.Node
	ast.Inspect(code, func(node ast.Node) bool {
		if node == nil {
			parents = parents[:len(parents)-1]
			return true
		}
		if skip[node] {
			return false
		}
		defer func() { parents = append(parents, node) }()
		name, ok := node.(*ast.Ident)
		if !ok || len(parents) == 0 {
			return true
		}
		variable, ok := info.Uses[name].(*types.Var)
		if !ok || variable.IsField() || variable.Pkg() == nil || variable.Parent() != variable.Pkg().Scope() {
			return true
		}
		// The expression the variable stands in: itself, or a selector naming it from another package.
		var expression ast.Expr = name
		index := len(parents) - 1
		if selector, ok := parents[index].(*ast.SelectorExpr); ok && selector.Sel == name {
			expression, index = selector, index-1
		}
		for index >= 0 {
			if parenthesized, ok := parents[index].(*ast.ParenExpr); ok {
				expression, index = parenthesized, index-1
				continue
			}
			break
		}
		if index < 0 {
			return true
		}
		if sliced, ok := parents[index].(*ast.SliceExpr); ok && sliced.X == expression {
			visit(variable, sliced)
			return true
		}
		if !SharesStorage(variable.Type(), map[types.Type]bool{}) {
			return true
		}
		switch parent := parents[index].(type) {
		case *ast.IndexExpr, *ast.IndexListExpr, *ast.RangeStmt, *ast.BinaryExpr, *ast.UnaryExpr, *ast.StarExpr:
			// Read through, ranged over, compared, its address taken (a write) or read through a pointer.
			return true
		case *ast.SelectorExpr:
			// A field read shares nothing; a method takes its receiver, a pointer method's address (a write).
			if selection := info.Selections[parent]; selection == nil || selection.Kind() == types.FieldVal {
				return true
			}
		case *ast.CallExpr:
			if builtin, ok := info.Uses[identOf(parent.Fun)].(*types.Builtin); ok {
				switch builtin.Name() {
				case "len", "cap":
					return true
				case "copy":
					if len(parent.Args) > 1 && parent.Args[1] == expression {
						return true
					}
				}
			}
		case *ast.AssignStmt:
			if slices.Contains(parent.Lhs, expression) {
				// The target of a write, which eachWrite names.
				return true
			}
		}
		visit(variable, node)
		return true
	})
}

// SharesStorage says whether a copy of a value of typ shares storage with it, so writing the copy writes the original:
// a pointer, slice, map, channel, function or interface, or a struct or array holding one.
func SharesStorage(typ types.Type, seen map[types.Type]bool) bool {
	if seen[typ] {
		return false
	}
	seen[typ] = true
	switch typ := typ.Underlying().(type) {
	case *types.Pointer, *types.Slice, *types.Map, *types.Chan, *types.Signature, *types.Interface:
		return true
	case *types.Array:
		return SharesStorage(typ.Elem(), seen)
	case *types.Struct:
		for field := range typ.Fields() {
			if SharesStorage(field.Type(), seen) {
				return true
			}
		}
	}
	return false
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

// linknames is a finding for each //go:linkname in a Go file of the tree under root naming a symbol of one of packages,
// which reaches its variables past every rule (Mentat's N5: another package's linkname wrote a's unexported variable on
// go1.27.1). As compilercensus walks it: never cohere, another module, nor a testdata, node_modules or hidden directory.
func linknames(root string, packages []string) []string {
	var found []string
	filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		relative, _ := filepath.Rel(root, path)
		if entry.IsDir() {
			if relative == "cohere" || entry.Name() == "testdata" || entry.Name() == "node_modules" || strings.HasPrefix(entry.Name(), ".") && relative != "." {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil || !strings.Contains(string(content), "go:linkname") {
			return nil
		}
		for number, line := range strings.Split(string(content), "\n") {
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) < 3 || fields[0] != "//go:linkname" {
				continue
			}
			for _, pkg := range packages {
				if strings.HasPrefix(fields[2], pkg+".") {
					found = append(found, fmt.Sprintf("%s:%d: a go:linkname names %s, which reaches past the census", filepath.ToSlash(relative), number+1, fields[2]))
				}
			}
		}
		return nil
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
