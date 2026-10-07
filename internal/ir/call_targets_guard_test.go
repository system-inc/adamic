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
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// x/tools is not a dependency here. go list supplies the module-aware export
// files, and go/types resolves selectors, including aliases and embedded fields.
// Entries name a file, enclosing Go function and IR field, so a migrated reader
// cannot leave an unrelated reader in that file exempt. Owners identify the unit
// responsible for routing the remaining analyses. Emission entries are permanent
// while they emit the call; even those must still contain their stated read.
type targetReader struct{ owner, reason string }

var targetReaders = map[string]targetReader{
	"internal/flow/build.go:CanThrow:ArraySort.Comparator":                               {"compiler", "MayThrow and library failure propagation, routed in step 2"},
	"internal/flow/infer.go:value:CallClosure.Closure":                                   {"runtime", "evaluates callee operands and applies analysis summaries"},
	"internal/fresh/fresh.go:directlyCalled:ArraySort.Comparator":                        {"runtime", "builds comparator call graph"},
	"internal/fresh/fresh.go:value:Call.Function":                                        {"runtime", "evaluates callee operands and applies analysis summaries"},
	"internal/fresh/fresh.go:value:CallClosure.Closure":                                  {"runtime", "evaluates callee operands and applies analysis summaries"},
	"internal/lower/exceptions.go:libraryFailure:ArraySort.Comparator":                   {"compiler", "MayThrow and library failure propagation, routed in step 2"},
	"internal/lower/exceptions.go:throwsOut:ArraySort.Comparator":                        {"compiler", "MayThrow and library failure propagation, routed in step 2"},
	"internal/native/class_inheritance.go:callCode:Call.Function":                        {"runtime", "emits the call"},
	"internal/native/class_inheritance_test.go:TestInheritanceMemoryPlans:Call.Function": {"runtime", "asserts memory plans for virtual calls"},
	"internal/native/element_borrow.go:unchanging:Call.Function":                         {"runtime", "checks callee mutations before element borrowing"},
	"internal/native/emit_arrays.go:comparator:ArraySort.Comparator":                     {"runtime", "emits the call"},
	"internal/native/emit_expressions.go:evaluate:ArraySort.Comparator":                  {"runtime", "emits the call"},
	"internal/native/emit_expressions.go:evaluate:Call.Function":                         {"runtime", "emits the call"},
	"internal/native/emit_expressions.go:evaluate:CallClosure.Closure":                   {"runtime", "emits the call"},
	"internal/native/emit_functions.go:arguments:Call.Function":                          {"runtime", "emits the call"},
	"internal/native/emit_functions.go:callThrough:CallClosure.Closure":                  {"runtime", "emits the call"},
	"internal/native/region.go:feedsRegion:Call.Function":                                {"runtime", "checks call results feeding regions"},
	"internal/native/region.go:freshValue:Call.Function":                                 {"runtime", "checks allocation region eligibility"},
	"internal/native/region.go:handRegion:Call.Function":                                 {"runtime", "selects callee region variant"},
	"internal/native/region.go:regionFor:Call.Function":                                  {"runtime", "selects allocation region for callee"},
	"internal/native/reuse.go:planReuse:ArraySort.Comparator":                            {"runtime", "plans consumed parameters and argument moves"},
	"internal/native/reuse.go:planReuse:Call.Function":                                   {"runtime", "plans consumed parameters and argument moves"},
}

func TestCallTargetReaders(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "list", "-export", "-deps", "-test", "-json", "./internal/native", "./internal/lower", "./internal/fresh", "./internal/flow")
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		if failure, ok := err.(*exec.ExitError); ok {
			t.Fatalf("go list: %s", failure.Stderr)
		}
		t.Fatal(err)
	}
	type listedPackage struct {
		ImportPath, Dir, Export string
		GoFiles, TestGoFiles    []string
	}
	packages := []listedPackage{}
	exports := map[string]string{}
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for {
		var pkg listedPackage
		err := decoder.Decode(&pkg)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		packages = append(packages, pkg)
		exports[pkg.ImportPath] = pkg.Export
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
		if pkg.ImportPath != "github.com/system-inc/adamic/internal/"+base || (base != "native" && base != "lower" && base != "fresh" && base != "flow") {
			continue
		}
		files := []*ast.File{}
		for _, name := range append(pkg.GoFiles, pkg.TestGoFiles...) {
			file, err := parser.ParseFile(fset, filepath.Join(pkg.Dir, name), nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			files = append(files, file)
		}
		info := &types.Info{Selections: map[*ast.SelectorExpr]*types.Selection{}}
		config := types.Config{Importer: imports}
		if _, err := config.Check(pkg.ImportPath, fset, files, info); err != nil {
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
	for reader, entry := range targetReaders {
		if entry.owner != "compiler" && entry.owner != "runtime" || entry.reason == "" {
			t.Errorf("invalid reader metadata: %s", reader)
		}
		if !seen[reader] {
			t.Errorf("stale call-target allowlist entry %s; remove it", reader)
		}
	}
}
