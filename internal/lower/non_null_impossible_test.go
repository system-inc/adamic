package lower

import (
	"context"
	"errors"
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
				"let value: number = " + operand + "!; value=1; console.log(`${value}`);",
				"const value: number = " + operand + "!;",
				"class Box { value: number = " + operand + "!; }",
				"function f(value: number = " + operand + "!): number {return value;}",
				"let value=1; value=" + operand + "!;",
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
