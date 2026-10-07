Landing refreshed with lint area's finding-edit serialization and captured-path fixes; no new helpers or claims.
Shared area e667e3e1 incorporated by merge 2a40af93, preserving exact current main 71d7e491 ancestry and all owned commits.
Affected lint oracle selection PASS 362.908s; setup PASS 56.847s, nproc 5; runtime RegExp reproducer still explicitly refused.
No new semantic mutant run or credit; all twelve batch13 mutants remain documented with their prior passing oracle evidence.
Stopped at the unchanged shared dynamic-pattern compiler gap; full repository gate, profile selection and other correctness suites not rerun.

The latest final fetch confirms current main 71d7e491b3c9724f7a0e2ee754592149e7f9790b and lint area e667e3e1dbdfd1b9125c3256961bfbc8ec31946b are ancestors. Area was merged rather than replaying 250 divergent commits, retaining exact main ancestry and the inherited changes without modifying shared files. Publish only codex/lint-helpers-02; no main or area push.

With /workspace/adamic-tools/env.sh sourced:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/slot02-required-typescript ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestCompilerAndStage1Agree|TestNodeTableIsLinkOnly)$' -count=1 -v -timeout=30m > /tmp/slot02-landing-e667-oracle.log 2>&1
```

Observed: TestRulesAgree PASS 140.32s, 13,071,542 identical bytes across Go, source Node, emitted JavaScript and sanitized native. TestCompilerAndStage1Agree PASS 167.59s across 452 files, 20,771,465 identical bytes using actual pinned TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. TestNodeTableIsLinkOnly PASS 54.95s across 2,154 rows, 13,087,427 identical bytes with and without unattached nodes. Package PASS 362.908s, no skipped tests. The inherited explicit parser recovery limitations remain visible in the complete log; this is not a claim of full recovery support.

bash cloud/setup.sh completed in 56.847s on five CPUs, Go 1.27.1, Node 24.19.0, clang 20.1.8. Setup, merge, oracle and final-fetch logs are losslessly archived and decompression-checked. Tests wrote directly to logs, never through a pipe.

Refreshed all twenty origin helper branches and read all nineteen claims. The next highest available concrete helper remains regexp.Compile with four consumers, after excluding the delivered comments bundle. It is not claimed by any helper branch. The required new RegExp(pattern, 'u') refuses a runtime source in internal/lower/regexp.go. The exact owned reproducer still exits one with stage 0 can't lower RegExp with a nonconstant pattern yet; see evidence/regexp.log.gz and ../regexp-compile-blocker/REPORT.md for Node success, the sanitized constant control and unchanged refusal test. No shared compiler workaround or handwritten matcher was introduced.

All thirty-nine prior helper claims remain completed and pushed. Their source/compiler inputs did not change in this eight-file shared lint merge, so the 154 previously caught helper mutants were not rerun or newly credited. The affected inherited lint selection was rerun. No new helper prerequisite was removed. No profile or full-repository gate claim is made. Per the user's instruction to stop at other blockers without editing shared files, work stops at this reproducer.
