package lower

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdamicNullishAssertionsAreRefused(test *testing.T) {
	test.Parallel()
	for _, extension := range []string{"a"} {
		for _, operand := range []string{"undefined", "null"} {
			for _, source := range []string{
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
				"const box: {value: number; other: number} = {value: " +
					operand +
					"!, get other(): number { return 1; }};",
				"console.log(" + operand + "!);",
				"const absent = " + operand + "; console.log(absent!);",
				"function f(value: " + operand + "): void { console.log(value!); }",
			} {
				test.Run(extension+"/"+operand+"/"+source, func(test *testing.T) {
					test.Parallel()
					path := filepath.Join(test.TempDir(), "main."+extension)
					if err := os.WriteFile(path, []byte(source), 0644); err != nil {
						test.Fatal(err)
					}
					program, err := load.Load([]string{path})
					if err != nil {
						test.Fatal(err)
					}
					_, err = Lower(context.Background(), program)
					if strings.Contains(source, "="+operand+"!") || strings.Contains(source, "value: "+operand+"!") {
						var notYet *NotYet
						if !errors.As(err, &notYet) {
							test.Fatalf("want unsupported placeholder result/storage, got %v", err)
						}
						return
					}
					var refused *Refused
					if !errors.As(err, &refused) {
						test.Fatalf("want exactly-nullish Refused, got %v", err)
					}
					if refused.What != "the non-null assertion !" ||
						refused.Fix != "write ?? panic('why it can't be missing'), or narrow and handle the missing case" {
						test.Fatalf("wrong diagnostic: %v", err)
					}
				})
			}
		}
	}
}
