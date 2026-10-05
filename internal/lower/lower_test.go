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
		{"Array.from's undefined as boolean | undefined", "const made = Array.from({ length: 2 }, (value: boolean | undefined, index: number) => index);\n", "main.a:1:41: stage 0 can't lower a function value taking boolean | undefined yet"},
		{"Array.from's undefined as a union", "const made = Array.from({ length: 2 }, (value: string | number | undefined, index: number) => index);\n", "main.a:1:41: stage 0 can't lower a function value taking string | number | undefined yet"},
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
		{"a set of number | undefined", "const maybes = new Set<number | undefined>();\n", "main.a:1:16: stage 0 can't lower a Set of number | undefined (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) yet"},
		{"a set's entries spread", "const seen = new Set(['a']);\nconst pairs = [...seen.entries()];\n", "main.a:2:19: stage 0 can't lower a Set's entries ([element, element] pairs) outside a for...of yet"},
		{"?.[] on a string", "function first(word: string | undefined): string {\n\treturn word?.[0] ?? 'none';\n}\n", "main.a:2:9: stage 0 can't lower ?.[] on a string yet"},
		{"a destructured parameter beside a default", "function sum([a, b]: readonly [number, number], scale = 1): number {\n\treturn (a + b) * scale;\n}\nconsole.log(`${sum([1, 2])}`);\n", "main.a:1:1: stage 0 can't lower a destructured parameter beside a parameter with a default yet"},
		{"an array of targets seen as an array of Weak", "import type { Weak } from 'adamic';\ninterface Item {\n\treadonly name: string;\n}\nconst items: Item[] = [{ name: 'a' }];\nconst seen: Weak<Item>[] = items;\n", "main.a:6:28: stage 0 can't lower an array or map of Item[] seen as one of Weak<Item>[] (one keeps its elements weakly, the other doesn't) yet"},
		{"a function value capturing what it initializes", "function run(): number {\n\tconst countdown = (from: number): number => (from <= 0 ? 0 : 1 + countdown(from - 1));\n\treturn countdown(3);\n}\nconsole.log(`${run()}`);\n", "main.a:2:8: stage 0 can't lower a function value that captures the variable its own initializer declares yet"},
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
		{"Array.from's undefined typed as a number", "const made = Array.from({ length: 2 }, (value: number, index: number) => value + index);\n", "main.a:1:41: Adamic 0.1 refuses a first Array.from parameter typed number, which is undefined every time;"},
		{"a parent pointer not declared weak", "interface TreeNode {\n\treadonly name: string;\n\tparent: TreeNode | undefined;\n\treadonly children: readonly TreeNode[];\n}\nconst root: TreeNode = { name: 'root', parent: undefined, children: [] };\n", "main.a:3:2: Adamic 0.1 refuses TreeNode.parent, a mutable field of type TreeNode | undefined, which can reach back to the TreeNode holding it: a cycle reference counting can't free; declare it parent: Weak<TreeNode> (import type { Weak } from 'adamic')"},
		{"a doubly linked list with a strong next", "import type { Weak } from 'adamic';\ninterface ListNode {\n\tprev: Weak<ListNode>;\n\tnext: ListNode | undefined;\n}\nconst first: ListNode = { prev: undefined, next: undefined };\n", "main.a:4:2: Adamic 0.1 refuses ListNode.next, a mutable field of type ListNode | undefined, which can reach back to the ListNode holding it"},
		{"a closure in a cell it captures", "function run(): number {\n\tlet countdown = (from: number): number => from;\n\tcountdown = (from: number): number => (from <= 0 ? 0 : 1 + countdown(from - 1));\n\treturn countdown(3);\n}\nconsole.log(`${run()}`);\n", "main.a:2:6: Adamic 0.1 refuses 'countdown', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; write the function as a function declaration (function countdown() {})"},
		{"a mutable array of children", "interface TreeNode {\n\treadonly children: TreeNode[];\n}\nconst root: TreeNode = { children: [] };\nroot.children.push({ children: [] });\n", "main.a:2:2: Adamic 0.1 refuses TreeNode[], an array whose elements can reach back to an array like it: a cycle reference counting can't free; declare the elements weak, Weak<TreeNode>[]"},
		{"a callback that captures what holds it", "class Button {\n\tlabel = 'ok';\n\tonClick: () => string = () => '';\n}\nfunction wire(): string {\n\tconst button = new Button();\n\tbutton.onClick = () => button.label;\n\treturn button.onClick();\n}\nconsole.log(wire());\n", "main.a:3:2: Adamic 0.1 refuses Button.onClick, a mutable field of type () => string, which can reach back to the Button holding it"},
		{"a map whose values reach back", "interface Room {\n\treadonly doors: Map<string, Room>;\n}\nconst hall: Room = { doors: new Map<string, Room>() };\nhall.doors.set('self', hall);\n", "main.a:2:2: Adamic 0.1 refuses Map<string, Room>, a map whose values can reach back to a map like it: a cycle reference counting can't free; declare the values weak, Map<string, Weak<Room>>"},
		{"a map whose keys reach back", "interface Room {\n\treadonly name: string;\n\treadonly visits: Map<Room, number>;\n}\nconst visits = new Map<Room, number>();\nconst hall: Room = { name: 'hall', visits };\nvisits.set(hall, 1);\n", "main.a:3:2: Adamic 0.1 refuses Map<Room, number>, a map whose keys can reach back to a map like it"},
		{"a set whose elements reach back", "interface Step {\n\treadonly name: string;\n\treadonly seen: Set<Step>;\n}\nconst seen = new Set<Step>();\nconst start: Step = { name: 'start', seen };\nseen.add(start);\n", "main.a:3:2: Adamic 0.1 refuses Set<Step>, a set whose elements can reach back to a set like it"},
		{"a map of sets that reach back", "interface Room {\n\treadonly name: string;\n\treadonly groups: Map<string, Set<Room>>;\n}\nconst groups = new Map<string, Set<Room>>();\nconst hall: Room = { name: 'hall', groups };\ngroups.set('all', new Set([hall]));\n", "main.a:3:2: Adamic 0.1 refuses Map<string, Set<Room>>, a map whose values can reach back to a map like it"},
		{"a cycle through a subtype", "interface Animal {\n\treadonly name: string;\n}\ninterface Dog extends Animal {\n\treadonly owner: Person;\n}\nclass Person {\n\tpet: Animal | undefined = undefined;\n}\nconst person = new Person();\nconst dog: Dog = { name: 'Rex', owner: person };\nperson.pet = dog;\n", "main.a:8:2: Adamic 0.1 refuses Person.pet, a mutable field of type Animal | undefined, which can reach back to the Person holding it"},
		{"a cycle only a later instantiation closes", "interface Plain {\n\treadonly size: number;\n}\ninterface Listener {\n\treadonly onEvent: () => number;\n}\nclass Box<Item> {\n\titem: Item;\n\tconstructor(item: Item) {\n\t\tthis.item = item;\n\t}\n\tlisten(): Listener {\n\t\treturn { onEvent: () => (this.item === undefined ? 0 : 1) };\n\t}\n}\nconst plain = new Box<Plain>({ size: 1 });\nconst listening = new Box<Listener>(plain.listen());\nlistening.item = listening.listen();\n", "main.a:8:2: Adamic 0.1 refuses Box<Listener>.item, a mutable field of type Listener, which can reach back to the Box<Listener> holding it"},
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
