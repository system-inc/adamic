# Meter stall investigation, 2026-11-05, fq9ycej

Stopped with evidence, without a first bad commit. The prescribed 15-minute no-output cutoff fires while this reproduction is still making progress. Its SIGQUIT stack is in the latent overlay's state copier, not a wait. Neither declaration singled out by the observations is a demonstrated infinite loop. Reporting a first bad commit from the initial three-minute replay predicate would be incorrect.

## Reproduction

Read the complete `stage3/meter/twice-daily.sh` before running anything. Built the script's latent overlay and entry overlay from `1f7df6a2e8145391dea19da63a172d5385638e39`, with the pinned recursive submodules. Ran apply from that revision against TypeScript v6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`. Apply completed successfully. Its product key was `4cbb2bbf3b4958c239cdcb044388f2036419c99bc4493f473f0fd8ae68f65aaa`.

The measured command was:

```sh
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 \
  /tmp/meter-stall-20261105/entry-census \
  /tmp/meter-stall-20261105/adapted/src/tsc/tsc.ts \
  /tmp/meter-stall-20261105/latent.jsonl
```

Stdout and stderr went to `latent.log`. Existing optional `LATENT_TIMES` and `LATENT_PROGRESS_FILE` sidecars were enabled. Full mode does not initialize the speculative progress hook, so the latter produced no progress file. The watchdog reset its deadline on growth of the original JSONL or log, not the timing sidecar. It waited 900 seconds without growth, sampled `/proc`, and sent SIGQUIT. Total elapsed time was 919.43 seconds, including load before metadata output and CPU sampling after the deadline. Go exited with status 2. See `watch.py`, `full-timing.json`, and `latent.log.gz`.

The output contains exactly one JSONL record, the initial checker-clean measurement metadata, 5,893 bytes with these scratch paths. `latent.log` was empty before SIGQUIT. This matches the reported shape of the operations artifact, but does not reproduce its claimed stationary lowering unit: 905 attempted-unit timing records arrived during our run.

## Where the cutoff caught the process

`Latent` in the saved `latent-hook.go.txt` buffers every `LatentFile` until `latentRun` returns, then emits all records. Therefore the last JSONL record cannot identify the next declaration. After metadata, the internal traversal begins with the entry's import at `tsc.ts:1:1`; that is not evidence that this import is stuck.

The actual last completed timed unit at cutoff was `TypeFlags`, `src/compiler/types.ts:6335:1`, taking 0.384 seconds. The next declaration in that file's top-level order is the type alias `DestructuringPattern`, `src/compiler/types.ts:6437:1`. At SIGQUIT, goroutine 1 was taking the statement snapshot before attempting this non-function declaration:

```text
runtime.ifaceeq
lower.latentPrivateIRField                 latent_state.go:1527
lower.latentCopyIR[...].func1               latent_full.go:111
lower.latentCopyIR[...].func1               latent_full.go:166
lower.latentCopyIR[...].func1               latent_full.go:117
lower.latentCopyIR[...].func1               latent_full.go:85
lower.latentCopyIR[...]                    latent_full.go:189
lower.latentCopy_lowering                  latent_state.go:266
lower.latentCopyPointer_lowering           latent_state.go:721
(*lowering).latentSnapshot                 latent_full.go:193
(*lowering).latentStatement                latent_full.go:237
latentFullSelected.func3                  latent_units.go:195
```

The four `/proc/<pid>/stat` samples gained 898 user-plus-system ticks at 100 Hz over 6.212 seconds: 8.98 CPU seconds, or 1.446 cores. This is busy CPU work, not an idle wait or deadlock. The goroutine was runnable. A single snapshot stack does not prove a nonterminating loop. See `proc-stat.txt`, `cpu-summary.json`, `last-unit-records.jsonl`, and the complete dump.

Replaying the actual next declaration alone on the same bad-endpoint binary completed in 26.40 seconds including load and registration. The selected unit is recorded in `unit-next.jsonl.gz`.

## The misleading candidate and stopped bisect

Earlier, the timing stream paused after `tracing.ts:210:9`, the nested `indexFromOne`. The next declaration was `tracingEnabled.dumpTypes`, `tracing.ts:218:5`. A replay of that exact unit completed on `f87d7884807161162c98a2d0ef3b5339e16b315f` in 41.25 seconds including load and registration. A replay on `1f7df6a2` reached its 180-second SIGQUIT deadline.

That initially looked like the requested one-unit predicate, so a path-filtered bisect was started over `internal/lower internal/load internal/flow internal/fresh`. Before any midpoint was classified, the full-pass timing stream proved that `dumpTypes` had completed on `1f7df6a2`: 156.776 wall seconds and 151.309 CPU seconds for the body attempt. The isolated test also includes load and registration, which the body timer excludes. The 180-second cutoff therefore confounded slow completion with a hang. The bisect was stopped and reset instead of publishing a false first bad commit. Its first selected midpoint, `e28e4bc8f99d410f8ec055166ac642f9788c36b9`, was interrupted during its build and is unclassified.

The isolated timeout dump's path was `NameResolver.Resolve` through `proveModuleReads`, `viewFieldReceiverTypeID`, `checkedViewReceiverField`, `view`, `structuralViewCast`, and `cast`, while lowering `dumpTypes`. It was not a generator resume loop or a blocked channel. This is a different observation from the full-pass snapshot dump; they must not be merged into one claimed cause.

The same replay completed with a five-minute cap in 145.12 seconds including setup. Its selected unit is `dumpTypes`, and its exit status is 0. See `unit-bad-300.log`, `unit-bad-300.jsonl.gz`, and `per-commit-timings.csv`.

The replay driver uses the existing `LatentReplay` selection, preserving the entry-first resolved file order. Its deliberately nonexistent finding signature produces an expected "signature did not reproduce" message after completion. Completion means the selected attempt returned; it does not mean the compiler admitted the unit or produced IR. The returned records confirm the selected units.

## Limits and stopping reason

No first bad commit, culprit diff, or actual nonterminating unit has been established. The original operations JSONL and live process were not available in this checkout; the supplied byte count alone does not locate a declaration because of output buffering. The mandatory reproduction cutoff stopped our process while it was still completing units. The two isolated candidates are not evidence of the original 1 hour 42 minute stationary unit. Continuing the bisect with the disproved predicate would answer a different question.

These observations do not establish a hang in the plain compiler on a real program, nor prove that the original incident is only a meter defect. No plain-compiler hang claim is made. No compiler or meter fix was attempted.

The evidence includes both SIGQUIT dumps, the full partial JSONL and timing stream, last records, the partial bisect log, candidate commit list, per-commit timings, replay driver, watchdog, exact overlay source explaining buffering and snapshots, preparation logs, and binary hashes. All census processes started for this investigation were stopped or exited before publication. The exploratory whole-pass good control was terminated after its single-unit control completed; it is not claimed to have completed a whole census.
