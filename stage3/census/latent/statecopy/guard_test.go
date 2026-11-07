package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLatentStatecopySortedMapProgramMembership(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	for name, source := range map[string]string{
		"internal/ir/map.go":               "package ir; type SortedMap[K comparable,V any] map[K]V",
		"internal/lower/lower.go":          "package lower; type lowering struct {plan *programMembership}",
		"internal/lower/program_region.go": "package lower; type programMembership struct {types ir.SortedMap[int, []bool]}",
	} {
		path := filepath.Join(root, name)
		if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0700); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	defer func() {
		if failure := recover(); failure != nil {
			test.Fatalf("program_region.go: programMembership.types: %v", failure)
		}
	}()
	result := string(generate(root, filepath.Join(root, "internal/lower/lower.go")))
	if !strings.Contains(result, "make(ir.SortedMap[int, []bool]") || !strings.Contains(result, "make([]bool") {
		test.Fatal(result)
	}
}

// The overlay adds only maps of checker-owned nodes and bools to lowering.
// All production state follows the same root and recursive struct decisions as generation.
func TestLatentStatecopyCoversEveryLoweringStateField(test *testing.T) {
	test.Parallel()
	// Reads: internal/lower/*.go
	// Reads: internal/ir/*.go
	root, absolutePathError := filepath.Abs("../../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	g := loadGenerator(root, filepath.Join(root, "internal/lower/lower.go"))
	done := map[string]bool{}
	for {
		progressed := false
		for name := range g.needed {
			if done[name] {
				continue
			}
			done[name] = true
			progressed = true
			structure := g.types[name].(*ast.StructType)
			for _, field := range structure.Fields.List {
				for _, id := range field.Names {
					if name == "lowering" &&
						id.Name == "impliedTargetQuery" &&
						printed(field.Type) == "func(*ast.Node) *checker.Type" {
						continue
					}
					if name == "lowering" &&
						id.Name == "sourceIdentities" &&
						printed(field.Type) == "map[*ast.Node]ir.SourceIdentity" {
						continue
					}
					func() {
						defer func() {
							if failure := recover(); failure != nil {
								test.Errorf("%s.%s (%s): %v", name, id.Name, fieldSource(root, name), failure)
							}
						}()
						g.copy(field.Type, "value."+id.Name)
					}()
				}
			}
		}
		if !progressed {
			break
		}
	}
}

func fieldSource(root, name string) string {
	paths, _ := filepath.Glob(filepath.Join(root, "internal/lower/*.go"))
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		positions := token.NewFileSet()
		file, parseFileError := parser.ParseFile(positions, path, nil, 0)
		if parseFileError != nil {
			continue
		}
		for _, decl := range file.Decls {
			if group, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range group.Specs {
					if spec, ok := spec.(*ast.TypeSpec); ok && spec.Name.Name == name {
						return positions.Position(spec.Pos()).String()
					}
				}
			}
		}
	}
	return name
}

// The shared identity map is passed through, never deep-copied: a lowering's copy reads and fills the one map every
// lowering of a latent run shares (#pv935yw). Mutant: the generator's sourceIdentities case dropped, so the copy gets
// result.sourceIdentities from a fresh map.
func TestLatentStatecopySharesSourceIdentities(test *testing.T) {
	test.Parallel()
	// Reads: internal/lower/*.go
	// Reads: internal/ir/*.go
	root, absolutePathError := filepath.Abs("../../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	result := string(generate(root, filepath.Join(root, "internal/lower/lower.go")))
	if !strings.Contains(result, "func latentCopy_lowering(") {
		test.Fatal("no lowering copier was generated")
	}
	if strings.Contains(result, "result.sourceIdentities=") || strings.Contains(result, "result.sourceIdentities =") {
		test.Fatal("the lowering copier deep-copies sourceIdentities, which every lowering of a latent run shares")
	}
}

// An unexported field added to an exported IR struct is refused by name when the copier is generated, so the
// census refuses once, at its build, instead of panicking in every unit that copies the struct (#fwbnwn3: on
// Oct 20 Program.methodTables did that to 9,818 of the tsc entry's 10,705 full-mode units). Mutant: the
// refusal in loadGenerator dropped, so this program generates and the field goes uncopied.
func TestLatentStatecopyRefusesAnUnnamedPrivateIRField(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	for name, source := range map[string]string{
		"internal/ir/ir.go":       "package ir; type Program struct {Functions []int; methodTables *int; addedLater map[string]int}",
		"internal/lower/lower.go": "package lower; type lowering struct {result *ir.Program}",
	} {
		path := filepath.Join(root, name)
		if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0700); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	defer func() {
		failure := recover()
		if failure == nil {
			test.Fatal("an unexported IR field missing from irPrivateFields was not refused")
		}
		if text, ok := failure.(string); !ok ||
			!strings.Contains(text, "Program.addedLater") ||
			!strings.Contains(text, "ir.go:1") {
			test.Fatalf("the refusal must name the field and where it's declared: %v", failure)
		}
	}()
	generate(root, filepath.Join(root, "internal/lower/lower.go"))
}

// irPrivateFields names exactly the unexported fields of exported IR structs on this tree: none missing, which
// loadGenerator refuses, and none stale, so the table can't keep a policy for a field the IR dropped. Mutant: an
// entry for a field internal/ir doesn't declare.
func TestLatentStatecopyNamesEveryPrivateIRField(test *testing.T) {
	test.Parallel()
	// Reads: internal/lower/*.go
	// Reads: internal/ir/*.go
	root, absolutePathError := filepath.Abs("../../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	g := loadGenerator(root, filepath.Join(root, "internal/lower/lower.go"))
	declared := map[string]bool{}
	for _, key := range g.privateIRFields {
		declared[key] = true
	}
	for key := range irPrivateFields {
		if !declared[key] {
			test.Errorf("irPrivateFields names %s, which internal/ir no longer declares", key)
		}
	}
	generated := string(generate(root, filepath.Join(root, "internal/lower/lower.go")))
	for _, want := range []string{
		"func latentPrivateIRField(",
		"func latentSharedIRPointer(",
		"reflect.TypeFor[*ir.GeneratedType]()",
		`field == "methodTables"`,
	} {
		if !strings.Contains(generated, want) {
			test.Errorf("the generated copier lacks %s", want)
		}
	}
}

// The independent audited contract prevents accidentally dropping a carried memo
// from the generator's table. Unrelated lowering fields are intentionally ignored.
func carryFactGuardGaps(root, lowerInput string) []string {
	audited := map[string]string{
		"entriesRecords":             "map[*ast.Node]bool",
		"entriesProgramRecords":      "bool",
		"entriesProgramRecordsKnown": "bool",
	}
	g := loadGenerator(root, lowerInput)
	declared := map[string]string{}
	for _, field := range g.types["lowering"].(*ast.StructType).Fields.List {
		for _, id := range field.Names {
			declared[id.Name] = printed(field.Type)
		}
	}
	gaps := []string{}
	for name, want := range audited {
		if policy, ok := loweringFieldPolicies[name]; !ok || policy != programFact {
			gaps = append(gaps, "carried lowering."+name+" is missing from the programFact table")
		}
		if got, ok := declared[name]; !ok {
			gaps = append(gaps, "carried lowering."+name+" no longer exists")
		} else if got != want {
			gaps = append(gaps, "carried lowering."+name+" type changed: got "+got+", audited "+want)
		}
	}
	for name := range loweringFieldPolicies {
		if _, ok := audited[name]; !ok {
			gaps = append(gaps, "unaudited carried lowering."+name)
		}
	}
	// A type mutant must report its field, rather than panic in generation.
	if len(gaps) > 0 {
		return gaps
	}
	generated := string(generate(root, lowerInput))
	for name, ty := range audited {
		if ty == "map[*ast.Node]bool" &&
			!strings.Contains(generated, "result."+name+" = func(value map[*ast.Node]bool)") {
			gaps = append(gaps, "carried lowering."+name+" lost its owned snapshot map copy")
		}
		for _, assignment := range []string{name + ": value." + name, "value." + name + " = facts." + name} {
			if !strings.Contains(generated, assignment) {
				gaps = append(gaps, "carried lowering."+name+" is dropped from capture/install: "+assignment)
			}
		}
	}
	// Fresh first-error units do not take a unit snapshot. Their shared helper
	// must own the installed map too, not merely rely on the generated copier.
	helper, readError := os.ReadFile(filepath.Join(root, "stage3/census/latent/full.go.txt"))
	if readError != nil ||
		!strings.Contains(string(helper), "local.entriesRecords = maps.Clone(local.entriesRecords)") {
		gaps = append(gaps, "carried lowering.entriesRecords lost its owned installation map copy")
	}
	return gaps
}

func TestLatentCarryFactsAuditedFields(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs("../../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for _, gap := range carryFactGuardGaps(root, filepath.Join(root, "internal/lower/lower.go")) {
		test.Error(gap)
	}
}

func TestLatentCarryFactsIgnoresUncarriedFields(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs("../../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	input, readError := os.ReadFile(filepath.Join(root, "internal/lower/lower.go"))
	if readError != nil {
		test.Fatal(readError)
	}
	path := filepath.Join(test.TempDir(), "lower.go")
	source := strings.Replace(string(input), "type lowering struct {", "type lowering struct {\n futureMemo bool", 1)
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	if gaps := carryFactGuardGaps(root, path); len(gaps) > 0 {
		test.Fatalf("an unrelated lowering field must retain the default unit snapshot: %v", gaps)
	}
}

// Type-check the actual copier and generated policies from source, then prove
// the private-field copy decisions from their ASTs. This test builds nothing.
// LATENT_STATECOPY_MUTANT selects an old full.go.txt saved under a mutant filename.
var latentStatecopyMutant = os.Getenv("LATENT_STATECOPY_MUTANT")

// Trimmed test binaries do not embed GOROOT. The standard source importer uses
// build.Default and offers no custom Context. Resolve its read-only source root
// once during package initialization, before any tests run in parallel.
var latentProbeSourceRootError = func() error {
	if build.Default.GOROOT != "" {
		return nil
	}
	for _, directory := range filepath.SplitList(os.Getenv("PATH")) {
		executable, resolveSymlinksError := filepath.EvalSymlinks(filepath.Join(directory, "go"))
		if resolveSymlinksError != nil {
			continue
		}
		root := filepath.Dir(filepath.Dir(executable))
		if info, statError := os.Stat(
			filepath.Join(root, "src", "reflect", "type.go"),
		); statError == nil && !info.IsDir() {
			build.Default.GOROOT = root
			return nil
		}
	}
	return fmt.Errorf("source importer needs an installed Go source tree on PATH or GOROOT")
}()

func TestLatentSpeculativePrivateIRFields(test *testing.T) {
	test.Parallel()
	// Reads: internal/lower/*.go
	// Reads: internal/ir/*.go
	if latentProbeSourceRootError != nil {
		test.Fatal(latentProbeSourceRootError)
	}
	root, absolutePathError := filepath.Abs("../../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	generated := generate(root, filepath.Join(root, "internal/lower/lower.go"))
	full := filepath.Join(root, "stage3/census/latent/full.go.txt")
	if latentStatecopyMutant != "" {
		full = latentStatecopyMutant
	}
	extract := func(path string, source any, names map[string]bool) string {
		fs := token.NewFileSet()
		file, parseFileError := parser.ParseFile(fs, path, source, 0)
		if parseFileError != nil {
			test.Fatal(parseFileError)
		}
		result := ""
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && names[fn.Name.Name] {
				result += printed(fn) + "\n"
			}
			if group, ok := decl.(*ast.GenDecl); ok && group.Tok == token.VAR {
				for _, spec := range group.Specs {
					if v, ok := spec.(*ast.ValueSpec); ok && names[v.Names[0].Name] {
						result += printed(group) + "\n"
					}
				}
			}
		}
		return result
	}
	source := "package probe\nimport (\"reflect\"; \"os\"; \"github.com/system-inc/adamic/internal/ir\")\n"
	source += extract("generated.go", generated, map[string]bool{
		"latentPrivateIRField":      true,
		"latentSharedIRPointer":     true,
		"latentSpecCopyFunction":    true,
		"latentSpecIRMutableFields": true,
	})
	source += extract(full, nil, map[string]bool{"latentCopyIR": true, "latentEmptyArgumentCache": true})
	source += extract(filepath.Join(root, "stage3/census/latent/speculative.go.txt"), nil, map[string]bool{
		"latentSpeculativeEnabled":    true,
		"latentSpecImmutable":         true,
		"latentSpecImmutableTypes":    true,
		"latentSpecMutableFields":     true,
		"latentSpecMutableFieldCache": true,
	})

	// Check the static ownership assertion before type checking so the historical
	// mutant is rejected for sharing private fields, even if its old helpers drift.
	fullAST, parseFileError := parser.ParseFile(token.NewFileSet(), full, nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	generatedAST, parseFileError := parser.ParseFile(token.NewFileSet(), "generated.go", generated, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	if latentPrivateCopyProofError := latentPrivateCopyProof(
		fullAST,
		generatedAST,
	); latentPrivateCopyProofError != nil {
		test.Fatal(latentPrivateCopyProofError)
	}
	fs := token.NewFileSet()
	probe, parseFileError := parser.ParseFile(fs, "probe.go", source, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	configuration := types.Config{
		Importer: latentProbeSourceImporter{Importer: importer.ForCompiler(fs, "source", nil), root: root},
	}
	if _, checkError := configuration.Check(
		"github.com/system-inc/adamic/latentprobe",
		fs,
		[]*ast.File{probe},
		nil,
	); checkError != nil {
		test.Fatalf("assembled speculative copier does not type-check against real IR source: %v", checkError)
	}
}

func TestLatentSpeculativeRefusesUnnamedExportedIRField(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	path := filepath.Join(root, "internal/ir/ir.go")
	if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0700); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if writeError := os.WriteFile(
		path,
		[]byte("package ir; type Function struct { AddedLater []int }"),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	defer func() {
		failure := recover()
		if failure == nil ||
			!strings.Contains(fmt.Sprint(failure), "Function.AddedLater") ||
			!strings.Contains(fmt.Sprint(failure), "ir.go:1") {
			test.Fatalf("build refusal must name exported field and declaration: %v", failure)
		}
	}()
	generate(root, filepath.Join(root, "internal/lower/lower.go"))
}

func TestLatentSpeculativeBoundaryPolicyRequiredAtBuild(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs("../../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	g := loadGenerator(root, filepath.Join(root, "internal/lower/lower.go"))
	policies := map[string]irFieldPolicy{}
	for key, value := range irSpecFields {
		policies[key] = value
	}
	delete(policies, "Function.Boundary")
	defer func() {
		failure := recover()
		if failure == nil ||
			!strings.Contains(fmt.Sprint(failure), "Function.Boundary") ||
			!strings.Contains(fmt.Sprint(failure), "isn't named") {
			test.Fatalf("missing Boundary must fail at generation, naming it: %v", failure)
		}
	}()
	g.checkIRFields(token.NewFileSet(), policies)
}

func TestLatentSpeculativeEveryIRFieldPolicyIsCurrent(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs("../../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	g := loadGenerator(root, filepath.Join(root, "internal/lower/lower.go"))
	declared := map[string]bool{}
	for name, expression := range g.irTypes {
		if structure, ok := expression.(*ast.StructType); ok && ast.IsExported(name) {
			for _, field := range structure.Fields.List {
				for _, id := range field.Names {
					if id.IsExported() {
						declared[name+"."+id.Name] = true
					}
				}
			}
		}
	}
	for key := range irSpecFields {
		if !declared[key] {
			test.Errorf("stale speculative field policy %s", key)
		}
	}
	generated := string(generate(root, filepath.Join(root, "internal/lower/lower.go")))
	for _, want := range []string{
		"result.Boundary = latentCopyIR(value.Boundary, seen)",
		"result.Source = latentCopyIR(value.Source, seen)",
	} {
		if !strings.Contains(generated, want) {
			test.Errorf("generated function copier lacks %s", want)
		}
	}
}

// The source importer reads IR at its real location without a temporary module,
// changing the process directory or running a build. Its source root is frozen at
// package initialization; parallel tests only read that configuration.
type latentProbeSourceImporter struct {
	types.Importer
	root string
}

func (sourceImporter latentProbeSourceImporter) Import(name string) (*types.Package, error) {
	if name == "github.com/system-inc/adamic/internal/ir" {
		return sourceImporter.Importer.(types.ImporterFrom).ImportFrom("./internal/ir", sourceImporter.root, 0)
	}
	return sourceImporter.Importer.Import(name)
}

func latentProbeFunction(file *ast.File, name string) *ast.FuncDecl {
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == name {
			return function
		}
	}
	return nil
}

// Accept only the ownership-preserving control flow: a zero struct, and a
// dominating private-policy/continue guard before each indexed field write.
// This deliberately rejects new write shapes until their ownership is proved.
func latentPrivateCopyProof(full, generated *ast.File) error {
	fail := func(reason string) error { return fmt.Errorf("speculative private-field AST proof: %s", reason) }
	copyIR := latentProbeFunction(full, "latentCopyIR")
	if copyIR == nil {
		return fail("missing IR copier")
	}
	var structure, pointer *ast.CaseClause
	ast.Inspect(copyIR, func(node ast.Node) bool {
		clause, ok := node.(*ast.CaseClause)
		if ok && len(clause.List) == 1 {
			switch printed(clause.List[0]) {
			case "reflect.Struct":
				structure = clause
			case "reflect.Pointer":
				pointer = clause
			}
		}
		return true
	})
	if structure == nil ||
		len(structure.Body) != 4 ||
		printed(structure.Body[0]) != "result := reflect.New(v.Type()).Elem()" {
		return fail("struct copy must start empty, never from the source")
	}
	branch, ok := structure.Body[1].(*ast.IfStmt)
	if !ok ||
		printed(branch.Cond) != "speculative" ||
		branch.Init != nil ||
		branch.Else != nil ||
		len(branch.Body.List) != 3 {
		return fail("missing isolated speculative struct path")
	}
	first, ok := branch.Body.List[0].(*ast.ForStmt)
	if !ok ||
		printed(first.Init) != "index := 0" ||
		printed(first.Cond) != "index < v.NumField()" ||
		printed(first.Post) != "index++" {
		return fail("private policies must be consulted for every field")
	}
	second, ok := branch.Body.List[1].(*ast.RangeStmt)
	if !ok ||
		second.Tok != token.DEFINE ||
		printed(second.Key) != "_" ||
		printed(second.Value) != "index" ||
		printed(second.X) != "latentSpecMutableFields(v.Type())" {
		return fail("mutable fields must use the checked field inventory")
	}
	checkWrites := func(body *ast.BlockStmt, argument string) bool {
		if len(body.List) != 3 {
			return false
		}
		guard, ok := body.List[0].(*ast.IfStmt)
		if !ok ||
			guard.Init != nil ||
			guard.Else != nil ||
			printed(guard.Cond) != "latentPrivateIRField(v, index)" ||
			len(guard.Body.List) != 1 {
			return false
		}
		skip, ok := guard.Body.List[0].(*ast.BranchStmt)
		if !ok || skip.Tok != token.CONTINUE || skip.Label != nil {
			return false
		}
		// A private field cannot reach either assignment below that continue.
		writable, ok := body.List[1].(*ast.IfStmt)
		if !ok ||
			writable.Init != nil ||
			writable.Else != nil ||
			printed(writable.Cond) != "!result.Field(index).CanSet()" ||
			len(writable.Body.List) != 1 {
			return false
		}
		expression, ok := writable.Body.List[0].(*ast.ExprStmt)
		if !ok {
			return false
		}
		panicCall, ok := expression.X.(*ast.CallExpr)
		if !ok || printed(panicCall.Fun) != "panic" {
			return false
		}
		write, ok := body.List[2].(*ast.ExprStmt)
		if !ok {
			return false
		}
		call, ok := write.X.(*ast.CallExpr)
		return ok &&
			printed(call.Fun) == "result.Field(index).Set" &&
			len(call.Args) == 1 &&
			printed(call.Args[0]) == argument
	}
	if !checkWrites(first.Body, "v.Field(index)") ||
		!checkWrites(second.Body, "copyValue(v.Field(index))") ||
		printed(branch.Body.List[2]) != "return result" {
		return fail(
			"private fields must pass latentPrivateIRField before any source assignment; privateEmpty fields stay empty",
		)
	}
	// The struct case has no preceding source copy and returns its result. Shared
	// identities must return their original pointer before entering struct copying.
	if printed(structure.Body[3]) != "return result" || pointer == nil || len(pointer.Body) < 2 {
		return fail("missing copy result or pointer policy")
	}
	shared, ok := pointer.Body[1].(*ast.IfStmt)
	if !ok ||
		shared.Init != nil ||
		shared.Else != nil ||
		printed(shared.Cond) != "latentSharedIRPointer(v.Type())" ||
		len(shared.Body.List) != 1 ||
		printed(shared.Body.List[0]) != "return v" {
		return fail("shared private identities must use latentSharedIRPointer in speculative mode too")
	}
	if irPrivateFields["Program.methodTables"] != privateEmpty {
		return fail("methodTables must be a privateEmpty cache")
	}
	privateHelper := latentProbeFunction(generated, "latentPrivateIRField")
	if privateHelper == nil ||
		len(privateHelper.Body.List) == 0 ||
		printed(privateHelper.Body.List[0]) != "field := v.Type().Field(index).Name" {
		return fail("private policies must inspect the actual source field name")
	}
	for key, policy := range irPrivateFields {
		owner, field, _ := strings.Cut(key, ".")
		helper := latentProbeFunction(generated, "latentPrivateIRField")
		condition := fmt.Sprintf("v.Type() == reflect.TypeFor[ir.%s]() && field == %q", owner, field)
		if policy == privateShared {
			helper = latentProbeFunction(generated, "latentSharedIRPointer")
			condition = fmt.Sprintf("t == reflect.TypeFor[*ir.%s]()", owner)
		}
		if helper == nil {
			return fail("missing generated policy for " + key)
		}
		found := false
		for _, statement := range helper.Body.List {
			guard, ok := statement.(*ast.IfStmt)
			if !ok || printed(guard.Cond) != condition {
				continue
			}
			body := guard.Body.List
			if guard.Init == nil && guard.Else == nil && len(body) > 0 && printed(body[len(body)-1]) == "return true" {
				// An empty cache policy does exactly one thing: skip its zero-initialized
				// destination field. It must neither assign nor read the private source.
				found = policy != privateEmpty || len(body) == 1
			}
		}
		if !found {
			return fail("generated private policy does not preserve ownership for " + key)
		}
	}
	return nil
}
