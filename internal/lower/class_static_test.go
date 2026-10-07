package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestStaticMethodsLowerAsFunctions(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"class_static.a", "class_static_calls.a", "class_static_alias.a", "class_static_alias_tdz.a", "class_static_tdz.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("..", "oracle", "testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowerSource(t, string(source))
			if err != nil {
				t.Fatal(err)
			}
			for _, function := range program.Functions {
				if function.Name == "Stepper_static_down" {
					if len(function.Parameters) != 2 || program.Locals[function.Parameters[0]].Name != "this" || program.Locals[function.Parameters[1]].Name != "count" {
						t.Fatalf("static down must have its class receiver and count parameter, got %+v", function)
					}
					return
				}
			}
			t.Fatal("static down was not lowered")
		})
	}
}

func TestStaticUsesFollowCurrentClassRepresentation(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source, message string
		refused               bool
	}{
		{"generic class", `class Box<T> { static down(count: number): number { return count + 1; } } console.log(` + "`${Box.down(3)}`" + `);`, "", false},
		{"anonymous static field", `console.log(` + "`${(class { static readonly count = 4; }).count}`" + `);`, "ClassExpression", false},
		{"anonymous static method", `console.log(` + "`${(class { static down(count: number): number { return count + 1; } }).down(3)}`" + `);`, "a method of an anonymous class", false},
		{"static field", `class Stepper { static readonly count = 4; } console.log(` + "`${Stepper.count}`" + `);`, "", false},
		{"destructured static method", `class Stepper { static down(count: number): number { return count + 1; } } const { down } = Stepper; console.log(` + "`${down(3)}`" + `);`, "a method in object destructuring", true},
		{"static function value", `class Stepper { static down(count: number): number { return count + 1; } } const down = Stepper.down; console.log(` + "`${down(3)}`" + `);`, "down would lose its object", true},
		{"static this", `class Stepper { static down(count: number): number { return count + 1; } static twice(count: number): number { return this.down(count); } } console.log(` + "`${Stepper.twice(3)}`" + `);`, "", false},
		{"alias runtime value", `class Stepper { static down(count: number): number { return count + 1; } } const Alias = Stepper; console.log(` + "`${Alias === Stepper}`" + `);`, "", false},
		{"wider alias runtime value", `class Stepper { static down(count: number): number { return count + 1; } } const Alias: { down(count: number): number } = Stepper; console.log(typeof Alias);`, "", false},
		{"mutable alias", `class Stepper { static down(count: number): number { return count + 1; } } let Alias = Stepper; console.log(` + "`${Alias.down(3)}`" + `);`, "", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if probe.message == "" {
				if err != nil {
					t.Fatalf("want supported static use, got %v", err)
				}
				return
			}
			var gap *NotYet
			var refusal *Refused
			if (probe.refused && !errors.As(err, &refusal)) || (!probe.refused && !errors.As(err, &gap)) || !strings.Contains(err.Error(), probe.message) {
				t.Fatalf("want named diagnostic %q, got %v", probe.message, err)
			}
		})
	}
}

func TestClassConstructorTypesAreNotInstances(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(`class Stepper { static down(count: number): number { return count + 1; } }`), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	declaration := file.Statements.Nodes[0]
	l := &lowering{checker: checked, program: program}
	constructor := checked.GetTypeOfSymbol(checked.GetSymbolAtLocation(declaration.Name()))
	_, err = l.instantiate(declaration, constructor, declaration)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(err.Error(), "constructor type instead of an instance type") {
		t.Fatalf("want constructor-type gap, got %v", err)
	}
	if mapper := l.typeMapperOf(declaration, constructor); mapper != nil {
		t.Fatal("a constructor type must not supply instance type arguments")
	}
}
