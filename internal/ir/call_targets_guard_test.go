package ir_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/toolchain"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

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

// callTargetExports is the export data of the packages TestCallTargetReaders reads and everything they import, their
// tests included, a product built ahead so the test runs no go.
func callTargetExports(t *testing.T, selected string) string {
	t.Helper()
	pkg := "./internal/" + selected
	inputs := callTargetExportInputs(t, selected)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()
		goTool, err := toolchain.Path("go")
		if err != nil {
			return err
		}
		command := exec.CommandContext(ctx, goTool, "list", "-trimpath", "-buildvcs=false", "-export", "-deps", "-test", "-json", pkg)
		command.Dir = root
		command.Env = toolchain.Env()
		output, err := command.Output()
		if err != nil {
			return fmt.Errorf("export listing: %w", err)
		}
		decoder := json.NewDecoder(bytes.NewReader(output))
		var packages []buildcache.ListedPackage
		if err := os.Mkdir(filepath.Join(directory, "exports"), 0755); err != nil {
			return err
		}
		for {
			var pkg buildcache.ListedPackage
			if err := decoder.Decode(&pkg); err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			if relative, err := filepath.Rel(root, pkg.Dir); err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				pkg.Dir = filepath.ToSlash(relative)
			} else {
				pkg.Dir = ""
			}
			for _, names := range []*[]string{&pkg.GoFiles, &pkg.TestGoFiles, &pkg.XTestGoFiles} {
				kept := (*names)[:0]
				for _, name := range *names {
					if !filepath.IsAbs(name) {
						kept = append(kept, name)
					}
				}
				*names = kept
			}
			if pkg.Export != "" {
				data, err := callTargetExportMember(pkg.Export)
				if err != nil {
					return err
				}
				pkg.Export = fmt.Sprintf("exports/%d.a", len(packages))
				if err := os.WriteFile(filepath.Join(directory, pkg.Export), data, 0644); err != nil {
					return err
				}
			}
			packages = append(packages, pkg)
		}
		encoded, err := json.Marshal(packages)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "listing.json"), encoded, 0644)
	})
}

func callTargetExportInputs(t *testing.T, selected string) buildcache.Inputs {
	t.Helper()
	pkg := "./internal/" + selected
	inputs, err := buildcache.GoTestInputs("", "listing.json", pkg, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = "call-target-export-members " + selected
	inputs.Files = append(inputs.Files, "internal/ir/call_targets_guard_test.go")
	inputs.Flags = append(inputs.Flags, "go list -trimpath -buildvcs=false -export -deps -test -json; retain __.PKGDEF")
	return inputs
}

// go/importer reads __.PKGDEF, not executable object members. Keep that member
// in its original ar envelope, so imports use the same compiler export bytes
// without copying every dependency's machine code into each reader product.
func callTargetExportMember(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	header := make([]byte, 68)
	if _, err := io.ReadFull(file, header); err != nil {
		return nil, err
	}
	if string(header[:8]) != "!<arch>\n" || strings.TrimSpace(string(header[8:24])) != "__.PKGDEF" || string(header[66:68]) != "`\n" {
		return nil, fmt.Errorf("unexpected Go export archive %s", path)
	}
	size, err := strconv.Atoi(strings.TrimSpace(string(header[56:66])))
	if err != nil || size < 0 {
		return nil, fmt.Errorf("invalid export member size in %s", path)
	}
	data := make([]byte, len(header)+size+size%2)
	copy(data, header)
	if _, err := io.ReadFull(file, data[len(header):len(header)+size]); err != nil {
		return nil, err
	}
	if size%2 != 0 {
		data[len(data)-1] = '\n'
	}
	return data, nil
}

// Not parallel: discover each key before the archive cases start, avoiding
// four simultaneous dependency listings on one four-CPU worker. Archive cases
// still prepare themselves when selected alone; main supplies their input index.
func TestCallTargetExportInputsNative(t *testing.T) {
	callTargetExportInputs(t, "native")
}

func TestCallTargetExportInputsLower(t *testing.T) {
	callTargetExportInputs(t, "lower")
}

func TestCallTargetExportInputsFresh(t *testing.T) {
	callTargetExportInputs(t, "fresh")
}

func TestCallTargetExportInputsFlow(t *testing.T) {
	callTargetExportInputs(t, "flow")
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
