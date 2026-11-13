# Per-unit post-lowering pass ledger

This change records omissions only. It does not call any post-lowering pass.
The one ordered definition is `pass_ledger.go.txt`; `pass_ledger.go` constructs
owned metadata entries from it. The 30-step audit inventory now has a 31st
production step: `checkInspectorDoomedContracts`, after `voidDiscards`. A 32nd,
`checkFlowLiveness`, follows it (#p2n17c4 H3).

Each independently attempted unit gets `pass_ledger` beside its status and
instantiation metadata in the census JSONL. First-error, full and speculative
records use the common finalizer; replay records use it too. Deferred generic
and checker-excluded declarations have no independent pass measurement and do
not acquire a ledger merely because a caller instantiated them.

For a body without a lowering stop, entries are `not_run_scoped`, except
`findCycles`, the Program-region gate, plan application and the ownership query,
which are `not_run_whole_program`. A body stop, including a recovered partial
body or panic, takes precedence: every entry is
`not_reached_lowering_stop`. Syntax/contract scan findings alone are not body
lowering failures. No entry can be emitted as `executed` by this change.

`TestLatentPassLedgerMatchesMainTailPb47` follows main's `LowerWithOptions`
local call graph to the module-statements/unlowerable boundary and compares the
ordered tail with the table. It includes the prepend, async metadata/error
blocks and conditional Program/plan/ownership blocks, including their nested
calls. Dropping an actual table entry makes the unchanged guard fail. A separate
mutation covers nested plan-call drift. The Go cases call `t.Parallel()`, start
no subprocess, and run below 25 seconds.

The reporter requires the complete, ordered required-step vector with every
status `executed` before claiming whole-pass measurement. Missing, empty,
partial, duplicate, skipped and legacy ledgers fail closed. Existing finding
and file counters retain their numeric definitions; their displayed labels now
say **body lowering**. A previously finding-free entry observation no longer
implies “lowers whole.” The omission-as-execution reporter mutant fails the new
coverage-line assertion. The real overlay also exercises successful, stopped
and generic-deferred units.

The updated reporter renders the retained real main entry baseline
`4bba06cb2dc4a0046c11fc97e9953ec85ed0e638` as:

> 3186 units lower without a body stop; 0 have their whole-program passes measured

That number is a rendering of the retained baseline, not a fresh tsc census at
this change's compiler pin. The live reporter computes N from each new run's
unit/finding records. No native or whole-program acceptance is inferred.

The implementation base is main
`219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205`. The probe-isolation fix was still in
Loom's line at the start, so `3a370737` is cherry-picked as prerequisite
`e3dcb944c`; the ledger is one separate commit above it.

## Reader inventory

Searches covered literal latent-stream filenames, `census_records`, unit-field
access and whole-lowering claims across stage3. Active consumers fall into:

* `stage3/meter/report.py`: reads the one Go table through the shared
  `pass_ledger.py` reader, adds body/whole coverage and restricts whole claims.
* `stage3/census/pair/pair.py`: accepts a new or legacy ledger without calling
  metadata a change in body eligibility. Actual first-stop comparisons retain
  their existing definitions.
* `stream_records.py`, `unit_status.py`, census summarizers/recounts,
  `rerun2.py`, and replay readers: preserve extra fields or select named fields.
  The two corpus/report audits now omit only `pass_ledger` when comparing body
  eligibility; they retain it in complete records.
* Lane/shard and adaptation census readers: use named unit fields, source
  identities or counts; no closed unit-key schema rejects the new field.
  Historical evidence and frozen meter validation snapshots are unchanged.

The report regressions cover legacy/new streams, full executed vectors,
incomplete vectors and the false whole-measurement mutant. Pairing and stream
regressions retain the new field without manufacturing body regressions.
The compiler dependency entry remains load/lower/IR, which covers the real
census guard and main-source reads.

Validation: the whole latent Go tree passes (65 cases/subcases), as does its
real overlay form (70). The whole meter suite passes (85 cases, five existing
optional probes skipped), as do all 21 pairing cases and the latent stream
reader cases. The new Go cases complete below 25 seconds. The dependency-map
audit passes unchanged. A real emitted sample and child-status summary are
retained in `evidence/pass-ledger/`; its successful unit has 27 scoped omissions
and four whole-program omissions, its stopped unit has 31 not-reached entries,
and its deferred generic has no independent ledger.
