package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// Every refusal rule reads the slot it judges a value against through l.slotType, never from the
// checker, so a caller that sets a slot the checker doesn't give (a delayed Array write's element) and
// runs the registry over a value is judged by every rule, not by the ones a second list remembered
// (#b6m00y7). The walk starts at each rule the registry holds, found from its own function rather than
// from a list here, and follows every function and method of this package that it names, as a call or
// as a value, and every package variable it reads, so a rule registered later, or a helper any rule
// reaches, is held too. A name shared by several methods follows all of them, which errs toward
// reaching more.
func TestRefusalRulesReadTheirSlotThroughSlotType(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	functions := map[string][]*ast.FuncDecl{}
	variables := map[string][]ast.Expr{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, entry.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			switch declaration := declaration.(type) {
			case *ast.FuncDecl:
				functions[declaration.Name.Name] = append(functions[declaration.Name.Name], declaration)
			case *ast.GenDecl:
				for _, specification := range declaration.Specs {
					value, isValue := specification.(*ast.ValueSpec)
					if !isValue {
						continue
					}
					for index, name := range value.Names {
						if index < len(value.Values) {
							variables[name.Name] = append(variables[name.Name], value.Values[index])
						}
					}
				}
			}
		}
	}
	// reachedFrom is each function the walk reached, by the rule it reached it from first.
	reachedFrom := map[string]string{}
	var queue []string
	for _, rule := range refusalRules.rules {
		function := runtime.FuncForPC(reflect.ValueOf(rule.function).Pointer())
		if function == nil {
			t.Fatalf("refusal rule %s has no function to walk", rule.name)
		}
		name := strings.TrimSuffix(function.Name()[strings.LastIndex(function.Name(), ".")+1:], "-fm")
		if len(functions[name]) == 0 {
			t.Errorf("refusal rule %s registered %s, which isn't a function or method of this package the walk can read; register a named one", rule.name, function.Name())
			continue
		}
		if _, reached := reachedFrom[name]; !reached {
			reachedFrom[name] = rule.name
			queue = append(queue, name)
		}
	}
	reach := func(name, rule string) {
		if _, reached := reachedFrom[name]; !reached && (len(functions[name]) > 0 || len(variables[name]) > 0) {
			reachedFrom[name] = rule
			queue = append(queue, name)
		}
	}
	var direct []string
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		rule := reachedFrom[name]
		var bodies []ast.Node
		for _, declaration := range functions[name] {
			if declaration.Body != nil {
				bodies = append(bodies, declaration.Body)
			}
		}
		for _, value := range variables[name] {
			bodies = append(bodies, value)
		}
		for _, body := range bodies {
			ast.Inspect(body, func(node ast.Node) bool {
				switch node := node.(type) {
				case *ast.SelectorExpr:
					if node.Sel.Name == "GetContextualType" && name != "slotType" {
						direct = append(direct, files.Position(node.Pos()).String()+" in "+name+", reached from rule "+rule)
					}
					reach(node.Sel.Name, rule)
				case *ast.Ident:
					reach(node.Name, rule)
				}
				return true
			})
		}
	}
	sort.Strings(direct)
	if len(direct) > 0 {
		t.Errorf("a refusal rule reads the checker's contextual type directly, so a slot a caller sets never reaches it; read l.slotType instead:\n%s", strings.Join(direct, "\n"))
	}
	// The walk has to reach the lookup itself from the rules that judge a value against its slot, or it
	// proved nothing about them.
	for _, name := range []string{"slotType", "refuseWidening", "refuseVoidWidening", "refuseOptionalWidening", "classViewRefusal", "stringWideningRefusal"} {
		if _, reached := reachedFrom[name]; !reached {
			t.Errorf("the walk from the registry never reached %s", name)
		}
	}
}
