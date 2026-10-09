# Miscompile containment, part A

Step 09, task #63pvx2b. Six common-lowering guards stop eight differential witnesses before either backend can produce unsafe output. This contains implementation gaps; it does not implement the missing runtime behavior.

## Observations

The untouched base was e8ac330e. The delivery tree also incorporates current main f54b8bd3; that merge changes tests and evidence, not these lowering paths. See before.json for Node, native sanitizer and JavaScript results, and after.json for all nine negative witnesses through `c`, `js` and `build --sanitize` (27 path-bearing stops).

| Program | Node stdout | Native before | After |
| --- | --- | --- | --- |
| 9984394_error_spread | no name no message | Error bad 1, exit 0 | NotYet: Error storage spread |
| 4ddd17f_opt_3 | [] [m] | UBSan null string, exit 1 | NotYet: possibly undefined Error message |
| classfeat_maybe_setter | 3 | clang rejects non-scalar self-cast | NotYet: optional numeric setter |
| 4ddd17f_long_name | 2 | clang rejects overlength C literals | NotYet: property name over 4095 bytes |
| iterators_derived_symbol | 1,2 | computed-name compiler panic | NotYet: computed member in derived class |
| iterators_override_source | 300 then 3 | computed-name compiler panic | Same NotYet |
| d7054e9_this_arrow | undefined:NaN\|set | UBSan null string, exit 1 | Existing constructor escape refusal extended to closures |
| d7054e9_this_arrow_number | NaN then NaN,NaN | 1 then 1,2, exit 0 | Same refusal |

JavaScript already matched Node for the first four. Both iterator programs panicked before JavaScript emission. Both constructor witnesses produced wrong JavaScript output as well as unsafe native results.

Error subclass spread has an additional precise negative pin. That program previously stopped on an unrelated unsupported class-base shape, so it is not counted as another retired wrong answer. Library Error types, union constituents and nominal Error subclasses are classified; ordinary structural records remain admitted. Ordinary non-derived iterator classes remain admitted.

The constructor guard reuses the existing last-field-assignment readiness boundary. An arrow capturing this before that boundary is refused even if its only use is a field read. An arrow created after initialization remains admitted. No new runtime representation or native emitter changes are made.

## Verification

Commands run, with output stored here:

- `go test ./internal/oracle -run '^TestMiscompile2A' -count=1 -timeout 90s -v`: eleven top-level tests, Node pins, both backends for admitted controls, ASan/UBSan and leak checks. Final run is fixtures.log; test-seconds.json records each leaf. Earlier cold marker leaf took 44.60 s, controls 19.96 s; all below 60 s.
- `go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(miscompile_2a|census_never_rest_marker)' -count=1 -timeout 90s -v`: registered-fixtures.log.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -v -args -update-counts`: passed, 181.587 s, counts.log. This existing aggregate refresh is not a new test leaf.
- `go vet ./internal/lower ./internal/oracle`: vet.log.
- `python3 review/compiler/miscompile-fxspptb-2a/mutants.py`: six isolated revert mutants, all caught, mutants.log and individual logs.
- `go build -o /tmp/adamic-miscompile-after ./cmd/adamic`: build.log. CLI checks in after.json stop before C, JavaScript or sanitized builds.

The existing never-rest marker fixture also used the optional Error-message gap. Its Error construction now explicitly uses `message ?? ''`; its never-rest function signature is unchanged. marker-node.json confirms unchanged Node output, and TestMiscompile2AMarker holds both backends to it. Counts change only that row (one additional retain/release) and the new admitted controls row. Negative fixtures have no allocation counts.

Each mutant disables only one guard: Error spread, optional Error argument, optional numeric setter, long property name, derived computed member, or constructor capture. The corresponding pinned stop fails; the computed-member mutant restores the original compiler panic. The before results independently demonstrate the unsafe behavior prevented by each guard. No mutant passes because an unsafe program is accepted.

## Setup and limits

`nproc` is 5; CPU quota is 4. Setup timing lines: Node 0.070 s, Go 0.073 s, submodules 0.247 s, markdown dependencies 0.261 s (step skip 0.043 s), clang 0.556 s. Setup failed writing the Go build cache with `no space left on device`; setup.log records the exact failure. `go clean -cache` recovered space without deleting worktrees. A subsequent cold build exhausted the smaller /tmp filesystem; TMPDIR was moved to /workspace/miscompile-2a-tmp and builds were completed sequentially with the installed toolchain.

An attempted worktree cleanup was rejected by automatic approval review because it could destroy uncommitted work. It was not performed; only the safe Go cache cleanup was used.

Real enumerability-aware spread, optional Error runtime construction, setter C representation, long metadata literals and inherited computed-member lowering remain deferred. This unit deliberately stops their narrow shapes rather than repairing those implementations. No whole package or full gate was run. Integration lane results are recorded separately in lane-checks.log.
