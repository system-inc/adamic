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
		{"a class inside a function", "function make(): number {\n\tclass Box {\n\t\treadonly size = 1;\n\t}\n\treturn 1;\n}\nconsole.log(`${make()}`);\n", "main.a:2:2: stage 0 can't lower a class inside a function yet"},
		{"join on an array of functions", "const steps = [(): number => 1];\nconsole.log(steps.join(','));\n", "main.a:2:13: stage 0 can't lower join on an array of objects, arrays, maps or functions yet"},
		{"concat on an array of arrays", "const grid: number[][] = [[1]];\nconst more = grid.concat([[2]]);\n", "main.a:2:14: stage 0 can't lower concat on an array of arrays yet"},
		// An arrow in a branch the checker knows is dead returns never, and has no name to point at:
		// that was a nil dereference, a crash where a not-yet belongs (found by the fuzzer, seed 102).
		{"an arrow returning never", "let flag: boolean = false;\nif (flag) {\n\tconsole.log([0].filter((item) => flag).join(','));\n}\n", "main.a:3:25: stage 0 can't lower a function returning never yet"},
		{"Array.from of an array", "const copy = Array.from([1, 2], (value) => value);\n", "main.a:1:25: stage 0 can't lower Array.from of anything but { length } yet"},
		{"Array.from without a callback", "const holes = Array.from({ length: 2 });\n", "main.a:1:15: stage 0 can't lower Array.from with other than { length } and a callback yet"},
		{"Array.from with a named callback", "const wide = true;\nconst made = Array.from({ length: 2 }, wide ? (_, index) => index : (_, index) => -index);\n", "main.a:2:40: stage 0 can't lower Array.from with a callback that isn't an arrow function written in place yet"},
		{"assigning Array.from's undefined", "const made = Array.from({ length: 2 }, (value, index) => {\n\tvalue = index;\n\treturn index;\n});\n", "main.a:2:2: stage 0 can't lower assigning to a parameter that only ever receives undefined yet"},
		{"an array of boolean | undefined", "const answers: (boolean | undefined)[] = [true, undefined];\n", "main.a:1:42: stage 0 can't lower an array of true | undefined yet"},
		{"a captured boolean | undefined", "function run(): boolean {\n\tlet seen: boolean | undefined;\n\tconst mark = (): void => {\n\t\tseen = true;\n\t};\n\tmark();\n\treturn seen ?? false;\n}\nconsole.log(`${run()}`);\n", "main.a:4:3: stage 0 can't lower a boolean | undefined variable a function value captures yet"},
		{"a class field of a union", "class Shown {\n\tvalue: string | number = 1;\n}\nconsole.log(`${new Shown() === new Shown()}`);\n", "main.a:2:2: stage 0 can't lower a field of type string | number yet"},
		{"a template of a union with an object", "function pick(flag: boolean): number | { size: number } {\n\treturn flag ? 1 : { size: 2 };\n}\nconsole.log(`${pick(true)}`);\n", "main.a:4:16: stage 0 can't lower a template interpolating a union with an object, an array, a map or a function in it yet"},
		{"?.[] on a string", "function first(word: string | undefined): string {\n\treturn word?.[0] ?? 'none';\n}\n", "main.a:2:9: stage 0 can't lower ?.[] on a string yet"},
		{"a destructured parameter beside a default", "function sum([a, b]: readonly [number, number], scale = 1): number {\n\treturn (a + b) * scale;\n}\nconsole.log(`${sum([1, 2])}`);\n", "main.a:1:1: stage 0 can't lower a destructured parameter beside a parameter with a default yet"},
		{"an array of targets seen as an array of Weak", "import type { Weak } from 'adamic';\ninterface Item {\n\treadonly name: string;\n}\nconst items: readonly Item[] = [{ name: 'a' }];\nconst seen: readonly Weak<Item>[] = items;\n", "main.a:6:37: stage 0 can't lower a readonly Item[] seen as a readonly Weak<Item>[] (one keeps something weakly that the other keeps strongly) yet"},
		{"a function value capturing what it initializes", "function run(): number {\n\tconst countdown = (from: number): number => (from <= 0 ? 0 : 1 + countdown(from - 1));\n\treturn countdown(3);\n}\nconsole.log(`${run()}`);\n", "main.a:2:8: stage 0 can't lower a function value that captures the variable its own initializer declares yet"},
		{"a tuple passed where an array goes", "function total(values: readonly number[]): number {\n\treturn values.length;\n}\nconst pair: [number, number] = [3, 4];\nconsole.log(`${total(pair)}`);\n", "main.a:5:22: stage 0 can't lower a [number, number] seen as a readonly number[] (a tuple is held as an object, not an array, so far; write it as an array where it's made, or copy it into one: [pair[0], pair[1]]) yet"},
		{"a tuple given to a function value taking an array", "const pair: [number, number] = [3, 4];\nconst measure = (values: readonly number[]): number => values.length;\nconsole.log(`${measure(pair)}`);\n", "main.a:3:24: stage 0 can't lower a [number, number] seen as a readonly number[] (a tuple is held as an object, not an array, so far; write it as an array where it's made, or copy it into one: [pair[0], pair[1]]) yet"},
		{"tuples where arrays of arrays go", "type Pair = readonly [number, number];\nconst pairs: Pair[] = [[1, 2]];\nconst rows: readonly (readonly number[])[] = pairs;\n", "main.a:3:46: stage 0 can't lower a Pair seen as a readonly number[] (a tuple is held as an object, not an array, so far; write it as an array where it's made, or copy it into one: [pair[0], pair[1]]) yet"},
		{"a callback taking arrays given tuples", "type Pair = readonly [number, number];\nconst pairs: Pair[] = [[1, 2]];\nconst measure = (values: readonly number[]): number => values.length;\nconsole.log(pairs.map(measure).join(','));\n", "main.a:4:23: stage 0 can't lower a Pair seen as a readonly number[] (a tuple is held as an object, not an array, so far; write it as an array where it's made, or copy it into one: [pair[0], pair[1]]) yet"},
		{"a function value called through ?.", "const steps: (() => void)[] = [];\nconst step = steps[0];\nstep?.();\n", "main.a:3:1: stage 0 can't lower a call through ?. (an optional call) yet"},
		{"a generic function as a value", "function same<Item>(item: Item): Item {\n\treturn item;\n}\nconst copy: (value: number) => number = same;\n", "main.a:4:41: stage 0 can't lower a generic function as a value yet"},
		{"a try around repeat", "function line(count: number): string {\n\ttry {\n\t\treturn '-'.repeat(count);\n\t} catch {\n\t\treturn '';\n\t}\n}\nconsole.log(line(3));\n", "main.a:2:2: stage 0 can't lower a try around repeat, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) yet"},
		{"a try around a call that reaches toFixed", "function shown(value: number, digits: number): string {\n\treturn value.toFixed(digits);\n}\nfunction safe(value: number): string {\n\ttry {\n\t\treturn shown(value, 2);\n\t} finally {\n\t\tconsole.log('done');\n\t}\n}\nconsole.log(safe(1));\n", "main.a:5:2: stage 0 can't lower a try around toFixed, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) yet"},
		{"rethrowing a stored Error", "const failure = new Error('stored');\nfunction stop(): void {\n\tthrow failure;\n}\nstop();\n", "main.a:3:8: stage 0 can't lower throwing an Error that isn't made where it's thrown or caught by the catch around it yet"},
		{"an object seen with a field weak in one view only", "import type { Weak } from 'adamic';\ninterface Box {\n\treadonly label: string;\n}\ninterface Strong {\n\treadonly v: Box;\n}\ninterface Weakly {\n\treadonly v: Weak<Box>;\n}\nfunction run(n: number): void {\n\tconst box: Box = { label: `b${n}` };\n\tconst s: Strong = { v: box };\n\tconst w: Weakly = s;\n\tconsole.log(`${s.v === box} ${w.v === box}`);\n}\nrun(1);\n", "main.a:14:20: stage 0 can't lower a Strong seen as a Weakly (one keeps something weakly that the other keeps strongly) yet"},
		{"a function seen with a parameter weak in one view only", "import type { Weak } from 'adamic';\ninterface Box {\n\treadonly label: string;\n}\nfunction run(n: number): void {\n\tconst box: Box = { label: `b${n}` };\n\tconst f: (x: Box) => boolean = (x: Weak<Box>): boolean => x === box;\n\tconsole.log(`${f(box)}`);\n}\nrun(1);\n", "main.a:7:33: stage 0 can't lower a (x: Weak<Box>) => boolean seen as a (x: Box) => boolean (one keeps something weakly that the other keeps strongly) yet"},
		{"a function seen with a result weak in one view only", "import type { Weak } from 'adamic';\ninterface Box {\n\treadonly label: string;\n}\nfunction run(n: number): void {\n\tconst box: Box = { label: `b${n}` };\n\tconst f: () => Weak<Box> = (): Box => box;\n\tconst got = f();\n\tconsole.log(`${got === box}`);\n}\nrun(1);\n", "main.a:7:29: stage 0 can't lower a () => Box seen as a () => Weak<Box> (one keeps something weakly that the other keeps strongly) yet"},
		{"a spread seen with a field weak in one view only", "import type { Weak } from 'adamic';\ninterface Box {\n\treadonly label: string;\n}\ninterface Strong {\n\treadonly tag: string;\n\tv: Box;\n}\ninterface Weakly {\n\treadonly tag: string;\n\tv: Weak<Box>;\n}\nfunction run(n: number): void {\n\tconst box: Box = { label: `b${n}` };\n\tconst s: Strong = { tag: 't', v: box };\n\tconst w: Weakly = { ...s, tag: 'w' };\n\tconsole.log(`${w.v === box}`);\n}\nrun(1);\n", "main.a:16:25: stage 0 can't lower a Strong seen as a Weakly (one keeps something weakly that the other keeps strongly) yet"},
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
		{"var", "var old = 1;\n", "main.a:1:1: Adamic 0.1 refuses var; use const or let"},
		{"async", "async function wait(): Promise<void> {}\n", "main.a:1:1: Adamic 0.1 refuses an async function;"},
		{"!", "const map = new Map<string, number>();\nconst value = map.get('a')!;\n", "main.a:2:15: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case"},
		{"==", "const same = 1 == 1;\n", "main.a:1:16: Adamic 0.1 refuses ==;"},
		{"delete", "const box: { a?: number } = { a: 1 };\ndelete box.a;\n", "main.a:2:1: Adamic 0.1 refuses delete;"},
		{"a constructor calling a method before its fields are set (R2's, and R's half_built)", "class Scaled {\n\treadonly doubled: number;\n\treadonly base: number;\n\tconstructor(base: number) {\n\t\tthis.doubled = this.twice();\n\t\tthis.base = base;\n\t}\n\ttwice(): number {\n\t\treturn this.base * 2;\n\t}\n}\nconsole.log(`${new Scaled(21).doubled}`);\n", "main.a:5:18: Adamic 0.1 refuses this escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise); assign every field first, then use this"},
		{"a constructor handing this out before its fields are set (R's early_this)", "function describe(box: Box): string {\n\treturn box.label;\n}\nclass Box {\n\treadonly shown: string;\n\treadonly label: string;\n\tconstructor(label: string) {\n\t\tthis.shown = describe(this);\n\t\tthis.label = label;\n\t}\n}\nconsole.log(new Box('a').shown);\n", "main.a:8:25: Adamic 0.1 refuses this escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise); assign every field first, then use this"},
		{"a constructor storing this before its fields are set", "const registry: Pair[] = [];\nclass Pair {\n\tleft: string;\n\tright: string;\n\tconstructor() {\n\t\tthis.left = `left${1}`;\n\t\tregistry.push(this);\n\t\tthis.right = `right${2}`;\n\t}\n}\nconsole.log(new Pair().left);\n", "main.a:7:17: Adamic 0.1 refuses this escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise); assign every field first, then use this"},
		{"a constructor storing this, a field set only in a branch", "const registry: Pair[] = [];\nclass Pair {\n\tleft: string;\n\tright: string;\n\tconstructor(flag: boolean) {\n\t\tthis.left = 'left';\n\t\tif (flag) {\n\t\t\tthis.right = 'yes';\n\t\t} else {\n\t\t\tthis.right = 'no';\n\t\t}\n\t\tregistry.push(this);\n\t}\n}\nconsole.log(new Pair(true).right);\n", "main.a:12:17: Adamic 0.1 refuses this escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise); assign every field first, then use this"},
		{"a filter that decides by truthiness", "const kept = [1, 2].filter((value) => value);\n", "main.a:1:28: Adamic 0.1 refuses a filter callback that doesn't return a boolean;"},
		{"Array.from's undefined typed as a number", "const made = Array.from({ length: 2 }, (value: number, index: number) => value + index);\n", "main.a:1:41: Adamic 0.1 refuses a first Array.from parameter typed number, which is undefined every time;"},
		{"a parent pointer not declared weak", "interface TreeNode {\n\treadonly name: string;\n\tparent: TreeNode | undefined;\n\treadonly children: readonly TreeNode[];\n}\nconst root: TreeNode = { name: 'root', parent: undefined, children: [] };\nroot.parent = root;\n", "main.a:3:2: Adamic 0.1 refuses TreeNode.parent, a mutable field of type TreeNode | undefined, which can reach back to the TreeNode holding it: a cycle reference counting can't free, and the write at "},
		{"a doubly linked list with a strong next", "import type { Weak } from 'adamic';\ninterface ListNode {\n\tprev: Weak<ListNode>;\n\tnext: ListNode | undefined;\n}\nconst first: ListNode = { prev: undefined, next: undefined };\nfirst.next = first;\n", "main.a:4:2: Adamic 0.1 refuses ListNode.next, a mutable field of type ListNode | undefined, which can reach back to the ListNode holding it"},
		{"a closure in a cell it captures", "function run(): number {\n\tlet countdown = (from: number): number => from;\n\tcountdown = (from: number): number => (from <= 0 ? 0 : 1 + countdown(from - 1));\n\treturn countdown(3);\n}\nconsole.log(`${run()}`);\n", "main.a:2:6: Adamic 0.1 refuses 'countdown', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing"},
		{"a mutable array of children", "interface TreeNode {\n\treadonly children: TreeNode[];\n}\nconst root: TreeNode = { children: [] };\nroot.children.push(root);\n", "main.a:2:2: Adamic 0.1 refuses TreeNode[], an array whose elements can reach back to an array like it: a cycle reference counting can't free, and the write at "},
		{"a callback that captures what holds it", "class Button {\n\tlabel = 'ok';\n\tonClick: () => string = () => '';\n}\nfunction wire(): string {\n\tconst button = new Button();\n\tbutton.onClick = () => button.label;\n\treturn button.onClick();\n}\nconsole.log(wire());\n", "main.a:3:2: Adamic 0.1 refuses Button.onClick, a mutable field of type () => string, which can reach back to the Button holding it"},
		{"a map whose values reach back", "interface Room {\n\treadonly doors: Map<string, Room>;\n}\nconst hall: Room = { doors: new Map<string, Room>() };\nhall.doors.set('self', hall);\n", "main.a:2:2: Adamic 0.1 refuses Map<string, Room>, a map whose values can reach back to a map like it: a cycle reference counting can't free, and the write at "},
		{"a map whose keys reach back", "interface Room {\n\treadonly name: string;\n\treadonly visits: Map<Room, number>;\n}\nconst visits = new Map<Room, number>();\nconst hall: Room = { name: 'hall', visits };\nvisits.set(hall, 1);\n", "main.a:3:2: Adamic 0.1 refuses Map<Room, number>, a map whose keys can reach back to a map like it"},
		{"a set whose elements reach back", "interface Step {\n\treadonly name: string;\n\treadonly seen: Set<Step>;\n}\nconst seen = new Set<Step>();\nconst start: Step = { name: 'start', seen };\nseen.add(start);\n", "main.a:3:2: Adamic 0.1 refuses Set<Step>, a set whose elements can reach back to a set like it"},
		{"a map of sets that reach back", "interface Room {\n\treadonly name: string;\n\treadonly groups: Map<string, Set<Room>>;\n}\nconst groups = new Map<string, Set<Room>>();\nconst hall: Room = { name: 'hall', groups };\ngroups.set('all', new Set([hall]));\n", "main.a:3:2: Adamic 0.1 refuses Map<string, Set<Room>>, a map whose values can reach back to a map like it"},
		{"a cycle through a subtype", "interface Animal {\n\treadonly name: string;\n}\ninterface Dog extends Animal {\n\treadonly owner: Person;\n}\nclass Person {\n\tpet: Animal | undefined = undefined;\n}\nconst person = new Person();\nconst dog: Dog = { name: 'Rex', owner: person };\nperson.pet = dog;\n", "main.a:8:2: Adamic 0.1 refuses Person.pet, a mutable field of type Animal | undefined, which can reach back to the Person holding it"},
		{"a cycle only a later instantiation closes", "interface Plain {\n\treadonly size: number;\n}\ninterface Listener {\n\treadonly onEvent: () => number;\n}\nclass Box<Item> {\n\titem: Item;\n\tconstructor(item: Item) {\n\t\tthis.item = item;\n\t}\n\tlisten(): Listener {\n\t\treturn { onEvent: () => (this.item === undefined ? 0 : 1) };\n\t}\n}\nconst plain = new Box<Plain>({ size: 1 });\nconst listening = new Box<Listener>(plain.listen());\nlistening.item = listening.listen();\n", "main.a:8:2: Adamic 0.1 refuses Box<Listener>.item, a mutable field of type Listener, which can reach back to the Box<Listener> holding it"},
		{"a cycle a generic class makes but never names", "interface Tagged {\n\treadonly id: number;\n}\nclass Box<T> {\n\treadonly id: number;\n\tv: T | undefined;\n\tconstructor(v: T) {\n\t\tthis.id = 1;\n\t\tthis.v = v;\n\t}\n}\nclass Maker<T> {\n\treadonly tag: string;\n\tconstructor(tag: string) {\n\t\tthis.tag = tag;\n\t}\n\tmake(x: T): Tagged {\n\t\treturn this.tag.length > 0 ? new Box<T>(x) : new Box<T>(x);\n\t}\n}\ninterface Node {\n\treadonly name: string;\n\tbox: Tagged;\n}\nfunction make(k: number): number {\n\tconst empty = new Maker<number>(`e`);\n\tconst n: Node = { name: `n${k}`, box: empty.make(0) };\n\tconst maker = new Maker<Node>(`m`);\n\tn.box = maker.make(n);\n\treturn n.box.id + n.name.length;\n}\nconsole.log(`${make(1)}`);\n", "main.a:23:2: Adamic 0.1 refuses Node.box, a mutable field of type Tagged, which can reach back to the Node holding it"},
		{"a generic cell holding a function that captures it", "class Tie<T> {\n\treadonly tag: string;\n\tconstructor(tag: string) {\n\t\tthis.tag = tag;\n\t}\n\ttie(initial: T, make: (get: () => T) => T): T {\n\t\tlet cell: T = initial;\n\t\tconst get = (): T => (this.tag.length > 0 ? cell : cell);\n\t\tcell = make(get);\n\t\treturn cell;\n\t}\n}\nfunction run(n: number): number {\n\tconst tie = new Tie<() => number>(`t${n}`);\n\tconst f = tie.tie(\n\t\t(): number => 0,\n\t\t(get: () => () => number): (() => number) => (): number => (n > 100 ? get()() : n),\n\t);\n\treturn f();\n}\nconsole.log(`${run(1)}`);\n", "main.a:7:7: Adamic 0.1 refuses 'cell', a variable a function value captures and can be reached from what it holds"},
		{"reduce without an initial value", "const sum = [1, 2].reduce((total, value) => total + value);\n", "main.a:1:13: Adamic 0.1 refuses reduce without an initial value;"},
		{"Math.random", "const roll = Math.random();\n", "main.a:1:14: Adamic 0.1 refuses Math.random;"},
		{"Math.random, not called", "const roll = Math.random;\n", "main.a:1:14: Adamic 0.1 refuses Math.random;"},
		{"a method read off its object", "class Counter {\n\tcount = 0;\n\tincrement(): number {\n\t\tthis.count += 1;\n\t\treturn this.count;\n\t}\n}\nconst counter = new Counter();\nconst detached: () => number = counter.increment;\n", "main.a:9:32: Adamic 0.1 refuses a method read as a value (increment would lose its object, and this with it); call it in an arrow that keeps the object: () => counter.increment() (unbound-method)"},
		{"this.method read as a value", "class Counter {\n\tcount = 0;\n\tincrement(): number {\n\t\treturn 1;\n\t}\n\tsteps(): number[] {\n\t\treturn [1, 2].map(this.increment);\n\t}\n}\nconsole.log(new Counter().steps().join(','));\n", "main.a:7:21: Adamic 0.1 refuses a method read as a value (increment would lose its object, and this with it); call it in an arrow that keeps the object: () => this.increment() (unbound-method)"},
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

// pets is the two types every invariance probe widens: a Dog is an Animal with more to read.
const pets = `import type { Weak } from 'adamic';
interface Animal {
	readonly name: string;
}
interface Dog extends Animal {
	readonly bark: string;
}
interface Tag {
	readonly tag: string;
}
`

// A mutable location is invariant in 0.1 (adamic/invariant-mutable): a value is refused where it's
// seen through a type that can write what it can't hold, at any depth, through a type parameter's
// constraint and through an intersection too. tsc accepts every one, and on Node each writes a cat
// among the dogs.
func TestAMutableLocationSeenWiderIsRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name   string
		source string
		want   string
	}{
		{"an array seen as a wider array", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const animals: Animal[] = dogs;
animals.push({ name: 'Tom' });
`, "main.a:12:27: Adamic 0.1 refuses a value of type Dog[] seen as Animal[], which can write Animal where Dog is read; make the wider type readonly"},
		{"a mutable field seen as a wider field", `interface Kennel {
	pet: Dog;
}
interface Pen {
	pet: Animal;
}
function swap(kennel: Kennel): void {
	const pen: Pen = kennel;
	pen.pet = { name: 'Tom' };
}
`, "main.a:18:19: Adamic 0.1 refuses a value of type Kennel seen as Pen, which can write Animal where Dog is read"},
		{"a map's values seen wider, passed", `function add(animals: Map<string, Animal>): void {
	animals.set('Tom', { name: 'Tom' });
}
const dogsByName = new Map<string, Dog>();
add(dogsByName);
`, "main.a:15:5: Adamic 0.1 refuses a value of type Map<string, Dog> seen as Map<string, Animal>"},
		{"a function seen as taking a narrower array", `const count: (dogs: Dog[]) => number = (animals: Animal[]): number => animals.push({ name: 'Tom' });
`, "main.a:11:40: Adamic 0.1 refuses a value of type (animals: Animal[]) => number seen as (dogs: Dog[]) => number, which can write Animal where Dog is read"},
		{"an array of arrays seen wider behind readonly", `function widen(packs: readonly Dog[][]): readonly Animal[][] {
	return packs;
}
`, "main.a:12:9: Adamic 0.1 refuses a value of type readonly Dog[][] seen as readonly Animal[][], which can write Animal where Dog is read"},
		{"a type parameter's constraint: a Narrow into a Pack slot", `function mix<Pack extends Animal[], Narrow extends Pack>(pack: Pack, narrow: Narrow): number {
	const slot: Pack = narrow;
	slot.push({ name: 'Tom' });
	return pack.length;
}
`, "main.a:12:21: Adamic 0.1 refuses a value of type Narrow seen as Pack, a type parameter whose constraint Animal[] can be written, so it can write what Narrow can't hold; take it as Narrow, or constrain Pack to something readonly"},
		{"an explicit type argument widens a mutable parameter", `function adopt<Pack extends Animal[]>(pack: Pack): number {
	pack.push({ name: 'Tom' });
	return pack.length;
}
const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
adopt<Animal[]>(dogs);
`, "main.a:16:17: Adamic 0.1 refuses a value of type Dog[] seen as Animal[], which can write Animal where Dog is read; make the wider type readonly"},
		{"an inferred type argument makes a generic write unsafe", `function adopt<Pack extends Animal[]>(pack: Pack): number {
	pack.push({ name: 'Tom' });
	return pack.length;
}
const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
adopt(dogs);
`, "main.a:12:12: Adamic 0.1 refuses instantiating a generic function makes a value of type { name: string; } written where Dog is read; make the collection readonly, or use a type parameter for the value being written"},
		{"a type parameter's constraint behind a readonly property", `interface Held<Pack> {
	readonly pack: Pack;
}
function mix<Pack extends Animal[], Narrow extends Pack>(narrow: Narrow): number {
	const held: Held<Pack> = { pack: narrow };
	held.pack.push({ name: 'Tom' });
	return held.pack.length;
}
`, "main.a:15:35: Adamic 0.1 refuses a value of type Narrow seen as Pack, a type parameter whose constraint Animal[] can be written"},
		{"a type parameter whose constraint is narrower, seen as the wider array", `function widen<Pack extends Dog[]>(pack: Pack): Animal[] {
	return pack;
}
`, "main.a:12:9: Adamic 0.1 refuses a value of type Pack seen as Animal[], which can write Animal where Dog is read"},
		{"an intersection's array seen as an intersection's", `function tagged(dogs: Dog[] & Tag): number {
	const animals: Animal[] & Tag = dogs;
	return animals.push({ name: 'Tom' });
}
`, "main.a:12:34: Adamic 0.1 refuses a value of type Dog[] & Tag seen as Animal[] & Tag, which can write Animal where Dog is read"},
		{"an array of targets seen as an array of Weak", `const items: Animal[] = [{ name: 'Rex' }];
const seen: Weak<Animal>[] = items;
`, "main.a:12:30: Adamic 0.1 refuses a value of type Animal[] seen as Weak<Animal>[], which can write Weak<Animal> where Animal is read"},
		{"an object seen with a mutable field weak in one view only", `interface Strong {
	v: Animal;
}
interface Weakly {
	v: Weak<Animal>;
}
const strong: Strong = { v: { name: 'Rex' } };
const weakly: Weakly = strong;
weakly.v = undefined;
`, "main.a:18:24: Adamic 0.1 refuses a value of type Strong seen as Weakly, which can write Weak<Animal> where Animal is read"},
		{"a readonly field turned back into a writable one", `const kennel: { pet: Dog } = { pet: { name: 'Rex', bark: 'woof' } };
const view: { readonly pet: Animal } = kennel;
const pen: { pet: Animal } = view;
pen.pet = { name: 'Tom' };
`, "main.a:13:30: Adamic 0.1 refuses a value of type { readonly pet: Animal; } seen as { pet: Animal; }, whose readonly field pet becomes writable: a readonly field may hold something narrower than Animal, which a write of Animal would replace; keep pet readonly in the type it's seen as"},
		{"a readonly field made writable by a mapped type", `type Writable<T> = { -readonly [K in keyof T]: T[K] };
interface View {
	readonly pet: Animal;
}
function open(view: View): Writable<View> {
	return view;
}
`, "main.a:16:9: Adamic 0.1 refuses a value of type View seen as Writable<View>, whose readonly field pet becomes writable"},
		{"a readonly array field turned writable", `function open(view: { readonly pets: readonly Animal[] }): { pets: readonly Animal[] } {
	return view;
}
`, "main.a:12:9: Adamic 0.1 refuses a value of type { readonly pets: readonly Animal[]; } seen as { pets: readonly Animal[]; }, whose readonly field pets becomes writable"},
		{"a readonly field turned writable inside a readonly array", `interface Held {
	readonly pet: Animal;
}
interface Open {
	pet: Animal;
}
function open(held: readonly Held[]): readonly Open[] {
	return held;
}
`, "main.a:18:9: Adamic 0.1 refuses a value of type readonly Held[] seen as readonly Open[], whose readonly field pet becomes writable"},
		{"a readonly field of a primitive turned writable", `function open(view: { readonly name: string }): { name: string } {
	return view;
}
`, "main.a:12:9: Adamic 0.1 refuses a value of type { readonly name: string; } seen as { name: string; }, whose readonly field name becomes writable"},
		{"a shorthand property", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const list = dogs;
const pen: { list: Animal[] } = { list };
`, "main.a:13:35: Adamic 0.1 refuses a value of type Dog[] seen as Animal[], which can write Animal where Dog is read"},
		{"a parameter's default", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
function adopt(animals: Animal[] = dogs): number {
	return animals.push({ name: 'Tom' });
}
`, "main.a:12:36: Adamic 0.1 refuses a value of type Dog[] seen as Animal[], which can write Animal where Dog is read"},
		{"a class field's initializer", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
class Pen {
	animals: Animal[] = dogs;
}
`, "main.a:13:22: Adamic 0.1 refuses a value of type Dog[] seen as Animal[], which can write Animal where Dog is read"},
		{"an object spread", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const kennel = { pets: dogs };
const pen: { pets: Animal[] } = { ...kennel };
`, "main.a:13:35: Adamic 0.1 refuses a value of type { pets: Dog[]"},
		{"an as", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const animals = dogs as Animal[];
`, "main.a:12:17: Adamic 0.1 refuses a value of type Dog[] seen as Animal[], which can write Animal where Dog is read"},
		{"a union target", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const slot: Animal[] | string = dogs;
`, "main.a:12:33: Adamic 0.1 refuses a value of type Dog[] seen as string | Animal[], which can write Animal where Dog is read"},
		{"a union source", `interface Cat extends Animal {
	readonly lives: number;
}
function pick(either: Dog[] | Cat[]): Animal[] {
	return either;
}
`, "main.a:15:9: Adamic 0.1 refuses a value of type Cat[] | Dog[] seen as Animal[], which can write Animal where Cat is read"},
		{"a conditional tsc reduced to the wider branch", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const animals: Animal[] = [{ name: 'Tom' }];
const either = dogs.length > 0 ? dogs : animals;
`, "main.a:13:34: Adamic 0.1 refuses a value of type Dog[] seen as Animal[], which can write Animal where Dog is read"},
		{"a literal's element tsc reduced to the wider", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const animals: Animal[] = [{ name: 'Tom' }];
const lists = [dogs, animals];
`, "main.a:13:16: Adamic 0.1 refuses a value of type Dog[] seen as Animal[], which can write Animal where Dog is read"},
		{"a return tsc reduced to the wider", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const animals: Animal[] = [{ name: 'Tom' }];
function either(flag: boolean) {
	if (flag) {
		return dogs;
	}
	return animals;
}
`, "main.a:15:10: Adamic 0.1 refuses a value of type Dog[] seen as Animal[], which can write Animal where Dog is read"},
		{"a method's return", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const source = { list: (): Dog[] => dogs };
const shelter: { list(): Animal[] } = source;
`, "main.a:13:39: Adamic 0.1 refuses a value of type { list: () => Dog[]"},
		{"a method's parameter, bivariant in tsc", `interface Handler {
	handle(animal: Animal): void;
}
const dogHandler = { handle: (dog: Dog): void => console.log(dog.bark) };
const handler: Handler = dogHandler;
`, "main.a:15:26: Adamic 0.1 refuses a function taking Dog seen as one taking Animal (tsc relates a method's parameters both ways), so it can be handed what it can't take"},
		{"a tuple's later element", `const pair: [Animal, Dog] = [{ name: 'Tom' }, { name: 'Rex', bark: 'woof' }];
const list: Animal[] = pair;
`, "main.a:12:24: Adamic 0.1 refuses a value of type [Animal, Dog] seen as Animal[], which can write Animal where Dog is read"},
		{"a class target", `class Box<T> {
	item: T;
	constructor(item: T) {
		this.item = item;
	}
}
const dogBox = new Box<Dog>({ name: 'Rex', bark: 'woof' });
const animalBox: Box<Animal> = dogBox;
`, "main.a:18:32: Adamic 0.1 refuses a value of type Box<Dog> seen as Box<Animal>, which can write Animal where Dog is read"},
		{"a plain object seen as a class", `class Pen {
	pet: Animal = { name: 'Tom' };
}
const kennel = { pet: { name: 'Rex', bark: 'woof' } };
const pen: Pen = kennel;
`, "main.a:15:18: Adamic 0.1 refuses a value of type { pet: { name: string"},
		{"a fresh copy's elements seen wider", `interface Kennel {
	pet: Dog;
}
const kennels: Kennel[] = [{ pet: { name: 'Rex', bark: 'woof' } }];
const pens: { pet: Animal }[] = kennels.slice();
`, "main.a:15:33: Adamic 0.1 refuses a value of type Kennel[] seen as { pet: Animal"},
		{"an annotated destructuring", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const kennel = { pets: dogs };
const { pets }: { pets: Animal[] } = kennel;
`, "main.a:13:38: Adamic 0.1 refuses a value of type { pets: Dog[]"},
		{"an intersection's array seen as a plain array", `function tagged(dogs: Dog[] & Tag): number {
	const animals: Animal[] = dogs;
	return animals.push({ name: 'Tom' });
}
`, "main.a:12:28: Adamic 0.1 refuses a value of type Dog[] & Tag seen as Animal[], which can write Animal where Dog is read"},
		{"a destructuring assignment's element seen wider", `function pack(dogs: Dog[]): readonly [Dog[], number] {
	return [dogs, dogs.length];
}
let animals: Animal[] = [];
let count = 0;
[animals, count] = pack([{ name: 'Rex', bark: 'woof' }]);
`, "main.a:16:2: Adamic 0.1 refuses a value of type Dog[] seen as Animal[], which can write Animal where Dog is read"},
		{"a destructuring assignment's element seen wider inside a readonly array", `function pack(packs: Dog[][]): readonly [Dog[][], number] {
	return [packs, packs.length];
}
let packs: readonly Animal[][] = [];
let count = 0;
[packs, count] = pack([[{ name: 'Rex', bark: 'woof' }]]);
`, "main.a:16:2: Adamic 0.1 refuses a value of type Dog[][] seen as readonly Animal[][], which can write Animal where Dog is read"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, pets+probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.Error(), probe.want) {
				t.Errorf("got %v, want a refusal starting %q", err, probe.want)
			}
		})
	}
}

// Each refused shape's sound neighbor isn't refused: seen through something that can't write, or
// only as itself. Stage 0 may not lower some of them yet (a value of a type parameter's type, an
// intersection with an array), and that's a different promise from a refusal.
func TestAViewThatCantWriteIsNotRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name   string
		source string
	}{
		{"a destructuring declaration of a readonly field", `interface Badge {
	readonly label: string;
	readonly count: number;
}
const badge: Badge = { label: 'gold', count: 3 };
const { label, count: total } = badge;
console.log(` + "`${label} ${total}`" + `);
`},
		{"a destructuring assignment's elements going into wider names", `function swap(pair: readonly [string, number]): readonly [number, string] {
	return [pair[1], pair[0]];
}
let either: string | number = 0;
let maybeCount: number | undefined = undefined;
[maybeCount, either] = swap(['twelve', 12]);
console.log(` + "`${maybeCount ?? -1} ${either}`" + `);
`},
		{"an array seen as a readonly wider array", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const animals: readonly Animal[] = dogs;
console.log(` + "`${animals.length}`" + `);
`},
		{"a readonly field seen wider", `interface Kennel {
	readonly pet: Dog;
}
interface Pen {
	readonly pet: Animal;
}
const kennel: Kennel = { pet: { name: 'Rex', bark: 'woof' } };
const pen: Pen = kennel;
console.log(pen.pet.name);
`},
		{"a readonly constraint: a Narrow into a Pack slot", `function mix<Pack extends readonly Animal[], Narrow extends Pack>(pack: Pack, narrow: Narrow): number {
	const slot: Pack = narrow;
	return pack.length + slot.length;
}
`},
		{"a type parameter as itself", `function keep<Pack extends Animal[]>(pack: Pack): Pack {
	const slot: Pack = pack;
	return slot;
}
`},
		{"a type parameter narrowed from undefined, as itself", `function pick<Pack extends Animal[] | undefined>(pack: Pack): number {
	if (pack !== undefined) {
		const slot: Pack = pack;
		return slot === undefined ? 0 : slot.length;
	}
	return 0;
}
`},
		{"an intersection's array seen as a readonly array with the same tag", `function tagged(dogs: Dog[] & Tag): number {
	const animals: readonly Animal[] & Tag = dogs;
	return animals.length;
}
`},
		{"a readonly field seen as a readonly field", `const view: { readonly pet: Animal } = { pet: { name: 'Rex' } };
const same: { readonly pet: Animal } = view;
console.log(same.pet.name);
`},
		{"a writable field seen as a readonly one", `const pen: { pet: Animal } = { pet: { name: 'Rex' } };
const view: { readonly pet: Animal } = pen;
console.log(view.pet.name);
`},
		{"a readonly field seen as its own type", `interface View {
	readonly pet: Animal;
}
function keep(view: View): View {
	const same: View = view;
	return same;
}
`},
		{"a readonly field copied into a writable one", `const view: { readonly pet: Animal } = { pet: { name: 'Rex' } };
const pen: { pet: Animal } = { pet: view.pet };
pen.pet = { name: 'Tom' };
console.log(view.pet.name);
`},
		{"a copy made by slice", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const animals: Animal[] = dogs.slice();
animals.push({ name: 'Tom' });
`},
		{"a copy made by map", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const animals: Animal[] = dogs.map((dog) => dog);
`},
		{"a copy made by filter", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const animals: Animal[] = dogs.filter((dog) => dog.bark.length > 0);
`},
		{"a conditional of fresh arrays", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const animals: Animal[] = dogs.length > 5 ? [{ name: 'Tom' }] : [];
`},
		{"a Map made from pairs", `const pairs: [string, Dog][] = [['Rex', { name: 'Rex', bark: 'woof' }]];
const byName = new Map<string, Dog>(pairs);
console.log(` + "`${byName.size}`" + `);
`},
		{"a conditional of fresh copies", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const animals: Animal[] = dogs.length > 5 ? dogs.slice() : [];
animals.push({ name: 'Tom' });
`},
		{"a new Map passed", `function add(animals: Map<string, Animal>): void {
	animals.set('Tom', { name: 'Tom' });
}
add(new Map<string, Dog>());
`},
		{"a new Map", `const byName: Map<string, Animal> = new Map<string, Dog>();
`},
		{"a union target that can't write", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const slot: readonly Animal[] | string = dogs;
`},
		{"a shorthand property that can't write", `const dogs: Dog[] = [{ name: 'Rex', bark: 'woof' }];
const list = dogs;
const view: { list: readonly Animal[] } = { list };
`},
		{"a destructuring with no type", `interface View {
	readonly pet: Animal;
}
const view: View = { pet: { name: 'Rex' } };
const { pet } = view;
console.log(pet.name);
`},
		{"a method returning the same type", `const source = { list: (): readonly Animal[] => [] };
const shelter: { list(): readonly Animal[] } = source;
`},
		{"a class seen as itself", `class Box<T> {
	item: T;
	constructor(item: T) {
		this.item = item;
	}
}
const dogBox = new Box<Dog>({ name: 'Rex', bark: 'woof' });
const again: Box<Dog> = dogBox;
console.log(again.item.bark);
`},
		{"an intersection seen as itself", `function tagged(dogs: Dog[] & Tag): number {
	const same: Dog[] & Tag = dogs;
	return same.length;
}
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, pets+probe.source)
			var refused *Refused
			if errors.As(err, &refused) {
				t.Errorf("refused a view that can't write: %v", err)
			}
		})
	}
}

// A tuple is held as an object, so stage 0 can't see one as an array yet: at every place a tuple can
// flow into an array slot, at any depth, it says so with the way around it, and never lowers the
// object as an array (R's c3.a printed a pointer as the array's length).
func TestATupleSeenAsAnArrayIsNotYet(t *testing.T) {
	t.Parallel()
	const seen = "seen as a readonly number[] (a tuple is held as an object, not an array, so far; write it as an array where it's made, or copy it into one: [pair[0], pair[1]]) yet"
	for _, probe := range []struct {
		name   string
		source string
		want   string
	}{
		{"an initializer", "const pair: [number, number] = [3, 4];\nconst values: readonly number[] = pair;\n", "main.a:2:35: stage 0 can't lower a [number, number] " + seen},
		{"a cast literal", "const values: readonly number[] = [3, 4] as [number, number];\n", "main.a:1:35: stage 0 can't lower a [number, number] " + seen},
		{"an assignment", "const pair: [number, number] = [3, 4];\nlet values: readonly number[] = [];\nvalues = pair;\n", "main.a:3:10: stage 0 can't lower a [number, number] " + seen},
		{"an argument", "function total(values: readonly number[]): number {\n\treturn values.length;\n}\nconst pair: [number, number] = [3, 4];\nconsole.log(`${total(pair)}`);\n", "main.a:5:22: stage 0 can't lower a [number, number] " + seen},
		{"a return", "function widen(pair: [number, number]): readonly number[] {\n\treturn pair;\n}\n", "main.a:2:9: stage 0 can't lower a [number, number] " + seen},
		{"a field", "const pair: [number, number] = [3, 4];\nconst held: { readonly values: readonly number[] } = { values: pair };\n", "main.a:2:64: stage 0 can't lower a [number, number] " + seen},
		{"an element", "const pair: [number, number] = [3, 4];\nconst rows: (readonly number[])[] = [pair];\n", "main.a:2:38: stage 0 can't lower a [number, number] " + seen},
		{"an arrow's body", "const make: () => readonly number[] = () => [3, 4] as [number, number];\n", "main.a:1:45: stage 0 can't lower a [number, number] " + seen},
		{"a map's value", "const pair: [number, number] = [3, 4];\nconst rows = new Map<string, readonly number[]>();\nrows.set('a', pair);\n", "main.a:3:15: stage 0 can't lower a [number, number] " + seen},
		{"a slot that may be undefined", "const pair: [number, number] = [3, 4];\nconst values: readonly number[] | undefined = pair;\n", "main.a:2:47: stage 0 can't lower a [number, number] " + seen},
		{"an array of tuples seen as an array of arrays", "const pairs: [number, number][] = [[3, 4]];\nconst rows: readonly (readonly number[])[] = pairs;\n", "main.a:2:46: stage 0 can't lower a [number, number] " + seen},
		{"a tuple field seen as an array field", "const held: { readonly values: [number, number] } = { values: [3, 4] };\nconst seen: { readonly values: readonly number[] } = held;\n", "main.a:2:54: stage 0 can't lower a [number, number] " + seen},
		{"a function returning a tuple seen as returning an array", "const make = (): [number, number] => [3, 4];\nconst widened: () => readonly number[] = make;\n", "main.a:2:42: stage 0 can't lower a [number, number] " + seen},
		{"an array method called on a tuple", "const pair: [string, string] = ['a', 'b'];\nconsole.log(pair.join('-'));\n", "main.a:2:13: stage 0 can't lower join on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), probe.want) {
				t.Errorf("got %v, want a not-yet ending %q", err, probe.want)
			}
		})
	}
}

// shelter is a class with a method, for the method-value probes.
const shelter = `class Shelter {
	readonly name: string;
	constructor(name: string) {
		this.name = name;
	}
	admit(pet: string): string {
		return ` + "`${this.name} took ${pet}`" + `;
	}
}
const shelter = new Shelter('Haven');
`

// A method read as a value loses its object, and this is undefined when it's called (R's
// method_value.a panicked "compiler bug" natively). 0.1 refuses it, with the arrow that keeps it.
func TestAMethodReadAsAValueIsRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name   string
		source string
		want   string
	}{
		{"held in a variable", "const admit = shelter.admit;\n", "main.a:11:15: Adamic 0.1 refuses a method read as a value (admit would lose its object, and this with it); call it in an arrow that keeps the object: (pet) => shelter.admit(pet) (unbound-method)"},
		{"passed as a callback", "console.log(['Rex'].map(shelter.admit).join());\n", "main.a:11:25: Adamic 0.1 refuses a method read as a value (admit would lose its object"},
		{"in parentheses", "const admit = (shelter.admit);\n", "main.a:11:16: Adamic 0.1 refuses a method read as a value (admit would lose its object"},
		{"read through this", "class Desk {\n\treadonly greeting: string = 'hi';\n\tgreet(): string {\n\t\treturn this.greeting;\n\t}\n\tlater(): () => string {\n\t\treturn this.greet;\n\t}\n}\n", "main.a:17:10: Adamic 0.1 refuses a method read as a value (greet would lose its object, and this with it); call it in an arrow that keeps the object: () => this.greet()"},
		{"an interface's method", "interface Greeter {\n\tgreet(name: string): string;\n}\nfunction detach(greeter: Greeter): (name: string) => string {\n\treturn greeter.greet;\n}\n", "main.a:15:9: Adamic 0.1 refuses a method read as a value (greet would lose its object"},
		{"an array's method", "const names: string[] = [];\nconst add = names.push;\n", "main.a:12:13: Adamic 0.1 refuses a method read as a value (push would lose its object"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, shelter+probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.Error(), probe.want) {
				t.Errorf("got %v, want a refusal containing %q", err, probe.want)
			}
		})
	}
	for _, neighbor := range []struct {
		name   string
		source string
	}{
		{"called on its object", "console.log(shelter.admit('Rex'));\n"},
		{"called through parentheses", "console.log((shelter.admit)('Rex'));\n"},
		{"called in an arrow", "console.log(['Rex'].map((pet) => shelter.admit(pet)).join());\n"},
		{"a field holding a function", "const holder: { readonly admit: (pet: string) => string } = { admit: (pet) => shelter.admit(pet) };\nconst admit = holder.admit;\nconsole.log(admit('Rex'));\n"},
	} {
		t.Run("not "+neighbor.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, shelter+neighbor.source)
			var refused *Refused
			if errors.As(err, &refused) {
				t.Errorf("refused a method that keeps its object: %v", err)
			}
		})
	}
}
