Built: isAlwaysTruthyTest in its own .a file, supplied numeric kinds and recursive parenthesis facts.
Commits: claim 7448c575e4144033ca729bb384572d8f693732ac; implementation follows this report on codex/lint-helpers-from-lint-wave1-12.
Checks: actual Go, source Node, emitted JavaScript and ASan/UBSan native agree on 30,458 observations and 182,218 bytes; focused package PASS 20.113s.
Mutants: zero-bigint row 30,003, skip-parentheses row 29,569 and empty-string row 57, each compiling and exiting cleanly on every backend before comparison catches it.
Uncovered: four CFG consumers lose one prerequisite each, zero final blockers; complete CFG/findings/fix integration and the full repository gate are not claimed.

API: isAlwaysTruthyTest(input: TruthyNode): boolean. TruthyNode holds presence, numeric kind, scanner text and a zero-or-one inner array for parentheses. The caller supplies the node facts; no helper parser lookup or kind-name comparison occurs. The adapter projects actual parser fields, including each parenthesis rather than pre-skipping it. It never includes an expected Go result in port inputs. The helper skips parentheses, accepts true and regex-literal kinds, uses canonical numeric text for zero, the delivered arbitrary-precision normalizer for BigInt zero, and string emptiness. Regex syntax is always truthy; no regex matcher is implemented.

The actual Go numeric domain is parenthesized expression 218, true keyword 111, regex literal 13, numeric literal 8, BigInt literal 9 and string literal 10. Cohere is pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db. Go expectations call the actual private helper through an oracle-only export overlay; the cohere worktree is unchanged. Every node in the 2,119 captured fixture sources is observed, plus 120 parsed boundary sources covering three parenthesis depths and 40 expressions: canonical numeric zeros, exponent/hex/separated forms, huge BigInts, Unicode/empty strings, regex nodes, templates, unary expressions, arrays, objects and function expressions. These are helper contract comparisons over consumer fixture ASTs, not complete native rule findings runs.

A measured discrepancy from the apparent nil branch matters: Go SkipParentheses(nil) panics before the later nil check executes. Therefore the port explicitly panics on absent input. A separate test runs actual Go and all three port backends with nil, asserting nonzero exit, no successful stdout and a diagnostic. Panic text/stack traces are backend-specific and are not claimed byte-identical. Normal successful streams are byte-identical. No silent false result is accepted for nil.

The initial fixture decoder incorrectly tried to put findings arrays in map[string]string; the decoder now selects only File/Source/Rule. Initial compiler diagnostics also enforced type-only import syntax and rejected a non-null assertion; both are corrected using a type import and checked access with panic. No failing attempt is credited as a passing mutant.

Selection checked 569 origin refs and all 20 distinct helper claim files. All greater-fan-out concrete helpers were reserved or delivered, and this four-consumer helper tied the maximum unclaimed count. Claim was pushed before code. Main remains c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06; the named shared harness ab70f38d4 has not landed there. Both own branches contain this main; the rule branch remains parked with its exact shared blockers documented. The existing exact nextBuildCount bigint-return lowering gap remains undelivered.

Commands, after sourcing /workspace/adamic-tools/env.sh:

```sh
go test ./stage1/cohere/lint/helpers/wave12 -run '^TestAlwaysTruthy' -count=1 -v > /tmp/wave12-truthy-oracle.log 2>&1
go test ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=20m > /tmp/wave12-truthy-full.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-truthy-vet.log 2>&1
```

Setup: Go/clang/Node/submodules ready 0s, cache warm 62s, total 62s; nproc 5, four-core quota, 17.6 GB. Only owned helper and claim files are changed. No shared harness edits, main/area pushes or new throughput claim.

Final owned package PASS 78.442s; vet clean. All six delivered helpers and their existing negative controls remain green after this new port.
