"""Render all measured counts, deltas, evidence and limits."""
import json
from pathlib import Path
root=Path(__file__).resolve().parent;result=json.loads((root/'REPORT.json').read_text());runs=result['runs'];base=runs[0]
if result.get('meter_version') == 3:
 import runpy
 runpy.run_path(str(root/'meter3.py'), run_name='__main__')
 raise SystemExit(0)
if 'baseline' in result:
 import runpy
 runpy.run_path(str(root/'write_rerun2.py'))
 raise SystemExit(0)
LABEL='measured on a checker-rejected program'
def esc(s):return str(s).replace('|','\\|').replace('\n','<br>')
text=f'''Built: a scratch-only census that measures eligible top-level functions and every top-level statement across six configurations.
Base: {base['main']}; exact scratch heads and source/binary/overlay hashes are in REPORT.json.
Commands/results: six builds and six corpus measurements exit 0; all census counts are {LABEL}.
Mutants: extra NotYet on probes and real tsc, body-scope, misattribution, output guards, and report artifacts caught.
Limits: first lowering error per unit, diagnosed-body skips, generic/isolation limits, and no native execution claim.

# Scope and method

**Every census number in this report and its JSON is measured on a checker-rejected program.**
This includes checker counts, unit counts, reason/file totals, skipped counts, and
feature deltas. Branch/source SHAs and setup timings are provenance. These
observations cannot establish which programs compile or preserve JavaScript semantics.

TypeScript 6.0.3 is pinned to `050880ce59e30b356b686bd3144efe24f875ebc8`.
The 78 compiler roots are the 77 original sources plus
`diagnosticInformationMap.generated.ts`. `apply.sh` ran with adaptations 10 and 20
merged: type imports changed 72 files / 3,719 lines; optional declarations changed
26 files / 406 lines; combined, 73 files / 4,125 lines. All configurations use the
same adapted bytes. `data/apply.log.gz`, `data/patch-set.md`, and the SHA-256 manifest
preserve this evidence. The two JSON input files are not source roots.

The project is checked once per configuration with all roots together. The overlay
preserves every checker diagnostic but exposes the rejected program only through
`LatentLoad`. It matches each top-level function body's byte interval against raw
checker diagnostic spans in that file. A diagnosed body is skipped and counted;
a signature-only diagnostic leaves its body eligible. Functions on the same line
remain distinct. The raw spans and unit eligibility ledger permit an independent recount.

For each file, the refusal visitor continues past findings and skips diagnosed
top-level functions. Every eligible function and every top-level statement gets a
fresh lowering state. A returned error or recoverable panic is recorded and the next
unit runs. Project declarations, globals, class static storage, and supported enum
values are registered without lowering sibling bodies. Sibling function signatures
are prepared on use. Checker-diagnosed dependency bodies encountered by generic/class
lowering become separate `SkippedDependency` boundaries, excluded from actual
NotYet/Refused counts. Registration failures are rediscovered at actual reads.

Counts deduplicate `(kind, location, reason, exact diagnostic text)` across attempts.
A shared dependency failure is attributed to its actual diagnostic file, and raw
events retain each owning unit and refusal/lowering phase. REPORT.json includes all
unique finding texts/locations, per-file reasons/deltas, and source/binary/overlay
hashes. `data/<configuration>.jsonl.gz` preserves all raw events, checker spans and
units. `outside_roots` accounts for unlocated ordinary errors as well as findings
outside the compiler roots. All NotYet/Refused sites have structured locations.

The refusal scan records every visited refusal site. Lowering still returns its
first error within each unit. These are observed latent findings, not an exhaustive
list of every potential error in every function.

# Measurement only

Only `stage3/census/latent/` is committed on the delivery branch. Measurement edits are
scratch Go overlays, never production edits. Ordinary `Load` and `LoadOverlay`
are disabled in the census binary; `lower.Lower` always returns nil IR and an explicit
measurement error. The driver imports no emitter/backend. Every corpus run enables
`LATENT_ASSERT_NO_OUTPUT`, checking both disabled APIs. The ordinary driver fails
unless overlaid. No scratch branches are pushed, and no native output is produced.

# Configurations and conflict resolution

Baseline is main plus the stage3 pipeline, original census and adaptations 10/20.
Each individual configuration adds the named feature alone to that same baseline.
The cumulative run adds taste, flags, namespaces and nested in that order. Feature
branches include their ancestors; comparisons concern their actual merged heads.

| Configuration | Feature commit(s) | Resolved scratch head | Never-pushed branch |
|---|---|---|---|
'''
for r in runs:text+=f"| {r['name']} | {'<br>'.join(k+': '+v for k,v in r['features'].items()) or 'baseline'} | {r['head']} | {r['branch']} |\n"
text+='''
Individual flags/namespaces conflicts combine main's accessor and nominal checks
with the feature enum/namespace rules. Cumulative resolutions also combine taste's
syntax support, flag safeguards, namespace state/traversal, and nested sibling-call
and capture rules. `resolve_scratch.py` contains exact observed conflict recipes,
accepts only `/tmp` trees on `scratch/latent-*` branches, and refuses unknown conflicts.
Merge/resolution logs are in data/. These are scratch integration decisions, not
compiler changes committed by this unit. Conflicting oracle count rows retain the
current scratch ledger with new feature rows appended; that ledger is unused by the
census and is not claimed as validated oracle evidence. All resolved compilers build.

Dependency pins: cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`,
typescript-go `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
Scratch trees share the initialized checkout through a symlink. Builds use
`-buildvcs=false` because initial VCS stamping did not recognize the symlink as a
worktree submodule. This changes build metadata, not compiler semantics.

# Totals

Every number below is **measured on a checker-rejected program**. Dependency skips
are separate measurement boundaries. Bodyless function declarations are attempted
and can produce the lowerer's ordinary missing-body diagnostic.

| Configuration | Checker diagnostics | Units | Functions attempted | Bodies skipped | Statements attempted | NotYet | Refused | Dependency skips | Errors/panics | Raw events |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
'''
for r in runs:
 u=r['units'];c=r['lowering_counts'];text+=f"| {r['name']} | {r['checker_total']} | {u['total']} | {u['attempted_functions']} | {u['skipped_checker_body']} | {u['attempted_statements']} | {c['NotYet']} | {c['Refused']} | {c['SkippedDependency']} | {c['error']}/{c['panic']} | {r['recorded_events']} |\n"
text+='''
The namespace and cumulative runs each record one ordinary error:
`lower: src/compiler/parser.ts:1472:9: the checker gave a declaration no symbol`.
Its error type exposes no structured location, so it is retained in the unlocated
`outside_roots` bucket and its attempting unit remains in the raw event. It is not
counted as NotYet or Refused. No configuration records a panic.

# Feature deltas

All deltas are **measured on a checker-rejected program**, relative to main.
Negative count deltas mean fewer observed sites. Exact removed/added identities
expose shifted first errors. Common-unit deltas restrict events to units eligible
in both configurations, separating changes in eligibility from changes in blockers.
These observations do not prove successful compilation or feature semantics.

| Configuration | NotYet delta | Refused delta | Removed sites | Added sites | Skipped-body delta | Newly eligible units | Common-unit NotYet delta | Common-unit Refused delta |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
'''
for r in runs[1:]:
 d=r['feature_lowering_delta'];text+=f"| {r['name']} | {d['counts']['NotYet']} | {d['counts']['Refused']} | {d['removed_sites']} | {d['added_sites']} | {d['skipped_bodies']} | {d['newly_eligible_units']} | {d['common_eligible_counts_delta']['NotYet']} | {d['common_eligible_counts_delta']['Refused']} |\n"
text+='\nLargest reason decreases and increases, all **measured on a checker-rejected program**:\n'
for r in runs[1:]:
 d=r['feature_lowering_delta']['per_reason']; changes=sorted((n,k) for k,n in d.items() if n<0)[:5]+sorted(((n,k) for k,n in d.items() if n>0),reverse=True)[:5]
 text+='\n- '+r['name']+': '+'; '.join(esc(k)+f' ({n:+d})' for n,k in changes)+'.\n'
text+='''
# Per file

Every cell is **measured on a checker-rejected program**, showing `NotYet / Refused`.
JSON also includes checker/skip counts, per-file reasons, and per-file deltas.
Zero own findings do not imply the rejected program compiles.

| File under src/compiler | Main N/R | Taste N/R | Flags N/R | Namespaces N/R | Nested N/R | Cumulative N/R |
|---|---:|---:|---:|---:|---:|---:|
'''
for f in result['source']['files']:
 name=f['file'];text+='| '+name.removeprefix('src/compiler/')+' | '+' | '.join(f"{r['per_file'][name]['NotYet']} / {r['per_file'][name]['Refused']}" for r in runs)+' |\n'
text+='''
# Per reason

Every cell is **measured on a checker-rejected program**. Reasons are the exact
lowerer strings, including concrete types. SkippedDependency rows are measurement
boundaries, not compiler blockers. Complete diagnostic text/locations are in JSON.

| Reason | Main | Taste | Flags | Namespaces | Nested | Cumulative |
|---|---:|---:|---:|---:|---:|---:|
'''
for reason in sorted(set().union(*(r['per_reason'] for r in runs))):text+='| '+esc(reason)+' | '+' | '.join(str(r['per_reason'].get(reason,0)) for r in runs)+' |\n'
text+='''
# Mutants and validation

`audit.py` checks continuation through two failing functions, two refusal sites in
one function, an imported sibling call, diagnosed-body skips, signature-only
eligibility, same-line function ranges, and measurement labels. An overlay-only
NotYet at `one.a:3:1` adds one finding in one.a and leaves two.a identical. A
misattribution mutant fails that equality. Expanding body scope to include the
signature wrongly skips `signatureOnly`; the eligibility check catches it.
`data/audit.log` records these checks. Synthetic probes are scratch .a programs,
not TypeScript-derived runtime fixtures or Node/native comparisons.

`audit_corpus.py` plants the extra NotYet in the real eligible tsc function
`getModuleInstanceState`, `src/compiler/binder.ts:330:1`. Binder's unique-site count
rises by exactly one; all other 77 files' entire records and every unit's eligibility
remain unchanged. `data/corpus-mutant-audit.json` records every file's delta.
`data/corpus-mutant-run.log` preserves the run. The baseline contains no mutant.

`audit_output_guards.py` builds two scratch compiler mutants. Non-nil IR from Lower
is caught by `measurement returned usable IR`; exposing the permissive loader through
ordinary Load is caught by `measurement loader exposed an output program`. Both
audit runs fail as intended, with exact guard panic logs preserved in data/.

`audit_report.py` independently verifies adapted source hashes, root coverage,
checker spans, body eligibility, measurement labels, unique findings, attempt events,
per-file reasons and feature deltas. Mutants alter a source hash, drop a file, inflate
checker/finding totals, change a reason/per-file/unit count, remove the label, change the outside-root bucket, alter
a feature delta/common-unit count, and flip a raw skipped-body status. Every mutant is caught.
`data/report-audit.log` records their failures. The actual overlaid scratch compiler
passes `go vet`; all Python scripts pass syntax compilation. All test outputs go to logs.

Initial setup: Go 1.27.1 ready in 0s; clang 20.1.8, Node 24.19.0 and submodules ready
by 1s; cache warming 119s; total 119s. `nproc` is 5; cgroup quota is four CPUs.
`data/setup.log` preserves timing lines.

Commands used after the scratch merges and adaptation:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/run_comparisons.py /workspace/adamic /tmp/tsc-latent-adapted /tmp/latent-complete-runs /tmp/latent-rejected-worktrees.json > /tmp/latent-complete-comparisons.log 2>&1
python3 stage3/census/latent/audit.py /tmp/latent-complete-runs/main-census > /tmp/latent-complete-audit.log 2>&1
python3 stage3/census/latent/audit_corpus.py /tmp/latent-complete-runs /tmp/tsc-latent-adapted > /tmp/latent-complete-corpus-audit.log 2>&1
python3 stage3/census/latent/audit_output_guards.py /tmp/latent-comparisons/main /tmp/latent-complete-runs/main-overlay /tmp/latent-output-guards > /tmp/latent-output-guards.log 2>&1
python3 stage3/census/latent/summarize.py /tmp/latent-complete-runs /tmp/tsc-latent-adapted > /tmp/latent-complete-summary.log 2>&1
python3 stage3/census/latent/write_report.py > /tmp/latent-complete-report.log 2>&1
python3 stage3/census/latent/audit_report.py /tmp/tsc-latent-adapted > /tmp/latent-complete-report-audit.log 2>&1
cd /tmp/latent-comparisons/main
go vet -overlay=/tmp/latent-complete-runs/main-overlay/overlay.json ./stage3/census/latent/tool > /tmp/latent-complete-vet.log 2>&1
```

`prepare_scratch.py REPOSITORY NEW_TMP_DIRECTORY UNIQUE_BRANCH_PREFIX` reproduces
individual/cumulative scratch configurations and writes worktrees.json for the runner.
Main, feature, and preparation commits are pinned by REPORT.json. Apply adaptations from the new main scratch tree, then measure every
configuration on that one tree. Observed resolution recipes were run; a second full
scratch preparation cycle was not repeated.

# Limits and unmeasured work

Lowering stops at its first returned error within each unit. The refusal visitor
can expose multiple syntax refusals, while type-dependent helpers can still return
their first error. Generic declarations are attempted without invented substitutions.
Symbol registration is best effort; some reads may reflect isolated context rather
than successful whole-program lowering. Diagnosed dependency bodies remain skipped
measurement boundaries. The final module-order, ownership, and backend passes are
omitted; local operations may still produce their own cycle/readiness refusals.
Panics/ordinary errors are explicitly separated from NotYet/Refused.

I did not run TypeScript's upstream suite, native-vs-Node comparisons, or the full
Adamic compiler gate. This binary cannot produce runnable native output, and every
adapted configuration remains checker-rejected. The ledger records latent lowering
observations under the stated rules, not native tsc readiness or correctness.
'''
(root/'REPORT.md').write_text(text)
