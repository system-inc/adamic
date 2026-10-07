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
