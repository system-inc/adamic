# function-without-body

Source: `src/compiler/path.ts:310` in TypeScript 6.0.3 (`050880ce59e30b356b686bd3144efe24f875ebc8`), with main's registered stage3 adaptations applied.

Keep the string overload and its implementation at lines 310 and 312. Omit the branded Path overload and documentation. Helpers implement only the POSIX inputs in the driver; no URL or drive path is exercised.

Historical main census: **191** `a function without a body` sites (checker-rejected measurement).
Current fixture: **NotYet**, `a function without a body`.

[main.a](main.a) runs through `oracle/node.mjs` and prints exactly [expected.stdout](expected.stdout), with empty stderr and exit 0. The shared status records the full diagnostic including its line and column.
