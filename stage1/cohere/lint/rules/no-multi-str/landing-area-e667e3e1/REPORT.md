Integrated current lint area e667e3e1 with current main 71d7e491 retained; ledger unchanged and no owned rule source change.
TestOwnedWitnesses PASS: 127500 bytes identical on Go, Node, emitted JavaScript and sanitized native (33.41s).
TestRulesAgree FAIL (67.93s) and TestNodeTableIsLinkOnly FAIL (30.62s): explicit parser-recovery refusal for case-936/repository/source/Octal.ts, semicolon offset 10.
The combined parity command exits 1 in 131.965s; no check was skipped, relaxed or edited.
No new helper claimed: landing remains blocked by the documented Adamic parser-recovery gap; full gate, 17 external checks and throughput remain unverified.

Command: source /workspace/adamic-tools/env.sh; go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestNodeTable.*)$' -count=1 -v -timeout=20m. Full log is parity.log. Setup used GOPROXY=https://proxy.golang.org|direct and passed on 5 CPUs; timing lines are preserved in setup.log. The minimal reproducer remains var a = 01.5; followed by a newline, as recorded in landing-main-71d7e491. Go recovers; Adamic still refuses after recovery classification.

All eight owned TestMutants subtests PASS after successful compile and execution on Node, emitted JavaScript and sanitized native: module_name_ignored, optional_parameter_ignored, nexus-import-require-node-namespace_semantic, no-multi-str_semantic_mutant, no-nonoctal-decimal-escape_semantic_mutant, no-octal_semantic_mutant, structure-network-no-invalidate-cache-literal-key_semantic and structure-network-no-string-literal-query_semantic. Each is caught by output comparison. Command: go test ./stage1/cohere/lint -run TestMutants/<alternation of those eight names>$ -count=1 -v -timeout=20m; full raw output and exact elapsed times are in mutants.log.
