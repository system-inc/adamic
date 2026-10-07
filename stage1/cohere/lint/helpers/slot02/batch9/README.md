# CFG barrier and linking helpers

One Go helper per .a file. state.a supplies an acyclic numeric block arena, with -1 representing a nil pointer. Block identities, ordered successor IDs, reachability, incoming state, barrier values and nil-versus-present barrier storage are explicit. The barriers array is private graph state; callers must not retain independently resliced Go header aliases through this adapter.

- set_cycle_barrier.a: nil source and negative index are no-ops. A nil barrier slice stays nil for false, otherwise allocates false entries through successor length. An existing slice grows with false padding to successor length. The requested index is then replaced. Valid callers pass an index within the resulting slice; the Go out-of-range panic domain is outside semantic parity coverage.
- link_with_cycle_barrier.a: nil endpoint returns -1 without dependencies or state changes. Otherwise take the edge index before append, call appendSuccessor, then setCycleBarrier, then set target incoming when source is reachable. Duplicate and self edges are preserved. It returns the old successor count.
- link.a: forward both original identities exactly once to linkWithCycleBarrier with false; discard the numeric result.

appendSuccessor is an explicit external dependency. Its separately owned inline backing-array/capacity implementation is not ported here; the isolated adapter reproduces ordered successor values and pointer identity. The actual Go replay initializes the two inline slots before invoking the private methods. Capacity and externally retained slice-header alias behavior are not exposed by this contract. LinkWithCycleBarrier likewise receives the barrier setter explicitly, so registration can wire the delivered implementation without a shared harness edit.

The fixture workflow overlays the ordinary Go rule-testing entry point and typed fixture entry point, captures every consuming rule's runtime sources, and runs the Core and React rule packages. The standalone Go oracle parses every distinct source and uses real IndexRoots and Build with empty hooks to capture method-entry graph states. This includes roots beyond each rule's filtered listener selection; it is helper coverage over every consumer fixture, not whole-rule findings parity or a claim to have instrumented each consumer's exact filtered graph invocation. Instrumentation inserts capture/trace calls at method entry; the original method bodies remain unchanged. Replays call the real private methods with observed states and explicit controls.

The oracle captures all four consumers, deduplicates graph-state inputs, and exercises nil endpoints, self/duplicate edges, empty/nonempty storage, negative/zero/tail indices, lazy allocation, false padding, true/false replacement, reachable/unreachable sources and prior incoming values. Expected output is removed before Adamic reads inputs. Source Node, emitted JavaScript and ASan/UBSan native must finish without stderr and agree before comparison to Go. Mutants must also compile and finish; crashes do not count.

From the repository root, source /workspace/adamic-tools/env.sh, then run:

```
python3 stage1/cohere/lint/helpers/slot02/batch9/testdata/regenerate.py > /tmp/slot02-batch9-regeneration.log 2>&1
go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch9$' -count=1 -v -timeout=20m > /tmp/slot02-batch9.log 2>&1
```

No new rule, regexp matcher, shared registration, Diagnostic model or compiler implementation changes. See REPORT.md for complete gate results, mutant witnesses and limits.
