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
		{"?. on a string", "const words = ['a'];\nconsole.log(`${words[0]?.length ?? 0}`);\n", "main.a:2:16: stage 0 can't lower optional chaining on a string yet"},
		{"concat on an array of arrays", "const grid: number[][] = [[1]];\nconst more = grid.concat([[2]]);\n", "main.a:2:14: stage 0 can't lower concat on an array of arrays yet"},
		{"a rest parameter", "function sum(...values: number[]): number {\n\treturn values.length;\n}\nconsole.log(`${sum(1, 2)}`);\n", "main.a:1:14: stage 0 can't lower a parameter that isn't a plain name yet"},
		{"Array.from of an array", "const copy = Array.from([1, 2], (value) => value);\n", "main.a:1:25: stage 0 can't lower Array.from of anything but { length } yet"},
		{"Array.from without a callback", "const holes = Array.from({ length: 2 });\n", "main.a:1:15: stage 0 can't lower Array.from with other than { length } and a callback yet"},
		{"Array.from with a named callback", "const wide = true;\nconst made = Array.from({ length: 2 }, wide ? (_, index) => index : (_, index) => -index);\n", "main.a:2:40: stage 0 can't lower Array.from with a callback that isn't an arrow function written in place yet"},
		{"assigning Array.from's undefined", "const made = Array.from({ length: 2 }, (value, index) => {\n\tvalue = index;\n\treturn index;\n});\n", "main.a:2:2: stage 0 can't lower assigning to a parameter that only ever receives undefined yet"},
		{"an array of boolean | undefined", "const answers: (boolean | undefined)[] = [true, undefined];\n", "main.a:1:42: stage 0 can't lower an array of true | undefined yet"},
		{"a field of type boolean | undefined", "interface Answer {\n\tyes: boolean | undefined;\n}\nconst answer: Answer = { yes: true };\nconsole.log(`${answer.yes ?? false}`);\n", "main.a:5:16: stage 0 can't lower a field of type boolean | undefined yet"},
		{"a function value returning boolean | undefined", "const pick = (): boolean | undefined => true;\n", "main.a:1:14: stage 0 can't lower a function value returning boolean | undefined yet"},
		{"a captured boolean | undefined", "function run(): boolean {\n\tlet seen: boolean | undefined;\n\tconst mark = (): void => {\n\t\tseen = true;\n\t};\n\tmark();\n\treturn seen ?? false;\n}\nconsole.log(`${run()}`);\n", "main.a:4:3: stage 0 can't lower a boolean | undefined variable a function value captures yet"},
		{"an array of a union", "const mixed: (string | number)[] = ['a', 1];\n", "main.a:1:36: stage 0 can't lower an array of string | number yet"},
		{"a field of a union", "interface Shown {\n\treadonly value: string | number;\n}\nconst shown: Shown = { value: 1 };\nconsole.log(`${shown.value}`);\n", "main.a:5:16: stage 0 can't lower a field of type string | number yet"},
		{"a union stored in a field", "interface Shown {\n\tvalue: string | number;\n}\nconst shown: Shown = { value: 1 };\nshown.value = 'one';\n", "main.a:5:1: stage 0 can't lower storing string | number in a field yet"},
		{"a class field of a union", "class Shown {\n\tvalue: string | number = 1;\n}\nconsole.log(`${new Shown() === new Shown()}`);\n", "main.a:2:2: stage 0 can't lower a field of type string | number yet"},
		{"a union a function value takes", "const show = (value: string | number): string => `${value}`;\n", "main.a:1:15: stage 0 can't lower a function value taking string | number yet"},
		{"a map of a union", "const values = new Map<string, string | number>();\n", "main.a:1:16: stage 0 can't lower a Map of string | number yet"},
		{"a template of a union with an object", "function pick(flag: boolean): number | { size: number } {\n\treturn flag ? 1 : { size: 2 };\n}\nconsole.log(`${pick(true)}`);\n", "main.a:4:16: stage 0 can't lower a template interpolating a union with an object, an array, a map or a function in it yet"},
		{"a set of objects", "const shapes = new Set<{ size: number }>();\n", "main.a:1:16: stage 0 can't lower a Set of { size: number; } (a Set holds strings or numbers so far) yet"},
		{"a set's entries", "const seen = new Set(['a']);\nfor (const pair of seen.entries()) {\n}\n", "main.a:2:20: stage 0 can't lower for...of over a Set's entries ([element, element] pairs) yet"},
		{"a set from a string", "const letters = new Set('abc');\n", "main.a:1:25: stage 0 can't lower new Set from a string (an array is what it takes so far) yet"},
		{"a template interpolating an object", "const point = { x: 1 };\nconsole.log(`${point}`);\n", "main.a:2:16: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet"},
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
