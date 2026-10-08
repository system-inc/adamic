# Wave 20 checker continuation

All 15 claimed rules stopped at an unavailable shared helper or checker question. These are source-inspection blockers, not newly executed parity results. No new typed descriptor, private checker, copied shared helper or semantic mutant was committed.

| Rule | Required Go call and reproducer |
| --- | --- |
| nexus-no-process-exit-after-output | [nexus-no-process-exit-after-output.md](nexus-no-process-exit-after-output.md) |
| nexus-no-uncleared-race-timeout | [nexus-no-uncleared-race-timeout.md](nexus-no-uncleared-race-timeout.md) |
| nexus-require-blocking-standard-streams | [nexus-require-blocking-standard-streams.md](nexus-require-blocking-standard-streams.md) |
| no-floating-promises | [no-floating-promises.md](no-floating-promises.md) |
| no-implied-eval | [no-implied-eval.md](no-implied-eval.md) |
| no-meaningless-void-operator | [no-meaningless-void-operator.md](no-meaningless-void-operator.md) |
| prefer-promise-reject-errors | [prefer-promise-reject-errors.md](prefer-promise-reject-errors.md) |
| prefer-regex-literals | [prefer-regex-literals.md](prefer-regex-literals.md) |
| prefer-rest-params | [prefer-rest-params.md](prefer-rest-params.md) |
| react-hooks-set-state-in-effect | [react-hooks-set-state-in-effect.md](react-hooks-set-state-in-effect.md) |
| react-hooks-set-state-in-render | [react-hooks-set-state-in-render.md](react-hooks-set-state-in-render.md) |
| react-hooks-static-components | [react-hooks-static-components.md](react-hooks-static-components.md) |
| react-jsx-fragments | [react-jsx-fragments.md](react-jsx-fragments.md) |
| react-jsx-no-constructed-context-values | [react-jsx-no-constructed-context-values.md](react-jsx-no-constructed-context-values.md) |
| react-jsx-no-undef | [react-jsx-no-undef.md](react-jsx-no-undef.md) |

The old native analyses and historical evidence remain on the original wave branch. They are not unified-harness ports and are not counted as newly matched upstream cases. The existing no-self-compare unified descriptor is already integrated on the fetched area tip.

Landing base: area `c4bdc23fa86d55cf7e579989201c11258f4d3a62`, merged into wave 20 at `02ff654e2ab6e0b83709ad7c6af2f754e1ea6ad9`. The provisional area-based branch codex/typeaware-wave-20-checker has no committed rule changes and will not be pushed.

## Toolchain

`bash cloud/setup.sh` initially failed after a checkout changed during its dependency scan: `open bridge/tsgo/checker/accessed_property.go: no such file or directory`. Retrying on the stable area checkout passed: Go ready 0.111s; Node ready 0.135s; submodules ready 0.344s; markdown dependencies ready 0.390s; clang ready 0.865s; Go build ready 302.364s; cache warm 304.706s; done 305.470s. nproc 5, CPU quota 400000/100000 (four cores), reported memory 17.6 GB. Logs are kept in scratch; no credentials or environment dumps are archived.

## Validation command

The merged wave branch is tested with `GOFLAGS=-buildvcs=false GOMAXPROCS=4 ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave20-typescript ADAMIC_LINT_BENCH=1 ADAMIC_LINT_PROFILE_DIR=/workspace/wave20-checker-landing-profiles ADAMIC_LINT_PROFILE_SNAPSHOTS=/workspace/wave20-checker-landing-profiles go test -json -count=1 -timeout 60m ./stage1/cohere/lint`. Registry generation and `go vet ./stage1/cohere/lint/...` run first. The cohere worktree uses the existing pinned submodule via a symlink; buildvcs is disabled because Git rejects a symlink as a submodule working tree. The shared harness builds the checker archive; no private validation driver is used.

## Results

Registry PASS: 75 descriptors. Owned adapter gofmt PASS (empty output). Vet FAIL and full lint package FAIL on the old wave bridge API mismatch: [LANDING_BLOCKER.md](LANDING_BLOCKER.md). Package counts: 0 PASS, 1 FAIL, 0 SKIP. No test cases ran (0 PASS/0 FAIL/0 SKIP); therefore TestOwnedWitnesses, TestMutants and TestRulesAgree did not execute. There were no observed skipped tests; TestCheckerBridgeRefusalPending would skip for the missing TSGoError prelude if compilation succeeded, but that is not an observed result of this run.

Full lint wall time 144.698s; nproc 5; one-minute load start 7.925, end 5.061, median 9.543, peak 11.590. [Summary](evidence/lint-summary.json), [full JSON log](evidence/lint.jsonl), [vet](evidence/vet.txt), [registry](evidence/registry.txt), [gofmt](evidence/gofmt.txt). All required lint input variables were set; no input was omitted to skip a check.

New upstream cases matched: 0 for each of the fifteen blocked rules. New semantic mutants compiled/run/caught: none, because no rule was ported and the merged package could not build. No Node, emitted JavaScript, sanitized native, released-handle or throughput success is claimed. Historical private analyses were not rerun. Stopped on the missing shared dependencies and the merged old bridge API mismatch; no further rules claimed.
