Retained all seven canonical takeover rules on current area b28757f33 and main 71d7e491; no rule source changes needed.
Fresh evidence committed on codex/lint-wave1-06-land; final pushed sha reported alongside wave1-07.
TestMutants PASS 211.986s; subsequent TestOwnedWitnesses and TestRulesAgree PASS 98.005s.
All seven semantic mutants caught by independent Go comparison on source Node, emitted JavaScript and sanitized native.
Full registry mutants and full repository gate not run; inherited explicit parser-recovery exclusions remain outside coverage.

Current branch includes area b28757f339f3253ed3796a4dd6138d9ee736d887 and current main 71d7e491b3c9724f7a0e2ee754592149e7f9790b. This turn changes only evidence. All seven descriptors already have node:true, handed-node visits and the required typed options adapters. No shared file was authored or weakened.

Commands, sourced /workspace/adamic-tools/env.sh, output directly to committed logs:
go test ./stage1/cohere/lint -run '^TestMutants$/^(wrong-diagnostic-id(#01)?|wrong-global-type-name|omit-empty-export-fix|wrong-enclosing-scope|configured-restriction-ignored|report-left-reference)$' -count=1 -v -timeout 20m
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree)$' -count=1 -v -timeout 15m

Mutants: wrong-diagnostic-id (Next beforeInteractive), wrong-diagnostic-id#01 (Next css tags), wrong-global-type-name, omit-empty-export-fix, wrong-enclosing-scope, configured-restriction-ignored, report-left-reference. Each executes successfully and differs from upstream Go on all three runtimes. Final rule parity: 13,188,475 identical bytes; positive witnesses: 120,417 identical bytes. Original wave1-07's formerly surviving option-gated mutants were repaired independently in its own directories. No new helper or rule claim; no push to main or area.
