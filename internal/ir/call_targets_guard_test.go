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
	"internal/native/region.go:environmentEscapes:CallClosure.Closure":                                {"runtime", "follows carrier lifetime during evaluation; target identity comes from ClosureTargets"},
	"internal/native/arguments_length.go:spreadArguments:Call.Function":                               {"runtime", "names the spread call signature"},
	"internal/flow/build.go:CanThrow:ArraySort.Comparator":                                            {"compiler", "MayThrow and library failure propagation, routed in step 2"},
	"internal/lower/exceptions.go:libraryFailure:ArraySort.Comparator":                                {"compiler", "MayThrow and library failure propagation, routed in step 2"},
	"internal/lower/exceptions.go:throwsOut:ArraySort.Comparator":                                     {"compiler", "MayThrow and library failure propagation, routed in step 2"},
	"internal/lower/class_static_guard_test.go:TestClassStaticInitializerCallIsEmitted:Call.Function": {"compiler", "names the function under test"},
	"internal/native/class_inheritance.go:callCode:Call.Function":                                     {"runtime", "names the function to call"},
	"internal/native/emit_arrays.go:comparator:ArraySort.Comparator":                                  {"runtime", "names the function to call"},
	"internal/native/emit_expressions.go:evaluate:Call.Function":                                      {"runtime", "names the function to call"},
	"internal/native/emit_expressions.go:evaluate:CallClosure.Closure":                                {"runtime", "names the function to call"},
	"internal/native/emit_functions.go:arguments:Call.Function":                                       {"runtime", "names the function to call"},
	"internal/native/emit_functions.go:callThrough:CallClosure.Closure":                               {"runtime", "names the function to call"},
	"internal/native/loop_borrow_test.go:TestGlobalArgumentLending:Call.Function":                     {"runtime", "names the function under test"},
	"internal/native/loop_borrow_test.go:TestLoopCallCoverage:Call.Function":                          {"runtime", "names the function under test"},
}

func checkCallTargetReaders(t *testing.T, selected string) {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	listingProduct := callTargetExports(t, selected)
	encoded, err := os.ReadFile(filepath.Join(listingProduct, "listing.json"))
	if err != nil {
		t.Fatal(err)
	}
	var packages []buildcache.ListedPackage
	if err := json.Unmarshal(encoded, &packages); err != nil {
		t.Fatal(err)
	}
	exports := map[string]string{}
	for index, pkg := range packages {
		if pkg.Export != "" {
			pkg.Export = filepath.Join(listingProduct, pkg.Export)
		}
		if pkg.Dir != "" {
			pkg.Dir = filepath.Join(root, pkg.Dir)
		}
		packages[index] = pkg
		exports[pkg.ImportPath] = pkg.Export
		// External tests may use declarations from export_test.go. Import their
		// package's test variant rather than the production-only export file.
		if path, _, variant := strings.Cut(pkg.ImportPath, " ["); variant && (path == pkg.ForTest || exports[path] == "") {
			exports[path] = pkg.Export
		}
	}
	fset := token.NewFileSet()
	imports := importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) { return os.Open(exports[path]) })
	irPackage, err := imports.Import("github.com/system-inc/adamic/internal/ir")
	if err != nil {
		t.Fatal(err)
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
	for _, pkg := range packages {
		if strings.Contains(pkg.ImportPath, "[") || !strings.HasPrefix(pkg.ImportPath, "github.com/system-inc/adamic/internal/") {
			continue
		}
		base := filepath.Base(pkg.Dir)
		if base != selected {
			continue
		}
		if pkg.ImportPath != "github.com/system-inc/adamic/internal/"+base || (base != "native" && base != "lower" && base != "fresh" && base != "flow") {
			continue
		}
		for group, names := range [][]string{append(pkg.GoFiles, pkg.TestGoFiles...), pkg.XTestGoFiles} {
			if len(names) == 0 {
				continue
			}
			packagePath := pkg.ImportPath
			if group == 1 {
				packagePath += "_test"
			}
			files := []*ast.File{}
			for _, name := range names {
				file, err := parser.ParseFile(fset, filepath.Join(pkg.Dir, name), nil, 0)
				if err != nil {
					t.Fatal(err)
				}
				files = append(files, file)
			}
			info := &types.Info{Selections: map[*ast.SelectorExpr]*types.Selection{}}
			config := types.Config{Importer: imports}
			if _, err := config.Check(packagePath, fset, files, info); err != nil {
				t.Fatal(err)
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
						path, err := filepath.Rel(root, fset.Position(selector.Pos()).Filename)
						if err != nil {
							t.Fatal(err)
						}
						key := filepath.ToSlash(path) + ":" + functionName + ":" + field
						seen[key] = true
						if _, allowed := targetReaders[key]; !allowed {
							t.Errorf("unapproved call-target read %s at %s; use CallTargets or ClosureTargets", key, fset.Position(selector.Pos()))
						}
						return true
					})
				}
			}
		}
	}
	for reader, entry := range targetReaders {
		if entry.owner != "compiler" && entry.owner != "runtime" || entry.reason == "" {
			t.Errorf("invalid reader metadata: %s", reader)
		}
		if strings.HasPrefix(reader, "internal/"+selected+"/") && !seen[reader] {
			t.Errorf("stale call-target allowlist entry %s; remove it", reader)
		}
	}
}

// callTargetExports is the export data of the packages TestCallTargetReaders reads, their tests included, and of what
// they import directly, a product built ahead so the test runs no go. go/importer opens only those: the rest of -deps
// (tsc's export data, most of it) is what a direct import's signature reaches, carried in its own export data (#mtvmz1w).
func callTargetExports(t *testing.T, selected string) string {
	t.Helper()
	return buildcache.GoDirectExports(t, "./internal/"+selected, "./internal/leakcheck")
}

// Not parallel: prepare this export group before parallel reader checks on four CPUs.
func TestCallTargetExportInputsNative(t *testing.T) {
	callTargetExports(t, "native")
}

// Not parallel: prepare this export group serially before parallel reader checks.
func TestCallTargetExportInputsLower(t *testing.T) {
	callTargetExports(t, "lower")
}

// Not parallel: prepare this export group serially before parallel reader checks.
func TestCallTargetExportInputsFresh(t *testing.T) {
	callTargetExports(t, "fresh")
}

// Not parallel: prepare this export group serially before parallel reader checks.
func TestCallTargetExportInputsFlow(t *testing.T) {
	callTargetExports(t, "flow")
}

func TestProduct_CallTargetReaderArchivesNative(t *testing.T) {
	t.Parallel()
	callTargetExports(t, "native")
}

func TestProduct_CallTargetReaderArchivesLower(t *testing.T) {
	t.Parallel()
	callTargetExports(t, "lower")
}

func TestProduct_CallTargetReaderArchivesFresh(t *testing.T) {
	t.Parallel()
	callTargetExports(t, "fresh")
}

func TestProduct_CallTargetReaderArchivesFlow(t *testing.T) {
	t.Parallel()
	callTargetExports(t, "flow")
}

func TestCallTargetReadersNative(t *testing.T) {
	t.Parallel()
	checkCallTargetReaders(t, "native")
}

func TestCallTargetReadersLower(t *testing.T) {
	t.Parallel()
	checkCallTargetReaders(t, "lower")
}

func TestCallTargetReadersFresh(t *testing.T) {
	t.Parallel()
	checkCallTargetReaders(t, "fresh")
}

func TestCallTargetReadersFlow(t *testing.T) {
	t.Parallel()
	checkCallTargetReaders(t, "flow")
}
