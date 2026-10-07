# Adamic's memory model

Status: **draft, written October 5, 2026 by @system_adamic** (#w318mqs), with strings already built on it (`2a2e30b`, `967d121`) and objects and arrays next. The rule above everything here: **no garbage collector**, not as a fallback, not for now. If a design needs one, the design is wrong.

## What lives where

- **Numbers and booleans** are values: a `double` and a `bool`, copied, never allocated.
- **Strings, objects, arrays, maps and closures** live on the heap behind a reference, and each carries a reference count.
- A **constant** the program spells out (`'Fizz'`, a literal object of constants, later) is static and immortal: its count is 0, and retain and release skip it. Every program's `''`, `'true'` and `'false'` are immortal too.

## Worked examples

Examples 1 to 5 were measured at commit `9c04abcc0569eaae387c5cb467201fec3c9f093c` (task #w318mqs).
Example 6 was regenerated at merge commit `fb07d6bae0fdd27fbf6066a6f9c2a02dd27391ab`,
which includes the shared-slice append fix `7b6f986` and its counts row `aeb1708`.
These are complete programs in `internal/oracle/testdata/memory_examples/`.
C below is copied from `adamic c <file>` at the stated commit, with intervening lines
omitted; identifiers are unchanged. Each row comes from `adamic build <file> -o <binary> --count`
and one run of the binary, in the same column order as `internal/oracle/counts.md`.
Allocations count heap values, not allocator buffers; retains and releases count calls,
even for NULL and immortal values. Freeing a container releases its contents internally,
without adding a release call to the counter.

The compiled fixtures are registered with the Node oracle, ASan, UBSan and the leak
check. The two refused programs run on Node in `TestMemoryExamplesRefused` and must
be refused before C generation: they have no generated C or counted-build row.

### 1. Prepend, walk and map a list

Source: `internal/oracle/testdata/memory_examples/list.a`

```ts
interface Item {
	readonly value: number;
	readonly next: Item | undefined;
}
function map(item: Item | undefined): Item | undefined {
	if (item === undefined) return undefined;
	return { ...item, value: item.value * 2, next: map(item.next) };
}
function sum(item: Item | undefined): number {
	return item === undefined ? 0 : item.value + sum(item.next);
}
function run(): void {
	let head: Item | undefined = undefined;
	for (let value = 1; value <= 3; value++) {
		head = { value, next: head };
	}
	console.log(`${sum(head)}`);
	const mapped = map(head);
	console.log(`${sum(mapped)}`);
}
run();
```

Generated C, selected lines in emission order:

```c
adamic_release(adamic_local_0_item);
/* ... */
bool adamic_temporary_2 = ((adamic_local_0_item->heap.references == 1 && !adamic_weak_held(adamic_local_0_item)) && !adamic_local_0_item->frozen);
adamic_object * adamic_temporary_3 = (adamic_temporary_2 ? adamic_retain(adamic_local_0_item) : adamic_object_copy(adamic_local_0_item));
/* ... */
adamic_object * adamic_temporary_8 = (adamic_object *)(adamic_temporary_2 ? adamic_temporary_6->reference : adamic_retain(adamic_temporary_6->reference));
/* ... */
adamic_temporary_6->reference = NULL;
/* ... */
adamic_object * adamic_temporary_9 = adamic_function_0_map(adamic_temporary_8);
/* ... */
adamic_release(adamic_temporary_12->reference);
adamic_temporary_12->reference = adamic_temporary_9;
/* ... */
adamic_release(adamic_local_0_item);
/* ... */
adamic_object * adamic_temporary_18 = ((adamic_object *)adamic_object_data_field(adamic_local_1_item, "next", &adamic_cache_17)->reference);
/* ... */
adamic_object * adamic_local_2_head = NULL;
/* ... */
adamic_object * adamic_temporary_23 = adamic_object_new(&adamic_shape_0);
/* ... */
adamic_temporary_23->slots[1].reference = adamic_retain(adamic_local_2_head);
/* ... */
adamic_local_2_head = adamic_temporary_23;
adamic_release(adamic_temporary_24);
/* ... */
adamic_object * adamic_temporary_28 = adamic_local_2_head;
adamic_local_2_head = NULL;
adamic_object * adamic_temporary_29 = adamic_function_0_map(adamic_temporary_28);
/* ... */
adamic_release(adamic_local_4_mapped);
adamic_release(adamic_local_2_head);
```

| Fixture | Allocations | Frees | Retains | Releases | Peak live | In regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/oracle/testdata/memory_examples/list.a | 7 | 7 | 6 | 16 | 5 | 0 |

The output is `6` then `12`: three nodes are allocated while prepending, and the map's three spreads reuse them because the dead local `head` is moved into a consumed parameter, each node is unique, unfrozen and has no Weak handle, and each replaced `next` field moves into the recursive call.
The other four allocations and frees are two number strings and two concatenations for output, giving seven allocations and seven frees, with peak five (three nodes plus two output strings) and no regions.
The six retains are three prepend links (including NULL) and three reuse retains; the six `next` reads across the two walks are borrowed, since the node they come from stays alive for the call (a stable strong field chain), so they take no count. The sixteen releases are three old heads, four consumed map arguments (including NULL), three replaced links, four output temporaries and two locals at scope exit.

### 2. A tree whose parent is Weak

Source: `internal/oracle/testdata/memory_examples/tree.a`

```ts
import type { Weak } from 'adamic';
interface Tree {
	readonly value: number;
	parent: Weak<Tree>;
	readonly children: readonly Tree[];
}
function leaf(value: number): Tree {
	return { value, parent: undefined, children: [] };
}
function branch(value: number, children: readonly Tree[]): Tree {
	const root: Tree = { value, parent: undefined, children };
	for (const child of children) child.parent = root;
	return root;
}
function up(node: Tree): number {
	const parent = node.parent;
	return node.value + (parent === undefined ? 0 : up(parent));
}
function run(): void {
	const root = branch(1, [branch(2, [leaf(3)]), leaf(4)]);
	const first = root.children[0];
	if (first !== undefined) {
		const last = first.children[0];
		if (last !== undefined) console.log(`${up(last)}`);
	}
}
run();
```

Generated C, selected lines in emission order:

```c
adamic_weak * adamic_temporary_1 = adamic_weak_of(NULL);
adamic_array * adamic_temporary_2 = adamic_array_new(0, true);
adamic_object * adamic_temporary_3 = adamic_object_new(&adamic_shape_0);
/* ... */
adamic_temporary_3->slots[1].reference = adamic_temporary_1;
adamic_temporary_3->slots[2].reference = adamic_temporary_2;
/* ... */
adamic_object * adamic_temporary_10 = adamic_object_new(&adamic_shape_0);
/* ... */
adamic_temporary_10->slots[2].reference = adamic_retain(adamic_local_2_children);
/* ... */
adamic_weak * adamic_temporary_17 = adamic_weak_of(adamic_local_4_root);
/* ... */
adamic_temporary_18->reference = adamic_temporary_17;
if (adamic_temporary_20 != NULL) adamic_release(adamic_temporary_20);
/* ... */
adamic_weak * adamic_temporary_23 = adamic_retain(((adamic_weak *)adamic_object_field(adamic_local_3_node, "parent", &adamic_cache_22)->reference));
adamic_object * adamic_temporary_24 = (adamic_object *)adamic_retain(adamic_weak_target(adamic_temporary_23));
adamic_weak * adamic_temporary_25 = adamic_weak_of(adamic_temporary_24);
/* ... */
adamic_object * adamic_temporary_28 = (adamic_object *)adamic_retain(adamic_weak_target(adamic_local_6_parent));
/* ... */
adamic_object * adamic_temporary_29 = (adamic_object *)adamic_retain(adamic_weak_target_present(adamic_local_6_parent));
/* ... */
adamic_release(adamic_local_9_last);
/* ... */
adamic_release(adamic_local_8_first);
adamic_release(adamic_local_7_root);
```

| Fixture | Allocations | Frees | Retains | Releases | Peak live | In regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/oracle/testdata/memory_examples/tree.a | 12 | 12 | 28 | 31 | 12 | 0 |

The output is `6`: strong child edges own the tree, while a parent's Weak edge owns a counted handle rather than its target; `runtime/weak.c` makes one handle per target and reuses it for siblings (two handles here, not three).
The twelve allocations and frees are four nodes, four child arrays, two handles and two output strings; peak twelve occurs during output, and no statement uses a region.
The twenty-eight retains and thirty-one release calls protect child arrays, loop elements, construction temporaries and Weak reads, including temporary strong target reads; `last` and `first` are released before `root`, so releasing the root's final count then frees every remaining node, array and handle through `runtime/heap.c`'s iterative teardown, as the leak check verifies.

### 3. The same tree with a strong parent

Source: `internal/oracle/testdata/memory_examples/refused/tree.a`

```ts
interface Tree {
	readonly value: number;
	parent: Tree | undefined;
	readonly children: readonly Tree[];
}
function leaf(value: number): Tree {
	return { value, parent: undefined, children: [] };
}
function branch(value: number, children: readonly Tree[]): Tree {
	const root: Tree = { value, parent: undefined, children };
	for (const child of children) child.parent = root;
	return root;
}
function up(node: Tree): number {
	const parent = node.parent;
	return node.value + (parent === undefined ? 0 : up(parent));
}
function run(): void {
	const root = branch(1, [branch(2, [leaf(3)]), leaf(4)]);
	const first = root.children[0];
	if (first !== undefined) {
		const last = first.children[0];
		if (last !== undefined) console.log(`${up(last)}`);
	}
}
run();
```

Refusal from `adamic c` (repository absolute prefix removed):

```text
adamic: internal/oracle/testdata/memory_examples/refused/tree.a:3:2: Adamic 0.1 refuses Tree.parent, a mutable field of type Tree | undefined, which can reach back to the Tree holding it: a cycle reference counting can't free, and the write at internal/oracle/testdata/memory_examples/refused/tree.a:11:32 may close one (the value written may reach what it's written into, where run is called from the top level); declare it parent: Weak<Tree> (import type { Weak } from 'adamic'), which doesn't count and reads undefined once what it points to is freed; or make it readonly; or write into it only values this function made, or only into what it made (adamic/cycle-capable)
```

No C is emitted; `adamic build ... --count` is refused too, so there are no counts.

Changing only the import and `parent` type makes the back edge strong: the parent's children can reach it again, so the assignment in `branch` can close a cycle.
The exact fix for this program is to restore `import type { Weak } from 'adamic'` and `parent: Weak<Tree>`, as in example 2; merely making the parent readonly would require a different builder and would not by itself prove every constructor write safe.

### 4. Captured const, captured let, and a captured function

Source: `internal/oracle/testdata/memory_examples/closures.a`

```ts
function run(): void {
	const fixed = 7;
	const readFixed = (): number => fixed;
	let changing = 7;
	const readChanging = (): number => changing;
	changing = 9;
	console.log(`${readFixed()} ${readChanging()}`);
}
function countdown(n: number): number {
	return n === 0 ? 0 : 1 + countdown(n - 1);
}
run();
console.log(`${countdown(3)}`);
```

Generated C, selected lines in emission order:

```c
adamic_cell *adamic_local_1_fixed_cell = adamic_cell_new((adamic_value){.number = (0x1.cp+02)}, false);
adamic_closure * adamic_temporary_1 = adamic_closure_new(adamic_function_2_closure, 1);
adamic_temporary_1->cells[0] = adamic_retain(adamic_local_1_fixed_cell);
/* ... */
adamic_cell *adamic_local_3_changing_cell = adamic_cell_new((adamic_value){.number = (0x1.cp+02)}, false);
adamic_closure * adamic_temporary_2 = adamic_closure_new(adamic_function_3_closure, 1);
adamic_temporary_2->cells[0] = adamic_retain(adamic_local_3_changing_cell);
/* ... */
adamic_local_3_changing_cell->value.number = (0x1.2p+03);
/* ... */
adamic_release(adamic_local_4_readChanging);
adamic_release(adamic_local_3_changing_cell);
adamic_release(adamic_local_2_readFixed);
adamic_release(adamic_local_1_fixed_cell);
/* ... */
double adamic_temporary_13 = self->cells[0]->value.number;
/* ... */
double adamic_temporary_15 = self->cells[0]->value.number;
```

| Fixture | Allocations | Frees | Retains | Releases | Peak live | In regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/oracle/testdata/memory_examples/closures.a | 9 | 9 | 2 | 9 | 7 | 0 |

The current design uses a cell for both captured `const` and captured `let`: `const fixed = 7` gets `adamic_cell_new`, just like `let changing`, and both closures read `self->cells[0]`, consistent with the [Cycles section](#cycles-found-by-the-compiler-broken-by-weak), which includes captures of either binding kind in its cell-based cycle rule.
The output is `7 9` then `3`, and nine allocations and frees comprise two cells, two closures and five output strings, with peak seven (four capture values plus the first statement's three strings) and no regions.
Two retains give the closures ownership of their cells, and nine release calls drop five output strings and four scope owners; the function-declaration fix below allocates no function value or cell for its recursion.

Source: `internal/oracle/testdata/memory_examples/refused/closure.a`

```ts
function run(): void {
	let countdown: (n: number) => number = (n) => n;
	countdown = (n) => n === 0 ? 0 : 1 + countdown(n - 1);
	console.log(`${countdown(3)}`);
}
run();
```

Refusal from `adamic c` (repository absolute prefix removed):

```text
adamic: internal/oracle/testdata/memory_examples/refused/closure.a:2:6: Adamic 0.1 refuses 'countdown', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; write the function as a function declaration (function countdown() {}), which captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable)
```

No C is emitted; `adamic build ... --count` is refused too, so there are no counts.

This closure captures the local variable that owns the closure, so it would keep its own cell alive forever.
Replace it with the `function countdown(n: number): number` declaration in the accepted program above: recursion calls a named C function directly and creates no self-capture cycle.

### 5. A graph per loop item, summarized

Source: `internal/oracle/testdata/memory_examples/regions.a`

```ts
interface Graph {
	readonly value: number;
	readonly next: Graph | undefined;
}
function build(value: number): Graph {
	return { value, next: tail(value + 1) };
}
function tail(value: number): Graph {
	return { value, next: undefined };
}
function sum(graph: Graph): number {
	if (graph.next === undefined) return graph.value + 0;
	return graph.value + graph.next.value;
}
function run(): void {
	let total = 0;
	for (let item = 0; item < 3; item++) {
		total += sum(build(item));
		const held = build(item);
		total += sum(held);
	}
	console.log(`${total}`);
}
run();
```

Generated C, selected lines in emission order:

```c
static adamic_object * adamic_function_0_build_in(adamic_region *region, double adamic_local_0_value);
/* ... */
adamic_object * adamic_temporary_1 = adamic_function_1_tail((adamic_local_0_value + (0x1p+00)));
adamic_object * adamic_temporary_2 = adamic_object_new(&adamic_shape_0);
/* ... */
adamic_temporary_2->slots[1].reference = adamic_temporary_1;
/* ... */
static adamic_object * adamic_function_0_build_in(adamic_region *region, double adamic_local_0_value) {
/* ... */
adamic_object *adamic_temporary_4 = adamic_function_1_tail_in(region, (adamic_local_0_value + (0x1p+00)));
adamic_object *adamic_temporary_5 = adamic_object_new_in(region, &adamic_shape_0);
/* ... */
adamic_temporary_5->slots[1].reference = adamic_temporary_4;
/* ... */
adamic_object *adamic_temporary_9 = adamic_object_new_in(region, &adamic_shape_0);
/* ... */
adamic_region adamic_temporary_25 = ADAMIC_REGION;
adamic_object *adamic_temporary_26 = adamic_function_0_build_in(&adamic_temporary_25, ((double)adamic_local_4_item));
double adamic_temporary_27 = adamic_function_2_sum(adamic_temporary_26);
/* ... */
adamic_region_end(&adamic_temporary_25);
adamic_object * adamic_temporary_28 = adamic_function_0_build(((double)adamic_local_4_item));
adamic_object * adamic_local_5_held = adamic_temporary_28;
double adamic_temporary_29 = adamic_function_2_sum(adamic_local_5_held);
/* ... */
adamic_release(adamic_local_5_held);
```

| Fixture | Allocations | Frees | Retains | Releases | Peak live | In regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/oracle/testdata/memory_examples/regions.a | 14 | 8 | 0 | 5 | 2 | 6 |

The output is `18`: each of three `total += sum(build(item))` statements uses its own region and frees its two nodes together, while `const held = build(item)` uses the heap because `held` survives its declaration statement.
Fourteen allocations are twelve nodes and two output strings: six nodes go with regions, eight values are freed individually, peak live is two, and zero retains plus five releases reflect moves into returned fields, borrowed `sum` parameters, three heap owners and two output temporaries.
This is a region for one statement inside a loop, not for the whole iteration: values kept across statements still use the heap; regions for an entire call or request and arrays in regions remain outside this example's supported pattern (`internal/native/region.go`, `runtime/region.c`).

### 6. Append a string, slice it, then append to the slice

Source: `internal/oracle/testdata/memory_examples/strings.a`

```ts
function run(): void {
	let text = '';
	for (let index = 0; index < 160; index++) text += 'x';
	let slice = text.slice(16, 144);
	slice += '!';
	console.log(`${text.length} ${slice.length} ${text.charAt(144)} ${slice.charAt(128)}`);
}
run();
```

Generated C, selected lines in emission order:

```c
static adamic_string adamic_string_0 = ADAMIC_STRING("");
static adamic_string adamic_string_1 = ADAMIC_STRING("x");
static adamic_string adamic_string_2 = ADAMIC_STRING("!");
static adamic_string adamic_string_3 = ADAMIC_STRING(" ");
/* ... */
adamic_string * adamic_local_0_text = adamic_retain(&adamic_string_0);
/* ... */
adamic_local_0_text = adamic_string_append(adamic_local_0_text, 1, (adamic_string *const[]){&adamic_string_1});
/* ... */
adamic_string * adamic_temporary_2 = adamic_string_slice(adamic_local_0_text, (0x1p+04), (0x1.2p+07), true);
adamic_string * adamic_local_2_slice = adamic_temporary_2;
adamic_local_2_slice = adamic_string_append(adamic_local_2_slice, 1, (adamic_string *const[]){&adamic_string_2});
/* ... */
adamic_release(adamic_local_2_slice);
adamic_release(adamic_local_0_text);
```

| Fixture | Allocations | Frees | Retains | Releases | Peak live | In regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/oracle/testdata/memory_examples/strings.a | 14 | 14 | 2 | 17 | 5 | 0 |

Nine owned string allocations grow the 160 one-byte appends geometrically; appends within the available capacity reuse the unique owned string, and `runtime/string_share.c` makes one shared slice header for this 128-byte slice (its owner's header and bytes are under eight times the slice's).
Appending to this shared slice copies because its bytes belong to its owner and its capacity is zero: `runtime/string_append.c` requires `length + added <= capacity` before writing in place, fixed in `7b6f986` and held by `shared_slice_append.a`; Node and native now both print `160 129 x !`, with the generated C above unchanged at the merged runtime.
The fourteen allocations and frees are nine grown owners, one slice header, one owned copy for the slice append and three output strings (two numbers and one concatenation); the two characters are the runtime's immortal ASCII strings and allocate nothing. Peak is five, with no regions, two retains (initial immortal empty string and slice owner) and seventeen releases (nine replaced loop strings, the replaced slice header, five output temporaries and the two locals).

### Checking the documented counts

`TestMemoryExampleCountsMatchDocumentation` builds each compiled example counted afresh
(no observation-cache reuse), runs it with the counts table's 8 MiB stack, and compares
all six numbers with its row above; it also requires each example registered with Node.
`TestCountsAreRecorded` separately holds these rows in `internal/oracle/counts.md`.
Mutant: change the list's documented allocations from `7` to `8`, keeping the source
and counts table unchanged. The documentation test fails only `list.a`, reporting the
measured row with `7`; restoring `7` makes it pass. No compiler change is needed.

Run: `go test ./internal/oracle -run '^TestMemoryExampleCountsMatchDocumentation$' -count=1 -v`.
The mutant exits 1; after restoration it exits 0. Refusal-check mutants replace the
strong-parent probe with example 2 and the self-capture probe with the function-declaration
fix in example 4: `go test ./internal/oracle -run '^TestMemoryExamplesRefused$/tree$' -count=1 -v`
and the same command ending in `/closure$` each exit 1 with
`want cycle refusal with fix, got <nil>`; the restored refusal probes pass.

## What the runtime is

All paths below are under `internal/native/runtime/`.

- **Heap header:** `adamic.h` defines `adamic_heap`: `size_t references`, kind tag and
  `uint32_t slab`. `heap.c` retains, releases and iteratively frees values and their contents.
- **Size-class allocator:** `heap.c` serves values up to 256 bytes in 16-byte classes
  from 64 KiB chunks, recycles empty chunks across classes, and uses malloc for larger
  values. Sanitized builds use malloc/free per value so ASan and LeakSanitizer see them.
- **Strings:** `string.c` and `string_build_impl.h` own UTF-8/WTF-8 storage;
  `adamic.h` defines immortal constants; `string_share.c` holds owners for shared
  slices; `string_append.c` grows owned strings. `string_index.c` keeps UTF-16 length,
  checkpoints, cursor and compact UTF-16-view caches; slices have their own caches,
  and in-place append invalidates the old index and length. Shared-slice append
  copies into owned storage, preserving the owner's bytes.
- **Arrays and maps:** `array.c` holds typed elements in a growable buffer; `map.c`
  holds keys and values in an ordered hash table, with typed wrappers in `map_set.c`.
  Their element and entry ownership is released by `heap.c`.
- **Regions:** `region.c` bump-allocates statement-scoped objects, releases heap
  references they hold and frees its blocks on `adamic_region_end`; region values
  have count zero while alive. `internal/native/region.go` proves the call pattern
  and emits heap and `_in` variants.
- **Weak handles:** `weak.c` shares a counted handle per target, using a side table;
  the target is uncounted by the handle and freeing it invalidates the handle.
- **Sharing across threads:** the header's count is plain today. See
  [docs/concurrency.md](concurrency.md) (branch `codex/concurrency`) for how only
  shared values acquire atomic counts; `adamic.h` and `heap.c` are the current
  single-threaded implementation, not an implementation of that design.

## Counting

Every heap value starts at 1. `adamic_retain` adds one, and `adamic_release` subtracts one and frees at zero. The compiler inserts every retain and release; nothing at runtime guesses who's still looking.

The conventions stage 0 uses today, chosen for being obviously right rather than fast:

| Where a reference is | Who owns it |
|---|---|
| A value made mid-statement | the statement; released when it ends |
| A variable | the variable; released when reassigned or when its scope ends, by `break`, `continue` or `return` included |
| A parameter | the callee, which retains on entry and releases on every way out |
| A return value | the caller, which receives it owned |
| A field or element | the object or array that holds it; released when that's freed |
| A module's top-level variable | the program; released when main finishes, the last declared first (modules in reverse of their evaluation order), so the leak check sees what it reached. Not on a panic, which stops where it stands |

Counts are plain `size_t`, not atomic, because 0.1 is single-threaded. When concurrency lands (#p286ycm), a value shared across threads is one of the shareable, immutable kinds, and its count becomes atomic at the moment it's shared (Lean 4 does exactly this), so single-threaded code never pays for atomics.

## Freeing

Releasing an object releases what it holds. A long chain (a linked list a million long) would overflow the C stack if freed recursively, so freeing works from a list rather than by recursion. That's a correctness rule, not an optimization, and stage 0 must have a fixture that frees a chain far deeper than the stack.

## Reuse in place (Perceus)

The counts are what let Adamic mutate in place without anyone seeing:

- When a value whose count is exactly 1 is consumed to build a new value of the same shape, the new value takes over the old memory. `{ ...tree, left: insert(tree.left, value) }` on a tree nobody else holds rewrites one node instead of copying it. That's Perceus, from Koka; Lean 4 does the same.
- `push` on an array with count 1 grows it in place, and on a shared one copies first.
- **Identity is never observable through reuse.** Reuse happens only when nothing else holds the old value, so nothing can compare against it. Program 9 in docs/0.1.md pins this: while `before` still holds the old tree, the path is copied and the shared subtree stays shared.

**What's built (#6zt3jmn).** In stage 0, the one place a value is consumed to build one of its own shape is the spread: `{ ...source, field: value }` always has the source's shape. When the source's count is exactly 1, the spread takes over its object: the literal's fields are written over its own, and nothing is allocated or freed. Uniqueness is checked at runtime, as Perceus does. What the compiler proves (`internal/native/reuse.go`) is that taking the object can't be seen:

- **The source is owned here.** It's a local, or a consumed parameter. A borrowed parameter's count is its caller's, so a count of 1 there means the caller still holds it, and it's never reused.
- **The source is dead.** Liveness over `internal/flow`'s graph says it isn't read after the instruction, or the instruction gives its variable a new value.
- **Nothing else reads it.** In that instruction, the source is read only by the spread and by reads of its fields inside the literal, and each field the literal replaces is read at most once. While the fields' values are evaluated, nothing but the literal can reach the object, so nothing can change it, and JavaScript's order (the spread's fields read first) holds without a copy.
- **A replaced field moves out.** Its one read moves the value out instead of retaining it. That's what lets `insert(tree.left, value)` find the next node unique too.

A parameter that's such a source is **consumed**, not borrowed. Its caller hands over a reference: a statement's temporary as it is, anything else retained. The callee releases it on every way out but the one where its memory became the result.

An argument a consumed parameter takes is **moved** when nothing reads its variable after the call. For a local, that means it's dead there. For a global, the statement must be assigning it anew, and nothing the call can reach may read or write it or call a function value. That's program 9's `tree = insert(tree, value)`.

**What the counts show.**

- **`09_tree.ts`:** allocations and frees fell from 33 to 19, and retains from 199 to 161. Its first nine inserts rebuild their path in place. The tenth runs while `before` holds the root, so its path is copied and the right subtree stays shared: "right subtree shared: true" still prints, held by the oracle.
- **`reuse.a`:** 86 allocations fall to 77 (5 in a loop that reuses one object every pass, 4 down a uniquely held list), measured against the same build with the plan turned off.
- **Nothing else moved.** The other spreads in the fixtures are of globals, or of values still live.

**How it's held.** `reuse.a` is a case for every way taking an object could be seen: its source read after, held by someone else, its replaced field read twice, and a moved global read by the callee. Beside those are cases where it's taken and mustn't change what prints. Mutants, against the whole oracle:

- **Reusing a shared object** (the count check dropped): stdout differs.
- **Ignoring liveness:** stdout differs, in `reuse.a` and in the existing `maybe_number_slots.a`.
- **Moving a replaced field read twice:** UBSan reports a null dereference.
- **Moving a global the callee reads:** UBSan reports a null dereference.
- **Moving a field out of a shared object:** stdout differs.

Liveness is held the way the graph's edges are. `TestLivenessHoldsOnEveryPath` walks every call's points backward, and wherever the next thing to touch a variable is a read, liveness must say it's live: 3,417,833 checks over every program. Liveness that doesn't flow back around a loop failed 1,147,791 of them, in 36 programs.

**Three holes, found in review (integration 5), and closed.** Reviewer R's probes are now oracle fixtures (`reuse_*.a`):

- **A borrowed parameter was moved.** `forward(point) { return bump(point) }` moved `point` into `bump`'s consumed parameter, but `point`'s count was the caller's. `bump` saw a count of 1 and reused the caller's object, then freed it. A move now hands over only a count the function owns.
- **A moved global was read by a sibling argument.** In `tree = insert(tree, size())`, `size()` read `tree` after it was moved out (nulled). The check looked only inside the callee. It now looks at every call in the statement.
- **A `Weak` was ignored.** A `Weak` reaches its target without counting, so a count of 1 doesn't mean nothing else can reach it. The `Weak`'s holder could write the object while it was being taken over, and find the new object in the old one's place after. Reuse now also requires that no `Weak` points at the value (`adamic_weak_held`, one comparison in a program with no `Weak`).

Each has a mutant that puts the old behavior back:

| Mutant | Caught by |
|---|---|
| A borrowed parameter moved | ASan, in `reuse_forward.a` |
| Only the callee checked before moving a global | UBSan null dereference, in `reuse_global_sibling.a` |
| Weak handles ignored | stdout differs, in both `Weak` fixtures |

**Found on the way:** the spread itself was miscompiled. Native copied the source's object after evaluating the fields' values, so `{ ...point, x: moveY(point) }` showed a write `moveY` made to `point.y`. Node prints `1 0 99`, native printed `1 99 99`. It's fixed, and `spread_snapshot.a` holds it.

**Reuse for arrays.** There are three more places a value is consumed to build one of its own kind. Each is taken over under the same rules: owned here, dead after the instruction, and read once in it.

- **`map` writes in place.** `const ys = xs.map(f)`, when `xs` is unique, writes each result over the element it came from, in `xs` itself. That needs three more conditions: the callback is written right there, it takes no third argument (the array, which would see its elements replaced as the map went), and its results sit in the array's slots as the elements do.
- **A spread first in an array literal is appended to.** `[...xs, more]`, when `xs` is unique, is `xs` with `more` appended. Lowering already requires the spread's elements to be the literal's.
- **A discarded `splice` makes no array.** A `splice` whose result nothing uses lets go of what it removed instead of returning it (`adamic_array_remove`).

`push`, `pop`, `fill` and `reverse` never copied to begin with: JavaScript mutates an array in place even when it's shared, so they had nothing to gain.

**What the counts show.**

- `normalize.a` makes 10 fewer allocations and its peak live falls from 16 to 11.
- `sort_releases.a` makes 1 fewer, with its peak live from 13 to 11.
- `splices.a` makes 2 fewer.
- `reuse_arrays.a`, new, makes 78 against 84 with the array plan turned off.

Releases rise where `map` writes in place. That's how the counter is wired, not more work: freeing an array lets go of its elements inside the runtime without calling `adamic_release`, so it isn't counted, and replacing elements in place calls it once each.

The `trees` and `sort` benchmarks have none of these patterns, and their allocations didn't move. Since B's run, their retains have fallen from the earlier units: `trees` from 375,302,854 to 272,979,306, and `sort` from 6,002,000 to 2,000,004. `trees` allocates 68 million plain objects it never spreads, which is the arenas' work, below.

**How it's held.** `reuse_arrays.a` puts each array reuse beside the cases where it would be seen. Mutants, each caught:

- `map` reusing a shared array: stdout differs.
- Ignoring liveness: stdout differs, in `reuse_arrays.a`, `reuse.a` and `maybe_number_slots.a`.
- Ignoring a callback's third argument: stdout differs.
- The spread reusing a shared array: stdout differs.
- `adamic_array_remove` not letting go of what it removed: the leak check fails, in `reuse_arrays.a` and the existing `splices.a`.

**Not yet:** a spread inside a closure is reused only when its source is the closure's own local. A closure's parameters stay owned, so they are never unique. `filter`, `slice` and `concat` of a dying array still allocate, and `sort` still sorts a copy even when its array is unique and no comparator can reach it.

**A spread of a value that may be undefined** (reviewer R, round 8b, older than reuse). JavaScript's `{ ...undefined }` is `{}`, but natively the copy read the NULL source's shape, and once spreads became reuse sources the uniqueness check read its count first. Now, where the checker says the source may be undefined, lowering records the source type's fields the literal doesn't give (`ObjectLiteral.Empty`). For a NULL source the emitter makes a new object with those fields, each set to undefined (NULL for a reference, the packed undefined for a number), and writes the literal's own fields into it. Every read Adamic has then gives what JavaScript reads from `{}`. Reuse checks for NULL before it looks at the count. A boolean or `Weak` field the literal doesn't give says NotYet, since neither has an undefined of its own in a slot. `spread_undefined.a` covers it, reused and copied, with fields and without, and spread again. Three mutants, each caught there: the copy without the NULL path (UBSan), the reuse check without it (UBSan), and the empty object's number fields left 0 (no sanitizer fires; stdout differs, `0 0` for `0 none`).

## Borrowed parameters

Status: **built**, as designed on paper on October 5, 2026 by stream A2 (#r3s6nwg's follow-on, with #5jck546 answered at the end). The design is kept below as written, and what landing it measured follows each part it predicted.

### What a parameter costs today

The callee retains every reference parameter on entry and releases it on every way out (the table above). That's two calls per reference parameter per call, and the counts table (`internal/oracle/counts.md`) shows them: 3,402 retains and 5,163 releases across the 54 fixtures.

### The caller already keeps every argument alive

Borrowing has two halves. The callee mustn't keep the reference past the call without taking its own count, and the caller must keep the value alive until the call returns. Stage 0's emitter already does the second half, for its own reasons (the comment on `native.C`). Every argument is a C expression that stays valid to the end of its statement:

- a value made mid-statement is a temporary the statement owns, released when the statement ends;
- a field read, a global read and a captured variable's read are retained into such a temporary the moment JavaScript reads them, so a write later in the same statement (by the callee too) can't free them;
- a plain local is written only by its own function, so it can't change while that function is waiting on a call.

The first half is done too. Every place that keeps a reference past its statement takes a count of its own: a store into a local (`store` retains, then releases the old value), a `return` (retained into the result), a field write, an element write, `push`, and a map's `set` (through `held`). A parameter read flows into those like any other value.

So the callee's retain on entry protects against exactly two things:

1. **A reassigned parameter.** `store` releases the old value, and a borrowed parameter's old value is the caller's, not the callee's. Releasing it would free what the caller still holds.
2. **A closure's parameter.** A closure is also called from loops the emitter writes for the array methods, and `map`'s loop hands it `elements[index]` in place, unretained. A callback that writes that element (`array[index] = other`) would free its own parameter. (`forEach`, `filter`, `find`, `findIndex`, `some`, `every` and `reduce` each retain the element before the call, and the sort holds its own count of every element it compares, so those are safe. `map` is the one exception, and the rule below stays conservative for every closure until that's fixed.)

### The rule

A reference parameter of a named function or a method, `this` included, is **borrowed** unless some `ir.Assign` anywhere in the program writes it. A closure's parameters stay owned.

A borrowed parameter is neither retained on entry nor released on the way out. Nothing else in the emitter changes.

### How the IR knows

- `ir.Assign` is the only statement that writes a local that already exists. `Declare`, `ForOf` and a pattern's `Binding` always make a new one. Only statements hold statements, so one walk over every function's body and over `Main` finds every write. Locals are numbered across the whole program, so a write from inside a closure (through its cell) is the same index as the parameter it writes.
- The walk goes in lowering, as a small pass of its own (`internal/lower/borrow.go`), and its answer goes on the IR (`Local.Borrowed`). Lowering decides what the program means and the emitter does what it's told, as with `Checked` and `Captured`. The JavaScript backend ignores it.
- The emitter skips the entry retain and the exit release for a borrowed parameter. A `store` into one is a compiler bug, and it says so out loud (a Go panic) rather than emitting a release of something it doesn't own.
- A default parameter is already two locals: the incoming argument and the local the body declares from it. The incoming one is never written, so it's borrowed, and the declared one is an ordinary local.

### What the counts table should show

A prototype of exactly this rule (the walk, plus the emitter change, nothing else) passed the whole oracle on Linux: stdout, stderr and exit codes against Node, ASan and UBSan, and LeakSanitizer on every fixture that finishes. Its counts, against the table at `d6af38d`:

- **Allocations, frees and peak live didn't move on any of the 54 rows.** Borrowing changes no lifetime. If one of those columns moves, borrowing is doing something it shouldn't.
- **Retains and releases fell by the same amount on every row**: the number of reference parameters of named functions and methods the run passes. 19 fixtures moved, and the other 35 have no such call. In total, retains went from 3,402 to 2,958 (444 fewer, 13%) and releases from 5,163 to 4,719.
- The largest: `09_tree.ts` 300 to 199 retains (`insert`, `collect` and `height` recurse on borrowed trees), `case_mapping.a` 267 to 171, `05_wordcount.ts` 194 to 142, `06_stack.ts` 148 to 112 (`this` in every method, and `push`'s item), `splices.a` 96 to 60.

When it lands, the table's diff should be exactly that: retains and releases down by equal amounts, and nothing else moving.

**Landed, it was.** By then the table had grown to 60 rows (main's new fixtures and the input fixtures), plus the two fixtures below. Measured against the table just before borrowing: 20 rows moved, every one by equal retains and releases and nothing else. Allocations, frees and peak didn't move on any row. Retains fell from 19,283 to 18,828 (455 fewer) and releases from 52,171 to 51,716. Each of the 54 rows the prototype measured moved by exactly what the prototype said, 444 in all, and the other 11 are `read_files.a`, added since.

### Proving it can fail

Two mutants, each needing a fixture that builds its strings at runtime (literals are immortal):

1. **Borrow a reassigned parameter.** A function that reassigns a string parameter it was handed, called with a string the caller goes on using, makes `store` release the caller's string. ASan has to catch a use after free or a double free.
2. **Borrow a closure's parameter.** A `map` callback that overwrites the element it was given, then reads its parameter, makes the parameter dangle. ASan has to catch it. That fixture is also what would let closure parameters be borrowed later. `map`'s loop would retain each element it hands out, as the other visits already do, or the aliasing analysis below would prove the callback can't write the array.

Both are oracle fixtures: `borrow_reassigned.a` and `borrow_map_overwrite.a`. The mutants, run against the whole oracle:

- **Borrowing reassigned parameters** is stopped first by the emitter's own guard: a store into a borrowed parameter is a Go panic, a compiler bug said out loud.
- **The same, with that guard removed:** ASan caught a heap-use-after-free in `borrow_reassigned.a`, where `framed` read the string `louder`'s store had freed. `functions.a` failed too, on the leak check, where a reassigned parameter's new value was never let go.
- **Borrowing closures' parameters:** ASan caught a heap-use-after-free in `borrow_map_overwrite.a`'s callback, which read the `name` the overwrite had freed. Nothing else failed.

### Lent reads: borrowing extended to what a statement reads

A reference read from a place a call could write (a global, a captured variable, a field) is retained the moment JavaScript reads it. Something later in the statement might write that place and free the value before it's used. B measured the cost in the tokenizer benchmark: every `text.charCodeAt(position)` of the global `text` was a retain and a release.

The count buys nothing when nothing can run between the read and its use. So a read is **lent**, taken without a count (`internal/native/borrow.go`), when both of these hold:

- It's a direct operand of a **consumer**: an operation done with its operands once it's evaluated, whose result is a number, a boolean, a new value, or one holding its own count. A conditional, `??` or a cast hands its operand on as its own value, so it isn't one.
- Every operand of that consumer is **pure**: it writes no variable, field, element or map, calls no code, and frees nothing. That's a fixed list of the IR's operations. Anything off the list, including an operation added later, isn't pure.

Purity is the proof here, not the alias graph. A global or a captured variable isn't a value in the flow graph, so its range says nothing about it. One more thing makes the proof hold in C: a consumer's result is often a C expression its parent evaluates later, possibly after a sibling's call. So a consumer that lent a read has its result pinned to a temporary right after its operands, where JavaScript evaluates it.

**What the counts show.**

- The tokenizer benchmark's retains fell from 10,048,047 to 2,407,586, and its releases by the same 7,640,461. Allocations are unchanged.
- In the oracle's table, 63 rows moved, all retains and releases. Every row fell by equal amounts except the three fixtures that panic (`08_results.ts`, `writes_past_end.a`, `maybe_boolean_panic.a`). Those are counted where they stopped, and the release that would have matched a skipped retain comes after the panic.

**Time didn't move much.** On `go run ./bench -only tokenizer -rounds 7` in the noisy cloud container (load around 2), the best native run went from 0.867 s to 0.824 s, against Node's 0.27 s. A non-atomic retain and release is cheap, and the tokenizer's time is elsewhere. Also, the benchmark's text is degenerate: Node and native both print 1 identifier and 597,482 numbers, so its identifier path barely runs.

**How it's held.** `lent_reads.a` puts a call that reassigns what's read beside the read, or beside its consumer: a global, a field and a captured variable. The strings are built at runtime and held only by what's reassigned. Mutants, each caught by an ASan heap-use-after-free:

- A call counted as pure: in `lent_reads.a`, and in the existing `read_order.a` and `adversarial_order.a`.
- A closure call counted as pure: in `lent_reads.a` and `adversarial_order.a`.
- A consumer's result not pinned: in `lent_reads.a`.

### Where it goes next: reuse in place pulls the other way

Perceus needs the callee to **own** what it reuses. `insert` in program 9 can rewrite `{ ...tree, left }` in place only if it holds the only count on `tree`, and under the rule above `tree` is borrowed, because `insert` never reassigns it. Lean 4's rule is the one to grow into. A parameter is owned when the body consumes it (reuses its memory, or passes it to an owned parameter), and borrowed otherwise. That's computed to a fixpoint over the call graph, and a call that passes a borrowed value to an owned parameter retains at the call. With reuse, the counts table will show both directions at once. Retains come back on the parameters reuse needs, and allocations and frees fall where reuse fires. That's why the table has both kinds of column.

### Which IR this is built on (#5jck546)

cohere carries a port of the React Compiler's HIR (`cohere/internal/lint/ecmascript/high_level_intermediate_representation`, 134 files). I read its types (`high_level_intermediate_representation.go`), its lowering (`lower.go`), aliasing effects (`effects.go`), mutable ranges (`ranges.go`) and non-escaping scopes (`prune_non_escaping_scopes.go`). It's a control-flow graph per source function, with three-address instructions over named values and single assignment on request (`ssa.go`). It infers aliasing effects per instruction (13 of React's 19 kinds), plus mutable ranges as half-open intervals. It's careful work, and it measures its own gaps.

What it lacks for a compiler, by its own account:

- **Proven types.** `Identifier.Type` is left nil. The seam is the AST node, to ask the checker. Adamic needs the representation after monomorphization (`Stack<string>` is its own copy), and the HIR is one graph per source function, not per instantiation.
- **Ownership.** It has no owned or borrowed, no consume, no retain or release. `EffectCapture` ("may be retained by something that outlives this reference") is the nearest fact. It's derived for React's questions from a signature table keyed by syntactic name (`EffectGapTypeDirectedShapes`), with an unknown call assumed to mutate and capture every operand.
- **The interprocedural question.** "Does this callee keep my argument?" is borrow inference's whole question. It's exactly `EffectGapInterproceduralParameters`, which the HIR doesn't implement, and whose cost "to a general consumer asking 'does this callee mutate my argument' is total".
- **Refusing.** Lowering never abandons a function. What it can't model becomes `UnsupportedNode`, a value with unknown effects, and classes are among those today. That's right for a linter. Adamic's doctrine is `NotYet` with where and what, so a compiler pass meeting `UnsupportedNode` would have to refuse anyway.
- **One meaning.** Adamic's lowering decides JavaScript's evaluation order, the snapshots and the inserted checks once, and the oracle holds that. Building the HIR from the AST too would be a second lowering, a second place for the order to go wrong, and nothing would hold it.

What it has that Adamic will need: a control-flow graph with single assignment, and the aliasing graph and mutable ranges built on it. Perceus needs those for last uses and for where drops go, and arenas need them for escape.

**Recommendation: lift its passes onto Adamic's IR. Don't build the HIR.** Adamic keeps one lowering, from the checked AST to its typed IR. When a pass needs control flow, Adamic builds a graph from its own IR, which is already structured (`If`, `Loop`, `ForOf`, `Switch`, `Break`, `Continue`, `Return`), so the blocks fall out mechanically. Then cohere's single-assignment construction and its aliasing and ranges machinery get lifted onto that graph. React's signature table is replaced by facts the IR already carries. Every runtime operation is its own IR node with one known effect, written down once. Every direct call names its function, whose summary is computed from its body. Only a call through a closure value falls back to the conservative default. Borrowed parameters, as designed above, need none of this: they land on today's tree IR. The decision starts to cost something at Perceus.

That recommendation came from reading, not from measuring. What would settle it was lifting `ssa.go`'s `Construct` onto a graph built from Adamic's IR, and counting how much of it carries over unchanged. Kirk decided for lifting, and that has now been done and measured.

### The lift, measured

`internal/flow` is the graph. `Build` makes one per IR function and one for the top level. Each instruction points back at the IR statement or condition it came from and lists the locals it reads and the one it writes. That's all single assignment needs, and evaluation order stays decided in one place. Globals and captured variables aren't values in the graph: any call may write a global, and any closure holding a cell may write a captured variable. cohere leaves its `LoadGlobal` unrenamed for the same reason.

cohere's `ssa.go`, `ssa_eliminate.go`, `ssa_verify.go` and `graph.go` (at 715ba94) were copied in and then edited only where Adamic's graph differs. Counted line by line against the originals (`difflib`, comments and blank lines apart):

| File | Code lines kept | Removed or changed | New | Comments kept |
|---|---:|---:|---:|---:|
| `ssa.go` (`Construct`) | 182 of 226 | 44 | 3 | 123 of 235 |
| `ssa_eliminate.go` | 81 of 82 | 1 | 0 | 36 of 37 |
| `ssa_verify.go` | 222 of 229 | 7 | 2 | 70 of 91 |
| `graph.go` | 108 of 188 | 80 | 2 | 34 of 62 |
| **All four** | **593 of 725 (82%)** | **132** | **7** | **263 of 425 (62%)** |

What the 132 were:

- 57 are cohere's evaluation order. It's deferred, not incompatible: nothing lifted reads it yet, and mutable ranges will bring it back.
- About 20 are structural fallthroughs, which Adamic's terminals don't have. Removing them keeps the order a reverse postorder, which is what single assignment needs. Ranges will want cohere's loop-body-first order back, and a fallthrough on `If` with it.
- 17 rename the `Returns` place. An Adamic return's value is read by an instruction instead.
- About 15 handle context stores. A captured variable isn't in the graph at all.
- 7 copy a place's effect, reactivity and source range, which an Adamic place doesn't carry.
- The rest recurse into nested functions. An Adamic closure is its own IR function, with its own graph.

The algorithms themselves came over unchanged: Braun's lookup, sealing and incomplete phis, redundant-phi elimination to a fixed point, and the dominance verifier. What Adamic wrote is the graph (`flow.go`, 235 lines) and the builder from its IR (`build.go`, 280 lines). The answer to #5jck546 holds up: the passes lift, and the part that's Adamic's own is the part that should be, the graph made from the IR that carries the proven types.

**How it's held.** `TestEveryFunctionIsInSingleAssignment` builds and constructs every function of every oracle program, plus `internal/flow/testdata/joins.a` (every shape of join). That's 217 functions, 52 phis and 640 uses. Three checks run on each, and none of them is built from the construction:

- cohere's verifier: every value is defined once, and every use is dominated by its definition.
- Reaching definitions, computed the textbook way over the graph before construction. A use that one definition reaches must name that definition's value. A use that several reach must name a phi whose operands come to exactly those definitions.
- A count of the IR's reads made by `fmt`'s `%#v` rather than by `Build`'s walk, so a read the builder drops shows up.

Three mutants, each caught by the check aimed at it:

- A loop header that skips its incomplete phi failed reaching definitions 127 times and the verifier never. The first predecessor dominates the header, so dominance alone can't see a loop-carried value lost.
- Collapsing every phi failed the verifier 17 times and reaching definitions 140 times.
- A builder walk that misses reads inside slices failed the read count 76 times.

**The graph's shape is held too.** Reaching definitions runs over the graph `Build` made, so a wrong edge (a `break`, a `continue` or a case test going to the wrong block) would fool it. `TestEveryPathNodeTakesIsInTheGraph` holds the edges to what runs. The JavaScript backend can mark every point the graph has an instruction for (`javascript.Options`, which adds nothing unless asked: the default output was compared byte for byte on all 59 programs). Each program runs on Node with the marks, and every call's sequence of points must walk its graph: next in the block, or first in a block the terminal reaches through blocks that run nothing. A program that finishes must end every call at a return. Over 63 programs, 158,821 points were walked (86 programs and 3,173,486 points since main's new fixtures). Mutants, each failing this test and none failing the single-assignment test:

- A loop's `continue` sent to the loop's exit failed 4 programs.
- A loop's `break` sent to its update failed 1.
- A for...of's `continue` sent to its exit failed 3.
- A case test sent to the wrong case's body failed 5.

A switch case falling through into the next case's body is not a mutant at all. Every 0.1 case ends in `break` or `return`, so that edge would leave a block nothing reaches, and the live graph doesn't change.

### Mutable ranges and aliasing, lifted and measured

The second lift brings over cohere's alias graph and mutable ranges, React's `InferMutationAliasingRanges`, with evaluation order back to number them. Counted the same way against cohere's files at 715ba94 (code lines; the first three rows are this lift):

| File | Kept | Removed or changed | New |
|---|---:|---:|---:|
| `ranges.go` (alias graph, `mutate`, both halves) | 682 of 809 (84%) | 127 | 0 |
| `graph.go` (with evaluation order back) | 132 of 188 (70%) | 56 | 2 |
| `effects.go` (the effect types) | 120 of 944 (13%) | 824 | 0 |
| `ssa.go`, `ssa_eliminate.go`, `ssa_verify.go` (the first lift) | 485 of 537 | 52 | 5 |

- **`ranges.go`:** the 127 are 79 for `MutationSites`, which counts HIR instruction shapes, 10 for nested functions, 9 for `StoreContext`, 14 for React's frozen parameters plus the context and `Returns` places, 11 for React's frozen closures, and 5 for the return terminal. The alias graph, the `mutate` worklist and both halves of the pass came over unchanged.
- **Evaluation order:** 23 of the 57 lines came back unchanged. The other 34 are two lines for each of the 17 terminal kinds Adamic doesn't have.
- **`effects.go`:** only the types came over, on purpose. What makes effects in cohere (React's signature table keyed by the callee's name, and the HIR instruction switch, 824 lines) is replaced by Adamic's own `infer.go`, 341 lines. Every runtime operation is its own IR node there, so `push`, `set` and `splice` say exactly what they write. A call (a function, a closure, a runtime loop calling back) does what cohere assumes of an unknown one. It also mutates every value that has escaped this function: a parameter, a value from a global or captured variable, a call's result, or anything handed to a call or stored into an escaped value. Strings, numbers and booleans are created primitive and take part in nothing.
- **One adaptation:** a value read out of a container goes through a temporary that `infer.go` mints and creates from the container. Adamic's graph has no temporaries of its own, and only `CreateFrom` carries a mutation back to the container transitively.
- **Inert for now:** cohere's freeze machinery came over unchanged, but nothing emits a `Freeze` yet. Adamic's `readonly` types could: a value the checker proved readonly can't be mutated, which is the fact a freeze states.

**How it's held: the check runs against what Node does.** `TestEveryMutationIsInItsRange` runs every program on Node with every point marked. Before each point, every tracked variable's object is printed whole (every field, element and map entry it reaches). A variable that holds the same object as at the previous point, which now prints differently, was mutated by the instruction there. The value that variable held there must have a range containing that instruction. That holds the alias graph to the truth: a mutation through a value it didn't know aliased lands outside a range. Over 86 programs, 29,341 mutations were checked. None fell outside its range, none was on a range the pass left unset, and no range was invalid.

`internal/flow/testdata/mutations.a` exercises every way one value can be mutated through another: through a field, an element, a for...of element, a map value, either side of a conditional, a container holding it, a callee, and `push`. Writing it found a real error: a value read out of `holder ?? {...}` was aliased to the container non-transitively, so `counter`, which the container held, didn't have the write in its range. The temporary above is the fix.

Six mutants, each caught:

| Mutant | Mutations outside a range |
|---|---:|
| A write into a container doesn't mutate it | 29,249, in 10 programs |
| A read of a variable isn't the same object | 29,277, in 11 |
| A call doesn't mutate escaped values | 20, in 3 |
| `mutate` doesn't travel back to what a value was created from | 7, in `mutations.a` |
| A for...of element isn't part of what the loop iterates | 3 |
| The mixed read above goes back to an `Alias` | 3 |

The last three were caught only once `mutations.a` existed. Before it, no program mutated through those paths in a function whose values nothing else reached.

**Not covered:** the check sees mutations of tracked variables only. Globals and captured variables aren't values in the graph, and a variable a call reaches only through one isn't checked at that call. The check also counts a change in anything a variable's object reaches, which is exactly what a transitive range claims and more than a direct one does. A direct range that's too short for a mutation it only reaches indirectly would show up here as a failure, not a pass, so the check errs toward failing.

## Cycles: found by the compiler, broken by Weak

Compiler-generated async protocol layouts have one narrow exception: the canonical, unexported provenance identities in `internal/ir/async.go`, accepted by `fresh.RuntimeBreaksCycles` and checked at `lowering.findCycles` in `internal/lower/cycles.go`. These are IR/runtime types, not source checker declarations. Their internal frame -> waiting Promise -> reaction -> frame edges are broken by settlement, pending-subscription cancellation and normal-exit registry teardown in `runtime/async.c`; reference counting alone is insufficient. The [async design](concurrency-async.md#promise-representation-counts-and-cycles) specifies the order. Names, structural shape, annotations and user brands never confer this exemption. A user class named exactly `adamic_generated_frame_0`, `adamic_async_promise` or `adamic_async_reaction` still goes through the ordinary slot/write checks; the spoof tests and name-based mutants hold that boundary. User payloads and captures do not inherit it. The first async lowering admits only primitive payloads and bypasses synchronous lifetime proofs across suspension; additional reference payloads need the normal cycle/fresh-write analysis, without asking callers to put Weak on runtime-owned protocol edges.

Reference counting can't free a cycle, and a garbage collector is refused (no cycle collector, ever: decided by @system_adamic, task #gsz351g). What's known:

- **Immutable data is acyclic.** A value built from `readonly` parts can only point at values that already existed when it was made, so no `readonly` structure can ever reach itself. Immutable by default is the first answer.
- **A cycle needs a write.** It forms only when an existing object's mutable slot (a mutable field, an array or map element, a closure's captured variable) is set to something that can reach that object back.
- **That's visible in the types.** A mutable slot of type `T` inside type `S` can close a cycle only if `T` can reach `S` through the type graph. The checker holds that graph, so every *cycle-capable* slot can be found at compile time.

What stage 0 does (`internal/lower/cycles.go`):

1. **The finder** walks the whole program's type graph after lowering and refuses every cycle-capable slot that isn't declared `Weak`, with the fix (`adamic/cycle-capable`). A slot is a field (a `readonly` one too, since its constructor writes it: below), an element of an array that isn't `readonly`, a `Map`'s value, or a variable a function value captures (`let` or `const`: the cell is written after the closure captured it). Reaching follows an object's fields and the fields of every object type in the program that can be seen as it (a `Dog` seen as an `Animal` brings its `owner` along), an array's elements, a map's keys and values, and, through a function type, the variables captured by every function value in the program that can be seen as it, since a type doesn't say what a function captured. A `Weak` is not followed. A value just made (a literal, a call's result, `new`) is examined only where it's kept, since nothing writes through it before.
2. **What that means in practice.** A parent pointer must be `Weak`. A mutable child array of the same type (`children: TreeNode[]`) is cycle-capable too (`node.children.push(root)`), so a tree's children are `readonly`, or `Weak`, or written only with what the relaxation below proves can't close a cycle. Either link of a doubly linked list alone can close a cycle (`a.next = a`), and the types can't tell `next` from `prev`, so both are `Weak` and something else (an array) owns the nodes. A closure kept in a variable it captures (`let countdown = ...; countdown = (n) => countdown(n - 1)`) is a cycle; the fix is a function declaration, which captures nothing. A callback field is refused only when some function value in the program captures a variable that can reach the object holding it.
3. **Spelling.** `import type { Weak } from 'adamic'`, then `parent: Weak<TreeNode>`, `Weak<TreeNode>[]`, `Map<string, Weak<TreeNode>>`, or a `let` of a `Weak` type. `Weak<Target>` is `(Target & WeakBrand) | undefined`: tsc accepts it, a plain `TreeNode` assigns into it, and a read is a `TreeNode` once narrowed, so it reads as `TreeNode | undefined`. The brand gives weak slots a type of their own, so the compiler sees them in the type graph, not just in the syntax. An array of `TreeNode` seen as an array of `Weak<TreeNode>` (which tsc allows) says NotYet: one holds targets and the other handles.
4. **Natively** (`runtime/weak.c`), a `Weak` slot holds a counted handle, shared by every weak slot pointing at one target, never the target itself, so it doesn't count. Freeing a target tells its handle, which then reads `undefined`. A side table from target to handle is how freeing finds it, consulted only while any handle exists, so a program without `Weak` pays nothing.
5. **In the JavaScript backend**, a `Weak` is a plain reference, which is what Node does with the source.

**Where native and Node differ, by design:** a `Weak` read after its target's last strong holder let go is `undefined` natively, while on Node the collector keeps the target as long as the `Weak` points at it. A read the checker had narrowed to present panics natively instead of reading freed memory. Programs that read a `Weak` only while its target is held strongly (a child's parent, while the tree is held) mean the same on both; the oracle's fixtures are such programs, and `internal/oracle/weak_test.go` pins the difference itself.

### Relaxing the finder for fresh writes

The type rule is sound and, for builders of trees, costly: every parser pushes freshly made children into a mutable `children: Node[]`, and a `Node[]` whose elements can reach a `Node[]` is cycle-capable, so it's refused, and stage 1's ports (285K lines of lint rules behind them) must build each child list into a local array and freeze it `readonly`. This is the design for letting such a slot stand undeclared when the program provably never closes a cycle through it. It needs A2's control-flow graph (`internal/flow`: single assignment per function, and aliasing effects, Create, Capture, Alias and Mutate, with values escaped when a parameter, a global, a captured variable, a call's result, or a value handed to a call). The design is kept as written; what was built, and how it's held, follows it.

**What it proves.** A cycle needs a last write, the one that closes it, and every slot on a cycle is cycle-capable by type, which the finder already knows. So it's enough that at every write into a cycle-capable slot, the value written can't reach the slot's holder. Then no write closes a cycle, and by induction over the program's writes none ever forms. A slot is relaxed (allowed undeclared) only when every write into it in the whole program is proven so; any one write that isn't keeps the refusal, now naming that write. Writes are: a field assigned or set in a literal, class initializer or constructor (through any view of the holder, as the finder's `related` already matches views); an array's push, index write, `fill`, `splice`, literal and spread; a map's `set` and literal entries; and a captured variable's assignment (a cell's holder is the closure, which has always escaped, so cells are never relaxed).

**The two proofs.** Within the function that makes the write, over the whole function at once (flow-insensitively, so a capture on a later loop iteration counts), either suffices:

1. *A fresh value.* Everything the written value can reach was made in this function (a literal, `new`, or a call to a function summarized as returning fresh values, below), nothing in it has escaped (been handed to a call, stored in a global or a cell, or into an escaped value) anywhere in the function, and the holder is never captured into any of it. Then the value reaches only things this function made and saw everything done to, and the holder isn't among them.
2. *A fresh holder.* The holder was made in this function and hasn't escaped before the write, and nothing in the function ever captures the holder into the value or anything it reaches. Whatever the value reaches that was made elsewhere can't reach the holder, since nothing made elsewhere has had a chance to hold it. This is the parser's case: `const node = { kind, children: [] }; node.children.push(parseChild());` is proven without knowing anything about `parseChild`.

A function is *summarized as returning fresh values* when its result, by proof 1 inside it, reaches only what it made, its parameters' strong contents excepted, and then a call's result is fresh at a call site whose arguments in those parameters are themselves fresh there. A Weak edge is no capture at all (it keeps nothing), so a parent pointer declared `Weak` doesn't spoil either proof.

**Where it falls back to the refusal.** Any write neither proof covers: a holder that is a parameter (a helper `addChild(parent, child)`; extending proof 2 across calls, by checking every call site, is the next step if the ports need it), a holder that escaped before the write, a value read from a global, a cell or a call not summarized, a value or holder handed to a call that could tie them together, values a runtime callback produces (`map`, `forEach`, a `sort` comparator, whose writes the graph can't see into), and anything inside a generic class judged per instantiation where the flow graph has none.

**How it's held honest.** Every way to sneak a cycle past a proof is a probe that must stay refused: the holder captured into the value before the write; the same on the loop's next iteration; the value and the holder handed to one call; the holder stored in a global, then reached from the value through a function; a function summarized as fresh that really returns its argument, or something read from a global; a write through another view of the holder; a fresh-looking value that spreads an old one (`{ ...old }`); and each runtime write (`splice`, `fill`, `concat`, spread, `Array.from`). Every program it accepts runs under the leak check, which is the ground truth: a hole that lets a cycle through leaks, and LeakSanitizer reports it (with globals cleared at exit, so a cycle behind one shows). Each mutant drops one condition of a proof (ignore captures after the write, take every call's result as fresh, ignore escapes into globals, skip the views) and must turn at least one of those probes from refused into accepted and leaking.

**Built (stream B3).** `internal/fresh` is the proof, `internal/lower/fresh.go` ties it to the finder. Where it departs from the design above, it's on the side of refusing:

- **One interpretation, both proofs.** Each function is interpreted over its graph (`internal/flow`, exception edges included), flow-sensitively to a fixed point, so a capture on a later pass of a loop reaches the write through the back edge. The abstract objects are allocation sites (a literal, `new Map()`, `split`'s array, an object a summarized call makes), each split by recency into its newest instance and its older ones, plus *outside*: anything the function didn't make, or made and let escape. The state is what each variable and each object's fields, elements and entries may hold (weak updates only), and which objects have escaped: handed to a call or a callback, stored in a global or a captured variable, thrown, or stored into anything outside or escaped, with everything they reach, through a `Weak` too, since whoever holds the `Weak` can read it. A write is proven when no object the value reaches strongly may be the holder: the confined objects it reaches aren't the holder, and if it reaches anything outside or escaped, the holder is confined. A fresh value and a fresh holder are the two ways that holds.
- **The holder is the object written into.** In `node.children.push(child)` that's the array `node.children` holds, found through the field, so a node whose `children` was replaced by an outside array isn't taken as fresh.
- **Recency.** A site's newest instance is fresh when it's made, whatever its older ones did: a loop that fills a node and then lets it escape into a list proves the next pass's node. The older instances keep every escape.
- **Summaries take their parameters as outside.** A function's summary is the confined objects its result reaches at its exit (after any `finally`), by where each was first made; a call makes them anew. A parameter's contents are outside in it, not "fresh at a call site whose arguments are fresh"; arguments always escape. A call through a function value or a runtime callback (`map`, `forEach`, `Array.from`, a `sort` comparator) is outside, with everything handed to it escaped. The fixed point over the call graph starts at "returns nothing" and falls back to "returns something outside" after 32 rounds.
- **Every slot a write could be writing.** Lowering records each write's holder type (`ir` nodes carry a `Site`). A write that isn't proven keeps refused every cycle-capable slot of its kind (a field by its name, an array's elements, a map's keys and values, a set's elements) whose holder is related to its own, as the finder matches views; one whose holder type is generic or unknown keeps every slot of its kind refused, and an IR node the proof doesn't know keeps every slot refused. A slot with no unproven write stands, including one nothing writes at all (a parent pointer only ever set in literals).
- **The refusal names the write.** It still points at the slot, and adds which write may close the cycle and why: `..., and the write at main.a:7:2 may close one (the value written may reach what it's written into)`. Cells are refused as before.

**How it's held.** 24 probes in `internal/oracle/testdata/fresh_refused`, each a program Node runs whose line marked `// closes the cycle` really closes one: the design's list, plus an index write, `map`'s callback, a closure that captures the holder, a constructor that adds itself to its parent, a holder that escapes on the way to a `catch`, a node an earlier pass of a loop let escape, a summarized function whose result escapes in a `finally` after its `return`, and, since Sets hold objects and Maps take them as keys, a set given itself and a map keyed by its holder. `TestFreshWriteProbesStayRefused` requires each refused, naming that write; one accepted is built, run, and leak-checked, and the failure says what LeakSanitizer found. `fresh_parser.a` (a recursive descent parser pushing each child into the node it's parsing, gap 4's and gap 5's shapes from stage 1, and a helper given the parent with a `Weak` parent pointer) and `fresh_writes.a` (every kind of write, a loop, a constructor, a write after a catch) run three ways and leak nothing. `TestEveryWriteIsRecordedAndKnown` holds that every write in every oracle program is one lowering recorded and the proof knows: 314 writes in 148 programs, 295 proven. Mutants, each against the probes, and each caught by LeakSanitizer reporting the cycle (Node and native print the same and exit 0, so only the leak check sees it):

| Mutant | Probes accepted, and leaking |
|---|---|
| Every value taken as fresh (reaching outside ignored) | `one_call`, `wraps_argument`, `returns_argument` |
| Escapes before the write ignored (an escaped holder taken as confined) | the same three |
| A call through a function value or callback taken as fresh | `map_callback`, `array_from`, `closure_captures` |
| A summary's outside taken as fresh | `one_call`, `wraps_argument`, `returns_argument`, `returns_global`, `global_then_function`, `caught`, `older_instance`, `escapes_after_return` |
| Escapes into globals ignored | `older_instance`, `global_then_function`, `returns_global`, `escapes_after_return` |
| Views skipped (a write matched only to its own holder type) | `another_view` |
| Loops not iterated to a fixed point | `next_iteration`, `older_instance` |
| Recency forgetting an older instance's escape | `older_instance` |
| A summary ignoring what escaped by the function's exit | `escapes_after_return` |
| `set.add` stored without being judged | `set_add` |
| A write lowering didn't record (`TestEveryWriteIsRecordedAndKnown`) | 15 writes reported |

A weaker form of the recency mutant, which drops the escape but still re-closes escapes after the rename, isn't caught: the array a node holds is renamed first, while the node is still escaped, and re-closing puts the array's escape back. That mutant is masked rather than caught, and no probe reaches the difference.

**Not covered yet.** Writes in a closure see every captured variable as outside. The holder proof across calls, below, came after.

**A hole reviewer R found (round 12), and closed.** The finder skipped `readonly` fields, which was sound while every write into a mutable cycle-capable slot was refused: nothing half-built could be put where it reaches back. With the relaxation, a constructor can push `this` into its parent's children (proven: `this` is fresh then) and then set its own `readonly parent` to the parent, which closes the cycle, and that write was never asked about. R's `ctor_readonly.a` compiled and leaked 993 bytes under LeakSanitizer, silently in a release build; `ctor_wrapped.a` (`this` inside a literal pushed) and `ctor_set.a` (`this` added to a set) were the same hole. Now a `readonly` field is a slot like any other. The only write a `readonly` field has is its constructor's, and that write is judged: `this.parent = parent` with `this` still confined is proven, as a fresh holder, and stands; set after `this` has gone into the parent, it's refused with the fix `declare it parent: Weak<Node>`. A `readonly` field whose type can't reach back is no slot at all, as before. R's three probes are in `fresh_refused`, each naming the `this.parent` write. The mutant that puts the skip back makes all three compile and leak (three LeakSanitizer reports) and fails `TestReadonlyFieldsAreJudgedByTheirConstructorsWrites`.

### The holder proof across calls

`function addChild(parent: Node, child: Node): void { parent.nodes.push(child); }` is refused today: inside it, both the holder and the value are parameters, so neither proof holds, and a caller handing it a node it made learns nothing either, since every argument escapes into a call. Whether that push closes a cycle depends on what the caller passes, so the design is to judge it there.

**Parameters as placeholders.** A function only ever called directly (by `ir.Call`, never as a sort comparator, and not a closure, whose calls the graph can't see) is interpreted with each reference parameter as a placeholder object `P(i)`, and each field `f` read from one as a second placeholder `D(i, f)`, which stands for everything the caller's argument reaches through that field: a read further down is `D(i, f)` again, so whatever the function does with something deep in its argument (lets it escape, writes into it) is said of the whole of it. A function with a parameter some closure captures keeps its parameters outside. Placeholders are escaped from the start, so every judgment the function makes about itself stays exactly as conservative as it is now, with a parameter as outside. What changes is what the function's summary can say about them.

**The summary becomes a transfer.** Besides what the function returns, it records, in terms of placeholders, the function's own fresh objects by origin (now including those it stored only into placeholders, which the caller makes anew and stores where the edges say), and outside:

1. *Edges:* what the function stored into a placeholder's fields, elements or entries.
2. *Leaks:* which placeholders it let escape for real: handed to a call that isn't summarized, stored in a global or a captured variable, thrown, or stored into outside or into something leaked. A placeholder that starts escaped is not a leak; only what the function does is.
3. *Clobbers:* whether it called anything not summarized, which may write anything it can reach into the caller's arguments.
4. *Deferred writes:* each write the function couldn't prove, with its holder and value.

Its result may now name a placeholder too: `same(node)` returns the argument, precisely, rather than outside.

**At a call site,** the caller replays the summary in its own state: each `P(i)` becomes the argument, and each `D(i, f)` everything the argument reaches through `f` in the state after the edges are added (repeated until nothing more is added, so a callee that wrote through one parameter and read through another, which the caller passed the same object for, is seen); each fresh object is made anew, as now; outside stays outside. It adds the edges (a store into something exposed lets what's stored escape, as any store does), lets the leaked arguments escape, and, when the callee clobbers, marks its own placeholders as possibly holding outside. Arguments no longer escape just by being passed. Then it judges each deferred write against its state after the call: a write proven at this call site is done here; one that isn't is deferred again, into the caller's own summary, when the caller has placeholders, and is refused otherwise, naming the write in the callee and the call. A write is proven when every place it was deferred to proves it. A write in a function that never leaves (neither returns nor throws) stays refused.

**Why it's sound.**

- *Placeholders are outside for everything the function decides alone.* A placeholder is escaped, its fields read as `D` and, after any call that clobbers, as outside too. So a write the function proves without its callers is one it would prove with every parameter outside, which is what's proven today. Aliasing between parameters (`addChild(node, node)`) can't fool it: no judgment rests on two placeholders being different objects.
- *The caller sees no less than happened.* Every effect the callee had on what the caller can reach goes through its arguments (edges, leaks) or through what's already escaped in the caller (a global, a captured variable), which the caller already reads as outside. A clobber can only reach the caller's arguments or escaped objects, and only the caller's own placeholders are read more precisely than outside, so those are what it marks.
- *Judging after the call is conservative.* The abstract heap and the escaped set only grow, so if the value could reach the holder when the callee wrote, it still can in the state after the call. `D(i, f)` is translated as everything the argument reaches through `f` after the replay, which covers what it reached when the call began and whatever the callee added, through any parameter.
- *Every way a function runs is a call site that judges it.* Writes are deferred only out of functions called by `ir.Call` alone. A function used as a sort comparator, or a closure, keeps its parameters outside and defers nothing. A function nobody calls never runs its writes.
- *Summaries are found to a fixed point as before, deferred writes included.* Each is keyed by the write it came from, and its holder and value range over finite sets (origins, a placeholder per parameter and per field read from one, outside), so the fixed point exists. When it's given up on, every function goes back to outside parameters and nothing is deferred.

**What it should prove:** `addChild(node, parseItem())` from a function that made `node`; the same through a helper that calls `addChild` in turn (deferred twice); `addChild(root, new Node())` at the top level, where the value is fresh; and a helper that makes the child, adds it, and returns it.

**What must stay refused, as probes:** `addChild(node, node)`; `addChild(node, wrapper)` where `wrapper` holds `node`; a parent the caller let escape into a global, given a child that reads it; a callee that stores the parent in a global before the caller writes into it; a callee that calls a closure which writes the parent; a callee that stores the parent into the child (`child.owner = parent`) and then pushes; a callee used as a sort comparator; a callee that pushes and then throws, with the write after the catch; a deferred write that's fine at one call site and not at another; a grandchild write two fields deep; and a callee that returns its argument, now named precisely, still refused when the caller writes it into itself. Each a program whose marked write closes a cycle, so LeakSanitizer sees it if accepted. Mutants: a placeholder that starts confined, replay without the edges, leaks not replayed, clobbers ignored, a comparator treated as a direct call, and a deferred write taken as proven without its call sites judging it.

**Built (stream B3), and where it departs from the above.**

- *A field and what's below it are two placeholders.* Reading field `f` of `P(i)` gives `F(i, f)`, which a caller translates as exactly what its argument holds there; reading anything of `F(i, f)` gives `S(i, f)`, everything below, which a caller translates as everything that reaches, strongly or weakly. A holder `parent.nodes` is then the array itself at the call site, not the whole subtree, which after the push would include the child just pushed. A write two fields deep (`parent.nodes[0].nodes.push(...)`) has the whole subtree as its holder, so it's proven only for a value that reaches nothing outside and nothing in it.
- *What a deferred write's value reaches outside is judged by what had escaped before the call.* Replaying a push into a global holder lets the pushed value escape, and that escape is the write's own effect, not a way the value could have reached the holder. While the callee runs, only code it can't see into could tie its arguments to anything, so before the call is what counts, unless the callee clobbers, and then everything escaped after the replay counts. Whether the holder is exposed is judged after the replay: the callee may have stored it into something escaped itself.
- *A function's exits are its returns and its throws,* so a write before a throw is in its summary. One that never leaves keeps its deferred writes refused.

**How it's held.** 12 more probes in `internal/oracle/testdata/fresh_refused/call_*.a`, the list above (with the child holding its parent written into the child's literal, so the probe is about the push and not about a second cycle-capable field). Each stays refused naming its marked write. `fresh_calls.a` runs three ways leak-clean: a parser built on `addChild`, `addPair` deferring through `addChild`, a helper that makes a child with a `Weak` parent pointer and returns it, and pushes at the top level into a global holder of fresh values. `TestEveryWriteIsRecordedAndKnown`: 314 writes in 148 programs, 295 proven. Mutants, against the probes:

| Mutant | Probes accepted, and leaking |
|---|---|
| A placeholder that starts confined | 11, among them `call_same`, `call_wrapper`, `call_owner`, `call_twice_deferred`, `call_grandchild` (10 leak reports; the 11th was refused at another write) |
| Replay without the edges | `call_grandchild` |
| Leaks not replayed | `call_leaks_parent`, `caught` |
| A comparator treated as a direct call | `call_comparator` |
| A deferred write dropped instead of judged by callers | 35, every probe but one (34 leak reports; the 35th was refused at another write) |
| Clobbers ignored | `clobber_link_loop` |
| What had escaped before the call ignored | `escaped_before` |
| `load()` not adding outside for a clobbered placeholder | none: changed fields covered by replayed leaks or earlier escapes; `TestClobberedPlaceholderReadsCoveredAtCaller` |
| `everything()` not adding outside for a clobbered placeholder | none: the same coverage for every copied field; `TestClobberedPlaceholderReadsCoveredAtCaller` |

The clobbers and earlier-escapes rows once read "none: masked", on the argument that every link code the callee can't see could add is itself a write into a cycle-capable slot, refused where that code is judged. That's false: the link can be a write the proof accepts, a fresh child pushed into a parent, made by a function value or through a global the callee can't see, and only together with the later write does it close a cycle. Integration's reading of aee89b2 (#fxspptb) showed it: without either condition `clobber_link_loop.a` (a function value adds the child) and `escaped_before.a` (a closure adds it through a global the parent escaped to) compile, and leak 800 objects each under LeakSanitizer. Both are probes now, with eight more from the same reading that hold other conditions in `internal/fresh` no probe reached: `everything()`'s exposed reads, a handler's state before and after, a destructured `for...of` keeping its element, and `reverse`, `fill`, `sort`, `map.set` and `set.add` returning their receiver (`spread_outside`, `throw_keeps_old`, `map_entries_pattern`, `alias_*`).

The two placeholder-read merges are different from that clobbers judgment. Each read still includes a field placeholder without the merge (`fresh.go:660-703`), and placeholders are exposed (`state.exposed`, lines 433-435), so removing it cannot prove an extra write locally. For an opaque call to change the concrete object read, it must be able to reach that object: through an operand, or through something already outside. `call` escapes every operand (lines 991-996); `escape` and `escapeObject` follow strong and weak contents (lines 370-409); `summarize` records leaked placeholders (lines 1794-1799); and `made` replays those leaks to a fixed point before judging or returning (lines 1510-1536). A concrete read from an exposed object already adds outside in the other branch of `load` or `everything`. Access through globals and captured variables reads outside (`value`, lines 1019-1021); objects stored there have escaped (`define`, lines 882-884). Visible stores into exposed objects propagate exposure (`storeInto`, lines 439-452, and `closeEscapes`, lines 414-429); summary edges do the same during replay (lines 1520-1529), so access introduced through an alias is covered too. A caller's own placeholders can defer again, but their leaks travel through the next summary until concrete arguments are reached. If opaque code cannot reach the object, it cannot change its fields, so the extra outside names no new real possibility. This argument uses reachability, not the false claim that an opaque call's linking write must itself be refused. The deferred judgment still needs both earlier escapes and the clobbers clause (lines 1476-1479 and 1543-1547), including links added by accepted writes.

`clobbered_loads_test.go` checks both reads with and without their added outside, for a parameter handed to an opaque call strongly or weakly, access through a child with a weak back link, and an object already escaped at the caller. The translated result includes outside and the deferred write stays refused. `TestClobberedReadProgramKeepsRefusal` also runs a small IR program through the graph and summary fixed point, comparing a read after the callback with one saved before it, which retains just the placeholder. These are checks of the covering mechanisms, not a leak observation: removing either merge still accepts none of the refusal probes, so that test never enters its leak check. Nor are all judgments equal: an unrelated callback sets clobbered too. `TestClobberedPlaceholderReadCanLosePrecision` shows an empty field of a confined argument then produces an extra refusal with either merge; without the added outside, that safe write is proven. The proof's behavior is kept unchanged.

## Exceptions, designed into counting

`throw`, `try`, `catch` and `finally` are 0.2's (#zek5q21). Unwinding is where reference counting is hardest: every frame a throw leaves holds references (variables, a statement's temporaries, a loop's held array or map iterator), and each must be let go exactly once on the way out, or the program leaks or frees twice. This is the design; the first cut is below it.

### What a thrown value is

JavaScript throws any value. Adamic, for now, throws only an `Error`: `throw new Error(message)`, or the error a `catch` caught, thrown again. Anything else is refused, with the fix `throw new Error(String(value))`. The reason is the catch side: what a `catch` binds is `unknown`, and holding any value there means a value of every kind at once, which 0.2 doesn't have yet. An `Error` is an object like any other, with two fields, `name` and `message`, counted like any object; `stack` isn't there, since its text is the engine's and differs between runs. A `catch (error)` narrows it with `error instanceof Error`, and then reads `error.message` and `error.name`. `TypeError` and the rest wait for library failures to be catchable (below), and subclasses of `Error` for class inheritance.

### Two ways to unwind, measured

1. **Cleanup paths.** One pending-exception word in the runtime. `throw` sets it and jumps to its frame's handler: the innermost `try` around it in the same function (letting go of what every scope between holds, and the statement's temporaries), or, with none, the function's way out, which lets go of everything the frame holds and returns a zero value. After every call to a function that can throw, one test of the word, and the same jump. Which functions can throw is known from the call graph, so a function that can't pays nothing, and a program without `throw` is the code it was. `finally` is emitted at every way out of its `try`: falling through, `return`, `break`, `continue` and a throw, as the emitter already emits releases at every way out of a scope.
2. **`setjmp` and a cleanup stack.** Each `try` is a `setjmp`, and a throw is a `longjmp` to the innermost one. Since `longjmp` skips the frames between, each of those frames must have put every reference it owns on a runtime cleanup stack as it took it, and taken it off as it let go, so the throw can release what's on the stack above the `try`'s mark. A call costs nothing, but every owned reference in any function a throw might pass through costs a push and a pop, and every `try` entered costs a `setjmp`.

Measured, on a shared 4-vCPU cloud container (Intel Xeon @ 2.80GHz, load about 1, clang 18.1.3 `-O2`, best of 7 interleaved rounds; noisy, so under about 5% means little):

- **In the mechanisms alone** (`bench/unwinding/unwinding.c`, plain C, every frame able to be thrown through, 4,000 rounds of a recursion 14 deep):

  | Workload | no exceptions | cleanup paths | `setjmp` and a cleanup stack |
  |---|---|---|---|
  | trees, allocating (as `trees.ts`) | 2.492 s | 2.520 s (+1%) | 2.668 s (+7%) |
  | calls, no allocation, tiny frames | 0.258 s | 0.302 s (+17%) | 0.289 s (+12%) |

  Throwing from the deepest frame every 64th round costs neither anything measurable. On tiny frames the cleanup stack's push and pop is cheaper than a test after each of two calls; once a frame does real work, both vanish into it, the cleanup paths more completely.
- **In what Adamic emits**, with every function marked as one a throw can leave, so every call is followed by its test (a local build, not committed): `trees` 4.211 s against 3.966 s with the tests, `nbody` 1.083 against 1.074, `spectral_norm` 0.462 against 0.469, `sort` 0.464 against 0.483, `word_count` 0.233 against 0.226. Every difference is inside the noise, either way.

**Decided: cleanup paths.** In real code their cost didn't show at all, and only a function a throw can leave pays it; a program with no `throw` is the code it was. They keep the C structured, so clang optimizes it and the sanitizers see every release, where a `longjmp` skips past frames the count then has to be told about, and past the runtime's own loops (`sort`, a map's iteration) holding references of their own. The cleanup stack won one microbenchmark, on frames that do nothing but call; if real code ever shows the test, the call graph is the place to shave it (a callee that can't throw needs none).

### Staying byte for byte with Node

- **Uncaught.** An error no `catch` takes ends the program as a panic: `adamic: panic: ` and `String(error)` (`Error: boom`, or `Error` when the message is empty), exit 70, everything written to stdout before it flushed first. That's what the oracle's runner already makes of an uncaught throw on Node, so the native binary, the JavaScript backend and the source agree byte for byte. A `finally` runs on the way out, on both sides, before the panic is written.
- **The JavaScript backend** emits JavaScript's own `try`, `catch`, `finally`, `throw` and `new Error`.
- **A panic is never caught.** `panic`, and every check the compiler inserts, ends the program where it stands. On Node the oracle's runtime throws to stop the program, which a `catch` in the source could take; so `panic` there writes its line at once, then silences everything the program writes after it and makes the exit 70, so a `catch` or `finally` that runs after a panic on Node can't be seen, as it never runs natively.
- **Library failures are panics, for now.** In JavaScript, `'x'.repeat(-1)` throws a `RangeError` a `catch` can take; natively it panics, with the same text. Until the runtime raises those through the same pending word as `RangeError` objects (the next step), a `try` that can reach one whose argument isn't a constant that can't fail (`repeat`, `toFixed`, `toPrecision`, `toExponential`, `toString(radix)`, `normalize`, `new Array(length)`, `Array.from({ length })`) says NotYet. Running out of stack, a string past the length limit, and the temporal dead zone throw from nearly anywhere; inside a `try` they stay panics natively while Node could catch them, which is noted here and not refused.

### The first cut

`throw new Error(message)`, a caught error thrown again, and `try` with `catch`, `finally` or both, wherever statements go (functions, methods, constructors and the top level). A `catch` may leave its binding out. A function value (an arrow function) may throw, and so may a sort's comparator, named or written in place. Which function value a call reaches isn't known, so when any function value in the program can throw, every call through one is followed by the test of the word (`ir.Program.ClosuresMayThrow`, worked out with `MayThrow` until neither changes). The loops the emitter writes around a callback (`map`, `filter`, `find`, `some`, `forEach`, `reduce`, `Array.from`, and `Map` and `Set` `forEach`) test it after each call and let go of the element they hold across it; what they made so far is the statement's temporary and goes with the rest. `map` never writes over its array in place for a callback that can throw. The sort is the one loop in C: a comparator that throws `longjmp`s out of the sort's own frames, which hold no references, to a `setjmp` in a frame whose locals nothing changes after it, and the sort lets go of the references it took (a snapshot, since a merge halfway done may hold one twice and another not at all) and writes nothing back, so the array is as it was, as V8 leaves it. `return`, `break` and `continue` inside a `finally` override what the `try` was doing, a return, a throw, a break or a continue: the held result and the pending error each wait in a scope of the finally's, so leaving it lets go of them, and the flow graph routes the jump out from the finally's own block. Held by `closures_throw.a`, `closures_throw_uncaught.a` and `finally_leaves.a`.

### Exceptions in the graph

The analyses that move and reuse values (`internal/flow`, and reuse in place on it) must see every way a statement can end, a throw included, or they prove a value dead that a `catch` still reads. Reviewer R's round 7 found the graph had no edges for exceptions at all: on the integration that brought `try` in, any program with `try` or `throw` failed to compile (`flow: no graph for ir.Try`).

- **The edges.** An instruction with a call to a function that can throw (`ir.Function.MayThrow`) ends its block with a `MayThrow` terminal: one edge to what follows, one to the handler. `throw` goes to the handler. The handler is the innermost `catch`, or, without one, the `finally`, or, with no `try` around it, a block ending in `Throw`, the function's exceptional exit, which a trace counts as leaving as `Return` does. A `finally` ends in a `Choose` over every way into it can go on: after the `try`, the next handler out (when a throw is being carried through it), and each `return`, `break` or `continue` routed through it. Liveness then sees that a `catch` still reads what is live there.
- **What reuse learned.** A variable that a statement assigns is only overwritten (and its old value dead) when the statement can't throw; one that can throw leaves the old value for whoever takes the throw. A global is never moved into a call in a statement that can throw: a catch, in this function or a caller, may read it and find it NULL. And a temporary handed to a consumed parameter stays the statement's own to let go until every argument is evaluated, since an argument after it can throw and then the call is never made.
- **How it's held.** The trace check's runtime stops at a panic (a Node panic is a throw a catch could take, so the trace after it isn't native's), and the graph's paths must include every path Node took through `try`, `catch` and `finally`. R's probe `move_throw.a` is a fixture: a global moved into a call that throws, a local the catch reads, a spread whose field throws, a dead-looking local the catch reads, and a field taken out of a reused spread before a later field throws. The mutant that drops the `MayThrow` edges fails it under UBSan (a NULL read in the catch) and fails `internal/flow`'s own tests.

**When regions meet exceptions (R, round 8).** A statement with a region that a throw leaves ends the region on its way out. `checkThrown` lets go of the statement's temporaries, then ends the statement's region, then jumps. Otherwise the region's blocks, and the heap strings its values hold, leak. Bringing the two together turned up three joins that broke without a word:

- A call handed a region returned early, before the test of the pending exception that follows any call that can throw. A throw out of a region version went unseen: native went on with a NULL result, and the program printed the wrong totals.
- The region's escape analysis didn't know `ir.Defined`, and took a narrowed parameter's field reads as keeping it. No region was planned anywhere.
- Lent reads didn't count `ir.Defined` as pure, so an operand under it was retained. The counts table showed it: `regions.a` went to 0 in regions, and lent reads went quiet across the table.

`regions_throw.a` throws out of a region version mid-statement, after nodes are made in the region: at the top level inside a `try`, out of a function into its caller's catch, and through a `finally`, each followed by a region that finishes. Mutant: no region end on the throw path, caught by the leak check (8,627 bytes in 9 allocations: the region's blocks and its label strings).

## Arenas

Some work allocates a lot and frees it all at once: one request, one file checked by cohere. For that, an arena: allocations bump a pointer, and the arena frees everything in one go at the end. A value allocated in an arena must not outlive it, and proving that is escape analysis. The lowering IR's aliasing analysis (#5jck546) is where that comes from. Arenas are for stage 1 (cohere in Adamic), where cohere's own measurements already show that with the collector off, fresh allocation is the cost.

### The design (#272q6cv, on paper)

Status: **built** (`internal/native/region.go`, `runtime/region.c`), as designed below. What building it measured is at the end of this section. The benchmark it answers is `bench/trees.ts`. It makes 68,332,244 objects, each its own `malloc` and `free`. Node and Bun bump-allocate them in a young generation and drop each dead tree at once. Reuse in place can't help: nothing there is consumed to build its own shape.

**What lives in a region.** A region is scoped to one statement, alive while it runs and freed when it ends. It holds the values the statement creates, directly or through calls, that are dead when it ends. In `total += check(build(depth))`, every node `build` makes is read by `check`, which keeps none of them, and nothing holds the tree after the statement. All of them can be allocated in the statement's region and freed with it: no `free` per node, no count reaching zero node by node.

**Which values those are: three facts, two of them interprocedural.**

1. **Where a function's result comes from.** A function's allocation site is **returned-only** when what it makes flows only into the function's return value, or into fields and elements of other returned-only values. That's a question about the alias graph (`internal/flow`): every edge out of the site's value is an `Assign` or `Capture` into the return, or into another returned-only value. A function whose returned value comes only from returned-only sites, and from calls to such functions, **returns fresh**. In `build`, both object literals are returned-only, and its recursive calls feed its return, so `build` returns fresh. That's a fixed point over the call graph, recursion included.
2. **Whether a parameter escapes.** A parameter escapes when its value can be stored where it outlives the call: a global, a captured variable, a field or element of anything not itself made in the call, the return value, or a callee's escaping parameter. `check` only reads its parameter's fields and returns a number, so its parameter doesn't escape. This is the escape the effect inference already tracks (`infer.go`'s escaped set), refined from "handed to any call escapes" to "handed to a callee whose parameter escapes".
3. **Whether the statement's fresh values die with it.** Liveness (`flow.LiveOut`) says nothing reads the variable they're in after the statement. The alias graph says they're never stored into an escaped value. A fresh value passed to a call reaches only that call's non-escaping parameters.

**How a region reaches the allocations: as a parameter, not a global.** Which region a value belongs to is a property of the call, not of the allocation site, so the region is passed explicitly, as in Tofte and Talpin's region inference.

- A function that returns fresh takes a hidden `adamic_region *` parameter.
- Its returned-only sites allocate from that region, or from the heap when it's `NULL`.
- Its calls that feed its return pass the region on. Every other call passes `NULL`.

A "current region" global would be simpler and wrong. A function called during the statement for another reason, say one that stores a fresh object into a global, would allocate it in the region, and it would dangle once the region ended.

**Counting in a region.** A value made in a region is immortal while it lives: its count is 0, so retain and release skip it, as they do the program's constants. That's already how the runtime treats a constant, so no runtime path learns anything new. When the statement ends, the region walks its values and releases what they hold outside the region (a heap string in a field, say), then frees its blocks in one go. A region value's reference to another region value is a release of an immortal, which costs nothing.

**What could go wrong, and what catches it.** The one failure is a region value still reachable after its region ends: an escape the analysis missed. In the sanitized build the region's blocks are real `malloc` blocks, freed at the statement's end, so a dangling read is an ASan use-after-free on the oracle's run. The mutants this must be held by:

- A region for a statement whose fresh value is assigned to a variable live after it.
- A callee whose parameter does escape (stores it into a global) treated as non-escaping.
- A returned-only site whose value is also stored into a field of a parameter.
- The region passed down to a call that doesn't feed the return.

Each needs a fixture that builds its strings and objects at runtime. Each must fail on ASan or the leak check, and nothing else.

**What the counts table should show.** A new column, **in regions**: values made in a region and let go with it, not freed one by one. `allocations` stays every value made, `frees` becomes the values freed one at a time, and a finished program satisfies allocations = frees + in regions. For `trees`:

- **In regions:** the nodes of every `check(build(depth))`, 66,759,382. That's 68,332,244 less the stretch tree (2^20 - 1 = 1,048,575 nodes) and the long-lived tree (2^19 - 1 = 524,287), which globals hold.
- **Frees:** 1,572,862, those two trees.
- **Retains and releases:** unchanged as calls. On a region value they cost a branch and touch no count. Removing the calls for values statically known to be in a region is a later step.
- **Peak live:** unchanged. A region frees at the statement's end, which is when each of those trees died anyway.

**What building it measured.** The design held, with one correction to its prediction. On `bench/trees.ts`, counted:

| | allocations | frees | in regions | retains | releases | peak live |
|---|---:|---:|---:|---:|---:|---:|
| Before | 68,332,244 | 68,332,244 | 0 | 272,979,306 | 204,647,140 | 2,097,149 |
| After | 68,332,244 | 1,572,900 | 66,759,344 | 272,979,306 | 204,647,140 | 2,097,149 |

- **In regions:** 66,759,344 is exactly the nodes of every `check(build(depth))`. It's the sum of the counts the program prints for its eight depths.
- **The correction:** the prediction above (66,759,382 in regions and 1,572,862 frees) took every allocation besides the two long-lived trees for a node. It forgot the 38 strings the program builds to print, which are freed one at a time.
- **Unchanged, as designed:** retains, releases and peak.

**Time barely moved.** `go run ./bench -only trees -rounds 5`, best of 5, in the noisy cloud container (load about 4):

| | native | Node | native vs Node |
|---|---:|---:|---:|
| Before | 2.727 s | 1.602 s | 1.70x |
| After | 2.519 s | 1.464 s | 1.72x |

Native's 8% is the same as Node's own run-to-run swing. Native's memory stayed at 97.9 MB, because the two long-lived trees are on the heap.

So 66.8 million `malloc`s and `free`s were cheap here: glibc's small-object cache serves same-sized nodes quickly. What's left is, by inference rather than profiling, the 273 million retain calls. Each is a call into the runtime even when it does nothing for a region value, plus the recursion itself. Not emitting retains and releases on values statically known to be in a region is the next step for this benchmark.

**How it's held.** `regions.a` puts one honest region beside every way a value could outlive its statement:

- A callee keeps its parameter, keeps something read out of it, returns it, or stores it into another object.
- A function keeps what it makes as well as returning it, or keeps a fresh value it doesn't return.
- A fresh value is declared into a variable.
- One statement has a region and also makes a value that escapes.

Each kept value is read after its statement. Mutants, against the whole oracle:

| Mutant | Caught by |
|---|---|
| No parameter escapes | ASan heap-use-after-free |
| A fresh function hands its region to calls that don't feed its return | ASan heap-use-after-free |
| Every fresh call in a statement that has a region gets it | ASan heap-use-after-free, on the one-statement-two-values line |
| Every statement with a fresh call gets a region | ASan, in `regions.a` and the existing `weak_parent.a` and `08_results.ts` |
| A region's end doesn't release what its values hold | the leak check |
| A region's end doesn't free its blocks | the leak check |

Freeing the blocks at the region's end is what poisons them: the blocks are ordinary `malloc` blocks, so ASan sees any later read as a use after free.

**What it leaves for later.**

- A region per loop iteration, per function call (a value made and dropped inside one call), and per request (stage 1's cohere, a file at a time).
- Arrays in regions. Their element buffers grow by `realloc`, which a bump allocator can't do in place.

### Region values, uncounted

A value in a region needs no count, but a fresh function can't know whether its caller handed it a region or the heap, so every retain and release on what it makes was still a call, a branch inside the runtime that did nothing. So a fresh function is now emitted twice (`region.go`, `regionVariants`):

- **The heap version** is the function as it was before regions: plain `adamic_object_new`, its calls to the heap versions of other fresh functions. A call handed no region calls it.
- **The region version** (`_in`) takes a region that is never NULL. Its returned literal is made there, and the calls feeding that literal, or its return, go to region versions in turn. So every one of those values is statically in the region. Each is held in a temporary the statement doesn't own (`regionValue`), so it takes no release at the statement's end, a field holding it is written without a retain, and returning it takes none. Anything else a region literal holds, such as a heap string, is still retained, and the region's end releases it.
- **The statement that has a region** calls the region version with `&region`, and doesn't own what it returns either.

Measured on `bench/trees.ts`, counted:

| | allocations | frees | in regions | retains | releases | peak live |
|---|---:|---:|---:|---:|---:|---:|
| Before | 68,332,244 | 1,572,900 | 66,759,344 | 272,979,306 | 204,647,140 | 2,097,149 |
| After | 68,332,244 | 1,572,900 | 66,759,344 | 139,810,138 | 71,128,452 | 2,097,149 |

Retains and releases each fell by about 133 million. That's three of each per inner node (two fields and the return, against the two call results and the literal) and one per leaf (the return, against the literal), over the 66.8 million region nodes. Nothing else moved. What's left of the retains is mostly `retain(NULL)` for a leaf's two `undefined` fields, plus the long-lived trees, which are on the heap.

Time, `go run ./bench -only trees -rounds 5`, best of 5 at load about 2.6: native 1.968 s against Node's 1.423 s, so **1.38x Node, down from 1.72x**. Interleaved against the previous commit's binary, best of 5: 2.504 s before, 2.004 s after. Memory is unchanged at 97.9 MB.

In the oracle's table, only `regions.a` moved: retains 624 to 517, releases 708 to 596.

Two mutants, each run against `regions.a` and each caught:

| Mutant | Caught by |
|---|---|
| A heap call to a fresh function is taken as in a region (its result uncounted) | the leak check |
| A region literal holds every field without a retain, its heap strings included | ASan heap-use-after-free, on a label string the region's end let go of |

### Cheaper statement-region end (region-end)

Built on main `50045bd`, separately from the unmerged iteration arenas.
`07efb19` changes only existing statement regions. The shared allocator is
inline so its two entry points incur no extra helper call at release `-O2`
(checked in clang assembly). heap.c and count.c were untouched. Region
planning's freshness/escape rules and heap-object zeroing stay unchanged.
The supplied M4 Max sample motivated this unit; measurements below are fresh
Linux observations, not a reproduction of that macOS profile.

**Outside children decide whether there is a release walk.** A region starts
with holds_outside false. A region literal's same-region fields and constant
undefined need no marker. For each other reference field, the emitter calls
adamic_region_hold before storing it, whether kept retains or moves the
reference. A non-NULL reference with a nonzero count sets the bit; dynamic
immortals leave it clear. When the bit is clear, region_end does not visit any
object or field. When set, it keeps the existing child-release walk, while
all blocks still exist. Main already filters immortal/regional children in
class_inheritance.c's release_field, so removing the walk changes no release
counts. The bit and allocator choice are metadata, not counted allocations.

A nonescaping consumer can still write a heap child into its parameter.
region.go therefore computes a separate conservative fixed point over direct
and virtual call targets: a reference property write, callback/opaque
operation, or a callee containing one forces the statement's bit before it
runs. This summary does not grant any new allocation an escape proof. Reads
and plain literal allocation are harmless; unknown operations force cleanup.
It can conservatively retain a walk for a write to an unrelated object.
Class allocation always forces cleanup, since subsequent constructor stores
are not restricted to same-region references.

**Weak cleanup is independent.** weak.c's adamic_weak_forget_region checks
whether the table is empty once, then visits its slots and invalidates targets
whose addresses are in this region's used block ranges. Removing the last
entry frees the table and resets its capacity, which ends the loop. The
release walk runs first, since freeing outside children can remove handles.
Pure regions still invalidate Weak targets before freeing blocks. With no
Weak, there is one empty-table check per end instead of one call/check per
object. With Weak, this is a table-capacity scan with block-range membership
checks; dense Weak workloads were not timed here.

**Only a fully filled ordinary literal skips zeroing.** Its field expressions,
conversions and possible throwing calls finish before allocation. Then the
emitter uses adamic_object_new_filled_in and writes every shape slot's active
member. Only nonthrowing retains, the marker and scalar stores occur between
allocation and completion; nothing publishes the object in that interval.
A throw in a field expression therefore finds no half-filled parent object.
Earlier completed children are still well-formed and cleaned on the existing
throw path. Constructors conservatively use adamic_object_new_in, which
zeroes slots and forces cleanup; spreads and heap allocation keep their old
paths. The filled allocator is not permission to move allocation ahead of
field evaluation.

**How it is held.** region_end.a has pure trees, runtime string children, a
transitive nonescaping consumer that replaces a field with a heap string,
a later field that throws after a child has been built, and a successful
region after the catch. It runs against Node through the oracle. The current
planner refuses escaping Weak stores, so TestRegionEndWeakTargets exercises
that runtime contract directly: targets in different blocks, an unrelated
heap target, and a region whose target is the last entry in the Weak table.
A missed target is actually read after the region ends, not merely compared
with NULL. TestRegionEndThrowInitialization also runs the .a witness against
Node under both sanitizers with an aligned malloc fill byte of 240, so an
alignment diagnostic does not mask the ASan invalid read.

| Mutant, restored after its run | Caught by |
|---|---|
| adamic_region_hold does not set holds_outside | LeakSanitizer: regions.a leaks 3,864 bytes in 56 allocations; region_end.a also leaks |
| batch Weak cleanup omits regional targets | ASan heap-use-after-free in TestRegionEndWeakTargets |
| filled literal allocation moved before its field expressions, across a throw | ASan SEGV invalid read in release_field in TestRegionEndThrowInitialization |
| later consumer reference writes do not force cleanup | LeakSanitizer: region_end.a leaks 73 bytes in one allocation |

These are targeted mutant runs, not whole-oracle mutant runs. The zeroing
mutant first failed the ordinary oracle under UBSan's alignment check; the
aligned-fill run above supplies the requested ASan witness. No mutant is
credited only for an answer difference or a compiler failure. The ordinary
new fixture and existing region/throw/constructor-capture fixtures passed.
Regenerating the counts table added only the new fixture's row; every
existing row remained byte-for-byte unchanged.

**Measured 2026-10-06.** Before is main `50045bd`; after is `1921f57`.
Both use identical sources and release flags (`native.Flags`, clang 20.1.8,
`-O2`, no LTO). CSS is absent on this main too: use the frozen `6476d7a` port
and `ccfd64f` css-release evidence from the previous survey. Neither side
includes ccfd64f's separate release fast path. CSS consumes the same 4,952
shared inputs with `print_main.ts shared.txt count`; JSON consumes the same
1,096 inputs with `main.ts --cases cases.txt`, producing 63,704,214 bytes.
Trees is the existing depth-18 benchmark, 68,332,244 allocations; regions.a is
one ordinary fixture run, including its three-round small-tree loop.

Counts, one counted build/run per snapshot (after rebuilt at the inline commit):

| driver | version | allocations | frees | in regions | retains | releases | peak live |
|---|---|---:|---:|---:|---:|---:|---:|
| trees | Before | 68,332,244 | 1,572,900 | 66,759,344 | 67,982,686 | 67,982,728 | 2,097,149 |
| trees | After | 68,332,244 | 1,572,900 | 66,759,344 | 67,982,686 | 67,982,728 | 2,097,149 |
| regions.a | Before | 316 | 260 | 56 | 219 | 409 | 50 |
| regions.a | After | 316 | 260 | 56 | 219 | 409 | 50 |
| CSS parse/print | Before | 23,347,421 | 23,347,421 | 0 | 85,979,081 | 89,181,243 | 145,450 |
| CSS parse/print | After | 23,347,421 | 23,347,421 | 0 | 85,979,081 | 89,181,243 | 145,450 |
| JSON format | Before | 104,362,078 | 104,362,078 | 0 | 302,478,215 | 301,718,362 | 13,759,795 |
| JSON format | After | 104,362,078 | 104,362,078 | 0 | 302,478,215 | 301,718,362 | 13,759,795 |

All six columns are unchanged on every measured driver. Trees' 66,759,344
regional values now avoid both the object/field walk and slot zeroing; its
remaining retain/release calls in check are outside this unit. regions.a's
regional nodes hold runtime heap labels, so they still require child cleanup.
CSS and JSON execute no statement regions on either snapshot: this unit does
not address the real drivers' missing whole-input coverage. JSON's generated
C is byte-for-byte identical; CSS's changed region variants have no executing
caller in this corpus.

Time is best of five alternating before/after pairs. The native C fork/wait4
launcher measures process startup, file I/O and output; hash comparison and
counting are outside the timer. RSS is the native child's peak on its fastest
run, not the Python driver's inherited peak. All test/build/sanitizer jobs
finished before timing. Shared Linux x86-64 EPYC 9V74, five visible CPUs with a
four-CPU quota, 16 GiB cgroup memory limit, Go 1.27.1 and Node 24.19.0. Load is
the one-minute average at each driver's start and end. Setup reported Go,
clang, Node and submodules ready at 0 s, cache warm and done at 66 s; nproc 5.

| driver | Before seconds | After seconds | time change | Before RSS MiB | After RSS MiB | load, start to end |
|---|---:|---:|---:|---:|---:|---:|
| trees | 1.631854 | 1.283309 | -21.36% | 129.0 | 129.0 | 0.04 to 0.25 |
| regions.a | 0.000752 | 0.000763 | +1.48% | 0.8 | 0.8 | 0.25 to 0.25 |
| CSS parse/print | 1.801035 | 1.775470 | -1.42% | 23.4 | 23.3 | 0.25 to 0.47 |
| JSON format | 11.742555 | 11.893889 | +1.29% | 1338.9 | 1338.9 | 0.47 to 0.93 |

Trees improves 21.36% with unchanged rounded RSS. CSS's -1.42% and JSON's
+1.29% are observations with overlapping run ranges, not evidence of a
formatter benefit or a new lifetime cost. The sub-millisecond regions.a
+1.48% is especially dominated by startup and scheduling. No new M4 Max
profile or dense-Weak timing is claimed; constructors remain conservative.

**Checks and reproduction.** Formatting and `go vet ./...` passed. The full
uncached repository gate passed at 07efb19, including all stage1 ports. After
the inline-only refinement, the complete native/fresh/oracle gate passed
uncached again at 1921f57, and all four mutants were rerun with the catches
above. The final measured CSS build passed all 24,076 archived Go inputs in
both print-option sets under ASan, UBSan and LeakSanitizer, empty stderr,
3,172,898 and 3,227,590 identical bytes. Final JSON passed those sanitizers on
all 1,096 measured inputs, empty stderr and 63,704,214 Node/Go-identical bytes.
Every timed and counted answer matches the fresh Node answer hash.

```
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/region-end/full-gate.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/native/... ./internal/fresh/... ./internal/oracle > /tmp/region-end/final-gate.log 2>&1
go test -count=1 -timeout 30m ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts > /tmp/region-end/counts-update.log 2>&1
```

Build each compiler with `go build -o <snapshot>-adamic ./cmd/adamic` from
its revision, then each existing entry with `adamic build <entry> -o <binary>`,
and once with `--count`. For trees use bench/trees.ts; for the fixture use
internal/oracle/testdata/regions.a. Recover CSS and the fixed corpora as in
ccfd64f's css-release report. CSS shared input SHA256 is
faf62ec1e054167e192cdb427fd689dfde759dfa0cbaa9dc59253d54bd42eef7;
JSON input SHA256 is
bc394e563cbb5a4dccc264285c69825fabc21552b56ac6b13b7708008f26fba1.
The final compiler snapshot was rebuilt after inlining, with counts and
answers rechecked before timing.

Artifacts are in `/tmp/region-end`: results-final.json records all five pairs,
RSS, loads, exact arguments, counts and output hashes; timing.log keeps every
round. prepare.py, rebuild-after.py, timing.py and sanitize-corpora.py retain
commands. mutants.py and mutants-final-run.log retain the four targeted
mutations, controls and restoration. full-gate.log, final-gate.log,
sanitized-corpora-final.log, vet-final.log and gofmt-final.log retain checks.
The sanitizer script links the exact native.Flags archive (keyed by all
runtime inputs, flags and clang identity). No runtime source hook outside
region.c, weak.c and adamic.h was needed.

### No retain of the constant undefined

Retain passes over the null pointer, but each `adamic_retain(NULL)` was still a call. Every place the emitter retains a value it keeps now goes through `retained` (`emit.go`). When the value's C is the null pointer constant, however it's parenthesized or cast (an `undefined`, a missing argument, `undefined` boxed into a union), it is kept as it is, with no call. Before this, the 0.1 fixtures and the oracle's held 50 such calls: 27 into object fields, 13 into globals and locals, 4 into array elements, 6 elsewhere. None are left.

Measured on `bench/trees.ts`, counted, against 926feae:

| | retains | releases |
|---|---:|---:|
| Before | 71,827,454 | 3,145,768 |
| After | 3,145,726 | 3,145,768 |

That's 68,681,728 fewer retains: the two `undefined` children of every leaf. Nothing else moved. The merge of main had already taken `trees` from 139,810,138 retains and 71,128,452 releases to the "before" row: lent reads now see through `ir.Defined`, so `check` reads its tree's children without a count.

**Releases didn't fall, on any row, and can't.** 22 rows of the oracle's table moved, each by retains alone, from 1 (`main.ts`) to 112 (`regions.a`). Allocations, frees and peak moved on none. A retain of the null constant has no counted release to match. When it went into a field or an element, the slot is let go of by the runtime when its holder is freed (`let_go` in `heap.c`), which isn't a counted release. When it went into a variable or a global, the release that comes later is of whatever the variable holds by then, and that release still happens.

Two mutants:

| Mutant | Result |
|---|---|
| Any C containing `NULL` taken for the constant | changes no fixture's C: no value that reaches `retained` contains `NULL` but the constant, so it proves nothing |
| Any cast pointer, `((T *)(name))`, taken for the constant | changes only `unions.a`, where a string or object boxed into a union loses its retain: ASan heap-use-after-free |

### Moving a statement's temporary into what keeps it

With regions uncounted and `undefined` free, what was left on `trees` was the heap version of `build`, for the two long-lived trees. Measured first: 3,145,722 of the 3,145,726 retains were its own. Each inner node retained its two children's call results into its fields and itself into its return, three retains, then let go of the three temporaries when the statement ended, three releases. Each leaf did it once, for its return. 786,430 inner nodes times 3 plus 786,432 leaves is exactly that.

A temporary the statement owns, handed to a place that keeps it, is now moved there with its count (`kept` in `emit.go`): no retain, and the statement doesn't let go of it at its end. Only two places qualify, because they can't let go of what they hold before the statement ends:

- **A field of an object literal**, which the statement made and owns. This covers a spread, reused or copied, too.
- **What a function returns.**

A variable or a global doesn't qualify. A call later in the same statement could assign it, letting go of the value while the statement still reads it.

On `trees`, counted: retains 3,145,726 to 4, releases 3,145,768 to 46, the same 3,145,722 off each. Allocations, frees and peak didn't move. In the oracle's table 63 rows moved, every one by equal retains and releases, 1,001,548 of each in all.

Time, `go run ./bench -only trees -rounds 5`, best of 5 at load about 3: native 1.512 s, Node 1.518 s, **1.00x Node**, at a third of Node's memory (98.0 MB against 300.9). Interleaved against the commit before this one: 1.625 s to 1.559 s.

Two mutants, each run against the whole oracle and each caught by ASan alone:

| Mutant | Caught by |
|---|---|
| A value the statement doesn't own is moved too, so nothing retains it | heap-use-after-free, in 28 fixtures |
| An owned temporary is moved but still let go of at the statement's end | heap-use-after-free, in 63 fixtures |

### Declarations, assignments and `??` take what the statement owns

`nbody` made 8 objects and 52,000,118 retains. Read in its C first: every `const other = bodies[j] ?? new Body(...)` retained the element into a temporary, retained that again as what `??` makes, and retained it a third time into `other`. The statement then let go of the first two, and `other`'s scope of the third. That's three pairs where one does. With `body` the same way, it comes to about 51 pairs a step.

Three moves, each where nothing can run between the move and the end of what owned the count:

- **A declaration or a statement's assignment** takes the statement's own temporary (`taken` in `emit.go`). The store is the statement's last write, so nothing after it can assign the variable while the statement still reads the value. That's the case `kept` left out: an assignment inside a larger expression, where a later call could.
- **`??` passes its left operand on** when the statement owns it. When the left operand isn't there, it's NULL, with nothing to let go of.
- **`??` passes a fallback it made on** with its count, in its own branch.

Counted:

| | retains | releases |
|---|---:|---:|
| `nbody` before | 52,000,118 | 52,000,122 |
| `nbody` after | 22,000,050 | 22,000,054 |
| `word_count` | 7,140,552 to 6,100,372 | 6,140,609 to 5,100,429 |
| `tokenizer` | 2,407,586 to 2,405,063 | 2,402,543 to 2,400,020 |

`trees` lost its last 2 of each, and `sort` 2 of each. In the oracle's table 114 rows moved, 1,025,184 retains and 1,025,233 releases in all. Allocations, frees and peak moved on none.

**15 rows fell by more releases than retains**, from 1 more (`string_positions.a`) to 7 (`string_index.a`). That's the second move: when `??` found its left operand absent, the statement used to release that NULL temporary at its end, a counted call. Proof: with a release of the left operand put back in the absent branch, all 114 rows fell by equal amounts.

Time, best of 5 at load about 3: `nbody` 0.393 s to 0.242 s interleaved against the commit before; against Node, 0.241 s to 0.763 s, **0.32x Node**, in 4.0 MB to Node's 72.1. Bun is at 0.127 s. The 22 million retains left are one per element read, the count a variable keeps on what it was given. Borrowing a local from an array nothing can change while the local lives is the next step there.

Three mutants, each run against the whole oracle and each caught by ASan alone:

| Mutant | Caught by |
|---|---|
| An assignment takes a value the statement doesn't own, so nothing retains it | heap-use-after-free, in 3 fixtures |
| `??` passes its left operand on, and the statement still lets go of it | heap-use-after-free, in 19 fixtures |
| `??` passes its fallback on, and still lets go of it | heap-use-after-free, in 3 fixtures |

### A variable borrowed from an array (designed)

What's left on `nbody` is one retain and one release per element a variable is given: `const other = bodies[j] ?? new Body(...)` keeps a count on the object for as long as `other` lives. The array already holds that object, so if nothing can take it out of the array while `other` lives, the variable can borrow it as a parameter borrows from its caller.

**What is borrowed.** A local `v`, declared with `const` or never assigned after its declaration, not captured by a closure, whose initializer is `a[i]` or `a[i] ?? fallback`. Here `a` is a local or parameter of the same function that is never assigned anywhere in it and isn't captured, so `a` names the same array, alive, for the whole function. A global or a captured `a` doesn't qualify: a call can assign either.

**The proof: nothing in `v`'s live range can change the array.** `v`'s live range is the rest of the block that declares it, every nested statement included. Every expression and statement there must be one that can't take an element out of any array, through any alias. That's a fixed list, never a guess about what's safe:

- **Allowed:** what `pure` in `borrow.go` allows (reads, arithmetic, field and element reads, string work); a field store into an object (`SetProperty`, which lets go of the field's old value, never an array's slot); declarations and assignments of other variables; a `push` (it adds, and moving the buffer doesn't move the objects); an Error allocation, with its message and name expressions checked too; and direct, nonvirtual calls to named functions whose bodies, and everything they call, are made only of allowed things.
- **A change, which stops the borrow:** a store into an element (`SetIndex`); `pop`, `shift`, `splice` (also discarded, `adamic_array_remove`), `fill`, `sort`, `reverse`; anything reuse could do in place to an array (`map`, a spread of an array, `[...a, x]`); any call through a function value, and every runtime operation that calls one (`map`, `forEach`, `reduce`, `filter`, `find`, a `sort` comparator, `Array.from` with a function); a call to a named function that does any of these, at any depth; every virtual call, whose static signature does not prove what an override does; and an assignment to `a` itself.
- **A throw.** A throw out of the live range ends `v` without a release. A borrowed `v` has none to give, so a throw path needs nothing. A call that can throw is judged by what it does, like any other.

**What else must know.** A borrowed `v` holds no count, so nothing may treat it as owned:
- reuse in place must not take it over (its object's count is the array's, so it looks unique);
- a call must not move it into a consumed parameter;
- its scope must not release it.

It's marked like a borrowed parameter, and reuse already refuses those.

**The fallback.** With `a[i] ?? fallback`, when the element is missing `v` is the fallback, which is fresh and owned. A hidden owner holds it, NULL when the element was there, and the scope lets go of the owner. So a borrowed `v` with a fallback still costs a release call per declaration, of NULL in the common case, but no retain and no count traffic on the object.

**Probes, each a use-after-free or wrong output if the proof is wrong** (`borrow_element.a`, strings built at runtime):

1. `a[0] = other` while `v = a[0]` lives, then `v.label` read.
2. `a.splice(0, 1)`, and `a.pop()`, while `v` lives.
3. A named function that stores into the array, two calls deep, called while `v` lives.
4. A closure that stores into the array, called while `v` lives.
5. `a` reassigned while `v` lives, with the old array's last count in `a`.
6. `a.map(...)` while `v` lives, where `a` is dead after, so reuse maps it in place.
7. `{ ...v, label }` while `v` lives: reuse in place would write into the array's own object. The result is wrong output, not a sanitizer report, so the oracle's stdout catches it.
8. `v` handed to a function that consumes its parameter.
9. The fallback taken, so the owner has something to let go of (the leak check).

Each probe must keep its count (be refused borrowing). A mutant that drops each rule from the list must fail the probe it guards, by ASan, the leak check, or stdout for probe 7.

**Built** (`element_borrow.go`). The plan runs before reuse's, marks each borrowing variable `Borrowed`, and the declaration is emitted as the element itself with no count. Writing the probes found one thing about the probes themselves: in one function, each probe's later neighbours were changes too, which refused the borrow on their own. So the first run of the mutants left five rules uncaught. Each probe is now a function of its own.

In `borrow_element.a`, probes 1 to 6 are refused. Probes 7 and 8 borrow, and are held by reuse and moves refusing a `Borrowed` variable. Probe 9 and `total`'s loop borrow. One mutant per rule, each caught on that fixture:

| Mutant | Caught by |
|---|---|
| A store into an element isn't a change (probe 1) | ASan heap-use-after-free |
| `splice` and `pop` aren't changes (probe 2) | ASan heap-use-after-free |
| A call is never a change, whatever its callee does (probe 3) | ASan heap-use-after-free |
| A call through a function value isn't a change (probe 4) | ASan heap-use-after-free |
| The array's variable may be assigned (probe 5) | ASan heap-use-after-free |
| `map` isn't a change (probe 6) | ASan heap-use-after-free |
| Reuse takes over a borrowed variable (probe 7) | stdout: `7 item0* item0*`, the array's object written over |
| A borrowed variable is moved into a consumed parameter (probe 8) | ASan heap-use-after-free |
| The fallback's owner isn't let go of (probe 9) | the leak check |

On `nbody`, counted: retains 22,000,050 to 7,000,019. Releases stayed at 22,000,054: the fallback's owner is let go of once per declaration, NULL every time here. No other row of the oracle's table moved, and the other benchmarks didn't either. Time, best of 5 interleaved against 4636a33 at load about 2.5: 0.235 s to 0.222 s, small enough that the noise could hide it. The release of an owner that is NULL is the next thing to take out of `nbody`'s loop.

**The owner's release, only when it holds something.** The scope now lets go of a fallback's owner behind a test for NULL (`mostlyNull` in `emit.go`, read by `releaseScopes`, so the throw path does the same). `nbody`, counted: releases 22,000,054 to 7,000,023, now level with its 7,000,019 retains. In the oracle's table only `borrow_element.a` moved: releases 134 to 129, its five owners that stayed NULL. Time, best of 5 interleaved against 73bbae9 at load about 3: 0.237 s to 0.220 s, again inside what the noise could hide. Mutant: the test inverted (`== NULL`), so an owner holding a fallback is never let go of. Caught by the leak check on probe 9.

### Borrowed element reads checked against current main, October 6, 2026

**Observed on `ef3d907`: the indexed reads in `nbody` already borrow.** No condition of
`borrowable` or `changes` fails for either `body = bodies[i] ?? ...` or
`other = bodies[j] ?? ...` in `advance`, or for the corresponding declarations in `energy`.
The indices may be loop variables; the declarations may be inside loop bodies. The locals and
array parameter are neither captured nor reassigned. `SetProperty` writes the body's fields,
not the array's slots, and is already allowed. The fallback's constructor is a direct call
whose body only allocates and writes fields, so its summary is unchanging. `Math.sqrt` is pure.
`TestNbodyIndexedElementsBorrow` holds all five indexed declarations, including `sun`, to this
answer and to C without releases of those borrowed locals.

The generated C for the inner loop, with its actual names (arithmetic between the shown parts
omitted):

```c
adamic_value *adamic_temporary_54 = adamic_array_at(adamic_local_4_bodies, ((double)adamic_local_23_j));
adamic_object * adamic_local_24_other = adamic_temporary_54 == NULL ? NULL : (adamic_object *)adamic_temporary_54->reference;
adamic_object * adamic_local_24_other_owner = NULL;
if (adamic_local_24_other == NULL) {
    adamic_object * adamic_temporary_55 = adamic_function_4_Body_new((0x0p+00), (0x0p+00), (0x0p+00), (0x0p+00), (0x0p+00), (0x0p+00), (0x0p+00));
    adamic_local_24_other_owner = adamic_temporary_55;
    adamic_local_24_other = adamic_local_24_other_owner;
}
/* Field arithmetic and SetProperty stores use body and other without counting them. */
if (adamic_local_24_other_owner != NULL) {
    adamic_release(adamic_local_24_other_owner);
}
```

In the real run every lookup succeeds, so the owner stays NULL and its release is skipped.
The outer `body` has the same form.

**The 7,000,019 retains have a different source.** Every one of the million `advance` calls
contributes five pairs for its `for...of` bindings, one pair for the array that loop's iterator
holds, and one pair for reading the global `bodies` into the call's temporary. That is exactly
7,000,000 pairs. The remaining 19 retains are five constructor returns, five array pushes,
six from `offsetMomentum`'s iterator and bindings, and three global reads for `offsetMomentum`
and the two `energy` calls. The 23 remaining releases are five constructor-local releases,
five constructor-result releases after the array pushes, six iterator/binding releases in `offsetMomentum`, three
argument-temporary releases, two energy-string releases, the final global array release,
and the release of the initially NULL global array slot when `bodies` is initialized.

The `for...of` path in `emit_statements.go` calls `declareLocal(..., false)` unconditionally;
`declareLocal` in `emit_locals.go` retains and schedules its reference for release. A `ForOf`
is not an `ir.Declare`, so this pass cannot select it, and setting `Borrowed` alone would not
remove that retain or release. Likewise the global argument read is emitted by `read` in
`emit_locals.go`, not by this pass. Removing these pairs needs work outside this unit's
territory. `borrow.go`'s pure-consumer rule must not simply call every function pure: a call
can change a later operand's source. Its consumes list is unchanged.

**Two changes within the territory.** Making an Error (`ir.MakeError`) is now an allowed
operation; the existing expression walk still checks its message and name. A caught throw
can therefore sit between a borrow and its last use, and an exceptional exit can end a borrow
without releasing the array's element. `borrow_element_throw.a` exercises both, a numeric
field write in an indexed loop, a pop in a catch before the last use (refused borrowing), and
a pop through a direct call in the Error's message (also refused).

A virtual call now stops borrowing. The old summary used only `Call.Function`, its static
signature. In `borrow_element_virtual_store.a`, that signature reads only the array's length,
while the implementation replaces its element. Keeping the array alive, and prohibiting its
move into a consumed parameter, does not keep an element alive after an explicit replacement.
The conservative fix checks `Call.Virtual == 0` before consulting the direct-call summary.
Resolving all possible overrides could recover some borrows later; it is not proven here.

**Mutants actually run.** Each began from the final compiler independently, and the source was
restored after every run. Oracle mutants ran uncached, against Node, release C and sanitized C;
each test invocation wrote its complete output to a separate log. No mutant was killed by
clang's warnings or by a lowering refusal.

| Mutant | Check that caught it |
|---|---|
| Drop the `SetIndex` change | ASan heap-use-after-free, `borrow_element.a`, `probe1` |
| Drop the array-variable assignment exclusion | ASan heap-use-after-free, `borrow_element.a`, `probe5` |
| Allow `ArrayPop` and `ArraySplice` | ASan heap-use-after-free, `borrow_element.a`, `probe2` |
| Ignore the direct callee's changing summary | ASan heap-use-after-free, `borrow_element.a`, `probe3`, and `borrow_element_throw.a`, `messageInvalidates` |
| Allow `CallClosure` | ASan heap-use-after-free, `borrow_element.a`, `probe4` |
| Drop `Call.Virtual == 0` | ASan heap-use-after-free, `borrow_element_virtual_store.a`, `inspect` |
| Schedule the borrowed item in `leavesBorrowScope` for scope cleanup | ASan heap-use-after-free, `borrow_element_throw.a`, after its throw and catch |
| Remove the harmless Error allocation allowance | `TestThrowElementBorrowPlan`: both positive borrows missing |
| Disable declaration borrowing | `TestNbodyIndexedElementsBorrow`: 0 declarations instead of 5 |

All nine tests exited 1 under their mutant. The first seven reported ASan heap-use-after-free;
the last two failed their intended assertions. The throw cleanup mutant is deliberately
restricted to `leavesBorrowScope`, so a normal-path cleanup failure cannot mask the exceptional
path being tested. The existing overwrite, reassignment, removal, and closure probes are
separate functions, so one probe's change cannot mask another guard.

**Counts and time.** The final compiler emits exactly the same `nbody` C as `ef3d907`
(`cmp` exits 0). Both counted runs report 8 allocations, 8 frees, 7,000,019 retains,
7,000,023 releases, peak 7 and 0 in regions. No nbody count was removed by this unit.

On this Linux 6.18.44 container, AMD EPYC 9V74, clang 20.1.8, Go 1.27.1 and Node 24.19.0,
`nproc` reports 5 and the cgroup quota is 4 CPUs. `bash cloud/setup.sh` printed: Go ready 0s,
clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 104s, done 104s. The
printed environment file is `/workspace/adamic-tools/env.sh`, sourced for every build/test shell.

Separate `go run ./bench -only nbody -rounds 5` runs gave native bests of 0.704s before
(load 3.47 to 8.48) and 0.426s after (load 1.64 to 1.54). Setup's cache warming overlapped
the first run, and its rounds ranged from 0.704s to 2.129s; those two bests cannot establish
a compiler speedup. With both native binaries built first and then interleaved for five
rounds, alternating which ran first, bests were **0.451295s before and 0.461960s after**,
load 1.20 to 1.18 (one-minute load; five-minute 1.79 to 1.78, fifteen-minute 1.04 throughout).
Both printed `-0.169075164` then `-0.169086185`. The C is identical; the timing difference
is an observation of run noise, not an optimization result.

The whole oracle counts table was regenerated, not just the borrowing fixtures. Every row
that differs from `ef3d907` is below, including new rows. All other rows are unchanged.

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| `borrow_element_virtual_move.a` before | 26 | 24 | 10 | 22 | 12 | 2 |
| `borrow_element_virtual_move.a` after | 25 | 23 | 11 | 26 | 11 | 2 |
| `borrow_element_throw.a` new | 38 | 38 | 34 | 54 | 8 | 0 |
| `borrow_element_virtual_store.a` new | 13 | 11 | 9 | 15 | 7 | 2 |

Why the existing row moves: both invocations of `go` now retain and release `first`, adding
2 pairs. Its array no longer lends, so both virtual calls move the array into their consumed
parameter instead of retaining it, removing 2 retains. In the Mapper invocation, that array
is now unique: the existing map reuse branch takes the array over, adding 1 retain and
2 releases for the replaced slots. The net is **1 more retain and 4 more releases**. No
second array is allocated for the map, so allocations, frees and peak each fall by 1.
The generated C diff shows the two new element retains, the moves that NULL `items`, and
the same Mapper code whose uniqueness branch now runs. The two constructor objects in
regions are unchanged. The new rows count the new throw and virtual-replacement programs;
they have no earlier row to compare.

**Commands and limits.** The focused native checks and uncached borrowing oracle pass:

```sh
go test ./internal/native -run 'TestNbodyIndexedElementsBorrow|TestThrowElementBorrowPlan|TestPassThroughsAreNotConsumers' -count=1 > /tmp/borrowed-native-focused.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/borrow_element' -count=1 -timeout 30m > /tmp/borrowed-oracle.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/borrowed-counts.log 2>&1
```

`gofmt -l cmd internal` printed nothing, `git diff --check` printed nothing, and
`go vet ./... > /tmp/borrowed-vet.log 2>&1` exited 0 with no diagnostics.

The complete gate was also started:

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/borrowed-gate.log 2>&1
```

It was stopped using the worker-gate exception after more than eleven minutes, while the
remaining stage-1 corpus tests were still running (the stopped command's exit was 143).
Before stopping, it reported `ok` for **all tests in internal/native (251.846s) and all tests
in internal/oracle (220.546s)**, plus the bridge (510.450s), flow (143.547s), freshness
(50.080s), lowering (35.021s), loading (1.777s), regexp, fuzz and command test packages.
The full repository gate is therefore **partial**, not a claimed full pass. The touched
packages passed uncached, the complete oracle passed uncached, and the separately filtered
borrowing oracle passed uncached as well. Remaining stage-1 packages were not fully verified
by this unit. The log preserves the completed package results.

Each wrote to a log: `/tmp/borrowed-native-focused.log`, `/tmp/borrowed-oracle.log`,
`/tmp/borrowed-counts.log`. The mutant runner (`ADAMIC_GATE_UNCACHED=1 python3 /tmp/borrowed-mutants.py`) wrote
`/tmp/borrowed-mutants.log` and one
`/tmp/borrowed-mutant-<name>.log` per mutant in the table. Bench logs are
`/tmp/borrowed-before-bench.log`, `/tmp/borrowed-after-bench.log` and
`/tmp/borrowed-paired-bench.log`; full generated benchmark C is
`/tmp/borrowed-before.c` and `/tmp/borrowed-after.c`.

Not covered by this unit: borrowing `for...of` bindings, avoiding the iterator's own count,
borrowing global call arguments, or resolving virtual override effects. Those need different
emission or call-effect proofs. The borrowed declaration rule stays conservative on array
stores, reassignment, removal, closure calls, callbacks and operations it cannot prove.
### Borrowed loop bindings and global call arguments (built, October 7, 2026)

Starting at `ef3d907`, nbody's indexed declarations already borrow. Its remaining
7,000,019 retains come from three emission paths:

| Source | Repeated retains | Why the baseline counts it |
|---|---:|---|
| `advance`'s five `for...of` bindings | 5,000,000 | `forOf` always calls `declareLocal(..., false)`, which retains every reference binding and schedules its release |
| `advance`'s iterator array | 1,000,000 | `forOf` holds the array for the loop, including reassignment and unwinding |
| The global `bodies` argument to `advance` | 1,000,000 | `arguments` calls `value`; a call is not a pure consumer, so `read` snapshots the global with a retain |

The other 19 retains are ten constructor/array-insertion counts, six for
`offsetMomentum`'s iterator and bindings, and three more global arguments.
The baseline release count is 7,000,023. Here is the actual generated C, with
unrelated field operations omitted:

```c
/* Before, advance's loop. */
adamic_array * adamic_temporary_86 = adamic_retain(adamic_local_4_bodies);
for (size_t adamic_temporary_87 = 0; adamic_temporary_87 < adamic_temporary_86->length; adamic_temporary_87++) {
    adamic_object * adamic_local_30_body = adamic_retain((adamic_object *)adamic_temporary_86->elements[adamic_temporary_87].reference);
    /* Read and write body's x, y, z fields. */
    adamic_release(adamic_local_30_body);
}
adamic_release(adamic_temporary_86);
/* Before, the main loop's call. */
adamic_array * adamic_temporary_194 = adamic_retain(adamic_global_2_bodies);
adamic_function_2_advance(adamic_temporary_194, (0x1.47ae147ae147bp-07));
adamic_release(adamic_temporary_194);
/* After: the binding and argument snapshot hold no count. */
adamic_object * adamic_local_30_body = (adamic_object *)adamic_temporary_86->elements[adamic_temporary_87].reference;
adamic_array * adamic_temporary_194 = adamic_global_2_bodies;
adamic_function_2_advance(adamic_temporary_194, (0x1.47ae147ae147bp-07));
```

**The loop proof.** `planElementBorrows` also plans a plain array loop binding,
using `borrowable`'s local, capture, type, and assignment facts, and `changes` on
the entire body. Both the binding and array must be unassigned throughout the
function; this is deliberately stronger than checking only the body. The binding
is marked `Borrowed` before reuse runs, and the array is marked lending. The
emitter gives the binding the element pointer and schedules no release for it.
The iterator retains its array and releases it on every exit. Field stores and
pushes are safe: they do not remove the element. Index stores, removal,
reordering, in-place array reuse, unknown callbacks, closure calls, and mutating
named calls still refuse the borrow. Virtual calls are refused because the
static method's summary does not cover overrides. This uses the same virtual
guard as the preceding borrowed-element unit, now merged into this branch.

`MakeError` itself cannot remove an element, so it is allowed, with its operands
still checked. A throw caught inside the body preserves the borrowed pointer;
a throw leaving the body releases the iterator, never the borrowed binding.
Maps, strings, regex iterators, destructuring bindings, closures, captured
sources, and global sources do not receive the new binding optimization.

**The iterator count.** Its hold remains the one array count this loop explicitly
owns. The million iterator holds in nbody are unchanged. For an iterable already
owned by the statement, `kept` transfers that count into the iterator instead of
retaining it again and releasing the temporary. For a local or borrowed parameter,
`kept` retains as before. Eliminating nbody's iterator hold would need a separate
proof that another owner survives every body exit; this unit keeps the hold.

**The global argument proof.** `lentArgument` accepts only a direct reference
global read handed to a borrowed, unconsumed parameter. Every later argument must
be pure, and `touches` must return false for every `CallTargets` implementation.
It recursively checks reachable calls and refuses unknown closure and callback
effects. It also refuses reads of the global, which is stronger than the needed
no-write proof. Every target must have a borrowed, unconsumed parameter. The
snapshot and any initialization check remain, preserving JavaScript evaluation
order. A return of the parameter takes its own count as before. Captured reads,
pass-through casts and narrowings, unions, consumed parameters, effectful later
arguments, and calls reaching the global keep their counts. Neither `pureKind`
nor `consumes` was widened; `TestPassThroughsAreNotConsumers` passes unchanged.

**Measurements.** clang 20.1.8, Go 1.27.1, Node 24.19.0, Linux 6.18.44,
AMD EPYC 9V74, `nproc` 5, CPU quota 4. `cloud/setup.sh` reported Go ready 0s,
clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 15s,
done 15s. Five before/after native `-O2` runs were interleaved without sanitizers;
one-minute load was 0.39 both before and after. Best times were 0.423612s before
and 0.403722s after, a 4.7% difference small enough for cloud noise to hide.
All ten runs printed `-0.169075164` and `-0.169086185`.

| nbody | Before | After |
|---|---:|---:|
| Allocations / frees | 8 / 8 | 8 / 8 |
| Retains | 7,000,019 | 1,000,011 |
| Releases | 7,000,023 | 1,000,015 |
| Peak live / in regions | 7 / 0 | 7 / 0 |

Exactly 5,000,005 binding pairs and 1,000,003 argument pairs disappeared.
The 1,000,001 iterator holds remain, plus ten constructor/insertion retains.
The release total additionally includes two formatted strings, final global
cleanup, and the initial NULL global replacement.

**Fixtures and mutants.** `borrow_loop.a` builds labels at runtime and independently
exercises index stores, pop, splice, named mutation, closure mutation, an overriding
mutator, reassignment, push, consumed binding reuse, capture, assignment, temporary
iterables, and both caught and escaping throws. `borrow_global_call.a` exercises a
safe borrow, deep global replacement, replacement by a later argument, and an
overriding callee replacing the global. Both run against source Node, backend Node,
release native, ASan, UBSan, and LeakSanitizer.

A push alone cannot free the current element. Nor can reassignment of the array
variable while the iterator owns the old array. Thus there is no honest ASan
use-after-free fixture for those operations alone. Push is allowed and tested;
the conservative reassignment refusal is held by the plan assertion, while
removing the iterator hold makes the reassignment fixture fail under ASan.
Binding capture and assignment refusals, and the lending bookkeeping assertion,
are independently checked rather than claimed as sanitizer failures.
The first captured-binding mutant survived because a closure invocation separately
refused borrowing. Removing that invocation isolated the guard; its rerun failed.

Run `source /workspace/adamic-tools/env.sh` and
`python3 internal/native/testdata/run-loop-borrow-mutants.py > /tmp/loop-mutants.log 2>&1`.
It restores every mutation, writes individual logs under `/tmp/adamic-loop-mutants`,
and fails if a mutant survives or only breaks a build. Twenty loop/global mutants all
failed: thirteen ASan reports and seven compiler behavior assertions. The merged runner
also tests the virtual guard on the earlier indexed-element fixture,
`borrow_element_virtual_store.a`, as `element-virtual-call`.

| Mutant | Catcher |
|---|---|
| Ignore `SetIndex` | ASan, `storeBody` |
| Allow `ArrayPop` | ASan, `popBody` |
| Allow `ArraySplice` | ASan, `spliceBody` |
| Ignore changing named callees | ASan, `callBody` |
| Allow closure calls | ASan, `closureBody` |
| Allow virtual calls using only the static summary | ASan, `virtualBody` |
| Allow an assigned array variable | `TestLoopBorrowPlan`, `reassignBody` |
| Allow an assigned binding | `TestLoopBorrowPlan`, `assignedBinding` |
| Allow a captured binding | `TestLoopBorrowPlan`, `capturedBinding` |
| Do not mark the binding borrowed | ASan, `bindingReuse` |
| Do not mark the array lending | `TestLoopBorrowPlan`, missing lending fact |
| Disable loop borrowing | `TestLoopBorrowPlan` and `TestNbodyBorrowedLoopC` |
| Schedule a borrowed binding's release on a throw | ASan, `throwBody` |
| Omit the iterator retain in `reassignBody` | ASan, `reassignBody` |
| Leave the transferred iterator count in the statement cleanup | ASan, `freshBody` |
| Ignore global `touches` | ASan, `replace` |
| Ignore later-argument purity | ASan, `read` |
| Check only the static global callee | ASan, `Replacer.read` |
| Refuse safe error construction | `TestLoopBorrowPlan`, `throwBody` |
| Disable global lending | `TestGlobalArgumentLending` |

**Every moved counts row.** Across 136 existing rows, allocations and frees fell
by one each, retains by 17,238, releases by 17,234, and the sum of peaks by one;
region counts did not move. Two new rows were added. The complete diff follows.
A/F/R/L/P/G mean allocations, frees, retains, releases, peak, and in regions.
Each cause was measured by disabling that feature alone and regenerating the
entire counts table. Parenthesized R/L numbers are final minus that counterfactual,
so interacting causes need not add up. Bindings remove loop-element counts;
global arguments remove safe call snapshot pairs; iterator transfer removes
redundant temporary pairs. Shared safety is safe error construction plus the
virtual guard. In `exceptions.a`, error construction permits borrowing; in
`borrow_element_virtual_move.a`, the guard prevents the first borrow, permits
safe array moves into consumed parameters, and lets unique mapping reuse an array.
That row's +1 retain/+4 releases and one fewer allocation/free/peak are the same
shared-guard effects measured in the preceding unit. `return_panic_fires.a` stops
before its statement releases temporaries, so its retain and release deltas differ.

| Fixture | Before A/F/R/L/P/G | After A/F/R/L/P/G | Measured cause |
|---|---|---|---|
| `library_object_keys.a` | 88/88/35/75/24/0 | 88/88/32/72/24/0 | iterator transfer (-3 R, -3 L) |
| `library_object_is.a` | 163/163/29/198/7/0 | 163/163/21/190/7/0 | iterator transfer (-8 R, -8 L) |
| `library_object_order.a` | 84/84/31/71/44/0 | 84/84/30/70/44/0 | iterator transfer (-1 R, -1 L) |
| `class_oct6_deep.a` | 43/43/23/55/20/0 | 43/43/22/54/20/0 | iterator transfer (-1 R, -1 L) |
| `class_oct6_parameters.a` | 43/43/16/44/10/0 | 43/43/15/43/10/0 | iterator transfer (-1 R, -1 L) |
| `class_oct6_release.a` | 372/372/180/441/26/0 | 372/372/177/438/26/0 | iterator transfer (-3 R, -3 L) |
| `class_oct6_subclass_holder.a` | 76/76/55/116/12/0 | 76/76/54/115/12/0 | iterator transfer (-1 R, -1 L) |
| `library_array_with.a` | 68/68/57/123/10/0 | 68/68/43/109/10/0 | global arguments (-13 R, -13 L), iterator transfer (-1 R, -1 L) |
| `library_array_flat_map.a` | 50/50/66/115/17/0 | 50/50/62/111/17/0 | bindings (-4 R, -4 L) |
| `library_array_flat.a` | 33/33/84/100/19/0 | 33/33/55/71/19/0 | bindings (-24 R, -24 L), global arguments (-5 R, -5 L) |
| `library_array_spliced.a` | 72/72/59/133/12/0 | 72/72/40/114/12/0 | global arguments (-18 R, -18 L), iterator transfer (-1 R, -1 L) |
| `library_array_copy_within.a` | 615/615/141/754/7/0 | 615/615/126/739/7/0 | global arguments (-4 R, -4 L), iterator transfer (-11 R, -11 L) |
| `library_array_search.a` | 87/87/11/98/11/0 | 87/87/10/97/11/0 | iterator transfer (-1 R, -1 L) |
| `json_stringify_scalars.a` | 100/100/38/170/5/0 | 100/100/37/169/5/0 | iterator transfer (-1 R, -1 L) |
| `json_stringify_options.a` | 74/74/30/83/6/0 | 74/74/29/82/6/0 | iterator transfer (-1 R, -1 L) |
| `json_stringify_numbers.a` | 66/66/2/108/3/0 | 66/66/1/107/3/0 | iterator transfer (-1 R, -1 L) |
| `library_fnexpr_loops.a` | 182/182/197/269/82/0 | 182/182/194/266/82/0 | iterator transfer (-3 R, -3 L) |
| `library_for_in.a` | 79/79/135/123/65/0 | 79/79/130/118/65/0 | iterator transfer (-5 R, -5 L) |
| `library_for_in_keys.a` | 38/38/110/79/34/0 | 38/38/108/77/34/0 | iterator transfer (-2 R, -2 L) |
| `library_for_in_live.a` | 39/39/63/80/21/0 | 39/39/59/76/21/0 | iterator transfer (-4 R, -4 L) |
| `internal/load/testdata/0.1/compile/03_shapes.ts` | 13/13/17/25/7/0 | 13/13/16/24/7/0 | iterator transfer (-1 R, -1 L) |
| `internal/load/testdata/0.1/compile/05_wordcount.ts` | 49/49/73/78/26/0 | 49/49/71/76/26/0 | iterator transfer (-2 R, -2 L) |
| `internal/load/testdata/0.1/compile/06_stack.ts` | 39/39/61/87/7/0 | 39/39/60/86/7/0 | iterator transfer (-1 R, -1 L) |
| `internal/load/testdata/0.1/compile/07_modules/main.ts` | 14/14/14/26/10/0 | 14/14/9/21/10/0 | bindings (-4 R, -4 L), global arguments (-1 R, -1 L) |
| `internal/load/testdata/0.1/compile/08_results.ts` | 26/25/37/49/4/0 | 26/25/36/48/4/0 | iterator transfer (-1 R, -1 L) |
| `internal/load/testdata/0.1/compile/09_tree.ts` | 19/19/85/126/16/0 | 19/19/82/123/16/0 | global arguments (-2 R, -2 L), iterator transfer (-1 R, -1 L) |
| `bitwise_sweep.a` | 4152674/4152674/2076/4154757/9/0 | 4152674/4152674/1260/4153941/9/0 | iterator transfer (-816 R, -816 L) |
| `objects.a` | 58/58/59/110/18/0 | 58/58/47/98/18/0 | bindings (-5 R, -5 L), global arguments (-5 R, -5 L), iterator transfer (-2 R, -2 L) |
| `maps_and_text.a` | 59/59/41/94/10/0 | 59/59/39/92/10/0 | iterator transfer (-2 R, -2 L) |
| `sorting.a` | 61/61/57/75/38/0 | 61/61/55/73/38/0 | iterator transfer (-2 R, -2 L) |
| `classes.a` | 16/16/24/42/10/0 | 16/16/16/34/10/0 | global arguments (-8 R, -8 L) |
| `indexing.a` | 33/33/13/46/6/0 | 33/33/12/45/6/0 | iterator transfer (-1 R, -1 L) |
| `casts.a` | 9/9/17/25/6/0 | 9/9/16/24/6/0 | iterator transfer (-1 R, -1 L) |
| `cast_fails.a` | 2/0/5/3/2/0 | 2/0/4/2/2/0 | iterator transfer (-1 R, -1 L) |
| `updates.a` | 27/27/27/57/9/0 | 27/27/23/53/9/0 | global arguments (-4 R, -4 L) |
| `string_index.a` | 89/89/29/126/7/0 | 89/89/28/125/7/0 | iterator transfer (-1 R, -1 L) |
| `visits.a` | 95/95/125/209/23/0 | 95/95/124/208/23/0 | iterator transfer (-1 R, -1 L) |
| `searches.a` | 106/106/82/174/19/0 | 106/106/81/173/19/0 | iterator transfer (-1 R, -1 L) |
| `defaults.a` | 38/36/17/52/6/2 | 38/36/15/50/6/2 | global arguments (-2 R, -2 L) |
| `strings_more.a` | 128/128/30/164/11/0 | 128/128/29/163/11/0 | iterator transfer (-1 R, -1 L) |
| `library_math_number_math.a` | 1280/1280/50/1331/8/0 | 1280/1280/25/1306/8/0 | iterator transfer (-25 R, -25 L) |
| `library_math_number_convert.a` | 199/199/197/365/7/0 | 199/199/196/364/7/0 | iterator transfer (-1 R, -1 L) |
| `library_math_number_prototype.a` | 71/71/2/74/8/0 | 71/71/1/73/8/0 | iterator transfer (-1 R, -1 L) |
| `navigation.a` | 162/162/40/189/16/0 | 162/162/36/185/16/0 | iterator transfer (-4 R, -4 L) |
| `number_formats.a` | 235/235/46/263/18/0 | 235/235/45/262/18/0 | iterator transfer (-1 R, -1 L) |
| `precision_range.a` | 4/3/2/5/2/0 | 4/3/1/4/2/0 | iterator transfer (-1 R, -1 L) |
| `radixes.a` | 696/696/97/332/25/0 | 696/696/95/330/25/0 | iterator transfer (-2 R, -2 L) |
| `timsort.a` | 2769/2769/3779/6232/308/0 | 2769/2769/3778/6231/308/0 | iterator transfer (-1 R, -1 L) |
| `sort_top_level.a` | 33/33/80/97/19/0 | 33/33/79/96/19/0 | iterator transfer (-1 R, -1 L) |
| `splices.a` | 74/74/41/108/12/0 | 74/74/29/96/12/0 | global arguments (-12 R, -12 L) |
| `array_from.a` | 94/94/57/132/25/0 | 94/94/56/131/25/0 | iterator transfer (-1 R, -1 L) |
| `weak_parent.a` | 86/86/264/313/41/0 | 86/86/253/302/41/0 | global arguments (-8 R, -8 L), iterator transfer (-3 R, -3 L) |
| `doubly_linked.a` | 90/90/545/578/49/0 | 90/90/525/558/49/0 | global arguments (-20 R, -20 L) |
| `fresh_parser.a` | 294/294/338/508/34/0 | 294/294/292/462/34/0 | global arguments (-4 R, -4 L), iterator transfer (-42 R, -42 L) |
| `fresh_writes.a` | 330/330/409/525/132/0 | 330/330/369/485/132/0 | global arguments (-4 R, -4 L), iterator transfer (-36 R, -36 L) |
| `fresh_calls.a` | 147/147/268/323/39/0 | 147/147/239/294/39/0 | global arguments (-2 R, -2 L), iterator transfer (-27 R, -27 L) |
| `exceptions.a` | 133/133/171/246/20/0 | 133/133/159/234/20/0 | bindings (-5 R, -5 L), global arguments (-3 R, -3 L), iterator transfer (-4 R, -4 L), shared safety (-5 R, -5 L) |
| `param_assigned_in_try.a` | 46/46/32/67/12/0 | 46/46/31/66/12/0 | global arguments (-1 R, -1 L) |
| `invariance_readonly.a` | 30/30/39/52/16/0 | 30/30/38/51/16/0 | iterator transfer (-1 R, -1 L) |
| `tuples_kept.a` | 31/31/54/74/13/0 | 31/31/52/72/13/0 | iterator transfer (-2 R, -2 L) |
| `undefined_strings.a` | 15/15/21/43/6/0 | 15/15/19/41/6/0 | global arguments (-2 R, -2 L) |
| `maybe_booleans.a` | 48/48/51/101/11/0 | 48/48/47/97/11/0 | global arguments (-3 R, -3 L), iterator transfer (-1 R, -1 L) |
| `unions.a` | 1449/1449/1111/2573/16/0 | 1449/1449/1107/2569/16/0 | bindings (-3 R, -3 L), iterator transfer (-1 R, -1 L) |
| `maybe_number_slots.a` | 484/484/305/798/16/0 | 484/484/304/797/16/0 | iterator transfer (-1 R, -1 L) |
| `case_mapping.a` | 203/203/20/217/16/0 | 203/203/19/216/16/0 | iterator transfer (-1 R, -1 L) |
| `undefined_elements.a` | 20/20/46/43/9/0 | 20/20/45/42/9/0 | iterator transfer (-1 R, -1 L) |
| `map_zero_keys.a` | 29/29/24/49/8/0 | 29/29/23/48/8/0 | iterator transfer (-1 R, -1 L) |
| `adversarial_exits.a` | 120/120/80/175/25/0 | 120/120/63/158/25/0 | bindings (-13 R, -13 L), global arguments (-3 R, -3 L), iterator transfer (-1 R, -1 L) |
| `adversarial_iteration.a` | 59/59/120/152/21/0 | 59/59/118/150/21/0 | iterator transfer (-2 R, -2 L) |
| `string_positions.a` | 436/436/301/607/134/0 | 436/436/300/606/134/0 | iterator transfer (-1 R, -1 L) |
| `class_layouts.a` | 70/70/112/159/27/0 | 70/70/103/150/27/0 | global arguments (-8 R, -8 L), iterator transfer (-1 R, -1 L) |
| `size_class_churn.a` | 675173/675173/270039/675213/8005/0 | 675173/675173/270038/675212/8005/0 | iterator transfer (-1 R, -1 L) |
| `class_as_interface.a` | 372/372/360/525/60/0 | 372/372/356/521/60/0 | global arguments (-1 R, -1 L), iterator transfer (-3 R, -3 L) |
| `optional_class_method.a` | 40/40/47/74/9/0 | 40/40/44/71/9/0 | iterator transfer (-3 R, -3 L) |
| `spread_snapshot.a` | 22/22/4/25/11/0 | 22/22/3/24/11/0 | global arguments (-1 R, -1 L) |
| `reuse.a` | 76/76/57/122/16/0 | 76/76/55/120/16/0 | global arguments (-2 R, -2 L) |
| `regions.a` | 316/260/219/409/50/56 | 316/260/213/403/50/56 | global arguments (-6 R, -6 L) |
| `borrow_loop.a` | new | 88/86/94/142/7/2 | new fixture |
| `borrow_global_call.a` | new | 18/16/10/23/6/2 | new fixture |
| `borrow_element_virtual_move.a` | 26/24/10/22/12/2 | 25/23/11/26/11/2 | shared safety (1 R, 4 L) |
| `spread_undefined.a` | 45/45/27/65/10/0 | 45/45/26/64/10/0 | iterator transfer (-1 R, -1 L) |
| `trig_reduction.a` | 133/133/14/148/7/0 | 133/133/1/135/7/0 | iterator transfer (-13 R, -13 L) |
| `sets.a` | 224/224/13016/13113/65/0 | 224/224/13011/13108/65/0 | bindings (-4 R, -4 L), iterator transfer (-1 R, -1 L) |
| `library_map_set.a` | 170/170/219/320/24/0 | 170/170/135/236/24/0 | global arguments (-84 R, -84 L) |
| `library_map_set_keys.a` | 76/76/70/135/15/0 | 76/76/64/129/15/0 | global arguments (-6 R, -6 L) |
| `library_map_set_construct.a` | 180/180/229/275/52/0 | 180/180/228/274/52/0 | iterator transfer (-1 R, -1 L) |
| `maybe_collections.a` | 64/64/166/225/20/0 | 64/64/165/224/20/0 | iterator transfer (-1 R, -1 L) |
| `collections.a` | 363/363/395/565/101/0 | 363/363/391/561/101/0 | iterator transfer (-4 R, -4 L) |
| `power_of_two_string.a` | 19/19/2/22/4/0 | 19/19/1/21/4/0 | iterator transfer (-1 R, -1 L) |
| `declared_later.a` | 56/56/76/128/15/0 | 56/56/75/127/15/0 | iterator transfer (-1 R, -1 L) |
| `return_panic.a` | 13/13/14/27/9/0 | 13/13/13/26/9/0 | bindings (-1 R, -1 L) |
| `return_panic_fires.a` | 5/0/8/7/5/0 | 5/0/6/6/5/0 | global arguments (-2 R, -1 L) |
| `from_codes.a` | 408/408/56/416/12/0 | 408/408/53/413/12/0 | global arguments (-1 R, -1 L), iterator transfer (-2 R, -2 L) |
| `bitwise.a` | 9304/9304/1406/9379/45/0 | 9304/9304/1369/9342/45/0 | iterator transfer (-37 R, -37 L) |
| `tuple_values.a` | 119/119/126/207/37/0 | 119/119/122/203/37/0 | global arguments (-2 R, -2 L), iterator transfer (-2 R, -2 L) |
| `generic_values.a` | 22/22/39/52/10/0 | 22/22/33/46/10/0 | global arguments (-6 R, -6 L) |
| `undefined_references.a` | 17/17/37/52/11/0 | 17/17/34/49/11/0 | bindings (-3 R, -3 L) |
| `utf8_view.a` | 59/59/34/80/7/0 | 59/59/33/79/7/0 | iterator transfer (-1 R, -1 L) |
| `search_from.a` | 292/292/168/329/22/0 | 292/292/159/320/22/0 | iterator transfer (-9 R, -9 L) |
| `shared_slices.a` | 12675/12675/5999/12774/87/0 | 12675/12675/5992/12767/87/0 | global arguments (-5 R, -5 L), iterator transfer (-2 R, -2 L) |
| `string_append.a` | 249/249/95/285/49/0 | 249/249/89/279/49/0 | bindings (-6 R, -6 L) |
| `search_from_sweep.a` | 133/133/34/155/10/0 | 133/133/33/154/10/0 | iterator transfer (-1 R, -1 L) |
| `integer_format.a` | 106826/106826/58737/146871/20/0 | 106826/106826/48056/136190/20/0 | iterator transfer (-10681 R, -10681 L) |
| `reuse_spread_method_alias.a` | 8/8/8/15/6/0 | 8/8/7/14/6/0 | iterator transfer (-1 R, -1 L) |
| `try_assignments.a` | 37/37/9/42/6/0 | 37/37/8/41/6/0 | iterator transfer (-1 R, -1 L) |
| `library_string_conversion.a` | 15/15/4/21/3/0 | 15/15/3/20/3/0 | iterator transfer (-1 R, -1 L) |
| `library_string_indices.a` | 639/639/200/845/8/0 | 639/639/56/701/8/0 | global arguments (-129 R, -129 L), iterator transfer (-15 R, -15 L) |
| `library_string_raw.a` | 45/45/61/95/11/0 | 45/45/58/92/11/0 | global arguments (-3 R, -3 L) |
| `library_map_set_setops.a` | 590/590/347/666/17/0 | 590/590/336/655/17/0 | global arguments (-11 R, -11 L) |
| `regexp.a` | 455/455/369/422/62/0 | 455/455/364/417/62/0 | iterator transfer (-5 R, -5 L) |
| `sweeps/regexp_methods.a` | 448046/448046/131098/340026/68/0 | 448046/448046/126442/335370/68/0 | iterator transfer (-4656 R, -4656 L) |
| `regexp_exec.a` | 179/179/82/160/18/0 | 179/179/77/155/18/0 | global arguments (-5 R, -5 L) |
| `regexp_match.a` | 83/83/85/103/18/0 | 83/83/84/102/18/0 | iterator transfer (-1 R, -1 L) |
| `class_features_static.a` | 110/110/84/191/28/0 | 110/110/59/166/28/0 | global arguments (-25 R, -25 L) |
| `class_features_static_private.a` | 42/42/71/112/12/0 | 42/42/53/94/12/0 | global arguments (-18 R, -18 L) |
| `class_features_private.a` | 47/47/37/65/15/0 | 47/47/29/57/15/0 | global arguments (-8 R, -8 L) |
| `class_features_accessors.a` | 67/67/53/110/21/0 | 67/67/47/104/21/0 | global arguments (-6 R, -6 L) |
| `class_features_twice.a` | 26/26/6/35/6/0 | 26/26/4/33/6/0 | global arguments (-2 R, -2 L) |
| `class_features_retained.a` | 48/48/57/88/18/0 | 48/48/42/73/18/0 | global arguments (-15 R, -15 L) |
| `class_features_distinct.a` | 48/48/50/81/14/0 | 48/48/48/79/14/0 | global arguments (-2 R, -2 L) |
| `class_inheritance.a` | 150/150/29/170/23/0 | 150/150/24/165/23/0 | global arguments (-4 R, -4 L), iterator transfer (-1 R, -1 L) |
| `class_inheritance_exceptions.a` | 29/29/22/42/8/0 | 29/29/19/39/8/0 | iterator transfer (-3 R, -3 L) |
| `class_inheritance_memory.a` | 48/42/22/60/11/6 | 48/42/20/58/11/6 | global arguments (-2 R, -2 L) |
| `class_inheritance_generic.a` | 89/89/104/167/39/0 | 89/89/85/148/39/0 | global arguments (-18 R, -18 L), iterator transfer (-1 R, -1 L) |
| `class_inheritance_interface.a` | 37/37/45/71/13/0 | 37/37/44/70/13/0 | iterator transfer (-1 R, -1 L) |
| `class_inheritance_conditional.a` | 164/164/148/251/35/0 | 164/164/146/249/35/0 | global arguments (-2 R, -2 L) |
| `user_iterators.a` | 669/669/466/915/67/0 | 669/669/465/914/67/0 | iterator transfer (-1 R, -1 L) |
| `literal_optional_shapes.a` | 36/36/40/68/9/0 | 36/36/39/67/9/0 | iterator transfer (-1 R, -1 L) |
| `e4eec87_u03_discriminated_undefined.a` | 10/10/14/20/6/0 | 10/10/12/18/6/0 | global arguments (-1 R, -1 L), iterator transfer (-1 R, -1 L) |
| `e4eec87_f1_field_narrowed.a` | 1/0/2/3/1/0 | 1/0/1/2/1/0 | global arguments (-1 R, -1 L) |
| `e4eec87_f1_field_present.a` | 6/6/2/9/3/0 | 6/6/1/8/3/0 | global arguments (-1 R, -1 L) |
| `object_prototype.a` | 142/142/317/440/22/0 | 142/142/165/288/22/0 | global arguments (-150 R, -150 L), iterator transfer (-2 R, -2 L) |
| `regexp_cycle_fields.a` | 61/61/134/132/44/0 | 61/61/132/130/44/0 | global arguments (-2 R, -2 L) |
| `regexp_cycle_collections.a` | 31/31/86/90/23/0 | 31/31/85/89/23/0 | iterator transfer (-1 R, -1 L) |
| `utf8_sweep.a` | 38498/38498/7702/38502/7706/0 | 38498/38498/7701/38501/7706/0 | iterator transfer (-1 R, -1 L) |
| `arguments.a` | 77/77/15/69/27/0 | 77/77/14/68/27/0 | iterator transfer (-1 R, -1 L) |
| `read_arguments.a` | 15/15/12/19/9/0 | 15/15/11/18/9/0 | iterator transfer (-1 R, -1 L) |
| `walk.a` | 222/222/123/294/36/0 | 222/222/117/288/36/0 | iterator transfer (-6 R, -6 L) |

**Validation run.** Test output went to files, never through a pipe. The final
unmutated commands and results were:

```text
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/native ./internal/oracle
ok github.com/system-inc/adamic/internal/native 111.817s
ok github.com/system-inc/adamic/internal/oracle 108.826s
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/borrow_|TestLoopBorrowPlan|TestNbodyBorrowedLoopC|TestGlobalArgumentLending|TestPassThroughsAreNotConsumers|TestCountsAreRecorded' -count=1 -timeout 30m
ok github.com/system-inc/adamic/internal/native 0.188s
ok github.com/system-inc/adamic/internal/oracle 13.163s
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
ok github.com/system-inc/adamic/internal/oracle 12.211s
go vet ./...
exit 0, no output
gofmt -l cmd internal
exit 0, no output
git diff --check
exit 0, no output
```

The whole repository gate was not run; the complete two touched packages were.
Iterator hold elimination, borrowing from captured/global sources, closure-body
binding optimization, destructuring, and write-only global effect summaries were
not covered. The existing stable-array and borrowed-parameter requirements remain
conservative. The owned-iterable mutant was further isolated to retain local
sources correctly but leave an owned temporary scheduled for statement cleanup;
`freshBody` failed under ASan. This proves the transfer itself, independently of
the reassignment hold test. Logs are `/tmp/loop-packages.log`,
`/tmp/loop-worker.log`, `/tmp/loop-counts.log`, `/tmp/loop-vet.log`,
`/tmp/loop-mutants-final.log`, `/tmp/loop-mutant-transfer-final.log`, and
`/tmp/loop-bench.log`. The per-feature measurements are
`/tmp/loop-counts-no-{bindings,global-arguments,iterator-transfer,shared-safety}.md`.

### Merge verification, October 7, 2026

Merged `658dbc4` (`codex/borrowed-element-reads`) into `75d3be5`
(`codex/borrowed-loop-bindings`) with two parents, without rebasing. Both fixture
sets and both borrowing reports remain. The shared `MakeError` allowance and
virtual-call guard are preserved with the loop binding and lending changes.
`TestNbodyIndexedElementsBorrow` now counts only `ir.Declare` entries in the shared
plan: its five indexed declarations remain checked independently of the two new
loop bindings. The throwing declaration test also ignores non-declaration entries.

After all mutants were restored, the focused native checks and sanitizer oracle
passed for `borrow_element_virtual_store.a`, `borrow_element_throw.a`,
`borrow_loop.a`, and `borrow_global_call.a` (native 0.228s, oracle 0.800s).
The original twenty mutants plus `element-virtual-call` were all caught:
fourteen ASan reports and seven compiler assertions. The additional virtual guard
mutant fails with ASan heap-use-after-free in `inspect` on
`borrow_element_virtual_store.a`, proving the earlier main bug remains covered.
The first focused run overlapped a deliberate mutant and was discarded; the
reported run was repeated after restoration, before the complete package gate.

```text
source /workspace/adamic-tools/env.sh
python3 internal/native/testdata/run-loop-borrow-mutants.py > /tmp/loop-merge-mutants.log 2>&1
all 21 caught; exit 0
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
ok github.com/system-inc/adamic/internal/oracle 12.325s
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/native ./internal/oracle
ok github.com/system-inc/adamic/internal/native 123.226s
ok github.com/system-inc/adamic/internal/oracle 123.011s
go vet ./...
exit 0, no output
gofmt -l cmd internal
exit 0, no output
git diff --check
exit 0, no output
```

The regenerated table adds only the two restored fixture rows; existing rows,
including the loop/global fixtures, did not change from `75d3be5`. The earlier
fixture counts remain 38/38/34/54/8/0 for throw and 13/11/9/15/7/2 for virtual
store, in allocations/frees/retains/releases/peak/regions order. Complete native
and oracle packages were run, not the whole repository gate. Final logs are
`/tmp/loop-merge-focus-final.log`, `/tmp/loop-merge-mutants.log`,
`/tmp/loop-merge-packages.log`, `/tmp/loop-merge-counts.log`, and
`/tmp/loop-merge-vet.log`.

### The loop's array borrows the same owner, October 7, 2026

Starting from `6c6417a` (`codex/borrowed-loop-bindings`), `forOf` now uses the
existing `elementBorrows` fact for both the binding and its iterator array.
`loopBorrowable` requires a plain array loop with a lendable reference binding,
and `borrowable` requires that the binding and the source array are locals of
this function, uncaptured and never assigned. A parameter qualifies as a local.
The body's `changes` check proves that no expression, statement, or reachable
nonvirtual named callee can take an element out of any array. The source is
marked lending before reuse runs, so a call cannot move its variable's count
into a consumed parameter. The source's owner therefore outlives every iterator
step. The emitter snapshots the pointer but takes no iterator count and schedules
no iterator release, on the normal path or on any early or exceptional exit.
No new purity or call-effect allowance was added.

This deliberately reuses the whole element proof. Assignment anywhere in the
function refuses it, even outside the body; assigned or captured bindings,
non-reference elements, destructuring, closures, maps, strings, regex iterators,
globals, captured sources, and fresh-call sources keep their existing holds.
An operation not on the existing allowed list still refuses the borrow.

**A throw is an exit, not an invalidation.** An unchanged local's scope or a
parameter's caller still holds the array until the iterator has ended, including
when a throw ends it. Thus `throwBody` can borrow both binding and array. A throw
alone cannot honestly demonstrate that a redundant iterator hold is needed.
`throwHeldBody` instead reassigns its source and immediately throws; its iterator
must own the old array and release that count during unwinding. Its mutant omits
the retain but preserves the unwind release, which ASan catches on the throw
path before another length test can run. This specifically proves exceptional
cleanup ownership, separately from the four full hold-removal mutants.

**Fixtures and mutants.** The existing `borrow_loop.a` has independent
`reassignBody` and `freshBody` probes. It now also has `closureReassignBody`
(a called closure writes the captured array variable), `globalReassignBody`
(a named call writes the global source), `throwHeldBody`, and `parameterBody`
(a caller-held fresh argument borrowed safely for the entire callee loop).
Labels and arrays are made at runtime. `TestLoopArrayHoldC` checks the actual
iterator temporary in eleven fixture functions and both nbody loop functions,
including positive local, parameter, push, and throw cases. Existing plan tests
continue to require the borrowed binding and lending facts.

Run `python3 internal/native/testdata/run-loop-array-hold-mutants.py` with the
setup environment sourced. Every mutation is restored even on failure; complete
outputs are in `/tmp/adamic-loop-array-hold-mutants/<mutant>.log`.
All seven mutants were caught, independently:

| Mutant | Catcher |
|---|---|
| Remove the entire iterator hold in `reassignBody` | ASan heap-use-after-free, next loop length read |
| Remove the entire iterator hold in `closureReassignBody` | ASan heap-use-after-free, next loop length read |
| Remove the entire iterator hold in `globalReassignBody` | ASan heap-use-after-free, next loop length read |
| Remove the entire iterator hold in `freshBody` | ASan heap-use-after-free, first loop length read after statement cleanup |
| Omit the retain, keep unwind cleanup in `throwHeldBody` | ASan heap-use-after-free, unwind release |
| Schedule iterator cleanup for borrowed `throwBody` | ASan heap-use-after-free, array read after its catch |
| Disable hold elision, preserving binding borrows | `TestLoopArrayHoldC`, redundant owned iterators |

**Nbody, observed.** The only C changes are the array retain and release in
`offsetMomentum` and in `advance`. The latter runs one million times and the
former once, removing exactly 1,000,001 pairs. The ten remaining retains are
five constructor returns and five array insertions. The fourteen releases also
include the two formatted energy strings, final global cleanup, and the initial
NULL global replacement.

| nbody | Before | After |
|---|---:|---:|
| Allocations / frees | 8 / 8 | 8 / 8 |
| Retains | 1,000,011 | 10 |
| Releases | 1,000,015 | 14 |
| Peak live / in regions | 7 / 0 | 7 / 0 |

clang 20.1.8, Go 1.27.1, Node 24.19.0, Linux 6.18.44, AMD EPYC 9V74.
`nproc` reports 5, cgroup quota 4 CPUs. `bash cloud/setup.sh` completed:
Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, cache warm
105s, done 105s. The generated environment file is
`/workspace/adamic-tools/env.sh`, sourced in every build and test shell.

`go run ./bench -only nbody -rounds 5` reported best native times 0.442s
before and 0.412s after, and identical answers against Node (0.719s and
0.651s). The baseline overlapped setup's cache warming, with one-minute load
9.34 to 11.38; the after run was at 3.30 to 2.79. These are not comparable
speedup measurements. Both native binaries were then interleaved for seven
rounds, alternating which ran first, with no other test workload running.
Best times were **0.407087s before and 0.410165s after**, at one-minute load
2.79 to 2.65. All fourteen runs printed `-0.169075164` then `-0.169086185`.
There is no observed timing improvement despite the deterministic count reduction.

**The entire counts diff.** All 308 rows were regenerated twice: with the new
fixture and hold elision disabled, then with the final compiler. This isolates
fixture additions from the optimization. Only the twelve rows below differ from
`6c6417a`; the other 296 are unchanged. A/F/R/L/P/G mean allocations, frees,
retains, releases, peak live, and values in regions. The C diff was inspected
for each row to identify the loops listed, including the lowered library helpers.
Every optimizer-only change is an equal reduction of retains and releases,
38 of each in total; no allocation, free, peak, or region changes.

| Fixture | Before A/F/R/L/P/G | After A/F/R/L/P/G | Cause of removed hold pairs |
|---|---|---|---|
| `library_array_flat_map.a` | 50/50/62/111/17/0 | 50/50/60/109/17/0 | 2 loop entries in the generated string-array `array_flat_map` helper |
| `library_array_flat.a` | 33/33/55/71/19/0 | 33/33/41/57/19/0 | 14 loop entries in generated `array_flat` helpers, including nested arrays |
| `internal/load/testdata/0.1/compile/07_modules/main.ts` | 14/14/9/21/10/0 | 14/14/8/20/10/0 | 1 call to `pathLength` in `07_modules/geometry.ts` |
| `objects.a` | 58/58/47/98/18/0 | 58/58/45/96/18/0 | 2 calls to `firstNamed` |
| `exceptions.a` | 133/133/159/234/20/0 | 133/133/156/231/20/0 | 3 calls to `find`, including its throw paths |
| `unions.a` | 1449/1449/1107/2569/16/0 | 1449/1449/1106/2568/16/0 | 1 entry to `lookups` |
| `adversarial_exits.a` | 120/120/63/158/25/0 | 120/120/60/155/25/0 | 3 calls to `find`, including continue and return paths |
| `borrow_loop.a` | 88/86/94/142/7/2 | 128/126/125/196/14/2 | 5 entries: `pushBody`, `throwBody`, `caughtBody`, `bindingReuse`, `parameterBody`; new fixture work measured separately |
| `sets.a` | 224/224/13011/13108/65/0 | 224/224/13010/13107/65/0 | 1 call to `count` over its array parameter; Set iterator holds are unchanged |
| `return_panic.a` | 13/13/13/26/9/0 | 13/13/12/25/9/0 | 1 call to `firstWord`; the separate panic-firing fixture is unchanged |
| `undefined_references.a` | 17/17/34/49/11/0 | 17/17/30/45/11/0 | 4 calls to `find`, including early returns |
| `string_append.a` | 249/249/89/279/49/0 | 249/249/88/278/49/0 | 1 entry to `seams` over local `pieces` |

With elision disabled but the expanded fixture present, `borrow_loop.a` is
128/126/130/201/14/2. Its four added probes and the global initializer therefore
add **40 allocations, 40 frees, 36 retains, 59 releases, and 7 peak live** relative
to the old fixture. The new global array stays live while the closure-reassignment
probe runs, contributing to the higher peak. Elision removes five pairs from that
expanded fixture, giving 128/126/125/196/14/2. This is additional fixture work,
not a lifetime change in the optimization. Summed across the whole checked-in
table, the net diff is +40 allocations, +40 frees, -2 retains, +21 releases,
+7 in the sum of peaks, and no change in regions. There are no added rows.

**The previous borrowing guards were rerun.**
`python3 internal/native/testdata/run-loop-borrow-mutants.py` still catches every
one of its 21 independent mutants on the final compiler. Each exits 1, with
14 sanitizer failures and 7 behavior assertions; no build-warning kill counts.
Together with the new runner, all 28 mutants were caught (20 ASan, 8 assertions).
The runner's exact names and observed catchers in this run are:

| Mutants | Catcher |
|---|---|
| `set-index`, `pop`, `splice`, `named-call`, `closure-call`, `binding-owned`, `throw-owned`, `iterator-hold`, `global-touches`, `later-argument`, `virtual-call`, `global-virtual-targets`, `iterator-transfer`, `element-virtual-call` | ASan heap-use-after-free in the borrowing oracle fixtures |
| `array-assigned`, `binding-assigned`, `binding-captured`, `lending-fact`, `borrow-disabled`, `error-construction` | `TestLoopBorrowPlan` (also emitted-C assertions for disabled borrowing) |
| `global-disabled` | `TestGlobalArgumentLending` |

**Validation and limits.** The final unmutated commands wrote their full output
to log files and passed:

```text
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/oracle -run 'TestLoopBorrowPlan|TestLoopArrayHoldC|TestNbodyBorrowedLoopC|TestNativeAgreesWithNode/internal/oracle/testdata/borrow_loop' -count=1 -timeout 30m
ok github.com/system-inc/adamic/internal/native 0.158s
ok github.com/system-inc/adamic/internal/oracle 4.772s
python3 internal/native/testdata/run-loop-array-hold-mutants.py
7 caught; exit 0
python3 internal/native/testdata/run-loop-borrow-mutants.py
21 caught; exit 0
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
holds kept counterfactual: ok github.com/system-inc/adamic/internal/oracle 10.862s
final compiler: ok github.com/system-inc/adamic/internal/oracle 10.462s
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/native ./internal/oracle
ok github.com/system-inc/adamic/internal/native 110.598s
ok github.com/system-inc/adamic/internal/oracle 107.531s
go vet ./...
exit 0, no output
gofmt -l cmd internal
exit 0, no output
git diff --check
exit 0, no output
```

This includes the entire touched packages, source Node and backend Node,
release native, ASan, UBSan, LeakSanitizer, and the complete recorded counts
check. The full repository gate was not run under the worker-gate exception;
stage-1 packages were only compiled by setup and vetted. This unit does not
weaken the existing whole-function assignment or element-mutation conditions,
optimize closure bodies, or remove holds from non-reference, destructured,
global, captured, map, string, or regex sources.

Logs: `/tmp/loop-hold-setup.log`, `/tmp/loop-hold-focus.log`,
`/tmp/loop-hold-mutants.log`, `/tmp/loop-hold-regression-mutants.log`,
`/tmp/loop-hold-counts-kept.log`, `/tmp/loop-hold-counts.log`,
`/tmp/loop-hold-packages.log`, `/tmp/loop-hold-vet.log`,
`/tmp/loop-hold-before-bench.log`, `/tmp/loop-hold-after-bench.log`, and
`/tmp/loop-hold-paired-bench.log`. The counterfactual table is
`/tmp/loop-hold-counts-kept.md`; preserved generated C is
`/tmp/loop-hold-before.c` and `/tmp/loop-hold-after.c`.

### Calls that can free a loop's array (coverage, October 7, 2026)

`borrow_loop_calls.a` sits on `2b63ccb`, which is `6c6417a` plus the iterator
borrowing a stable local or parameter. Both proofs are one fact, so one fixture
covers both. The strings are built at runtime. A callee that pops (including
through another parameter holding the same array, a branch, a try, or a deeper
call), spreads, maps, stores an index, or is reached virtually keeps the
iterator's count and retains the binding. A callee that reads a length, forwards
to such a read, or only pushes does not: the binding is borrowed and the
iterator takes no count. `TestLoopCallCoverage` checks that plan and the
iterator temporary for all 26 loops.

Two elided loops are cases a call really can free, and another owner still holds
the array. `scanCopy` copies a global into a local and then calls a function
that replaces the global; the local's count is the owner. `scanLent` takes the
global as a borrowed parameter and calls the function that replaces it.
`reportQuiet` does lend its global, because that callee never touches it.
`scanLent` does not: lending follows the callee and sees the assignment, so the
call retains. The loop then borrows. Replacing the global drops the global's
count and leaves the argument's. Both labels print, and the global's length is
then 0. A field copied into a local survives a call that replaces the field.
Iterating the field expression directly does not borrow; the iterator holds
that read.

| Mutant | Catcher |
|---|---|
| A direct call counts as unchanging | ASan heap-use-after-free in `popLocal`, the element `remove` freed |
| Global lending does not follow callees | ASan heap-use-after-free in `scanLent`, after `resetLent` releases the global |

Run `source /opt/adamic-tools/env.sh` and
`python3 internal/native/testdata/run-loop-call-mutants.py`. Each mutant is
restored. Logs are `/tmp/adamic-loop-call-mutants/<mutant>.log`. Both exited 1
with that ASan report and no other failure mode.

The new counts row is 263/261/189/355/25/2. `TestCountsAreRecorded -update-counts`
regenerated the whole table after merging `codex/loop-array-hold` into `e8ba3d5`.
Call targets from that main changed the rows the two sides had disagreed on:
`library_object_keys.a` 32/72 retains/releases, `library_object_is.a` 21/190,
`class_as_interface.a` 345/510, `optional_class_method.a` 44/71,
`class_inheritance_memory.a` 20/58, `class_inheritance_generic.a` 85/148,
`class_inheritance_interface.a` 44/70, `class_inheritance_conditional.a` 146/249,
`devirtualize.a` 23/63, `user_iterators.a` 465/914, and `prompt_then_read.a`
2/2/2/4/2/0. The direct-call mutant now skips `CallTargets` instead of the old
`Virtual == 0` test, which this main no longer has. It still frees the element
under `popLocal`.

```text
source /opt/adamic-tools/env.sh
go test ./internal/native -run 'TestLoopCallCoverage|TestLoopArrayHoldC|TestLoopBorrowPlan|TestGlobalArgumentLending|TestNbodyBorrowedLoopC' -count=1 -timeout 10m
ok github.com/system-inc/adamic/internal/native 0.266s
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
ok github.com/system-inc/adamic/internal/oracle 22.942s
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/borrow_loop' -count=1 -timeout 30m
ok github.com/system-inc/adamic/internal/oracle 0.943s
python3 internal/native/testdata/run-loop-call-mutants.py
call-unchanging and touches-ignores-callees caught; exit 0
```

The oracle run is source Node, the JavaScript backend, release native, ASan,
UBSan, and LeakSanitizer for this fixture. The whole repository gate was not
run. Closure-body binding optimization, destructuring, and write-only global
effect summaries remain uncovered, as they were before this fixture.

### Strong field borrow chains

A strong field read can borrow transitively from a stable root (`borrow.go`,
`element_borrow.go`). For example, `node.parent`, `node.parent.kind`, and
`node.children` can be read without taking a count; an uncaptured, unassigned
local initialized by such a read can borrow for its scope too. Existing borrowed
parameters carry `Local.Borrowed`; an unassigned ordinary local already holds its
own count. A captured or global root does not qualify.

The proof checks the whole function, including enclosing blocks, loops, catches
and finallies. No assignment can replace the root. No statement or reachable
callee can replace any field on the chain. Named calls use every `CallTargets`
implementation, recursive bodies included; callback calls use `ClosureTargets`,
with `Unknown` refusing the borrow. Field names conservatively match every holder,
not just related types, since the native IR does not carry that relation. Unknown
operations and object spreads refuse. Emission also refuses a root that reuse
consumes, moves or takes over. Borrowing declarations mark their variables
`Borrowed` and their root as lending before reuse is planned.

A borrowed chain is snapshotted at its source read. A store or an owned return
still takes a count through the existing ownership helpers. A captured declaration
still creates an owning capture cell. An argument to an existing borrowed
parameter needs no temporary retain when the chain survives the function's calls;
a consumed argument still takes the count its convention requires.

This proof follows strong property loads and `Defined` only. It does not infer an
owner through a weak target read, a call result, an optional chain or an indexed
load. Weak targets need a separate strong owner; accessor results need the
compiler's borrowed-return provenance. The pass does not change parameter
conventions, for-of emission, accessor returns or devirtualization.

The nine `borrow_chain_*.a` fixtures run against source on Node, native with
sanitizers, and the JavaScript backend. The tree-walk fixture loses nine pairs;
the chained-argument fixture goes from one retain to zero. Refusals keep the count
that protects the saved parent, override argument or capture. Mutants borrowing
across a field write, a reassigned root, a writing override or an Unknown callback,
and a generated capture cell with its retain removed, each fail under ASan.
Disabling declaration borrowing fails the positive planner control; changing a
recorded retain fails the counts gate.

The current-main 321-fixture comparison removes 514 retains and 509 releases, with no
rises or changes to allocation, free, peak or region counts. Five extra removed
retains are absent optional string reads in `literal_optional_shapes.a`: their
NULL values were passed on by `??` without a corresponding temporary release.
Batch8 removes 15,334,163 pairs on the 77-file compiler corpus, while its best
release time moves only 2.4 percent. Full evidence, commands and remaining
accessor/iterator work are in [the borrow-chains report](../internal/native/performance/borrow-chains/REPORT.md).

## Strings, specifically

UTF-8 bytes, immutable, counted. JavaScript programs see UTF-16 (`length`, indexes, `<`), so the runtime keeps UTF-16 behavior over UTF-8 storage: an ASCII-only flag makes the common case free, and other strings compute the mapping when first asked. Lone surrogates (which UTF-8 can't hold) are stored as WTF-8 and written out as U+FFFD, as Node does. Program 10 in docs/0.1.md is the fixture for all of it.

Long non-ASCII strings keep a byte checkpoint every 32 UTF-16 units and a cursor at the last code point found (`runtime/string_index.c`). Sequential reads walk from that cursor; nearby backward reads choose it when it is closer than the checkpoint and walk backward over UTF-8 continuation bytes. The cursor always names a code point's first unit, so either half of a supplementary point is found from the same bytes. `codePointAt` decodes the location once, and slices reuse their already located boundary halves.

An indexed string with no supplementary points also keeps a compact UTF-16 view, two bytes per unit, for direct `charCodeAt` and `codePointAt` reads. Lone surrogates are BMP units and can use this view. ASCII strings need neither cache; short strings and stack pieces still walk without an index. The view is made when the long string's index is first built, freed with that index, and invalidated with it before an in-place append. This trades extra cache bytes and one decoding pass for cheaper repeated reads. Measurements and the Node sweeps are recorded in [UTF-16 views](utf16-views.md).

Since a string is immutable, a slice can read its parent's bytes in place, as a Go string does, and hold a reference to the parent (its owner) so they outlive it (runtime/string_share.c, fixture `shared_slices.a`). A slice of a slice holds the first one's owner, so no chain grows. Its caches (units and index) are its own, counted from its own first byte. A small slice is copied only when its ultimate owner's header and byte capacity exceed eight times the slice's header and byte length. Spare append capacity counts toward that bound. There is no minimum slice length: short views can share small parents. Constants pin no counted storage, so their slices share freely. A stack piece without a count or literal marker is copied, including a whole slice, because no view can keep its bytes alive. Empty slices use the immortal empty string. ASCII indexing uses 128 immortal one-unit strings, so it allocates and copies nothing; longer slices retain their byte owner. UTF-16 slices carry their already known unit length, and a split surrogate boundary still builds the necessary WTF-8 half.

`text = text + more` on a local nothing else can write while `more` is evaluated (not a global, not captured) is an append (runtime/string_append.c, fixture `string_append.a`). When the local holds the only reference, a count of one, no one can see the string change, so the append writes after its last byte if there's room, and resets what was cached about the old bytes (units and index). Otherwise it copies into a new string with twice the room. A loop of n appends allocates about log n times, not n. A string built that way can hold up to twice its bytes. A shared slice, a constant and a string some slice reads are never written: their room is 0, or their count is more than one.

## How this is held honest

- Every native fixture runs under ASan and UBSan (use-after-free, overflow, undefined behavior) and, separately, under a leak check for anything never freed: LeakSanitizer on Linux; on macOS the counted build, whose allocations must be its frees and its values in regions, and `leaks --atExit` on that binary for what the runtime takes from malloc outside the counts. (`leaks` alone can't see a value: the size-class allocator's chunks stay reachable from its own table.) A mutant that drops releases is caught by the leak check (`2a2e30b`).
- Every change to counting or reuse is checked by the oracle against Node, byte for byte.
- Retains and releases per fixture are counted (`adamic build --count`), and every oracle fixture's counts are checked in as `internal/oracle/counts.md`. A change that moves them fails the gate until the table is updated, so its effect shows in review as a diff of numbers, not a feeling.
