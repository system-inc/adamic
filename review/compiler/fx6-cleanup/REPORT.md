Built: pinned item 139 diagnostics and cleaned remaining dead code/comments for views fix-forward 4 and 5, items 141 and 153.  
Commit: implementation 915a1f94473ee3c253291e92a963cd33047585b2, on candidate-4 b49c150bd0151c7d78eb42c1a6b3acefe7d95bf1.  
Commands/results: view/cast selections, TestCallTargetReaders, counts, vet and integration lane checks pass; counts.md is unchanged.  
Mutant: blanking ViewWhere fails P17, P48 and P64 only on their source-location assertions; restored diagnostics pass.  
Limits: probes were reconstructed; no full-package/full-repository gate, future lane implementation or runtime layout change.

Item 139 was already fixed on this base: string-key and object-destructuring reads use readViewMember, which writes program.Where(node) into ViewWhere. No diagnostic production code changed. The reconstructed .a probes demand an unsupported ReadonlyMap collection contract through a string key (p17), destructuring (p48) and a literal-typed key (p64). Their CLI refusals print the actual source path and respectively 6:16, 6:9 and 7:16. These are reconstructed witnesses, not claims about the unavailable attachment bytes. The compiler owns these diagnostic locations; the existing view tests supply the Node/native agreement checks.

The final new test leaf times are P17 0.16s, P48 0.14s and P64 0.10s, including each load. Each is a parallel top-level test. The reversible mutant runner changes only the assignment in view_member_read.go to an empty string. All three fail with ': Adamic 0.1 refuses checked view read ... unsupported collection contract', rather than a build failure. Its patch is blank-view-where.patch; mutant.log records every failure and mutant-result.log records the successful catch. The source was restored, verified clean against Git, and diagnostics-restored.log and views-final.log are green.

The base-wide git grep census in dead-before.log proves interfaceConstructions has no caller; interfaceLiteralFields, interfaceWrite and interfaceScalarShape are used only by that dead walker. Those four functions were deleted. viewInterfaceFields has no caller, viewObjectFields is used only by it, viewDataType only calls itself, and lazyReadRefusal has no caller. Those functions were deleted too, including the now-empty view_interfaces.go. The unused mixed-union hook variable was removed; internMixedUnionViewContract remains live through buildUnionReadContract and its tests. dead-after.log has no remaining references to the deleted names.

The redundant family assignments in view_lazy.go were removed: arrays already returned through unionAggregateContract; non-scalar or unavailable unions already returned through the earlier union adapter. A scalar union escaping that adapter contains no phantom intersections, so viewRepresentation and representation agree and the second union condition is false.

The named future lane was inspected without merging it: compiler/views-v4 at 97c69648f71a76492897b22cd3cb63f85947e964 assigns viewArrayContractHook and viewCallableContractHook and calls viewInterfaceType. These three symbols and their strict adapter boundaries are retained. v4-census.log records those references and shows no external V4 callers of the deleted helpers. On this base the array/callable strict hook branches are dormant, and their comments now identify the future attachment points. Dictionary and intersection hooks and the fail-closed viewErasureProof placeholder are conservatively retained for V5/V6. No branches named compiler/views-v5 or compiler/views-v6 were advertised, so their implementations were not inspected. Retaining the intersection hook for future dictionary intersections is a conservative assumption, not an observed assignment.

Comments now describe mixed-union interning through shared dispatch, support checks at contract construction/completion, and the native predicate selecting field-based membership without a common literal discriminant. The interface_cast.go registration comment and native view_unions_mixed.go comments were already corrected on this base and needed no further comment change.

Exactly two runtime-owned files changed, and both changes are comments only:

- internal/native/runtime/view_unions_untagged.h: the plain slot probe returns false for missing/uninitialized slots or unsupported receivers, but true with Unknown for present unsupported storage; snapshots are borrowed.
- internal/native/runtime/object.c: current snapshot callers pass absent=false; the retained branch would return undefined for a missing own slot if a future caller passed true. Required reads use the readiness failure path.

internal/native/runtime/adamic.h is untouched, including its allocation layout and comments. Its unused tail reserves count*sizeof(size_t), eight bytes per slot on this target. Removing it would reduce each requested plain-object allocation by that amount, possibly changing allocator size classes, region capacity and peak storage. The initialized and representation byte-tail offsets would stay the same if only the size formula changed. Allocation/free counts do not necessarily fall when requested sizes fall. This is a runtime-owner decision; no such change or performance measurement was made here.

All commands ran from the repository root after sourcing /workspace/adamic-tools/env.sh, with test output redirected to the named logs. Final verification:

```text
timeout 180 go test ./internal/lower ./internal/native ./internal/javascript -run 'View|MixedUnion|Untagged|TaggedInterface|CheckedCast' -count=1 -v -timeout 90s
lower 6.737s; native 0.686s; javascript 0.369s; exit 0 (views-final.log)
timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s
ir 10.227s; exit 0 (call-target-readers-final.log)
timeout 240 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 90s -args -update-counts
oracle 52.818s; exit 0; counts.md unchanged (counts.log)
timeout 120 go vet ./internal/lower ./internal/native ./internal/javascript
exit 0, no output (vet.log)
timeout 180 python3 review/compiler/fx6-cleanup/run-location-mutant.py
blank ViewWhere caught by P17, P48 and P64 source-location assertions; exit 0
timeout 180 go build -o /tmp/fx6-adamic ./cmd/adamic
exit 0 (cli-build.log)
/tmp/fx6-adamic c internal/lower/testdata/view_diagnostics/oct9_views_p17.a
/tmp/fx6-adamic c internal/lower/testdata/view_diagnostics/oct9_views_p48.a
/tmp/fx6-adamic c internal/lower/testdata/view_diagnostics/oct9_views_p64.a
all exit 1 with the exact source locations above (cli-p17.log, cli-p48.log, cli-p64.log)
git diff --check
exit 0
```

Earlier green selections were 'View|MixedUnion|Untagged' across the same three packages (11.670s/16.094s/0.690s, views.log), and 'TaggedInterface|CheckedCast' in lower/native (0.737s/0.363s, interface-tests.log). Focused diagnostics ran with '^TestViewDiagnosticP', -count=1 -v -timeout 90s (0.156s before mutation; 0.218s after restoration). The first TestCallTargetReaders run passed in 24.229s. These were selected tests, not whole-package runs.

Integration's required command was run after committing:

```text
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
lane checks 10.2 s: gofmt and tools on 52 Go files, t.Parallel on 3 test packages; a-check 2 .a files; vet 3 packages
```

Fetches and the lane-checker process had hard shell limits. This checkout initially tracked only main, so candidate-4, devtools/fast-gate and cloud/merge-tree were explicitly fetched into their remote-tracking refs. The branch was created from the unit-specific candidate-4 tip, overriding the generic start-from-main instruction. No other worker branch was merged. The lane checker includes inherited candidate changes in its 52-file census.

Setup used GOPROXY='https://proxy.golang.org|direct' and timeout 600 bash cloud/setup.sh. setup.log records: Go ready 0.062s, Node ready 0.078s, clang ready 0.396s, markdown dependencies ready 1.017s, submodules ready 125.357s, Go build ready 376.504s, test binaries deferred 376.608s, build cache warm 376.609s, done 376.641s. nproc=5; cpu.max=400000 100000 (four CPU quota). Go 1.27.1, clang 20.1.8 and Node 24.19.0 were used.

Setup succeeded. An early test attempt before submodule restoration could not load cohere/TypeScript/tsc/go.mod, and a cold compile attempt hit its timeout 120 shell limit without test results. Initial dictionary witnesses stopped at the pre-existing index-signature refusal, so the final witnesses use a collection obligation to reach lazy read diagnostics. The first full counts refresh failed on missing @types/node 25.3.3 (75.816s; counts-missing-node-types.log). timeout 120 npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund installed the existing lockfile's three packages (node-types-setup.log), after which counts passed. No dependency manifest changed.

This contributes the remaining diagnostic regression guards and conservative cleanup toward the requested views fix-forward 4/5 work. It does not implement V4 callables, V5 dictionaries, V6 erasure, inspect unnamed future lanes, change runtime executable code, or run the shared fast gate locally. Pure deletions and comment edits have no additional mutant.
