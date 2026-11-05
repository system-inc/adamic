# 99l per-file compiler census

Compiler: fresh origin/main `99449b40bc54c2539b7b30f2840371b332957b0f`. Deliverable base:
`c2b6304cc949b68b1dbe72fbf3577a9039fb950d` (99k). Both dependency pins are in
compiler-provenance.json. Main's adaptation set is unchanged since 99k's
main base; before is main plus 99k and after adds only 99l.

The ordinary stage3/census/tool was built directly with go build from main.
No overlays, refusal suppression, backend, or partial-root fallback were used.
Both sequential streams include every one of the 79 compiler source roots,
non-source inputs, and the separate aggregate row. Non-source and aggregate
rows are retained in the raw streams and excluded from the 79-root denominator.

Exactly 73 baseline roots first stop at core.ts:381:26. All 73 move past it:

| Roots | Kind | New first stop | Reason |
| ---: | --- | --- | --- |
| 73 | Refused | src/compiler/core.ts:409:37 | a value of type U[] \| undefined seen as {}[] \| undefined, which can write {} where U is read |

These are first lowering stops after clean checking, not native compilation
successes. census-results.json records every root, including the other six.
Raw JSONL and driver logs are retained compressed beside this report.
