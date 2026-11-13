# SourceFileLike position member

TypeScript pin: 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3).

Retype the internal optional `SourceFileLike.getPositionOfLineAndCharacter` method as an optional function-valued property, explicitly unioned with undefined. Preserve its JSDoc, `@internal`, parameter names, optional parameters, and return type exactly. The only runtime readers in src are scanner.ts:469 (presence) and :470 (receiver-preserving call).

`71-writable-views` owns only `lineMap`, not the complete interface; this is a new adaptation. `owners.json` pins the entire interface before and after all earlier adaptations. Stock TypeScript 6.0.3 parses it; unknown owner text, duplicate owners, parse errors, and half-adapted text fail before any write. A second application writes nothing. Other source bytes and line endings are preserved. Both commented and comment-free JavaScript emits are checked before writing.

`mutant.cjs TREE` restores the optional method on that same pinned owner. It is a measurement mutant, not part of normal apply.

Commands, logs, full censuses, the upstream declaration builds, API guard, type check, restoring-mutant observation and final lane result are retained on [the evidence branch](https://github.com/system-inc/adamic/tree/typescript/sourcefile-like-position-evidence/stage3/census/sourcefile-like-position). No evidence products belong in this directory. Nothing is submitted to Loom.
