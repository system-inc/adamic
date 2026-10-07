Built: three separate .a helpers: normalizeBigIntLiteral, isBreakableStatement and the CFG newBlock allocator.
Commits: normalization 183b04f397b488620fc821aad73a3bec959974d3; predicate b37b3eb3ee0a56687e4a681a807b4696c0983b56; allocator claim bdbaf59f25a2ba92555a99f7a3a8a02b634db009; rule parking 565a0d45349fec52f1c7d6de73a8cd4449a81d50.
Checks: complete owned helper package PASS 52.416s and vet clean; three new helpers match actual Go on 67,982 observations / 2,133,286 bytes across source Node, emitted JavaScript and sanitized native.
Mutants: all six new semantic controls and eight existing controls caught; each new mutant compiles and exits cleanly before failing actual Go comparison on every backend.
Uncovered: removes three dependencies for four CFG consumers, zero final blockers; exact nextBuildCount remains blocked at bigint return lowering; shared rule integration and the full repository gate remain uncovered.

API: normalizeBigIntLiteral(text: string): string. It removes one trailing lowercase n, preserves the Go base-zero integer grammar including signed octal and prefix underscore rules, converts arbitrarily large integers with a little-endian decimal digit array, and returns the exact suffix-stripped text on invalid input. No bigint compiler primitive or shared harness change is required. It is independent of the existing nextBuildCount blocker, which remains explicit and undelivered.

Actual Go's private function supplies every expected answer through an oracle-only export overlay; cohere's worktree is not edited. Consumer capture runs every named Test function in core/array_callback_return_test.go, core/consistent_return_test.go, core/no_unreachable_loop_test.go and react/rules_of_hooks_test.go. All tests pass. Captured unique rule/file/source records: array-callback-return 251, consistent-return 75, no-unreachable-loop 1,624, react-hooks/rules-of-hooks 169. These are the four frozen readiness consumers. The tested original fixtures invoke the normalization function zero times; the wrapper's empty trace is an observation, not proof of literal-path coverage. Their 2,119 complete source texts are also passed through the helper as invalid/fallback controls, together with 8,864 explicit valid/malformed/arbitrary-size controls. This proves the helper contract on those inputs, not complete CFG or findings parity.

Controls cover every base prefix, signs, underscore boundary placement, legacy octal, invalid digits, lowercase/uppercase hex, zero and negative zero, 2^53/64/127/256 boundaries, 257-bit binary, long hexadecimal, 1,024-digit decimal, exhaustive short malformed strings and seeded generated spellings. Observation streams encode UTF-16 units so fallback Unicode does not get hidden by terminal output.

The initial capture incorrectly assumed the upstream trace file would exist even when there were no calls. It failed after all consumer tests passed; because the shell lacked set -e, a comparison attempt then failed on the missing generated corpus. The capture now handles an empty trace explicitly and the rerun uses set -e. Neither initial failure receives passing credit; both logs are retained.

Commands, after sourcing /workspace/adamic-tools/env.sh:

```sh
python3 stage1/cohere/lint/helpers/wave12/testdata/capture_control_flow.py > /tmp/wave12-cfg-capture.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/wave12 -run '^TestBigIntNormalization' -count=1 -v -timeout=15m > /tmp/wave12-cfg-bigint.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-cfg-vet.log 2>&1
```

The capture script derives the exact test-name lists from those four files and writes separate full consumer logs under evidence/cfg-consumers-*.log. Source and emitted JavaScript run on Node; native runs with ASan and UBSan. Each mutant is compiled separately, exits cleanly with no stderr, then fails the actual-Go output comparison. New Adamic files use .a.

Setup: Go/clang/Node/submodules ready at 0s, cache warm 40s, total 40s; nproc 5, four-core quota and 17.6 GB. Cohere is pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db. Current main and the compiler are unchanged from the previous F801-LANDING.md gate. Those previous rule and helper results remain recorded, not rerun or relabeled as complete integration.

Selection inspected 531 fetched origin refs and all 17 unique helper claim variants. Comments are already delivered under the older HELPERS.md claim. All greater-fan-out concrete symbols were reserved, so this four-consumer helper tied the maximum. The claim and selection ledger precede code. Only this unit's helper directory and claim were edited; no main or area branch push. No new throughput measurement or final rule readiness is inferred.

## Numeric breakable predicate

isBreakableStatement(node: BreakableNode): boolean reads a supplied {present, kind} fact record. A false presence bit denotes Go nil; the kind is the pinned parser's numeric SyntaxKind. It reads no kind names and performs no parser lookup. The exact actual Go domain is switch 256, while 248, do 247, for 249, for-in 250, for-of 251. The caller can supply the numeric field of its already-read node.

Claim 9a4b8818e3c06918f3f6718dd11090534806c401 was pushed before code, after refreshing 541 origin refs and 19 distinct helper claims. This helper serves the same four consumers listed above and removes no final blocker alone. The adapter parses all 2,119 captured sources with actual Go and visits every AST node; filename extensions preserve the fixture script kind, while relative names receive an absolute normalized /corpus prefix required by the parser. It projects only presence and kind. Controls include every numeric kind through KindCount plus one and nil paired with every kind. No Go result is placed in the facts given to Adamic.

30,267 consumer AST/kind/nil observations produce 179,169 identical bytes on actual Go, source Node, emitted JavaScript and ASan/UBSan native. PASS 9.422s. Two mutants compile and exit cleanly, then only Go comparison catches missing-switch at row 446 and nil-loop at row 1 on all three backends. The earlier bigint suite was also rerun after expanding the owned Go export overlay and passed; vet is clean.

Unsuccessful attempts remain in evidence: the adapter initially rejected relative source filenames; its first fix shadowed the path package; the observation entry initially logged a boolean against the string-only console contract; and the first numeric constants were wrong because the manual enum-line count missed declarations. Actual Go/source Node comparison caught the numeric error at row 381. Constants were then taken from the actual Go numeric observation. None of those failed attempts is credited as a passing mutant.

Commands, after sourcing the setup environment:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/wave12 -run '^(TestBreakableStatementMatchesGo|TestBigIntNormalization)' -count=1 -v -timeout=15m > /tmp/wave12-cfg-breakable-initial.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/wave12 -run '^TestBreakableStatementMatchesGo$' -count=1 -v -timeout=15m > /tmp/wave12-cfg-breakable.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-cfg-breakable-vet.log 2>&1
```

The earlier bigint passing lines are preserved in cfg-breakable-initial.log even though that command ultimately failed in the new adapter. The final focused predicate check passes. This is helper-level parity using actual Go AST facts, not an independent Adamic parser or complete CFG/rule findings claim. The source helpers are separate .a files; all new harness code remains under this owned helper directory.

## CFG block allocation

newBlock<E>(builder: BlockBuilder<E>): ControlFlowBlock<E> allocates from eight-object chunks, sets the next index, appends the same object to the ordered block list and returns it. Each chunk slot is distinct, zero-state arrays are independent, and later event writes remain visible through returned/stored aliases without leaking to other slots. The source retains the recursive successors type and generic events. Actual loading, lowering and both backends accept this recursive type; no compiler gap is inferred merely from recursion. This allocates empty graph blocks and does not construct cyclic successor graphs or prove the whole graph builder.

Claim bdbaf59f was pushed before code after checking 548 origin refs and 19 unique claim blobs. The public source is control_flow_new_block.a, with its separate observation entry block_main.a. Initially written under gaps while primitive support was unknown, the candidate was moved to the active directory only after four-backend comparison passed. No gap implementation for this allocator remains. The type-only BlockBuilder import must be marked import type for Node's source runner; the initial ordinary import failed and its log is preserved. A provisional gap check then correctly failed because the primitive lowered, and was replaced by positive backend comparison.

Actual Go consumer fixture families were rerun with an oracle-only wrapper around the private newBlock method. All 2,119 captured consumer inputs pass their original assertions. The trace contains all block counts 1 through 163, plus zero as a control, for 164 tested construction sizes. Go's actual method is invoked independently in the oracle; observations record index, ordered block count, chunk count, reachable/hasIncoming/final/thrown flags, empty event/successor/barrier state, returned/stored identity, fresh-slot identity and subsequent built-string event reads through stored aliases. No hand-authored expected allocator replaces Go.

The complete 26,732 allocation/state/identity observations yield 950,119 identical bytes on actual Go, source Node, emitted JavaScript and ASan/UBSan native. The alias mutant changes chunk[index % 8] to chunk[0]. It compiles and exits cleanly on every backend; only Go output comparison catches row 4. String events are constructed at runtime. The fixture tests are not a native complete CFG or consumer findings/fix run. Go nil slices are represented as empty arrays for the tested operations; slice capacity/nil introspection, private inline successor-buffer aliasing and int32 index overflow beyond the observed allocation sizes are not claimed.

Commands, after sourcing /workspace/adamic-tools/env.sh:

```sh
ADAMIC_CFG_CAPTURE_HELPER=block python3 stage1/cohere/lint/helpers/wave12/testdata/capture_control_flow.py > /tmp/wave12-cfg-block-capture.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=20m > /tmp/wave12-cfg-full.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-cfg-final-vet.log 2>&1
```

The final complete owned package passes in 52.416s. Fresh component times: bigint normalization 2.82s plus mutants 6.47s; block allocator 5.99s; numeric predicate including mutants 8.78s. Existing Tailwind helpers and their controls also pass. The exact counter gap check confirms Node still matches actual Go's nine signed-64-bit observations while lowering refuses gaps/atomic-counter.a:4:10 with stage 0 can't lower a function returning bigint yet. Its compiling number approximation is caught at row 5 on all three backends. The counter remains explicitly undelivered. This is an observed primitive boundary, not proof that every alternate representation is impossible. Work stops there without changing shared compiler/harness files or claiming a seventh helper. Other helpers remain unclaimed; no exhaustion claim is made.

New helper dependency handoff: array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks each lose the three new symbols from their frozen readiness helper lists. All four retain other listed prerequisites, so zero final blockers are removed. See control_flow_readiness.json for the owned residual lists. The separate existing Tailwind helpers still serve their six listed consumers. Shared readiness.json, registration and comparison files are unchanged. No new throughput sample or complete repository gate is claimed.
