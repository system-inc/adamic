# Census: carry program record facts

**Byte identity is proven on the subset only: binder.ts, emitter.ts, and utilities.ts. The changed census has not completed an all-79-files comparison.** The original 16 GiB failure remains in [full-baseline-blocker.md](full-baseline-blocker.md), [baseline-resources.json](baseline-resources.json), and [full-baseline-partial-timings.csv](full-baseline-partial-timings.csv).

The existing `LATENT_ONLY_FILE` exact absolute filename filter selects attempts while keeping the complete 79-file program, checker diagnostics, declarations, and sibling bindings. Each selected file ran in a separate process. Utilities contains the two requested units; binder and emitter have many expression attempts, and emitter produced 11 record-storage read findings (utilities: 18). All completed subset runs stayed below 2.4 GiB observed peak RSS; **no file was dropped**. Every run, including profiling, answer logging, and the intermediate Boolean-only run, is listed in [resources.csv](resources.csv). One preliminary old binary was terminated after a scratch editing assertion failed; it is marked `pre-extension-aborted`, not used for comparison.

## Measured answers

The CPU profile's sampled chain is `enumNeverValue → property → entriesProgramHasRecords → entriesProgramRecordsWalk → entriesRecordLiteral → entriesRecordLiteralWalk`. The two original hotspots reach the same aggregate record answer, not two independent global Booleans. See [profile counts/costs](profile-walks.csv), [pprof top](utilities-profile-top.txt), and [call edges](utilities-profile-walk-chain.txt) and [sampled stack](enum-record-stack.txt). Times below are instrumented wall costs, nested costs must not be added together.

| Answer / work | Calls across subset | Mean seconds/call | Equality and decision |
|---|---:|---:|---|
| `entriesProgramHasRecords` cold `entriesProgramRecordsWalk` | 90 | 3.067210 | `true` in every one of 71 asking units; carry `entriesProgramRecords` and its known flag. |
| `entriesRecordLiteralWalk` / `entriesRecords` verdicts | 112,925 | 0.002432 | The expensive aggregate walk fills this map. 89,229 unit/literal observations, 1,256 literals, zero disagreements, eight consistently true literals. Carry its completed map too; otherwise storage-type discovery repeats the expensive aggregate work. |
| `entriesRecordStorageTypes` | 132 | 0.011878 | Actual checker-type identity slices differ within a run: 23 distinct answers/23 utility units, 11/12 binder units, 34/36 emitter units. **Do not carry** either this slice or its known flag. |
| `runPredicateFixedPoint("kind")` | 327 | 0.072248 | Stable logged result within each run; below the 0.1 s mean threshold, leave unit-local. Eleven calls exceeded 0.1 s; no other mode ran. |
| `assignedSymbol` table | 74 | 0.000810 | Stable logged map within each run; below threshold, leave unit-local. |
| `factoryFieldOrigin` table | 3,016 | 0.000024 | Always empty in the logged subset asks; below threshold, leave unit-local. |

`objectLiteral` (112 calls, mean 1.134413 s) includes nested record-walk cost and depends on the literal/unit; it is not an independent program answer. `view` (1,246 calls, mean 0.069081 s; 148 over 0.1 s) registers unit-owned views and depends on source/target/node. `predicateReferenceDirections` (98 calls, mean 0.027914 s; two over 0.1 s) is reference-dependent. None is carried. The single literal walk over 0.1 s is retained in the cost ledger. These exhaust the instrumented direct whole-program walkers with calls over 0.1 s, plus the indirect predicate fixed point checked against pprof.

[Boolean answers](program-record-answers.csv), [literal verdicts and asking-unit counts](literal-verdicts.csv), [all literal observations](literal-observations.jsonl.gz), [candidate equality counts](answer-equality.csv), and an actual [storage-type counterexample](storage-types-counterexample.json) retain the evidence. Comparison of pointer-bearing answers is within one process, never between processes. All three scratch CPU-profile runs and all three literal-proof runs also produced byte-identical baseline JSONL.

## Carry and guard

Full-mode census discovers the record answer once on a disposable lowering, before unit copies. The typed statecopy policy table captures only `entriesRecords`, `entriesProgramRecords`, and `entriesProgramRecordsKnown`, and installs them on each file base. The Boolean scalars copy by value; each unit and statement snapshot retains its own typed map copy. No checker-type slice is shared. The disposable state's module-read proofs are discarded; the overlay establishes those effects on the unit's original first eligible literal/aggregate query, with a snapshot-owned pending flag so rollback restores that timing. One cold record walk now supplies all units of a full census invocation, including across file bases. Legacy and speculative traversal are unchanged.

`TestLatentCarryFactsRequiresEveryLoweringFieldPolicy` calls `t.Parallel()`, uses AST parsing and in-process generation only, and passed in 0.099 s. It requires a policy for all 86 production lowering fields, so a newly added memo cannot evade the guard by using an unexpected name; stale fields, omitted carried fields, missing capture/install, and lost owned map copying fail by field name. A repository-wide function-definition search finds only this test. The planted `entriesRecords` policy-deletion mutant failed, was restored, and is documented in [carry-facts-mutants.md](carry-facts-mutants.md).

`go test ./stage3/census/...` passed in full. Statecopy vet, the overlay non-test lower source vet, and overlay driver vet passed; logs are committed. An initial extra package-wide overlay vet attempted production tests and failed on `lowerCheckedWithCycleTypes`, which this measurement overlay intentionally removes. The successful explicit non-test source vet covers the changed overlay without those incompatible production tests; its exact argv is in [overlay-source-vet-command.json](overlay-source-vet-command.json). No file under `internal/` was changed, and no profiling code is committed.

## Timing and identity

| Metric | Baseline | Carried |
|---|---:|---:|
| Sum of three census wall times | 675.345 s | 363.531 s |
| Largest per-run peak RSS | 2,378,200 KiB | 2,417,132 KiB |
| Sum of per-unit lowering wall times | 623.650 s | 303.692 s |
| Sum of per-unit process CPU times | 774.263 s | 445.228 s |
| Attempted units | 1470 | 1470 |

Total wall time is **1.858× faster (46.17% less)**; summed unit lowering time is 2.054× faster. Measurements include cold load/register work and one record probe per filtered process. Wall times use a monotonic clock with 0.5 s completion polling; peak RSS is the maximum observed `/proc/PID/status` VmHWM/VmRSS, sampled every 0.5 s. Runs were sequential. Single baseline/final observations establish measured drops, not confidence intervals.

`getLiteralText`: 4.179028 → 0.486105 s (8.597×). `isTypeDeclaration`: 0.436000 → 0.442929 s (0.984×), essentially unchanged on this pin; the historical 5.2 s was not reproduced for that unit.

Both combined `latent.jsonl` files have SHA-256:

```text
513666b504176e75324695584240a4e8e5b24e63b2404b5fd5766c897ccaf018
```

The combined ledger keeps one identical checker header and the three file records in binder/emitter/utilities order, assembled from three independent filtered processes. [latent.sha256](latent.sha256) also records each individual before/after file hash. Raw findings and per-unit timing JSONL remain under `/workspace/scratch/subset/{baseline,carried}/`. [timing-table.csv](timing-table.csv) contains all 1,470 paired units; the largest twenty drops are below and in [top-20.csv](top-20.csv).

| Unit (file:line) | Before s | After s | Drop s | Before/after |
|---|---:|---:|---:|---:|
| `KindVariableStatement` (utilities.ts:9044) | 65.597 | 2.102 | 63.495 | 31.21× |
| `emitLiteral` (emitter.ts:2118) | 7.362 | 0.091 | 7.271 | 80.65× |
| `emitFiles` (emitter.ts:752) | 16.624 | 9.410 | 7.215 | 1.77× |
| `emitTripleSlashDirectives` (emitter.ts:4313) | 4.254 | 0.355 | 3.900 | 11.99× |
| `emitExportAssignment` (emitter.ts:3734) | 4.236 | 0.359 | 3.877 | 11.79× |
| `isExternalOrCommonJsModule` (utilities.ts:2638) | 3.882 | 0.052 | 3.831 | 75.29× |
| `bindJSDocTypeAlias` (binder.ts:2114) | 4.651 | 0.836 | 3.814 | 5.56× |
| `emitJSDocHeritageTag` (emitter.ts:4166) | 3.939 | 0.137 | 3.802 | 28.81× |
| `emitJSDocSeeTag` (emitter.ts:4134) | 3.954 | 0.199 | 3.755 | 19.88× |
| `projectReferenceIsEqualTo` (utilities.ts:803) | 3.804 | 0.050 | 3.754 | 76.13× |
| `createDiagnosticMessageChainFromDiagnostic` (utilities.ts:2481) | 3.802 | 0.051 | 3.751 | 74.05× |
| `isDiagnosticWithDetachedLocation` (utilities.ts:8637) | 3.786 | 0.048 | 3.737 | 78.39× |
| `writeDirectives` (emitter.ts:4330) | 3.839 | 0.115 | 3.724 | 33.31× |
| `emitJSDocPropertyLikeTag` (emitter.ts:4250) | 4.019 | 0.298 | 3.722 | 13.51× |
| `getLiteralText` (utilities.ts:1980) | 4.179 | 0.486 | 3.693 | 8.60× |
| `moduleResolutionIsEqualTo` (utilities.ts:810) | 3.729 | 0.052 | 3.677 | 71.10× |
| `addDeclarationToSymbol` (binder.ts:635) | 4.236 | 0.571 | 3.665 | 7.42× |
| `emitLeadingComments` (emitter.ts:5976) | 3.862 | 0.206 | 3.657 | 18.79× |
| `shouldEmitBlockFunctionBodyOnSingleLine` (emitter.ts:3476) | 4.029 | 0.374 | 3.655 | 10.76× |
| `emitJSDocImportTag` (emitter.ts:4140) | 4.107 | 0.453 | 3.653 | 9.06× |

Machine: `98b2b7d59531`, `nproc=5`, 17.6 GB reported RAM, cgroup CPU quota `400000 100000` (4 CPUs), 16 GiB memory limit, no swap. The original all-files baseline was killed at 16,616,968 KiB peak child RSS after 2,409.291 s. Compiler baseline is pinned to main's original fetched tip `9c607b11c6d7a6afad29853db5d86b488baaa08e`; this resumes evidence commit `69b4b3bd7205c51dc5b94a1790c3b02d91b9a2b5` on the same branch. Remote main subsequently advanced; these measurements are of the stated pin, not that later revision. TypeScript/adapted-tree HEAD is `050880ce59e30b356b686bd3144efe24f875ebc8`; Cohere is `06af2e399c6c1fb93fd570b774027942747400b4`, its TypeScript submodule `d92d9bfee114c80be2c375d72edae966176e3a4f`. Go 1.27.1, Node 24.19.0, clang 20.1.8. [Adapted-source hashes](adapted-source-shas.csv) cover all 79 input files; [binary hashes](binaries.sha256) identify the measured executables.

[commands.md](commands.md) gives the exact measurement/build/check commands and the **unfiltered baseline-versus-changed command for a house box with more memory**. Do not interpret this subset equality as the unperformed all-79-files proof.
