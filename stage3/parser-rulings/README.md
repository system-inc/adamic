# Parser rulings: factory completion

This is the conservative fixed-literal surface of roadmap step 24, rule 1. The three staged-cast reductions from `codex/step24-factories` at `3027b839` now compile and match source Node in both backends. Factory names are not trusted. Fresh local object literals reserve their declared fixed fields with unreadable slots, then ordinary assignments establish readiness. Reads of a deferred field stop with exit 70 and name the field and source expression. A required field whose type admits undefined may hold an initialized undefined value.

The completion summary intersects all branch paths and all escape snapshots. Direct aliases share one allocation root. Return, store and passing a reference are escape boundaries. Unknown control, calls, closures, captured roots and opaque allocator construction retain checks rather than grant completion. The existing readiness CFG uses these summaries for direct factory calls and their local results. The IR allocation remains the ordinary fixed object allocation, with each missing field marked Uninitialized; ownership, reuse and regions see that allocation rather than separate field allocations.

This does not yet implement the exhaustive interprocedural proof required for the full node factory. In particular, a generic or patched allocator's returned object is not certified merely from its asserted subtype. Calls into allocation helpers and update factories need ownership-aware effect summaries before their syntax fields can be certified. Optional staged fields are refused pending step 17's presence representation (#z00sxvc). Presence observers are also refused while reserved absent slots lack observable presence bits. An existing object assertion remains subject to the original cast refusal.

## Acceptance table and observations

[factory-pairs.tsv](factory-pairs.tsv) has one row for every required runtime factory/field pair in the supplied 492-factory population. The acceptance target is checked for the 15 deferred names (745 pairs) and proven for other names (3,692 pairs). All 4,437 runtime pairs are currently refused before lowering; 1,476 phantom-brand pairs are excluded. The target is separate from both the scout's evidence status and actual compiler observation. The 428 explicit undefined pairs are a subset of the pairs involving those names, not the total number of checked targets. Phantom brands are counted separately and are not runtime fields. [factory-summary.json](factory-summary.json) pins the input bytes and records all counts; [factory_table.py](factory_table.py) regenerates both artifacts.

The full parser driver refuses during Load at `src/compiler/builder.ts:1246:69`: `Path | undefined` is not assignable to `string`. This is the first stop both before and after this change, so every real-source pair remains recorded as refused before factory lowering. No pair is promoted from source evidence to observed native proof. The reduced programs below provide the measured completion and checked-read surface.

| Reduction | Proven fields at factory return | Checked fields at factory return |
| --- | --- | --- |
| createBinaryExpression | kind, left, operatorToken, right | none |
| createVariableDeclaration | kind, name, initializer | none |
| createNumericLiteral | kind, text, transformFlags | none |
| createSourceFile | kind | bindDiagnostics |

A caller's later diagnostics assignment makes that caller's read ready. Reading before that assignment stops loudly. Source Node prints `before` then `undefined`; the checked native and JavaScript backends print `before` then stop with exit 70. This is an intentional ruling check, not a Node output match.

## Focused verification

All command output is redirected to logs. No whole package or full gate was run.

```
go test ./internal/lower -run 'TestParserFactoryCertificates|TestStagedFactoryRefusals|TestFactoryCompletionBeforeEscape|TestUncheckableCastsStayRefused|TestCheckedCastProofAndElision|TestReadinessElisionRequiresDominatingAssignment' -count=1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/parser_factory_|TestParserFactoryDeferredRead' -count=1 -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
```

The oracle covers four successful source Node comparisons, both backends, native ASan/UBSan and leak checks, plus the intentional deferred-read stop. Independent mutants union branch readiness, ignore store escape, remove initializer compatibility, allow optional fields, allow presence observers and remove the deferred-read check. Their focused tests fail on the intended proof/refusal/output assertions. The read-check mutant executes successfully and incorrectly prints `before\n0\n`, demonstrating a wrong-output catcher rather than a compiler or sanitizer accident.

Counts add the five parser factory rows. The existing logical_and_reference_maybe row moves without changing its numbers. The stale taste/17_binder_flow row disappears because its fixture was already marked non-lowering on this base; this change does not alter its classification. Allocation counts for the loud-stop fixture include its live allocation at process termination and do not imply normal-exit cleanup.

Setup passed with Go 1.27.1, Node 24.19.0 and clang 20.1.8. Cumulative timing lines: Go 0.028s, Node 0.031s, submodules 0.071s, markdown 0.090s, clang 0.189s, build 42.857s, deferred test binaries 43.106s, warm 43.107s, done 43.134s. nproc is 5.

Rules 2 through 5 are not implemented by this change. The full factory's interprocedural acceptance targets remain pending; this is not a claim that the parser's five rulings are complete.
