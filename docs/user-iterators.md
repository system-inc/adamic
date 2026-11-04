# User-defined iteration in stage 0

An explicit iterator uses the existing ownership model: the consumer holds the iterable, the
iterator, the cached `next` function, and each result while it reads it. A literal method receives
its object as `this`; it does not capture its owner. Other captures remain reference-counted cells
and participate in the cycle finder. Class methods retain their existing static dispatch.

`for-of` is lazy. It caches `next` once, tests `done` before reading `value`, and gives every loop
iteration a fresh binding. Break, return, and a throw from the body call the current `return`
method when there is one. Exhaustion, continue, and a failure from `next` do not. An incoming throw
wins over a throw from `return`; a close failure replaces a break or return completion.

Spread consumes the iterator to exhaustion. `Array.from` maps between steps and closes if the
mapper throws. Destructuring declarations take only the requested steps, skip value reads for
holes, keep exhaustion sticky, evaluate defaults in binding order, and close a partially consumed iterator.
Rest consumes what remains. Tuple yields also support destructured loop bindings. Assignment
and parameter destructuring of custom iterables remain stage-0 gaps. A yielded type that excludes
`undefined` needs a default in a fixed binding: TypeScript otherwise pretends early exhaustion cannot produce `undefined`.

The supported protocol has required, zero-argument methods with concrete class or literal/arrow
origins and represented object results. `done` is a required boolean; `value` is a required field
with a single-slot representation. Optional methods, structural method signatures which erase
receiver conventions, protocol replacement, iterator object spreads, and incompatible views are
explicit stage-0 gaps. `Array.from` does not yet accept `thisArg` or mappers with more than two
parameters; mapper arguments must preserve their native representations. Known constructors,
literal factories, and immutable factory aliases preserve method origins.
Generic class views also need proven, invariant type arguments because their methods are
monomorphized. Other views use conservative whole-program shape checks: an unrelated compatible shape can
make a view unsafe. Distinct discriminants distinguish such shapes. Built-in consumers
continue to use their existing lowering; this unit does not expand the existing array/string
`Array.from` or destructuring forms.

## Generators need owned suspended frames

Generators can fit reference counting without a collector, but not by keeping a C stack frame
alive or materializing all yields. A future lowering needs a heap frame whose program counter
and typed live slots form a discriminated state machine. Every value live across a suspension
must be owned or proven immortal; a borrowed caller value cannot survive its caller. Captured
locals and frame fields must enter the cycle analysis, including frames returned through erased
iterator views. Frame destruction releases held values mechanically; it must not run user
`finally` code, since JavaScript's garbage collection does not close an abandoned generator.

`next`, `return`, and `throw` need explicit resume modes and pending completions. A generator can
yield from `finally`, so `return` need not finish it. Delegating `yield*` needs its own iterator
records and completion precedence. Reentrant resume must raise the same catchable failure as
Node, rather than corrupt the frame. Each resume path must release exactly the slots it leaves
and preserve the ones needed by a later resume. No collector is proposed as a fallback.

The first slice (#qk8rztp S1c) builds that frame for one shape: a named, non-generic `function*`
declared at a module's top level, with plain parameters, whose body yields only as a statement
(`yield value;`) and never inside `try`, consumed by a for-of directly over a call to it, a spread
or `Array.from` of the call, or made and dropped by a call as a statement (docs/generators.md). The body runs on runtime/generator.h's counted frame, every
local in a cell, cut into regions at each yield as an async body is at each await; a yield kills
the cells its resume never reads. Each consumer steps the call by its origin, whatever iterator
type it is declared to return, and a step that throws is tested before done. The consuming loop's block
holds the generator, so every way out of it lets go of the frame once, which is the whole of
IteratorClose while no `finally` exists. Every other generator form stops as not yet built
(NotYet, not Refused: generators are designed and intended): expressions, methods, nested and
generic declarations at rule 2500, and in their own words at lowering a generator held anywhere
but its consumer, destructuring, `new Set` or `new Map` of one, an `Array.from` mapper, `yield*`, a
yield's result, a yield inside `try`, `return` with a value, a closure capturing a body local, and a
consumer inside an async function or a generator.
