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
