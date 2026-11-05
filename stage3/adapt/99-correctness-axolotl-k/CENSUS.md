# Full per-file census for 99k

Pinned S2a compiler: `d231af0d34a4a145b14139dbc76d74f87e478b98`. Current main compiler: `6761c7374aa7169c73602b6a74bc281b879a640d`.
Both use cohere `06af2e399c6c1fb93fd570b774027942747400b4` and its checker `d92d9bfee114c80be2c375d72edae966176e3a4f`.

The ordinary meter driver, stage3/census/tool, was built directly with go build from each checkout. No lowering overlay, refusal suppression, backend, or partial-root fallback was used. Each run includes all 79 compiler source roots individually and the separate aggregate row. Non-source input rows and the aggregate are retained in the compressed raw streams and excluded from the 79-file denominator.

The before and after input trees use the initial main adaptation set. Only core.ts and builder.ts differ, through this adaptation. Main already contains d231af0d and a later S2a method fix, so the exact requested d231af0d binary was also measured independently. The current-main after results match the pinned S2a after results for every root.

Exactly **73 of 79** before roots first stop at core.ts:356:37 on the double cast. All 73 move past it. Their new first stops, grouped by reason:

| Roots | Kind | New first stop | Reason |
| ---: | --- | --- | --- |
| 73 | Refused | src/compiler/core.ts:381:26 | a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold |

The core.ts:381:26 call is in flatten, whose text this adaptation does not edit. These are observed lowering stops after a clean checker, not native compilation successes. census-results.json lists every root and all other first stops. The raw JSONL, driver logs, compiler hashes, commands, and dependency pins are retained beside this report.
