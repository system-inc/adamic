# Generic functions used as values

Ruled by @system_adamic on #s4e4reh, step 16 (#s4e4reh, wall #rkxvjfx). The contextual specialization and identity rule is approved, with the const-alias
exception and higher-rank refusal below. Based on `compiler/after-chain-fixed-verify` at `50654a40`. Builds on its generic substitutions and instantiated body
relations, and on V4's identity-preserving checked callable views.

## Survey

The [step 16 ledger](../review/compiler/lowering-chain/source-member-4/step-16-generics/baseline.json.gz)
actually records **26 attempted roots and 14 distinct diagnostic sites**, the
reverse of the brief's wording. The 4,031 hidden bytes are historical attribution,
not bytes retired by this proposal. Roots include imported helpers and overload
attempts; neither root count nor site count is a count of executable programs.
The table covers every root/site pair for this exact reason. Positions refer to
the ledger's pinned TypeScript source (`050880ce`). `EqualityComparer<S>` means
`(a: S, b: S) => boolean`. No surveyed site uses an explicit instantiation
expression. A default inside a generic body becomes concrete only when the
containing function has a concrete instantiation, never from its constraint.

| Attempted root(s) | Value site and function | What fixes the binder |
| --- | --- | --- |
| binder.ts:635 | core.ts:220, `equateValues` through `appendIfUnique` | `EqualityComparer<Declaration>` in the instantiated `contains` default |
| binder.ts:1373 | core.ts:220, `equateValues` | `EqualityComparer<FlowNode>` |
| checker.ts:2698 | core.ts:220, `equateValues` through uniqueness helpers | `EqualityComparer<Declaration>` through nested duplicate-location collection |
| checker.ts:2791 | core.ts:220, `equateValues` through `pushIfUnique` | `EqualityComparer<Declaration>` |
| checker.ts:5229, checker.ts:5249 | core.ts:220, `equateValues` through nested `visit` | `EqualityComparer<Symbol>` |
| checker.ts:5737, checker.ts:5762 | core.ts:220, `equateValues` through nested symbol-table walk | `EqualityComparer<SymbolTable>` |
| checker.ts:6101 | core.ts:220, `equateValues` through `appendIfUnique` | `EqualityComparer<LateVisibilityPaintedStatement>` |
| checker.ts:6769 | core.ts:220, `equateValues` in type-to-node walk | `EqualityComparer<Type>` for `context.inferTypeParameters` |
| checker.ts:7718 | core.ts:220, `equateValues` | `EqualityComparer<ReverseMappedSymbol>` for `context.reverseMappedStack` |
| core.ts:1370 | core.ts:1370, `equateValues` | Default slot `EqualityComparer<T>`; outer T must be fixed by a caller. This standalone generic census attempt supplies none |
| core.ts:814 | core.ts:814, `equateValues` | Default slot `(a: T, b: T, index: number) => boolean`; caller fixes outer T. Extra index is ignored as on Node |
| utilities.ts:10629 | utilities.ts:10629, `equateValues` | Default slot `EqualityComparer<T>`; caller fixes outer T |
| core.ts:2377 | core.ts:2378, `identity` | Declared return `GetCanonicalFileName`, through conditional arm: T = string |
| core.ts:2442 | core.ts:2442, `identity` | Default slot `GetCanonicalFileName`: T = string |
| factory/parenthesizerRules.ts:675 | factory/parenthesizerRules.ts:676, `identity` | Contextual arrow result `(leftSide: Expression) => Expression`: T = Expression. Other fields in this root need their own proofs; the census stops at this first site |
| program.ts:1075 | program.ts:1076, `getTypeReferenceResolutionName` | Object field `getName(entry: FileReference \| string): string`: T = FileReference \| string, not merely its coincidentally identical constraint |
| sourcemap.ts:820 | sourcemap.ts:821, `identity` | Object field `(input: DocumentPosition) => DocumentPosition`: T = DocumentPosition; the second field is the same shape |
| transformers/classFields.ts:1956 | transformers/classFields.ts:2071, `startOnNewLine` | Array `forEach` slot `(value: Expression, index: number, array: Expression[]) => void`: T = Expression. Callback result is discarded, not converted to undefined inside its body |
| path.ts:1039, path.ts:1045, path.ts:1047 | path.ts:1049, `identity` | **Nothing at this value site**: checker infers the local as `GetCanonicalFileName`, but the unannotated conditional has no contextual type for this arm. Later `getPathComponentsRelativeTo` calls require string, but do not supply value-site context. Fix: annotate the local `getCanonicalFileName: GetCanonicalFileName` |
| checker.ts:9845 | debug.ts:251, `assertIsDefined` | **Nothing**: `AnyFunction = (...args: never[]) => void` is a stack marker, not a binder inference proof. Do not guess never/unknown or erase its assertion. Fix: use a concrete proven marker function, or a separately designed identity-only marker representation |
| factory/utilities.ts:1370 | factory/utilities.ts:1372, `enter` | **Nothing**: a switch case compares function identity; it supplies no concrete call signature |
| factory/utilities.ts:1392 | factory/utilities.ts:1394, `enter` | **Nothing**: array element `BinaryExpressionState` is itself `<TOuterState, TState, TResult>(...) => number`. Outer TState fixes neither the slot's fresh binders nor all of enter's binders. Fix: make the state type/array concrete in all three enclosing arguments, including identity comparisons |

Thus 20 roots reach contextual slots (three require a concrete outer caller),
and six reach the four unresolved sites. This is a source classification, not
an after-census or a claim that the entire contextual roots will compile.

## Ruled rule

At a generic function **value expression**, obtain one complete checker-derived
substitution from an explicit `f<S>` instantiation or a contextual, nongeneric
callable signature. Compose with enclosing substitutions before testing that
every declaration binder is fixed. Use checker type identities, not spelling,
ABI similarity, a constraint, or observations of later calls. A represented
union or unknown is acceptable only if it is actually fixed by the context and
the instantiated body/relations and representation are supported. An unused
binder is still unresolved unless explicitly supplied or resolved by the checker
from a declared default. Do not invent a default when inference provides none.

Yield one monomorphized callable instance for that expression. Check the actual
instantiated body, parameters, results, predicates, mutable relations and nominal
bounds using the dependency's body-relation rules. Then apply ordinary callable
variance and ABI adaptation, including ignored extra callback arguments,
optional/default/rest parameters and discarded results. No `as` assertion or
`satisfies` syntax substitutes for these proofs. Conditional arms and object
fields may inherit a contextual signature; a generic container whose **element
is a concrete function type** is fine. An element with a generic call signature
is not. Ambiguous overload/context selection or unsupported concrete layouts
remain NotYet with the unresolved site and a fix.

A const alias of a generic function declaration retains its binders. Each direct
call through the alias is monomorphized independently, so the first example
prints `4 x`. Passing, storing or returning the alias is an escape: apply the
contextual specialization rule at that expression. Mutable aliases are outside
this exception.

Higher-rank slots are **Refused** in both extensions, naming every free slot
binder and a concrete-slot fix. Monomorphization has no instance to select; an
erased generic entry would trust a type it cannot prove.

If binders remain free, use **Refused**, with file:line:column, the flow path
into the slot, each unresolved binder, and an annotation/instantiation fix.
This is a ruled language restriction on otherwise sound polymorphic storage. Do not
silently specialize only the first later call or use one untyped native entry.

```a
function identity<T>(value: T): T { return value; }
const number: (value: number) => number = identity; // proposed accepted
const text = identity<string>;                    // proposed accepted
const generic = identity;                        // admitted const alias
console.log(`${generic(4)} ${generic('x')}`);       // Node: 4 x
```

```a
function identity<T>(value: T): T { return value; }
const box: { readonly run: <U>(value: U) => U } = { run: identity };
console.log(`${box.run(4)} ${box.run('x')}`); // Node: 4 x; refused at run
// Fix: concrete number/string slots assigned identity at their own value sites.
```

Both programs are sound TypeScript; the first is admitted by the alias exception.
Refusal of the second means unsupported polymorphic storage, not a type lie. Proposed diagnostic example:
`main.a:2:56: refused: identity -> box.run retains generic U; T has no single
instantiation; fix: use a concrete callable slot or identity<number>`.
The same specialization/refusal rule applies to `.ts` and `.a`. Existing checked
boundary policy remains: a `.ts` boundary eligible for V4 checking may insert
its checks; `.a` must meet its proven-type contract. Neither extension gets an
erased generic function fallback. Node running unchanged TypeScript accepts the
polymorphic examples; Adamic's two backends would refuse before emission.

## Identity and V4

Specialization selects code, not a new source function object. Repeated reads of
the same declaration/environment and same specialization compare equal, as V4
requires for adapters. Node also requires equality across **different**
specializations of the same source function. Imports, aliases, checked views,
Map/Set keys and unknown boxing must preserve that source identity. Distinct
function creations in distinct factory activations remain distinct; repeated
reads within one activation retain its identity and captured environment.

```a
function identity<T>(value: T): T { return value; }
const first: (value: number) => number = identity;
const again: (value: number) => number = identity;
const text: (value: string) => string = identity;
const left: unknown = first, right: unknown = text;
console.log(`${first === again} ${left === right}`); // Node: true true
```

Carry source identity separately from adapter entry/signature. Intern reusable
adapters without replacing closure-creation identity; ownership of captures and
cycle proofs still apply. If an identity observation cannot yet be represented,
stop with NotYet rather than compare specialization entry pointers.

V4's checked callable view checks a concrete parameter/result contract at the
call boundary, with its prescribed order, paths and failure behavior. It cannot
infer an arbitrary generic binder from `typeof function`, certify a generic body,
or turn the state-machine slot into a universally callable value. Instantiate
and prove first, then attach the checked view where V4 permits it; keep every
check on either backend and preserve identity through the view. Generic
predicate/assertion signatures require their separate body proofs, including
Debug's stack-marker use. An adapter cannot launder a trusted predicate or an
invariant mutable/nominal mismatch.

## First implementation after the ruling

Prioritize concrete comparer/default slots: eleven roots hit core.ts:220;
core.ts:814 and utilities.ts:10629 share its default shape. Use number and
runtime-built string callers, aliases/repeated reads, an outer generic default,
extra callback arguments, and different specializations. Hold each implemented
root reduction to source Node in native and JavaScript, with a valid wrong-result
or wrong-identity mutant caught by output, plus unresolved-binder and body-lie
negative mutants. Keep each new test leaf below 60 seconds. Remeasure exact
root/site rows on the dependency source snapshot; historical hidden bytes stay
historical. General polymorphic storage, identity-only markers, backward flow
inference and the binary state machine are outside that first implementation.

## Implemented and measured

The comparer/default reductions and immutable declaration aliases are implemented,
including alias readiness and concrete escapes. Higher-rank slots are refused in
both extensions with binder names and fixes; a monomorphic callable's argument
and result contracts are checked at their own annotations and escape sites,
without treating their recursive object graphs as that callable's own binders.
Source identity is preserved across represented specializations and callable views.

The [same-input census](../review/compiler/generic-values/census-comparison.md)
now observes this family at 11 roots and eight sites, from 26 roots and 14 sites.
The [complete admission proof](../review/compiler/generic-values/FOURTH.md)
checks all five newly accepted corpus programs against Node in both backends.
Explicit instantiation expressions, nested generic callable escapes and unknown
callable descriptors remain NotYet. Contexts without a single concrete signature
remain refused; the later parenthesizer overload site is recorded in the census.
