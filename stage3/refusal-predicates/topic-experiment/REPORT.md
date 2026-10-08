# Predicate topic cherry-pick experiment

Base: `337aa466bf7b519ffdc50c6f16e976a330a4a4f5` (`origin/compiler/area-next`). Branch: `codex/notyet-predicates-topic`.

Five of the six own non-merge commits were applied. The views-dependent lowering commit was attempted, counted, and skipped. Neither the views merge `684dc1dc` nor the non-null merge `0ac2a1f5` was imported. No views code was imported separately.

## Conflict counts

Across all six attempts: **15 distinct conflicted paths, 5 textual conflict hunks, and 11 file-location conflicts**. The latter have no textual markers and must not be counted as zero conflicts. If counting each location conflict as one structural unit, there are 16 conflict units total. Compared with the supplied full-merge count of 24 files, this is 9 fewer files (37.5%). The full merge's hunk count was not supplied, so no hunk comparison is claimed.

| Source commit | Result | Conflicted files | Textual hunks | File-location conflicts |
|---|---|---:|---:|---:|
| 50446255 | f3a12e1e | 0 | 0 | 0 |
| c137f65f | 42b601ed | 0 | 0 | 0 |
| adb405d5 | 13759a9c | 4 | 5 | 0 |
| dfced069 | Skipped: dependencies absent | 11 | 0 | 11 |
| 2e78986b | 1c699f22 | 0 | 0 | 0 |
| c191f97b | 38dd6d21 | 0 | 0 | 0 |

`adb405d5` required textual resolutions in `internal/lower/expression.go` (1), `internal/lower/predicates.go` (2), `internal/lower/refusals.go` (1), and `internal/oracle/counts.md` (1). Existing area contextual conversions and predicate-argument checks were retained. The foundation's flow proof was mapped onto the area's `predicateFlowProof` / `proveFlowPredicate` API; the new helper in `internal/lower/predicates_contracts.go` was renamed `closedPredicateRefusal` to preserve the area's existing `provePredicate` summary API. The area's existing count row was retained alongside the two own fixture rows. `refusals.go` has no final diff after resolution.

`dfced069` produced eleven fixture-location conflicts because Git inferred directory moves from source history whose merges were deliberately excluded. Exact paths, raw cherry-pick output, and per-path marker counts are in `ledger.json` and the pick logs.

## Deferred dependencies

Commit **dfced0692dc60eeb9bc6db05a4094eaa596c280d** needs the views symbols `(*lowering).structuralViewCast` and `(*lowering).structuralViewIntersection`, absent from this area. Its `(*lowering).view` dependency also requires the shared lazy admission implementation (`view_lazy.go` and its callers); the area has an older eager scalar contract implementation under that name. It additionally calls `(*lowering).checkedAssertionSource`, supplied by the excluded checked non-null work and absent here. None were imported. The whole commit was skipped rather than substituting different checking semantics.

The two later reporting commits are preserved as historical evidence from the source branch. In particular, **VIEW_REPORT.md's 120/122 coverage does not describe this partial topic**. No new coverage against the refusal table was measured in this conflict experiment. Open structural predicates remain refused here.

## Validation

All output is saved in adjacent log files. Commands run from the topic worktree:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
go test ./internal/lower -run '^TestPredicateContractRefusals$' -count=1 -v
go test ./internal/oracle -run '^(TestPredicateCheckedNarrowing|TestPredicateOpenContractsStayRefused)$' -count=1 -v
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/predicate_(callback_contract|helper_return).a$' -count=1 -v
```

Results: build passed; lower refusals passed (0.172s); narrowing/open-contract tests passed (17.083s); both positive fixtures passed against Node, JavaScript, release native and sanitized native (0.684s). No whole package or full gate was run. No new rule or fixture was introduced by this experiment; existing count rows were retained. Mutant evidence in the cherry-picked reports is historical; mutants were not rerun for this conflict experiment.

Initial setup stalled during the nested TypeScript submodule network clone and was terminated. The exact pinned submodule was then checked out using the original worktree's local Git objects as an alternate, through Git rather than copying source. Setup retry passed: node ready 0.047s, go 0.048s, markdown ready 0.123s (step 0.015s), submodules 0.128s, clang 0.319s, Go build 254.347s, deferred test binaries 254.537s, cache 254.539s, done 254.609s. `nproc`: 5; cgroup CPU quota: 4.

Parser follow-up work remains paused for this requested experiment. The topic is intentionally partial and is not a replacement for the merged source branch's views implementation.
