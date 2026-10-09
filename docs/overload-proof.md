# Per-overload body proof

Contract for #mmqp7em, #z7vszfz and #f2ng8fc, toward roadmap step 05's
hidden result boundaries and step 16's overload proofs. This specifies the
required behavior; it does not claim the implementation already provides it.

## One overload, one input domain

For each overload whose result is narrower than its implementation annotation,
analyze the original implementation body in a separate proof environment.
Bind each implementation parameter by position to the corresponding overload's
admitted argument type, including undefined for optional or omitted arguments.
Other parameters narrow too, even when their types happen to be unchanged.
Never substitute the types of one observed call: the proof covers every argument
combination admitted by that overload. Check input compatibility first using
Adamic's contravariant parameter relation; narrowing cannot repair an overload
whose arguments the implementation cannot accept.

Generic binders must be related consistently and proved for every admitted
instantiation. Defaults run when an argument is undefined, before the body;
rest parameters represent the admitted remaining arguments. Destructuring must
preserve the corresponding component types and default evaluation. If these
bindings cannot be modeled soundly, the narrower result remains unproved.
Only the proof environment changes. The implementation, argument evaluation
order, object identity, captures and number of invocations stay the same.

## Paths and results

Walk statements in source execution order, tracking parameter and local values,
assignments and branch facts. Follow if/else, conditional expressions and switch
(including fallthrough), and propagate all live states at joins. Loops require a
sound invariant over every iteration, including zero iterations where possible.
Honor break, continue, return, throw and finally destinations. Discard a path
only when its impossibility follows from the narrowed input domain and valid
flow facts. An unsupported construct or exhausted proof budget is unproved,
never evidence of unreachability.

At each reachable normal exit, relate the actual returned expression's type
under that path's facts to the overload's promised result with Adamic's own
sound relation. Bare return and reachable end mean undefined. A return through
a local uses its reaching assignments, not its wider annotation alone. Readonly
fields relate covariantly; writable fields and mutable payloads remain invariant.
Every field must hold, including result.kind and result.value. Preserve existing
storage and ownership compatibility requirements: a type proof alone cannot
reinterpret a differently represented field or copy a shared result.

Assignments replace facts; alias writes and effects invalidate facts they can
change. A wider annotation is neither a narrower proof nor a reason to reject
an independently proved returned value. Casts and trusted TypeScript predicates
cannot supply evidence without their existing Adamic verification.

An escaping early throw has no normal result obligation. A catch may turn it
into a normal return, and finally may replace a return or throw, so both must
be walked. A proved nonreturning call ends its normal path; an ordinary call
that might throw still has its normal continuation.

## Receivers, closures and calls

**Proposed conservative answers where the ruling is silent:** an explicit this
parameter narrows to the overload receiver contract separately from positional
arguments. An implicit method receiver keeps its established class type; an
unknown or incompatible receiver cannot justify a narrower result.

Nested function and arrow returns belong to those functions, not the outer
implementation. Creating a closure does not execute its body. Called closures
and helpers need a known target and checked result/effect contract; unknown
calls invalidate affected flow facts. Captured writes count even when hidden
behind aliases. Returning a closure requires proving its callable contract,
not assuming that today's captured value survives future calls.

A call to another overload may use a narrower result only after that callee's
body proof is established for its admitted inputs (or an existing sound result
contract already suffices). A TypeScript call-site check is not a static proof
for Adamic. Recursive proof dependencies cannot certify each other by assuming
the promises under investigation; without an independent sound summary they
remain unproved. Ordinary checked helper results can suffice, as with evaluate's
TemplateExpression helper. These choices deliberately permit conservative
refusal where the implementation cannot establish the required facts.

## Admission and refusal

The step 05 TypeScript hatch retains its call-site check of the failed readonly
field when proof is absent: Block/ConciseBody checks result.kind, and
EvaluatorResult checks result.value. Failure names the overload, call site,
field and both types; emitted checks remain visible in --explain-checks and
counts. Adamic admits a narrower overload without an overload-result check
only after proof succeeds; otherwise it refuses and names the failing result
path and return location or unresolved proof obligation. All overload promises
must hold before an overloaded Adamic value escapes through a factory or alias.

The reduced positive witnesses must retain the decisive correlations:
transformAsyncFunctionBody's non-arrow domain returns a complete Block despite
its ConciseBody annotation; evaluate's TemplateExpression domain returns only
string or undefined in result.value despite the wider EvaluatorResult annotation.
Both .a witnesses must agree with source Node and emit no overload-result check.

These narrowed bodies remain refused, even if a sample invocation is harmless:

```typescript
interface Result<T> { readonly value: T; }
function evaluate(kind: 'template', bad: boolean): Result<string | undefined>;
function evaluate(kind: 'template' | 'numeric', bad: boolean): Result<string | number | undefined> {
    if (bad) return { value: 1 };
    return { value: 'text' };
}
```

The diagnostic must name result.value and the reachable numeric return.
Likewise, returning an expression body from a non-arrow input fails result.kind;
reassignment, captured mutation, a wider helper result or reachable undefined
fallthrough cannot be hidden by the overload's declaration.

## Required evidence

Run both positive witnesses against Node and assert absence of inserted overload
checks. Run a negative variant with a reachable wider return and assert refusal
at its field path. Mutate the proof to initialize parameters with the full
implementation types instead of overload types, restore it after testing, and
require the focused fixture suite to fail.

**Evidence clarification:** full-domain analysis is more conservative than
narrow-domain analysis. Alone it cannot make a correctly refused negative
program become accepted; it should wrongly refuse a positive witness. The
brief's requirement that this mutant makes the refusal fixture fail is interpreted
as a fixture test containing both admission and refusal controls. Report which
positive control kills it, separately from the wider-return refusal. If a strictly
negative-only mutant failure was intended, that needs a different mutant; this
contract does not claim an impossible acceptance from broadening the proof domain.
