Rebased eight surviving owned rules onto lint area b46914832d70e00847d82d5d221ab7bb24040c53, containing main c7991b90.
Unchanged dedup ledger applied; four losing active copies remain retired.
Expanded unified TestOwnedWitnesses passes in 27.085s, 103579 identical bytes across Go, Node, emitted JavaScript and sanitized native.
All eight owned semantic mutants are caught only by comparison on all three Adamic paths; TestMutants passes in 155.040s.
Broad TestRulesAgree fails in 49.269s at case-936/Octal.ts, expected semicolon at 10; no new claims, full gate or 17 required-check proof.

Reproducer: source /workspace/adamic-tools/env.sh and go test ./stage1/cohere/lint -run ^TestRulesAgree$ -count=1 -v -timeout=20m. Full failure log retained beside this report. Corpus expansion shifts the previously observed octal failure from case-906 to case-936. This remains an explicit parser refusal, not complete parity. No check was skipped, relaxed or removed.

Fresh green commands use the same environment and flags with ^TestOwnedWitnesses$ and the eight named TestMutants subtests. Logs record successful mutant execution and output differences. Shared merge conflicts took exact current area files. Owned source is unchanged; area changes to context, driver, tests and registrations are accepted unchanged. The parser blocker remains outside owned rule scope and the harness-only parking exception. No new helper claim is eligible under the landing cap.
