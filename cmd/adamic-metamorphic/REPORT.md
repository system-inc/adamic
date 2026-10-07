# Metamorphic variants of the oracle's fixtures

The hand-written fixtures TestNativeAgreesWithNode runs reach far more of the emitter than the
randomized generator does. `adamic-metamorphic` rewrites those trusted programs so they mean the same
thing by another road, keeps a variant only when Node prints exactly what it printed for the original
(stdout, stderr and exit code, byte for byte), and then holds a compiler to Node on the variant the way
the oracle holds it on the fixture. Measured on the developer-tools tip 2adf65c2 (the base of this
branch), on a 16-core Mac with a load average between 38 and 77, every run at parallel 6, Node
v24.14.1, Apple clang 21.

## How it works

`internal/metamorphic` makes the variants; `cmd/adamic-metamorphic` drives them.

Every edit is made at the boundaries of nodes typescript-go's parser found, with the checker naming
what each identifier refers to (the same parser and checker adamic-reduce shrinks through), never by a
pattern over the text. A transform proposes sites; a site is kept only when the program with it, and
with every site kept before it, still passes the checker and stage 0 lowers it (`load.LoadOverlay`
and `lower.Lower` in process). A refused site is counted and dropped, never a finding. At most 24 sites
are tried (chosen by a stride when there are more) and 12 kept per transform per program.

1. **wrap**: the module's statements move into `function metamorphicMain(): void` called once at the
   end, so its variables become locals and its functions closures. Imports, types, classes and enums
   stay at the top level, with whatever they reach by name, to a fixed point, because stage 0 lowers no
   class inside a function. Stage 0 also lowers no function declaration inside a function, so moved
   functions become `const name = function name(...) {...}` at the top of the body, in their order
   (hoisted, as the declarations were). If stage 0 refuses that, arrows are tried, then functions stay
   at the top level with every variable they reach. A module that exports is skipped.
2. **alias**: a const read twice or more after its statement gets `const nameAlias = name;` right
   after the statement, and every read from the second on reads the alias.
3. **method**: a run of statements moves into `class MetamorphicBlockN { run(...): void { ... } }`
   at the top level, called once where the run was. What the run reads from an enclosing function
   arrives as parameters of the same name, typed by the checker at the first read, so the text is
   unchanged; only constants can be parameters, so a reassigned variable must be declared inside the
   run. A run can't return, break or continue out, use this or super, or declare what's read after it.
4. **identity**: values handed on (initializers, returned values, arguments, operands, elements,
   property values, template parts) go through `metamorphicIdentity<T>(value: T): T`, and every other
   site through a closure called on the spot, `(() => value)()`.
5. **temporaries**: a statement's own expression has its computed operands (a binary operator's sides,
   not `&&`, `||` or `??`; a call's arguments when the callee reads only names; a template's parts)
   moved into consts declared before the statement, in evaluation order; and a const with no annotation
   read once, in the next statement, at a place evaluated exactly once, goes back in parentheses.

**combined** composes them: wrap, method, alias, temporaries, identity.

Each kept variant, and each original, then runs through a checkout's own command line as the oracle
builds it: `adamic js` (the checker's and stage 0's refusals come from here, sorted by fuzz's
`Refusal`), `adamic build --sanitize`, `adamic build` and `adamic build --count`, so the checkout's own
runtime and flags are what's tested. The judge is fuzz's (`fuzz.Compare`, the judge split from its
leak run so a caller can run its platform's leak check): Node against native under ASan and UBSan and
against the JavaScript backend, for a crash, an inserted check or a disagreement. Then the release
build against the sanitized one, and for a program Node finishes with exit 0, the oracle's leak check
(on macOS the counted build's counts, then `leaks --atExit`; on Linux LeakSanitizer). The judge calls a
program printing over a megabyte unfit, a rule for generated programs; three fixtures print that much
(bitwise_sweep, large_output, sweeps/regexp_methods), so their stdout is judged by length and SHA-256.

```
adamic-metamorphic generate -transforms wrap -fixtures 20 -import-free -o variants
adamic-metamorphic generate -transforms wrap,alias,method,identity,temporaries,combined -o variants
adamic-metamorphic run -root <checkout> -variants variants -o results.jsonl
```

## First measurement: wrap, 20 import-free fixtures

The 20 are the fixtures at index i * 250 / 20 (every 12th or 13th) of the 250 import-free fixtures in
the oracle's sorted list.
The first version moved functions in as declarations: 11 kept, 9 refused by stage 0 ("a function
inside a function (a closure)"). The second moved them in as function expressions: 17 kept, 3 refused.
The final version falls back to keeping functions at the top level: 20 kept. All 20 originals and all
20 variants agree on the tip.

| fixture | try 1: functions as declarations | try 2: function expressions | final | functions moved as | original | variant | generate s | run s |
|---|---|---|---|---|---|---|---|---|
| dedication/dedication.a | kept | kept | kept | function expressions | agreed | agreed | 0.09 | 1.34 |
| internal/oracle/testdata/array_from.a | refused: a function inside a function | refused: not yet: reading order | kept | kept at the top level (expressions and arrows refused) | agreed | agreed | 0.14 | 1.45 |
| internal/oracle/testdata/borrow_element_virtual_store.a | refused: a function inside a function | kept | kept | function expressions | agreed | agreed | 0.09 | 1.31 |
| internal/oracle/testdata/class_oct6_deep.a | refused: a function inside a function | kept | kept | function expressions | agreed | agreed | 0.09 | 1.32 |
| internal/oracle/testdata/effects.a | refused: a function inside a function | refused: not yet: assigning to an Identifier | kept | kept at the top level (expressions and arrows refused) | agreed | agreed | 0.13 | 1.44 |
| internal/oracle/testdata/functions.a | refused: a function inside a function | kept | kept | function expressions | agreed | agreed | 0.10 | 1.33 |
| internal/oracle/testdata/json_stringify_keys.a | kept | kept | kept | function expressions | agreed | agreed | 0.10 | 1.36 |
| internal/oracle/testdata/library_array_flat.a | kept | kept | kept | function expressions | agreed | agreed | 0.10 | 1.36 |
| internal/oracle/testdata/library_for_in_live.a | kept | kept | kept | function expressions | agreed | agreed | 0.10 | 1.48 |
| internal/oracle/testdata/library_map_set_visit.a | kept | kept | kept | function expressions | agreed | agreed | 0.11 | 1.40 |
| internal/oracle/testdata/library_object_keys.a | kept | kept | kept | function expressions | agreed | agreed | 0.10 | 1.34 |
| internal/oracle/testdata/long_literals.a | kept | kept | kept | function expressions | agreed | agreed | 0.10 | 1.52 |
| internal/oracle/testdata/narrowed_methods.a | refused: a function inside a function | refused: not yet: assigning to an Identifier | kept | kept at the top level (expressions and arrows refused) | agreed | agreed | 0.15 | 0.69 |
| internal/oracle/testdata/objects.a | refused: a function inside a function | kept | kept | function expressions | agreed | agreed | 0.09 | 1.44 |
| internal/oracle/testdata/regexp_match.a | refused: a function inside a function | kept | kept | function expressions | agreed | agreed | 0.09 | 1.45 |
| internal/oracle/testdata/replace_all_large.a | kept | kept | kept | function expressions | agreed | agreed | 0.11 | 1.56 |
| internal/oracle/testdata/search_halves.a | kept | kept | kept | function expressions | agreed | agreed | 0.10 | 1.34 |
| internal/oracle/testdata/splice_empty.a | kept | kept | kept | function expressions | agreed | agreed | 0.10 | 1.34 |
| internal/oracle/testdata/string_too_long.a | kept | kept | kept | function expressions | agreed | agreed | 0.09 | 0.67 |
| internal/oracle/testdata/tuples_kept.a | refused: a function inside a function | kept | kept | function expressions | agreed | agreed | 0.10 | 1.34 |

## All fixtures: each transform, then combined (tip 2adf65c2)

277 fixtures (the oracle's list of programs stage 0 lowers). Generation is in process (checker and
stage 0 per tried site, then Node once for the original and once per variant); run is every way the
oracle runs a program. "Sites refused" counts tries the checker or stage 0 refused; "node differs" is a
variant Node printed differently for, discarded.

| transform | fixtures | kept | skipped | node differs | sites kept | sites refused | generate s/fixture | run s/variant | agreed | checked | unfit | not yet | invalid | finding | crash |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| original | 277 | 277 | 0 | 0 | 0 | 0 | 0.06 | 1.96 | 269 | 7 | 0 | 0 | 0 | 1 | 0 |
| wrap | 277 | 275 | 1 | 1 | 276 | 155 | 0.11 | 1.46 | 266 | 7 | 0 | 0 | 0 | 2 | 0 |
| alias | 277 | 194 | 83 | 0 | 761 | 57 | 0.12 | 1.46 | 188 | 5 | 0 | 0 | 0 | 1 | 0 |
| method | 277 | 277 | 0 | 0 | 707 | 35 | 0.14 | 1.96 | 268 | 7 | 0 | 0 | 0 | 2 | 0 |
| identity | 277 | 273 | 4 | 0 | 2485 | 89 | 0.26 | 1.42 | 261 | 7 | 0 | 0 | 0 | 5 | 0 |
| temporaries | 277 | 271 | 4 | 2 | 1682 | 4 | 0.20 | 2.37 | 263 | 7 | 0 | 0 | 0 | 1 | 0 |
| combined | 277 | 274 | 0 | 3 | 6277 | 404 | 0.62 | 1.45 | 258 | 7 | 0 | 0 | 0 | 9 | 0 |

The most common refusals: wrap, stage 0 can't yet lower a closure assigning (52) or incrementing (12) a
captured variable, or reading one in some positions (57), so 75 fixtures keep their functions at the
top level (arrows never rescued one that function expressions lost); alias, the alias loses a
narrowing (TS18048, 19); method, stage 0 refuses a parameter of a recursive array type (9) and a type
the checker prints but the top level can't name (TS2304, 5); identity, the checker's narrowing or
contextual typing lost (TS18048, TS18046, TS18047). The node-differs variants are the fixtures that
test evaluation order (borrow_defined_lent_field, reuse_arrays under temporaries) and the temporal dead
zone (dead_zone under wrap): the Node gate is what keeps them out.

The one original finding, navigation.a ("native stdout differs": `atanh 0.09999999999999999` on Node,
`0.1` natively, and an end angle differing in its last digit), fails the same way unchanged on this
machine's Node v24.14.1, so its variants failing with it are not counted as findings.

### Findings by transform

Every finding is "clang refused the C", fuzz's cc kind: a valid program the compiler emits C for that
doesn't compile. None is a silent miscompile, a sanitizer report or a leak. Counting variants (each
fixture's variant under a transform once), without navigation:

| transform | findings | A self-comparison | B this or super in an arrow | C one generic at two representations |
|---|---|---|---|---|
| wrap | 1 | fills | | |
| alias | 0 | | | |
| method | 1 | fills | | |
| identity | 4 | | class_oct6_deep, class_oct6_release, class_oct6_subclass_holder | generic_values |
| temporaries | 0 | | | |
| combined | 8 | fills | borrow_element_super_move, class_oct6_deep | call_targets_region, regexp_match, regexp_unicode, throw_keeps_old_value, undefined_references |

All 14 reduced with adamic-reduce against the tip (GOFLAGS=-trimpath, parallel 6), signature derived
from the variant. The first line is the derived signature.

### A. A local compared with itself is C that clang refuses (-Wtautological-compare)

An array that is a local (after wrap or method) compared with what a call on it returned, `zeros.fill(99) === zeros`, is emitted as `(adamic_local_2_zeros == adamic_local_2_zeros)`, and `-Werror` turns clang's self-comparison warning into an error. At the top level, where the original keeps `zeros` as a global, the same comparison compiles. A valid program the compiler cannot build.

**wrap, fills.a** (28 lines, 965 bytes to 4 lines, 131 bytes; 49 candidates in 7.2s)

First line: `cc:error: self-comparison always evaluates to true [-Werror,-Wtautological-compare]`

```ts
function metamorphicMain(): void {
const zeros = new Array<number>(5).fill(0);
console.log(`${''} ${zeros.fill(99) === zeros}`);
}
```

**method, fills.a** (28 lines, 984 bytes to 7 lines, 171 bytes; 52 candidates in 8.9s)

First line: `cc:error: self-comparison always evaluates to true [-Werror,-Wtautological-compare]`

```ts
class MetamorphicBlock {
	run(): void {
		const zeros = new Array<number>(5).fill(0);
console.log(`${''} ${zeros.fill(99) === zeros}`);
	}
}
new MetamorphicBlock().run();
```

**combined, fills.a** (56 lines, 1931 bytes to 10 lines, 275 bytes; 112 candidates in 12.2s)

First line: `cc:error: self-comparison always evaluates to true [-Werror,-Wtautological-compare]`

```ts
class MetamorphicBlock2 {
	run(): void {
		const zeros = new Array<number>(5).fill(0);
		const zerosAlias = (() => zeros)();
const metamorphicTemporary5 = `${''} ${zerosAlias.fill(99) === zerosAlias}`;
	}
}
function metamorphicMain(): void {
new MetamorphicBlock2().run();
}
```

### B. this or super read inside an arrow in a subclass is an undeclared C identifier

An arrow called on the spot inside a subclass's method or field initializer (identity's closure form, `(() => super.read())()` or `(() => this.label)()`) emits a read of `adamic_local_N_this` or `adamic_local_N_this_cell` that the enclosing function never declared. Every original of these fixtures builds and agrees; every reduced program still needs the arrow.

**identity, class_oct6_deep.a** (40 lines, 1853 bytes to 19 lines, 456 bytes; 196 candidates in 11.2s)

First line: `cc:error: use of undeclared identifier 'adamic_local_N_this'`

```ts
class First {
    readonly seed: number;
    readonly label: string;
    constructor() {
        this.seed = 0;
        this.label = '';
    }
    describe(): string { return ''; }
}
class Second extends First {
}
class Third extends Second {
}
class Fourth extends Third {
    override describe(): string { return (() => super.describe())(); }
}
function show(value: First): void {
    console.log(`${true}/${true}/${true}/${value instanceof Fourth}`);
}
```

**identity, class_oct6_release.a** (44 lines, 1911 bytes to 21 lines, 648 bytes; 272 candidates in 15.3s)

First line: `cc:error: use of undeclared identifier 'adamic_local_N_this_cell'`

```ts
class Base {
    readonly label: string;
    constructor(index: number) { this.label = ''; }
    read(): string { return ''; }
}
class Child extends Base {
    readonly values: readonly string[];
    constructor(index: number) { super(0); this.values = []; }
}
class Grandchild extends Child {
    readonly details: { };
    constructor(index: number) { super(0); this.details = { }; }
}
class Leaf extends Grandchild {
    readonly extra: string;
    constructor(index: number) { super(0); this.extra = ''; }
    override read(): string { return (() => super.read())() + (() => this.extra)(); }
}
function make(): Base {
    return new Leaf(0);
}
```

**identity, class_oct6_subclass_holder.a** (29 lines, 1283 bytes to 17 lines, 497 bytes; 243 candidates in 24.9s)

First line: `cc:error: use of undeclared identifier 'adamic_local_N_this'`

```ts
class Parent {
    readonly label: string;
    child: Child | undefined = undefined;
    constructor(label: string) { this.label = ''; }
    read(): string { return ''; }
}
class Child extends Parent {
    readonly detail: string;
    constructor(label: string) { super(''); this.detail = ''; }
}
class Grandchild extends Child {
    override read(): string { return (() => super.read())(); }
}
function family(): void {
    const parent = new Parent('');
    parent.child = new Grandchild('');
}
```

**combined, borrow_element_super_move.a** (47 lines, 1309 bytes to 22 lines, 408 bytes; 172 candidates in 13.4s)

First line: `cc:error: use of undeclared identifier 'adamic_local_N_this'`

```ts
class MetamorphicBlock2 {
	run(): void {
		const metamorphicTemporary = new Mapper().twice();
	}
}
interface Box {
}
class Worker {
	run(items: Box[]): number {
		return 0;
	}
}
class Mapper extends Worker {
	twice(): string {
		const items: Box[] = [];
		const itemsAlias = items;
		return (() => `${''} ${super.run(itemsAlias)}`)();
	}
}
function metamorphicMain(): void {
new MetamorphicBlock2().run();
}
```

**combined, class_oct6_deep.a** (71 lines, 2794 bytes to 25 lines, 546 bytes; 232 candidates in 19.2s)

First line: `cc:error: use of undeclared identifier 'adamic_local_N_this_cell'; did you mean 'adamic_local_N_this'?`

```ts
class MetamorphicBlock2 {
	run(): void {
		const show = function show(value: First): void {
    const metamorphicTemporary2 = (() => `${true}/${true}/${true}/${value instanceof Fourth}`)();
};
	}
}
class First {
    readonly seed: number;
    readonly label: string;
    constructor() {
        this.seed = 0;
        this.label = '';
    }
}
class Second extends First {
    readonly copy = (() => this.label)();
}
class Third extends Second {
}
class Fourth extends Third {
}
function metamorphicMain(): void {
new MetamorphicBlock2().run();
}
```

### C. One generic function at two reference representations is one C function (-Wincompatible-pointer-types)

metamorphicIdentity<T> called with values of different C representations in one program (an object and a closure, an array and a map, an array and a closure) is emitted as one C function, so one of the calls passes the wrong pointer type. Every original of these fixtures builds and agrees. Which part of instantiation conflates them is inferred from the C, not traced.

**identity, generic_values.a** (57 lines, 1755 bytes to 20 lines, 600 bytes; 152 candidates in 14.3s)

First line: `cc:error: incompatible pointer types passing 'adamic_map *' (aka 'struct adamic_map *') to parameter of type 'adamic_array *' (aka 'struct adamic_array *') [-Werror,-Wincompatible-pointer-types]`

```ts
function metamorphicIdentity<T>(value: T): T {
	return value;
}
interface Named {
	readonly name: string;
}
function arrayOf<Value>(value: Value): Value[] {
	const values: Value[] = [];
	return metamorphicIdentity(values);
}
function mapOf<Value>(value: Value): Map<string, Value> {
	const values = new Map<string, Value>();
	return values;
}
function setOf<Value>(value: Value): Set<Value> {
	const values = metamorphicIdentity(new Set<Value>());
	return values;
}
const named: Named = { name: '' };
console.log(`${arrayOf(named)[0]?.name} ${mapOf(named).get('value')?.name} ${setOf(named).size}`);
```

**combined, call_targets_region.a** (80 lines, 2150 bytes to 32 lines, 641 bytes; 277 candidates in 17.7s)

First line: `cc:error: incompatible pointer types passing 'adamic_object *' (aka 'struct adamic_object *') to parameter of type 'adamic_closure *' (aka 'struct adamic_closure *') [-Werror,-Wincompatible-pointer-types]`

```ts
function metamorphicIdentity<T>(value: T): T {
	return value;
}
class MetamorphicBlock2 {
	run(value: Item): void {
		metamorphicIdentity(value);
	}
}
class MetamorphicBlock3 {
	run(): void {
		const read = function read(): void {
};
		const readAlias = metamorphicIdentity(read);
const metamorphicTemporary4 = (() => new Keeper())();
	}
}
interface Item {
}
class Reader {
    read(value: Item): number {
        return 0;
    }
}
class Keeper extends Reader {
    override read(value: Item): number {
        new MetamorphicBlock2().run(value);
        return 0;
    }
}
function metamorphicMain(): void {
new MetamorphicBlock3().run();
}
```

**combined, regexp_match.a** (59 lines, 2228 bytes to 16 lines, 453 bytes; 205 candidates in 12.5s)

First line: `cc:error: incompatible pointer types passing 'adamic_object *' (aka 'struct adamic_object *') to parameter of type 'adamic_array *' (aka 'struct adamic_array *') [-Werror,-Wincompatible-pointer-types]`

```ts
function metamorphicIdentity<T>(value: T): T {
	return value;
}
class MetamorphicBlock2 {
	run(): void {
		const show = function show(match: RegExpMatchArray | null): string {
	return '';
};
		const showAlias = show;
const metamorphicTemporary7 = showAlias(metamorphicIdentity('zzz'.match(/(?<x>a)|(b)/g)));
const iterator = metamorphicIdentity('a ba'.matchAll(/(?<x>a)|(b)/dg));
	}
}
function metamorphicMain(): void {
new MetamorphicBlock2().run();
}
```

**combined, regexp_unicode.a** (61 lines, 2789 bytes to 15 lines, 340 bytes; 131 candidates in 13.1s)

First line: `cc:error: incompatible pointer types passing 'adamic_array *' (aka 'struct adamic_array *') to parameter of type 'adamic_closure *' (aka 'struct adamic_closure *') [-Werror,-Wincompatible-pointer-types]`

```ts
function metamorphicIdentity<T>(value: T): T {
	return value;
}
class MetamorphicBlock2 {
	run(): void {
		const show = function show(): string {
	return '';
};
		const showAlias = metamorphicIdentity(show);
const points = metamorphicIdentity('a🌍b'.match(/./gu));
	}
}
function metamorphicMain(): void {
new MetamorphicBlock2().run();
}
```

**combined, throw_keeps_old_value.a** (65 lines, 2063 bytes to 20 lines, 509 bytes; 164 candidates in 9.1s)

First line: `cc:error: incompatible pointer types passing 'adamic_object *' (aka 'struct adamic_object *') to parameter of type 'adamic_closure *' (aka 'struct adamic_closure *') [-Werror,-Wincompatible-pointer-types]`

```ts
function metamorphicIdentity<T>(value: T): T {
	return value;
}
class MetamorphicBlock2 {
	run(): void {
		const fail = function fail(): Tree {
	return { value: 0, label: '' };
};
		const failAlias = metamorphicIdentity(fail);
const moveThenAssign = function moveThenAssign(): string {
	let local: Tree = { value: 7, label: '' };
		const other = metamorphicIdentity(local);
		return '';
};
	}
}
type Tree = { value: number; label: string };
function metamorphicMain(): void {
new MetamorphicBlock2().run();
}
```

**combined, undefined_references.a** (70 lines, 2263 bytes to 18 lines, 408 bytes; 153 candidates in 13.4s)

First line: `cc:error: incompatible pointer types passing 'adamic_closure *' (aka 'struct adamic_closure *') to parameter of type 'adamic_object *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]`

```ts
function metamorphicIdentity<T>(value: T): T {
	return value;
}
class MetamorphicBlock2 {
	run(): void {
const box: Box = { };
const boxAlias = metamorphicIdentity(box);
const twice = metamorphicIdentity(doubler(true));
	}
}
function doubler(flag: boolean): ((value: number) => number) | undefined {
	return undefined;
}
interface Box {
}
function metamorphicMain3(): void {
new MetamorphicBlock2().run();
}
```

## Cost

Generating: 0.06 seconds per fixture to place and run the original on Node, 0.11 to 0.26 per fixture
for one transform, 0.62 for combined; every transform for all 277 fixtures took 65 seconds of wall
time at parallel 6. Running: 1.4 to 2.4 seconds of wall time per program in its own slot (four
compiles, four runs, the leak check), so the 1,841 programs (277 originals, 1,564 variants) took 521 to
848 seconds per checkout at parallel 6 under this load. Reducing: 7 to 25 seconds per finding.

## Blind run

Five sealed fault patches, applied with `git apply` onto worktrees of origin/main at 39638d9e (not
read, printed, grepped or diffed), each worktree's adamic built with -trimpath, every original and
every kept variant (the same 1,841 programs) run through it, and through the unpatched 39638d9e as
the control. A program catches a fault when it fails on the fault's worktree and either passes on the
control or fails there with another key. The runs were made before the megabyte digest; the three
fixtures that print that much were then run again, every transform, on every checkout and on the tip,
with the digest (all agree everywhere), and the first caught fixture of each fault was run again to
record its first line.

The control first. Two originals fail on 39638d9e itself: navigation.a (as on the tip) and
typeof_null.a ("native stdout differs", fixed by the tip), with their variants; and the A, B and C
variant findings above are all present there too.

| transform | fixtures | kept | skipped | node differs | sites kept | sites refused | generate s/fixture | run s/variant | agreed | checked | unfit | not yet | invalid | finding | crash |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| original | 277 | 277 | 0 | 0 | 0 | 0 | 0.06 | 2.53 | 268 | 7 | 0 | 0 | 0 | 2 | 0 |
| wrap | 277 | 275 | 1 | 1 | 276 | 155 | 0.11 | 1.52 | 265 | 7 | 0 | 0 | 0 | 3 | 0 |
| alias | 277 | 194 | 83 | 0 | 761 | 57 | 0.12 | 2.89 | 188 | 5 | 0 | 0 | 0 | 1 | 0 |
| method | 277 | 277 | 0 | 0 | 707 | 35 | 0.14 | 2.88 | 267 | 7 | 0 | 0 | 0 | 3 | 0 |
| identity | 277 | 273 | 4 | 0 | 2485 | 89 | 0.26 | 3.12 | 260 | 7 | 0 | 0 | 0 | 6 | 0 |
| temporaries | 277 | 271 | 4 | 2 | 1682 | 4 | 0.20 | 1.83 | 262 | 7 | 0 | 0 | 0 | 2 | 0 |
| combined | 277 | 274 | 0 | 3 | 6277 | 404 | 0.62 | 2.38 | 257 | 7 | 0 | 0 | 0 | 10 | 0 |

| fault | caught by the original fixtures | caught by variants the originals missed | first finding |
|---|---|---|---|
| blind-01 | 7 | 0 | class_as_interface.a, native stdout line 15: expected "0", native "-1" |
| blind-02 | 1 | 0 | exceptions.a, native stdout line 13: Node prints "finally round 1" after "finally round 0", native skips it |
| blind-03 | 7 | 0 | json_stringify_numbers.a, native stdout line 10: expected "1e-7", native "0.0000001" |
| blind-04 | 3 | 0 | library_object_assign_fields.a, native stdout line 3: expected "9\|10\|6\|5\|7", native "10\|9\|6\|5\|7" |
| blind-05 | 1 | 0 | library_array_with.a, crash: native AddressSanitizer: heap-use-after-free |

Variants caught every fault too, but only on fixtures whose original already caught it: blind-01 42
variants (7 under each transform), blind-02 6, blind-03 37 (alias 2, the rest 7), blind-04 18,
blind-05 6. No fault was caught only by a variant.

## What this shows, and what it doesn't

Observed: on the tip, the variants found three families of C that clang refuses (14 variants, 13
fixtures) that the 277 originals don't reach, each reduced to a program of 4 to 32 lines. On the five
sealed faults the variants added no catch the originals didn't already make.

Not covered: fixtures stage 0 doesn't lower; the second files of the two multi-file fixtures (only the
main file is rewritten); the transforms run on the tip's checker and stage 0, so a variant the blind
base can't lower would be discarded there (none was: no variant came to "not yet" or "invalid" on any
checkout). The leak check, the release comparison and the megabyte digest were proven able to fail
only by unit tests and a mutant of the counts comparison (caught), not by a blind fault.
`TestFuzzerSharesRuntimeLibrary` in internal/fuzz fails on this Mac before and after this change, with
macOS ASan's "detect_leaks is not supported"; the judge's own leak run is unchanged, and this tool runs
the oracle's macOS leak check instead.
