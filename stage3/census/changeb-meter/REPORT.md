# Paired latent census: change B against A

Both runs completed all **79 files**, with **4,017 timed units each**, exit 0. The existing meter `report.py.lowering_summary` validated both ledgers’ coverage and distinct-site counts. The two raw ledgers are **byte-for-byte identical**, SHA-256 `2c2bd81803ebe5e57a81eba37b39f1784850af0fe29eb987da25f6c79d4fcee8`.

| Distinct sites | A | B |
|---|---:|---:|
| NotYet | 689 | 689 |
| Refused | 2,135 | 2,135 |
| SkippedDependency | 3 | 3 |
| error | 0 | 0 |
| panic | 0 | 0 |

**No changed reason rows, new stops, missing A units, or A-lowering/B-stopped units.** The requested [delta](census-delta.json), [changed-reason table](reason-changes.md), [unit regression list](lowering-regressions.md), and [meter validation](validation.json) record this. Identical ledgers also cover attempted units that already had stops, beyond the successful-A-unit regression check.

All view/predicate reason names, counts and examples appear in [view-predicate-rows.md](view-predicate-rows.md). In particular, `checked predicate overload target TypeElement`, `checked predicate overload target LiteralExpression`, `checked predicate overload target ClassElement`, and `checked predicate overload target LateBoundDeclaration` each remain 1 → 1. `checked predicate membership over an open tag domain` remains 45 → 45; `checked view field modifiers of type NodeArray<Modifier> | NodeArray<ModifierLike> | undefined` remains 1 → 1. No reason contains “branded supertype.” No view read/write stop was introduced.

| Timing/resources | A | B | B−A / B:A |
|---|---:|---:|---|
| Sum of unit wall seconds | 596.824203 | 590.417948 | -6.406255 s / 0.989266× |
| Sum of unit CPU seconds | 641.769465 | 638.837079 | -2.932386 s |
| Whole-run wall seconds | 744.742924 | 737.639238 | -7.103686 s |
| Peak RSS, KiB | 4,777,400 | 4,837,964 | +60,564 KiB |

The [largest twenty B−A units](timing-top20.md) show absolute seconds and B/A ratios; [all units](timing-table.md) and [machine-readable timing](timing-table.json) retain CPU and wall observations. Largest increase: `hasTabstop`, +1.020541 s (1.2420×), with the same optional-field-read NotYet on both sides.

Noise controls: 1,215 imports, interfaces and type aliases with identical findings, whose own statements have no checked-predicate body to exercise B. Median absolute wall difference **2.621 ms**, p95 **30.524 ms**, median B/A **0.9932×**. These include common registration overhead. Scheduling and GC can produce larger outliers. One simultaneous pair shows no aggregate slowdown; it does **not** isolate the cost of registering each supertype’s subtype views, prove those views were reached, or establish a speedup. First-error stops can mask later behavior. This is the requested checker-rejected, first-error census, not proof of whole-program lowering or native execution.

Box: `061be98c179a`, nproc 5, CPU quota 400000/100000 µs (4 CPUs), RAM 18,440,100 KiB, no swap. Both binaries ran concurrently against the one unchanged adapted tree. Both had `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1`, plus separate `LATENT_TIMES` outputs; each had a 5,400-second timeout, neither hit it. [Box](box.json), per-run resource JSON, and [source hashes](tree.sha256) are retained.

Exact SHAs and commands are in [COMMANDS.md](COMMANDS.md); binary SHA-256s are in [binaries.sha256](binaries.sha256). A = `e0391d48aab3227d3aa5f4c93da39e3fbb2513be`; B = `8c7afca2230e38952f3bf51357338a6b0f21ea8f`; main adaptation = `4e45d80b59bdd03c054d7611c50e6b89cace1a0b`; upstream = `050880ce59e30b356b686bd3144efe24f875ebc8`. Neither compiler SHA had the recorder, so the same meter-only changes from `1bab347531e8c238d97940ea843ad3c166509e18` were applied to both. No `internal/` files were changed. Complete latent and meter directories were compared byte-for-byte across worktrees.

Preparation overlapped setup, adaptation and builds. Timed measurements started after all completed. GNU time was unavailable; the successful [launcher](run-paired.py) uses Linux `wait4` for resource accounting. Preparation retries and the unused failed shell launcher are documented in COMMANDS.md. Logs are archived here. To rerun analysis, run `python3 analyze.py` beside the gzip ledgers; its meter import expects the prepared A worktree.
