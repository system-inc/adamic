Rebased the eight surviving wave15 rules onto main 71d7e491, retaining lint-area bb2ece564 and its recovery classification.
Owned witnesses pass: 104634 bytes match Go, Node, emitted JavaScript and sanitized native; sources unchanged.
All eight semantic mutants are caught by output comparison on all three Adamic execution paths; TestMutants PASS 403.288s.
Broad TestRulesAgree explicitly refuses case-936/Octal.ts after recoveryRows, with expected semicolon at 10; parser recovery is outside this unit.
No new claims; full gate, the 17 required external checks and throughput were not rerun.

Commands run with /workspace/adamic-tools/env.sh sourced:

```sh
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree)$' -count=1 -v -timeout=20m > /tmp/wave15-cleanup-parity.log 2>&1
go test ./stage1/cohere/lint -run '^TestMutants/(module_name_ignored|optional_parameter_ignored|nexus-import-require-node-namespace_semantic|no-multi-str_semantic_mutant|no-nonoctal-decimal-escape_semantic_mutant|no-octal_semantic_mutant|structure-network-no-invalidate-cache-literal-key_semantic|structure-network-no-string-literal-query_semantic)$' -count=1 -v -timeout=20m > /tmp/wave15-cleanup-mutants.log 2>&1
```

Combined parity command exits 1: TestRulesAgree FAIL 275.43s, TestOwnedWitnesses PASS 33.68s, package 309.139s. The eight mutant subtests exit successfully and compare different output on Node, emitted JavaScript and sanitized native. Cold builds follow the explicitly requested Go cache cleanup.

The minimal raw input is Octal.ts.txt: `var a = 01.5;`. Retained minimal probe logs show Go exit 0 with noOctal at 8..10 and Node exit 70 even for an explicitly recovery-marked row. These two minimal logs are from the earlier classification investigation, not new runs; parity.log is the fresh full comparison after the area classification fix landed. Both lint_test.go and profile_test.go use recoveryRows. No classifier or shared harness changes were authored.

Cleanup: go clean -cache -testcache, the obsolete detached harness probe worktree and named worker scratch builds/test directories were removed. Pushed evidence and other workers' files were preserved. Writable capacity rose from zero to 25 GB on /workspace and 6.9 GB on /tmp immediately after cleanup.
