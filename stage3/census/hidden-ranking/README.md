- Fixed the census copier: only ir.Program.argumentFacts is reset to nil on snapshots.
- Main d72728e5: **3,981,038 / 10,618,056 hidden bytes (37.493097%)**.
- Delta from 69501280: **+326,158 bytes**, +3.064438 percentage points.
- Panic-hidden bytes: **3,647**; copier panic records: **0**.
- Real rollback/fresh-facts fixture and allowlist-removal mutant pass; no native-compilation or closure-only gain claim.

Compiler pin: `d72728e570fe09d91cf55b37b564dbad0fefc24d`. Measurement-only fix is in `stage3/census/latent/full.go.txt`.
No production compiler, cache implementation or checker code was changed.
[RESULT.json](RESULT.json) includes every reason and boundary; [DELTA.json](DELTA.json)
includes every file and reason's before/after values. Current hidden ranges and the
ten largest regions are in [hidden-source](../hidden/README.md).

Main's newer adaptations change source bytes. A second census uses exactly all 82
frozen input hashes from the 69501280 run, to expose compiler-only changes without
confounding adaptation edits. Even that delta includes all intervening compiler
features, so it does not isolate closures alone.

| Metric | 69501280 | d72728e5 main input | d72728e5 frozen control |
|---|---:|---:|---:|
| Source bytes | 10,615,807 | 10,618,056 | 10,615,807 |
| Hidden bytes | 3,654,880 | 3,981,038 | 3,970,749 |
| Hidden share | 34.428659% | 37.493097% | 37.404118% |
| Hidden-byte delta | 0 | +326,158 | +315,869 |
| Panic-hidden bytes (exclusive outermost cause) | 3,647 | 3,647 | 3,647 |
| Copier panic records | not remeasured | 0 | 0 |

The named allowlist contains only `ir.Program.argumentFacts`. The copied Program
starts with that pointer nil; the next facts reader recomputes it. Every other
unexported field still panics, including an unrelated structure with a field also
named argumentFacts. The old broken meter's approximately 10,550 cache panics are
not counted as compiler refusals or lowerable source; the repaired runs above
replace that invalid measurement.

`panic_hidden_bytes` uses the same outermost-cause partition as compiler reasons.
Every panicked boundary participates in the hidden union; independent examination
is still subtracted. The union of residual bytes under **any** panic boundary is
6,657;
3,010 of those bytes are also enclosed by another outer cause.
Those shared bytes remain in the total, once, and do not get duplicate panic credit.
A separate byte-mask audit checks the panic union. Removing the synthetic panic
span is caught by its known 30-byte answer. Copier panic-hidden bytes are zero.

**Remaining census limitation:** d72728e5 still declares `typeMapper struct{}`.
Its generated `latentCopyPointer_typeMapper` allocates a new empty mapper; it
does not preserve the checker-owned identity as the 69501280 fix did. The
argumentFacts exception repairs snapshot-cache failures, but does not repair
this other known source of invented-any labels and checker-context pollution.
Counts are therefore recorded census observations, not validated source-any
attribution or proof of closure gains. Both pins' differences remain visible.

The top 30 Refused/NotYet reasons below exclude checker and panic buckets from the
compiler-reason list, but include their bytes in the total. Attribution is potential
census exposure, not a measured fix-alone counterfactual. Owners use the same pinned
refusal/NotYet tables as before; unavailable owners remain null. Whole-body boundary
examples locate the selection, with exact diagnostic locations retained in JSON.

| Rank | Kind and exact reason | Hidden bytes | Boundaries | Up to three file:line examples | Owner |
|---:|---|---:|---:|---|---|
| 1 | NotYet: a value of type __String | 273,443 | 376 | factory/nodeFactory.ts:492<br>binder.ts:509<br>transformers/typeSerializer.ts:450 | unavailable |
| 2 | NotYet: a function returning CapturedThis | 132,395 | 2 | transformers/es2015.ts:488 | unavailable |
| 3 | NotYet: a value of type InitializedVariableDeclaration | 120,003 | 5 | transformers/module/module.ts:174<br>transformers/generators.ts:336 | unavailable |
| 4 | NotYet: a computed field name | 113,387 | 4 | visitorPublic.ts:623<br>parser.ts:505<br>scanner.ts:135 | unavailable |
| 5 | NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class | 101,409 | 2000 | factory/emitHelpers.ts:148<br>parser.ts:10625<br>transformers/module/system.ts:536 | unavailable |
| 6 | NotYet: a value of type PrivateIdentifierInExpression | 97,864 | 3 | transformers/classFields.ts:354 | unavailable |
| 7 | NotYet: a function returning ImmediatelyInvokedArrowFunction | 93,640 | 2 | transformers/esDecorators.ts:295 | unavailable |
| 8 | NotYet: a function returning any | 89,754 | 2056 | utilities.ts:1384<br>commandLineParser.ts:2478<br>moduleSpecifiers.ts:897 | unavailable |
| 9 | Refused: overload 1 of writeTokenText result void cannot be served by implementation result number | 79,336 | 10 | emitter.ts:1211 | unavailable |
| 10 | Refused: overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem | 63,225 | 4 | transformers/declarations.ts:259 | unavailable |
| 11 | NotYet: a value of type ParameterPropertyDeclaration | 61,355 | 3 | transformers/ts.ts:235 | unavailable |
| 12 | NotYet: a value of type Path | 60,277 | 148 | resolutionCache.ts:579<br>watchUtilities.ts:572<br>resolutionCache.ts:348 | unavailable |
| 13 | Refused: overload 1 of transformFunctionBody result Block cannot be served by implementation result ConciseBody | 42,328 | 4 | transformers/es2018.ts:154 | unavailable |
| 14 | NotYet: a value of type PrimitiveLiteral | 38,462 | 3 | expressionToTypeNode.ts:174 | unavailable |
| 15 | Refused: a method in object destructuring | 35,437 | 12 | sys.ts:991<br>sys.ts:593<br>utilities.ts:11319 | unavailable |
| 16 | NotYet: a function returning T &#124; undefined | 32,292 | 159 | moduleNameResolver.ts:1166<br>moduleSpecifiers.ts:713<br>utilities.ts:2923 | unavailable |
| 17 | NotYet: a value of type any | 30,670 | 85 | program.ts:4891<br>commandLineParser.ts:3059<br>moduleNameResolver.ts:1803 | unavailable |
| 18 | Refused: overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody | 28,857 | 3 | transformers/es2017.ts:119 | unavailable |
| 19 | NotYet: a destructured name that isn't plain | 27,051 | 3 | utilities.ts:11494 | unavailable |
| 20 | Refused: a cast the runtime can't check | 25,433 | 926 | sys.ts:419<br>moduleSpecifiers.ts:1123<br>utilities.ts:3435 | internal/lower/cast_proof.go / lowering.castProof |
| 21 | NotYet: a value of type ResolvedConfigFileName | 20,045 | 13 | tsbuildPublic.ts:932<br>tsbuildPublic.ts:1890<br>tsbuildPublic.ts:608 | unavailable |
| 22 | NotYet: a function returning T | 18,147 | 106 | transformer.ts:248<br>utilities.ts:12348<br>executeCommandLine.ts:983 | unavailable |
| 23 | Refused: var | 15,286 | 65 | parser.ts:1568<br>scanner.ts:1061<br>parser.ts:1600 | unavailable |
| 24 | Refused: overload 1 of parenthesizeConciseBodyOfArrowFunction result Expression cannot be served by implementation result ConciseBody | 14,306 | 4 | factory/parenthesizerRules.ts:54 | unavailable |
| 25 | NotYet: a value of type readonly T[] &#124; undefined | 13,990 | 329 | factory/utilities.ts:718<br>transformers/module/system.ts:439<br>transformers/destructuring.ts:466 | unavailable |
| 26 | NotYet: for...of over an object | 13,952 | 120 | transformers/utilities.ts:194<br>moduleNameResolver.ts:2716<br>binder.ts:432 | unavailable |
| 27 | Refused: overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram &#124; undefined | 12,172 | 5 | builder.ts:1673<br>builderPublic.ts:223<br>builderPublic.ts:184 | unavailable |
| 28 | NotYet: an array of never | 12,066 | 84 | utilities.ts:9044<br>commandLineParser.ts:3335<br>utilities.ts:9841 | unavailable |
| 29 | NotYet: a function returning Path | 11,548 | 18 | watchUtilities.ts:118<br>program.ts:1514<br>program.ts:1499 | unavailable |
| 30 | NotYet: a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } | 10,567 | 2 | programDiagnostics.ts:89 | unavailable |

Other causes: SkippedDependency: 87,747 bytes, checker: 1,740,798 bytes, conflict: 422 bytes, panic: 3,647 bytes.

Validation and reproduction

Toolchain: Node 24.19.0, Go 1.27.1, clang 20.1.8; `nproc` = 5 (CPU quota 4).
`GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh` logged timing
lines Node 0.016s, Go 0.018s, submodules 0.043s, markdown 0.047s and clang
0.101s, then stopped during dependency export with the pre-existing full `/tmp`.
Sourcing `/workspace/adamic-tools/env.sh` and setting `TMPDIR` to workspace
scratch allowed the census build and targeted tests to finish.

Ran `bash stage3/apply.sh <main-adapted>`, `make_overlay.py`, `gofmt` on the
generated overlay and `go build -buildvcs=false -overlay=<overlay.json>
./stage3/census/latent/tool`. Both full census runs used `LATENT_FULL=1
LATENT_ASSERT_NO_OUTPUT=1 <census> <src/compiler> <full.jsonl>`. Main's stock
ledger used `NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node
units.cjs <src/compiler> <stock.json>`; the frozen run used the unchanged stock
ledger from ec0b16c0. Retrieve hidden.py, audit.py, prove.py, units.cjs and
ranking.py from ec0b16c0; their arithmetic and owner tables are unchanged.
Union failed/selected/skipped spans per file, subtract independently examined
regions, then partition residual intervals under their outermost causes.
Panic union is the intersection of residual hidden ranges with the union of
all panic boundary spans, checked independently with byte masks.

`cache_copy_audit.py` ran only
`go test ./internal/lower -run '^TestLatent(CacheRollbackFacts|UnknownPrivateField)$'`
with the guarded overlay, `-buildvcs=false -count=1 -v`. The actual failed
initializer rolled back, subsequent lowering succeeded, and packed facts matched
a fresh program. Removing the allowlist entry produced the argumentFacts panic;
skipping all private fields failed the unknown-private-field assertion.

Passed 8 arithmetic tests, 4 ranking tests (including the nested-byte mutant),
the 141-byte real census integration fixture, both full-ledger byte-mask audits
and all 9 arithmetic/report mutants. The synthetic duplicate/nested panic
fixture gave exactly 30 hidden bytes; dropping panic spans was caught.
`latent/audit.py` passed with its scope, injected-refusal and attribution mutants.
The condition-aware full audit from f1502d13 passed, including first-error-only
and failed-state-retention mutants; main's obsolete condition-refusal expectations
were not used or edited. All test output went to logs in
[evidence/d72728e5](evidence/d72728e5). No whole-package confirmation, full gate,
native oracle, lane or emitted-JavaScript comparison was run: this unit changes
the census copier and records measurements, not production compilation.
