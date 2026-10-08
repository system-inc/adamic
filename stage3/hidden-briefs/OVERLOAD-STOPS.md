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

There is no observed lowering boundary inside hidden-01's excluded body. Its
outer transformDeclarations also stops at declarations.ts:612:90 on
`an indirect value of an overload requiring a checked implementation boundary`;
that location is outside both assigned intervals. The two independent failures
must not be conflated.

Hidden-13 additionally has a checker-excluded implementation body. Its first
checker diagnostic is es2017.ts:764:38: `error TS18048: 'outerParameter' is possibly
'undefined'.` Further diagnostics occur at 765:25, 765:58, 767:76 and 770:44.
The checker raises these through `checkNonNullTypeWithReporter`; the census
excludes the body through `latentFullSelected` / `LatentOwnDiagnosticsIn`.
Fixing the overload signature alone would not prove this checker-excluded body.

Hidden-14's enclosing createEvaluator independently refuses object destructuring
at utilities.ts:11316:35 through `objectBindings`: `Adamic 0.1 refuses a method in
object destructuring; call it on its receiver or wrap that call in an arrow;
destructuring would lose this`. This is outside the assigned interval.

The initial measurement on the merged compiler encountered the area's new private
IR emitter cache, argumentFacts. A measurement-only change now permits precisely
its nil state and still refuses a populated cache or any other private field.
The repeated baseline contains normal lowering and checker observations, not
those initial instrumentation panics. No checker option was weakened.

[Raw baseline](../../docs/overload-results/groups/structural/before.jsonl.gz)
contains all units, exact diagnostics, checker exclusions and boundary spans.
