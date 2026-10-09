# non-null-assertion

Source: `src/compiler/core.ts:1098` in TypeScript 6.0.3 (`050880ce59e30b356b686bd3144efe24f875ebc8`), with main's registered stage3 adaptations applied.

first keeps the adapted body, including the indexed non-null assertion introduced by the strict-index adaptation. Debug.assert is a small checking harness; both inputs are nonempty.

Historical main census: **484** `the non-null assertion !` sites (checker-rejected measurement).
Current fixture: **Refused**, `the non-null assertion !`.

[main.a](main.a) runs through `oracle/node.mjs` and prints exactly [expected.stdout](expected.stdout), with empty stderr and exit 0. The shared status records the full diagnostic including its line and column.
