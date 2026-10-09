# any-value

Source: `src/compiler/utilities.ts:8147` in TypeScript 6.0.3 (`050880ce59e30b356b686bd3144efe24f875ebc8`), with main's registered stage3 adaptations applied.

compareDataObjects keeps its complete body and original dynamic parameter annotations. Exercise equal and unequal numeric properties. The real any parameters are deliberate negative fixture input, not a proposed adaptation.

Historical main census: **30** `a value of type any` sites (checker-rejected measurement).
Current fixture: **NotYet**, `a value of type any`.

[main.a](main.a) runs through `oracle/node.mjs` and prints exactly [expected.stdout](expected.stdout), with empty stderr and exit 0. The shared status records the full diagnostic including its line and column.
