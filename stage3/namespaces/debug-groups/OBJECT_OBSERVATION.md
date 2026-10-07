Four computed namespace-object sites examined; live callable export slots remain NotYet.
Node prints live:disabled:false then live; a snapshot mutant prints different output.
The production container-guard mutant is caught by TestDebugNamespaceObservationBoundary, not by a build warning.
No sites accepted: this group remains 4/4 and the tracked total remains 13.
Narrow sound acceptance needs one live container, typed dynamic callable slots, identity and staged initialization; no language decision was made.

The exact program, source mutant and build diagnostic are in object_observation.a,
object_observation.json and object_observation.results.json. The original four
meter sites are debug.ts:169:49, 170:22, 189:56 and 190:14. They participate in
saving, replacing and restoring assertion functions. A snapshot plus static
qualified calls would disagree with Node. Supporting a namespace used as a value
must also preserve detached original functions and private function declarations.
This exceeds the original unit's small sound container boundary, so its existing
explicit NotYet remains. Casts through any in the unchanged upstream body remain
an independent soundness refusal.

Commands run with the configured Go/Node/clang toolchain:

```sh
python3 stage3/namespaces/debug-groups/run.py object_observation > /tmp/debug-group-object.log 2>&1
go test ./internal/lower -run '^TestDebugNamespaceObservationBoundary$' -count=1 > /tmp/debug-group-object-test.log 2>&1
```

The compiler mutant changes only the namespace-object observation guard to false,
runs the same test, and restores namespaces.go in finally. The test fails with
`Debug container boundary lost`, getting a later `reading Debug` NotYet instead
of the required explicit missing-container reason. The semantic source mutant
redirects computed accesses to a snapshot while keeping qualified calls live;
it is independently caught by source Node stdout. No native support is claimed.

Setup succeeds in 34.572 s: Go build 34.205 s, warm cache 34.526 s; nproc 5,
CPU quota 4. Current origin/main is already included. Only the feature branch
is pushed; the exact push SHA is recorded in the response and subsequent group ledger.
