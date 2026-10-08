# Debugger behavior decision

One syntax site and one census refusal remain at src/compiler/debug.ts:198:9.
Current main refuses the minimal `debugger;` program, so this kind has not landed.

Deleting Debug.fail's debugger changes behavior with an attached inspector. The
proof extracts the actual source function, runs it on Node with Debugger enabled,
then runs the same body with only `debugger;` deleted. The original pauses once;
the deletion mutant never pauses. Both throw the same Debug Failure. probe error.
This rules out claiming complete behavior preservation from a normal oracle run.

Minimal program for @system_adamic_typescript:

```a
debugger;
console.log("done");
```

Without an attached debugger, both versions print done. With an attached
debugger, only the original has its deliberate breakpoint. Decide whether
preserving compiler results may explicitly exclude inspector breakpoints before
removing this site. The adapter keeps it. This report sends no external message.

```sh
source /workspace/adamic-tools/env.sh
NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules" node stage3/adapt/41-explicit-any-remaining/debugger-proof.cjs MAIN_TREE > debugger-proof.log 2>&1
```

Observed counts: one before, one after, zero adapted. The deleting-debugger
mutant is caught by the inspector pause-count comparison. No native inspector
claim is made.
