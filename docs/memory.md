# Adamic's memory model

Status: **draft, written October 5, 2026 by @system_adamic** (#w318mqs), with strings already built on it (`2a2e30b`, `967d121`) and objects and arrays next. The rule above everything here: **no garbage collector**, not as a fallback, not for now. If a design needs one, the design is wrong.

## What lives where

- **Numbers and booleans** are values: a `double` and a `bool`, copied, never allocated.
- **Strings, objects, arrays, maps and closures** live on the heap behind a reference, and each carries a reference count.
- A **constant** the program spells out (`'Fizz'`, a literal object of constants, later) is static and immortal: its count is 0, and retain and release skip it. Every program's `''`, `'true'` and `'false'` are immortal too.

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

Reference counting can't free a cycle, and a garbage collector is refused (no cycle collector, ever: decided by @system_adamic, task #gsz351g). What's known:

- **Immutable data is acyclic.** A value built from `readonly` parts can only point at values that already existed when it was made, so no `readonly` structure can ever reach itself. Immutable by default is the first answer.
- **A cycle needs a write.** It forms only when an existing object's mutable slot (a mutable field, an array or map element, a closure's captured variable) is set to something that can reach that object back.
- **That's visible in the types.** A mutable slot of type `T` inside type `S` can close a cycle only if `T` can reach `S` through the type graph. The checker holds that graph, so every *cycle-capable* slot can be found at compile time.

What stage 0 does (`internal/lower/cycles.go`):

1. **The finder** walks the whole program's type graph after lowering and refuses every cycle-capable slot that isn't declared `Weak`, with the fix (`adamic/cycle-capable`). A slot is a field that isn't `readonly`, an element of an array that isn't `readonly`, a `Map`'s value, or a variable a function value captures (`let` or `const`: the cell is written after the closure captured it). Reaching follows an object's fields and the fields of every object type in the program that can be seen as it (a `Dog` seen as an `Animal` brings its `owner` along), an array's elements, a map's keys and values, and, through a function type, the variables captured by every function value in the program that can be seen as it, since a type doesn't say what a function captured. A `Weak` is not followed. A value just made (a literal, a call's result, `new`) is examined only where it's kept, since nothing writes through it before.
2. **What that means in practice.** A parent pointer must be `Weak`. A mutable child array of the same type (`children: TreeNode[]`) is cycle-capable too (`node.children.push(root)`), so a tree's children are `readonly`, or `Weak`. Either link of a doubly linked list alone can close a cycle (`a.next = a`), and the types can't tell `next` from `prev`, so both are `Weak` and something else (an array) owns the nodes. A closure kept in a variable it captures (`let countdown = ...; countdown = (n) => countdown(n - 1)`) is a cycle; the fix is a function declaration, which captures nothing. A callback field is refused only when some function value in the program captures a variable that can reach the object holding it.
3. **Spelling.** `import type { Weak } from 'adamic'`, then `parent: Weak<TreeNode>`, `Weak<TreeNode>[]`, `Map<string, Weak<TreeNode>>`, or a `let` of a `Weak` type. `Weak<Target>` is `(Target & WeakBrand) | undefined`: tsc accepts it, a plain `TreeNode` assigns into it, and a read is a `TreeNode` once narrowed, so it reads as `TreeNode | undefined`. The brand gives weak slots a type of their own, so the compiler sees them in the type graph, not just in the syntax. An array of `TreeNode` seen as an array of `Weak<TreeNode>` (which tsc allows) says NotYet: one holds targets and the other handles.
4. **Natively** (`runtime/weak.c`), a `Weak` slot holds a counted handle, shared by every weak slot pointing at one target, never the target itself, so it doesn't count. Freeing a target tells its handle, which then reads `undefined`. A side table from target to handle is how freeing finds it, consulted only while any handle exists, so a program without `Weak` pays nothing.
5. **In the JavaScript backend**, a `Weak` is a plain reference, which is what Node does with the source.

**Where native and Node differ, by design:** a `Weak` read after its target's last strong holder let go is `undefined` natively, while on Node the collector keeps the target as long as the `Weak` points at it. A read the checker had narrowed to present panics natively instead of reading freed memory. Programs that read a `Weak` only while its target is held strongly (a child's parent, while the tree is held) mean the same on both; the oracle's fixtures are such programs, and `internal/oracle/weak_test.go` pins the difference itself.

### Relaxing the finder for fresh writes (designed, not built)

The type rule is sound and, for builders of trees, costly: every parser pushes freshly made children into a mutable `children: Node[]`, and a `Node[]` whose elements can reach a `Node[]` is cycle-capable, so it's refused, and stage 1's ports (285K lines of lint rules behind them) must build each child list into a local array and freeze it `readonly`. This is the design for letting such a slot stand undeclared when the program provably never closes a cycle through it. It needs A2's control-flow graph (`internal/flow`: single assignment per function, and aliasing effects, Create, Capture, Alias and Mutate, with values escaped when a parameter, a global, a captured variable, a call's result, or a value handed to a call). Nothing below is built yet.

**What it proves.** A cycle needs a last write, the one that closes it, and every slot on a cycle is cycle-capable by type, which the finder already knows. So it's enough that at every write into a cycle-capable slot, the value written can't reach the slot's holder. Then no write closes a cycle, and by induction over the program's writes none ever forms. A slot is relaxed (allowed undeclared) only when every write into it in the whole program is proven so; any one write that isn't keeps the refusal, now naming that write. Writes are: a field assigned or set in a literal, class initializer or constructor (through any view of the holder, as the finder's `related` already matches views); an array's push, index write, `fill`, `splice`, literal and spread; a map's `set` and literal entries; and a captured variable's assignment (a cell's holder is the closure, which has always escaped, so cells are never relaxed).

**The two proofs.** Within the function that makes the write, over the whole function at once (flow-insensitively, so a capture on a later loop iteration counts), either suffices:

1. *A fresh value.* Everything the written value can reach was made in this function (a literal, `new`, or a call to a function summarized as returning fresh values, below), nothing in it has escaped (been handed to a call, stored in a global or a cell, or into an escaped value) anywhere in the function, and the holder is never captured into any of it. Then the value reaches only things this function made and saw everything done to, and the holder isn't among them.
2. *A fresh holder.* The holder was made in this function and hasn't escaped before the write, and nothing in the function ever captures the holder into the value or anything it reaches. Whatever the value reaches that was made elsewhere can't reach the holder, since nothing made elsewhere has had a chance to hold it. This is the parser's case: `const node = { kind, children: [] }; node.children.push(parseChild());` is proven without knowing anything about `parseChild`.

A function is *summarized as returning fresh values* when its result, by proof 1 inside it, reaches only what it made, its parameters' strong contents excepted, and then a call's result is fresh at a call site whose arguments in those parameters are themselves fresh there. A Weak edge is no capture at all (it keeps nothing), so a parent pointer declared `Weak` doesn't spoil either proof.

**Where it falls back to the refusal.** Any write neither proof covers: a holder that is a parameter (a helper `addChild(parent, child)`; extending proof 2 across calls, by checking every call site, is the next step if the ports need it), a holder that escaped before the write, a value read from a global, a cell or a call not summarized, a value or holder handed to a call that could tie them together, values a runtime callback produces (`map`, `forEach`, a `sort` comparator, whose writes the graph can't see into), and anything inside a generic class judged per instantiation where the flow graph has none.

**How it's held honest.** Every way to sneak a cycle past a proof is a probe that must stay refused: the holder captured into the value before the write; the same on the loop's next iteration; the value and the holder handed to one call; the holder stored in a global, then reached from the value through a function; a function summarized as fresh that really returns its argument, or something read from a global; a write through another view of the holder; a fresh-looking value that spreads an old one (`{ ...old }`); and each runtime write (`splice`, `fill`, `concat`, spread, `Array.from`). Every program it accepts runs under the leak check, which is the ground truth: a hole that lets a cycle through leaks, and LeakSanitizer reports it (with globals cleared at exit, so a cycle behind one shows). Each mutant drops one condition of a proof (ignore captures after the write, take every call's result as fresh, ignore escapes into globals, skip the views) and must turn at least one of those probes from refused into accepted and leaking.

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

`throw new Error(message)`, a caught error thrown again, and `try` with `catch`, `finally` or both, wherever statements go (functions, methods, constructors and the top level). A `catch` may leave its binding out. Still NotYet: a function value (an arrow function) that can throw, since the runtime calls them from inside its own loops (`map`, `sort`) where a throw would have to unwind C it doesn't own; and `return`, `break` or `continue` inside a `finally`, which in JavaScript overrides what the `try` was doing.

### Exceptions in the graph

The analyses that move and reuse values (`internal/flow`, and reuse in place on it) must see every way a statement can end, a throw included, or they prove a value dead that a `catch` still reads. Reviewer R's round 7 found the graph had no edges for exceptions at all: on the integration that brought `try` in, any program with `try` or `throw` failed to compile (`flow: no graph for ir.Try`).

- **The edges.** An instruction with a call to a function that can throw (`ir.Function.MayThrow`) ends its block with a `MayThrow` terminal: one edge to what follows, one to the handler. `throw` goes to the handler. The handler is the innermost `catch`, or, without one, the `finally`, or, with no `try` around it, a block ending in `Throw`, the function's exceptional exit, which a trace counts as leaving as `Return` does. A `finally` ends in a `Choose` over every way into it can go on: after the `try`, the next handler out (when a throw is being carried through it), and each `return`, `break` or `continue` routed through it. Liveness then sees that a `catch` still reads what is live there.
- **What reuse learned.** A variable that a statement assigns is only overwritten (and its old value dead) when the statement can't throw; one that can throw leaves the old value for whoever takes the throw. A global is never moved into a call in a statement that can throw: a catch, in this function or a caller, may read it and find it NULL. And a temporary handed to a consumed parameter stays the statement's own to let go until every argument is evaluated, since an argument after it can throw and then the call is never made.
- **How it's held.** The trace check's runtime stops at a panic (a Node panic is a throw a catch could take, so the trace after it isn't native's), and the graph's paths must include every path Node took through `try`, `catch` and `finally`. R's probe `move_throw.a` is a fixture: a global moved into a call that throws, a local the catch reads, a spread whose field throws, a dead-looking local the catch reads, and a field taken out of a reused spread before a later field throws. The mutant that drops the `MayThrow` edges fails it under UBSan (a NULL read in the catch) and fails `internal/flow`'s own tests.

**When regions meet exceptions (R, round 8; regions aren't on this branch yet).** A statement with a region that a throw leaves must end the region on its cleanup path, the same way it lets go of its temporaries in `checkThrown`, or the region's blocks and the heap values its objects hold leak. When the two merge, that's a fixture (a fresh call that throws mid-statement, inside a `try` and out of a function) and a mutant that leaves the region open on the throw path, caught by the leak check.

## Arenas

Some work allocates a lot and frees it all at once: one request, one file checked by cohere. For that, an arena: allocations bump a pointer, and the arena frees everything in one go at the end. A value allocated in an arena must not outlive it, and proving that is escape analysis. The lowering IR's aliasing analysis (#5jck546) is where that comes from. Arenas are for stage 1 (cohere in Adamic), where cohere's own measurements already show that with the collector off, fresh allocation is the cost.

### The design (#272q6cv, on paper)

Status: **designed October 5, 2026, by stream A2; not built.** The benchmark it answers is `bench/trees.ts`. It makes 68,332,244 objects, each its own `malloc` and `free`. Node and Bun bump-allocate them in a young generation and drop each dead tree at once. Reuse in place can't help: nothing there is consumed to build its own shape.

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

**What it leaves for later.**

- A region per loop iteration, per function call (a value made and dropped inside one call), and per request (stage 1's cohere, a file at a time).
- Arrays in regions. Their element buffers grow by `realloc`, which a bump allocator can't do in place.
- Not emitting retains and releases on values statically known to be in a region.


## Strings, specifically

UTF-8 bytes, immutable, counted. JavaScript programs see UTF-16 (`length`, indexes, `<`), so the runtime keeps UTF-16 behavior over UTF-8 storage: an ASCII-only flag makes the common case free, and other strings compute the mapping when first asked. Lone surrogates (which UTF-8 can't hold) are stored as WTF-8 and written out as U+FFFD, as Node does. Program 10 in docs/0.1.md is the fixture for all of it.

## How this is held honest

- Every native fixture runs under ASan and UBSan (use-after-free, overflow, undefined behavior) and, separately, under a leak check (`leaks --atExit` on macOS, LeakSanitizer on Linux) for anything never freed. A mutant that drops releases is caught by the leak check (`2a2e30b`).
- Every change to counting or reuse is checked by the oracle against Node, byte for byte.
- Retains and releases per fixture are counted (`adamic build --count`), and every oracle fixture's counts are checked in as `internal/oracle/counts.md`. A change that moves them fails the gate until the table is updated, so its effect shows in review as a diff of numbers, not a feeling.
