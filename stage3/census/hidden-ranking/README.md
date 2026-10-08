- Built an exact-reason hidden-byte ranking on the frozen ed6e2975 census.
- Partitioned 3,970,749 hidden bytes exactly once under their outermost stopping causes.
- Refusal/NotYet reasons receive 2,148,422 bytes; other causes receive 1,822,327 bytes.
- Four synthetic accounting tests pass; the nested-byte double-count mutant is caught.
- Owners use the pinned tables where supplied; NotYet ownership and actual fix-alone outcomes remain unavailable.

[RESULT.json](RESULT.json) contains all 1,708 refusal/NotYet reasons, including
zero-credit reasons, all other causes, every raw boundary/skip record and every
attributed half-open byte segment. This unit reuses the complete
[hidden-source measurement](../hidden/README.md), compiler
`ed6e29751ee47d86fad450cd1674139883bc0f70`. It does not remeasure a newer compiler.
The topic branch starts at hidden-source `f1502d130bc0b4440f29b93b76b7d18bff3f6a60`;
no compiler feature branches were merged.

Every one of the 17,643 Boundary records is matched by exact diagnostic text to
its recorded Refused, NotYet, panic/error or SkippedDependency cause. Panic/error
records sharing the same panic text are one panic cause. The 147 checker-skipped
body records and 16 dependency body records are included as checker causes,
matching the original hidden-source union. Frozen per-file hidden ranges are
recomputed from the census and stock catalogue before attribution.

For each residual hidden byte, a strictly enclosing boundary takes precedence
over nested boundaries, including independently attempted nested units. Equal
spans with the same cause merge. Equal spans with differing causes, or crossing
outer spans with differing causes, receive a conflict bucket and no individual
reason credit. Checker-rejected enclosing bodies win over nested lowerer reasons.
Successfully examined bytes have already been removed and receive no credit.
All reasons retain their exact diagnostic text; there is no family normalization.

`bytes_revealed_if_fixed_alone` is the requested outermost-cause attribution:
bytes potentially exposed to the census after removing that outer reason. It is
an estimate, not a measured compiler counterfactual. Inner blockers, other checks
in the same construct and the checker can still stop a rerun. The estimate does
not claim all credited bytes would then lower successfully. Conflicting causes
are retained separately because the ledger cannot choose one cause for them.

`boundary_count` counts distinct (file, start, end) spans carrying that exact
reason, including fully examined or shadowed spans. `raw_boundary_records`
retains repetitions across attempts; `contributing_boundary_count` counts spans
that actually receive outermost credit. Examples are up to three distinct
boundary file:lines, preferring the largest credited boundaries. Fewer than
three means fewer contributing file:lines exist. Whole-body examples name the
boundary selection; its actual failure diagnostic and unit are retained in JSON.

Owner sources are frozen at refusal table
`d35a81d36fdafccf827bad0f572d311b2a0d4deb` and NotYet table
`dc6b1529ae9d2a2210672e105c8bb6374619a59d`. The refusal final and baseline tables
provide production file/function owners for exact matches. The supplied NotYet
table has disposition/why fields but **no owner column**; its owners are null.
Reasons absent from the supplied tables also retain null owners. No owners are
guessed from related templates. The frozen inputs and their hashes are in
evidence/ and RESULT.json. 17 exact reasons have supplied owners.

The top 30 refusal/NotYet reasons follow, sorted by descending credited bytes.
Other causes are excluded from this compiler-reason list and accounted below.

| Rank | Kind and exact reason | Bytes revealed if fixed alone | Boundaries | Three file:line examples (when available) | Owner from supplied tables |
|---:|---|---:|---:|---|---|
| 1 | NotYet: a value of type __String | 273,934 | 377 | factory/nodeFactory.ts:492<br>binder.ts:509<br>transformers/typeSerializer.ts:450 | unavailable: NotYet table has no owners |
| 2 | NotYet: a function returning CapturedThis | 132,345 | 2 | transformers/es2015.ts:488 | unavailable: NotYet table has no owners |
| 3 | NotYet: a value of type InitializedVariableDeclaration | 120,003 | 5 | transformers/module/module.ts:174<br>transformers/generators.ts:336 | unavailable: NotYet table has no owners |
| 4 | NotYet: a computed field name | 113,387 | 4 | visitorPublic.ts:623<br>parser.ts:505<br>scanner.ts:135 | unavailable: NotYet table has no owners |
| 5 | NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class | 101,409 | 1985 | factory/emitHelpers.ts:148<br>parser.ts:10625<br>transformers/module/system.ts:536 | unavailable: NotYet table has no owners |
| 6 | NotYet: a value of type PrivateIdentifierInExpression | 97,864 | 3 | transformers/classFields.ts:354 | unavailable: NotYet table has no owners |
| 7 | NotYet: a function returning ImmediatelyInvokedArrowFunction | 93,640 | 2 | transformers/esDecorators.ts:295 | unavailable: NotYet table has no owners |
| 8 | NotYet: a function returning any | 89,627 | 2060 | utilities.ts:1384<br>commandLineParser.ts:2478<br>moduleSpecifiers.ts:897 | unavailable: NotYet table has no owners |
| 9 | Refused: overload 1 of writeTokenText result void cannot be served by implementation result number | 79,336 | 10 | emitter.ts:1211 | unavailable: exact reason absent |
| 10 | Refused: overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem | 63,225 | 4 | transformers/declarations.ts:259 | unavailable: exact reason absent |
| 11 | NotYet: a value of type ParameterPropertyDeclaration | 61,355 | 3 | transformers/ts.ts:235 | unavailable: NotYet table has no owners |
| 12 | NotYet: a value of type Path | 60,277 | 148 | resolutionCache.ts:579<br>watchUtilities.ts:572<br>resolutionCache.ts:348 | unavailable: NotYet table has no owners |
| 13 | Refused: overload 1 of transformFunctionBody result Block cannot be served by implementation result ConciseBody | 42,328 | 4 | transformers/es2018.ts:154 | unavailable: exact reason absent |
| 14 | NotYet: a value of type PrimitiveLiteral | 38,462 | 3 | expressionToTypeNode.ts:174 | unavailable: NotYet table has no owners |
| 15 | Refused: a method in object destructuring | 35,437 | 12 | sys.ts:991<br>sys.ts:593<br>utilities.ts:11316 | unavailable: exact reason absent |
| 16 | NotYet: a function returning T &#124; undefined | 32,292 | 159 | moduleNameResolver.ts:1166<br>moduleSpecifiers.ts:713<br>utilities.ts:2923 | unavailable: NotYet table has no owners |
| 17 | NotYet: a value of type any | 30,827 | 89 | program.ts:4891<br>commandLineParser.ts:3059<br>moduleNameResolver.ts:1803 | unavailable: NotYet table has no owners |
| 18 | Refused: overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody | 28,857 | 3 | transformers/es2017.ts:119 | unavailable: exact reason absent |
| 19 | NotYet: a destructured name that isn't plain | 27,051 | 3 | utilities.ts:11491 | unavailable: NotYet table has no owners |
| 20 | Refused: a cast the runtime can't check | 25,281 | 936 | sys.ts:419<br>moduleSpecifiers.ts:1123<br>utilities.ts:3435 | internal/lower/cast_proof.go / lowering.castProof |
| 21 | NotYet: a value of type ResolvedConfigFileName | 20,029 | 13 | tsbuildPublic.ts:932<br>tsbuildPublic.ts:1890<br>tsbuildPublic.ts:608 | unavailable: NotYet table has no owners |
| 22 | NotYet: a function returning T | 18,147 | 106 | transformer.ts:248<br>utilities.ts:12345<br>executeCommandLine.ts:983 | unavailable: NotYet table has no owners |
| 23 | Refused: var | 15,286 | 65 | parser.ts:1568<br>scanner.ts:1061<br>parser.ts:1600 | unavailable: exact reason absent |
| 24 | Refused: overload 1 of parenthesizeConciseBodyOfArrowFunction result Expression cannot be served by implementation result ConciseBody | 14,306 | 4 | factory/parenthesizerRules.ts:54 | unavailable: exact reason absent |
| 25 | NotYet: a value of type readonly T[] &#124; undefined | 13,990 | 329 | factory/utilities.ts:718<br>transformers/module/system.ts:439<br>transformers/destructuring.ts:466 | unavailable: NotYet table has no owners |
| 26 | NotYet: for...of over an object | 13,952 | 120 | transformers/utilities.ts:194<br>moduleNameResolver.ts:2716<br>binder.ts:432 | unavailable: NotYet table has no owners |
| 27 | Refused: overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram &#124; undefined | 12,172 | 5 | builder.ts:1673<br>builderPublic.ts:223<br>builderPublic.ts:184 | unavailable: exact reason absent |
| 28 | NotYet: an array of never | 12,119 | 94 | utilities.ts:9041<br>commandLineParser.ts:3335<br>utilities.ts:9838 | unavailable: NotYet table has no owners |
| 29 | NotYet: a function returning Path | 11,548 | 18 | watchUtilities.ts:118<br>program.ts:1514<br>program.ts:1499 | unavailable: NotYet table has no owners |
| 30 | NotYet: a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } | 10,567 | 2 | programDiagnostics.ts:89 | unavailable: NotYet table has no owners |

Other causes, also counted once:

| Cause | Hidden bytes |
|---|---:|
| SkippedDependency | 87,747 |
| checker | 1,730,511 |
| conflict | 422 |
| panic | 3,647 |

Reproduce from the repository root:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/hidden-ranking-setup.log 2>&1
source /workspace/adamic-tools/env.sh
python3 stage3/census/hidden-ranking/test_ranking.py > /tmp/hidden-ranking-tests.log 2>&1
python3 stage3/census/hidden-ranking/ranking.py > /tmp/hidden-ranking-result.log 2>&1
python3 stage3/census/hidden-ranking/render.py
```

The committed fixture is a synthetic full census with an outer [10,60) boundary,
a nested [20,30) boundary, a duplicate outer observation and a disjoint [70,90)
boundary. Independent examination removes [40,50). The byte-by-byte oracle
expects exactly 40 outer-reason bytes and 20 inner-reason bytes, total 60;
the nested span receives zero extra credit. Separate tests cover checker ancestors,
conflicting equal spans and crossing spans. The mutant increments the credited
total by one only when a nested boundary is active. It is caught by
`each hidden byte must count exactly once`. The proof runner exits zero only
when that assertion rejects the mutant. All four tests passed; the ranking
command reproduced every original hidden range and partitioned 3,970,749 bytes.
Logs are retained in evidence/. No native/oracle fixture or counts.md changed.
No whole packages or full gate were run.

Setup cumulative timings: Node 0.018s, Go 0.020s, markdown ready 0.055s,
submodules 0.056s, clang 0.112s, Go build 30.167s, test binaries deferred 30.234s,
cache warm 30.235s, done 30.259s. `nproc` is 5; CPU quota is 4.
