Built: normalizeBigIntLiteral in its own .a file, preserving actual Go arbitrary-precision normalization and invalid fallback.
Commits: claim d5dd996cbfaba9911b393467b8c43376e9be1add was pushed before implementation; rule parking 565a0d45349fec52f1c7d6de73a8cd4449a81d50; compiler base f8013f0baac41ddc340d76f83bddde38536a8f07.
Checks: 10,983 rows / 1,003,998 UTF-16 observation bytes match actual Go, source Node, emitted JavaScript and ASan/UBSan native; tests PASS 10.944s and vet clean.
Mutants: hex-digit-value, negative-zero and invalid-fallback all compile/run; only Go output comparison catches each on every backend (rows 2137, 2124 and 2152).
Uncovered: removes one dependency for four rules, zero final blockers; original consumer tests make no normalization calls; whole CFG/rule integration, arbitrary invalid UTF-8 and the full gate are not claimed.

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
