# Return, throw and yield routing

Each helper has a separate `.a` file. All three return immediately when the current block is unreachable. Node/block identities are opaque arena handles, not numeric AST kinds; -1 represents nil.

- makeReturn asks returnFrame for an index. A nonnegative index must name an existing frame: set returnedAny true, then link current to finallyEntry. With no frame, mark current final. Then make an unreachable continuation.
- makeThrow asks throwFrame for an index. For a valid frame, set thrownAny true before selecting throwTarget and linking. Otherwise mark current thrown. Then make an unreachable continuation.
- makeYield performs both return routing and throw routing in that order, keeping the same current block through both. It then allocates a fresh block, links current to it, and enters it. This records both ways the generator's caller can terminate it before ordinary continuation.

Dependencies preserve upstream behavior: frame lookups return -1 or a valid index and do not mutate the stack; throwTarget receives the actual frame; link, marking, allocation, entry and unreachable-continuation operations obey Go's graph contracts. Invalid dependency indices are refused, not a valid rule input. State uses a private frame array; no graph/node ownership cycles or shared harness models are introduced.

The temporary Go overlay keeps all three helper bodies unchanged, wrapping actual calls during Go CFG Build of every root in all 2,119 runtime sources from the four consumers. Dependency wrappers record only direct calls, suppressing nested backend calls, and annotate each call with current frame flags. Expected verdicts, frame mutations, frame lookup/target results and continuation identities come from actual Go. Test-only Adamic dependencies replay those backend observations and trace their calls. This proves these routing compositions and ordering; it does not port the dependency implementations or complete graph traversal.

Controls add generator/delegated-yield, nested-label, try/finally, return/throw roots, depths 0..3, all 64 catch/finally/position/prior-flag masks, and both reachable states for each helper. Full sources plus controls yield 2,466 verdict/trace lines. Go, source Node, emitted JavaScript and ASan/UBSan native agree. Twenty-four compiling mutants cover reachability, first-frame selection, flags, routing, block arguments and continuation steps. Mutants must exit 0 without stderr before changed output is counted.

Source /workspace/adamic-tools/env.sh, then run `go test ./stage1/cohere/lint/helpers/slot03 -run TestBatch16 -count=1 -v -timeout=20m` with direct file logging. Regenerate with `python3 stage1/cohere/lint/helpers/slot03/batch16/testdata/regenerate.py`. Exact consumers and remaining blockers are in readiness.json. No regex or new lint listener is added.
