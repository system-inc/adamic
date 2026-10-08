# array-callback-return

Port of Go core array_callback_return.go and array_callback_return_callee.go: callback positions, logical/conditional/IIFE climbing, static method names, async/generator guards, all three options, ordered reports and suggestion edits. End reachability comes from control_flow_graph.Build with empty hooks and Graph.EndReachable; arenas are disposed as units. No checker facts are required. The mutation inverts the end-reachability judgment.

Validation: 288 unique upstream source/options cases / 316 recordings at cohere 7945d102a6c18dd36adf9114a758ce646e8b2359. Full discovered registry: 7,201 combinations / 14,996,773 identical bytes on Go, source Node, emitted JavaScript and ASan/UBSan native (TestRulesAgree, 314.23s). All owned witnesses agree across those backends. The end-reachability mutant is caught on Node and emitted JavaScript by the area harness, and on Node and sanitized native by the parallel rule-local TestNativeRuleMutant; mutated Node/native output is identical. No upstream case of this rule is excluded.

Reproduce with /workspace/adamic-tools/env.sh sourced: run testdata/upstream_count.py; go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses|TestMutants)$/^invert_CFG_end_reachability_' -count=1 -timeout=3h; go test ./stage1/cohere/lint/rules/array-callback-return -count=1 -timeout=3h.
