# Environment presence at sys.ts:1965:12

TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8. #baobab item 11.

Drop only the getEnvironmentVariable member-presence test from `if (sys && sys.getEnvironmentVariable)`, leaving `if (sys)`. This is JavaScript-changing and value-identical on the measured lane domain: every truthy sys holds this required System method as a function. The complete controlled lane oracle counted nine such visits, zero truthy values without a function, and zero falsy visits; 106,366 tests passed with only the sanctioned API row. No adaptation was constructed before that count completed.

The block passes sys to setCustomPollingValues and calls sys.getEnvironmentVariable("NODE_ENV") with its receiver intact to select the assertion level. Polling configuration reads the TSC_WATCH_POLLINGINTERVAL, TSC_WATCH_POLLINGCHUNKSIZE and TSC_WATCH_UNCHANGEDPOLLTHRESHOLDS low/medium/high environment settings. The next, separate block reads sys.debugMode. A labeled hypothetical sys lacking the method now enters the block and throws TypeError at the NODE_ENV call; it is outside the measured lane domain.

Ownership search covered owners.json, targets.json and every adapt.cjs. Existing sys.ts pins own other declarations, not this top-level if statement. owners.json pins that whole statement before and after with exact text and SHA-256. Parse errors, unknown owner text, half-adapted owners, duplicates and mixed line endings fail before writing. Replay writes nothing. Both commented and comment-free transpiled JavaScript must be byte-identical after accounting for exactly the authorized condition spelling; all other emitted text is unchanged.

`mutant.cjs TREE` restores exactly the original test for measurement. Proofs and raw evidence are on [typescript/env-presence-1965-evidence](https://github.com/system-inc/adamic/tree/typescript/env-presence-1965-evidence/stage3/census/env-presence-1965). Nothing is submitted to Loom.
