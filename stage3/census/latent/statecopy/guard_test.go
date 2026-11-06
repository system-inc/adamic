package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLatentStatecopySortedMapProgramMembership(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for name, source := range map[string]string{
		"internal/ir/map.go":               "package ir; type SortedMap[K comparable,V any] map[K]V",
		"internal/lower/lower.go":          "package lower; type lowering struct {plan *programMembership}",
		"internal/lower/program_region.go": "package lower; type programMembership struct {types ir.SortedMap[int, []bool]}",
	} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	defer func() {
		if failure := recover(); failure != nil {
			t.Fatalf("program_region.go: programMembership.types: %v", failure)
		}
	}()
	result := string(generate(root, filepath.Join(root, "internal/lower/lower.go")))
	if !strings.Contains(result, "make(ir.SortedMap[int, []bool]") || !strings.Contains(result, "make([]bool") {
		t.Fatal(result)
	}
}

// The overlay adds only maps of checker-owned nodes and bools to lowering.
// All production state follows the same root and recursive struct decisions as generation.
func TestLatentStatecopyCoversEveryLoweringStateField(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/*.go
	// Reads: internal/ir/*.go
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
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
					if name == "lowering" && id.Name == "impliedTargetQuery" && printed(field.Type) == "func(*ast.Node) *checker.Type" {
						continue
					}
					if name == "lowering" && id.Name == "sourceIdentities" && printed(field.Type) == "map[*ast.Node]ir.SourceIdentity" {
						continue
					}
					func() {
						defer func() {
							if failure := recover(); failure != nil {
								t.Errorf("%s.%s (%s): %v", name, id.Name, fieldSource(root, name), failure)
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
		file, err := parser.ParseFile(positions, path, nil, 0)
		if err != nil {
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
func TestLatentStatecopySharesSourceIdentities(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/*.go
	// Reads: internal/ir/*.go
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	result := string(generate(root, filepath.Join(root, "internal/lower/lower.go")))
	if !strings.Contains(result, "func latentCopy_lowering(") {
		t.Fatal("no lowering copier was generated")
	}
	if strings.Contains(result, "result.sourceIdentities=") || strings.Contains(result, "result.sourceIdentities =") {
		t.Fatal("the lowering copier deep-copies sourceIdentities, which every lowering of a latent run shares")
	}
}

// An unexported field added to an exported IR struct is refused by name when the copier is generated, so the
// census refuses once, at its build, instead of panicking in every unit that copies the struct (#fwbnwn3: on
// Oct 20 Program.methodTables did that to 9,818 of the tsc entry's 10,705 full-mode units). Mutant: the
// refusal in loadGenerator dropped, so this program generates and the field goes uncopied.
func TestLatentStatecopyRefusesAnUnnamedPrivateIRField(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for name, source := range map[string]string{
		"internal/ir/ir.go":       "package ir; type Program struct {Functions []int; methodTables *int; addedLater map[string]int}",
		"internal/lower/lower.go": "package lower; type lowering struct {result *ir.Program}",
	} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	defer func() {
		failure := recover()
		if failure == nil {
			t.Fatal("an unexported IR field missing from irPrivateFields was not refused")
		}
		if text, ok := failure.(string); !ok || !strings.Contains(text, "Program.addedLater") || !strings.Contains(text, "ir.go:1") {
			t.Fatalf("the refusal must name the field and where it's declared: %v", failure)
		}
	}()
	generate(root, filepath.Join(root, "internal/lower/lower.go"))
}

// irPrivateFields names exactly the unexported fields of exported IR structs on this tree: none missing, which
// loadGenerator refuses, and none stale, so the table can't keep a policy for a field the IR dropped. Mutant: an
// entry for a field internal/ir doesn't declare.
func TestLatentStatecopyNamesEveryPrivateIRField(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/*.go
	// Reads: internal/ir/*.go
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	g := loadGenerator(root, filepath.Join(root, "internal/lower/lower.go"))
	declared := map[string]bool{}
	for _, key := range g.privateIRFields {
		declared[key] = true
	}
	for key := range irPrivateFields {
		if !declared[key] {
			t.Errorf("irPrivateFields names %s, which internal/ir no longer declares", key)
		}
	}
	generated := string(generate(root, filepath.Join(root, "internal/lower/lower.go")))
	for _, want := range []string{"func latentPrivateIRField(", "func latentSharedIRPointer(", "reflect.TypeFor[*ir.GeneratedType]()", `field == "methodTables"`} {
		if !strings.Contains(generated, want) {
			t.Errorf("the generated copier lacks %s", want)
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
		if ty == "map[*ast.Node]bool" && !strings.Contains(generated, "result."+name+" = func(value map[*ast.Node]bool)") {
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
	helper, err := os.ReadFile(filepath.Join(root, "stage3/census/latent/full.go.txt"))
	if err != nil || !strings.Contains(string(helper), "local.entriesRecords = maps.Clone(local.entriesRecords)") {
		gaps = append(gaps, "carried lowering.entriesRecords lost its owned installation map copy")
	}
	return gaps
}

func TestLatentCarryFactsAuditedFields(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, gap := range carryFactGuardGaps(root, filepath.Join(root, "internal/lower/lower.go")) {
		t.Error(gap)
	}
}

func TestLatentCarryFactsIgnoresUncarriedFields(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	input, err := os.ReadFile(filepath.Join(root, "internal/lower/lower.go"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "lower.go")
	source := strings.Replace(string(input), "type lowering struct {", "type lowering struct {\n futureMemo bool", 1)
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if gaps := carryFactGuardGaps(root, path); len(gaps) > 0 {
		t.Fatalf("an unrelated lowering field must retain the default unit snapshot: %v", gaps)
	}
}

// Compile the actual speculative copier and the generated private-field answers together.
// LATENT_STATECOPY_MUTANT selects an old full.go.txt saved under a mutant filename.
func TestLatentSpeculativePrivateIRFields(t *testing.T) {
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("build-refusal", TestLatentStatecopyRefusesAnUnnamedPrivateIRField)
	generated := generate(root, filepath.Join(root, "internal/lower/lower.go"))
	full := filepath.Join(root, "stage3/census/latent/full.go.txt")
	if mutant := os.Getenv("LATENT_STATECOPY_MUTANT"); mutant != "" {
		full = mutant
	}
	extract := func(path string, source any, names map[string]bool) string {
		fs := token.NewFileSet()
		file, err := parser.ParseFile(fs, path, source, 0)
		if err != nil {
			t.Fatal(err)
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
	source += extract("generated.go", generated, map[string]bool{"latentPrivateIRField": true, "latentSharedIRPointer": true, "latentSpecCopyFunction": true, "latentSpecIRMutableFields": true})
	source += extract(full, nil, map[string]bool{"latentCopyIR": true, "latentEmptyArgumentCache": true})
	source += extract(filepath.Join(root, "stage3/census/latent/speculative.go.txt"), nil, map[string]bool{"latentSpeculativeEnabled": true, "latentSpecImmutable": true, "latentSpecImmutableTypes": true, "latentSpecMutableFields": true, "latentSpecMutableFieldCache": true})
	folder := t.TempDir()
	files := map[string]string{
		"go.mod":  fmt.Sprintf("module github.com/system-inc/adamic/probe\ngo 1.27.0\nrequire github.com/system-inc/adamic v0.0.0\nreplace github.com/system-inc/adamic => %s\n", root),
		"copy.go": source,
		"copy_test.go": `package probe
import ("testing"; "reflect"; "github.com/system-inc/adamic/internal/ir")
func TestPrivateCopy(t *testing.T) {
 t.Setenv("LATENT_SPECULATIVE","1")
 p := &ir.Program{}
 p.MethodsNamed("witness")
 if reflect.ValueOf(p).Elem().FieldByName("methodTables").IsNil() { t.Fatal("source cache was not populated") }
 copied := latentCopyIR(p,map[any]any{})
 if !reflect.ValueOf(copied).Elem().FieldByName("methodTables").IsNil() { t.Fatal("speculative copy shares methodTables") }
 if reflect.ValueOf(p).Elem().FieldByName("methodTables").IsNil() { t.Fatal("copy changed source cache") }
 schema := &ir.JSONDecodeSchema{Nodes: []ir.JSONDecodeNode{{Kind:"witness", LiteralUnits:[]uint16{65}}}}
 functions := []ir.Function{{Boundary:ir.FunctionBoundary{Parameters:[]ir.BoundaryParameter{{Schema:schema}},Return:schema}}}
 clone := latentCopyIR(functions,map[any]any{})
 if clone[0].Boundary.Return == schema || clone[0].Boundary.Parameters[0].Schema != clone[0].Boundary.Return {t.Fatal("boundary schema ownership or graph identity lost")}
 clone[0].Boundary.Return.Nodes[0].LiteralUnits[0]=66
 if schema.Nodes[0].LiteralUnits[0]!=65 {t.Fatal("boundary schema aliases source")}
 identity := &ir.GeneratedType{}
 if latentCopyIR(identity,map[any]any{}) != identity { t.Fatal("shared pointer identity changed") }
}
`,
	}
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(folder, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "-count=1", ".")
	command.Dir = folder
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("speculative copier: %v\n%s", err, output)
	}
}

func TestLatentSpeculativeRefusesUnnamedExportedIRField(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "internal/ir/ir.go")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("package ir; type Function struct { AddedLater []int }"), 0600); err != nil {
		t.Fatal(err)
	}
	defer func() {
		failure := recover()
		if failure == nil || !strings.Contains(fmt.Sprint(failure), "Function.AddedLater") || !strings.Contains(fmt.Sprint(failure), "ir.go:1") {
			t.Fatalf("build refusal must name exported field and declaration: %v", failure)
		}
	}()
	generate(root, filepath.Join(root, "internal/lower/lower.go"))
}

func TestLatentSpeculativeBoundaryPolicyRequiredAtBuild(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	g := loadGenerator(root, filepath.Join(root, "internal/lower/lower.go"))
	policies := map[string]irFieldPolicy{}
	for key, value := range irSpecFields {
		policies[key] = value
	}
	delete(policies, "Function.Boundary")
	defer func() {
		failure := recover()
		if failure == nil || !strings.Contains(fmt.Sprint(failure), "Function.Boundary") || !strings.Contains(fmt.Sprint(failure), "isn't named") {
			t.Fatalf("missing Boundary must fail at generation, naming it: %v", failure)
		}
	}()
	g.checkIRFields(token.NewFileSet(), policies)
}

func TestLatentSpeculativeEveryIRFieldPolicyIsCurrent(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
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
			t.Errorf("stale speculative field policy %s", key)
		}
	}
	generated := string(generate(root, filepath.Join(root, "internal/lower/lower.go")))
	for _, want := range []string{"result.Boundary = latentCopyIR(value.Boundary, seen)", "result.Source = latentCopyIR(value.Source, seen)"} {
		if !strings.Contains(generated, want) {
			t.Errorf("generated function copier lacks %s", want)
		}
	}
}
