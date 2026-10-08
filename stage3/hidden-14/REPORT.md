Recorded roadmap step 30's evaluator-result witnesses and exact pinned head replay; no overload mechanism changed.
Base: 4885cec50290686df487b62aac47c85d871ed40c on codex/hidden-14-evaluator-overload-result; preparation is a local commit, awaiting codex/overload-results.
Exact replay exited 0; focused source/refusal tests passed; counts refresh passed after installing the existing pinned Node types.
Number-return input mutant failed the source Node comparison; its separate negative fixture remained Refused at the result promise.
Native ASan/UBSan/leak and JavaScript-backend comparisons, region remeasurement and delivery push remain pending the shared dependency.

## Observations

The branch starts at the requested origin/compiler/area-next-fixtures tip.
CLAUDE.md and its required language/memory documents were read. An isolated
388096e6 worktree regenerated the adapted compiler input with stage3/apply.sh.
All 82 file hashes match that pin's RESULT.json; the verified hashes are in
[evidence/source-hashes.json](evidence/source-hashes.json). utilities.ts is
`ef43309e71a5bff946d868de2753b73de6b3cae07e5f215e2a2f1281ef0406ef`.
The existing cohere submodule was referenced by symlink, not copied.

The exact replay on the area tip selects evaluate and reproduces Refused at
utilities.ts:11320:5 with the complete original reason:

```text
overload 1 of evaluate result EvaluatorResult<string | undefined> cannot be served by implementation result EvaluatorResult<string | number | undefined>
```

The [replay log](evidence/replay.log.txt) records exit-0 reproduction and its
measurement-only context. Replay load/register/lower was 1.856995652s and total
2.139712501s. A preliminary replay overlapped adaptation and is excluded; the
reported replay ran after all 82 source hashes matched.

The sound .a fixture is the supplied readonly result witness. Source Node prints
`node\n`, empty stderr, exit 0. Its independently saved negative changes the
implementation to return `{ value: 7 }`: Node prints `7\n`, empty stderr, exit 0;
the checker accepts the source and lowering refuses its overload result. The
negative lives in internal/load/testdata/0.1/refuse, already excluded from cohere
soundness checks for deliberately wrong inputs. Neither fixture is registered
for native comparison before the shared overload mechanism is available.

The number-return input mutant was run independently against the sound witness:
changing only `return { value };` to `return { value: 7 };` made the focused source
comparison exit 1 at stdout `7\n` versus expected `node\n`. The negative refusal
probe exited 0 and logged the exact overload-result Refused diagnostic. The
sound witness was restored. [Mutant](evidence/number-mutant.log.txt),
[negative refusal](evidence/negative-refusal.log.txt), and
[restored fixtures](evidence/fixtures.log.txt) retain these observations.
This proves the input mutant's comparison and current refusal; it does not claim
that the pending sound specialization has run natively.

## Commands and counts

All test output went directly to log files. Only focused selectors ran:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestHiddenBoundaryEvaluatorOverloadResult' -count=1 -v
# PASS, oracle 0.261s

go test ./internal/oracle -run '^TestHiddenBoundaryEvaluatorOverloadResultSource/testdata$' -count=1 -v
# input mutant: FAIL, oracle 0.168s, intended stdout assertion

go test ./internal/oracle -run '^TestHiddenBoundaryEvaluatorOverloadResultNegative$' -count=1 -v
# PASS, exact Refused diagnostic

go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
# PASS, oracle 38.984s
```

The first counts refresh exited 1 because the base requires @types/node 25.3.3
in stage3/api/node_modules. `npm ci --prefix stage3/api --ignore-scripts
--no-audit --no-fund` installed the existing lockfile's three packages, exit 0;
no package manifests changed. The retry passed. Both runs' logs are retained.

Every counts diff is inherited registration cleanup, with no numeric change:

- logical_and_reference_maybe.a moved to its registration order; its
  allocations/frees/retains/releases/peak/regions remain 8/8/11/22/4/0.
- stage3/fixtures/taste/17_binder_flow.a lost its stale 51/51/0/51/7/0 row:
  taste_stage3_test.go already registers it with lowers=false on this base,
  and counts includes only lowering fixtures.

No evaluator counts row is added yet; there is no compiled evaluator result
claimed. [Counts log](evidence/counts.log.txt) records the successful refresh.

Setup used the required GOPROXY fallback. Its cumulative timing lines were
submodules 0.074s, Go 0.081s, Node 0.087s, markdown ready 0.236s (validated
installed bytes; step 0.026s), clang 0.463s, Go build 47.203s, test binaries
deferred 47.405s, cache warm 47.406s, done 47.433s. nproc=5, CPU quota=4,
Go 1.27.1, Node 24.19.0, clang 20.1.8. The printed environment file was
/workspace/adamic-tools/env.sh. [Setup log](evidence/setup.log.txt).

## Handoff and limits

The coordination instruction assigns the mechanism to Codex 01a11518-9272,
codex/overload-results. The uncommitted second-proof prototype was discarded;
no internal/lower, internal/native, splitter or central oracle harness edits
remain. Earlier prototype checks and a preliminary false-registration harness
failure are not claimed as final validation. A full baseline census had already
started before coordination; it was stopped, exit 143, and its partial output is
excluded from measurement. No whole-package tests or full gate ran.

The original region is utilities.ts [455532,462634), old hidden intersection
7,102 bytes. New hidden intersection and byte difference are **not measured**.
The current first boundary is still the exact overload-result refusal above.
No revealed-byte claim follows from this fixture preparation.

Per coordination, stop here. After overload-results pushes, rebase this local
preparation onto that shared branch, enable the sound fixture's three-way oracle,
rerun the number-return mutant/refusal, refresh counts, and remeasure the same
hash-pinned interval with the hidden census. Report its union intersection and
next boundary, then push the finished delivery branch once. Until then the
string-only promise stays refused rather than silently broadened.
