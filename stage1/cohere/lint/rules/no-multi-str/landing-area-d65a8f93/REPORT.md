Rebased the eight surviving owned rules onto lint area d65a8f931c98655936ae04c6899f38f14862b73e, containing current main 39638d9e.
The dedup ledger is unchanged; the five contested winners and three unique owned rules remain active, and four losing copies remain retired.
Unified TestOwnedWitnesses passes in 44.835s with 59204 identical bytes across Go, Node, emitted JavaScript and sanitized native.
All eight surviving semantic mutants are caught only by differing output after successful execution on all three Adamic paths; TestMutants passes in 202.887s.
No new claims: broad parity retains the recorded Octal.ts parser blocker; full repository gate and throughput were not rerun.

Commands with /workspace/adamic-tools/env.sh sourced: go test ./stage1/cohere/lint -run ^TestOwnedWitnesses$ -count=1 -v -timeout=20m, followed by the eight named TestMutants subtests with -count=1 -v -timeout=20m. Logs retained beside this report. Exact current area versions resolved the repeated three shared-file conflicts; no shared implementation was authored. The owned rule and bridge sources are unchanged from aacbf2b4. Runtime changes were accepted unchanged.

The area delta modifies heap.c, string_build_impl.h and string_search_impl.h plus tests and performance evidence. It does not change the parser, shared lint harness or dedup ledger. The last broad TestRulesAgree attempt remains recorded in landing-area-7481e032: explicit parser refusal at case-906/Octal.ts, expected semicolon at 10. That broad test was not rerun for this runtime-only delta, and full broad green is not claimed. The parser blocker is outside owned rule scope and outside the harness-only parking exception.
