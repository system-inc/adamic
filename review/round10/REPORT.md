# Stream R, round ten: stream C's integer loop counters, and a try that borrowing can't see into

I ran everything on Linux x86-64 with Go 1.27.0, clang 18.1.3 and Node v24.21.0. `review/fxspptb/probe.sh` runs
each probe four ways:

1. Node on the source.
2. Native under ASan, UBSan and LeakSanitizer.
3. Native as `adamic build` makes it.
4. The JavaScript backend.

## Finding

**On cloud/integrate-7-flow, a parameter reassigned only inside a `try` crashes the compiler.** I saw it on
3ab8f94 and again on b2ade88.

`try_assign.a`'s `relabel(box: Box, ...)` does `box = { ... }` inside a `try`. `adamic c` panics:
`native: a store into the borrowed parameter box` (emit.go:500).

The cause: `findAssigned` (lower/borrow.go:34) walks If, Loop, Block, ForOf and Switch, but not the
exceptions' `ir.Try`. So the write inside the try isn't seen, the parameter is borrowed, and the emitter then
refuses the store. It's loud, not a miscompile, but a valid program doesn't compile.

The fix is to visit Try's Body, Catch and Finally in `findAssigned`. That one function also decides borrowing
and loop counters (below), so both depend on it.

## Stream C's integer loop counters (cloud/hx74vgw-loops, 952898d): no findings

Read: lower/counters.go and `adamic_array_at_integer`.

- A counter is kept in an int64 only when all of these hold: it is declared in the loop with a whole constant
  within 2^53, not -0; its one update is `+ 1`; nothing else in the body writes it; it's neither captured nor
  global.
- The bound must be a size, a whole constant (≤ 2^53 for `<`, below it for `<=`), or a variable that nothing
  in the program writes. `findAssigned` runs over every function, closures included.
- Every read of the counter as a number is `(double)`, so comparisons and arithmetic are JavaScript's.
- The integer index checks `index < 0` before its unsigned bound.

`counters.a` agrees on all four runs, with six counters kept as integers in the generated C. It covers:

- a bound variable written by a closure during the loop: it keeps a double;
- a start of -2 indexing an array;
- a start folded from `0.1 * 10`;
- an inner bound that is the outer counter;
- an array popped while the loop reads its length;
- a counter running to 2^53 - 1 by `<` and `<=`, printed with `+ 1` and `+ 2`;
- counters as Map keys;
- `1 / counter` at 0.

The guards whose mutants no oracle can catch are those that only decide termination at 2^53: the `<=` limit,
and a written bound. A loop JavaScript never ends can't be compared.

When the counters meet exceptions, `findAssigned` must see into Try. Otherwise a counter written inside a
`try` in the loop body is kept in an integer, and the emitter refuses the write: loud again.
`try_bound.a`, a bound assigned only inside a try, agrees on integrate-7-flow, which has no counters yet.

Counting down isn't on the branch yet.

## B3's relaxation of the cycle finder

Not pushed yet: no branch carries it.
