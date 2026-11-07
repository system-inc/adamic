# arrayOf proof

Base: f0271c6a7b2e92900eb7046d8246cfb89f52cfb6.
TypeScript: 050880ce59e30b356b686bd3144efe24f875ebc8.

| Check | Result |
| --- | --- |
| Prior owner search | 0 matches before writing |
| Count before construction | 398 units, 860 non-negative integer calls, all other classes 0, counterexamples 0 |
| Whole lane and independent reconciliation | Pass; 106366 passes, 1 sanctioned failure, 0 skips; exact identities and baseline diff |
| Full apply on main | Exit 0, product packaged; 81504 inputs match measured baseline |
| Full apply on main plus arrayof | Child exit 0, product packaged; outer-shell discrepancy retained separately |
| Whole-owner and no-write contract | 11 checks pass; LF/CRLF, drift refusals, foreign edit, bytes and mtime |
| Paired replay of every adapter | 84 outcomes; 83 identical, arrayof changes one edit to zero edits |
| Ordinary adamic types before and after | All 79 sorted roots; both exit 0; zero diagnostics |
| Source comparison | 746 files; only the exact pinned arrayOf owner differs |
| Complete core JavaScript, comments kept and removed | Only arrayOf differs |
| All 9 emitted compiler JavaScript files | Only _tsc.js and typescript.js differ, only at arrayOf; all hunks displayed |
| typescript.d.ts and tsserverlibrary.d.ts | Both byte-identical |
| Full core.ts census | 271 units each; 59 sites before, 58 after; exactly F2 gone; 0 new sites anywhere |
| Original-owner mutant | Complete 59-site baseline restored, including core.ts:1339:12 |
| Archive integrity | Internal and external SHA-256 manifests; read-back passes |
| Process cleanup | All measurement, build, census and replay workers stopped |

JAVASCRIPT-HUNKS.md shows every nonempty source and emitted JavaScript hunk.
PROOF.json includes hashes, complete results, the count table and checkpoint.
The three packed evidence archives retain complete raw records and commands.
The failed packaging/layout preflights and the outer-shell status discrepancy
are recorded explicitly and excluded from successful command claims.
