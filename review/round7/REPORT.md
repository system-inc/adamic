# Stream R, round seven: B2's exceptions

I read the design in docs/memory.md ("Exceptions, designed into counting") and probed the first cut (37c3052).
Everything ran on Linux x86-64 with Go 1.27.0, clang 18.1.3 and Node v24.21.0. `review/fxspptb/probe.sh` runs
each probe four ways:

1. Node on the source.
2. Native under ASan, UBSan and LeakSanitizer.
3. Native as `adamic build` makes it (-O2).
4. The JavaScript backend.

These probes need a branch with exceptions to compile.

## Confirmed findings, worst first

### 1. On cloud/integrate-7 (9984394), no program with `try` or `throw` compiles

`smallest.a` is a function that throws, called inside a try:

    function f(): void { throw new Error(`x${1}`); }
    try { f(); } catch (error) { console.log(`caught`); }

On integrate-7 the compiler itself panics: `panic: flow: no graph for ir.Try`. Other programs give
`no graph for ir.Throw`. The panic comes from A2's `flow.Build` (build.go:155), called by the reuse planner
(`planReuse`, reuse.go:73), which runs for every program. The merge brought B2's `ir.Try` and `ir.Throw`
statements to A2's graph builder, which has no case for them.

So `go test ./internal/oracle -run .../exceptions` fails on integrate-7, along with every exceptions fixture.

**What the fix has to bring.** When the graph learns `Try` and `Throw`, every call that can throw needs an edge
to its catch, and a throw needs one too. Liveness, and so every Perceus move and every reuse in place, must see
that a catch can read a variable after the call.

Without that edge, a variable moved into a call that then throws is NULL in the catch, where JavaScript still
has it. A spread taken over and then interrupted by a throw shows the catch a half-written object.
`move_throw.a` is the probe for that fix; it covers three cases:

- `tree = insert(tree, fail(true))` with a catch that reads the global `tree`;
- a local moved into a consumed parameter whose callee throws, then read in the catch;
- `{ ...item, label: fail() }` with a catch that reads `item`.

On B2's branch, which has no Perceus, all three agree on every run. I couldn't run them where Perceus is, since
integrate-7 doesn't compile them.

### 2. `this` escaping a constructor before its fields are set: a field typed `string` is really undefined

This is older than exceptions, and exceptions widen it.

`early_this.a` has no exceptions. A constructor sets `left`, pushes `this` into a registry, and calls a function
that reads `this.right` before the constructor sets it:

| Run | Result |
|---|---|
| Node and the JavaScript backend | `left1 undefined undefined`, then `left1 right2 string` |
| Native, sanitized | `string.c:58: runtime error: member access within null pointer`, exit 1 |
| Native, -O2 | segfault, exit 139 |

It's the same on my branch's older main, so the hole predates exceptions. tsc's `strictPropertyInitialization`
checks only that a field is set by the constructor's end, not that nobody reads it first.

`half_built.a` is the exceptions version. The constructor pushes `this` and then throws before setting `right`.
The half-built object outlives the throw in the registry, with the field unset for good:

| Run | Result |
|---|---|
| Node | `caught half1`, then `left1 undefined undefined` |
| Native, sanitized | UBSan at string.c:58 |
| Native, -O2 | segfault, exit 139 |

A fix would refuse `this` escaping a constructor, or a call that receives it, before every field is set. Or it
would give every field a value before the constructor's body runs. The first matches what Adamic does
elsewhere: types are true, so a read can't see a value the type doesn't allow.

## What held, on B2's branch (97a5dbf), each agreeing on all four runs and leak-clean

- **`unwind.a`:**
  - A throw inside a for...of over a Map: the open iteration ends, and the map can be written after the catch.
  - A finally that makes and lets go of strings of its own on the way out.
  - A finally that throws over a pending error: the second error wins, and the first is let go.
  - A constructor that throws halfway through its fields, with nobody else holding `this`: no leak.
  - A Weak live across a throw and read in the catch.

  Its for...of over an array never actually threw (`out.length` never hit 6), so `uncaught.a` covers array
  loops instead.
- **`uncaught.a`:** 2,000 buffered lines, then a throw three calls deep through nested for...of loops over an
  array and a Map, inside a try with only a finally. With stdout and stderr in one stream (`2>&1`), native and
  Node are byte-identical: every line, then `finally ran`, then `adamic: panic: Error: deep 0 54`, exit 70.
- **`return_finally.a`:** a `return` whose finally throws. The returned string is let go and the error is
  caught.
- **A callback that can throw** (in a `map`) is NotYet, as the design says.

## Not covered

- The Perceus and exceptions interaction itself, since integrate-7 doesn't compile any try (finding 1).
- Library failures inside a try, which are NotYet, and panics inside a try, which the design says stay panics.
- Nothing ran on macOS or arm64.
