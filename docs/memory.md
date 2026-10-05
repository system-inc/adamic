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

## Arenas

Some work allocates a lot and frees it all at once: one request, one file checked by cohere. For that, an arena: allocations bump a pointer, and the arena frees everything in one go at the end. A value allocated in an arena must not outlive it, and proving that is escape analysis. The lowering IR's aliasing analysis (#5jck546) is where that comes from. Arenas are for stage 1 (cohere in Adamic), where cohere's own measurements already show that with the collector off, fresh allocation is the cost.

## Strings, specifically

UTF-8 bytes, immutable, counted. JavaScript programs see UTF-16 (`length`, indexes, `<`), so the runtime keeps UTF-16 behavior over UTF-8 storage: an ASCII-only flag makes the common case free, and other strings compute the mapping when first asked. Lone surrogates (which UTF-8 can't hold) are stored as WTF-8 and written out as U+FFFD, as Node does. Program 10 in docs/0.1.md is the fixture for all of it.

## How this is held honest

- Every native fixture runs under ASan and UBSan (use-after-free, overflow, undefined behavior) and, separately, under a leak check (`leaks --atExit` on macOS, LeakSanitizer on Linux) for anything never freed. A mutant that drops releases is caught by the leak check (`2a2e30b`).
- Every change to counting or reuse is checked by the oracle against Node, byte for byte.
- Retains and releases per fixture are counted (`adamic build --count`), and every oracle fixture's counts are checked in as `internal/oracle/counts.md`. A change that moves them fails the gate until the table is updated, so its effect shows in review as a diff of numbers, not a feeling.
