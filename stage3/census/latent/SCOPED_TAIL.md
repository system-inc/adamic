# Opt-in disposable scoped tail

Parent: `8685b6f42c854030fe71656b7b4ce87d349ed5da`, the approved ledger on
`e3dcb944c74000e5098138b818187bdc8597d111`. The compiler source base is main
`219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205`.

`LATENT_TAIL=1` enables the 13 scoped operations selected by `Scoped` in the
single ordered table in `pass_ledger.go.txt`. The unit retains its statement IR
as Main. Its tail takes `l.latentSnapshot()`, sets entry/source, and discards the
snapshot after observing the passes. No finalized state returns to the unit or
registration base. All IR copying still uses statecopy's checked field inventory
and its explicit private-field policies; no policy exception was added.

Clean operations record `executed_scoped`; a refusal records `stopped_scoped`
with its full error. Later scoped operations record
`not_reached_earlier_pass_stop`. Panics have their own tail finding and ledger
status. The other 18 steps retain their omissions, including four whole-program
steps. These observations do not establish whole-program acceptance.

Tail findings carry `phase: tail` and `pass`. The meter keeps their counts and
root-order artifact evidence separate from body lowering. Its default tail line
is `Scoped tail: not run.` A measured zero is distinct. Required omitted steps
continue to prevent a whole-program claim. The daily runner's default is
unchanged.

## Verification

`TestLatentScopedTailRecordStopOwnsSnapshotLoom03dg` uses production lowering and
the generated snapshot copier on a small admitted discriminated-union assertion
whose union-field read stops at `checkRecordGuardedReads`. It checks the finding,
ledger order/censoring, and equality of the unit state before and after the tail.
`TestLatentScopedTailEntryFindingLoom03dg` runs that fixture through the actual
first-error census in the overlay suite. Both tests run in isolated children of
the current test binary, call `t.Parallel`, and have bounded child lifetimes;
neither starts a compiler build. The exact guarded-read reason is:

> a union or optional field read in a program with record storage

The actual copy-back and enabled-tail bypass mutations both turn the unchanged
state guard red. Their lines and assertions are recorded in
[the mutant evidence](../../meter/scoped-tail-mutants.md).

The whole latent Go tree, its real overlay suite/build, latent stream readers,
the whole meter suite, the pairing suite, and the compiler dependency audit were
run. `compiler-dependencies.json` now explicitly declares `internal/flow` for
this package. The normalizer is the only new compiler component used directly.

## Measurement

The first-error observation covers all **81** resolved sources and **4,566** top-level units of the checker-clean tsc entry. **3,186** units lower without a body stop; **0** have their whole-program passes measured. There are **2** scoped-tail stops and **0** panics. All twelve children exited **0**, before their 1,700-second deadlines. Every attempted unit has the ordered 31-step ledger; the audit checks clean execution, stop censoring, and preserved omissions.

| Scoped step | Calls | Stopped units | Panics | Timed pass wall (s) | Prototype stops |
| --- | ---: | ---: | ---: | ---: | ---: |
| `forwarderValues` | 3186 | 0 | 0 | 0.007453 | 0 |
| `finishClassCalls` | 3186 | 0 | 0 | 0.010585 | 0 |
| `finishAccessors` | 3186 | 0 | 0 | 0.002471 | 0 |
| `finishToPrimitive` | 3186 | 0 | 0 | 0.010243 | 0 |
| `checkRecordGuardedReads` | 3186 | 2 | 0 | 0.002303 | 2 |
| `flow.NormalizeAsync` | 3184 | 0 | 0 | 0.369700 | 0 |
| `asyncEntrySource` | 3184 | 0 | 0 | 0.000995 | 0 |
| `checkViewOptionalWrites` | 3184 | 0 | 0 | 0.005247 | 0 |
| `checkHostMemberWrites` | 3184 | 0 | 0 | 0.002080 | 0 |
| `asyncPayloadFailures` | 3184 | 0 | 0 | 0.000532 | 0 |
| `checkViewLogicalWrites` | 3184 | 0 | 0 | 0.278805 | 0 |
| `borrow` | 3184 | 0 | 0 | 0.743206 | 0 |
| `counters` | 3184 | 0 | 0 | 0.625744 | 0 |

Both stops have the exact reason:

> a union or optional field read in a program with record storage

| Unit | Read site |
| --- | --- |
| `isTypeDeclaration` — `src/compiler/utilities.ts:11026:1` | `src/compiler/utilities.ts:11038:20` |
| `isTypeOnlyImportDeclaration` — `src/compiler/utilitiesPublic.ts:1536:1` | `src/compiler/utilitiesPublic.ts:1539:60` |

Every paired body finding and every unit field outside the pass ledger is identical. The stock pairing tool independently reports two new tail findings, zero lowered/moved findings, and zero body first-stop deltas in both raw and distinct-read views (`evidence/scoped-tail/pair-check.json`). These are added **tail** findings; the body first-stop population is unchanged. The actual report lines are:

```text
3186 units lower without a body stop; 0 have their whole-program passes measured
Scoped tail: 2 units stopped (checkRecordGuardedReads: 2).
```

| Cost | Tail off | Tail on | Observed difference |
| --- | ---: | ---: | ---: |
| Sum of census unit-clock wall (s) | 2250.477979 | 2462.639733 | +212.161754 (9.43%) |
| Sum of child process wall (s) | 5251.029134 | 5363.398966 | +112.369832 (2.14%) |
| Maximum child peak RSS (MiB) | 2262.257812 | 2402.609375 | +140.351562 |

Direct pass timers sum to **2.059364 s**. They use `time.Now`/`time.Since`, as the census unit timer does. Unit times include the disposable snapshot and tail bookkeeping; the direct pass sum excludes them. Child wall includes loading, independent refusal scans, reporting and serialization. Peak RSS is the maximum child high-water mark, not a sum.

At most four census children ran concurrently; verification also overlapped some early chunks. These are paired observations on this box, with one observation per group. The wall and peak differences are not isolated causal estimates of snapshot or individual-pass cost. Per-child costs and actual exits are retained in `evidence/scoped-tail/summary.json`. No aggregate relation trailer or global checker-cache size is fabricated from segmented observations.

The original [prototype report](https://github.com/system-inc/adamic/blob/4cadb42b6f48af2e4a99f334da775f308ea88818/stage3/census/latent-whole-passes/REPORT.md), on audit base `a93ac0b94`, also found these two guarded-record-read sites and zero stops in the other enabled steps. Its broader 271-stop result included 269 stops in passes omitted here. It omitted independent refusal scans and preceded probe isolation; its wider pass set also censored later invocations. Its directly timed work for this selected subset was approximately 0.906 s. The current result does not measure the omitted checks.

The input hashes, runtime overlay hashes and toolchain are in `evidence/scoped-tail/provenance.json`; `unit-ledger.json.gz` retains every measured unit ledger, including per-pass times. The [evidence reproduction notes](evidence/scoped-tail/README.md) describe the bounded file groups and positive guard.
