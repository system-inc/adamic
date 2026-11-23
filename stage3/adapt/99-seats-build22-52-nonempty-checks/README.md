# Family 52 one-argument `some` checks

Nov 22, 2026, build22-52, #bbssdcy. Adamic ruled this JavaScript-changing adaptation as sketched. Compiler accepts under the widened split. The adapter checks the exact stock TypeScript 6.0.3 overload and local declarations by symbol before editing.

| Site in TypeScript 6.0.3 | Local declaration | Inline check |
| --- | --- | --- |
| `transformers/destructuring.ts:171:14` | `let expressions: Expression[] \| undefined;` | `expressions === undefined \|\| expressions.length === 0` |
| `visitorPublic.ts:526:9` | `const declarations = context.endLexicalEnvironment();` | `declarations !== undefined && declarations.length > 0` |
| `visitorPublic.ts:548:9` | `const declarations = context.endBlockScope();` | `declarations !== undefined && declarations.length > 0` |

At all three sites stock TypeScript resolves the call to this bodiless overload.

```ts
export function some<T>(array: readonly T[] | undefined): array is readonly T[];
```

Row 150 requires a true-direction check on that signature's unproven claim. The edit removes the call and its signature claim, so row 150 has no claim left to check and row 132 has no overload result left to check. The branches narrow by the checker's own definedness flow. The named bindings are plain locals. The length operand reads the same binding again without evaluating its initializer or running code to look up the binding.

- At `transformers/destructuring.ts:171:14`, the guard returns for undefined or zero length. After the guard the checker knows `expressions !== undefined`, and the existing push runs only for a nonempty array.
- At `visitorPublic.ts:526:9`, `endLexicalEnvironment()` runs once at the existing declaration. The branch sees the local as defined and runs only for a nonempty array.
- At `visitorPublic.ts:548:9`, `endBlockScope()` runs once at the existing declaration. The branch sees the local as defined and runs only for a nonempty array.

The existing implementation in the fully adapted base's `core.ts` follows its `else` when the one-argument call supplies no predicate. The earlier type-only assertion on `array[i]!` is in the two-argument path.

```ts
export function some<T>(array: readonly T[] | undefined, predicate?: (value: T) => boolean): boolean {
    if (array !== undefined) {
        if (predicate !== undefined) {
            for (let i = 0; i < array.length; i++) {
                if (predicate(array[i]!)) {
                    return true;
                }
            }
        }
        else {
            return array.length > 0;
        }
    }
    return false;
}
```

| Argument | `some` and the visitor checks | `!some` and the destructuring guard |
| --- | --- | --- |
| Undefined | false | true |
| Empty array | false | true |
| Nonempty array | true | false |

Rows 24 and 32 govern the array and brand checks the original shape runs into. The inline tests read the existing array length and never inspect a brand or infer storage from a predicate target. Row 133 covers unread overload results, while these original calls supplied their results to a branch. Rows 144 and 150 govern the predicate check and its directions. Row 153 governs reads through a branded call's claim. The inline branches use the local's existing array type after the definedness test.

## Regression test and the next stops

Run `node stage3/adapt/99-seats-build22-52-nonempty-checks/test.cjs <pristine TypeScript v6.0.3 checkout>` with the locked `stage3/api` packages installed. The test runs the real adapter in a scratch tree and compares the emitted visitor bodies with stock. It checks return identity, declaration array ownership and mutation, and callback order for undefined, empty and nonempty arrays in both visitors.

Both round 1 census passes recorded these first stops after clearing family 52:

- `transformers/destructuring.ts:185:9`, unit `flattenDestructuringAssignment` at `:99:1`, family 53. The census says “a checked predicate target Expression that narrows ArrayBindingPattern, which isn't its supertype.” Other type names appear in the meter under the same template from `internal/lower/predicates_branded.go:364`. Seats reads this as an exposed stop. Compiler confirms the exact-wording accounting at accept.
- `visitorPublic.ts:123:5`, the existing overload-parameter refusal family. The census says “overload 1 of visitNode parameter node cannot be served by implementation parameter node.” The seven `visitFunctionBody` units are at `:508:1`, `:513:1`, `:518:1`, `:519:18`, `:520:18`, `:521:18` and `:522:1`. The three `visitIterationBody` units are at `:540:1`, `:542:1` and `:543:1`.

## Rulings at `0efbdb083a60`

### Row 24

> **Array membership in checked views waits for views v3; inferring physical storage from the asserted target is unsound; fixed object-backed tuples are admitted; an unused optional view is admitted lazily and reading its optional alias is refused.** NotYet and Refuse. Undated, compiler ruling (unattributed); source: docs/views-v2-dispatch-review.md:5-7,23,43-45. Spec: not in the spec. Test: internal/oracle/checked_views_v2_source_test.go:199 TestCheckedViewUntaggedArrayPending.

### Row 32

> **A required array-brand member typed `undefined` is a phantom brand only if every observation of its presence (`in`, `hasOwn`, keys, spread, assign, `JSON.stringify`, writes) is refused. The same holds for an object brand, an extension of an object type (`interface TaggedNode extends Node { __tagged: undefined }`) or a branded supertype: every observation of its phantom member's presence is refused (`in`, `Object.hasOwn`, `hasOwnProperty` and `propertyIsEnumerable` by `Object.hasOwn`'s key rule, `Object.keys`, `values`, `entries` and `getOwnPropertyNames`, an object spread, `Object.assign`, `JSON.stringify`, a property or element write), and so is a read of the member (a dot, a binding or shorthand, an assignment pattern at any depth), which stopped at run time where Node reads `undefined`. A value whose construction stores the member (a literal that writes it, a class field) and an optional member are no phantom, and observing them lowers.** Refuse. Oct 7 2026, @system_adamic; source: #bw3xg7c (rsjzeans). Object brands Nov 12 2026, @system_adamic, under the two-hour rule, as this entry applied to a brand it didn't yet cover, not a new ruling; source: #z5d0jye, Language's q06 at main 0c705275. An `in` test on the phantom member answered false, as Node does, but tsc narrowed the false branch to `never` and lowering trusted it, so a `never` value seen as `{ kind: string }` read the number 1.0 as a string pointer (an ASan `SEGV` natively), through step 1's narrow overload and through a checked predicate alike. Built as #3hsrk2t (e42d56d2; the read road and hasOwnProperty's key rule 544df32e; nested assignment patterns 4bba06cb). Every other road to a lying `never` is :74's to stop, owed as #p6vawef. Spec: not in the spec. Test: arrays, internal/lower/unbranded_test.go:150 TestUnbrandedPresenceIsRefused, :186 TestUnbrandedPresenceControls and :268 TestUnbrandedPresenceHeldElsewhere (#gb0fqaa); objects, internal/lower/phantom_object_presence_test.go:14 TestPhantomObjectPresenceLanguage (q01 to q07 and p02, p11, q06 the `SEGV`), :80 TestPhantomObjectPresenceObservations (every road, .a and .ts), :109 TestPhantomObjectPresenceControls (a real field, an own method, a real class lowering), :125 TestPhantomObjectPresenceEdges, :185 TestPhantomObjectReadsAreRefused, :203 TestPhantomObjectPrototypeKeys, :220 TestPhantomObjectAgreeingReads and :252 TestPhantomObjectNestedAssignmentRead; mutants in internal/lower/unbranded.mutants.md (Object presence, Object reads and keyed prototype methods, Nested object assignment reads).

### Row 132

> **A call resolving to an overload whose declared result is narrower than the implementation's gets a check on exactly the failing field at that call site in .ts (exit 70, counted), erased where the body proves it; in .a it's refused until a per-overload proof exists; an overloaded function used as a value keeps its whole set, and the check follows the promise, not the path.** Check in .ts, Refuse in .a. Oct 7 to 8 2026, @system_adamic; source: #bw3xg7c (anrxnm5x; e1bwzhwk; 7njxf2yb), #wvhe1jz (xaeaekqg, vbx223yc). Spec: not in the spec. Test: internal/oracle/census_overload_result_test.go:40 TestCensusOverloadResultStop.

### Row 133

> **A narrow overload's result that no read reaches is erased and counted, never checked, amending the narrow overload result row above ("A call resolving to an overload whose declared result is narrower") to read as the branded-call row ("A call claiming a type that holds a branded S") and the checked predicate row ("In .ts an unproven user-defined predicate is a checked predicate") do: a call statement, the left side of a comma, a `void` operand and an arrow body whose value nobody reads. A discarded result is released in the implementation's representation, never the overload's, since the value comes back as the implementation's type.** Check in .ts. Nov 11 2026, @system_adamic; source: #z5d0jye, Language's consistency point on Optimizer's call-site check (10b6c33d, which checked a discarded result) against this morning's :138. Oct 19's principle: a lie no read sees changes nothing a program can observe. The guard that every narrow call is wrapped or erased sits where every road meets it, the call statement's included (Language's should, Lower's must). Ruled and not built. Spec: not in the spec. Test: owed by #z5d0jye's step 1 (a discarded narrow call counted and not checked, with its mutant; a discarded call whose overload promises a reference while the implementation returns a boxed union holding undefined, released as the implementation's type under ASan, with the mutant releasing it as the overload's type going red).

### Row 144

> **In .ts an unproven user-defined predicate is a checked predicate: the true branch checked, the false branch checked where it narrows, `asserts` checked on return, exit 70, counted.** Check in .ts; .a unchanged. Oct 8 2026, @system_adamic; source: #tzd3gjg (8t75jj6h), #bw3xg7c (gt8a54dx). Spec: What's refused in 0.1 (row user-defined type guards, the `.ts` split). Test: internal/oracle/predicates_checked_test.go:64 TestCheckedPredicateOracle (each fixture under internal/lower/testdata/predicates_checked in .ts checked mode, held to Node natively, sanitized and on the JavaScript backend, with a mutant per branch); internal/lower/predicates_checked_test.go:19 TestCheckedPredicateBranches; internal/oracle/predicates_checked_test.go:187 TestCheckedPredicateCountsAreRecorded; internal/conformance/testdata/refusals/type_guard.ts (lowers in .ts; type_guard.a stays refused). Landed October 25 as 0ee00f60 (#7v4rjr7), re-cut from 740adf19, which was built on October 9 and had never reached main.

### Row 150

> **In .ts a predicate whose claim isn't proven is a checked predicate in every declaration shape, not only a function declaration: methods, arrows, function expressions, and bodiless signatures (function types, interface method signatures, overload declarations). Each call's true direction checks the argument against the claim of the signature the call goes through. A bodiless signature carries no proof of its own, so a proven function called through an unproven signature is still checked. An arrow or function expression used as a value carries its check in its closure, through #vdt9f1q's forwarder. The false direction stays the walk's (#1761dah). A call through a method or interface signature may drop its check only when every body that call can reach in the closed world proves the claim: the base, every override, every class or object literal that implements the interface method, computed once per program before any check is dropped. If the set can't be closed, the check stays, and a body that proves its own answer never erases a check at a call that can reach another body. In .a an unproven claim stays refused.** Check in .ts; Refuse in .a. Oct 27 2026, @system_adamic; source: #1w5bve2 (d4qnh82k, and 7ztj70ar on Language's 46180crd, with Optimizer's dispatch set dmy1n320). Spec: not in the spec. Test: NONE (not built; after #x2gt39p and #3hxh5jm; fixtures per shape, each lying once, plus `class Liar extends V` overriding a proven `V.isS` and called through `V`, which must stop at the check; a mutant that skips one shape's check, and one that erases on the base's proof alone).

### Row 153

> **A call claiming a type that holds a branded S stands when either no narrowed read exists or every read is safe. No read exists (rule 1) when the call is an operand of a `||` or `&&` chain whose whole value is the returned expression of a function with a checked result claim, every other use of the argument's binding there is an argument to another checked-claim call in the same chain at a parameter typed as the generic P, and nothing else reads, stores, captures or returns it, walked into nested functions and classes; the `&&` form's next call takes the narrowed argument as an upcast to its own parameter, never a field read through the claim. Otherwise each member of the true-side read's type must be one of: (a) S itself, or S intersected only with S's supertypes, read through S's view (change B); (b) a member the argument's declared type already held when the narrowing dropped nothing, by the checker's own type identity at the narrowed reference, never a structural comparison; (c) a member from an unbranded arm with its own literal tag, which the call's check tests, unique by value among the arms, by the same helper the predicate verifier uses for duplicate enum aliases (#fvmqee2). Anything else stays NotYet in its words. A producer, a direct call of a program function returning an arrow (core.ts's orPredicate2 and orPredicate3), is accepted at the trusted plain-callback sites (findAncestor's and lib's find, seven checker.ts sites) because its returned arrow is a checked closure under the producer's own written claim, instantiated per instance: exactly what `(p): p is R1 | R2 => f1(p) || f2(p)` written by hand would be, checking its claim at each true return, and accepted as a checked closure passed directly is, with the same false-side treatment; nothing new is proven. A branded instance checks what a branded checked closure checks under (a), no more. Inside a producer, a call through a parameter keeps its check unconditionally, and R-T80's erasure never treats it as closed by collecting the predicates passed at call sites. All seven tsc sites read only the true side (Optimizer's run 2); checker.ts:10911 also meets the `?.` call #vx8qdwg builds.** Check in .ts; ruled and not built. Nov 7 2026, @system_adamic; source: #fx9waqy (687asbyp, on Optimizer's ak1g9xx8 with Compiler's restatement of rule 1 and Language's conditions: the walk into nested functions, type identity, the tag unique by value, and the producer's two; x31x54pk, road (B), replacing that ruling's flow proof of the producer, after Optimizer measured that the flow proof trusts no calls). Spec: not in the spec. Test: owed by Optimizer's head (change B's next, on Lower's #fx9waqy): NotYet pins `p => f1(p) || (p as R2).x !== undefined` and `p => f1(p) && (() => p.x)()`; an assignment-narrowed declaration and a union written in another order, each NotYet unless the ids match; X | Y over `K { First = 1, Alias = 1 }` stopping at the call; a lying arrow carried in a variable and passed as f1, stopping at the check on both backends; the same helper at an unbranded and a branded instance in one program, each judged on its own; the arrow's own check stopping a lying f2 at the trusted site on both backends, and a branded instance whose result findAncestor hands to a caller that reads a field, stopping at the view's check; and five mutants (rule 1 without the use walk, (b) admitting a dropped member, (c) without the tag, erasure by call-site collection, and the arrow's return check dropped, which reads the lie raw through findAncestor's result).
