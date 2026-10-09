# Proposed contracts for the next overload stops

This is a design request for a ruling, based on compiler 3a501d3a. It changes no
compiler acceptance. 01 and 13 remain checker stops under stricter options;
05 remains on checked views. The two proposals below concern only 14's escaped
overloaded function and 06's visitor input domain.

## 14: an overloaded function returned as a value

The site is utilities.ts:11445:12, `return evaluate`. Its declared public value
has both call signatures: TemplateExpression gives
EvaluatorResult<string | undefined>; Expression gives
EvaluatorResult<string | number | undefined>. The implementation declares the
wider result. A resolved call must retain the narrower signature even after the
function passes through createEvaluator's return value and a local alias.

Smallest reduction keeping the returned function, both signatures and failed
readonly field relation:

```ts
interface Result<T> { readonly value: T; }
function make(value: string | number | undefined) {
    function evaluate(expr: string): Result<string | undefined>;
    function evaluate(expr: string | number): Result<string | number | undefined>;
    function evaluate(expr: string | number): Result<string | number | undefined> {
        return { value };
    }
    return evaluate;
}
const evaluate = make(1);
console.log(`${evaluate('text').value}`);
```

Observed on Node v24.19.0: exit 0, stdout `1\n`, empty stderr. The declared
narrower result does not enforce itself. The current compiler stops at the
reduction's return, 8:12, with `stage 0 can't lower an indirect value of an
overload requiring a checked implementation boundary yet`.

Replacing the last two lines with the following also preserves a useful
identity observation:

```ts
const evaluate = make('word');
const alias = evaluate;
console.log(`${alias === evaluate}`);
console.log(`${alias('text').value}`);
console.log(`${alias(1).value}`);
```

Node exits 0 and prints `true\nword\nword\n`. Returning or aliasing the function
does not create a new function or discard its captures.

The value's logical callable type must remain the complete overload set:

```ts
type Evaluator = {
    (expr: string): Result<string | undefined>;
    (expr: string | number): Result<string | number | undefined>;
};
```

It cannot be replaced by the implementation's single wider signature while
callers continue to receive the narrower result type. Nor is the wider
implementation signature proof that the value satisfies both signatures. Its
runtime representation can stay the original counted closure; the compiler
must retain the public call signatures and the corresponding implementation
contract separately from that representation.

Proposed rule: allow the escaped value in a closed program when its underlying
implementation declaration is known and every call through the value retains a
resolved signature. Start with a single known implementation declaration per
value, including multiple factory instances with different captures. Do not
admit unknown targets or opaque host invocations by guessing their contract.
Returning the closure or copying its reference to an alias preserves its
identity and captures; do not create a replacement function at the return or
alias assignment.

At each resolved invocation, independently establish that the implementation
accepts that signature's arguments. Invoke the original closure once using the
implementation's argument and result representations. Preserve callee-before-
argument evaluation, argument order, the original argument count where
observable, and receiver semantics; retain a stop if the ABI cannot preserve
an observed distinction. Then apply the already ruled result boundary before
exposing the signature's narrower result. A proof erases that check. Otherwise,
in .ts, check exactly result.value for string or undefined after the call and
before any narrower read. The wider signature does not acquire that check merely
because another signature is narrower. Record each emitted check in the existing
checked-site report. Failure exits 70 naming overload, indirect call site,
field and both result types, as the direct-call rule already requires.

For .a, the exposed overloaded value needs per-overload return proofs for its
public promises. This proposal does not authorize the unproved narrower promise
or an unchecked function-type cast. Unknown target/signature relationships
remain NotYet; a known unserved result keeps its relation refusal. If a function
crosses a boundary where the compiler cannot instrument its calls, retain the
stop rather than lose the promised check.

The closed target and overload correspondence must survive the factory return
and local alias in this reduction. A checked entry for a direct function name
alone does not serve calls through that value. More general storage or target
sets require a separate extension preserving the same correspondence.

Requested ruling: may this complete overloaded value escape under those closed
call-site contracts, with .ts checks at invocation and .a requiring proof? This
would extend the earlier field-check ruling to calls through values, without
changing the check's placement or wrapping the observable function identity.

## 06: a visitor accepting TIn, consumed by an implementation annotated Node

The site is visitorPublic.ts:196:5, reached from esDecorators.ts:1250:13. The
overload couples NodeArray<TIn> with Visitor<TIn, Node | undefined>. The
implementation annotates its visitor as Visitor<Node, Node | undefined>.
Ordinary callback contravariance requires Node to fit TIn, which is false.
TIn extends Node establishes the opposite relation and cannot reverse it.

Smallest reduction keeping that relationship and an observable required field:

```ts
interface Base { readonly kind: number; }
interface Named extends Base { readonly text: string; }
function visit<T extends Base>(nodes: readonly T[], visitor: (node: T) => void): void;
function visit(nodes: readonly Base[], visitor: (node: Base) => void): void {
    for (const node of nodes) visitor(node);
}
visit<Named>([{ kind: 1, text: 'modifier' }], node => console.log(node.text));
```

Observed on Node: exit 0, stdout `modifier\n`, empty stderr. The current compiler
refuses at 3:53: `overload 1 of visit parameter visitor cannot be served by
implementation parameter visitor; callback contravariance at callback.parameter1
requires Base to fit T`. Its standard diagnostic includes the proposed fix to
accept every overload argument without mutable widening or bivariance.

Replacing the loop with this implementation statement changes only the
callback's input origin:

```ts
visitor({ kind: 0 });
```

Node still exits 0, but prints `undefined\n`: the callback receives an object
without its required text field. Thus an otherwise legal overload call does not
prove that the implementation invokes the supplied visitor safely.

The visitor value must retain `(node: TIn) => VisitResult<Node | undefined>`.
VisitResult here includes a single Node, undefined, or readonly Node[]. The
actual modifierVisitor accepts ModifierLike and returns a Modifier, undefined,
or a permitted node array. Its covariant result does not widen its input to
arbitrary Node. A callback accepting TIn is not an ordinary Visitor<Node>, even
if both closures use the same physical calling convention.

Proposed rule: prove the implementation's actual visitor invocation domain
under the overload's coupled arguments, then specialize the consumer to retain
that relationship. Its visitor slot is logically Visitor<TIn, Node | undefined>
within that specialization. Every reachable invocation must receive a proven
TIn. Reads of the original input array may establish that fact only while its
element contract survives aliases, writes and intervening calls. A readonly
annotation alone is not an immutability proof. Undefined reads must be excluded
before invoking the visitor. No global callback subtype relation changes.

The proof must follow helper calls. In the real source visitNodes forwards the
same nodes and visitor to visitArrayWorker, whose invocation is at
visitorPublic.ts:346. That argument comes from nodes[i + start] and is guarded
against undefined. The proposed proof carries the TIn element relationship
through that helper; it must not substitute the broader Node constraint for
TIn. Do not substitute TIn for every Node occurrence in the implementation:
only values established by the invocation-domain proof carry that refinement.
The callback's argument conversion must also preserve the existing storage and
ownership rules; a domain proof alone does not authorize a different field
layout or an unchecked view. A fabricated Node, an unproved helper argument, a write that invalidates
the element contract, or storing/passing the callback where it can be invoked
on arbitrary Node defeats the proof and retains the refusal. Mere forwarding
of the callback's wide annotation is insufficient.

For this proposed proof route there is no inserted visitor-input check: the
proof establishes the argument before each callback invocation. Apply the same
input-proof requirement to .ts and .a. If it cannot be established, report the
unproved callback argument path and invocation site, in addition to the failed
Node-to-TIn relation. The reduction's fabricated-node implementation must be
refused rather than made safe by trusting its declaration.

A future checked alternative would need a complete, reifiable TIn membership
contract, not just the Node constraint or a matching kind. Its check belongs
immediately before visitor(argument), before the callback reads its narrower
fields, and would need the resolved overload and caller's concrete TIn. Bare
TIn has no such runtime test. Full ModifierLike and other structural inputs
may need checked views; this proposal does not request that mechanism or move
05. Where the domain proof fails, keep the refusal until that separate contract
is ruled and available.

The callback input proof also does not prove visitNodes' narrow NodeArray<TOut>
result. The test predicate and every returned element, including the branch
that returns the original array, need their own result proof or an approved
check. Source assertion annotations or a boolean test's signature are not
independent evidence. NodeArray metadata and aliasing must remain preserved.
This proposal removes only the unjustified blanket input comparison when the
actual invocation domain is proved; it does not declare the whole consumer
sound merely because that first comparison passed.

Requested ruling: may an overload consumer serve Visitor<TIn, R> despite its
wider implementation annotation when all actual invocations are proved to use
TIn, including through helpers, with unproved domains still refused? I recommend
this proof route first, rather than introducing a generic input hatch.

## Evidence and obligations after a ruling

The reductions above were authored for this contract and run directly as .ts on
Node. Commands used `node /tmp/overload-next-contract/{14,14-valid,06,06-liar}.ts`
individually, saving each output to its own log. Every run exited 0; the outputs
are recorded above. The existing compiler was invoked on 14.ts and 06.ts to
confirm their stop families. No fixture registry, compiler source, census count
or other region was changed; no implementation or mutant is claimed here.

If approved, 14 needs valid and planted-wrong results through returned values
and aliases, identity/capture and evaluation-order observations, .a refusal
pairs, and a mutant removing the indirect call's field check. 06 needs the
original-array visitor and the fabricated-node refusal, forwarding through a
helper, and mutants erasing TIn to its Node constraint or admitting an unproved
invocation. Both groups need Node observations and the sanctioned refusal/stop
assertions in both backends with sanitizers before another implementation push.
