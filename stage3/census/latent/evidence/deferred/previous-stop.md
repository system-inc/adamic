# Incomplete paired measurement

The implementation and its live pins are saved. The requested final branch
`typescript/census-deferred` has not been pushed, and the final single commit on
`018b7c21` has not been cut: its required overstatement count is not available.
No lowered-to-stopped count or new-findings count is claimed.

The functional code checkpoint is `d5e3d405cc6f69ff83fb908afeb9ad048fd63316` on
`typescript/census-deferred-checkpoint`. Subsequent checkpoint commits only
preserve evidence and pairing support. Temporary checkpoint history must be
squashed onto `018b7c21` after the complete pair, without rewriting a pushed ref.
The final target branch remains available for that normal first push.

## Completed

- Production deferral inventory and priority analysis: read-first.md.
- All latent drivers, including instantiated generic dependency bodies, record
  the NotYet under deferred. Module fallbacks use deferred_scan and honor all
  ordinary scan refusals. Speculative deduplication retains both owners/phases.
- The real in-process Go pin covers a direct written-any cast and a concrete
  generic caller, with and without a later ordinary refusal. All three modes pass.
  The unchanged pin fails against the dropped-read mutant with findings=[],
  owned by callWrittenAny. No Go or compiler child is launched by the pin.
- Whole latent packages and statecopy guards pass. No lowering field was added
  or renamed, so the statecopy policy table is unchanged.
- Predicates census family and seam audits, meter report tests, and module-order
  audits pass. Pair tests also cover deferred first stops, the line-1 module/unit
  collision, and retention of nonlocated program relation reports.
- One adapted tree was built by apply.sh at merge base
  `71b126b06aa5840838330965f77b089fbffee3e4`, from TypeScript source pin
  `050880ce59e30b356b686bd3144efe24f875ebc8`. Its 711 source hashes are frozen.
- The untouched 018b7c21 baseline completed with child exit 0 in 3942 seconds,
  with 4059 attempted units and 504 deferred generic declarations over 81 source
  records. The complete stream, timing journal, source hashes and child status
  are retained. baseline-only-summary.json is explicitly not a paired result.

## Why the measurement stopped

Fresh setup, validation and the long baseline consumed the turn window. Early
head attempts exposed costly repeat predicate proofs and missing generic-body
coverage; those issues were fixed and the live pins rerun. The final head could
not complete in the remaining window. Every head and pair waiter was terminated
and reaped; incomplete and superseded artifacts are labeled and excluded.

The baseline also ends with an explicit relation-overflow report: TS2859,
excessive complexity comparing types '" "' and '{}'. It conservatively marks
all attempted units because onset was observed only at run end. Preserve that
report. Pairing now represents its nonlocated program finding with <program>,
not a fabricated source coordinate. Lowering/deferred unit first stops and
conservative whole-program relation attribution are different observations;
neither is proof of successful compilation.

Production internal/lower sources are unchanged. Therefore none of the 73
production first stops moves from the supplied core.ts:2599:15 reference. This
is a source/pass-order conclusion; the 73-root production gate was not rerun.

## Resume

Reuse the same adapted bytes at `/workspace/scratch/zb0q0zd-adapted`; verify all
source-hashes.json entries if the tree is regenerated. Use the pinned toolchain
and source as in commands.md. The complete baseline can be reused only with those
identical bytes and the same 018b7c21 compiler/checker pins.

Generate the current head overlay and its entry extension, then build the head:

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/zb0q0zd-ts
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/zb0q0zd-resume-overlay
python3 stage3/meter/entry_overlay.py /tmp/zb0q0zd-resume-overlay /tmp/zb0q0zd-resume-entry
go build -buildvcs=false -overlay=/tmp/zb0q0zd-resume-entry/overlay.json -o /tmp/zb0q0zd-head ./stage3/census/latent/tool
```

Write each command's output to a separate log and report its own child status.
Run the head with runner options before its positional prefix:

```sh
python3 stage3/census/latent/evidence/deferred/run.py --watch /workspace/adamic/stage3/census/latent/evidence/deferred/head-times.jsonl --rss-ceiling-kib 15204352 stage3/census/latent/evidence/deferred/first-error-head -- env -u LATENT_FULL -u LATENT_SPECULATIVE -u LATENT_ONLY_FILE GOMEMLIMIT=5GiB LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=/workspace/adamic/stage3/census/latent/evidence/deferred/head-times.jsonl /tmp/zb0q0zd-head /workspace/scratch/zb0q0zd-adapted/src/tsc/tsc.ts /workspace/adamic/stage3/census/latent/evidence/deferred/head.jsonl
python3 stage3/census/latent/evidence/deferred/finish_pair.py
```

The finisher accepts the compressed baseline, waits for both successful statuses,
checks every frozen source hash, runs stage3/census/pair, and renders counts.md
and counts.json. Inspect the newly stopped units, changed existing stops, panics,
new reasons by file, and both relation trailers. Put the actual overstatement
count in the final commit message. Its first line must remain exactly:

```text
never cut alone: lands with chg_pavttxhpyc992z320pnwrjtj2m, because both edit the latent driver templates
```

Only then squash the temporary local checkpoint commits onto 018b7c21 and push
one commit normally to typescript/census-deferred. Never force-push or push main.
