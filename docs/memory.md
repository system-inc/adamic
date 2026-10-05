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

Stage 0 doesn't reuse anything yet. It retains and releases on every assignment, the simple baseline. Reuse is added against that baseline, counted (retains and releases per fixture) and timed, and the oracle and the leak check keep it honest.

## Borrowed parameters (designed, not built)

Status: **a design on paper, October 5, 2026, by stream A2** (#r3s6nwg's follow-on, with #5jck546 answered at the end). Nothing here is in the compiler yet. The numbers come from a prototype that was measured and then thrown away.

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

### Proving it can fail

Two mutants, each needing a fixture that builds its strings at runtime (literals are immortal):

1. **Borrow a reassigned parameter.** A function that reassigns a string parameter it was handed, called with a string the caller goes on using, makes `store` release the caller's string. ASan has to catch a use after free or a double free.
2. **Borrow a closure's parameter.** A `map` callback that overwrites the element it was given, then reads its parameter, makes the parameter dangle. ASan has to catch it. That fixture is also what would let closure parameters be borrowed later. `map`'s loop would retain each element it hands out, as the other visits already do, or the aliasing analysis below would prove the callback can't write the array.

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

That recommendation comes from reading, not from measuring. What would settle it is lifting `ssa.go`'s `Construct` onto a graph built from Adamic's IR, and counting how much of it carries over unchanged.

## Cycles: the open question

Reference counting can't free a cycle, and a garbage collector is refused, so this is where the design has to earn it. What's known:

- **Immutable data is acyclic.** A value built from `readonly` parts can only point at values that already existed when it was made, so no `readonly` structure can ever reach itself. Immutable by default is the first answer.
- **A cycle needs a write.** It forms only when an existing object's mutable slot (a mutable field, an array or map element, a closure's captured `let`) is set to something that can reach that object back.
- **That's visible in the types.** A mutable slot of type `T` inside type `S` can close a cycle only if `T` can reach `S` through the type graph. The checker holds that graph, so every *cycle-capable* slot can be found at compile time.

The decision this is heading toward, for Kirk's read and not yet made:
1. **0.1 allows cycle-capable slots**, and every fixture runs under the leak check, so a cycle in a test fails it. (That's in place now: `leaks --atExit` on every native fixture.)
2. **Next:** the compiler finds every cycle-capable slot and requires it to be declared `weak`, or refuses it. A `weak` reference doesn't count, and reads as `undefined` once its target is freed. That's Swift's and Rust's answer, made checkable. A cycle collector is not on the table.
3. The survey (#db31nbn) can say how many slots in real code are cycle-capable before this is decided. Parent pointers, doubly linked lists and graphs are the usual ones.

## Arenas

Some work allocates a lot and frees it all at once: one request, one file checked by cohere. For that, an arena: allocations bump a pointer, and the arena frees everything in one go at the end. A value allocated in an arena must not outlive it, and proving that is escape analysis. The lowering IR's aliasing analysis (#5jck546) is where that comes from. Arenas are for stage 1 (cohere in Adamic), where cohere's own measurements already show that with the collector off, fresh allocation is the cost.

## Strings, specifically

UTF-8 bytes, immutable, counted. JavaScript programs see UTF-16 (`length`, indexes, `<`), so the runtime keeps UTF-16 behavior over UTF-8 storage: an ASCII-only flag makes the common case free, and other strings compute the mapping when first asked. Lone surrogates (which UTF-8 can't hold) are stored as WTF-8 and written out as U+FFFD, as Node does. Program 10 in docs/0.1.md is the fixture for all of it.

## How this is held honest

- Every native fixture runs under ASan and UBSan (use-after-free, overflow, undefined behavior) and, separately, under `leaks --atExit` (anything never freed). A mutant that drops releases is caught by the leak check (`2a2e30b`).
- Every change to counting or reuse is checked by the oracle against Node, byte for byte.
- Retains and releases per fixture are counted (`adamic build --count`), and every oracle fixture's counts are checked in as `internal/oracle/counts.md`. A change that moves them fails the gate until the table is updated, so its effect shows in review as a diff of numbers, not a feeling.
