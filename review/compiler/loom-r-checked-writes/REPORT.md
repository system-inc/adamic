Rebased checked writes through a wider view as one commit on compiler/loom-r-checked-writes.
Source: 196a08b46732c1de3a991893c10b5c5c651965f3. Base: dbdf977f5b8b0979d414ebe6f20e7aadd312e309.
193 focused roots / 203 events pass; flow corpus 854 leaves pass; counts, guards, consumer checks and vet pass.
All 65 recorded mutants caught: 19 standalone overlays, 45 embedded witnesses, one planted count row.
Not covered: whole-package/full gate, unrelated consumer suites, WASI, other operating systems, census or benchmark campaigns. No ahra, Loom or PR submission.

The delivery SHA is the pushed branch head, supplied in the final response. git fetch origin was followed by checkout of this exact pinned base, then git cherry-pick --no-commit of the source. Only its parent-to-commit delta was applied. No cohere source was copied. Existing source TypeScript probes were preserved; newly authored Adamic programs were not introduced. Conservative choice: the refused .a adapter witnesses belong in *_refused directories, while their original refusals remain enforced.

Conflicted hunks, each resolved once:

1. internal/ir/ir.go Program tail: retain main's Async *AsyncProgram and add CheckedWrites / CheckedElements / WriteChecks / WriteContracts / ContractTypeNames. The maps are SortedMap. Main f93a65fb introduced async and parallel lowering infrastructure.
2. internal/ir/ir.go expression tail: retain main's ParallelMap.Type() returning Array and add the source ContractContainer expression and Type() method. Both expression families remain registered and usable.
3. internal/native/emit_statements.go field metadata update: retain main's adamic_slot_index(object, slot), and retain the source's condition that a checked Boolean/MaybeBoolean store preserves the actual slot representation. Main fcf5c2ad packed shared slot caches into an atomic word; metadata must use the returned slot, not a mutable cache index. Also adapt the source-added boolean-store read and checked result-contract emission to this same API.
4. internal/oracle/counts.md has one tail insertion conflict: main adds the five instantiation_{find,first,last,callback,array}.a rows while the source appends checked-write rows. Discard the conflict wholesale, restore pinned main's entire table, then regenerate once. All five main instantiation rows are retained. counts-conflicted-hunks.txt records the complete hunk; the two Go conflict snapshots retain every marker. Main's newer count-product infrastructure and source-name normalization are unchanged.

Additional integration:

- Main's new async host result shape and five parallel C harnesses, plus prototypeOwnershipModel, need the new contracts pointer initialized to NULL. Their assertions and behavior are unchanged. The parallel and prototype consumers pass, including ASan/UBSan, counted/release variants, TSan and the field-cache / oversized-slot mutants.
- Seven source-added runtime lines in object.c use main's slot-derived indexing or zero initializer; the maybe-boolean read uses field_owned's actual owner. One additional main line in async_host_impl.h initializes contracts to NULL. EVERY before/after line is listed in runtime-line-clearance.md. All other source own runtime +/- lines are identical. source-runtime.patch, rebased-runtime.patch, runtime-comparison.json and runtime-delta-differences.patch preserve the complete comparison. These eight runtime line adaptations need the requested runtime clearance.
- Move checked-runtime-fields and checked-write-messages to checked-runtime-fields_refused and checked-write-messages_refused. Update their adapters/count keys and the three runtime input path literals. No refusal is relaxed and no test is deleted. The flow corpus is unchanged and passes.

Counts:

Exactly one -update-counts invocation on Linux / house clang 20.1.8:
  timeout 600 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -json -timeout=570s -args -update-counts
Result: PASS, package 203.835s; 1179 leaves, slowest 31.94s. Main's table format is retained; count digests live in measurement-product keys, not a hand-merged column.
92 changed existing rows, 120 new rows, zero removed. counts-row-changes.json enumerates EVERY row with six before/after counters and its attribution. After the two refused-directory key renames, 91 existing-row deltas and all 120 new-row values exactly match the source. Their existing rows add 230570 retain/release pairs. The sole additional existing row is main's bench/parallel_files.a:
  main   3952651 | 3952651 | 8151045  | 8654930  | 8839 | 0
  branch 3952651 | 3952651 | 11034629 | 11538514 | 8839 | 0
That adds 2883584 retain/release pairs. A direct overlay removing ONLY ArrayPush and spliceArguments receiver holds restores all six main counters exactly (counts-bench-direct.log, PASS 4.35s). No other existing-row counter changes. The table guard passes afterward. Regeneration was never repeated.

Guards and test grain:

- internal/ir TestEveryMapAProgramHoldsIsSorted PASS. Source's SortedMaps retained; both plain-map mutants fail the map/gob checks.
- internal/ir TestCallTargetReaders PASS 44.03s with -timeout=90s. Its native/lower/fresh/flow/leakcheck reader analysis passes.
- internal/lower TestNativeAgreementRowsNameRegisteredPrograms PASS. Literal registry remains valid. Main async and parallel lowering regressions pass too.
- internal/oracle TestRuntimeSetsAreEachUsed PASS 4.26s. No new runtime feature set, witness or compiler-consuming package is needed. Existing cloud/fast-gate/compiler-dependencies.json declarations cover the consumers.
- cmd/adamic-gate TestEveryTestIsParallelOrSaysWhy PASS 1.03s. Existing explicit driver exception is preserved; added tests call t.Parallel first.
- Oracle TestFamiliesOwnEveryCaseOnce, TestFamilyWrappersAreCurrent and TestCountsAreRecordedTable PASS.
- All 193 source-focused roots pass in 14 bounded commands, 203 events, maximum 24.38s. test-seconds.json gives every event's package, name and seconds; focused-tests.json gives the exact inventory. Node decides outputs; source .a refusals, TypeScript adapters, native sanitized/release output, JS output, leak/count balance and exact diagnostics are checked by the source's own tests.
- Flow: timeout 600 go test ./internal/flow -run '^TestFlow(Program|Corpus|MutationRanges_|GraphPaths_|Liveness_)' -count=1 -parallel=4 -json -timeout=570s. PASS 124.050s, 854 leaves, max 23.03s.
- Focused consumer checks pass in fresh, javascript, regexp, leakcheck, native, oracle, compilerinputs, fuzz, stage1/typescript/{parser,scanner}, stage3/fixtures and stage3/census/latent/statecopy. Main async Node agreements and native parallel checks pass. Each executed leaf is under 60s; the slowest completed supplementary leaf is 29.00s. Source code is unmodified after vet.
- Remaining direct consumers compile with -run '^$': bench/regex, bridge/tsgo, cloud/reports/decode-ascii, cmd/adamic-meter, cmd/adamic-test262, internal/compilerinputs, internal/lower/cycleledger, internal/native/runtimeproducts, stage3/census/tool, stage3/drivers/parser/probe. Three build-ignored standalone utilities compile by explicit file: stage3/drivers/parser/build-metrics/main.go, stage3/drivers/scanner/build-metrics.go and internal/native/testdata/parallel/build_measure.go. This is compilation coverage, not a claim that their unrelated suites ran.
- timeout 180 go vet ./cmd/adamic ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle ./internal/regexp ./stage1/cohere/values ./stage3/fixtures: exit 0.

Recorded mutants:

Each standalone overlay fails with its intended assertion, without compile errors or a test timeout. mutants.json and mutant-results.json list every mutation, exact selector, command, exit, catcher and elapsed wall time. The longest standalone invocation is 68.532s including Go compilation; its failing test is 32.62s. The 45 embedded mutants and their successful catchers are individually listed in embedded-mutants.json and focused-results.json. They cover erased write checks, literal sets, allocation/nested contracts, declaration messages and representation evidence. The planted row (number-fit.ts) claims 7 allocations instead of 6; TestCountsAreRecorded_026/stage3/checked-writes/number-fit.ts rejects both the measurement product and the fresh measurement (counts-mutant.log, exit 1, leaf 0.28s). Total: 65/65 caught. The two additional main cache mutants also pass their catchers; they are not included in 65.

Setup and corrected verification attempts:

export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh; source /workspace/adamic-tools/env.sh; export GOMAXPROCS=4. Linux x86_64, nproc=5, cpu.max=400000 100000 (four CPUs), Go 1.27.1, Node 24.19.0, clang 20.1.8.
Successful setup timing lines: node 0.021s, go 0.023s, submodules 0.056s, markdown dependencies 0.064s (step 0.007s), clang 0.133s, shared cache 0.700s, go build 37.727s, test binaries deferred 37.814s, build cache warm 37.815s, done 37.841s. Full timing/toolchain lines are retained in setup-resolved.log and toolchain.log.

The first setup overlapped the cherry-pick and its warming build saw unresolved conflict markers; it failed, then setup succeeded after resolution. The first driver check correctly refused main's unprepared command runtime archive; TestProduct_CommandRuntimeSanitized exposed the C cache/shape integration errors. After those fixes its product passes, and the driver caught the remaining generated result-contract cache index, which was fixed before the final focused pass. Recorded runtime mutant overlays were re-cut onto main and corrected for the same cache API so compile failures are not counted as mutant catches.

A supplementary native selection was too broad for a single 90s package process. Its completed leaves all passed (max 29s), but it timed out with field_cache, oversized_slot and only nested_exception/tsan unfinished. Only those unfinished checks were run separately; each passes. Completed tests, including existing non-leaf parents over 60s, were not rerun. No production test was weakened or deleted. The initially broad benchmark attribution probe was stopped during count-product preparation and replaced by the single direct measurement above. Initial reader compilation selected three build-ignored utility directories; compiling their explicit Go files succeeds. All initial failures and final results are reported, rather than silently treated as passes.

Exact programmatic command lists are in commands.json and the three runner scripts. Additional fixed commands and bounds are given above; guard logs preserve their test names and outputs. All evidence belongs here; overlay Go sources end in .go.txt. Every test output was redirected to its log, never piped. No full package suite or full gate was run.

Production diff whitespace checks pass (internal, cmd, stage1, stage3, cloud). A whole-staged-tree whitespace check reports intentionally preserved conflict markers and patch formatting in review evidence, including inherited source evidence; these are text artifacts, not compilable sources. Generated C names the entry basename (lower.go), and main's placeless product infrastructure is retained.
