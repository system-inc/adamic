# Latent relation overflow: task 8dbaea5

The end-of-run answer is unavailable, so whether tsc's checker overflowed remains unknown.

Base: `6f68dd0fc5ee78c0b1230482d2c2089f02b1856b`. TypeScript pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.
Code checkpoint on the requested branch: `e6bfbce5db390af989f4a3c0aec7e469fab0d6f7`.
The identical runtime source was already pushed before this run at checkpoint
`b2bbfc3f56167c7c1ecafbe8110da154f7a67cb1`.
Entry: `/tmp/latent-adapted/src/tsc/tsc.ts`, first-error mode, entry-first resolved reach.

- Run complete: false; supervised child exit: 124.
- RelationOverflow after load: `""`.
- RelationOverflow at run end: `null`.
- Per-unit probing: disabled after measuring its cost.
- Observed endpoint queries: 1.
- Completed unit timing rows: 3381.
- Serialized file records: 0. The complete catalogue and attempted-unit count are unavailable.
- Wall time: 1800.376 seconds; peak RSS: 3911.7 MiB.

The final run reached its 30-minute bound within the environment's roughly two-hour turn window. It is explicitly incomplete; no empty end answer or no-overflow conclusion is inferred. The code and all Go proofs are complete and checkpointed.

## Why per-unit checks are not the default

On checker.ts, 51 calls took **32.598782 seconds**,
mean **639.192 ms**, while unit lowering took
10.569839 seconds. Across the cost experiment,
3362 completed unit queries plus the load query consumed
**972.055675 seconds**. This was too costly.

That experiment was deliberately stopped after checker.ts. Its child received
SIGTERM (supervised exit -15; existing wrapper shell exit 241). It is incomplete
and has no end answer. Its timings are retained separately in the
`per-unit-cost-*` artifacts. The final driver defaults to load/end checks.
`LATENT_RELATION_PER_UNIT=1` remains an explicit measurement option.

If overflow is present after load, every attempted unit receives Lower's exact
NotYet wording. If it appears only at the end, every attempted unit is
conservatively marked, with attribution explaining the unknown onset. Skipped
and deferred units are not represented as independent attempts. No lowering
field was added, so no statecopy row is required.

## Proofs and validation

Every new Go test calls `t.Parallel()` first, has a repository-unique name and
never starts go or a compiler subprocess:

- `TestLatentCensusOverflowMarkReachesProgramAndEveryAttempt` checks the clean
  control and production's failed/overflowed relation-cache mark without TS2859,
  in both endpoint and per-unit modes, including serialized findings.
- `TestLatentCensusEndpointOverflowMarksAllAttemptsAndExplainsOnset` checks that
  end-only overflow marks all attempts and explicitly explains unknown onset.
- `TestLatentCensusEntryReachKeepsRootBeforeDependencies` checks entry ordering.
- `TestLatentRelationOverflowHooksAndProductionWording` guards the driver hooks
  and main's actual NotYet words.

The complete latent overlay suite exits 0. The final whole
`go test -count=1 -json ./stage3/census/...` invocation exits 0.
The gate's parallel/name checks, setup, npm ci, apply and final build exit 0.
The basic census audit exits 0.

The mutant removes only `program.RelationOverflow(typeChecker)`, substituting
an empty answer. `TestLatentCensusOverflowMarkReachesProgramAndEveryAttempt`
goes red with `planted mark disappeared`; child exit 1. The unchanged test emits
program and unit NotYet findings. Both logs are retained.

Each package was also tested in its entirety separately. Paths below are
relative to `stage3/census/`; these are the actual child exit codes:

| Package | Exit |
|---|---:|
| `latent/importclean` | 0 |
| `latent/overlaytext` | 0 |
| `latent/refusalrewrite` | 0 |
| `latent/refusalrewrite/cmd` | 0 |
| `latent/replay` | 0 |
| `latent/replay/worker` | 0 |
| `latent/statecopy` | 0 |
| `latent/statementrewrite` | 0 |
| `latent/timing` | 0 |
| `latent/tool` | 0 |
| `predicates/tools/tool` | 0 |
| `tool` | 0 |

The tool additionally exits 0 under the overlay with all three in-process proofs.
The older full audit exits 1 identically on the changed and unchanged base
census: it expects string/number conditions to be refused, but both produce only
`var`. Its full, first-error-mutant and rollback-mutant observations match after
removing new overflow fields and normalizing temporary paths. The comparison
hashes and both failure logs are retained.

`commands.md` records reproduction and checkpoint details. Raw observations,
timings, metrics, adapted source hashes and the adaptation table are alongside
this report. Scope is restricted to stage3/census/latent and its compiler
dependency declaration.

For this incomplete run, the source-hash manifest covers files represented in
the unit timing stream plus checker.ts. It is not a complete file census.
A follow-up should rerun the endpoint command with a longer wall allowance;
the existing adapted tree and tested binary avoid repeating the cost experiment.
