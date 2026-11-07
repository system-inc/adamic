package ir_test

import (
	"encoding/json"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// x/tools is not a dependency here. go list supplies the module-aware export
// files, built ahead as a product (callTargetExports), and go/types resolves
// selectors, including aliases and embedded fields.
// Entries name a file, enclosing Go function and IR field, so a migrated reader
// cannot leave an unrelated reader in that file exempt. Owners identify the unit
// responsible for routing the remaining analyses. Emission entries are permanent
// while they emit the call; even those must still contain their stated read.
type targetReader struct{ owner, reason string }

var targetReaders = map[string]targetReader{
	"internal/native/region.go:environmentEscapes:CallClosure.Closure": {
		"runtime",
		"follows carrier lifetime during evaluation; target identity comes from ClosureTargets",
	},
	"internal/lower/library_node_process.go:nodeProcessStorageABI:CallClosure.Closure": {
		"compiler",
		"inspects the property spelling to validate a possible performance host method ABI, not its function targets",
	},
	"internal/native/arguments_length.go:spreadArguments:Call.Function": {
		"runtime",
		"names the spread call signature",
	},
	"internal/flow/build.go:CanThrow:ArraySort.Comparator": {
		"compiler",
		"MayThrow and library failure propagation, routed in step 2",
	},
	"internal/lower/exceptions.go:libraryFailure:ArraySort.Comparator": {
		"compiler",
		"MayThrow and library failure propagation, routed in step 2",
	},
	"internal/lower/exceptions.go:throwsOut:ArraySort.Comparator": {
		"compiler",
		"MayThrow and library failure propagation, routed in step 2",
	},
	"internal/lower/class_static_guard_test.go:TestClassStaticInitializerCallIsEmitted:Call.Function": {
		"compiler",
		"names the function under test",
	},
	"internal/native/class_inheritance.go:callCode:Call.Function": {
		"runtime",
		"names the function to call",
	},
	"internal/native/emit_arrays.go:comparator:ArraySort.Comparator": {
		"runtime",
		"names the function to call",
	},
	"internal/native/emit_expressions.go:evaluate:Call.Function": {
		"runtime",
		"names the function to call",
	},
	"internal/native/emit_expressions.go:evaluate:CallClosure.Closure": {
		"runtime",
		"names the function to call",
	},
	"internal/native/emit_functions.go:arguments:Call.Function": {
		"runtime",
		"names the function to call",
	},
	"internal/native/emit_functions.go:callThrough:CallClosure.Closure": {
		"runtime",
		"names the function to call",
	},
	"internal/native/loop_borrow_test.go:TestGlobalArgumentLending:Call.Function": {
		"runtime",
		"names the function under test",
	},
	"internal/native/loop_borrow_test.go:TestLoopCallCoverage:Call.Function": {
		"runtime",
		"names the function under test",
	},
}

func checkCallTargetReaders(test *testing.T, selected string) {
	test.Helper()
	root, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	listingProduct := callTargetExports(test, selected)
	encoded, readError := os.ReadFile(filepath.Join(listingProduct, "listing.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var packages []buildcache.ListedPackage
	if decodeError := json.Unmarshal(encoded, &packages); decodeError != nil {
		test.Fatal(decodeError)
	}
	exports := map[string]string{}
	for index, listedPackage := range packages {
		if listedPackage.Export != "" {
			listedPackage.Export = filepath.Join(listingProduct, listedPackage.Export)
		}
		if listedPackage.Dir != "" {
			listedPackage.Dir = filepath.Join(root, listedPackage.Dir)
		}
		packages[index] = listedPackage
		exports[listedPackage.ImportPath] = listedPackage.Export
		// External tests may use declarations from export_test.go. Import their
		// package's test variant rather than the production-only export file.
		if path, _, variant := strings.Cut(
			listedPackage.ImportPath,
			" [",
		); variant && (path == listedPackage.ForTest || exports[path] == "") {
			exports[path] = listedPackage.Export
		}
	}
	fset := token.NewFileSet()
	imports := importer.ForCompiler(
		fset,
		"gc",
		func(path string) (io.ReadCloser, error) { return os.Open(exports[path]) },
	)
	irPackage, importError := imports.Import("github.com/system-inc/adamic/internal/ir")
	if importError != nil {
		test.Fatal(importError)
	}
	protected := map[types.Object]string{}
	for _, field := range []string{"Call.Function", "CallClosure.Closure", "ArraySort.Comparator"} {
		parts := strings.Split(field, ".")
		structure := irPackage.Scope().Lookup(parts[0]).Type().Underlying().(*types.Struct)
		for index := 0; index < structure.NumFields(); index++ {
			if structure.Field(index).Name() == parts[1] {
				protected[structure.Field(index)] = field
			}
		}
	}
	seen := map[string]bool{}
	for _, listedPackage := range packages {
		if strings.Contains(listedPackage.ImportPath, "[") ||
			!strings.HasPrefix(listedPackage.ImportPath, "github.com/system-inc/adamic/internal/") {
			continue
		}
		base := filepath.Base(listedPackage.Dir)
		if base != selected {
			continue
		}
		if listedPackage.ImportPath != "github.com/system-inc/adamic/internal/"+base ||
			(base != "native" && base != "lower" && base != "fresh" && base != "flow") {
			continue
		}
		for group, names := range [][]string{
			append(listedPackage.GoFiles, listedPackage.TestGoFiles...),
			listedPackage.XTestGoFiles,
		} {
			if len(names) == 0 {
				continue
			}
			packagePath := listedPackage.ImportPath
			if group == 1 {
				packagePath += "_test"
			}
			files := []*ast.File{}
			for _, name := range names {
				file, parseFileError := parser.ParseFile(fset, filepath.Join(listedPackage.Dir, name), nil, 0)
				if parseFileError != nil {
					test.Fatal(parseFileError)
				}
				files = append(files, file)
			}
			info := &types.Info{Selections: map[*ast.SelectorExpr]*types.Selection{}}
			configuration := types.Config{Importer: imports}
			if _, checkError := configuration.Check(packagePath, fset, files, info); checkError != nil {
				test.Fatal(checkError)
			}
			for _, file := range files {
				for _, declaration := range file.Decls {
					functionName := "<package>"
					if function, ok := declaration.(*ast.FuncDecl); ok {
						functionName = function.Name.Name
					}
					ast.Inspect(declaration, func(node ast.Node) bool {
						selector, ok := node.(*ast.SelectorExpr)
						if !ok {
							return true
						}
						selection := info.Selections[selector]
						if selection == nil || selection.Kind() != types.FieldVal {
							return true
						}
						field, guarded := protected[selection.Obj()]
						if !guarded {
							return true
						}
						// A bare assignment's left side writes the field; compound assignments and
						// address-taking can read it and intentionally remain guarded.
						writes := false
						ast.Inspect(declaration, func(node ast.Node) bool {
							assignment, ok := node.(*ast.AssignStmt)
							if ok && assignment.Tok == token.ASSIGN {
								for _, left := range assignment.Lhs {
									if left == selector {
										writes = true
									}
								}
							}
							return true
						})
						if writes {
							return true
						}
						path, relativePathError := filepath.Rel(root, fset.Position(selector.Pos()).Filename)
						if relativePathError != nil {
							test.Fatal(relativePathError)
						}
						key := filepath.ToSlash(path) + ":" + functionName + ":" + field
						seen[key] = true
						if _, allowed := targetReaders[key]; !allowed {
							test.Errorf(
								"unapproved call-target read %s at %s; use CallTargets or ClosureTargets",
								key,
								fset.Position(selector.Pos()),
							)
						}
						return true
					})
				}
			}
		}
	}
	for reader, entry := range targetReaders {
		if entry.owner != "compiler" && entry.owner != "runtime" || entry.reason == "" {
			test.Errorf("invalid reader metadata: %s", reader)
		}
		if strings.HasPrefix(reader, "internal/"+selected+"/") && !seen[reader] {
			test.Errorf("stale call-target allowlist entry %s; remove it", reader)
		}
	}
}

// callTargetExports is the export data of the packages TestCallTargetReaders reads, their tests included, and of what
// they import directly, a product built ahead so the test runs no go. go/importer opens only those: the rest of -deps
// (tsc's export data, most of it) is what a direct import's signature reaches, carried in its own export data
// (#mtvmz1w).
func callTargetExports(test *testing.T, selected string) string {
	test.Helper()
	return buildcache.GoDirectExports(test, "./internal/"+selected, "./internal/leakcheck")
}

// Not parallel: prepare this export group before parallel reader checks on four CPUs.
func TestCallTargetExportInputsNative(test *testing.T) {
	callTargetExports(test, "native")
}

// Not parallel: prepare this export group serially before parallel reader checks.
func TestCallTargetExportInputsLower(test *testing.T) {
	callTargetExports(test, "lower")
}

// Not parallel: prepare this export group serially before parallel reader checks.
func TestCallTargetExportInputsFresh(test *testing.T) {
	callTargetExports(test, "fresh")
}

// Not parallel: prepare this export group serially before parallel reader checks.
func TestCallTargetExportInputsFlow(test *testing.T) {
	callTargetExports(test, "flow")
}

func TestProduct_CallTargetReaderArchivesNative(test *testing.T) {
	test.Parallel()
	callTargetExports(test, "native")
}

func TestProduct_CallTargetReaderArchivesLower(test *testing.T) {
	test.Parallel()
	callTargetExports(test, "lower")
}

func TestProduct_CallTargetReaderArchivesFresh(test *testing.T) {
	test.Parallel()
	callTargetExports(test, "fresh")
}

func TestProduct_CallTargetReaderArchivesFlow(test *testing.T) {
	test.Parallel()
	callTargetExports(test, "flow")
}

func TestCallTargetReadersNative(test *testing.T) {
	test.Parallel()
	checkCallTargetReaders(test, "native")
}

func TestCallTargetReadersLower(test *testing.T) {
	test.Parallel()
	checkCallTargetReaders(test, "lower")
}

func TestCallTargetReadersFresh(test *testing.T) {
	test.Parallel()
	checkCallTargetReaders(test, "fresh")
}

func TestCallTargetReadersFlow(test *testing.T) {
	test.Parallel()
	checkCallTargetReaders(test, "flow")
}
