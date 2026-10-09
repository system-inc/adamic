# User-defined iteration in stage 0

An explicit iterator uses the existing ownership model: the consumer holds the iterable, the
iterator, the cached `next` function, and each result while it reads it. A literal method receives
its object as `this`; it does not capture its owner. Other captures remain reference-counted cells
and participate in the cycle finder. Class and literal methods dispatch through their runtime receiver and method tables.

`for-of` is lazy. It caches `next` once, tests `done` before reading `value`, and gives every loop
iteration a fresh binding. Break, return, and a throw from the body call the current `return`
method when there is one. A continue to an outer label also closes. Exhaustion, an ordinary continue,
and a failure from `next` or from reading `done` or `value` do not. An incoming throw
wins over a throw from `return`; a close failure replaces a break or return completion.

Spread consumes the iterator to exhaustion. `Array.from` maps between steps and closes if the
mapper throws. Destructuring declarations take only the requested steps, skip value reads for
holes, keep exhaustion sticky, evaluate defaults in binding order, and close a partially consumed iterator.
Rest consumes what remains. Tuple yields also support destructured loop bindings. Assignment
and parameter destructuring of custom iterables remain stage-0 gaps. A yielded type that excludes
`undefined` needs a default in a fixed binding: TypeScript otherwise pretends early exhaustion cannot produce `undefined`.

The supported object protocol accepts structural interfaces, class methods, literal methods,
arrow functions, method getters, optional `return`, and optional boolean `done`. Dispatch caches
`next` once with its receiver. It reads the current `return` only when closing. Discriminated
result unions can give the completion value a different representation from yielded values;
completion values never enter the body. Concrete protocol methods must take zero parameters
and return represented objects. Generic class views still need invariant native arguments.
Conservative whole-program checks refuse hidden incompatible methods and declared result
fields absent at runtime. Optional/default method parameters need argument padding before
admission. Object views of built-in storage and native iterator objects still need a protocol adapter; direct built-in
arrays, tuples, strings, maps, sets and their iterators keep their existing lowering.

Malformed `.ts` protocols and non-object close results await the named ruled-divergence
infrastructure, `compiler/per-backend-stops` at `5f3b2e36`, absent from this base. They must stop
with exit 70 until step 21 supplies catchable TypeError. The `.a` witnesses containing unchecked
casts are refused before code generation. This unit does not claim their `.ts` backend proof.
`Array.from` still does not accept `thisArg` or mappers with more than two parameters.

## Synchronous generators

Synchronous generators use ordinary counted heap frames, resume states and separate
next, return and throw modes. Parameters, captures and this are owned at the call;
defaults and destructuring run before the iterator is returned. Cancellation runs
pending finally blocks, while dropping an iterator only frees its held values.
Array, Set and known generator delegation retain the inner protocol state.

The supported paths, explicit refusals and Node and ownership proofs are in
[generators.md](generators.md). Async generators remain refused. A closure kept
in a suspended frame must not create a strong cycle back to that frame.
