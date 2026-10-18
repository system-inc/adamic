Re-cut slice 5 onto newest main, preserving all main readers and modern JSON behavior.
Base ca0d1e1500d5ac663b62e196b0b68e3b4f7047be; source 9947f26139e9a1fbbbff2d29949b0d9490b35a0b; delivery SHA is in the handoff.
Current inventory: native 485/485 and oracle 859/859 top-level tests have passing or built-in skip results under 60 s; all eight touched suites were invoked.
Stale-reader, borrowed-record guard, missing baseline product and wrong outcome pin mutants were caught; array-extras mutant remains a Node-only semantic catch.
Not covered: full repository/Loom gate, optional WASI/benchmarks and pending Program-region lifetime; no ahra or submission.

Conflict details and main intent:

1. internal/fresh/corpus_units_test.go, thirteen function-name hunks: main regenerated 859 selectable leaves and shifted indices. Keep that entire main corpus and its indices; migrate only the 13 old refusal paths and canonical function names (fresh-resolution.json names every old/new pair). Auto-merged path/binding arrays are reconstructed consistently from main. No reader is deleted: 859 before and after, compared with the old source 828. Main newer exclusion-aware generator remains, with source fresh-only moved-control globs. The coverage guard passes.
2. internal/native/runtime/json_stringify.c, first hunk: main a77cc07a added modern typed descriptor/array providers, key-aware toJSON, exception handling, owned-result release, cycles and borrowed record views. Keep those functions and their five-argument write_field interface. Add source omission only after validating descriptor/slot bounds, and do not inspect construction metadata when object->shape is NULL. The main record-view sanitizer test passes; removing the shape guard is caught by UBSan.
3. json_stringify.c, second hunk: retain descriptor order for ordinary and unshaped record objects. For construction objects, enumerate adamic_object_ordered and call the modern writer with slots. This replaces main direct descriptor loop only when construction order exists. Main newer checks/providers/cleanup are retained; no old four-argument writer is restored.

Counts were not hand-merged. Restore newest main table, then use Linux house clang 20.1.8. Full go test ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -timeout 10m -json -args -update-counts passed in 184.122 s, moving exactly the source original 62 keys (41 added, 21 modified). The full suite then identified main NodeArray boundary as newly accepted. Its retained reader now requires Node/backend agreement; an additional narrowly selected official updater invocation measured that one row. Final diff is 63 keys: 42 added, 21 modified. All 62 original keys remain this source change own; the only additional key is main newly admitted NodeArray fixture. Every row is listed below and in counts-rows.json. No Mac measurements.

Additional guards and consumer repairs found by full suites:

- construction.c contains only functions/local and per-object storage; add its runtime-file audit marker in testdata/runtime-statics/storage.txt. TestRuntimeStaticsAreListed passes, without relaxing its scanner.
- Source adds object helpers reached by regex reduced runtime. Add construction.c and library_object.c to regexWitnessRuntimeSources, keeping the complete transitive link closure. All 81 initially failing native top-level tests pass after repair in four bounded batches; no case is dropped.
- The existing array-extras runtime mutant injected an old four-argument write_value call; adapt it to main keyed API with NULL for this primitive-number probe. It now compiles, exits cleanly under sanitizers and disagrees with Node, rather than dying at clang.
- TSGo Unit00 initially took 81.04 s: 57.638 s filling source/baseline products and 23.28 s actual behavior. Separate that work into existing TestProduct_ units and forbid baseline filling in behavior. A missing-product mutant fails quickly with the explicit preparation error. Final source and baseline preparation and both cache modes pass under 60 s; final baseline is 19.517 s, final split test command 12.028 s. The unchanged over-budget workload was not repeated.
- Moves ASan producer initially took 66.89 s in one C translation unit. Use main split builder with four jobs for this variant and an explicit split recipe flag, keeping every source statement and runtime mode. The splitter safely refused emitter parallel prototypes; retain only the two known external parallel-map ABI prototypes in the shared header without renaming them. Other unknown declarations still refuse. Split and Jobs are program-only options, so normalize them out before asking main runtime supplier for the existing archive. Cold split producer passes in 20.64 s; Node, both thread modes and leak check pass in 21.41 s. Other variants are unchanged.
- Main NodeArray indexing witness now lowers. Keep its existing boundary reader, require source Node/JS/native sanitized/native release/leak agreement and register it as lowering in the generic oracle family. Final boundary test, complete native shard 113 and counts table pass.
- Main review/refused/non_null.a now lowers by the source literal-placeholder rule. Its body is byte-identical after moving to agree/. Source Node actually exits 0 and prints undefined. Both backends deliberately check declared-number consumption before String argument use and exit 70 with the same diagnostic. Assumption: preserve that source checked-use contract #9wc5q5j; use main explicit .intended/.expected.json mechanism. The pin is stdout empty, stderr "adamic: panic: placeholder 'value' is unset at argument via value\n", exit 70. Source placeholderAllowsUnset explicitly is narrower than acceptsUndefined even when Node tolerates a use. This is not reported as raw Node agreement. Review reader also requires the pin to differ from actual Node, so a stale pin fails. Wrong exit 71 is caught independently by native, JS and release comparisons.

All current tests and exact commands:

Each build/test shell sources /workspace/adamic-tools/env.sh, unsets GOCACHEPROG, uses GOMAXPROCS=4 and GOFLAGS including -p=4. Test output goes directly to JSONL logs. All commands have outer hard limits. Full package invocations use -count=1 -parallel=4 and -json.

- go test ./cmd/adamic/...: initial runtime products were not prepared ahead, and subprocess temporary XDG caches hid them. Prepare TestProduct_CommandRuntime first, set ADAMIC_BUILD_CACHE_DIR to the existing shared product root, then rerun complete suite: exit 0, 21.345 s. This is invocation setup, no weakened CLI test.
- go test ./internal/flow/..., ./internal/fresh/..., ./internal/ir/..., ./internal/javascript/... and ./internal/lower/... (-timeout 10m): complete suites exit 0.
- go test ./internal/native/... (-timeout 14m): complete suite reached every top-level test, 810.300 s. Initial declaration/link and outdated-mutant failures are retained. All 81 failed top-level tests pass in repaired batches; final split path tests validate the production extensions and both TSGo modes. Current inventory audit proves all 485 top-level results pass/skip, none missing or over 60 s.
- go test ./internal/oracle/... (-timeout 14m): package deadline after 840.076 s; 16 unfinished tests all pass in a bounded batch. Initial non_null/NodeArray classification failures are retained and corrected through stronger agreement/pinned-check readers. Final current inventory is all 859 top-level results pass/skip, none missing or over 60 s.
- go test ./internal/oracle -run ^TestReviewPrograms -count=1 -parallel=4 -timeout 150s -json: exit 0, 38.835 s, including the explicit intended difference.
- go test ./internal/oracle -run "^(TestElementAccessCompilerAreaBoundaries|TestNativeAgreesWithNode_113|TestCountsAreRecordedTable|TestRuntimeSetsAreEachUsed|TestMovesObjectsUseDedicatedProducts|TestMovesSourceKeyNamesCompiler)$" -count=1 -parallel=4 -timeout 120s -json: exit 0, 6.472 s.
- Final TestCallTargetReaders (-timeout 90s): pass, 13.71 s. Full ir also passes TestEveryMapAProgramHoldsIsSorted; full lower passes TestNativeAgreementRowsNameRegisteredPrograms; runtime sets and final cmd/adamic-gate TestEveryTestIsParallelOrSaysWhy pass. No new Go package needs a dependency declaration.
- go vet on all eight touched packages recursively: final exit 0. gofmt -l all changed Go files: no output.
- Pinned unchanged Gate.aCheck rebuilt final compiler and accepts 59/59 programs: 43 checked, 16 refused, 0 NotYet, 0 type errors. Source 58 files retained plus moved non_null seed. No new first-line header was necessary.

validation.json stores every per-log failure and repair, final inventory ownership and current status; test-seconds.json keeps all individual event times. It does not hide the initial 81.04/66.89 observations. Optional WASI, long benchmark and existing pending lifetime hooks skip through their own guards. Full gate was not run.

Mutant evidence:

- stale-reader overlay: restore one old missing refusal path. Coverage guard rejects it, exit 1.
- borrowed-record guard overlay: compile production writer with shape guard removed in main record harness. UBSan rejects null shape access; no clang-warning kill credit. Ordinary record path passes.
- missing-tsgo-product overlay: alter only baseline product name. Behavior fails in 0.643 s with must be built ahead error, rather than compiling the baseline.
- wrong outcome pin: temporarily change expected exit 70 to 71. All three backend comparisons reject it; original JSON bytes restored in finally.
- adapted original array-extras mutant: valid keyed-writer call reaches runtime and produces Node stdout disagreement with clean sanitizer/leak execution. Source in-suite mutants ran in complete suites. Historical standalone Python mutant scripts were not all repeated.

Exact runtime C/header clearance beyond source own signed lines:

Only internal/native/runtime/json_stringify.c has extra patch lines. Unchanged main lines are not counted; runtime-clearance.json compares signed line multisets against git diff 9947f261~1 9947f261. All other applied runtime C/header +/- lines match source own lines. The seven extra signed lines are:

```diff
+    // Borrowed record views have no shape or construction metadata.
+    // Reserved and unset shaped slots are omitted before reading their storage.
+    if (object->shape != NULL && (!adamic_object_present(object, field->slot) || !adamic_object_initialized(object)[field->slot])) { return; }
-			for (size_t index = 0; index < schema->count; index++) { write_field(w, object, &schema->fields[index], slots, depth, &written); }
+			if (object->shape == NULL || object->write_order == NULL) {
+				for (size_t index = 0; index < schema->count; index++) { write_field(w, object, &schema->fields[index], slots, depth, &written); }
+						if (schema->fields[index].slot == order[position]) { write_field(w, object, &schema->fields[index], slots, depth, &written); break; }
```

Full runtime-final.patch and json-resolution.patch retain context. The guard adds a shape check for main borrowed records; the order branch keeps main slots argument and unshaped descriptor order. These exact changes need runtime clearance.

Every changed counts row, before -> after A/F/R/L/P/G:

- `internal/oracle/testdata/4ddd17f_weak_single_narrowed.a`: absent -> 4/2/4/6/4/0
- `internal/oracle/testdata/backend_stop_catches.a`: absent -> 2/2/6/4/1/0
- `internal/oracle/testdata/d96d304_try_stack.a`: absent -> 0/0/0/0/0/0
- `internal/oracle/testdata/non_null_initialized.ts`: 0/0/0/0/0/0 -> 55/55/50/109/18/0
- `internal/oracle/testdata/non_null_literal_assignment.ts`: 0/0/0/0/0/0 -> 1/1/2/2/1/0
- `internal/oracle/testdata/non_null_static_initialized.ts`: 1/0/0/1/1/0 -> 9/9/12/22/6/0
- `internal/oracle/testdata/non_null_static_uninitialized.ts`: 1/0/0/1/1/0 -> 1/0/3/2/1/0
- `internal/oracle/testdata/non_null_uninitialized_capture.ts`: 0/0/0/0/0/0 -> 2/0/2/0/2/0
- `internal/oracle/testdata/non_null_uninitialized_const.ts`: 0/0/0/0/0/0 -> 0/0/2/1/0/0
- `internal/oracle/testdata/non_null_uninitialized_exception.ts`: 0/0/0/0/0/0 -> 1/1/3/2/1/0
- `internal/oracle/testdata/non_null_uninitialized_field.ts`: 1/0/0/0/1/0 -> 1/0/3/2/1/0
- `internal/oracle/testdata/non_null_uninitialized_interface.ts`: 1/0/1/0/1/0 -> 1/0/2/1/1/0
- `internal/oracle/testdata/non_null_uninitialized_iteration.ts`: 1/0/0/1/1/0 -> 6/1/10/8/6/0
- `internal/oracle/testdata/non_null_uninitialized_loop.ts`: 0/0/0/0/0/0 -> 0/0/1/0/0/0
- `internal/oracle/testdata/non_null_uninitialized_map_entry.ts`: 0/0/0/0/0/0 -> 0/0/1/1/0/0
- `internal/oracle/testdata/non_null_uninitialized_optional.ts`: 1/0/0/0/1/0 -> 3/3/2/6/3/0
- `internal/oracle/testdata/non_null_uninitialized_spread.ts`: 1/0/0/0/1/0 -> 2/0/5/5/2/0
- `internal/oracle/testdata/notyet_element_access/nodearray.a`: absent -> 0/0/0/0/0/0
- `internal/oracle/testdata/placeholder_nonnull_alias_reset.a`: absent -> 2/1/6/7/2/0
- `internal/oracle/testdata/placeholder_nonnull_assignment_result.a`: absent -> 0/0/2/3/0/0
- `internal/oracle/testdata/placeholder_nonnull_before_use.a`: absent -> 0/0/1/1/0/0
- `internal/oracle/testdata/placeholder_nonnull_field.a`: absent -> 9/9/15/23/5/0
- `internal/oracle/testdata/placeholder_nonnull_local.a`: absent -> 12/12/11/24/4/0
- `internal/oracle/testdata/placeholder_nonnull_null_before_use.a`: absent -> 0/0/2/1/0/0
- `internal/oracle/testdata/placeholder_nonnull_null_coalesce.a`: absent -> 8/8/15/27/4/0
- `internal/oracle/testdata/placeholder_nonnull_null_copy.a`: absent -> 26/26/102/121/7/0
- `internal/oracle/testdata/placeholder_nonnull_null_equality.a`: absent -> 4/4/10/16/2/0
- `internal/oracle/testdata/placeholder_nonnull_null_json.a`: absent -> 5/5/0/6/3/0
- `internal/oracle/testdata/placeholder_nonnull_null_loose.a`: absent -> 10/10/23/35/4/0
- `internal/oracle/testdata/placeholder_nonnull_null_optional.a`: absent -> 6/6/16/25/3/0
- `internal/oracle/testdata/placeholder_nonnull_null_saved_leak.a`: absent -> 1/0/4/4/1/0
- `internal/oracle/testdata/placeholder_nonnull_null_truthiness.a`: absent -> 1/1/7/10/1/0
- `internal/oracle/testdata/placeholder_nonnull_null_typeof.a`: absent -> 3/3/9/11/3/0
- `internal/oracle/testdata/placeholder_nonnull_observe.a`: absent -> 27/27/48/79/9/0
- `internal/oracle/testdata/placeholder_nonnull_return_assignment.a`: absent -> 0/0/2/3/0/0
- `internal/oracle/testdata/placeholder_nonnull_saved_leak.a`: absent -> 1/0/3/4/1/0
- `internal/oracle/testdata/placeholder_nonnull_scanner.a`: absent -> 4/4/9/8/3/0
- `stage3/interface-downcasts/default-boxed-write.ts`: 0/0/0/0/0/0 -> 5/5/5/12/4/0
- `stage3/interface-downcasts/default-read-before-set.ts`: 0/0/0/0/0/0 -> 1/0/4/2/1/0
- `stage3/interface-downcasts/default-staged.ts`: 0/0/0/0/0/0 -> 2/2/6/7/2/0
- `stage3/interface-downcasts/readiness-identifier-uninitialized.ts`: 0/0/0/0/0/0 -> 1/0/4/3/1/0
- `stage3/interface-downcasts/readiness-identifier.ts`: 0/0/0/0/0/0 -> 2/2/6/8/2/0
- `stage3/interface-downcasts/readiness-number-uninitialized.ts`: 0/0/0/0/0/0 -> 1/0/5/3/1/0
- `stage3/interface-downcasts/readiness-number.ts`: 0/0/0/0/0/0 -> 5/5/5/10/5/0
- `stage3/parser-ahead/factories/already-complete.a`: absent -> 1/1/4/4/1/0
- `stage3/parser-ahead/factories/complete.a`: absent -> 1/1/5/4/1/0
- `stage3/parser-ahead/factories/escaped.a`: absent -> 1/1/3/4/1/0
- `stage3/parser-ahead/node-array/fields.a`: absent -> 12/12/10/23/12/0
- `stage3/parser-ahead/rulings/factory-unset.a`: absent -> 3/3/13/20/3/0
- `stage3/parser-ahead/rulings/factory-use.a`: absent -> 1/0/2/3/1/0
- `stage3/parser-ahead/rulings/node-array-json.a`: absent -> 4/4/5/11/2/0
- `stage3/parser-ahead/rulings/node-array-keys.a`: absent -> 15/15/12/16/9/0
- `stage3/parser-ahead/rulings/node-array-length.a`: absent -> 4/4/4/9/2/0
- `stage3/parser-ahead/rulings/node-array-unset.a`: absent -> 4/4/18/29/4/0
- `stage3/parser-ahead/rulings/null-placeholder-pending.a`: absent -> 2/2/4/8/2/0
- `stage3/parser-ahead/rulings/own-key-order.a`: absent -> 28/28/17/37/10/0
- `stage3/parser-next/speculation/arrow-tryparse.a`: absent -> 3/3/8/11/2/0
- `stage3/parser-next/speculation/misfit.a`: absent -> 1/0/2/1/1/0
- `stage3/parser-next/speculation/named-lookahead.a`: absent -> 5/5/1/7/4/0
- `stage3/parser-next/speculation/named-tryparse.a`: absent -> 3/3/10/16/3/0
- `stage3/parser-next/speculation/truthiness-rewind.a`: absent -> 34/34/46/85/6/0
- `stage3/parser-next/stack/benchmark.a`: absent -> 7/7/0/17200008/4/0
- `stage3/parser-next/stack/legal.a`: absent -> 7/7/0/1728/4/0

Toolchain/setup: shared unit setup completed on slice 4 in 45.121 s after resolving an initial conflict-marker overlap. Linux Go 1.27.1, Node 24.19.0, house clang 20.1.8; nproc=5, cgroup four-CPU quota. Four retired Go-export products were reclaimed with exact paths recorded in cache-reclamation.json; current products and sources preserved. The environment restarted after validations; workspace/logs persisted. Recreated its missing /tmp/adamic-gate (1777) and finished inventory-only audits. No PR, force push, main push or Loom submission.
