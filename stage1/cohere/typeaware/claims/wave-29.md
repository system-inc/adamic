# Type-aware wave 29

Branch: `codex/typeaware-wave-29`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed rules, positions in the remaining checker-dependent by-volume ranking:

| Position | Rule | Compiler | Repository |
| --- | --- | ---: | ---: |
| 85 | id-denylist | 0 | 0 |
| 86 | id-match | 0 | 0 |
| 87 | nexus/concurrency-no-check-then-write | 0 | 0 |

Selection combines `validation-volume/compiler-all.counts` and
`validation-volume/repository-all.counts`, sorts by descending combined count
then lexical rule name, and excludes the 26 rules documented in README.md and
COVERAGE_REPORT.md. VOLUME_REPORT.md references these full count tables; its
inline table contains only the ten earlier additions. The separate coverage
inventory has 204 entries and is not the 197 checker-dependent population used
here.

Fetched all origin heads and checked stage1 source blobs and claim documents.
No implementation or claim for these rules was found. Config set membership is
not an implementation. No rule was skipped.

## Continuation after completion of the first claim

Original three rules completed, tested and pushed in 0fa8f823, including configured naming patterns and native Go-regex instruction execution. All origin heads were refreshed again before selection. Main and the bridge branch contain none of the following rule names in their stage1 Adamic/TypeScript implementations. None is mentioned in any origin typeaware claim Markdown blob. Ranking uses the same combined-volume/lexical ordering and the 26 documented baseline ports.

1. `no-restricted-globals`: combined volume 0.
2. `no-setter-return`: combined volume 0.
3. `no-shadow-restricted-names`: combined volume 0.

This update is committed and pushed before any new rule code is written. Earlier available candidates became claimed by other workers and were skipped during the refreshed scan. The remaining emitted-JavaScript checker-adapter gap is a shared harness dependency; native findings, fixes and suggestions are still held to Go.

Native implementations and comparison evidence for these three rules are in [wave-29-next/REPORT.md](../wave-29-next/REPORT.md). All configured and frozen-corpus comparisons, three rule mutants, raw shorthand facts, released handles and sanitizers pass. JSX parsing and the emitted-JavaScript checker adapter remain shared dependencies and are explicitly reported; no further claims were taken.

## Third batch

Previous six native rule ports are tested and pushed through 7f4eabb8. The shared JSX parser and emitted-JavaScript checker adapter limitations remain documented in wave-29-next/REPORT.md. A refreshed scan of 389 origin refs and 33 unique claim Markdown blobs found 26 baseline ports, 141 other claimed names, and 30 available names in the 197-rule volume ranking. The first three remaining names, all with combined default volume zero, are:

1. `react-hooks/set-state-in-effect`.
2. `react-hooks/set-state-in-render`.
3. `react-hooks/static-components`.

This claim update is committed and pushed before implementation. No other rules are claimed by this update.

Third-batch status: blocked, not ported. [wave-29-third/REPORT.md](../wave-29-third/REPORT.md) records positive Go findings, native JSX misclassification/failure and the separate missing native React SSA pipeline. No additional claims were taken.

Follow-up: static-components has a tested supplied-SSA native kernel (34 controls, two compiling mutants, sanitizer agreement), but no source-to-SSA entry. Published JSX parsing works in scratch. All three full ports remain blocked; no additional claim is taken. See the follow-up section of wave-29-third/REPORT.md.

## React analysis claims parked

As instructed by Ahra, the third-batch claims are parked and count as finished
for the landing-first cap. Their existing native kernels and blocker evidence
are pushed through 18ff19c8, rebased onto current main f8013f0b and re-green.

| Parked rule | Blocker |
| --- | --- |
| react-hooks/set-state-in-effect | Native source-to-high-level IR, SSA, hook/setter and capture analysis integration |
| react-hooks/set-state-in-render | Native source-to-high-level IR, SSA and render control/capture analysis integration |
| react-hooks/static-components | Native source-to-high-level IR, SSA, creation/capture analysis integration |

Cohere analysis modules are being ported to Adamic on #dnv6f2c. JSX support
is landing on area/stage1-lint. Parking does not claim full source-rule parity.

## Fourth batch

Fetched 529 origin refs and checked 33 unique typeaware claim blobs against the
197-rule combined-volume ranking. Excluding the 26 baseline ports and every
claimed name leaves 18 candidates. The first three are:

1. `react/jsx-fragments`.
2. `react/jsx-no-constructed-context-values`.
3. `react/jsx-no-undef`.

All have combined default volume zero. They do not use the parked native React
HIR/SSA/capture pipeline. Main and the bridge branch mention these names only
in inventories/counts, not native rule implementations. This claim is pushed
before new code. Shared JSX/numeric-node dependencies, if encountered, will be
reported explicitly without editing the shared parser or driver.

Fourth-batch status: partial kernels pushed, not full source ports. See
[wave-29-fourth/REPORT.md](../wave-29-fourth/REPORT.md). Numeric manifests and
three compiling kernel mutants pass; shared numeric JSX source preparation,
checker binding acquisition and context memo/escape analysis are not integrated.
No further claims are taken in this continuation.

## Required regex contract after rebase onto c01907a7

The no-handwritten-matchers instruction supersedes the earlier configured
id-match completion: its native VM is removed. Default empty-pattern mode is
re-green; configured matching uses the required new RegExp(pattern, 'u') source
and is blocked by native nonconstant RegExp lowering and raw Go/JS dialect
incompatibility. No fallback or full configured parity is claimed. See
[REGEX_CONTRACT.md](../wave-29-configured/REGEX_CONTRACT.md).
All other pushed full/partial profiles re-green on current main. The fourth JSX
batch remains partial with source/binding/memo blockers; no new claims are taken.
