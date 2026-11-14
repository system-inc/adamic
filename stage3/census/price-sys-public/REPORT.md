# Price sys public — measurement checkpoint

| Site | Road | Target gone | New findings | Declaration diff lines | Calls added per lane run |
|---|---|---|---|---|---|
| 3 | site-local (`bind`) | Yes | 0 | 0 | Counting pending |
| 3 | declaration | Not measured | Not measured | Not measured | Not measured |
| 10 | site-local | Not measured | Not measured | Not measured | Not measured |
| 10 | declaration | Not measured | Not measured | Not measured | Not measured |

Baseline: adamic `8e3a784aa5a601f57342c07b68f145fe56252c21`, detached private worktree. Upstream v6.0.3 confirmed at `050880ce59e30b356b686bd3144efe24f875ebc8`. Recursive shallow submodules initialized; stage3/api npm ci succeeded. Baseline apply and census binary preparation are running. No adaptation or production compiler change is included.

Upstream sys.ts:1380 is in `patchWriteFileEnsuringDirectory`, not the missing-entry watcher. It captures sys.writeFile, replaces sys.writeFile, and subsequently calls the captured function with `.call(sys, ...)`. An arrow reading the same object's writeFile at call time would recurse. There is no identity comparison of originalWriteFile in this function. Census and lane conclusions remain unmeasured.

## Upstream declaration experiment

Pristine upstream `npm ci` and `node_modules/.bin/hereby dts` succeeded. Each public member conversion was built separately, resetting source between variants. The generated `typescript.d.ts` has exactly one removed and one added line per road; `tsserverlibrary.d.ts` is byte-identical (empty diff). Exact diffs are adjacent to this report. These are pristine upstream measurements, not yet the adapted-tree declaration builds or 79-root checks.

The resolved upstream reader inventory is in `readers-upstream.json`: 526 explicit member accesses/destructuring entries, including homonyms; 35 resolve directly to the two requested public member declarations. This inventory includes harness and src test runners. Dynamic computed keys and structural implementations require further audit; this checkpoint does not claim a complete reader proof.

## Adapted baseline control

Baseline apply exited 0. Coordinates match the request: sys.ts:1385:31 and sys.ts:1612:77. Stock tsc 6.0.3 CLI passes the 79 compiler roots (zero diagnostics) with pinned Node ambient types. An initial stock API setup omitted those ambient types and was corrected; its missing-Node diagnostics are not compiler regressions.

The full single-file CommonJSBundle census exited 0 in 350.052 seconds, using LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1 and LATENT_ONLY_FILE selecting the complete sys.ts. It reports zero checker diagnostics and checker_rejected:false. Its historical generic measurement label does not override that header. Both requested refusal sites are present. There are 185 raw finding events, 59 unique non-Boundary findings: 41 NotYet and 18 Refused. The control stream, timings, apply log, exact tsc config/result and source manifest are retained. This is a census, not successful native compilation.

`READERS.md` lists the 35 resolved public-member references on the adapted baseline; the full 526-entry inventory includes homonyms and the harness/tests.

## Site 3 local form

Private candidate: `const originalWriteFile = sys.writeFile.bind(sys);`; the subsequent `.call(sys, ...)` remains unchanged. Its complete sys.ts census exits 0 (228.687 seconds): 58 unique findings, 41 NotYet and 17 Refused. Exactly the target refusal is gone; no finding is new or changed. Stock tsc over 79 roots exits 0. The one-line private candidate diff and full raw census are retained.

The source captures writeFile once, immediately replaces sys.writeFile, and later invokes originalWriteFile.call(sys, ...). There is no identity comparison or other escape of originalWriteFile. A live-receiver arrow would look up the replacement and recurse. Bind captures the function and receiver at the original read point, creates a new function identity and performs a bind call then. The witness retains receiver-sensitive calls after a later replacement and confirms the identity difference and naive-arrow recursion. This is a measured candidate, not a proposed adaptation. The lane count and broader observation review remain pending.
