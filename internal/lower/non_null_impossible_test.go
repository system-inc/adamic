package lower

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"testing"
)

func TestExactlyNullishAssertionsAreRefused(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{"a", "ts"} {
		for _, operand := range []string{"undefined", "null"} {
			for _, source := range []string{
				"function f(value: number = " + operand + "!): number {return value;}",
				"function f(): number { return " + operand + "!; }",
				"console.log(`${" + operand + "! + 1}`);",
				"console.log(`${" + operand + "! === undefined}`);",
				"let value=1; console.log(`${value=" + operand + "!}`);",
				"let value=1; const assigned=(value=" + operand + "!);",
				"let value=1; function f(): number { return value=" + operand + "!; }",
				"let value=1; value += " + operand + "!;",
				"let value=1; value=(" + operand + "! + 1);",
				"let value=1; const absent=" + operand + "; value=absent!;",
				"let value=1; let other=2; value=other=" + operand + "!;",
				"let value=1; const clear=(): number => value=" + operand + "!;",
				"const values=[1]; values[0]=" + operand + "!;",
				"class Box { value=1; } const box=new Box(); console.log(`${box.value=" + operand + "!}`);",
				"const value: number = true ? " + operand + "! : 1;",
				"const value: number = 1 || " + operand + "!;",
				"const values = [" + operand + "!];",
				"const values = [..." + operand + "!];",
				"const box: {value: number; other: number} = {value: " + operand + "!, get other(): number { return 1; }};",
				"console.log(" + operand + "!);",
				"const absent = " + operand + "; console.log(absent!);",
				"function f(value: " + operand + "): void { console.log(value!); }",
			} {
				t.Run(extension+"/"+operand+"/"+source, func(t *testing.T) {
					t.Parallel()
					path := filepath.Join(t.TempDir(), "main."+extension)
					if err := os.WriteFile(path, []byte(source), 0644); err != nil {
						t.Fatal(err)
					}
					program, err := load.Load([]string{path})
					if err != nil {
						t.Fatal(err)
					}
					_, err = Lower(context.Background(), program)
					var refused *Refused
					if !errors.As(err, &refused) {
						t.Fatalf("want exactly-nullish Refused, got %v", err)
					}
					if refused.What != "a non-null assertion whose operand is exactly "+operand || refused.Fix != "declare the variable optional and assign undefined" {
						t.Fatalf("wrong diagnostic: %v", err)
					}
				})
			}
		}
	}
}

func TestStandaloneDeinitializationUsesReadiness(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{"a", "ts"} {
		for _, operand := range []string{"undefined", "null"} {
			for _, source := range []string{
				"let value=4; value=" + operand + "!; console.log(`${value}`);",
				"let value=4; value! = " + operand + "!; console.log(`${value}`);",
				"let value=4; (value=(" + operand + "!)); console.log(`${value}`);",
				"const box={value:4}; box.value=" + operand + "!; console.log(`${box.value}`);",
				"const box={value:4}; box.value! = " + operand + "!; console.log(`${box.value}`);",
				"const box={value:4}; if(true) { (box.value=(" + operand + "!)); } console.log(`${box.value}`);",
			} {
				t.Run(extension+"/"+operand+"/"+source, func(t *testing.T) {
					t.Parallel()
					path := filepath.Join(t.TempDir(), "main."+extension)
					if err := os.WriteFile(path, []byte(source), 0644); err != nil {
						t.Fatal(err)
					}
					loaded, err := load.Load([]string{path})
					if err != nil {
						t.Fatal(err)
					}
					program, err := Lower(context.Background(), loaded)
					if err != nil {
						t.Fatal(err)
					}
					writes, reads, assertions := 0, 0, 0
					walk(program.Main, func(node any) bool {
						switch value := node.(type) {
						case ir.Assign:
							if value.Uninitialized {
								writes++
							}
						case ir.SetProperty:
							if value.Uninitialized {
								writes++
							}
						case ir.Read:
							if value.Readiness != "" {
								reads++
							}
						case ir.Property:
							if value.Readiness != "" {
								reads++
							}
						case ir.Coalesce:
							if value.Panic != nil {
								assertions++
							}
						}
						return true
					})
					if writes != 1 || reads != 1 || assertions != 0 {
						t.Fatalf("want one deinitialization and readiness read, no assertion: writes=%d reads=%d assertions=%d", writes, reads, assertions)
					}
				})
			}
		}
	}
}
