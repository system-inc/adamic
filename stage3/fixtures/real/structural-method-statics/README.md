# structural-method-statics

Source: `src/compiler/commandLineParser.ts:2239` in TypeScript 6.0.3 (`050880ce59e30b356b686bd3144efe24f875ebc8`), with main's registered stage3 adaptations applied.

Extract the real readFile callback, adding its contextual string parameter annotation. Reader and Host supply a deterministic file reader. Reader.marker preserves the program-with-statics context required by this exact reason; it is harness context, not a tsc host implementation.

Historical main census: **80** `a method call through a structural signature in a program with statics; use typeof the declaring class` sites (checker-rejected measurement).
Current fixture: **NotYet**, `a method call through a structural signature in a program with statics; use typeof the declaring class`.

[main.a](main.a) runs through `oracle/node.mjs` and prints exactly [expected.stdout](expected.stdout), with empty stderr and exit 0. The shared status records the full diagnostic including its line and column.
