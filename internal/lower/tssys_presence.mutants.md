# ts.sys's presence tests: mutants (#p2n17c4 T1)

Each is one edit, applied alone, confirmed with `git diff`, run against the tests it names, then put back. Lower's
tests are in internal/lower/tssys_presence_test.go, over two programs. TestTsSysSelfReadingLiteralStops pins tsc's
own shape, internal/oracle/testdata/tssys_presence, at its not yet: createDirectory reads the literal it's a method
of, a function value that captures the variable its own initializer declares (#dv99xzy). TestTsSysPresenceLowers
holds its twin, internal/oracle/testdata/tssys_presence_twin, whose createDirectory decides without the literal, to
lowering with every presence test admitted, and the oracle's internal/oracle/tssys_presence_test.go holds the twin
to Node. The twin is each mutant's witness. M-a and M-b are the design's (tssys-design.md, section 3). Neither has
run yet: each "Red" is what the code says until a run quotes it.

## M-a. Road 3 dropped: a function that never runs collected anyway

File: internal/lower/view_flow.go, collectAllocationFlowGraph.
Old: `		if !graph.running[i] {`
New: `		if false && !graph.running[i] {`
Run: `go test ./internal/lower -run '^TestTsSys' -count=1 -v`
Red: TestTsSysPresenceLowers, both rows, stops with a not yet at the first presence test judged, "whose receiver may
hold a value the program didn't build as a plain object literal (local or result has no tracked producer)": setSys,
which nothing calls, has its write `sys = s` collected, and its parameter s has no caller to give it a producer.
Pin 1: nothing. TestTsSysSelfReadingLiteralStops stays green, since tsc's shape stops at #dv99xzy while its
function is lowered, before the flow graph the mutant changes is built.

## M-b. ClosureTargets ignoring Direct

File: internal/ir/call_targets.go, ClosureTargets.
Old: `		if call.Direct > 0 {`
New: `		if false && call.Direct > 0 {`
Run: as M-a.
Red, by the design: in the twin, the arrow's call of getNodeSystem, nested in it, goes opaque, and the presence tests
stop with "opaque closure". By reading this branch it stays green: expression.go's nested call gives the arrow's
call the carrier MakeClosure{getNodeSystem}, not ClosureSelf, because the arrow's NestedParent (none) isn't
getNodeSystem's (the arrow), and ClosureTargets bounds a MakeClosure without Direct. Direct alone bounds a call only
where the carrier is ClosureSelf, a nested declaration calling a sibling in the same parent, which tsc's shape
doesn't have. wantNestedGetNodeSystemCall logs whether ClosureTargets bounds the call with its Direct taken away, so
the run says which. Pin 1: nothing. TestTsSysSelfReadingLiteralStops stays green, since tsc's shape stops at
#dv99xzy, a check made as the literal's capture is lowered, which asks nothing of ClosureTargets.
