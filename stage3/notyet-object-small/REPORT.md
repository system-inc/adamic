Built all sound object-small rules, with fixed-shape and index-signature refusals retained.
Commits: 17d024d4, 76316323, 43af2f34, c0179eda, 9104bc07, b33cb978, 600b55ca, c3d36182; non-null merge 002c5271; main merge e8ac282d.
Commands: focused Node/native/JavaScript oracle and touched-package checks passed; counts refreshed; logs are in evidence/.
Mutants: every mutation listed below failed its fixture; the stale union check test also runs an IR mutant in both backends.
Not covered: blocked compiler entry roots, open-key writes, heterogeneous/fixed tuples, NUL-bearing names; comparator is a census echo.

# Scope and bases

Own branch: codex/notyet-object-small. Area/compiler was resolved to b410340dc8f889b5799c3bc519117c63def3aa24. Census replay 9a1f14c5d994aa855625e7cfa295677060348fec was merged as be61cfbb. Table source: codex/stage3-notyet-table at 57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7. Requested checked-non-null c41c0e062e99da37820f822968d4df1b48cdaee7 was merged as 002c5271, preserving both sets of counts rows. Current main efe9f404 was merged as e8ac282d. Only this personal branch was pushed; no PR.

CLAUDE.md and object.go were read whole. Lower changes outside the assigned functions are small called helpers and representation/presence hooks, named in their commits. Remote worker histories were checked before touching helpers. No code was copied from cohere. New fixtures are .a.

Setup used GOPROXY=https://proxy.golang.org|direct and bash cloud/setup.sh, then source /workspace/adamic-tools/env.sh. nproc=5, CPU quota=4. Timing lines: node .021s, Go .033s, clang .243s, npm step .889s, markdown .958s, submodules 16.486s, build 230.082s, tests deferred 230.180s, cache warm 230.181s, done 230.208s. Setup succeeded; no workaround. Full setup output is preserved.

# Per-kind outcome, largest first

Counts below separate rule fixture coverage from original compiler roots reached. Replays measure lowering on a checker-rejected program and do not prove that the compiler project compiles or emits.

| SHA first | Kind | Outcome and covered roots |
| --- | --- | --- |
| c0179eda | union field storage (5) | Lowered. Tagged storage, actual scalar layouts, and narrowed-member checks. 3/5 roots directly clear: utilities 2461, 2475, 8745. Builder 1524 remains blocked by Path at 1520:95; utilities 8651 by earlier binary-expression/generic-function work. |
| 9104bc07, 600b55ca | assigning an element of a value (4) | Lowered finite keys naming existing required data fields. 0/4 original roots fully lower. Checker 19531 reaches the fixed-shape Refused (1/4 directly classified); debug 189 is blocked by generic K; sys 172 by its RHS element read; parser 10794 reaches an open-key NotYet. Optional own-field growth needs a ruling, not an expando implementation. |
| b33cb978, c3d36182 | tuple indexOf (2) | Lowered homogeneous trailing-rest tuples as arrays. 2/2 original roots reproduced before and no longer stop at indexOf; after non-null merge both reach declaration directly in a case at moduleSpecifiers 1389:13. |
| 17d024d4 | push with other than one value (2) | Lowered zero/multiple arguments, evaluating all arguments before appending. Fixture covered. 0/2 entry roots reached: both utilities roots are blocked by BinaryExpression value/value at 7750:9. |
| 43af2f34 | spreading arrays of other elements (2) | Lowered using the iterable element representation. 2/2 old stops removed across replays: utilities 12168 has no findings; scanner 3534 reached its next identity callback before the non-null merge, and now also encounters earlier generic callback/iteration stops. |
| 600b55ca | spread adds a field (1) | Lowered plain-object added-field spreads with exact own presence and preserved hidden fields. Tracing Args has an index signature, retained Refused for a ruling. 0/1 original entry roots directly reached after merge: Debug.assert reaches unknown at debug 213:28 first. |
| b33cb978, c3d36182 | tuple push (1) | Lowered homogeneous trailing-rest tuple push. 1/1 old tuple stop reproduced and removed; next stop remains unknown in debug 213:28. |
| 76316323 | slice optional numeric index (1) | Lowered undefined/NaN numeric bounds with exact defaults and evaluation order. 1/1 reaches next named stop: array of T at core 987:34. |
| no production change | comparator signature (1) | Cancelled census echo, 1/1 classified. Reproduced named stop; scratch-only arraySort hook prepares latentSignature(function), then the stop disappears. Repeated after non-null merge. Lazy census signature preparation was absent on this path; accepting a bad comparator was not needed. |
| b33cb978 | omitted ModuleSpecifierEnding tail (1) | Lowered true variable-length rest tuple literals without padding the tail. 1/1 directly clears; replay has no findings after merge. |

# Semantics and rulings

Union fields capture the stored tag before checking/unboxing a narrowed member. A stale narrowing panics in both backends instead of silently reading a different representation. Scalar object and public class layouts keep their actual storage identity through readonly union views and added-field spreads.

Computed writes capture receiver, key, and RHS once in JavaScript order, then dispatch only among finite required own fields. docs/0.1.md fixes an object's shape when made and refuses expandos; checker optional cache growth therefore stays Refused. Open string keys remain unsupported. Index signatures are refused by the existing doctrine (use Map); the tracing reduction proves that refusal and Node's observation separately. Ruling requested for both design cases, rather than silently widening the language.

Added-field spreads allocate a fresh merged own layout. Copy/getters happen before RHS evaluation; hidden fields survive readonly source views; absent optional sources do not acquire synthetic own properties. Existing throwing-accessor spread refusals remain in force. No existing runtime C file was changed. New separate helpers for runtime-owner review: internal/native/runtime/object_spread_extend.c and .h. Their cache holds only immutable, canonical ordered layout metadata, reachable for process lifetime, never object values. Canonicalization prevents repeated identical spreads from growing a chain of layouts. Field origins preserve scalar-versus-tagged slot interpretation.

NUL-bearing C field names remain an explicit NotYet for new computed writes and extending spreads, because current runtime names use C strings. Dedicated Node reductions and disabling-guard mutants prove both boundaries. Fixed/heterogeneous tuples keep their old object layout. Only homogeneous required-prefix/trailing-rest tuples with compatible representations become arrays. Their required reference reads now use the existing presence check after pop; Node's TypeError matches sanitized, release and JavaScript runs.

# Mutants actually run

Each source mutant was restored after its focused test failed. A failure from a type/build error was not accepted as evidence.

| Rule | Mutant | What caught it |
| --- | --- | --- |
| push argument order | reverse appended values | Node/native and JavaScript stdout mismatch |
| slice start | undefined start becomes one | both backend stdout mismatches |
| slice end | undefined end becomes zero | both backend stdout mismatches |
| iterable spread | re-query array-only element type | reduced fixture LowerNotYet |
| union slot | restore slotless union refusal | reduced fixture LowerNotYet |
| union scalar layout | remove actual layout discrimination | native numeric/boolean stdout mismatch |
| union unbox | bypass field unboxing | UBSan invalid bool |
| union narrowed member | remove generated IR member guard | in-test mutant returns native 1 / JavaScript wordswordswords1 instead of checked panic |
| computed write slot | choose wrong field | both backend stdout mismatches |
| computed write order | reorder captured arguments | both backend stdout mismatches |
| rest tuple representation | disable array recognition | reduced fixture lowering failure |
| rest tuple literal | pad absent tail | Node/native length and value mismatch |
| tuple indexOf | restore tuple method refusal | reduced fixture LowerNotYet |
| tuple push | restore tuple method refusal | reduced fixture LowerNotYet |
| tuple reference presence | omit array-tuple presence hook | UBSan null object access; release and Node exits differ |
| spread snapshot | copy after RHS | source side-effect stdout mismatch |
| spread field offsets | keep uniform offsets for added fields | ASan invalid reference/string access |
| spread storage origin | use merged shape instead of original field storage | ASan invalid scalar/reference access |
| canonical shape cache | disable reuse | native C harness exits 4 during repeated-layout assertion |
| optional source presence | synthesize absent source field | hasOwnProperty false versus true |
| added method layout | omit extension for added method | missing runtime field versus Node |
| index signature boundary | remove index-signature refusal | dedicated test receives nil instead of Refused |
| optional expando boundary | remove optional-field refusal | dedicated test receives nil instead of Refused |
| spread name boundary | disable NUL-name guard | dedicated test no longer gets expected NotYet |
| computed name boundary | disable NUL-key guard | dedicated test no longer gets expected NotYet |

# Commands and evidence

All test output went to files. No whole package suite or full gate was run. Earlier kind-specific fixture, package and counts logs are retained in evidence/.

Final focused commands (also rerun after merging main):

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestCheckedNonNull|TestObjectSmall.*Refusal|TestObjectSmallUnionFieldCheck|TestNativeAgreesWithNode/internal/oracle/testdata/object_small_' -count=1 -timeout 10m
go test ./internal/lower -run 'TestATupleSeenAsArrayIsNotYet|TestProvenRelations|TestNonNullAssertion|TestPrototypeHazardsBehindObjectViewsAreNotYet' -count=1
go test ./internal/native -run 'TestObjectSpreadExtendedLayouts|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestOptionalWriteMissingSlotRemainsChecked' -count=1
go test ./internal/ir -run TestCallTargetsIncludeEveryDescendant -count=1
go test ./internal/javascript -run '^$'
go test ./cmd/adamic -run TestNonNullExplainChecks -count=1
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts
```

JavaScript has no package tests; its build and every executable fixture's JavaScript oracle run passed. Counts update after spread passed in 23.609s and after tuple-pop in 26.925s. Before main merge, all oracle cases passed in 3.551s; cmd controls in 2.530s. Post-main checks passed: oracle 3.094s, lower .552s, native .666s, IR .041s; JavaScript package builds with no package tests. Large replay logs are losslessly gzip-compressed.

Replay command uses the exact compiler entry, not its directory:

```
go run ./stage3/census/latent/replay -project /tmp/object-small-adapted/src/tsc/tsc.ts -where /tmp/object-small-adapted/src/compiler/FILE:LINE:COL -kind NotYet -reason 'EXACT TABLE REASON'
```

Initial directory-root exploratory replays were superseded by entry-root runs. A removed signature exits 1 with 'signature did not reproduce'; that is evidence of the next stop, not a backend compilation claim. The checker optional-field refusal replay exits 0 with its new exact Refused signature. evidence/replay-findings.json summarizes the preserved raw post-merge logs. The sort proof uses a scratch Go overlay only, adding latentSignature(function) before reading the comparator's IR declaration; no production arraySort change or copied checker implementation.
