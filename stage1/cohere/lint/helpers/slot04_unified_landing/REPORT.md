Parked dynamic RegExp prerequisites by note; no blocked rule descriptor is owned or registered by this worker.
Landing implementation contains main 4e0bfda5 and lint integration bb2ece56; only the worker branch is pushed.
TestRulesAgree PASS 119.60s; TestOwnedWitnesses PASS 66.73s; all 40 registered mutants PASS.
Every mutant caught on Node, emitted JavaScript and sanitized native; no compiler refusal is mutant credit.
Helper-specific packages and the full repository gate were not rerun in this unified landing unit.

# Scope and parked work

This worker owns 66 completed helpers and no rule descriptor directories. PARKED.md names the four inventory consumers affected by unclaimed dynamic RegExp prerequisites, with the concrete .a reproducer and logs. There is no blocked rule module to move out of registration, and no unsupported regex implementation was delivered. Previous helper-specific verification and limitations remain in the earlier wave reports.

The branch was rebased without conflicts onto origin/area/stage1-lint bb2ece564842c4b2f909b9f75c27e74c2efa4f29, then merged current main 4e0bfda50a19c705a1aac0d9932e08483806d61c. The integrated witness-option transport is retained. No shared harness, registry or compiler implementation was edited. Main and area branches are untouched.

# Commands and evidence

Source /workspace/adamic-tools/env.sh before builds. Setup completed: Go 0s, clang 0s, Node 0s, submodules 0s, warm cache 100s, total 100s; nproc 5. Every test wrote directly to a log.

- go run ./cmd/lint-registry: PASS, registry-final.log.
- go test -count=1 -v -timeout=45m ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree)$': PASS 186.396s, final-witnesses-rules.log. Rule comparison agrees on 13053452 bytes; witnesses agree on 90589 bytes. Existing unsupported malformed-parser recovery is explicitly classified by the unchanged harness, not claimed as successful recovery parity.
- TestMutants ran with four disjoint filters recorded in group0.txt through group3.txt; all 40 descriptors appear exactly once in groups.json and all 40 subtests pass. final-mutants0.log: 528.071s; final-mutants1.log: 539.095s; final-mutants2.log: 500.553s; final-mutants3.log: 516.731s. Command for each: go test -count=1 -v -timeout=45m ./stage1/cohere/lint -run "$(cat groupN.txt)". Each temporary mutation compiles and runs successfully before ordinary Go comparison catches it on all three port runtimes.

The initial 20-minute combined sweep was interrupted because the complete mutant sweep needed a larger time budget. Two group runs were superseded when both protected bases advanced. Their logs are preserved under the names without final-, and are not final gate credit. No assertion, skip condition or mutant was changed. The 17 broader input-dependent comparisons and full repository gate were not selected or bypassed. The final response reports the pushed evidence commit once.
