Verified compiler sha: eb0427470cdc428b5de18bce73b49eeca6f5887f; 14 reproduced binder stops cleared, 11 executable per-kind reductions, 3 contract rulings, 2 echoes cancelled.
Commits: e662629c replay merge; 42207bc7 census witnesses; 593c3dbc union/optional inference; d448df93 callbacks; 69fb5eba maps/state; 1df0334a refusal probes; eb042747 current-main merge.
Commands/outputs: focused lower tests PASS (0.392s), uncached oracle PASS (2.596s), counts refresh PASS (28.656s); all 16 final replays no longer find the specified stop.
Mutants: all 10 caught, including all 11 executable per-kind fixtures for the old count guard and all 3 Node-held refusal probes for contract bypasses; individual failures in mutants.log.
Not covered: full compiler helpers or downstream stops; no IR/backend/runtime C changes; no territory skips remain; refused probes cannot reach backend sanitizers or counts.

Base and evidence

The explicit unit base was origin/area/compiler b410340dc8f889b5799c3bc519117c63def3aa24, already containing then-current main d65e2d5e. Replay 9a1f14c5d994aa855625e7cfa295677060348fec was merged. Table evidence came from codex/stage3-notyet-table 57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7. The latent-full snapshot's stage3/apply.sh produced /tmp/notyet-overloads-adapted; all 81 adapted compiler source hashes matched the table source manifest. These census measurements use a checker-rejected entry-root program, and are not evidence that the entire compiler typechecks or compiles. Compact original and final findings, including downstream panics, are in replays.json.

Current origin/main 749a69adbfae2a7bf22c1f0436d9ef73345c3070 was merged as eb042747 before final replay and focused re-greening. No conflicts. No whole package or full gate was run manually. Setup's prescribed build took 258.401s: Go 0.059s, Node 0.058s, clang 0.470s, markdown dependencies 1.071s, submodules 16.351s, Go build 258.256s, cache warm 258.369s. nproc=5, cpu.max=400000 100000 (4 CPU quota); Go 1.27.1, Node 24.19.0, clang 20.1.8. GOPROXY was set before setup. Setup succeeded.

Per kind (requested order; every row represents one listed root site)

“Lowered” means the additional-binder proof is implemented and the reduced fixture passes source Node, JavaScript backend Node, native release, ASan/UBSan and LeakSanitizer. It does not mean the complete compiler helper reaches native code. Every reproduced original stop is gone. Final replay intentionally exits 1 when the requested exact signature is absent; its JSON names the next stop.

| Sha | Kind | Root sites covered | Status | Fixture suffix (`internal/oracle/testdata/overload_*.a`) | Final source stop |
|---|---|---:|---|---|---|
| 593c3dbc | arrayToMap | 1 | lowered binder; next stop remains | array_to_map | Refused: overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 \| V2> |
| 593c3dbc | arrayToMultiMap | 1 | lowered binder; next stop remains | array_to_multimap | Refused: overload 2 of arrayToMultiMap result MultiMap<K, U> cannot be served by implementation result MultiMap<K, U \| V> |
| 593c3dbc | arrayToNumericMap | 1 | lowered binder; next stop remains | array_to_numeric_map | Refused: overload 2 of arrayToNumericMap result U[] cannot be served by implementation result (T \| U)[] |
| 1df0334a | createBinaryExpressionTrampoline | 1 | refused for a ruling | trampoline_contract | Refused: overload 1 of createBinaryExpressionTrampoline result (node: BinaryExpression) => TResult cannot be served by implementation result (node: BinaryExpression, outerState: TOuterState) => TResult |
| 1df0334a | createToken | 1 | refused for a ruling | token_contract | Refused: overload 1 of createToken result SuperExpression cannot be served by implementation result Mutable<Token<TKind>> |
| d448df93 | forEachAncestorDirectory | 1 | lowered binder; next stop remains | ancestor_directory | NotYet: a function returning T \| undefined |
| d448df93 | forEachLeadingCommentRange | 1 | lowered binder; next stop remains | leading_comment_range | NotYet: a function returning U \| undefined |
| d448df93 | forEachTrailingCommentRange | 1 | lowered binder; next stop remains | trailing_comment_range | NotYet: a function returning U \| undefined |
| d448df93 | getOriginalNode | 1 | lowered binder; next stop remains | original_node | NotYet: a function returning T \| undefined |
| 69fb5eba | mutateMap | 1 | lowered binder; next stop remains | mutate_map | Refused: a method in object destructuring; panic: statement panic: runtime error: invalid memory address or nil pointer dereference; error: statement panic: runtime error: invalid memory address or nil pointer dereference; Refused: a cast the runtime can't check; NotYet: a call through ?. (an optional call) |
| 69fb5eba | mutateMapSkippingNewValues | 1 | lowered binder; next stop remains | mutate_map_skipping_new | Refused: a method in object destructuring; NotYet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions |
| 69fb5eba | resolveTypeReferenceDirectiveNamesReusingOldState | 1 | lowered binder; next stop remains | resolve_type_names | NotYet: a value of type unknown; NotYet: a destructured name held otherwise than its field |
| 1df0334a | setSerializerContextAnd | 1 | refused for a ruling | serializer_contract | Refused: overload 2 of setSerializerContextAnd parameter cb cannot be served by implementation parameter cb |
| 69fb5eba | sortAndDeduplicate | 1 | lowered binder; next stop remains | sort_deduplicate | Refused: a cast the runtime can't check; panic: statement panic: runtime error: invalid memory address or nil pointer dereference; error: statement panic: runtime error: invalid memory address or nil pointer dereference |
| baseline | arrayFrom | 0 (1 listed, never reproduced) | skipped: cancelled echo (earlier refusal already blocked before change) | none | Refused: overload 1 of arrayFrom result U[] cannot be served by implementation result (T \| U)[] |
| baseline | group | 0 (1 listed, never reproduced) | skipped: cancelled echo (earlier refusal already blocked before change) | none | Refused: overload 2 of group parameter resultSelector cannot be served by implementation parameter resultSelector |

Implementation and conservative assumptions

censusOverload infers witnesses from both parameters and results when implementation binders outnumber overload binders. It retains positional alpha-renaming otherwise. An unobserved binder gets bottom as a candidate; constraints, strict parameter contravariance, invariant mutable representations, and covariant results still have to prove it valid. Bottom is not permission to erase an observable type. Existing strict relations remain in force.

The one shared production hook is internal/lower/generic.go inferTypes. A union with exactly one unknown binder can infer the target only if all other members fit that target. Multiple unknown binders stay unresolved. Optional inference removes explicit undefined while preserving rigid T, rather than replacing T with T & {}. Worker branch history was fetched and checked before touching this helper; no other worker's lower function was edited. Each commit names outside files. No C runtime helpers need runtime-owner review.

The three array reductions exercise the first identity overload, retaining additional mapper-result binders and generic implementations. They omit later mapper overloads that independently fail invariant result proofs. The numeric array is prepopulated before permutation writes to avoid an unrelated growth stop. The ancestor reduction preserves branded Path/string overloads and optional callback results but replaces traversal with one callback. Comment reductions retain four-argument callback binders and absent state without implementing scanner traversal. The original-node reduction preserves the nullable generic identity contract, not origin-chain traversal. Map reductions preserve the set-source overload and phantom map-value binder while omitting the separate map-source arm. Name resolution preserves FileReference/string overloads and computes a count, not filesystem resolution. Sorting keeps generic string ordering/deduplication, not the compiler's comparer machinery. These are reductions of the binder rule, not claims of whole-helper equivalence.

Rulings and remaining stops

createToken's reduced overload promises a required field missing at runtime (Node prints undefined); Adamic's result refusal stays. The trampoline result promises a one-argument function while implementation advertises a required second argument; the reduced JavaScript ignores it and Node prints node, but the strict arity proof still refuses. Serializer overload 2's callback requires an argument, while the implementation callback type admits omission; Node's reduced correlated branch prints node/arg, but accepting that correlation needs a separate proof/design ruling. All three are independently asserted by TestOverloadContractRulings in its own registration/test file.

The full three array helper sets also reach later overload-2 mutable result refusals. These must retain the refusal or gain a separate representation-safe proof; this change does not reinterpret mutable Map/array variance. arrayFrom overload 2 and group overload 3 are cancelled census echoes because the baseline replay already refused earlier overloads (arrayFrom overload 1 result; group overload 2 callback). Neither has an executable fixture or mutant claim.

The four optional-return source bodies reach the next named generic-return stop. Other named source stops include object method destructuring, unsupported generic map keys, unknown values, destructured-name representation, and unchecked casts. mutateMap and sortAndDeduplicate additionally expose census continuation panics after earlier stops; they are retained in evidence and are unresolved, not successful lowering claims.

Mutant proof table

| Mutant | What caught it |
|---|---|
| additional-binder-refusal | census_overload_binders plus every one of the 11 executable per-kind fixtures |
| parameter-inference omitted | census_overload_binders |
| bottom-witness replaced with implementation function type | census_overload_binders |
| constraint check disabled | TestCensusOverloadBinderGuards/constraint |
| parameter relation disabled | TestCensusOverloadBinderGuards/parameter plus serializer contract refusal probe |
| result relation disabled | TestCensusOverloadBinderGuards/result plus token and trampoline refusal probes |
| union-inference disabled | TestOverloadInferenceWitnesses/one_compatible_binder and all 3 array fixtures |
| known-union-member compatibility disabled | TestOverloadInferenceWitnesses/incompatible_known_member |
| multiple-union-binders ambiguity accepted | TestOverloadInferenceWitnesses/multiple_unknown_binders |
| optional-rigid-binder replaced by old nonnullable transform | TestOverloadInferenceWitnesses/optional_rigid_binder and ancestor fixture |

Every mutant run exited 1 with an intended test failure, not a build failure. Compiler files are restored in finally blocks. An early unused-variable mutation was rejected as a build failure and corrected before counting it. An overly broad optional-rigidity mutation survived three other callback reductions; only ancestor depends on that shared rule, so the runner requires its failure and the direct rigid-binder test. The other callbacks each fail the common count-guard mutant. No survived mutant is counted as caught.

Reproduction

Source /workspace/adamic-tools/env.sh before commands. validation.log records exact focused commands and outputs, setup timing, and the final per-kind replay summary. replays.json records exact site and reason arguments. run-mutants.py is the repeatable mutation runner; mutants.log includes each failing test. All test output was redirected to log files, never piped. Each positive fixture is registered from an overload-specific _test.go file. counts.md was refreshed after each positive group and after the main merge; refusal-only probes have no counted native build.
