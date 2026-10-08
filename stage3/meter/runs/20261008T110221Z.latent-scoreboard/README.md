# Stopped full latent scoreboard

Stopped at the user’s request on 2026-10-08T15:43:04.046793+00:00. Roadmap step 05’s speculative census supersedes this unit. No further census work is queued.

| Ref | Compiler SHA | Status |
| --- | --- | --- |
| origin/main | `efe9f4042049234e5a52639fe77b47c311fd530c` | Incomplete: 60/79 directory file records; tsc entry not started |
| origin/area/compiler | `b68b2fe1af2dd6ad369ad8dfbfedf66be5a42861` | Complete: compiler directory and tsc entry |
| origin/compiler/stage3-front-3 | `a36d1c0472649ef2ee549cc2505b93a84e893b59` | Incomplete: 44/79 directory file records; tsc entry not started |
| origin/codex/method-values | `f7de26dbb795d75010d7ebab8dd6357042d11c01` | Complete: compiler directory and tsc entry |
| origin/codex/conditions | `a4218d46fc5610442caefcf6d998baaaecb740b2` | Complete: compiler directory and tsc entry |

Completed exact-reason tables, grouped by owner:

- [area-compiler](area-compiler/TABLE.md): counts, actual-lowering columns, exclusions, boundaries, panics and errors; raw data and exact-reason CSVs alongside.
- [method-values](method-values/TABLE.md): counts, actual-lowering columns, exclusions, boundaries, panics and errors; raw data and exact-reason CSVs alongside.
- [conditions](conditions/TABLE.md): counts, actual-lowering columns, exclusions, boundaries, panics and errors; raw data and exact-reason CSVs alongside.

The main baseline did not finish. Main-to-branch delta tables and the claim that detached-method refusals fall by about 60 are unavailable. No partial stream contributes to any completed table. Incomplete streams are retained under each ref’s `incomplete/`, and the earlier main OOM attempt remains under `failed-attempts/`.

The completed method-values census records seven detached-method refusals in each scope. The completed conditions census records zero condition refusals in each scope, including actual lowering. Zero refers to eligible measured units, not excluded checker-diagnosed bodies or statements hidden by failed compounds.

Each completed measurement built its pinned ref’s own compiler with the 93f7f8e0 full latent overlay and stricter 9d534d3a state-shape guards. All refs use verified identical adaptation inputs and dependency pins. See [PROVENANCE.md](PROVENANCE.md) for toolchain timings, failures, guards, fixtures, mutants and limits.

This evidence-only commit contains this run directory. Draft scoreboard tools remain in the local workspace; no production compiler change is included. Local full-overlay fixtures and mutants, AST tests, no-output guards, scoreboard tests, compiler-selection checks and all 33 meter tests passed. Completed-scope audit logs are retained under `validation/`.
