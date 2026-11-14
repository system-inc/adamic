# Price sys public — measurement checkpoint

| Site | Road | Target gone | New findings | Declaration diff lines | Calls added per lane run |
|---|---|---|---|---|---|
| 3 | site-local | Not measured | Not measured | Not measured | Not measured |
| 3 | declaration | Not measured | Not measured | Not measured | Not measured |
| 10 | site-local | Not measured | Not measured | Not measured | Not measured |
| 10 | declaration | Not measured | Not measured | Not measured | Not measured |

Baseline: adamic `8e3a784aa5a601f57342c07b68f145fe56252c21`, detached private worktree. Upstream v6.0.3 confirmed at `050880ce59e30b356b686bd3144efe24f875ebc8`. Recursive shallow submodules initialized; stage3/api npm ci succeeded. Baseline apply and census binary preparation are running. No adaptation or production compiler change is included.

Upstream sys.ts:1380 is in `patchWriteFileEnsuringDirectory`, not the missing-entry watcher. It captures sys.writeFile, replaces sys.writeFile, and subsequently calls the captured function with `.call(sys, ...)`. An arrow reading the same object's writeFile at call time would recurse. There is no identity comparison of originalWriteFile in this function. Census and lane conclusions remain unmeasured.
