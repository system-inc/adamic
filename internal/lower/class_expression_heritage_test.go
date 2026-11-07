package lower

import (
	"errors"
	"strings"
	"testing"
)

// A class expression is class-like but isn't a ClassDeclaration, so the namespace walks and the class
// lowering that read its heritage clauses or members through AsClassDeclaration panicked ("interface
// conversion: ast.nodeData is *ast.ClassExpression, not *ast.ClassDeclaration"). Every class-like read
// goes through ClassLikeData, which both kinds share (#nkxtfkq), so each program stops at its own
// NotYet: stage 0 doesn't lower a class expression yet.
func TestClassExpressionHeritageIsNotYet(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source, where string }{
		{
			"implements",
			"interface Node { readonly kind: number; readonly pos: number; }\nconst Signature = class implements Node { kind = 171; pos = 1; };\nconsole.log(String(new Signature().pos));\n",
			":2:19",
		},
		{
			"extends a class expression",
			"const B = class { x = 1; };\nclass C extends B {}\nconsole.log(String(new C().x));\n",
			":1:11",
		},
		{
			"class expression extends",
			"class A { y = 2; }\nconst D = class extends A {};\nconsole.log(String(new D().y));\n",
			":2:11",
		},
		// Correctness's census stop (#57pc1at): with a namespace in the program, namespaceInitialization read
		// a class expression's heritage through AsClassDeclaration and the compiler crashed, exit 4. tsc's
		// _namespaces/ts.moduleSpecifiers.ts root stopped there past the Axolotl wall.
		{
			"namespace beside a class expression",
			"namespace N {\n    export const k = 1;\n}\nclass Base {\n    value(): number {\n        return N.k;\n    }\n}\nconst Derived = class extends Base {};\nconsole.log(String(new Derived().value()));\n",
			":9:17",
		},
		// The namespace walk still visits a class expression's extends clause: N.B is read before its
		// var is assigned, which Node throws on.
		{
			"namespace base read before its assignment",
			"namespace N {\n    export const Late = class extends N.B {};\n    export var B = class { v = 1; };\n}\nconsole.log(String(new N.Late().v));\n",
			"",
		},
	} {
		for _, extension := range []string{".ts", ".a"} {
			test.Run(probe.name+extension, func(test *testing.T) {
				test.Parallel()
				defer func() {
					if failure := recover(); failure != nil {
						test.Fatalf("panicked: %v", failure)
					}
				}()
				_, err := lowerSourceAs(test, probe.source, extension)
				if probe.where == "" {
					if err == nil {
						test.Fatalf("lowered, where Node throws on reading N.B before its assignment")
					}
					test.Logf("stops: %v", err)
					return
				}
				var notYet *NotYet
				if !errors.As(err, &notYet) ||
					notYet.What != "a ClassExpression" ||
					!strings.HasSuffix(notYet.Where, probe.where) {
					test.Fatalf("want NotYet a ClassExpression at %s, got %v", probe.where, err)
				}
			})
		}
	}
}

// Correctness's census program with a class declaration in place of the expression, the control
// (#57pc1at): the same namespace walk reads a ClassDeclaration's heritage through ClassLikeData, and it
// lowers, as on main, where Node prints 1.
func TestClassDeclarationHeritageBesideANamespaceLowers(test *testing.T) {
	test.Parallel()
	const source = "namespace N {\n    export const k = 1;\n}\nclass Base {\n    value(): number {\n        return N.k;\n    }\n}\nclass Derived extends Base {}\nconsole.log(String(new Derived().value()));\n"
	for _, extension := range []string{".ts", ".a"} {
		test.Run(extension, func(test *testing.T) {
			test.Parallel()
			if _, err := lowerSourceAs(test, source, extension); err != nil {
				test.Fatalf("the class declaration control must lower: %v", err)
			}
		})
	}
}
