# watch-functions-1532: stopped at implementation-read gate

Base: 8e3a784aa5a601f57342c07b68f145fe56252c21.
Upstream v6.0.3 verified: 050880ce59e30b356b686bd3144efe24f875ebc8.

The requested type-only widening is not justified. Upstream src/compiler/sys.ts:1030-1031 forwards the interval unchanged in DynamicPriorityPolling mode. The dynamic watcher forwards it again at :270. The queue selector at :349-358 handles only 250, 500 and 2000. At :361 the returned undefined is dereferenced for both an omitted interval and 123.

Ownership census: searched stage3/adapt owners.json, targets.json and every adapt.cjs. createSystemWatchFunctions is whole-pinned by 99-correctness-single-watcher-forwarder/owners.json. Its after text preserves the forwarding. createDynamicPriorityPollingWatchFile is whole-pinned by 99-correctness-timer-closures/owners.json. Its after text preserves the restricted selector and dereferences .files at owner-relative line 115. The counterexample extracts those exact selector/add functions from that after pin and removes only parameter annotations; it confirms both TypeErrors.

No adaptation or ledger change made. Per the task's explicit stop condition, census, restoring mutant, stock tsc, adamic types, apply, JavaScript comparison, declaration builds/diff and lane were not run. No implementation branch created or pushed. No Loom submission.

Reproduce: node watch-functions-1532-evidence/counterexample.cjs
