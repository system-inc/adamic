# Paired intersection-rule census

Completed: **79/79 files on both sides**, both exits 0; the meter's existing `report.py::lowering_summary` validates coverage and distinct-site counts. Both checker headers are identical. This is `LATENT_FULL=0`, measured on a checker-rejected program, with `LATENT_ASSERT_NO_OUTPUT=1`.

| Distinct sites | A | B |
|---|---:|---:|
| NotYet | 672 | 671 |
| Refused | 2180 | 2180 |
| error | 0 | 0 |
| panic | 0 | 0 |
| SkippedDependency (separate) | 3 | 3 |

**Zero units lower on A and stop or disappear on B.** The ledgers contain the same 4,543 units and identical unit metadata: 3,040 lower without a first-stop finding, 978 stop, 498 defer generic instantiation, and 27 skip checker-rejected bodies. Exactly one unit changes its first stop; none changes success/stop outcome.

`nullParenthesizerRules` (`factory/parenthesizerRules.ts:675:1`) moves from NotYet **“a checked predicate target LeftHandSideExpression over a generic argument TIn & ({} | null), read on the true side as Expression & LeftHandSideExpression”**, at `core.ts:1784:32`, to NotYet **“a template interpolating an object, an array, a map, a function or undefined”**, at `core.ts:1786:59`. The latter is the failure-message interpolation in `cast`, already a distinct site in A; its reason count stays 2 → 2. B therefore removes one distinct site and adds none.

The predicate/read row above is the **only reason row that changes: 1 → 0**. No view or branded-supertype row grows. [Every changed reason and its example location](reason-changes.md), [all 44 view/branded-supertype/predicate rows, including unchanged ones](view-predicate-rows.md), [the sole removed lowering site](removed-lowering-sites.md), [every moved unit](unit-moves.md), and [the empty regression list](lowering-regressions.md) are enumerated separately.

Source AST reading finds **78 direct calls in 22 top-level units**. [Every caller's A/B outcome, first stop and location](selected-unit-outcomes.md) is recorded, with all source calls in [direct-calls.json](direct-calls.json). Only `nullParenthesizerRules` moves. On each side two caller units lower (`getTsConfigObjectLiteralExpression`, `identifierToKeywordKind`), fourteen stop, and six skip checker-rejected bodies; skips are identified with their checker code and location and are not called successful lowering.

| Timing/resource measurement | A | B | B − A | B / A |
|---|---:|---:|---:|---:|
| Sum of 4,018 per-unit wall times, seconds | 587.536049 | 586.977423 | -0.558626 | 0.999049 |
| Sum of per-unit CPU times, seconds | 624.052572 | 634.428822 | +10.376250 | 1.016627 |
| Entire run wall time, seconds | 1022.550418 | 1015.994187 | -6.556231 | 0.993588 |
| Peak RSS, KiB | 6431564 | 5918008 | −513556 | 0.920151 |

The unit recorder includes registration, lowering and panic recovery; whole-run time also includes checking, refusal scans and output. Each attempted unit has exactly one timing on each side. [The 20 largest B-minus-A differences](timing-top20.md) give absolute seconds and B/A ratios; [all unit timings](unit-times.tsv) and the raw timing streams are retained. `nullParenthesizerRules` takes 0.118811 → 0.369940 seconds (+0.251128, 3.114×).

Noise controls: 168 import/export declarations (no receiver read is lowered) have median absolute paired difference **2.644 ms**, p90 **20.407 ms**, maximum **96.511 ms**; median B/A 0.9858. Their summed times are 9.127494 → 9.354933 seconds. The total lowering-wall difference is **−0.095%** in this single simultaneous pair; it does not establish a timing cost or a speedup. Larger per-unit differences can include registration, GC and scheduling noise.

Box: `d912081dbb9a`, `nproc=5`, CPU quota `400000/100000` (4 CPUs), host memory `18440100 kB`, cgroup memory limit `17179869184` bytes (16 GiB). Peak RSS is 6.134 GiB A / 5.644 GiB B. Both runs began together on this box and finished in about 17 minutes, below the independent 90-minute bounds. [Box/load](box.json) and [A](A-resource.json)/[B](B-resource.json) resource records are retained.

Pins: A **`1b516b1b53b2a49abd619f6c220516091222b350`**; B **`021e43f3e40d76fee81a62942d3b4280f846647e`**, sole parent A; TypeScript **`050880ce59e30b356b686bd3144efe24f875ebc8`** (v6.0.3). The shared adapted tree was built once from A. Both pins have identical stage3 tree `40025ede76f13f0dbd06e593e9a3ad1e4808f475` and already contain the `1bab347531e8c238d97940ea843ad3c166509e18` unit-times recorder; **no patch was needed**. `internal/` was unchanged in both worktrees. [Dependency provenance](provenance.json) and [79 adapted-source hashes](adapted-source-sha256.txt) identify the inputs.

Binary SHA256s:

- A: `3dbfad3e0f27cb49936363744b373b6a78529ea3d19d01062f75e505089e7f48`
- B: `3358a32e3c421c4ce27db6a137db61a5f358fb0b331874001df9ef51b05c32fe`

[Exact commands](COMMANDS.md), [build](build.sh), [paired runner](run-censuses.sh), [analysis](analyze.py), [meter validation](validation.json), and the requested [census-delta output](census-delta.json) accompany the gzipped ledgers. Preparation notes: setup succeeded in 965.589 seconds; an unnecessarily slow full-history upstream clone was canceled and replaced by a successful shallow clone of the exact pin. An initial launch wrapper found `/usr/bin/time` absent and exited 127 before either binary started; the successful run used Python's standard-library child-resource accounting with the same census commands and timeout. No compiler implementation was changed for this measurement.
