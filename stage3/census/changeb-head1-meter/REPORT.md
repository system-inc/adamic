# Optimizer head 1 against change B

Both first-error censuses completed over the same **79 files**; both pass the meter’s existing `report.py.lowering_summary` coverage/count validator. **Zero A-lowers/B-stops regressions**, zero errors, zero panics. Of 4,543 ledger units, 2,930 lower on A and 3,040 on B; 110 become lowering and 83 move to a different stop. All 2,930 A-lowering units remain lowering on B.

**These are latent measurements on a checker-rejected program**, with body-diagnostic skips, generic deferrals and first-stop limits; “lowers” means an attempted unit has no lowering-phase stop. This does not establish ordinary whole-program compilation or native output. A is change B, not an independent main run.

| Distinct sites | A: change B | B: optimizer head 1 |
|---|---:|---:|
| NotYet | 651 | 672 |
| Refused | 2180 | 2180 |
| SkippedDependency | 3 | 3 |
| error | 0 | 0 |
| panic | 0 | 0 |

Distinct key: `(kind, where, reason, text)`. Total NotYet/Refused: **2,831 → 2,852**. One site disappears and 22 appear as first stops move. Every changing reason and up to five examples per side are in [changed-reasons.md](changed-reasons.md); added/removed examples are also in its JSON. The requested original delta is [census-delta.md](census-delta.md).

View/predicate callouts: `checked view union arm awaits views-v3: array element kind` **0→1**; `checked predicate membership over an open tag domain` **47→50**; overload targets **AssignmentExpression<AmpersandAmpersandEqualsToken | BarBarEqualsToken | EqualsToken | QuestionQuestionEqualsToken> & ... 0→1**, **BindableStaticElementAccessExpression 0→2**, **JSDocTag 0→1**, **PrefixUnaryExpression & { operand: NumericLiteral; } 0→1**, and **PrefixUnaryExpression & { operator: SyntaxKind.PlusToken; operand: NumericLiteral; } 0→1**. `a branded supertype inside a checked predicate's array target` stays **2→2**. No checked-view-read/accessor or checked-view-write/accessor reason appears on either side. [view-predicate-reasons.md](view-predicate-reasons.md) names every matching row, including unchanged ones.

The sole A lowering-phase diagnostic site absent from B is **NotYet, `src/compiler/sourcemap.ts:474:9`, `a getter in a checked field contract` (1→0)**. Its full exact text is in [removed-lowering-sites.json](removed-lowering-sites.json). A-lowering units absent or stopped in B: **none**; the complete list is the empty [lowering-regressions.json](lowering-regressions.json). All 193 changed unit outcomes are in [unit-comparison.json](unit-comparison.json).

| Named unit | A first stop | B outcome | B narrowed accessor reread? |
|---|---|---|---|
| isOptionalJSDocPropertyLikeTag | getter contract at `sourcemap.ts:474:9` | lowers | No |
| isStatic | getter contract at `sourcemap.ts:474:9` | writing a checked object field without its source-slot type certificate at `src/compiler/binder.ts:1034:17` | No |
| isClassOrTypeElement | getter contract at `sourcemap.ts:474:9` | lowers | No |
| isLiteralTypeLiteral | getter contract at `sourcemap.ts:474:9` | lowers | No |
| notImplementedResolver | getter contract at `sourcemap.ts:474:9` | a () => never seen as a (node: ElementAccessExpression \| EnumMember \| PropertyAccessExpression) => string \| number \| undefined (one keeps something weakly that the other keeps strongly) at `src/compiler/emitter.ts:1172:23` | No |

All five A reasons above are exactly `a getter in a checked field contract`. Unit declaration locations and complete stops are in [named-units.json](named-units.json). `notImplementedResolver` is the variable statement at `emitter.ts:1147:1`, resolved from the ledger declaration metadata.

B’s `internal/lower/object.go:617` refusal is exactly **“a narrowed accessor reread, which can return a different value”**; its fix is “read the getter into a local once, then narrow and use that local”. None of the five B outcomes has that reason; it appears at no ledger site on either side.

| Timing on this box | A seconds | B seconds | B−A | B/A |
|---|---:|---:|---:|---:|
| Entire census wall | 2516.645 | 2645.436 | +128.790 | 1.0512 |
| Sum of per-unit wall | 1266.248 | 1415.541 | +149.294 | 1.1179 |
| Sum of per-unit CPU | 1407.740 | 1523.726 | +115.986 | 1.0824 |

4,018 units timed on each side. The top 20 by absolute B−A seconds, with names and ratios, are in [timing-top20.md](timing-top20.md); every unit is in [unit-times.csv](unit-times.csv). Largest: `hasTabstop` at `utilities.ts:11077:1`, **+22.774 s, 4.794×**. Unit timers include snapshot/signature/body work and exclude checker loading, file base registration, and other untimed work; their sums therefore differ from whole-run time.

Observed noise controls: 77 namespace import/export units contain no field, contract, or predicate lowering. Median absolute difference **10.5 ms**, p95 **83.7 ms**, maximum **463.7 ms**; median B/A **0.9976**, range **0.441–5.072**. Their total is **7.655→7.957 s**. This includes snapshot overhead, GC and scheduler noise; one paired run does not establish a statistical confidence interval. Details: [noise-floor.json](noise-floor.json).

Box: `c8fdb452e498`, `nproc=5`, cgroup CPU quota **4 cores**, host memory **18,440,100 kB**, cgroup memory limit **17,179,869,184 bytes**. Peak RSS: **A 5,705,864 KiB; B 6,397,164 KiB**. Both exit 0; neither reaches the 90-minute bound. The binaries launched together; A ran for 41m56.6s, B for 44m05.4s.

| Compiler | Exact SHA | Binary SHA-256 |
|---|---|---|
| A | `84fccdc821a3155ae66c8137c2834eaff5c6cb69` | `0d896e3261ff0bcb749f4e2754a06372935db41ed066a6e912e6cc5107b4e93b` |
| B | `4956436960b289f8791e2de0a3ba5cc22faa5eed` | `4ce266352b389427fa0b92bd58c5237799deb89c6f1c4e556d343d348dbf4f87` |

The tree was built once by A from TypeScript v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`. The 79-source hash manifest was checked unchanged after both runs. Both compiler SHAs already include the timing recorder; **no recorder patch was applied**. No `internal/` file was modified. Scratch worktrees share the same initialized pinned cohere dependency via symlinks.

Exact build/setup/compare commands, pins and minor setup retries are in [COMMANDS.md](COMMANDS.md). Paired invocation: [run-censuses.sh](run-censuses.sh), using the included Linux `wait4` wrapper because GNU time is absent. `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1`, separate `LATENT_TIMES`, simultaneous launch and independent 90-minute timeouts. [analyze.py](analyze.py) reproduces validation and the detailed tables from the raw scratch ledgers. Gzipped ledgers, unit-time streams, logs, box/resources, validation JSON, delta and tables are alongside this report. Only this evidence directory is committed.
