# Hidden boundary recut guards

Merged current main `bfe055330` in `5e3a8fcad`; no conflicts. This includes main's call-target reader table from PR #250.

## Call-target reader

`overloadSpecialization` used `CallClosure.Closure` both to inspect a possible implementation and to pass the original closure into the checked entry. Implementation analysis now uses `ClosureTargets`, requires a known single target, and keeps an indirect-call stop for unknown targets. The call implementation and signature lookup use `CallTargets`; multiple implementation targets retain an explicit NotYet stop. The remaining two closure-expression reads name the closure to invoke, preserving its identity and captures. Their allowlist entry is owned by `compiler`, with reason: "passes the original closure as the callee of the specialized overload entry".

The first reader run exposed two further `Call.Function` reads in the specialization key and parameter-layout lookup. Routing those through `CallTargets` made the complete reader census pass. Removing the new closure entry through a Go overlay fails `TestCallTargetReaders` on exactly the intended `CallClosure.Closure` read. The unmodified reader test passes again afterward. Mutant source is non-compilable `.go.txt` evidence.

## Closed cycle gaps

Before changing records, `TestFixturesCycles` ran source Node and the native oracle hook. Both newly compiling programs passed the native comparison; only their stale stage0 records failed. The guarded updater was then run with `-args -update`; it updates only when current and recorded Node agree and native agrees with Node. No source changed.

| Fixture | Old stage0 | New stage0 | Node and native behavior |
| --- | --- | --- | --- |
| 05_safe_load_read/main.a | NotYet: array of never | Compiles | stdout `0 0 true\n`, empty stderr, exit 0 |
| 06_import_order_mutant/main.a | NotYet: array of never | Compiles | empty stdout, stderr `adamic: panic: ReferenceError: Cannot access 'emptyArray' before initialization\n`, exit 70 |

The import-order case uses the existing Node oracle's error normalization. Its raw source error is a ReferenceError; compilation preserves the normalized oracle behavior. The native hook checks sanitized native and leaks on a successful exit. `cycles-record-changes.json` asserts that both changes are NotYet to Compiles and every non-stage0 field, including recorded output, stays identical. The other eight records remain unchanged; directory-callback's generic function-value gap remains NotYet. No output disagreement was observed.

The stale-record control is catchable: before update, both exact stage0 assertions fail despite passing Node and native comparisons. After update, all ten cycle cases pass without update mode.

## Requested validation

All commands have an outer 90-second limit, and every Go test has `-timeout 90s`. Output goes to files.

- `go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s -json`: passes. Restored-run timing is in `test-results.json`; the first green reader leaf was 27.930 seconds.
- `npm ci --prefix stage3/api`: passes before the updater and final fixture run.
- `go test ./stage3/fixtures -run '^TestFixturesCycles$' -count=1 -timeout 90s -json`: passes, package 17.307 seconds, top-level leaf 14.710 seconds. The earlier bounded updater also passes.
- `go test ./internal/lower -count=1 -timeout 90s -json`: the full requested package passes in 65.318 seconds, with 338 passing top-level tests; largest leaf 15.790 seconds.
- `go test ./internal/flow -run '^TestFlowCorpus' -count=1 -timeout 90s -json`: passes in 6.861 seconds, including all 18 remainder programs. This run sets `ADAMIC_UNIT_BUDGET=1`.
- Listed selectable corpus leaves and ran exact anchored selections with `-count=1 -timeout 90s -json`, `ADAMIC_GATE_UNCACHED=1`, in bounded batches of 80. All **831 leaves pass exactly once**, proved by identity and count assertions against the selection list. Exact commands, statuses and timings are in this directory.
- `git diff --check`: passes. No oracle fixture was added or changed, so counts regeneration is not needed.

Setup: Node ready 0.022s, Go ready 0.025s, Markdown and submodules ready 0.072s, clang ready 0.209s, Go build ready 39.561s, cache warm 39.703s, done 39.735s. `nproc=5`, cgroup quota four CPUs. `/workspace/adamic-tools/env.sh` was sourced in build/test shells.

The full repository gate, other stage3 fixture directories, and a whole-corpus byte census were not run in this repair. The requested full lower package and complete flow corpus were run.

Largest selectable flow leaf: 15.590 seconds. Restored reader leaf: 0.700 seconds.
