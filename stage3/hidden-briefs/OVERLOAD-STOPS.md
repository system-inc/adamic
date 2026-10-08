# Current overload region stops

Merged compiler baseline: `234467ff1cb968f3107746f7b17c368bc158ddba`, including
requested `compiler/area-next-fixtures` commit `f0c6e6fc`. Source pin `388096e6`;
all 82 adapted compiler file hashes are checked when computing the region counts.
The table names the first failed boundary in each assigned interval, including
checker body exclusion. A dependency or signature diagnostic need not lie inside
that interval. Such a location is explicitly identified, rather than invented
as an in-region lowering observation.

| Region | First boundary in the interval | Diagnostic location | Raising function | Owner | Exact diagnostic |
|---|---|---|---|---|---|
| 01 large, [68180,81805) | declarations.ts:1387:5 body excluded, [68264,90564) | declarations.ts:1678:110, beyond this interval | `latentFullSelected` / `LatentOwnDiagnosticsIn`; checker `isSignatureApplicable` | Checker, #k881crd | `error TS2345: Argument of type 'ExpressionWithTypeArguments \| undefined' is not assignable to parameter of type 'DeclarationDiagnosticProducing'. Type 'undefined' is not assignable to type 'DeclarationDiagnosticProducing'.` |
| 01 small, [82928,89827) | Same excluded body; first checker diagnostic inside this interval | declarations.ts:1678:110 | `latentFullSelected` / `LatentOwnDiagnosticsIn`; checker `isSignatureApplicable` | Checker, #k881crd | `error TS2345: Argument of type 'ExpressionWithTypeArguments \| undefined' is not assignable to parameter of type 'DeclarationDiagnosticProducing'. Type 'undefined' is not assignable to type 'DeclarationDiagnosticProducing'.` |
| 05 large, [35690,41768) | es2018.ts:828:9, [35690,35769) | visitorPublic.ts:123:5, dependency | `censusOverload` | Lowering | `Adamic 0.1 refuses overload 1 of visitNode parameter node cannot be served by implementation parameter node; make the implementation accept every value admitted by this overload, without mutable widening or bivariance` |
| 05 small, [144178,149926) | es2015.ts:3193:9, [144178,144257) | visitorPublic.ts:123:5, dependency | `censusOverload` | Lowering | `Adamic 0.1 refuses overload 1 of visitNode parameter node cannot be served by implementation parameter node; make the implementation accept every value admitted by this overload, without mutable widening or bivariance` |
| 06, [61324,72741) | partialTransformClassElement body entry, esDecorators.ts:1242:6 | esDecorators.ts:1239:9, parameter prologue before interval | `typeOf` | Lowering | `stage 0 can't lower a value of type TNode yet` |
| 13, [30237,37526) | es2017.ts:736:5 overload declaration | es2017.ts:736:5 | `censusOverload` | Lowering | `Adamic 0.1 refuses overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody; make the implementation result covariant with every overload result` |
| 14, [455532,462634) | evaluate body entry, utilities.ts:11322:81 | utilities.ts:11320:5, overload signature before interval | `censusOverload` | Lowering | `Adamic 0.1 refuses overload 1 of evaluate result EvaluatorResult<string \| undefined> cannot be served by implementation result EvaluatorResult<string \| number \| undefined>; make the implementation result covariant with every overload result` |

Exact multi-line checker text, without table whitespace normalization:

```text
transformers/declarations.ts:1678:110: error TS2345: Argument of type 'ExpressionWithTypeArguments | undefined' is not assignable to parameter of type 'DeclarationDiagnosticProducing'.
  Type 'undefined' is not assignable to type 'DeclarationDiagnosticProducing'.
```

The hidden-01 error comes from `clause.types[0]` at line 1678. Under
noUncheckedIndexedAccess the indexed value includes undefined, and strictNullChecks
rejects passing it to the required node parameter. Hidden-13's checker errors
similarly follow indexed reads of node.parameters[i] and outerParameters[i].
These are checker stops under the stricter options, #k881crd, rather than new
lowering result refusals.

There is no observed lowering boundary inside hidden-01's excluded body. Its
outer transformDeclarations also stops at declarations.ts:612:90 on
`an indirect value of an overload requiring a checked implementation boundary`;
that location is outside both assigned intervals. The two independent failures
must not be conflated.

Hidden-13 additionally has a checker-excluded implementation body. Its first
checker diagnostic is es2017.ts:764:38: `error TS18048: 'outerParameter' is possibly
'undefined'.` Further diagnostics occur at 765:25, 765:58, 767:76 and 770:44.
The checker raises these through `reportObjectPossiblyNullOrUndefinedError`,
called by `checkNonNullTypeWithReporter`; the census
excludes the body through `latentFullSelected` / `LatentOwnDiagnosticsIn`.
Fixing the overload signature alone would not prove this checker-excluded body.

Hidden-14's enclosing createEvaluator independently refuses object destructuring
at utilities.ts:11316:35 through `destructureFrom`: `Adamic 0.1 refuses a method in
object destructuring; call it on its receiver or wrap that call in an arrow;
destructuring would lose this`. This is outside the assigned interval.

The initial measurement on the merged compiler encountered the area's new private
IR emitter cache, argumentFacts. A measurement-only change now permits precisely
its nil state and still refuses a populated cache or any other private field.
The repeated baseline contains normal lowering and checker observations, not
those initial instrumentation panics. No checker option was weakened.

[Raw baseline](../../docs/overload-results/groups/structural/before.jsonl.gz)
contains all units, exact diagnostics, checker exclusions and boundary spans.

## After the structural result group

Compiler `8c7a0d8466268f2c1b61b235c701947d00fa13a3` admits a structural result
when the existing return proof proves every return under the overload's admitted
arguments. Readonly covariance permits widening; it never permits unchecked
narrowing of a wider result. Unproved results identify their failing component.

| Region | Revealed bytes | Next stop | Owner |
|---|---:|---|---|
| 01 large | 0 | declarations.ts:1387:5 body excluded by TS2345 at declarations.ts:1678:110, outside this interval | Checker |
| 01 small | 0 | declarations.ts:1678:110 TS2345 in that same excluded body | Checker |
| 05 large | 0 | es2018.ts:828:9 calls visitNode; visitorPublic.ts:123:5 parameter node refusal | Lowering, censusOverload |
| 05 small | 0 | es2015.ts:3193:9 calls visitNode; visitorPublic.ts:123:5 parameter node refusal | Lowering, censusOverload |
| 06 | 0 | esDecorators.ts:1242:6 body entry; esDecorators.ts:1239:9 TNode parameter representation | Lowering, typeOf |
| 13 | 0 | es2017.ts:736:5 Block result refusal at result.kind; implementation also checker-excluded | Lowering, censusOverload |
| 14 | 0 | utilities.ts:11322:81 body entry; utilities.ts:11320:5 EvaluatorResult refusal at result.value | Lowering, censusOverload |

Changed exact result diagnostics:

```text
transformers/es2017.ts:736:5: Adamic 0.1 refuses overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody at result.kind; prove every return for this overload's admitted arguments, preserving result variance
utilities.ts:11320:5: Adamic 0.1 refuses overload 1 of evaluate result EvaluatorResult<string | undefined> cannot be served by implementation result EvaluatorResult<string | number | undefined> at result.value; prove every return for this overload's admitted arguments, preserving result variance
```

The other exact diagnostics in the baseline table are unchanged. The complete
[comparison](../../docs/overload-results/groups/structural/regions.json) and
[next-stop records](../../docs/overload-results/groups/structural/next-stops.json)
retain compiler pin, source hashes, spans, raising functions and exact text.
All seven intervals remain fully hidden, totaling 58,158 bytes. The original full
compiler result bodies are still unproved; the accepted fixtures are the cases
whose admitted-argument return proof succeeds. No hidden-byte reveal is claimed.

## After field storage and constrained parameters

Compiler `dacbdb8cb2ce9033552aa5246a187b666b5cc395` reuses hidden-06's
constraint storage through the existing parameter typeOf path. Matching storage
is now required when reading a field of a union of overload result objects.
The full result.kind and result.value diagnostics remain admission refusals,
not property-read observations. Hidden-05 needs checked views and stays refused.

| Region | Hidden before | Hidden after | Revealed bytes | Next stop | Owner / raising function |
|---|---:|---:|---:|---|---|
| 01 large | 13,625 | 13,625 | 0 | declarations.ts:1387:5 excluded body; TS2345 at declarations.ts:1678:110 beyond interval | Checker, #k881crd; isSignatureApplicable |
| 01 small | 6,899 | 6,899 | 0 | Same excluded body; TS2345 at declarations.ts:1678:110 | Checker, #k881crd; isSignatureApplicable |
| 05 large | 6,078 | 6,078 | 0 | es2018.ts:828:9 calls visitNode; visitorPublic.ts:123:5 parameter node refusal | Lowering, censusOverload |
| 05 small | 5,748 | 5,748 | 0 | es2015.ts:3193:9; same visitNode parameter node refusal | Lowering, censusOverload |
| 06 | 11,417 | 3,369 | 8,048 | esDecorators.ts:1250:13 calls visitNodes; visitorPublic.ts:196:5 parameter visitor refusal | Lowering, censusOverload |
| 13 | 7,289 | 7,289 | 0 | es2017.ts:736:5 Block result refusal at result.kind; implementation also checker-excluded | Lowering, censusOverload |
| 14 | 7,102 | 7,102 | 0 | evaluate body entry; utilities.ts:11320:5 EvaluatorResult result refusal at result.value | Lowering, censusOverload |

The six unchanged exact diagnostics above still match the fresh census. The new
hidden-06 first boundary is [61665,61755), with this exact dependency diagnostic:

```text
visitorPublic.ts:196:5: Adamic 0.1 refuses overload 1 of visitNodes parameter visitor cannot be served by implementation parameter visitor; make the implementation accept every value admitted by this overload, without mutable widening or bivariance
```

[Comparison](../../docs/overload-results/groups/fields/regions.json),
[exact stop records](../../docs/overload-results/groups/fields/next-stops.json),
[raw census](../../docs/overload-results/groups/fields/after.jsonl.gz) and
[proof and validation report](../../docs/overload-results/groups/fields/REPORT.md).
All 82 adapted hashes remain pinned to 388096e6; total assigned hidden bytes
fall from 58,158 to 50,110. This is an intersection measurement on a
checker-rejected project, not a whole-compiler emission claim.
