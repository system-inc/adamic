# Paired Adamic census for #s8gy89r, family 1, cause U2

The comparison uses base `1a6b689e009ad27575887a18c938a6342a6cf2ca` and its direct child `793e2b25d267969957ecbb7ebfb0a6d551bade13` from `platforms/enum-tag-kind-helper`. Both were verified after fetching that named branch and `main` at `a59db917a3a401bd6712a703b94c821ff4864440`.

Stage 3 used TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Both checkouts initialized the same recursive submodule revisions. A file comparison of the applied directories, excluding Git metadata, was empty. The 312 adaptation checkpoint refs match, ending at tree `8f61e6699b44dd0e54523a760720a37acdd8cecb`. Both compilers read the base-applied tree.

This report is in progress. The completed measurement streams are published in [evidence](evidence/) as the runs finish.

## 1. Family 1 counts

Family 1 uses the exact reason “an object refinement using an open numeric enum as a literal tag”. Counts use the site identities in `stage3/census/pair`.

| Pass | Base distinct | Base raw | Head distinct | Head raw |
| --- | ---: | ---: | ---: | ---: |
| Compiler directory | 503 | 671 | Pending | Pending |
| `tsc.ts` entry | Pending | Pending | Pending | Pending |

The compiler-directory base recorded 2,983 findings across all reasons and zero panics. The historical 110-site cohort requires a per-site inventory to establish overlap.

## 5. Timing available so far

| Pass | Side | Process wall (s) | Total CPU (s) | Unit wall sum (s) | Unit CPU sum (s) | Timed units |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| Compiler directory | Base | 1,133.808 | 1,333.805 | 717.301 | 832.893 | 4,057 |

The process counters include checking, the source-file scans and stream output. The unit stream covers each attempted unit. Both latent passes are configured for CommonJSBundle, first-error mode, the scoped tail, and the no-output assertion.
