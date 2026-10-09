package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/go/packages"
	"log"
	"maps"
	"os"
	"path"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

const tsgoInternalPrefix = "github.com/microsoft/TypeScript/tsc/internal/"

type ExtraShim struct {
	// ExtraFunctions and ExtraMethods name upstream's unexported functions and methods to re-export by
	// `//go:linkname`, with no mirror, so the risk is the signature rather than a layout.
	//
	// # An upstream sync re-checks every name here
	//
	// A linkname binds by name and the declaration carries the signature it expects, so a sync that
	// renames one fails the link, and one that changes its parameters regenerates a new declaration
	// and fails at the caller. Before landing a sync, find each name on the new pin
	// (`git grep '^func (c \*Checker) <name>('`) and read its callers if the signature moved. The
	// newest, for lint's adamic rules: getNonMissingTypeOfSymbol (#sp4xwtj), instantiateType and
	// newTypeMapper (#ncheh9w), getCanonicalSignature and instantiateSignatureInContextOf (#bbtfx99), each
	// checked on d92d9bfee.
	ExtraFunctions []string
	ExtraMethods   map[string]([]string)

	// ExtraFields names unexported struct fields to generate a reader for.
	//
	// # Adding a name here obliges a row in TestShimFieldAccessorsReadTheFieldsTheyName
	//
	// A generated reader dereferences a mirror by field offset through `unsafe.Pointer`, so an
	// accessor that reads the WRONG field compiles, runs, and returns a plausible value. The only
	// thing that catches it is `internal/types/program/shim_layout_test.go`, which compares each
	// accessor's result against the checker's own name for what it should have returned.
	//
	// That test is a table, and a table does not fail because it is short. `Checker_stringType` was
	// added and shipped with the table still listing three accessors beside four fields: the suite
	// stayed green, the guard read as covering the group, and the gap opened in silence. Closed at
	// `0fd5095`, by whoever noticed rather than by anything in the workflow.
	//
	// So the obligation is stated here, at the file somebody edits to add a field, because nothing
	// downstream prompts it. Add the name, regenerate, then add the row, and prove the row can fail
	// by pointing it at a different accessor before trusting it.
	ExtraFields     map[string]([]string)
	IgnoreFunctions []string

	// TypeSubstitutions rewrites a type expression in a mirrored struct body.
	//
	// A mirror exists to reproduce a struct's memory layout so `unsafe` can reach a field by offset;
	// it never calls anything on these types. So a field whose type is unexported cannot be named
	// across the module boundary, but any type with identical size and alignment stands in perfectly.
	//
	// This is configuration rather than a code change because the substitutions are upstream's private
	// vocabulary, and it drifts. Four were needed for the move from typescript-go to
	// microsoft/TypeScript alone.
	//
	// PER-FIELD OFFSET IS THE CONTRACT. Identical size and alignment is the cheap proxy that catches
	// most violations of it, and verifyTypeSubstitutions below enforces that proxy, but the two are
	// not the same claim: a stand-in of the right width whose fields are in the wrong ORDER passes
	// every size comparison and still moves the fields that follow it. That was measured, not
	// reasoned about — swapping two 8-byte pointer fields in a mirror survived a whole-struct size
	// check while breaking the read. `reflect` can enumerate all 319 of Checker's fields with their
	// names and offsets even though they are unexported, so a per-field guard is possible; it has
	// not been built, and this comment is the record of that gap.
	//
	// A substitution that violates the proxy fails silently in the worst available way.
	// `checker.symbolArenaLinkStore` was substituted with `core.PagedLinkStore`, which is the store
	// half of it and omits the arena half, leaving the mirror 24 bytes short at field 99 of 319.
	// Every field after it then read at the wrong offset: `Checker_numberType` returned the `null`
	// type and `Checker_booleanType` returned `false`. Nothing crashed, no test failed, and the one
	// rule that depended on it went silent on every input. Anything mirroring a struct with more
	// than one field needs ExtraDeclarations below rather than a bare rename.
	//
	// WHERE a substitution appears decides whether its width is load-bearing at all, and the four
	// here are not four equal risks. Three of them are only ever reached inside a map type
	// (`map[*ast.Symbol]int32`, `map[uint64][]*ast.Symbol`), and a Go map is one pointer word
	// whatever its key and value widths are, so those three cannot shift an offset at any size.
	// Only `nodeLinkStore` at field 91 and `symbolArenaLinkStore` at field 99 sit inline, where
	// width is what the layout is made of. Ask where a type lands in the mirror before asking
	// whether its stand-in is the right size.
	TypeSubstitutions map[string]string

	// ExtraDeclarations are emitted verbatim into the generated file, ahead of the mirrors.
	//
	// TypeSubstitutions can only rewrite the head of a rendered type expression, so a generic
	// upstream type keeps its `[V]` suffix and the stand-in has to be a generic named type too. An
	// inline anonymous struct cannot carry that suffix, and the shim packages are wholly generated,
	// so there is nowhere else in them to put such a declaration. This is that place.
	ExtraDeclarations []string
}

func main() {
	packagesToShim := []string{
		"ast",
		"bundled",
		"checker",
		"compiler",
		"core",
		"scanner",
		"tsoptions",
		"tspath",
		"vfs",
		"vfs/cachedvfs",
		"vfs/osvfs",
	}

	packagesToShimFullNames := make([]string, len(packagesToShim))
	for i, pkg := range packagesToShim {
		packagesToShimFullNames[i] = tsgoInternalPrefix + pkg
	}

	packages, err := packages.Load(&packages.Config{
		// TODO: path relative to repo root
		Dir:  "./TypeScript-shim/compiler",
		Mode: packages.LoadSyntax,
	}, packagesToShimFullNames...)
	if err != nil {
		log.Fatalf("Error loading package: %v", err)
	}

	var shimHeaderBuilder strings.Builder
	var shimBuilder strings.Builder
	var tempBuffer bytes.Buffer

	for _, pkg := range packages {
		shimDirPath := path.Join("./TypeScript-shim/", strings.TrimPrefix(pkg.PkgPath, tsgoInternalPrefix))
		var extraShim ExtraShim
		extraShimFilePath := path.Join(shimDirPath, "extra-shim.json")
		if data, err := os.ReadFile(extraShimFilePath); err == nil {
			if err := json.Unmarshal(data, &extraShim); err != nil {
				fmt.Printf("error parsing %v: %v", extraShimFilePath, err)
				return
			}
		}
		if extraShim.ExtraMethods == nil {
			extraShim.ExtraMethods = map[string][]string{}
		}
		if extraShim.ExtraFunctions == nil {
			extraShim.ExtraFunctions = []string{}
		}
		if extraShim.ExtraFields == nil {
			extraShim.ExtraFields = map[string]([]string){}
		}
		if extraShim.TypeSubstitutions == nil {
			extraShim.TypeSubstitutions = map[string]string{}
		}

		// Refuse a bad substitution rather than emitting one. This runs before this package's own
		// shim is written, so a mirror whose stand-ins are the wrong width never reaches the tree.
		//
		// The abort is not transactional across packages: shims for packages processed earlier in
		// this loop are already on disk when it fires. That is deliberate rather than overlooked —
		// making it atomic would mean buffering every package before writing any, and the failure
		// it would prevent is a stale shim for an UNRELATED package, which the next successful run
		// overwrites. What must never exist is the bad mirror itself, and that is what this stops.
		if problems := verifyTypeSubstitutions(pkg, packages, extraShim); len(problems) > 0 {
			fmt.Printf("ERROR: %v declares type substitutions that do not preserve layout.\n",
				extraShimFilePath)
			for _, problem := range problems {
				fmt.Printf("  - %v\n", problem)
			}
			os.Exit(1)
		}

		for _, declaration := range extraShim.ExtraDeclarations {
			shimBuilder.WriteString(declaration)
			shimBuilder.WriteString("\n")
		}
		if extraShim.IgnoreFunctions == nil {
			extraShim.IgnoreFunctions = []string{}
		}

		// true if directly used, false otherwise
		importedPackages := map[string]bool{}

		importPackage := func(pkg string, directly bool) {
			if directly {
				importedPackages[pkg] = true
			} else if _, ok := importedPackages[pkg]; !ok {
				importedPackages[pkg] = false
			}
		}

		var qualifierOnlyPackageName types.Qualifier = func(p *types.Package) string {
			importPackage(p.Path(), true)
			return p.Name()
		}
		var qualifierEmptyPackageName types.Qualifier = func(p *types.Package) string {
			return ""
		}

		emitGoLinknameDirective := func(localName string, fn *types.Func) {
			// //go:linkname only allowed in Go files that import "unsafe"
			importPackage("unsafe", false)
			importPackage(pkg.Types.Path(), false)
			shimBuilder.WriteString("//go:linkname ")
			shimBuilder.WriteString(localName)
			shimBuilder.WriteByte(' ')
			shimBuilder.WriteString(fn.Pkg().Path())
			shimBuilder.WriteByte('.')
			if recv := fn.Signature().Recv(); recv != nil {
				shimBuilder.WriteByte('(')
				shimBuilder.WriteString(types.TypeString(recv.Type(), qualifierEmptyPackageName))
				shimBuilder.WriteByte(')')
				shimBuilder.WriteByte('.')
			}
			shimBuilder.WriteString(fn.Name())
			shimBuilder.WriteByte('\n')
		}

		emitLinkedFunction := func(fn *types.Func) bool {
			if fn.Signature().TypeParams() != nil {
				// https://github.com/golang/go/issues/60425
				// linking to functions with generics is not supported in go:linkname
				return false
			}
			name := cases.Title(language.English, cases.NoLower).String(fn.Name())
			emitGoLinknameDirective(name, fn)
			shimBuilder.WriteString("func ")
			shimBuilder.WriteString(name)
			types.WriteSignature(&tempBuffer, fn.Signature(), qualifierOnlyPackageName)
			shimBuilder.Write(tempBuffer.Bytes())
			tempBuffer.Reset()
			shimBuilder.WriteString("\n")
			return true
		}

		matchedExtraFunctions := make(map[string]bool, len(extraShim.ExtraFunctions))
		for _, name := range extraShim.ExtraFunctions {
			matchedExtraFunctions[name] = false
		}
		matchedExtraMethods := make(map[string](map[string]bool), len(extraShim.ExtraMethods))
		for name, methods := range extraShim.ExtraMethods {
			matchedExtraMethods[name] = make(map[string]bool, len(methods))
			for _, method := range methods {
				matchedExtraMethods[name][method] = false
			}
		}
		matchedExtraFields := make(map[string]bool, len(extraShim.ExtraFields))
		for name := range extraShim.ExtraFields {
			matchedExtraFields[name] = false
		}

		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			object := scope.Lookup(name)
			if !object.Exported() {
				fn, isFunc := object.(*types.Func)
				if _, exists := matchedExtraFunctions[name]; isFunc && exists {
					if emitLinkedFunction(fn) {
						matchedExtraFunctions[name] = true
					}
				}
				continue
			}

			printReexport := func(kind string) {
				importPackage(pkg.Types.Path(), true)
				shimBuilder.WriteString(kind)
				shimBuilder.WriteString(" ")
				shimBuilder.WriteString(name)
				shimBuilder.WriteString(" = ")
				shimBuilder.WriteString(pkg.Name)
				shimBuilder.WriteString(".")
				shimBuilder.WriteString(name)
				shimBuilder.WriteString("\n")
			}

			switch object.(type) {
			case *types.TypeName:
				typeName := object.(*types.TypeName)
				t := typeName.Type()
				named, isNamed := t.(*types.Named)
				if isNamed {
					_, nameWithTypeParams, _ := strings.Cut(types.TypeString(named, qualifierOnlyPackageName), ".")
					importPackage(pkg.Types.Path(), true)
					shimBuilder.WriteString("type ")
					shimBuilder.WriteString(nameWithTypeParams)
					shimBuilder.WriteString(" = ")
					shimBuilder.WriteString(pkg.Name)
					shimBuilder.WriteString(".")
					shimBuilder.WriteString(name)

					typeParams := slices.Collect(named.TypeParams().TypeParams())
					if len(typeParams) > 0 {
						// (*typeWriter)typeList
						shimBuilder.WriteByte('[')
						for i, typ := range typeParams {
							if i > 0 {
								shimBuilder.WriteByte(',')
							}
							shimBuilder.WriteString(typ.String())
						}
						shimBuilder.WriteByte(']')
					}

					shimBuilder.WriteString("\n")
				} else {
					printReexport("type")
				}

				if extraMethods, ok := matchedExtraMethods[name]; isNamed && ok {
					for method := range named.Methods() {
						methodName := method.Name()
						if _, exists := extraMethods[methodName]; !exists {
							continue
						}
						extraMethods[methodName] = true
						prefix := name + "_"
						emitGoLinknameDirective(prefix+methodName, method)
						funcDeclStr := types.ObjectString(method, qualifierOnlyPackageName)
						recvStart := 0
						recvEnd := 0
						paramsStart := 0
						for i, s := range funcDeclStr {
							if s == '(' {
								if recvStart == 0 {
									recvStart = i + 1
								}
								if recvEnd != 0 {
									paramsStart = i + 1
									break
								}
							}
							if s == ')' && recvEnd == 0 {
								recvEnd = i
							}
						}
						shimBuilder.WriteString("func ")
						shimBuilder.WriteString(prefix)
						shimBuilder.WriteString(funcDeclStr[recvEnd+2 : paramsStart])
						shimBuilder.WriteString("recv ")
						shimBuilder.WriteString(funcDeclStr[recvStart:recvEnd])
						if method.Signature().Params() != nil {
							shimBuilder.WriteString(", ")
						}
						shimBuilder.WriteString(funcDeclStr[paramsStart:])
						shimBuilder.WriteString("\n")
					}
				}

				if _, ok := matchedExtraFields[name]; isNamed && ok {
					importPackage("unsafe", true)

					matchedExtraFields[name] = true
					if err != nil {
						log.Fatalf("error formatting %v struct body: %v", name, err)
					}
					mirrorStructName := "extra_" + name

					var emitExtraStruct func(name string, s *types.Struct)
					emitExtraStruct = func(name string, s *types.Struct) {
						shimBuilder.WriteString("type extra_")
						shimBuilder.WriteString(name)
						shimBuilder.WriteString(" struct {")

						dependencies := [](struct {
							string
							*types.Struct
						}){}
						for field := range s.Fields() {
							shimBuilder.WriteString("\n  ")
							if !field.Embedded() {
								shimBuilder.WriteString(field.Name())
								shimBuilder.WriteByte(' ')
							}

							ptrType, ok := field.Type().(*types.Pointer)
							if ok {
								named, ok := ptrType.Elem().(*types.Named)
								if ok && !named.Obj().Exported() {
									strct, ok := named.Underlying().(*types.Struct)
									if ok {
										n := named.Obj().Name()
										dependencies = append(dependencies, struct {
											string
											*types.Struct
										}{n, strct})
										shimBuilder.WriteString("extra_")
										shimBuilder.WriteString(n)
										continue
									}
								}
							}

							fieldType := types.TypeString(field.Type(), qualifierOnlyPackageName)
							for from, to := range extraShim.TypeSubstitutions {
								fieldType = strings.ReplaceAll(fieldType, from, to)
							}
							shimBuilder.WriteString(fieldType)
						}
						shimBuilder.WriteString("\n}\n")

						for _, dep := range dependencies {
							emitExtraStruct(dep.string, dep.Struct)
						}
					}

					strct, ok := named.Underlying().(*types.Struct)
					if !ok {
						log.Fatalf("expected %v to be struct", name)
					}

					emitExtraStruct(name, strct)

					mappedFieldTypes := make(map[string]*types.Var, strct.NumFields())
					for field := range strct.Fields() {
						mappedFieldTypes[field.Name()] = field
					}

					for _, field := range extraShim.ExtraFields[name] {
						shimBuilder.WriteString("func ")
						shimBuilder.WriteString(name)
						shimBuilder.WriteByte('_')
						shimBuilder.WriteString(field)
						shimBuilder.WriteString("(v *")
						shimBuilder.WriteString(pkg.Name)
						shimBuilder.WriteByte('.')
						shimBuilder.WriteString(name)
						shimBuilder.WriteString(") ")

						fieldVar, ok := mappedFieldTypes[field]
						if !ok {
							log.Fatalf("expected struct %q to contain field %q", name, field)
						}
						shimBuilder.WriteString(types.TypeString(fieldVar.Type(), qualifierOnlyPackageName))
						shimBuilder.WriteString(" {\n")
						shimBuilder.WriteString("  return ((*")
						shimBuilder.WriteString(mirrorStructName)
						shimBuilder.WriteString(")(unsafe.Pointer(v))).")
						shimBuilder.WriteString(field)
						shimBuilder.WriteString("\n")
						shimBuilder.WriteString("}\n")
					}
				}
			case *types.Const:
				printReexport("const")
			case *types.Var:
				printReexport("var")
			case *types.Func:
				if !slices.Contains(extraShim.IgnoreFunctions, name) {
					funcType := object.(*types.Func)
					emitLinkedFunction(funcType)
				}
			}
		}

		exit := false
		for fnName, found := range matchedExtraFunctions {
			if found {
				continue
			}
			fmt.Printf("ERROR: couldn't find %v function\n", fnName)
			exit = true
		}
		for name, methods := range matchedExtraMethods {
			for methodName, found := range methods {
				if found {
					continue
				}
				fmt.Printf("ERROR: couldn't find %v.%v method\n", name, methodName)
				exit = true
			}
		}
		if exit {
			os.Exit(1)
		}

		// https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source
		shimHeaderBuilder.WriteString("\n// Code generated by internal/types/tools/generate_shims. DO NOT EDIT.\n\n")
		shimHeaderBuilder.WriteString("package ")
		shimHeaderBuilder.WriteString(pkg.Name)
		shimHeaderBuilder.WriteString("\n\n")
		importsList := slices.Collect(maps.Keys(importedPackages))
		slices.Sort(importsList)
		for _, imported := range importsList {
			shimHeaderBuilder.WriteString("import ")
			if !importedPackages[imported] {
				shimHeaderBuilder.WriteString("_ ")
			}
			shimHeaderBuilder.WriteString("\"")
			shimHeaderBuilder.WriteString(imported)
			shimHeaderBuilder.WriteString("\"\n")
		}
		shimHeaderBuilder.WriteString("\n")

		// The shim is printed through gofmt's own printer before it is written, so what lands in the tree is
		// what gofmt would leave it as. The builders above emit declarations one per line with no grouping,
		// and the files they wrote went unformatted into the tree, where a gofmt guard over it would fail on
		// output nobody may edit by hand. A source that does not parse is refused rather than written: the
		// printer cannot format it, and a shim that does not compile is worse than the stale one it replaces.
		shimGoPath := path.Join(shimDirPath, "shim.go")
		formatted, err := format.Source([]byte(shimHeaderBuilder.String() + shimBuilder.String()))
		if err != nil {
			log.Fatalf("error formatting %v, so it was not written: %v", shimGoPath, err)
		}
		if err := os.WriteFile(shimGoPath, formatted, 0o644); err != nil {
			log.Fatalf("error writing %v: %v", shimGoPath, err)
		}

		shimHeaderBuilder.Reset()
		shimBuilder.Reset()
	}
}

// verifyTypeSubstitutions refuses to emit a mirror whose stand-in types do not match the size and
// alignment of the upstream types they replace.
//
// This is the check that the `checker.symbolArenaLinkStore` defect needed and did not have. A
// mirror exists so `unsafe` can reach a field by offset, so a stand-in that is the wrong width
// shifts every field declared after it, and the read then returns a perfectly valid value of the
// right Go type that is simply not the field asked for. Nothing crashes and no test fails; the one
// rule downstream of it goes silent. The comment on TypeSubstitutions above has said IDENTICAL SIZE
// IS THE WHOLE CONTRACT since that defect was fixed, but a comment cannot enforce a contract, and
// the substitution that violated it was accepted by this generator without a word.
//
// So this runs before anything is written. A check here is worth more than a check in a test
// because it fails before the wrong code exists, and it covers substitutions no accessor reads
// today — a field read added later inherits a layout that was already proven rather than one that
// happened to be right.
//
// Alignment is checked alongside size because a stand-in can be the right width by accident and
// still place the fields after it wrong.
func verifyTypeSubstitutions(
	upstreamPackage *packages.Package,
	allPackages []*packages.Package,
	extraShim ExtraShim,
) []string {
	if len(extraShim.TypeSubstitutions) == 0 {
		return nil
	}

	sizes := types.SizesFor("gc", runtime.GOARCH)
	if sizes == nil {
		return []string{fmt.Sprintf("no go/types size model for GOARCH %q", runtime.GOARCH)}
	}

	packagesByPath := map[string]*types.Package{}
	packagesByName := map[string]*types.Package{}
	for _, loaded := range allPackages {
		packagesByPath[loaded.Types.Path()] = loaded.Types
		packagesByName[loaded.Types.Name()] = loaded.Types
	}

	// Stand-ins declared in ExtraDeclarations exist only as source text, so they are type-checked
	// into a synthetic package here in order to be measurable at all.
	declarationScope := typeCheckExtraDeclarations(extraShim.ExtraDeclarations, packagesByPath, sizes)

	// A generic stand-in must match for every instantiation, not one. These arguments differ in
	// width, alignment and pointer-ness, so a stand-in that only matches for one shape is caught.
	probeArguments := []types.Type{
		types.Typ[types.Int8],
		types.Typ[types.Int],
		types.NewPointer(types.Typ[types.Int]),
		types.NewArray(types.Typ[types.Int64], 5),
	}

	resolve := func(qualified string) types.Type {
		packageName, typeName, found := strings.Cut(qualified, ".")
		if !found {
			// An unqualified name is either a builtin or an ExtraDeclarations stand-in.
			if builtin := types.Universe.Lookup(qualified); builtin != nil {
				return builtin.Type()
			}
			if declarationScope != nil {
				if object := declarationScope.Lookup(qualified); object != nil {
					return object.Type()
				}
			}
			return nil
		}
		hostPackage, ok := packagesByName[packageName]
		if !ok {
			return nil
		}
		object := hostPackage.Scope().Lookup(typeName)
		if object == nil {
			return nil
		}
		return object.Type()
	}

	instantiate := func(base types.Type, argument types.Type) (types.Type, error) {
		named, ok := base.(*types.Named)
		if !ok || named.TypeParams().Len() == 0 {
			return base, nil
		}
		arguments := make([]types.Type, named.TypeParams().Len())
		for i := range arguments {
			arguments[i] = argument
		}
		return types.Instantiate(nil, named, arguments, false)
	}

	problems := []string{}
	for _, upstreamName := range slices.Sorted(maps.Keys(extraShim.TypeSubstitutions)) {
		standInName := extraShim.TypeSubstitutions[upstreamName]

		upstreamType := resolve(upstreamName)
		if upstreamType == nil {
			problems = append(problems, fmt.Sprintf(
				"substitution %q -> %q: could not resolve the upstream type, so its size was never "+
					"compared. An unresolvable substitution is not a passing one: if upstream renamed "+
					"or removed this type, the mirror is being built against a type that no longer "+
					"exists.", upstreamName, standInName))
			continue
		}
		standInType := resolve(standInName)
		if standInType == nil {
			problems = append(problems, fmt.Sprintf(
				"substitution %q -> %q: could not resolve the stand-in type. If it is declared in "+
					"ExtraDeclarations, that block failed to type-check; if it names a package, that "+
					"package is not among the ones loaded here.", upstreamName, standInName))
			continue
		}

		for _, argument := range probeArguments {
			upstreamInstance, err := instantiate(upstreamType, argument)
			if err != nil {
				continue
			}
			standInInstance, err := instantiate(standInType, argument)
			if err != nil {
				problems = append(problems, fmt.Sprintf(
					"substitution %q -> %q: the stand-in could not be instantiated at %s while the "+
						"upstream type could, so their shapes differ: %v",
					upstreamName, standInName, argument, err))
				break
			}

			upstreamSize := sizes.Sizeof(upstreamInstance)
			standInSize := sizes.Sizeof(standInInstance)
			upstreamAlign := sizes.Alignof(upstreamInstance)
			standInAlign := sizes.Alignof(standInInstance)
			if upstreamSize == standInSize && upstreamAlign == standInAlign {
				continue
			}
			problems = append(problems, fmt.Sprintf(
				"substitution %q -> %q at [%s]: upstream is size %d align %d, the stand-in is size "+
					"%d align %d (off by %d bytes).\n"+
					"    A mirror is read through unsafe.Pointer by field offset, so a stand-in of "+
					"the wrong width shifts every field declared after it and every accessor past "+
					"that point silently returns the wrong field.\n"+
					"    Fix the stand-in rather than this check: give it the same shape as "+
					"upstream, adding an ExtraDeclarations struct if upstream has more than one "+
					"field.",
				upstreamName, standInName, argument,
				upstreamSize, upstreamAlign, standInSize, standInAlign,
				upstreamSize-standInSize))
			break
		}
	}
	return problems
}

// typeCheckExtraDeclarations type-checks the ExtraDeclarations block on its own so the stand-ins
// declared there can be measured. It returns nil when there is nothing to check; a block that does
// not compile is reported by the caller as an unresolvable stand-in rather than silently skipped.
func typeCheckExtraDeclarations(
	declarations []string,
	packagesByPath map[string]*types.Package,
	sizes types.Sizes,
) *types.Scope {
	if len(declarations) == 0 {
		return nil
	}

	var sourceBuilder strings.Builder
	sourceBuilder.WriteString("package extradeclarations\n\n")
	// Named rather than blank imports: a blank import does not bind the package name, so a
	// declaration referring to `core.Arena` would not resolve.
	for _, path := range slices.Sorted(maps.Keys(packagesByPath)) {
		sourceBuilder.WriteString("import ")
		sourceBuilder.WriteString(packagesByPath[path].Name())
		sourceBuilder.WriteString(" ")
		sourceBuilder.WriteString(strconv.Quote(path))
		sourceBuilder.WriteString("\n")
	}
	sourceBuilder.WriteString("\n")
	for _, declaration := range declarations {
		sourceBuilder.WriteString(declaration)
		sourceBuilder.WriteString("\n")
	}

	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, "extra-declarations.go", sourceBuilder.String(), 0)
	if err != nil {
		return nil
	}
	configuration := types.Config{
		Sizes: sizes,
		Importer: importerFromMap(func(path string) (*types.Package, error) {
			if found, ok := packagesByPath[path]; ok {
				return found, nil
			}
			return nil, fmt.Errorf("package %q is not loaded", path)
		}),
		// The declarations reference imports the blank-import list above already covers; an
		// unused-import complaint would be noise rather than a layout problem.
		DisableUnusedImportCheck: true,
	}
	checked, err := configuration.Check("extradeclarations", fileSet, []*ast.File{parsed}, nil)
	if err != nil {
		return nil
	}
	return checked.Scope()
}

type importerFromMap func(string) (*types.Package, error)

func (i importerFromMap) Import(path string) (*types.Package, error) { return i(path) }
