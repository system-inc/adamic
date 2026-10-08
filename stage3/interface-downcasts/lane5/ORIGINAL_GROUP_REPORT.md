Built: original FileWatcher.close declaration/call controls and Program.getCanonicalFileName callback handoff controls, with exact known-void certificates.
Commits: follows b6e53fb4; this group adds two candidate contract certifications.
Commands/results: Node/release native/sanitized native/JavaScript witness controls pass; complete callable source gate, allocation counts, focused compiler guards and vet recorded in logs/original-group.
Mutants: independently erase native and JavaScript arity guards; both FileWatcher and canonical-name wrong-arity witnesses fail semantically in each run.
Limits: 4 candidate pairs/34 reads certified in fixtures; 304 pairs/1469 reads remain; full Program declarations and whole-program reachability are not certified.

Read counts come from the static inventory, not from the Unknown-fallback upper
bound. FileWatcher.close is rank 24 with 15 candidate reads, first at
sys.ts:1237:21. The fixture preserves the original complete FileWatcher declaration
from sys.ts:1453 and the watcher.close() call form. The raw source slot is unknown,
so the view must establish the producer's callable shape at the reached read.

Program.getCanonicalFileName is rank 32 with 8 candidate reads, first at
builderState.ts:325:82. The fixture preserves the callback handoff to
getReferencedFiles and the original GetCanonicalFileName type alias from
core.ts:2375. Its Program receiver declaration and helper body are reduced to
this member and callback invocation. This certifies the callable member contract,
not the complete original Program interface or original getReferencedFiles body.

The independent pristine TypeScript checkout is pinned to
050880ce59e30b356b686bd3144efe24f875ebc8. verify-original-witnesses.py verifies
all 1503 inventory UTF-16 read spans against that checkout, without newline
normalization or cohere source copying. original-witnesses.json retains the first
upstream witness for every candidate pair. These witnesses establish source
provenance; they are not an exact runtime-reachability census.

Both pairs have valid call, wrong number, wrong arity and wrong result controls.
Every wrong callable stops before invocation, exit 70, with field, expected
function type and found value/signature pinned. Node's permissive mismatched
arity/result behavior is checked independently. Native positives pass leak checks.
Known void is metadata 254, distinct from unknown zero and discard wildcard 255;
void contracts are exact, and valued producers cannot satisfy them. Method producer
metadata uses the same known-void encoding. Eight new count rows are appended.

Higher-ranked demand is not bypassed by fabricating ordinary callback objects for
native string/Map methods. The intrinsic-member-read probe reproduces the language's
unbound-method refusal, even though Node can read that value. The highest Unknown
candidate, Debug.assert (439 reads), currently hits adamic/no-type-predicate in
its declared callback signature. Both boundary probes have Node controls and pins;
latest integration is rechecked before declaring either a blocking prerequisite.

Commands:

```
python3 stage3/interface-downcasts/lane5/verify-original-witnesses.py /path/to/pinned/typescript
python3 stage3/interface-downcasts/lane5/run-original-mutants.py
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 10m -args -update-counts
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedView(Callable|StoredMarker)' -count=1 -timeout 10m
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -run 'TestViewCallable|TestPrepareViewCallable|Test.*CallTarget' -count=1 -timeout 10m
go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
```

The full repository gate is not claimed. The first all-source run failed only
because the new count rows had not yet been recorded; the restored gate was rerun
after updating them. The original declaration correction and pending-boundary
path correction were also followed by unmutated reruns, not hidden from evidence.
