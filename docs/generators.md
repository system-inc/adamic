# Synchronous generator functions

This is the step 20 contract for task #qk8rztp under #yvgst44. A synchronous
function*, generator expression or generator method creates a counted iterator
whose body is a resumable state machine over an ordinary heap frame object.
The same semantics apply to .a and .ts. No collector or retained native stack
is used.

## Call and ownership

FunctionDeclarationInstantiation happens at the call. Parameter defaults and
destructuring run in source order before the generator object exists. An error
there throws from the call, not from next(). The body starts only at next().
Every parameter, captured variable and this is retained into the frame at the
call. Generator parameters are owned for borrow inference. Frame locals own
all values that survive suspension; no borrow survives a yield. Program-region
members remain plain references under the existing region rules.

The frame is an ordinary counted object. Releasing its last owner releases its
held values. Destruction never runs source finally code. A dropped generator
therefore runs no finally, exactly as in JavaScript.

## Resume and cancellation

States distinguish suspended-start, suspended-yield, executing and completed.
next, return and throw implement GeneratorResume and GeneratorResumeAbrupt.
A resume while executing throws a catchable TypeError, as Node does. A completed
next returns { value: undefined, done: true }; a completed return returns its
argument with done: true; a completed throw throws its argument. return before
the first next completes without entering the body.

The first next argument is discarded, including any reference it holds. Later
next arguments become the value of the suspended yield expression. Yielded
values, return values and sent values have separate checked types T, TReturn
and TNext. No conversion trusts one of those promises as another.

return at a suspended yield injects a return completion and runs pending finally
blocks. throw injects a throw completion at that yield and reaches the enclosing
catch or finally. finally may itself yield; the pending completion then survives
until that finally finishes. An overriding completion in finally has JavaScript's
precedence. for-of closes on break, return or a throw from its body under the
existing IteratorClose rules. Exhaustion, continue and failure from next do not
close. An incoming throw takes precedence over a failure from return.

yield* owns the inner iterator record, obtains and caches next in specification
order, and delegates the inner protocol. It forwards sent values, return and
throw; array and Set delegates use their actual iterator behavior. Missing throw
performs the required close before throwing TypeError. Missing return resumes
the outer return completion. A delegate result must be an object, done is tested
before value, and a done value becomes the value of yield*. Delegation may
suspend while handling return or throw and retains its pending completion.

## Refusals

Every diagnostic includes the source file, line and column and the named path.

- Async generator: refused, "an async generator function"; fix: "use a synchronous
  generator; async resumption awaits the async ownership and cancellation model".
- A frame slot that can reach the generator itself or anything holding it:
  refused, "cycle-capable generator frame slot <path> without weak"; fix:
  "keep the back-reference weak, or remove the owning path back to this generator".
  This applies to mutable frame locals, retained parameters, captures and this
  under the existing object cycle rule (#gsz351g).
- A suspension construct without represented typed storage or a proved iterator
  protocol: NotYet, "generator suspension <path> has no represented storage";
  fix: "use represented values and an iterator whose protocol is known". The
  compiler must identify the unsupported path before emitting either backend.

The initial lowering also keeps these paths stopped:

- A used yield whose next type excludes undefined: "a generator yield result whose next argument is not proved present; use a
  next type admitting undefined". JavaScript permits next() without an
  argument. A narrower next promise needs a per-resume proof before an unbox can
  trust it. Yield, return and next storage remain separate.
- An unknown generator receiver or delegate: "generator suspension receiver has
  no proved generator factory origin" or "generator suspension delegate has no
  proved generator factory origin". Opaque structural protocol views do not prove
  the native receiver convention. Closed generator factories and their aliases
  preserve it.
- A protocol replacement through the generator or a wider object alias:
  "generator suspension protocol replacement needs a proved receiver convention".
  The factory does not prove the calling convention of a replacement method.
- A throw argument outside represented Error completions: "generator throw
  argument needs the represented Error completion protocol". Primitive abrupt
  completion values require a represented exception-value extension.
- A nested closure capturing a body local: "generator suspension nested capture
  <name> needs a weak frame reference". Retaining the frame from a closure kept
  in that frame would create a strong cycle; native stack cells cannot survive
  suspension as a substitute.
- Unproved initialization, captured per-iteration cells, labeled suspension
  targets, a yield in a switch case test, a short-circuit yield, and a string or non-array for-of inside a
  generator identify their named suspension path. Their existing eager lowering
  cannot be reused by leaving a borrowed local on the native stack.

A synchronous generator is never accepted by erasing yield, eagerly executing
its body or collecting its yields into an array. Unsupported paths remain
visible; they do not turn into partial execution or unchecked frame storage.

## Proof

Node decides values, order, errors and completion precedence. Both backends run
basic sequences, sent values, cancellation and injected throws, for-of close,
drop without close, array/Set/generator delegation, call-time defaults,
reentrant resume, and reductions of mapIterator and a checker generator.
Sanitizers and leak checks hold frame ownership. Mutants skip finally on return,
fail to forward delegate return, keep a borrow across yield, borrow a parameter
after its caller releases it, defer a default to next, leak the first next
argument into the body, and admit a cycle-capable slot without weak.
