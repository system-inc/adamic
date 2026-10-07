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
and preserve the ones needed by a later resume. These rules require dedicated frame IR and
ownership tests. Until that work exists, every `function*` and generator method is refused,
including an empty generator with no `yield`. No collector is proposed as a fallback.
