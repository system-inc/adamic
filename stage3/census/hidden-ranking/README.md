- Step 05: reran hidden source and ranking on compiler 69501280.
- Hidden: **3,654,880 / 10,615,807 bytes (34.428659%)**; delta **-315,869 bytes** from ed6e2975.
- Refusal/NotYet reasons receive 1,949,534 bytes; other causes receive 1,705,346 bytes.
- Four synthetic accounting tests pass; the nested-byte double-count mutant is caught.
- Owners use the pinned tables where supplied; NotYet ownership and actual fix-alone outcomes remain unavailable.

[RESULT.json](RESULT.json) contains all 1,712 refusal/NotYet reasons, including
zero-credit reasons, all other causes, every raw boundary/skip record and every
attributed half-open byte segment. This unit reruns the complete
[hidden-source measurement](../hidden/README.md), compiler `69501280a81259fb512edbb8dd0e52c6eb0d88c8`,
on area-next-drop 784b577a. The compiler is built in an isolated checkout;
no compiler features are merged into this reporting branch. All 82 adapted input
file hashes and byte lengths match the original ed6e2975 run.

[DELTA.json](DELTA.json) contains every file and exact reason's before/after
values and ranks, including disappeared and newly observed reasons. The original
artifacts remain at report commit 6c4fc1af.

| Metric | ed6e2975 | 69501280 | Delta |
|---|---:|---:|---:|
| Hidden bytes | 3,970,749 | 3,654,880 | -315,869 |
| Hidden share | 37.404118% | 34.428659% | -2.975459 percentage points |
| Blocked union bytes | 6,815,024 | 6,735,528 | -79,496 |
| Independently examined bytes | 2,844,275 | 3,080,648 | +236,373 |
| Any-return boundaries | 2,060 | 15 | -2,045 |
| Any-return credited bytes | 89,627 | 7,290 | -82,337 |
| Any-return rank | 8 | 34 | +26 |

The any-return bucket reproduces the compiler worker's 15 boundaries / 7,290
bytes. Its reduction is not the total hidden-byte delta: other reasons can stop
newly exposed statements. Neither change proves successful native lowering.
The total delta compares the entire ed6e2975 and 69501280 compiler pins; it does
not isolate the mapper fix from the intervening area changes.

Every one of the 20,696 Boundary records is matched by exact diagnostic text to
its recorded Refused, NotYet, panic/error or SkippedDependency cause. Panic/error
records sharing the same panic text are one panic cause. The 147 checker-skipped
body records and 29 dependency body records are included as checker causes,
matching the current hidden-source union. Current per-file hidden ranges are
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
evidence/ and RESULT.json. 18 exact reasons have supplied owners.

The top 30 refusal/NotYet reasons follow, sorted by descending credited bytes.
Other causes are excluded from this compiler-reason list and accounted below.

| Rank | Kind and exact reason | Bytes revealed if fixed alone | Boundaries | Three file:line examples (when available) | Owner from supplied tables |
|---:|---|---:|---:|---|---|
| 1 | NotYet: a value of type __String | 168,577 | 401 | factory/nodeFactory.ts:492<br>binder.ts:509<br>transformers/typeSerializer.ts:450 | unavailable: NotYet table has no owners |
| 2 | NotYet: a function returning CapturedThis | 129,056 | 3 | transformers/es2015.ts:488 | unavailable: NotYet table has no owners |
| 3 | NotYet: a value of type InitializedVariableDeclaration | 117,867 | 5 | transformers/module/module.ts:174<br>transformers/generators.ts:336 | unavailable: NotYet table has no owners |
| 4 | NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class | 117,289 | 2466 | factory/emitHelpers.ts:148<br>parser.ts:10625<br>transformers/module/system.ts:536 | unavailable: NotYet table has no owners |
| 5 | NotYet: a computed field name | 113,387 | 4 | visitorPublic.ts:623<br>parser.ts:505<br>scanner.ts:135 | unavailable: NotYet table has no owners |
| 6 | NotYet: a value of type PrivateIdentifierInExpression | 91,578 | 3 | transformers/classFields.ts:354 | unavailable: NotYet table has no owners |
| 7 | NotYet: a function returning ImmediatelyInvokedArrowFunction | 87,684 | 5 | transformers/esDecorators.ts:295 | unavailable: NotYet table has no owners |
| 8 | Refused: overload 1 of writeTokenText result void cannot be served by implementation result number | 62,495 | 10 | emitter.ts:1211 | unavailable: exact reason absent |
| 9 | NotYet: a value of type Path | 61,844 | 156 | resolutionCache.ts:579<br>watchUtilities.ts:572<br>resolutionCache.ts:348 | unavailable: NotYet table has no owners |
| 10 | NotYet: a value of type ParameterPropertyDeclaration | 59,747 | 3 | transformers/ts.ts:235 | unavailable: NotYet table has no owners |
| 11 | Refused: overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem | 59,499 | 4 | transformers/declarations.ts:259 | unavailable: exact reason absent |
| 12 | Refused: overload 1 of transformFunctionBody result Block cannot be served by implementation result ConciseBody | 41,822 | 4 | transformers/es2018.ts:154 | unavailable: exact reason absent |
| 13 | NotYet: a value of type PrimitiveLiteral | 37,568 | 3 | expressionToTypeNode.ts:174 | unavailable: NotYet table has no owners |
| 14 | Refused: a method in object destructuring | 34,480 | 12 | sys.ts:593<br>sys.ts:991<br>utilities.ts:11316 | unavailable: exact reason absent |
| 15 | NotYet: a value of type any | 29,361 | 93 | program.ts:4891<br>commandLineParser.ts:3059<br>moduleNameResolver.ts:1803 | unavailable: NotYet table has no owners |
| 16 | Refused: overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody | 28,699 | 3 | transformers/es2017.ts:119 | unavailable: exact reason absent |
| 17 | Refused: a cast the runtime can't check | 27,462 | 1047 | sys.ts:419<br>moduleSpecifiers.ts:1123<br>utilities.ts:3435 | internal/lower/cast_proof.go / lowering.castProof |
| 18 | NotYet: a destructured name that isn't plain | 26,845 | 3 | utilities.ts:11491 | unavailable: NotYet table has no owners |
| 19 | NotYet: a function returning T &#124; undefined | 24,802 | 87 | moduleNameResolver.ts:1166<br>moduleSpecifiers.ts:713<br>utilities.ts:2923 | unavailable: NotYet table has no owners |
| 20 | NotYet: a value of type ResolvedConfigFileName | 20,619 | 17 | tsbuildPublic.ts:932<br>tsbuildPublic.ts:1890<br>tsbuildPublic.ts:608 | unavailable: NotYet table has no owners |
| 21 | NotYet: a function returning T | 18,006 | 104 | transformer.ts:248<br>utilities.ts:12345<br>executeCommandLine.ts:983 | unavailable: NotYet table has no owners |
| 22 | NotYet: for...of over an object | 16,604 | 128 | transformers/utilities.ts:194<br>moduleNameResolver.ts:2716<br>transformers/utilities.ts:304 | unavailable: NotYet table has no owners |
| 23 | Refused: Object.entries | 15,966 | 1 | utilities.ts:1384 | unavailable: exact reason absent |
| 24 | Refused: var | 15,286 | 65 | parser.ts:1568<br>scanner.ts:1061<br>parser.ts:1600 | unavailable: exact reason absent |
| 25 | Refused: overload 1 of parenthesizeConciseBodyOfArrowFunction result Expression cannot be served by implementation result ConciseBody | 14,206 | 4 | factory/parenthesizerRules.ts:54 | unavailable: exact reason absent |
| 26 | NotYet: an array of never | 12,123 | 106 | utilities.ts:9041<br>commandLineParser.ts:3335<br>utilities.ts:9838 | unavailable: NotYet table has no owners |
| 27 | Refused: overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram &#124; undefined | 12,027 | 5 | builder.ts:1673<br>builderPublic.ts:223<br>builderPublic.ts:184 | unavailable: exact reason absent |
| 28 | NotYet: a function returning Path | 11,989 | 20 | watchUtilities.ts:118<br>program.ts:1514<br>program.ts:1499 | unavailable: NotYet table has no owners |
| 29 | NotYet: for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | 10,540 | 13 | moduleSpecifiers.ts:927<br>commandLineParser.ts:2788<br>moduleNameResolver.ts:2428 | unavailable: NotYet table has no owners |
| 30 | NotYet: a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } | 10,307 | 2 | programDiagnostics.ts:89 | unavailable: NotYet table has no owners |

Other causes, also counted once:

| Cause | Hidden bytes |
|---|---:|
| SkippedDependency | 87,831 |
| checker | 1,613,177 |
| conflict | 691 |
| panic | 3,647 |

Reproduce from the repository root:

```sh
# Build at the exact compiler pin in an isolated checkout; share its pinned cohere.
export GOPROXY='https://proxy.golang.org|direct'
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/hidden-step05-tmp
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/hidden-step05-overlay > "$TMPDIR/overlay.log" 2>&1
gofmt -w /tmp/hidden-step05-overlay/*.go
go build -buildvcs=false -overlay=/tmp/hidden-step05-overlay/overlay.json -o "$TMPDIR/census" ./stage3/census/latent/tool > "$TMPDIR/build.log" 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 "$TMPDIR/census" /tmp/hidden-adapted/src/compiler "$TMPDIR/full.jsonl" > "$TMPDIR/run.log" 2>&1
# From this reporting branch, using the frozen input verified against the original manifest:
NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules" node stage3/census/hidden/units.cjs /tmp/hidden-adapted/src/compiler "$TMPDIR/stock.json" > "$TMPDIR/stock.log" 2>&1
python3 stage3/census/hidden-ranking/step05.py /tmp/hidden-adapted/src/compiler "$TMPDIR/full.jsonl" "$TMPDIR/stock.json" --commit 69501280a81259fb512edbb8dd0e52c6eb0d88c8 > "$TMPDIR/result.log" 2>&1
python3 stage3/census/hidden-ranking/test_ranking.py > "$TMPDIR/ranking-tests.log" 2>&1
python3 stage3/census/hidden-ranking/test_step05.py > "$TMPDIR/delta-tests.log" 2>&1
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
command reproduced every current hidden range and partitioned 3,654,880 bytes.
The new known-answer delta fixture also catches a headline +1 mutant and a changed
source-hash mutant, plus a changed-denominator mutant. The hidden arithmetic suite,
its nine mutants, independent byte-mask audit and real nested recovery witness are rerun on the corrected binary.
Logs are retained in evidence/. No native/oracle fixture or counts.md changed.
No whole packages or full gate were run.

Setup on the reporting branch reached Go/Node 0.022s, submodules 0.053s,
markdown 0.060s, clang 0.121s, then failed during Go dependency export with a
full /tmp filesystem. Scratch setup also rejected its shared cohere symlink:
`expected submodule path cohere not to be a symbolic link`. The workaround
retains the initialized matching submodule and tools, and places Go scratch and
new measurement output on /workspace. Logs are retained in evidence/step05.
`nproc` is 5; CPU quota is 4. No production compiler files were edited.
