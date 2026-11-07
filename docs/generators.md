# Generators

A generator is a state machine over one counted heap frame, never a C stack frame kept alive and
never a list of its values made ahead (docs/rulings.md, the generators entries; #yvgst44, #qk8rztp).
This page says what stage 0 lowers today, what stays not yet built and why, and the rules the frame
keeps. It is held by `internal/oracle/generators_test.go` (every program in
`internal/oracle/testdata/generators/` against Node natively, sanitized and on JavaScript, and leak
checked) and `internal/lower/generators_test.go` (every form below that stays not yet built).

## What lowers

A `function*` lowers when it is:

- declared at a module's top level, by name, and not `async`; a generic one lowers once per
  instantiation, where it's called, as every generic function does, so two instantiations are two
  functions with frames of their own, and each instance's `T` is judged as below;
- or declared by name directly in the body of a function that is neither async nor a generator, with
  no type parameter of its own or in scope (#qk8rztp S3a). It is a nested function: bound with its
  siblings and hoisted, so a call written above it works, and its call makes its frame through the
  closure of the function that declared it, its carrier: a sibling's own, one made for the call in an
  arrow that forwards to it, or the parent's binding of it. Its body reads the parent's locals through
  the carrier's cells, live at each step, with the parent's dead-zone readiness, and may call its
  siblings, which may write them; a write inside it to a local it captured stops as not yet built;
- or written in place as an anonymous `function*` expression and called at once, through any
  parentheses, `(function* () { ... })()`, in a function that is neither async nor a generator, with
  no type parameter of its own or in scope (#qk8rztp S3b). It is a closure of its own, lowered once
  per expression, and its call makes its frame through a closure made for the call, its carrier,
  which the frame keeps as a nested generator's does; anything else around it, a comma, a
  conditional, a variable, `?.`, `.call`, a tag or `new`, is no call of it;
- given only plain, required parameters: no default, no `?`, no rest, no pattern, no `this`;
- declared, or inferred, to return `Generator<T, ...>`, `IterableIterator<T>`, `SetIterator<T>` or
  `MapIterator<T>`, where `T` is held in one present slot: a number, a boolean, a string, an object,
  an array, a map or a record;
- a body that yields only as a statement, `yield value;`, never inside `try`, and returns only with
  `return;` or by falling off its end.

Its call is consumed by one of:

- a `for...of` directly over the call, declaring one name with `const` or `let`;
- a spread of the call in an array literal, alone or among other elements;
- `Array.from` of the call, with no mapper;
- `new Set` of the call, its only argument, each value added in order;
- `new Map` of the call, its only argument, when what it yields is a `[key, value]` tuple of two
  required elements held as the Map's keys and values, each pair set in order;
- the call as a statement, which makes the generator and drops it, running none of the body.

A nested generator's call, and one written in place, is consumed only by a `for...of` over it, the
call as a statement, and the one-read `Iterable<T>` parameter below. Its spread, `Array.from`, `new
Set` and `new Map` stop as not yet built: their helper would make the frame, so it would be handed
the carrier, and as an ordinary function its parameters don't count as escaping in the environment
proof, which could put the parent's cells in its C frame while the generator held a closure into
them (Memory, #s6tt5xt).

Each consumer decides by the call's origin, never by the iterator type it is declared to return, so
a generator declared `SetIterator<T>` is stepped as a generator and never read as a Set's iterator.

A spread, `Array.from`, `new Set` and `new Map` each step the generator to its end and never read
its last result's value, so a `void` TReturn, what a plain `function*` infers and what tsc's
`mapIterator` declares, is never read there: the void rules let the call through when it is a
direct call to a generator this compiler lowers, the consumer's one argument (for the constructors,
the library's own `Map` and `Set`, judged by symbol, never a user class of that name or a subclass),
and what it yields holds no `void` at any depth, neither element of a pair included (rulings.md's
void-container classifier, as Adamic ruled on #qk8rztp; `generatorStepped`). Anything else that
reaches the TReturn, a stored generator among them, stays judged by the void rules.

A `for...of` over a parameter or local typed `Iterable<T>` or `IterableIterator<T>`, a view, lowers
when every value that can reach it is a Map's or a Set's `keys()`, `values()` or `entries()`, held
as the view reads `T`. It runs that iterator's own loop. `Iterable<string>` is held as a union,
since a string fits it too, and its loop reads the slot as the iterator through a checked unbox. A
parameter typed either `| undefined` is a view too, read past its undefined test (S2c-3).
Lowering proves it once every call is in (lower/iterable_views.go, held by
`internal/oracle/testdata/iterable_views/`). Any other origin, or one that holds its steps
otherwise, stops as not yet built: an array, a generator, a string, a Map itself, a program's own
iterator, the parameter of a function used as a value, of an object literal's method or of a host
callback, a field, a record's or a Map's value. `next(x)`, `return()` and `throw()` through a view
don't lower either.

A generator's direct call can also be passed straight into such a parameter, the argument for an
`Iterable<T>` or `IterableIterator<T>` parameter of a function the call names by its one body: a
module function, an overload set judged at its implementation, or a generic instance (#qk8rztp
S2c-1). The generator crosses into the parameter through `ir.IterableOf`, the one door from a
generator to an object's slot. The parameter's only use, counted in the lowered program, must be one
for-of over it, in a function that doesn't suspend, read `arguments`, assign it or let a closure
capture it. The parameter may be typed `Iterable<T> | undefined` or `IterableIterator<T> | undefined`
too, as tsc's `firstOrUndefinedIterator`'s is, when it is held as an object, whose undefined is a
missing pointer: any number of its `=== undefined` and `!== undefined` tests, before the for-of,
after it or in a loop's condition around it, are tests, not reads, and anything else that reads it,
a truthiness test, a null test, `typeof`, `??` or `?.` among them, still counts, as does a test
inside a closure, which captures it (#qk8rztp S2c-3, Adamic's ruling). Past its test the for-of
reads it through the narrowing's check, which the proof sees through and the loop keeps, so an
undefined the proof missed stops the program in that check's words. `| null`, `?` and a default
stay not yet built, and so does `Iterable<string> | undefined`, held as a union. That for-of becomes a dispatch on what the object is at run time: a generator steps by
its own next, a Map's or a Set's iterator by the collection loop's own step, so one parameter can be
fed both at different call sites, and anything else is a compiler bug that stops the program, naming
the site. Every way out of the loop finishes a generator, as Node's `return()` does, so a later pass
over the same parameter finds it done and its frame's cells are let go of at once; a collection's
iterator has no `return()` and is left as it is. An `Iterable<string>` view, held as a union since a
string fits it too, is read through a plain cast of the union's reference into the same dispatch,
whose kind tests stand where the checked unbox's would, so a string there still stops it loudly. A
view fed by generators of two functions stays not yet built.

An `IterableIterator<T>` parameter can also be stepped by hand, `x.next()` with no argument, as
checker.ts:21636's loop does, when every value that can reach it is a Map's or a Set's iterator
(#qk8rztp S2c-2a). The call keeps its receiver, the runtime-built iterator, whose `next` is its own
closure, so the library-member guard lets that one call by; a detached `next`, `next?.()`,
`return()` and `throw()` stay refused, and `next(x)` stays not yet built, as does a for-of beside
the nexts over the same parameter, an interleaving nothing has held to Node. The result's `done` and
`value` are read off the fresh object each call returns: `value` is `T` only where `done` has
narrowed it, and anywhere else it is `any`, which the any rules stop. A generator's call into such a
parameter stays not yet built, since its step hands back its value, not a result (S2c-2b).

## How it runs

- **The frame.** The call makes an `adamic_generator` (runtime/generator.h): a code record naming
  the resume function, a state, a status and one cell per local and parameter. It runs none of the
  body; the first step does. A nested generator's frame, and one written in place, also owns `self`,
  the carrier its call was made through, retained by its entry, so whatever made the carrier may let
  go of it before the first step (a carrier made for the call is let go of when its statement ends); finish releases it and leaves it `NULL`, so a done generator stops pinning its parent's
  cells, and the drop walk drops it at every state. A top-level generator's `self` stays `NULL`.
  Its entry is a closure's (`adamic_code`), and its resume reads the parent's cells through
  `self->cells`, as any closure's body does. The make-frame call is a `CallClosure` direct to a
  function that suspends, which the environment proof counts as keeping its carrier, so the parent's
  cells go on the heap.
- **The body.** flow.NormalizeAsync puts every local in a frame cell and every value a yield hands
  over in a snapshot, as it does an async function's; the body is cut into regions at each yield,
  as an async body is at each await, and the resume function switches on the state to the region it
  goes on in. The resume is emitted by the same block emitter as an async function's, with a stack
  check, named `adamic_generator_resume_<identity>`, and its code record is named as the
  declaration is.
- **A step.** `adamic_generator_next` runs the resume to its next way out and hands back the value,
  owned. A yield first hands over its value, then kills every counted cell its resume never reads
  again (flow.LiveIn), and suspends. A return or the body's end finishes the frame. A throw moves
  the error out of the frame into the pending exception and finishes it.
- **The consumer.** It tests the pending exception after each step, when the body can throw, before
  it reads done: a throw out of the body finishes the frame, so done would read true. The for-of's
  block holds the generator, so every way out of the loop (exhaustion, `break`, `continue` or a
  labeled jump to an outer loop, `return`, a throw from the body or the loop) lets go of it once.
  A spread, `Array.from`, `new Set` or `new Map` steps it in a helper made for the call site, whose
  local holds it; the Set or the Map is made from the values the helper collected. Node's
  constructors add each value before the next step, which is the same, since nothing can reach the
  Set or the Map before it's returned, and adding to a fresh one can't throw.

## The frame's rules

These are the generator ruling's four ownership precisions (docs/rulings.md, "Generators are a state
machine over a counted heap frame"), with how each is kept.

1. **Parameters are always owned.** The call retains every reference parameter into its cell, so a
   generator outlives the arguments its caller built for it (retained_param.a). No cell borrows
   another's value: an async frame lets a stable alias borrow its source's slot, but a yield's kill
   set can release a source while its alias is still live.
2. **Defaults run at the call.** Not yet built: a parameter with a default stops as not yet built.
3. **The first `next()`'s argument is dropped.** Every step is the compiler's own and sends nothing,
   and a yield's result, what a later `next(x)` sends in, stops as not yet built.
4. **Frame slots are under the cycle rule.** No slot outside its consumer holds a generator, and no
   closure captures a frame cell, so no cycle through a frame can be made. The cycle finder's
   MapIterator and SetIterator arm reads a type, which can't tell a collection's iterator from a
   generator declared as one; no slot it walks holds a generator's call
   (TestNoSlotTheIteratorArmWalksHoldsAGenerator). A nested generator's frame points out, through
   `self`, at its parent's cells; that edge closes no cycle while nothing those cells reach can hold
   a generator, which TestNoCapturedOrEnvironmentLocalCanHoldAGenerator reads off every generator
   and view fixture: no captured or environment local is a Generator, and no parameter one is handed
   into is captured. Owed at S2c-4, when generators become values, beside q5 on #qk8rztp (reachLinks
   linking a generator type to its frame's cells): reachLinks must also link a generator to its
   carrier's cells, the edge `self` makes, before that guard can be relaxed. The same day,
   region.go's environment proof must count every way a generator can outlive its call as keeping
   its carrier. Today a nested generator's carrier is safe because every consumer finishes with the
   frame before the parent returns, so the parent's cells never outlive what reads them, whatever
   placement the proof gives them; once a generator can be held past its call, both edges come due
   together.

A cell that holds a reference owns its value or holds `NULL`, never a stale pointer, so finishing the
frame and dropping it are each one walk over the counted cells, whatever the state. A dropped
generator runs no `finally`, as Node, which collects one without running it.

## What stays not yet built, and why

Every one of these stops as not yet built (NotYet, never Refused: generators are designed and
intended), at lowering, in its own words, wherever lowering reaches it (Lower, open question 7 on
#qk8rztp). A declaration of a form this slice doesn't lower stops where lowering reaches it, called
or not: its signature, a nested function's declaration, a function expression, or a block's
statements, each saying the form (written as an expression, a method, a static method, an object
literal's method, declared inside a function or a block, inside a namespace, without a name). No
function's body is lowered before its signature is written, so none can compile as an ordinary
function. A generic generator nothing instantiates is never lowered, so what its body holds costs
nothing, and so is a generator method of a class nothing in the whole program constructs, whose
static parts still run; the refusal walk still reads both, and refuses what is refused for good (an
async generator, rule 2400, among them).

- **A generator anywhere but its consumer:** stored in a variable, passed as an argument other
  than straight into a one-read `Iterable<T>` parameter, returned, put in an array or a field, read
  through a view, or the generator function used as a value. A
  generator object as a value needs `next`, `return` and `throw` as methods, a result object, a
  retain around each step, and the cycle finder seeing its cells as slots.
- **Destructuring a generator's call, and an `Array.from` mapper.** Destructuring takes only the
  steps its pattern asks and must close the rest; the mapper waits for the same value form. `new
Map` of a call whose yield isn't one `[key, value]` tuple, or whose pairs are held otherwise than
  the Map's, stops in the words `new Map` gives any other such source, and so does `new Map` or
  `new Set` of a generator that isn't a direct call. Destructuring a plain `function*` is refused before that,
  by adamic/no-void-read, which reads its `void` last value as a void position.
- **A generator expression, a method, and the nested declarations S3a doesn't take**: one that is
  generic or has a type parameter in scope, one declared in a block or in an async function or a
  generator, one used as a value, and one whose body writes a local it captured (S3a's next head),
  each in its own words. A generator expression stops unless it is S3b's, written in place and
  called: one stored first, one with a name of its own, which binds it inside its body, and one at a
  module's top level each stop in their own words; a method needs `this`.
- **A generic generator's instance whose `T` isn't one present slot**, a union or `unknown` among
  them, and a for-of in the body over an `Iterable<T>` parameter handed anything but a collection's
  iterator (the view above). tsc's own `mapIterator` handed an array stops, since an array is no
  origin a view admits, and handed a Map's `keys()` it lowers.
- **`yield*`.** Delegation forwards `next`, `return` and `throw` and closes its inner iterator.
- **A yield inside `try`**, and with it any `finally` around a yield. A throw or a return delivered
  at a yield, and `return()` running a `finally`, need resume modes the frame doesn't have yet.
- **`return value`**, a yield used as a value, `next(x)`, `return(v)` and `throw(e)`. Each is a part
  of the protocol whose values `T`, `TReturn` and `TNext` keep apart, and none is modeled yet.
- **A closure in the body capturing a local or a parameter**, and a nested declaration in the body.
  Its environment would be the generator's cells, which a canonical closure can't own (closure.c's
  canonical_owner stops one). This is the canonical environment a generator's cells would have to
  be; a nested generator's parent needs none, since its cells are already an environment's.
- **`arguments` in the body**, a consumer inside an async function or a generator, and a for-of
  over a generator's call in a module that awaits at its top level.
- **A Program region in a program with a generator.** A frame can't be a region member, and which
  types it reaches isn't proven.
