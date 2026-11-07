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

func TestSyntaxModuleDeclarationMutants(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"initializer", "scoped read"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/syntax_module_declarations.a"),
			)
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatal(err)
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
				test.Fatal("missing distinct namespace bindings")
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
				test.Fatal("mutant changed no namespace operation")
			}
			truth := onNode(test, path)
			generated := onJavaScriptBackend(test, program)
			if difference := disagreement(truth, generated); difference != "stdout differs" {
				test.Fatalf("JavaScript mutant: %q: %s", difference, generated.stderr)
			}
			result, binary := natively(test, program)
			if difference := disagreement(truth, result); difference != "stdout differs" {
				test.Fatalf("native mutant: %q: %s", difference, result.stderr)
			}
			if report := leaks(test, program, binary); report != "" {
				test.Fatal(report)
			}
			test.Log("caught in both backends by source Node stdout comparison")
		})
	}
}
