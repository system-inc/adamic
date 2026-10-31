# New B 4a1fe092 against A: paired latent census

Both censuses completed **79 / 79 files** (A / B), exit **0 / 0**. Both passed the existing meter’s `report.py.lowering_summary` coverage and distinct-site validator. This is the checker-rejected, first-error measurement; no usable IR or native output is claimed.

| Distinct (kind, where, reason, text) sites | A | B |
|---|---:|---:|
| NotYet | 689 | 685 |
| Refused | 2,135 | 2,135 |
| SkippedDependency | 3 | 3 |
| error | 0 | 0 |
| panic | 0 | 0 |

There are **6 changed kind/reason rows** and **0 A-successful units absent or stopped under B**. [Every changed reason, counts and up to five removed/added examples](reason-changes.md), the requested [census delta](census-delta.json), [every unit regression](lowering-regressions.md), and [new-stop context](new-stop-context.json) are retained. The context also covers B stops in units that already stopped under A.

Changed rows mentioning a view, branded supertype or predicate:

- `NotYet: a checked predicate target DynamicNamedDeclaration that narrows GetAccessorDeclaration, which is neither its supertype nor a subtype its views cover`: 0 → 1.
- `NotYet: checked predicate overload target ClassElement`: 1 → 0.
- `NotYet: checked predicate overload target JSDocPropertyLikeTag`: 1 → 0.
- `NotYet: checked predicate overload target LateBoundDeclaration`: 1 → 0.
- `NotYet: checked predicate overload target LiteralExpression`: 1 → 0.
- `NotYet: checked predicate overload target TypeElement`: 1 → 0.

The new stop is in `getAllAccessorDeclarations` at `utilities.ts:6838:9`. A already stopped in that unit at `sourcemap.ts:474:9`, `a getter in a checked field contract`, so this is a changed first stop, not a regression from a successfully lowered A unit. No new read-through-view or write-through-view reason was recorded. The rerun A ledger is byte-for-byte identical to the parent run A ledger (SHA-256 `2c2bd81803ebe5e57a81eba37b39f1784850af0fe29eb987da25f6c79d4fcee8`).

[All named view/predicate rows](view-predicate-rows.md), including unchanged rows, carry A/B counts and examples. **Distinct branded supertypes registered and total subtype views are unavailable from these ledgers/logs**, not zero: successful registrations are not recorded. See [registration-counts.json](registration-counts.json). No compiler instrumentation was added to obtain them.

| Timing/resources | A | B | B−A / B:A |
|---|---:|---:|---|
| Sum of unit wall seconds | 652.345070 | 654.044472 | +1.699402 s / 1.002605× |
| Sum of unit CPU seconds | 702.892103 | 703.441681 | +0.549578 s |
| Whole-run wall seconds | 809.307052 | 811.886167 | +2.579116 s |
| Peak RSS, KiB | 4,857,172 | 4,774,708 | -82,464 KiB |

[Top twenty by B−A](timing-top20.md) and [all 4,017 paired timed units](timing-table.md) give seconds and ratios; JSON preserves CPU time too. Empirical noise controls: **1,215** imports, interfaces and type aliases with identical findings, no predicate body in their own statements: median absolute delta **3.635 ms**, p95 **55.006 ms**, median B/A **1.0115×**. These include common registration overhead; scheduling and GC create outliers. One paired run does not isolate per-view cost or guarantee that a later expression was reached after a first stop.

Same box `061be98c179a`, nproc 5, CPU quota {'quota_us': 400000, 'period_us': 100000}, memory 18440100 kB. A was rerun beside new B, with simultaneous starts and separate `LATENT_TIMES`, `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1`, and 5,400-second limits. The shared tree hashes matched the previous pair before this run and matched again afterward.

A: `e0391d48aab3227d3aa5f4c93da39e3fbb2513be`; B: `4a1fe09255a6442aeb7befde57a6f14c5899e423`; unchanged adaptation main: `4e45d80b59bdd03c054d7611c50e6b89cace1a0b`; upstream: `050880ce59e30b356b686bd3144efe24f875ebc8`. The same meter-only recorder changes from `1bab347531e8c238d97940ea843ad3c166509e18` were applied to new B. Meter sources match A byte-for-byte; no `internal/` changes. [Exact commands](COMMANDS.md), [binary SHA-256s](binaries.sha256), gzip ledgers/logs, resource records and validated summaries are in this directory. Parent evidence is preserved unchanged.
