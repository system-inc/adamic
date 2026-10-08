Implemented census-only speculative recovery and depth tags in the existing latent overlay.
Base: `a5630a90`, including checker mapper fix `69501280` from the dispatched candidate.
Verified full/no-stubs totals: 19,168 sites, 10,482 NotYet and 8,686 Refused; output is byte-identical.
Speculative traversal completed 79 files; raw totals are complete, but depth tags failed independent verification.
This is a partial delivery: final depth tables and top 20 are withheld pending the binder.ts mismatch fix.

| Complete measurement | NotYet | Refused | Total | Verification |
|---|---:|---:|---:|---|
| Full baseline and no-stubs | 10,482 | 8,686 | 19,168 | Byte-identical across all 79 files |
| Speculative raw distinct sites | 52,596 | 13,650 | 66,246 | Traversal complete; depths unverified |

No files remain unfinished in the raw traversal. The walker reports zero unvisited nodes and 10,009,820 TypeScript source bytes examined. These coverage observations await the final independent result audit. All 79 files remain pending for the depth verification, listed individually in [depth-audit-pending.txt](depth-audit-pending.txt); binder.ts:926:17 is the first observed mismatch. No final depth or top-20 counts are claimed.

The 13:00 MDT deadline was missed. The earlier detached publisher failed before committing. This partial commit is explicitly authorized by the follow-up instruction. See [controls and every mutant](validation.md), [method](method.md), [control counts](counts.md), [partial result](PARTIAL.json), and [reproduction](reproduce.sh).
