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

Each consumer decides by the call's origin, never by the iterator type it is declared to return, so
a generator declared `SetIterator<T>` is stepped as a generator and never read as a Set's iterator.

## How it runs

- **The frame.** The call makes an `adamic_generator` (runtime/generator.h): a code record naming
  the resume function, a state, a status and one cell per local and parameter. It runs none of the
  body; the first step does.
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
   (TestNoSlotTheIteratorArmWalksHoldsAGenerator).

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

- **A generator anywhere but its consumer:** stored in a variable, passed as an argument, returned,
  put in an array or a field, read through a view, or the generator function used as a value. A
  generator object as a value needs `next`, `return` and `throw` as methods, a result object, a
  retain around each step, and the cycle finder seeing its cells as slots.
- **Destructuring a generator's call, and an `Array.from` mapper.** Destructuring takes only the
  steps its pattern asks and must close the rest; the mapper waits for the same value form. `new
  Map` of a call whose yield isn't one `[key, value]` tuple, or whose pairs are held otherwise than
  the Map's, stops in those words, and so does `new Map` or `new Set` of a generator that isn't a
  direct call. Destructuring a plain `function*` is refused before that,
  by adamic/no-void-read, which reads its `void` last value as a void position.
- **A generator expression, a method, a nested declaration.** A nested generator captures its
  parent's frame, which needs a canonical environment a generator's cells can be; a method needs
  `this`.
- **A generic generator's instance whose `T` isn't one present slot**, a union or `unknown` among
  them, and a parameter typed by a view, `Iterable<T>`, which a for-of in the body can't step until
  the erased view is built (#qk8rztp S2b). tsc's own `mapIterator` stops there, at its for-of.
- **`yield*`.** Delegation forwards `next`, `return` and `throw` and closes its inner iterator.
- **A yield inside `try`**, and with it any `finally` around a yield. A throw or a return delivered
  at a yield, and `return()` running a `finally`, need resume modes the frame doesn't have yet.
- **`return value`**, a yield used as a value, `next(x)`, `return(v)` and `throw(e)`. Each is a part
  of the protocol whose values `T`, `TReturn` and `TNext` keep apart, and none is modeled yet.
- **A closure in the body capturing a local or a parameter.** Its environment would be the
  generator's cells, which a canonical closure can't own (closure.c's canonical_owner stops one).
- **`arguments` in the body**, a consumer inside an async function or a generator, and a for-of
  over a generator's call in a module that awaits at its top level.
- **A Program region in a program with a generator.** A frame can't be a region member, and which
  types it reaches isn't proven.
