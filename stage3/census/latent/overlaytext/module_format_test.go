package overlaytext

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
	"testing"
)

// The sole known build-setting difference: independent units still use the
// conservative nil key. Resolving elementKey is a separate filed task.
var knownLoweringSettingExceptions = []string{"elementKey"}

// Context fields (program, checker, entry, plan, cycleTypes) belong to each
// attempt. This guard checks the declared build settings, rather than equating
// independent-unit context with production's whole-program context.
func loweringSettings(source, function string) (map[string]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "lowering.go", source, 0)
	if err != nil {
		return nil, err
	}
	var scope ast.Node = file
	if function != "" {
		scope = nil
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == function {
				scope = fn.Body
			}
		}
		if scope == nil {
			return nil, fmt.Errorf("missing %s", function)
		}
	}
	var literals []*ast.CompositeLit
	ast.Inspect(scope, func(node ast.Node) bool {
		if lit, ok := node.(*ast.CompositeLit); ok {
			if typ, ok := lit.Type.(*ast.Ident); ok && typ.Name == "lowering" {
				literals = append(literals, lit)
			}
		}
		return true
	})
	if len(literals) != 1 {
		return nil, fmt.Errorf("want one lowering literal, got %d", len(literals))
	}
	fields := map[string]ast.Expr{}
	for _, entry := range literals[0].Elts {
		pair, ok := entry.(*ast.KeyValueExpr)
		if !ok {
			return nil, fmt.Errorf("unkeyed lowering field")
		}
		key, ok := pair.Key.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("invalid lowering field")
		}
		fields[key.Name] = pair.Value
	}
	// Production sets some fields after its literal. Read every direct field
	// assignment in this function, including those following the deferred hook.
	if function != "" {
		ast.Inspect(scope, func(node ast.Node) bool {
			assignment, ok := node.(*ast.AssignStmt)
			if !ok || assignment.Tok != token.ASSIGN || len(assignment.Lhs) != len(assignment.Rhs) {
				return true
			}
			for i, target := range assignment.Lhs {
				field, ok := target.(*ast.SelectorExpr)
				if !ok {
					continue
				}
				owner, ok := field.X.(*ast.Ident)
				if ok && owner.Name == "lowering" {
					fields[field.Sel.Name] = assignment.Rhs[i]
				}
			}
			return true
		})
	}
	result := map[string]string{}
	for _, key := range []string{"moduleFormat", "elementKey"} {
		value, exists := fields[key]
		if !exists {
			continue
		}
		var text bytes.Buffer
		if err := format.Node(&text, token.NewFileSet(), value); err != nil {
			return nil, err
		}
		setting := text.String()
		if key == "moduleFormat" && setting == "moduleformat.Resolve(ModuleFormatCommonJSBundle, ModuleFormatEcmaScript)" {
			setting = "moduleFormat"
		}
		result[key] = setting
	}
	return result, nil
}

func moduleFormatGap(production, census string) error {
	if !reflect.DeepEqual(knownLoweringSettingExceptions, []string{"elementKey"}) {
		return fmt.Errorf("known exceptions must remain exactly elementKey")
	}
	p, err := loweringSettings(production, "lowerWithElementKey")
	if err != nil {
		return err
	}
	c, err := loweringSettings(census, "")
	if err != nil {
		return err
	}
	if p["elementKey"] != "key" || c["elementKey"] != "" {
		return fmt.Errorf("elementKey exception changed; resolve the filed task and remove the exception")
	}
	for _, exception := range knownLoweringSettingExceptions {
		delete(p, exception)
		delete(c, exception)
	}
	if p["moduleFormat"] != "moduleFormat" || !reflect.DeepEqual(p, c) {
		return fmt.Errorf("moduleFormat build settings disagree: production %v, latent %v", p, c)
	}
	return nil
}

func TestLatentModuleFormatMatchesProduction(t *testing.T) {
	p, err := os.ReadFile("../../../../internal/lower/lower.go")
	if err != nil {
		t.Fatal(err)
	}
	c, err := os.ReadFile("../units.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	if err := moduleFormatGap(string(p), string(c)); err != nil {
		t.Fatal(err)
	}
}

func TestLatentModuleFormatOmissionMutant(t *testing.T) {
	p, err := os.ReadFile("../../../../internal/lower/lower.go")
	if err != nil {
		t.Fatal(err)
	}
	c, err := os.ReadFile("../units.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	mutant := strings.Replace(string(c), "moduleFormat: moduleformat.Resolve(ModuleFormatCommonJSBundle, ModuleFormatEcmaScript), ", "", 1)
	if mutant == string(c) {
		t.Fatal("mutant did not change literal")
	}
	if err := moduleFormatGap(string(p), mutant); err == nil || !strings.Contains(err.Error(), "moduleFormat") {
		t.Fatalf("omission mutant survived: %v", err)
	}
}

func TestProductionModuleFormatAssignmentOmissionMutant(t *testing.T) {
	p, err := os.ReadFile("../../../../internal/lower/lower.go")
	if err != nil {
		t.Fatal(err)
	}
	c, err := os.ReadFile("../units.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	mutant := strings.Replace(string(p), "\tlowering.moduleFormat = moduleFormat\n", "", 1)
	if mutant == string(p) {
		t.Fatal("mutant did not delete production assignment")
	}
	if err := moduleFormatGap(mutant, string(c)); err == nil || !strings.Contains(err.Error(), "moduleFormat") {
		t.Fatalf("production assignment omission mutant survived: %v", err)
	}
}
