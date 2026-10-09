package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/syntax_module_declarations.a", true, false})
}

func TestSyntaxModuleDeclarationMutants(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"initializer", "scoped read"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/syntax_module_declarations.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			first, second := -1, -1
			var secondInitializer ir.Expression
			for _, statement := range program.Main {
				declaration, ok := statement.(ir.Declare)
				if !ok || program.Locals[declaration.Local].Name != "Element" {
					continue
				}
				literal, ok := declaration.Value.(ir.StringConstant)
				if !ok {
					continue
				}
				switch program.Strings[literal.Index] {
				case "Element":
					first = declaration.Local
				case "Fragment":
					second = declaration.Local
					secondInitializer = declaration.Value
				}
			}
			if first < 0 || second < 0 {
				t.Fatal("missing distinct namespace bindings")
			}
			changed := false
			if name == "initializer" {
				for index, statement := range program.Main {
					declaration, ok := statement.(ir.Declare)
					if ok && declaration.Local == first {
						declaration.Value = secondInitializer
						program.Main[index] = declaration
						changed = true
					}
				}
			} else {
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
					if read, ok := value.(ir.Read); ok && read.Local == first {
						read.Local = second
						changed = true
						return read
					}
					return value
				})
			}
			if !changed {
				t.Fatal("mutant changed no namespace operation")
			}
			truth := onNode(t, path)
			generated := onJavaScriptBackend(t, program)
			if difference := disagreement(truth, generated); difference != "stdout differs" {
				t.Fatalf("JavaScript mutant: %q: %s", difference, generated.stderr)
			}
			result, binary := natively(t, program)
			if difference := disagreement(truth, result); difference != "stdout differs" {
				t.Fatalf("native mutant: %q: %s", difference, result.stderr)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Log("caught in both backends by source Node stdout comparison")
		})
	}
}
