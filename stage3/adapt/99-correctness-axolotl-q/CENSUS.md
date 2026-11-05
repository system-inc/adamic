# 99q compiler census

The ordinary unmodified main compiler at d40c2470e measured all 79 TypeScript
compiler roots independently on the complete accepted stack. The measured src
tree is byte-identical to full apply; comparison.json verifies every file.
No measurement overlay, refusal suppression, backend change or partial-root
fallback is used in this census.

Grouped by exact reason, with each TypeScript root counted once:

| Roots | Kind | First stop | Reason |
| ---: | --- | --- | --- |
| 73 | Refused | `src/compiler/core.ts:736:36` | a cast the runtime can't check |
| 2 | NotYet | `src/compiler/commandLineParser.ts:293:35` (1); `src/compiler/diagnosticInformationMap.generated.ts:13:45` (1) | reading an enum before its runtime initialization; move the call after the enum declaration |
| 2 | accepted | - | - |
| 1 | NotYet | `src/compiler/debug.ts:243:59` | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| 1 | NotYet | `src/compiler/semver.ts:71:9` | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |

The shared next wall is deduplicateSorted's existing double cast of emptyArray
at core.ts:736:36. It is outside this adaptation's accepted owners. H02 removes
the deduplicate return relations, but does not make these roots lower successfully.
The core.ts root reaches debug.ts:243:59. Root traversal order still matters.

The two accepted roots are corePublic.ts and hostErrors.ts. census-results.json
retains every exact root, kind, location, reason and timing. The compressed JSONL
and stderr log retain all raw rows. Extra non-TypeScript rows and the ordinary
whole-program row are excluded from the 79-root table, but remain in the evidence.
The whole-program row reports the ordinary single-entry limit: stage 0 compiles
a program from one entry file, got 79.

The historical 99m census used an earlier main compiler. A fresh d40c2470e baseline
binder probe reproduces the line-726 deduplicate wall; the complete fresh core.ts
measurement observes three remaining historical relations rather than four.
Do not attribute differences at other roots versus 99m to these edits alone.

Both 79-root Adamic types checks have zero diagnostics. The independent site
censuses and rejected trials are described in README.md and PROOF.md.
