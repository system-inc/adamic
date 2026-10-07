# Latent census

The current report measures untouched `origin/main` and `origin/area/stage3`,
with no additional feature merges. Each uses its own `apply.sh` and adapted
source tree. [REPORT.md](REPORT.md) and [REPORT.json](REPORT.json) record NotYet
and Refused totals, both combined top-ten reason tables, and full file/reason
ledgers. All counts are measured on a checker-rejected program.

Use `unmerged4.py OUTPUT_DIRECTORY` to recount/render and `audit_unmerged4.py`
to verify untouched pins, source hashes, unique sites, rankings and body skips.
The runner accepts a per-configuration `adapted` path and `unmerged: true`;
it then verifies an unchanged head and empty feature set. Raw observations and
logs are in `data/unmerged4/`. The prior meter is archived in
`data/meter3-REPORT.md` and `data/meter3-REPORT.json.gz`.

The previous meter used pinned `origin/area/stage3` with its own adaptations
10, 20, 30–33, 40, 45 and 46, plus the latest cumulative-2 feature tips.
The archived meter3 reports contain the score out of 78,
the complete zero-diagnostic list, every file with 1–10 diagnostics and all its
codes/lines/causes, totals by code, and the latent lowering and variance ledgers.
All counts are measured on a checker-rejected program.

Recount/render with `meter3.py OUTPUT_DIRECTORY ADAPTED`; verify using
`audit_meter3.py ADAPTED`. `write_report.py` dispatches to the matching renderer.
The previous rerun report is preserved in `data/rerun2-REPORT.md` and
`data/rerun2-REPORT.json.gz`. Meter3 observations, pinned scratch integration,
adaptation hashes and audit logs are in `data/meter3/`. Source hashes identify
the exact adapted input; use a fresh output path for `apply.sh`.

The earlier rerun is archived in `data/rerun2-REPORT.md` and `data/rerun2-REPORT.json.gz`.
Both new configurations have **26/78 checker-clean files (33.33%)**, measured on
a checker-rejected program. The earlier six-configuration report is archived in
`data/previous-REPORT.md` and `data/previous-REPORT.json.gz`; its raw files remain
in data/. New raw observations, stock AST evidence, merge diffs and audits are
in `data/rerun2/`.

For this rerun, use `rerun2.py OUTPUT_DIRECTORY ADAPTED`, then
`write_rerun2.py`. `declarations.cjs ADAPTED MANIFEST_JSON OUTPUT_JSON` uses stock
TypeScript 6.0.3 to independently record declaration spans; set NODE_PATH to the
actual stage3 API cache. `audit_rerun2.py ADAPTED STOCK_DECLARATIONS_JSON` checks
checker totals/clean-file ratios, exact reason counts, nearest named declaration
owners of variance refusals, and deltas against the previous cumulative run.
Its mutants cover the new ratio and variance ownership checks. The older prepare,
summary and report-audit scripts target the original six-configuration format.
The old `write_report.py` entry point dispatches to the new renderer for this format.
Pinned configuration commits and all scratch integration details are in REPORT.json.


This measurement binary sees beyond a program's first refusal. It is built with a
scratch Go overlay; no production compiler source is edited or committed. All
census numbers are **measured on a checker-rejected program**. It does not compile
or run native programs.

The overlay preserves the checker's diagnostics and admits its rejected program
only through `LatentLoad`. Ordinary `Load`, `LoadOverlay`, and `lower.Lower` are
disabled. The latter always returns nil IR and an explicit measurement error.
The ordinary driver fails unless replaced by the overlay. The measurement driver
imports no backend and checks the output guards when `LATENT_ASSERT_NO_OUTPUT=1`.

For each file, it scans refusal syntax without stopping at the first finding,
then attempts each top-level statement and each top-level function whose own body
has no checker diagnostic. It skips and counts diagnosed function bodies, using
raw diagnostic byte ranges rather than displayed line numbers. Signature-only
errors remain eligible. Each attempted unit gets fresh lowering state, with
project declarations registered and sibling signatures prepared lazily. A returned
error or recoverable panic becomes a finding; processing continues. Diagnosed
dependency bodies produce separate `SkippedDependency` events, excluded from
compiler NotYet/Refused counts.

Findings retain kind, location, exact reason/text, attempting unit, phase, and the
measurement label. Headline counts deduplicate `(kind, where, reason, text)` across
all attempts and attribute sites to their actual diagnostic file. Raw JSONL keeps
the attempt context, eligibility ledger, and raw checker spans.

Build and audit from a compiler checkout:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/latent-overlay > /tmp/latent-overlay.log 2>&1
gofmt -w /tmp/latent-overlay/*.go
go build -buildvcs=false -overlay=/tmp/latent-overlay/overlay.json -o /tmp/latent-census ./stage3/census/latent/tool > /tmp/latent-build.log 2>&1
python3 stage3/census/latent/audit.py /tmp/latent-census > /tmp/latent-audit.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/latent-census /path/to/adapted/src/compiler /tmp/latent.jsonl > /tmp/latent-run.log 2>&1
```

`prepare_scratch.py REPOSITORY NEW_TMP_DIRECTORY UNIQUE_BRANCH_PREFIX` creates
main, four individual feature configurations, and one cumulative configuration.
It uses the recorded feature/main SHAs and resolves observed conflicts with
`resolve_scratch.py`. The resolver accepts only `/tmp` trees on `scratch/latent-*`
branches and fails on unknown conflicts. Preparation commits are pinned by REPORT.json as well. Scratch branches must never be pushed. Apply adaptations 10 and 20
from the prepared main tree before measuring all configurations on those same bytes.

`run_comparisons.py REPOSITORY ADAPTED OUTPUT_DIRECTORY WORKTREES_JSON` builds and
measures already resolved worktrees. WORKTREES_JSON is an array of
`{"name":"main","tree":"/tmp/.../main","features":[]}` records; each feature record
names its origin branch, and cumulative lists all four. The runner checks ancestry,
shares the initialized cohere checkout through a scratch symlink, builds with VCS
stamping disabled, and preserves all build/run logs. Independent configurations run
concurrently, with each compiler's measurement driver processing its files sequentially.

`summarize.py OUTPUT_DIRECTORY ADAPTED` writes REPORT.json and compressed raw JSONL.
`write_report.py` renders REPORT.md. Both include complete per-file/per-reason counts
and feature deltas. Common-eligible-unit deltas distinguish blocker changes from
checker eligibility changes. `audit_report.py ADAPTED` independently verifies source
hashes, raw counts, diagnostic/body span overlap, labels, attribution, and deltas,
and proves its checks can fail with artifact mutants.

`audit.py BINARY` covers continuation, multiple refusal sites, body skips,
signature-only errors, same-line functions, and imported sibling calls. The
`LATENT_MUTANT_FUNCTION` environment variable plants one extra overlay-only NotYet;
`LATENT_MUTANT_WHERE` optionally restricts it to one exact location.
`audit_corpus.py OUTPUT_DIRECTORY ADAPTED` verifies that mutant on a real eligible
tsc function: exactly one extra site in binder.ts and identical records everywhere
else. `audit_output_guards.py REPOSITORY OVERLAY_DIRECTORY NEW_TMP_DIRECTORY` builds
non-nil-IR and permissive-production-loader mutants and requires both guards to fail.
`LATENT_MUTANT_BODY_SCOPE` deliberately includes signatures in the body eligibility
range; the synthetic audit catches the incorrectly skipped function.

Limits: lowering still returns its first error within each unit. Generic declarations
are attempted without invented specializations. Symbol registration is best effort;
isolated context can differ from successful whole-program lowering. Final module
order, ownership, and backend passes are omitted. This is an observation ledger,
not proof of exhaustive blockers, successful compilation, or JavaScript semantics.
See REPORT.md for exact provenance, outcomes, mutants, and omitted validation.
