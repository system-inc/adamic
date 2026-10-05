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

Counts are plain `size_t`, not atomic, because 0.1 is single-threaded. When concurrency lands (#p286ycm), a value shared across threads is one of the shareable, immutable kinds, and its count becomes atomic at the moment it's shared (Lean 4 does exactly this), so single-threaded code never pays for atomics.

## Freeing

Releasing an object releases what it holds. A long chain (a linked list a million long) would overflow the C stack if freed recursively, so freeing works from a list rather than by recursion. That's a correctness rule, not an optimization, and stage 0 must have a fixture that frees a chain far deeper than the stack.

## Reuse in place (Perceus)

The counts are what let Adamic mutate in place without anyone seeing:

- When a value whose count is exactly 1 is consumed to build a new value of the same shape, the new value takes over the old memory. `{ ...tree, left: insert(tree.left, value) }` on a tree nobody else holds rewrites one node instead of copying it. That's Perceus, from Koka; Lean 4 does the same.
- `push` on an array with count 1 grows it in place, and on a shared one copies first.
- **Identity is never observable through reuse.** Reuse happens only when nothing else holds the old value, so nothing can compare against it. Program 9 in docs/0.1.md pins this: while `before` still holds the old tree, the path is copied and the shared subtree stays shared.

Stage 0 doesn't reuse anything yet. It retains and releases on every assignment, the simple baseline. Reuse is added against that baseline, counted (retains and releases per fixture) and timed, and the oracle and the leak check keep it honest.

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
- Retains and releases per fixture get counted before reuse is added, so its effect is a number, not a feeling.
