package latent

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func passLedgerTable(source []byte) ([]PassStep, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "pass_ledger.go.txt", source, 0)
	if err != nil {
		return nil, err
	}
	var result []PassStep
	found := 0
	for _, decl := range file.Decls {
		group, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range group.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok || len(value.Names) != 1 || value.Names[0].Name != "latentPostLoweringSteps" {
				continue
			}
			found++
			if len(value.Values) != 1 {
				return nil, fmt.Errorf("step table must have one initializer")
			}
			list, ok := value.Values[0].(*ast.CompositeLit)
			if !ok {
				return nil, fmt.Errorf("step table must be a literal")
			}
			for _, row := range list.Elts {
				literal, ok := row.(*ast.CompositeLit)
				if !ok {
					return nil, fmt.Errorf("step must be a literal")
				}
				step := PassStep{}
				for _, element := range literal.Elts {
					pair, ok := element.(*ast.KeyValueExpr)
					if !ok {
						return nil, fmt.Errorf("step must name its fields")
					}
					key, ok := pair.Key.(*ast.Ident)
					if !ok {
						return nil, fmt.Errorf("invalid step field")
					}
					switch key.Name {
					case "Name", "Production":
						text, ok := pair.Value.(*ast.BasicLit)
						if !ok || text.Kind != token.STRING {
							return nil, fmt.Errorf("step field must be a string")
						}
						word, err := strconv.Unquote(text.Value)
						if err != nil {
							return nil, err
						}
						if key.Name == "Name" {
							step.Name = word
						} else {
							step.Production = word
						}
					case "WholeProgram", "Scoped":
						boolean, ok := pair.Value.(*ast.Ident)
						if !ok || (boolean.Name != "true" && boolean.Name != "false") {
							return nil, fmt.Errorf("invalid whole-program flag")
						}
						if key.Name == "Scoped" {
							step.Scoped = boolean.Name == "true"
						} else {
							step.WholeProgram = boolean.Name == "true"
						}
					default:
						return nil, fmt.Errorf("unknown step field %s", key.Name)
					}
				}
				result = append(result, step)
			}
		}
	}
	if found != 1 || len(result) == 0 {
		return nil, fmt.Errorf("expected one nonempty step table, found %d", found)
	}
	seen := map[string]bool{}
	for _, step := range result {
		if step.Name == "" || seen[step.Name] {
			return nil, fmt.Errorf("empty or duplicate step %q", step.Name)
		}
		seen[step.Name] = true
	}
	return result, nil
}

func passPrinted(node ast.Node) string {
	var buffer bytes.Buffer
	if err := format.Node(&buffer, token.NewFileSet(), node); err != nil {
		panic(err)
	}
	return buffer.String()
}

// Follow LowerWithOptions's local function calls, so a production wrapper or
// helper rename cannot silently disconnect the guard from the actual tail.
func productionPassKeys(source []byte) ([]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "internal/lower/lower.go", source, 0)
	if err != nil {
		return nil, err
	}
	functions := map[string]*ast.FuncDecl{}
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil {
			functions[fn.Name.Name] = fn
		}
	}
	seen := map[string]bool{}
	var tails [][]string
	var visit func(string) error
	visit = func(name string) error {
		if seen[name] {
			return nil
		}
		seen[name] = true
		fn := functions[name]
		if fn == nil {
			return nil
		}
		moduleSeen := false
		for index, statement := range fn.Body.List {
			ast.Inspect(statement, func(node ast.Node) bool {
				if call, ok := node.(*ast.CallExpr); ok {
					if method, ok := call.Fun.(*ast.SelectorExpr); ok && method.Sel.Name == "moduleStatements" {
						moduleSeen = true
					}
				}
				return true
			})
			branch, ok := statement.(*ast.IfStmt)
			if !ok || !moduleSeen || passPrinted(branch.Cond) != "lowering.unlowerable != nil" {
				continue
			}
			if len(branch.Body.List) != 1 || passPrinted(branch.Body.List[0]) != "return nil, lowering.unlowerable" {
				return fmt.Errorf("unlowerable check changed")
			}
			var keys []string
			for _, tail := range fn.Body.List[index+1:] {
				if returned, ok := tail.(*ast.ReturnStmt); ok {
					if passPrinted(returned) != "return lowering, nil" {
						return fmt.Errorf("tail return changed: %s", passPrinted(returned))
					}
					continue
				}
				var call *ast.CallExpr
				switch node := tail.(type) {
				case *ast.ExprStmt:
					call, _ = node.X.(*ast.CallExpr)
				case *ast.IfStmt:
					if node.Init == nil {
						var calls []string
						ast.Inspect(node.Body, func(child ast.Node) bool {
							if invocation, ok := child.(*ast.CallExpr); ok {
								calls = append(calls, strings.TrimPrefix(passPrinted(invocation.Fun), "lowering."))
							}
							return true
						})
						keys = append(keys, "if "+passPrinted(node.Cond)+" [calls:"+strings.Join(calls, ",")+"]")
						continue
					}
					assignment, ok := node.Init.(*ast.AssignStmt)
					if ok && len(assignment.Rhs) == 1 {
						call, _ = assignment.Rhs[0].(*ast.CallExpr)
					}
				case *ast.AssignStmt:
					keys = append(keys, passPrinted(node))
					continue
				}
				if call == nil {
					return fmt.Errorf("unclassified production tail step: %s", passPrinted(tail))
				}
				keys = append(keys, "call:"+strings.TrimPrefix(passPrinted(call.Fun), "lowering."))
			}
			tails = append(tails, keys)
		}
		var children []string
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				if id, ok := call.Fun.(*ast.Ident); ok {
					children = append(children, id.Name)
				}
			}
			return true
		})
		for _, child := range children {
			if err := visit(child); err != nil {
				return err
			}
		}
		return nil
	}
	if functions["LowerWithOptions"] == nil {
		return nil, fmt.Errorf("LowerWithOptions missing")
	}
	if err := visit("LowerWithOptions"); err != nil {
		return nil, err
	}
	if len(tails) != 1 {
		return nil, fmt.Errorf("expected one reachable post-module/unlowerable tail, found %d", len(tails))
	}
	return tails[0], nil
}

func checkPassLedgerTail(table []PassStep, source []byte) error {
	actual, err := productionPassKeys(source)
	if err != nil {
		return err
	}
	var expected []string
	for _, step := range table {
		key := step.Production
		if key == "" {
			key = "call:" + step.Name
		}
		expected = append(expected, key)
	}
	if !reflect.DeepEqual(expected, actual) {
		return fmt.Errorf("post-lowering step names/order drift: table %v; main %v", expected, actual)
	}
	return nil
}

func readPassLedgerMain(t *testing.T) ([]PassStep, []byte) {
	t.Helper()
	source, err := os.ReadFile("pass_ledger.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	table, err := passLedgerTable(source)
	if err != nil {
		t.Fatal(err)
	}
	main, err := os.ReadFile("../../../internal/lower/lower.go")
	if err != nil {
		t.Fatal(err)
	}
	return table, main
}

func TestLatentPassLedgerMatchesMainTailPb47(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/lower.go
	table, main := readPassLedgerMain(t)
	if err := checkPassLedgerTail(table, main); err != nil {
		t.Fatal(err)
	}
}

func TestLatentPassLedgerDroppedStepMutantPb47(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/lower.go
	table, main := readPassLedgerMain(t)
	if err := checkPassLedgerTail(table[1:], main); err == nil || !strings.Contains(err.Error(), "names/order drift") {
		t.Fatalf("dropped-step mutant survived: %v", err)
	}
}

func TestLatentPassLedgerOnlyRecordsOmissionsPb47(t *testing.T) {
	t.Parallel()
	table, _ := readPassLedgerMain(t)
	for _, stopped := range []bool{false, true} {
		entries := OmittedPassLedger(table, stopped)
		if len(entries) != len(table) {
			t.Fatal("incomplete ledger")
		}
		whole := 0
		for i, entry := range entries {
			want := "not_run_scoped"
			if table[i].WholeProgram {
				want = "not_run_whole_program"
				whole++
			}
			if stopped {
				want = "not_reached_lowering_stop"
			}
			if entry.Name != table[i].Name || entry.Status != want {
				t.Fatalf("entry %d: %+v, want %s", i, entry, want)
			}
		}
		if whole != 4 {
			t.Fatalf("whole-program-only steps: %d", whole)
		}
	}
	// A unit owns its entries; changing one ledger cannot change another or the table.
	left, right := OmittedPassLedger(table, false), OmittedPassLedger(table, false)
	left[0].Status = "mutated"
	if right[0].Status != "not_run_scoped" {
		t.Fatal("ledgers alias")
	}
}

func TestLatentPassLedgerNestedCallDriftPb47(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/lower.go
	table, main := readPassLedgerMain(t)
	mutant := bytes.Replace(main, []byte("lowering.applyProgramMetadata(plan)"), []byte("lowering.unlistedPlanPass(plan)"), 1)
	if bytes.Equal(main, mutant) {
		t.Fatal("production plan call missing")
	}
	if err := checkPassLedgerTail(table, mutant); err == nil {
		t.Fatal("nested pass rename was not detected")
	}
}
