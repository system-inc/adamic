# value-condition

Source: `src/compiler/utilities.ts:10157` in TypeScript 6.0.3 (`050880ce59e30b356b686bd3144efe24f875ebc8`), with main's registered stage3 adaptations applied.

numberOfDirectorySeparators keeps its complete body. The regexp match array is used as a truthy condition; exercise both matching and nonmatching strings.

Historical main census: **56** `a value as a condition` sites (checker-rejected measurement).
Current fixture: **Refused**, `a value as a condition`.

[main.a](main.a) runs through `oracle/node.mjs` and prints exactly [expected.stdout](expected.stdout), with empty stderr and exit 0. The shared status records the full diagnostic including its line and column.
