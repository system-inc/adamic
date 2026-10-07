Re-landed nominal Error classes on the pinned batch 3 parent; all three audit failures are fixed-now.
Merge SHA: 701c5ad6a38432844425cd2ba353155940f863c6; parents a37ebdb0913eca9d3d6e8adbe8f01bb733da52eb and 75a30c4fbe2e26f57cce0fa924e8b20cf8f95926.
Validation: every affected package passes, including uncached native 421.276s and oracle 528.788s; Linux counts, vet and stage 3 byte audit pass.
Mutants: fourteen final root-cause mutants caught by their intended assertions; inherited exception and range mutants remain in the affected gate.
Not covered: the entire test262 corpus and the whole go test ./... gate; unsupported opaque reflection and host virtual formatting have named boundaries.

Compiles to Refused or NotYet in stage 3: none. No status record needs a change.

This unit uses codex/error-classes-counts-2 from the exact requested batch 3 commit, not current main. Only this feature branch will be pushed. Nothing has been pushed to main or an area/ branch. The merge preserves both histories and their tests rather than replaying duplicate patches.

## The three audited failures

| Probe | Reproduced merged behavior | Attribution | Root fix and final status |
| --- | --- | --- | --- |
| stage3/fixtures/nested-functions/06_parser_token_state.a | Recorded Compiles became Refused at 6:17 for a scanner capture cycle. Node still produced its initial/token sequence. | Batch 3's 990eb119f added closedFrameInput's whole-program proof. Incoming nominal Error classes add unused initializer bodies containing SetProperty; the proof incorrectly included those unreachable effects. | fixed-now: compute reachable direct, virtual and closure targets before judging their effects. Unknown closure targets conservatively include every implementation. Unused helpers cannot invalidate the proof; reachable mutation still does. The fixture remains Compiles and agrees with Node under both backends and sanitizers. |
| internal/oracle/testdata/coverage_error_causes.a | NotYet at 24:41: a function viewed as unknown/object required dynamic function descriptors. | Batch 3's 3edfb8fb9 admits unknown views only with complete descriptors. Incoming ErrorOptions cause boxing needs storage and identity, not descriptor reflection. | fixed-now: literal ErrorOptions cause boxing retains the exposed concrete type without requiring reflective descriptors. Concrete record reflection is preserved; a program combining an opaque cause and dynamic property reflection receives an explicit NotYet boundary, including through unknown aliases. Node's cause classification, null/undefined, scalar values and nominal identity match both backends. |
| test262/oracle uncaught exception handler | The import-cycle probe exited 7 with ReferenceError: AdamicPanic is not defined, overwriting the original TDZ ReferenceError. | Incoming 8465ed206 referenced AdamicPanic; batch 3's panic implementation exits directly and has no such class. | fixed-now: remove that reference. Real Node errors keep String(error), including host error codes. Generated error objects use standard Error.prototype.toString rather than accidentally calling a receiver-taking IR method as a JavaScript method. Exception identity, original text, earlier output and exit 70 are tested. |

Reproduction logs: /tmp/adamic-reland-before.log and /tmp/adamic-reland-runner-before.log. The initial cmd/adamic-test262 filter matched no tests; the actual Node import-cycle reproduction and the new permanent runner tests supply the evidence for the third failure.

The reachability conclusion is based on merged IR and the controlled unused-helper mutant. Counts causes below identify source constructs and analysis changes; they are not a sampled profile attributing each retain to a stack frame.

## Other required reconciliations

Batch 3's 611f6e093 made ir.Defined.Throws identify fallible TypeError checks. Incoming precision still retains an invariant check after proving its exceptional path impossible. Defined.Proven reconciles those facts: proven checks retain their runtime assertion but add no throw edge. Existing fallible record-reference checks are normalized to nominal TypeErrors before precision. Nullable String.prototype receiver checks also construct nominal TypeErrors directly. A new record/string-receiver fixture checks TypeError and Error identity after invalidation.

A checked string or function union helper has already proved its tag before returning a narrowed reference. Recognizing that precise IR shape removes a redundant presence throw edge. unions.a and narrowed_union_valid.a exactly match the batch 3 counts. Object tags are deliberately excluded because typeof null is object.

Standard new Error construction initializes its nominal literal directly instead of handing fields through a consuming allocator/initializer. Optional messages still default to the empty string, argument order is preserved, and user subclasses still use their ordinary constructor/super paths. The existing optional-message mutant now removes the default at direct construction, so it continues to exercise real input and is killed by the sanitizer and Node comparison.

The first complete oracle exposed a further merge incompatibility: node:fs's structural runtime errors failed nominal instanceof Error and silently skipped catches. Reproduced on node_fs_file_open.a in /tmp/adamic-reland-fs-review.log. Host errors now have the shared name/message/cause prefix and a derived layout owning code. The uniform field-offset proof includes that four-field runtime layout; its embedded-C guard caught the stale three-field proof, which was corrected. Their Error/TypeError/RangeError ancestry is registered before call finalization; cleanup releases prefix and code once. The node_fs_file_nominal.a fixture checks identity, absent cause, standard formatting and leak-free cleanup. error_host_uncaught.a checks the original Node host TypeError formatting on exit 70.

Virtual Error.toString in a program with throwing filesystem calls is explicitly NotYet (adamic/host-error-to-string), because Node's host subclasses can override it to include their code. Error.prototype.toString.call(error) is supported. This boundary is named in docs/memory.md and held by a mutant; it does not regress any stage 3 Compiles record.

Representation notice for the Error-as-a-value worker: internal/lower/error_classes.go is byte-identical to 75a30c4fb. Its nominal IDs, methods and name/message/cause layout are unchanged. The merge unifies the two null sentinels as adamic_null with adamic_kind_null; the incoming adamic_box_null spelling is removed. Native union narrowing maps that tag to the concrete null pointer. The filesystem runtime uses a derived host layout for its extra code field; it does not modify the built-in Error representation.

## Conflict resolution

All sixteen textual conflicts were resolved. flow/build.go, ir/ir.go and ir/call_targets_guard_test.go retain readiness, NodeFSFile, NoMove, Defined throws and both call-target reader lists. javascript/javascript.go retains captured-cell readiness and nominal errors. lower/class.go, expression.go and object.go retain accessor/deinitialization checks, enum/Never/void/undefined handling, main's structural views and incoming cause/null handling. lower/exceptions.go retains host MayThrow and nominal exceptions, with runtime range guards before finishClassCalls. lower/library_string.go retains main's receiver coercion and incoming range guards. lower/lower_test.go retains supported main behavior and removes obsolete incoming refusal expectations. native/emit_expressions.go, emit_locals.go and reuse.go retain nominal null handling, main's readiness invariants and both ownership protections. runtime/adamic.h retains environment and null kinds. oracle/adamic.mjs retains immediate panic and correct uncaught normalization. internal/oracle/counts.md was measured afresh on Linux.

No custom edit was made to the four restricted files (internal/native/emit.go, internal/lower/lower.go, internal/native/native.go, internal/oracle/oracle_test.go); their incoming changes remain part of the merge.

## Validation and environment

Linux, nproc 5, four-core CPU quota; Go 1.27.1, Node 24.19.0 and clang 20.1.8. cohere is the batch 3 pointer 7945d102a6c18dd36adf9114a758ce646e8b2359. TypeScript is d92d9bfee114c80be2c375d72edae966176e3a4f. Tests source /workspace/adamic-tools/env.sh and use world-traversable /tmp/adamic-gate.

bash cloud/setup.sh (/tmp/adamic-reland-setup.log): go ready 0.038s; node ready 0.041s; clang ready 0.364s; markdown dependency npm ci step 1.093s, ready 1.189s; submodules 6.424s; Go build 362.993s, cache-warm step 363.349s; done 363.392s. Stage 3's Node typings required npm ci --prefix stage3/api, recorded in /tmp/adamic-reland-node-types.log (about 6s).

The first broad attempts exposed missing Node typings and then a full 32 GiB overlay caused by accumulated Go cache entries. Only old Go cache files were deleted (25.05 GiB, 8,300 files), preserving current entries. Superseded own test process trees were stopped. Subsequent complete gates exposed the real filesystem and runner formatting incompatibilities described above. An environment restart interrupted the remaining native/oracle run; final sources and logs persisted, and those complete packages were restarted after the layout correction. These incomplete attempts are not counted as successful validation.

Final commands (all output redirected to the indicated logs):

```sh
ADAMIC_GATE_UNCACHED=1 go test -p 3 -count=1 -timeout 30m ./internal/ir ./internal/lower ./internal/javascript ./internal/flow ./internal/fresh ./internal/native ./internal/oracle ./stage1/cohere/graphql ./stage3/fixtures ./cmd/adamic-test262 > /tmp/adamic-reland-full-affected-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/lower > /tmp/adamic-reland-lower-final-retry.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/flow > /tmp/adamic-reland-flow-final-retry.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -p 3 -count=1 -timeout 30m ./internal/native ./internal/oracle ./stage1/cohere/graphql ./stage3/fixtures ./cmd/adamic-test262 > /tmp/adamic-reland-remaining-gate-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts > /tmp/adamic-reland-counts-after-layout.log 2>&1
gofmt -l cmd internal > /tmp/adamic-reland-gofmt-final.log 2>&1
go vet ./... > /tmp/adamic-reland-vet-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 python3 cloud/error-classes-reland-status.py > /tmp/adamic-reland-status-audit-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 python3 cloud/error-classes-reland-mutants.py > /tmp/adamic-reland-mutants-final-complete.log 2>&1
```

The broad run recorded lower/flow failures while those early probes read the superseded formatter, plus the stale native layout proof. The complete lower retry passed 196.568s and complete flow retry passed 453.327s after the final handler. The environment restart interrupted the broad run's remaining results. After correcting the layout proof, the final complete remaining gate exited 0: native 421.276s, whole uncached oracle 528.788s, GraphQL 79.782s, stage 3 56.039s and cmd/adamic-test262 182.037s. IR passed 1.185s and fresh passed 173.906s; JavaScript has no package tests and is exercised by the differential oracle. These complete package runs together re-green the full requested scope; no failed package result is counted as green.

Final Linux counts passed 58.834s. A byte comparison against the pre-layout-fix table exited 0. Formatting and all-package vet exited 0 with empty logs. Stage 3's constrained updater passed 79.807s and preserved all records. Thirteen root variants passed the full mutant script before the layout correction; the added fourteenth was run through the same isolated Go-overlay mechanism and caught by the exact runtime-layout assertion, recorded in /tmp/adamic-reland-layout-mutant.log. The checked-in script now includes all fourteen variants.

## Mutants

The repeatable cloud/error-classes-reland-mutants.py uses Go overlays and copied oracle runtimes; it does not alter production files. Each failure must reach its named assertion, not a compiler/build error. The thirteen original final variants passed /tmp/adamic-reland-mutants-final-complete.log; the added layout variant passed /tmp/adamic-reland-layout-mutant.log. Together all fourteen are caught. Per-mutant logs are /tmp/adamic-reland-mutants/<name>.log.

| Mutant | Check that catches it |
| --- | --- |
| host-layout-proof-stale | TestRuntimeFieldLayoutsAreIncluded against the embedded C declaration |
| host-error-identity-removed | Node output for node_fs_file_nominal.a and error_host_uncaught.a |
| host-error-override-accepted | TestHostErrorToStringIsNotYet's named boundary |
| checked-union-return-unproved | TestCheckedUnionReferenceCannotThrow, independent of recorded counts |
| string-receiver-typeerror-not-nominal | error_record_narrowing_identity.a's Node output |
| direct-construction-loses-message-default | coverage_error_optional_messages.a's Node/sanitizer comparison |
| record-typeerror-not-nominal | error_record_narrowing_identity.a's Node output |
| proven-read-still-throws | TestMayThrowPrecision's cannot-throw assertion, independent of counts |
| unused-helper-effects | TestClosedFrameInputRejectsMutation and the recorded Compiles scanner fixture |
| cause-descriptors-required-for-boxing | coverage_error_causes.a's required successful lowering and Node comparison |
| opaque-cause-reflection-accepted | TestOpaqueErrorCauseReflectionIsNotYet's named boundary |
| undefined-panic-class | TestOracleUncaughtExceptionNames, original ReferenceError/TypeError/RangeError and exit 70 |
| synthetic-error-format-bypassed | TestOracleUncaughtSyntheticErrorFormatting, original generated error and exit 70 |
| host-error-format-erased | TestOracleUncaughtHostErrorFormatting, Node's original host error code/text |

The inherited permanent TestErrorClassMutants, TestErrorReportingMutants, TestGeneratedErrorMutants, TestIntegrationCatchabilityMutants, TestRuntimeRangeMutants, TestReaderNullMutants and TestMergedGeneratedErrorMutants remain in the full oracle. These include under-approximated exception propagation, null-tag loss and incorrectly folded comparisons. The new precision assertions catch over-approximation without depending on counts drift.

## Counts and stage 3

[error-classes-reland-counts.md](error-classes-reland-counts.md) gives every moved existing row before/after and its source or analysis cause. Final table: 820 rows, 87 existing changes, 39 additions, no removals. All thirteen previously moved filesystem rows match batch 3 exactly after restoring their catches. The tree fixture improves from 33 allocations/86 retains to 19/48, with releases, peak and regions unchanged. Proven union reads match batch 3 exactly. nbody_runtime_fields.a removes one constructor handoff retain/release.

cloud/error-classes-reland-status.py permits changes only to stage0 outcome and diagnostic text, requires byte-identical recorded/current Node observations, and requires native agreement for Compiles. A formerly Compiles record cannot be downgraded by this updater. It verifies every byte outside stage0 after the run and lists any Compiles-to-Refused/NotYet changes first. The final audit passed in 79.807s (/tmp/adamic-reland-status-update.log): zero Compiles-to-Refused/NotYet changes, zero stage0 changes, and every byte outside stage0 preserved (/tmp/adamic-reland-status-audit-final.log). The 173 entries span 11 status files: 55 Compiles, 50 Refused, 48 NotYet, 19 Checker and one CheckedStop.

The complete test262 corpus and the entire go test ./... tree were not run. The full affected package gate, test262 runner package, stage 3 fixtures, all-package vet and Linux oracle counts are the selected coverage. Opaque dynamic cause reflection and host virtual formatting remain named NotYet boundaries rather than silent wrong code. No Error.captureStackTrace or Error-as-a-value worker changes were merged.
