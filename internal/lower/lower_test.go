package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func lowerSource(t *testing.T, source string) (*ir.Program, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return Lower(context.Background(), program)
}

func TestConsoleLowersToWriteLine(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "console.log('one');\nconsole.error(`two`);\nconsole.log('one');\n")
	if err != nil {
		t.Fatal(err)
	}
	want := []ir.Statement{
		ir.WriteLine{Stream: ir.Stdout, Value: ir.StringConstant{Index: 0}},
		ir.WriteLine{Stream: ir.Stderr, Value: ir.StringConstant{Index: 1}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.StringConstant{Index: 0}},
	}
	if len(program.Main) != len(want) || strings.Join(program.Strings, ",") != "one,two" {
		t.Fatalf("got %v with strings %q", program.Main, program.Strings)
	}
	for index := range want {
		if program.Main[index] != want[index] {
			t.Errorf("instruction %d: got %v, want %v", index, program.Main[index], want[index])
		}
	}
	if program.Source != "main.a" {
		t.Errorf("Source: got %q, want main.a", program.Source)
	}
}

func TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name   string
		source string
		want   string
	}{
		{"a closure", "function outer(): number {\n\tfunction inner(): number {\n\t\treturn 1;\n\t}\n\treturn inner();\n}\nconsole.log(`${outer()}`);\n", "main.a:2:2: stage 0 can't lower a function inside a function (a closure) yet"},
		{"a class inside a function", "function make(): number {\n\tclass Box {\n\t\treadonly size = 1;\n\t}\n\treturn 1;\n}\nconsole.log(`${make()}`);\n", "main.a:2:2: stage 0 can't lower a class inside a function yet"},
		{"join on an array of functions", "const steps = [(): number => 1];\nconsole.log(steps.join(','));\n", "main.a:2:13: stage 0 can't lower join on an array of objects, arrays, maps or functions yet"},
		{"an array of number | undefined", "const values: (number | undefined)[] = [1, undefined];\n", "main.a:1:40: stage 0 can't lower an array of number | undefined yet"},
		{"a function value returning number | undefined", "const pick = (): number | undefined => 1;\n", "main.a:1:14: stage 0 can't lower a function value returning number | undefined yet"},
		{"?. on a string", "const words = ['a'];\nconsole.log(`${words[0]?.length ?? 0}`);\n", "main.a:2:16: stage 0 can't lower optional chaining on a string yet"},
		{"concat on an array of arrays", "const grid: number[][] = [[1]];\nconst more = grid.concat([[2]]);\n", "main.a:2:14: stage 0 can't lower concat on an array of arrays yet"},
		{"a field of type number | undefined", "interface Spot {\n\tx: number | undefined;\n}\nconst spot: Spot = { x: 1 };\nconsole.log(`${spot.x ?? -1}`);\n", "main.a:5:16: stage 0 can't lower a field of type number | undefined yet"},
		{"a captured number | undefined", "function run(): number {\n\tlet best: number | undefined;\n\tconst pick = (): void => {\n\t\tbest = 1;\n\t};\n\tpick();\n\treturn best ?? -1;\n}\nconsole.log(`${run()}`);\n", "main.a:4:3: stage 0 can't lower a number | undefined variable a function value captures yet"},
		{"a rest parameter", "function sum(...values: number[]): number {\n\treturn values.length;\n}\nconsole.log(`${sum(1, 2)}`);\n", "main.a:1:14: stage 0 can't lower a parameter that isn't a plain name yet"},
		{"a default for a boolean", "function greet(loud = false): string {\n\treturn loud ? 'HI' : 'hi';\n}\nconsole.log(greet());\n", "main.a:1:16: stage 0 can't lower a default for a boolean parameter yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Error(), probe.want) {
				t.Errorf("got %v, want an error ending %q", err, probe.want)
			}
		})
	}
}

// What 0.1 refuses for good is said as a refusal with its fix, never as "not yet".
func TestWhatZeroOneRefusesIsRefusedWithAFix(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name   string
		source string
		want   string
	}{
		{"a string as a condition", "const name = 'x';\nif (name) {\n\tconsole.log(name);\n}\n", "main.a:2:5: Adamic 0.1 refuses a string as a condition; compare it explicitly, like name.length > 0 or count !== 0"},
		{"a number as a loop condition", "let count = 3;\nwhile (count) {\n\tcount -= 1;\n}\n", "main.a:2:8: Adamic 0.1 refuses a number as a condition; compare it explicitly, like name.length > 0 or count !== 0"},
		{"var", "var old = 1;\n", "main.a:1:1: Adamic 0.1 refuses var; use const or let"},
		{"async", "async function wait(): Promise<void> {}\n", "main.a:1:1: Adamic 0.1 refuses an async function;"},
		{"try", "try {\n} catch {\n}\n", "main.a:1:1: Adamic 0.1 refuses try;"},
		{"throw", "function stop(): void {\n\tthrow new Error('x');\n}\n", "main.a:2:2: Adamic 0.1 refuses throw;"},
		{"==", "const same = 1 == 1;\n", "main.a:1:16: Adamic 0.1 refuses ==;"},
		{"for...in", "for (const key in { a: 1 }) {\n\tconsole.log(key);\n}\n", "main.a:1:1: Adamic 0.1 refuses for...in;"},
		{"delete", "const box: { a?: number } = { a: 1 };\ndelete box.a;\n", "main.a:2:1: Adamic 0.1 refuses delete;"},
		{"a getter", "class Box {\n\tget size(): number {\n\t\treturn 1;\n\t}\n}\n", "main.a:2:2: Adamic 0.1 refuses a getter;"},
		{"a filter that decides by truthiness", "const kept = [1, 2].filter((value) => value);\n", "main.a:1:28: Adamic 0.1 refuses a filter callback that doesn't return a boolean;"},
		{"reduce without an initial value", "const sum = [1, 2].reduce((total, value) => total + value);\n", "main.a:1:13: Adamic 0.1 refuses reduce without an initial value;"},
		{"!", "const map = new Map<string, number>();\nconst value = map.get('a')!;\n", "main.a:2:15: Adamic 0.1 refuses the non-null assertion !;"},
		{"Math.random", "const roll = Math.random();\n", "main.a:1:14: Adamic 0.1 refuses Math.random;"},
		{"Math.random, not called", "const roll = Math.random;\n", "main.a:1:14: Adamic 0.1 refuses Math.random;"},
		{"an unchecked cast", "type Pet = { readonly kind: 'Cat' } | { readonly kind: 'Dog' };\ninterface Named {\n\treadonly name: string;\n}\nconst pets: readonly Pet[] = [{ kind: 'Cat' }];\nfor (const pet of pets) {\n\tconst named = pet as unknown as Named;\n}\n", "main.a:7:16: Adamic 0.1 refuses a cast the runtime can't check;"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.Error(), probe.want) {
				t.Errorf("got %v, want a refusal ending %q", err, probe.want)
			}
		})
	}
}
