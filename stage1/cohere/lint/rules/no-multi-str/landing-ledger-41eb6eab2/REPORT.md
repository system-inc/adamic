Integrated published harness 41eb6eab2 and applied its dedup ledger, retaining eight active owned rules on main b8fb957a.
Removed four losing active ports; archived historical evidence under no-multi-str/retired-wave15 and preserved surviving support imports.
TestOwnedWitnesses passes in 19.465s with 59204 identical bytes on Go, Node, emitted JavaScript and sanitized native.
All eight surviving rule mutants pass the unified comparison test in 159.850s, caught only by differing output after successful execution on all three Adamic paths.
TestRulesAgree remains blocked by the shared parser at case-906/Octal.ts, expected semicolon at 10; no new claims, full gate or throughput proof.

Ledger winners retained: @next/next/no-assign-module-variable, @typescript-eslint/default-param-last, no-multi-str, no-nonoctal-decimal-escape, no-octal. Other unique owned rules retained: nexus/import-require-node-namespace, structure/network-no-invalidate-cache-literal-key, structure/network-no-string-literal-query. Losers retired: @typescript-eslint/no-unnecessary-type-constraint, @typescript-eslint/prefer-as-const, @typescript-eslint/prefer-enum-initializers (winner wave1-08), structure/tailwind-no-physical-direction (winner wave1-05). Their rule modules, descriptors, Go registrations, messages and mutant descriptors are removed. Historical evidence is not an active port; the retained results.a support module serves three surviving rules.

The published ledger identifies batch source branches but does not assign a batch-only rule to wave1-15. No batch-only rule was claimed.

Owned support bridges now use published reportRange and SuggestionEdit/Suggestion APIs to preserve independent edit ranges, explicit finding ranges and complete suggestions. No shared implementation was authored. Three merge conflicts used the exact published harness versions of lint.ts, lint_test.go and testdata/oracle.go. Main and harness are both ancestors of this branch. Existing .a modules and registry kinds names are retained.

Commands with /workspace/adamic-tools/env.sh sourced: go run ./cmd/lint-registry; go test ./stage1/cohere/lint -run ^TestOwnedWitnesses$ -count=1 -v -timeout=20m; go test ./stage1/cohere/lint -run the eight named TestMutants subtests -count=1 -v -timeout=20m; go test ./stage1/cohere/lint -run ^TestRulesAgree$ -count=1 -v -timeout=20m. Logs record every named mutant and backend. The broad run found 2256 unique source/rule/options combinations, records separate unsupported recovery inputs, then fails with an explicit parser refusal at case-906, before complete parity can be established. This shared parser boundary is outside owned rule scope and is the remaining parking blocker. The earlier multiple-fix failure concerns a now-retired rule copy.

Helper branch 739c5677 remains green and pushed on unchanged main b8fb957a; previous cache claim stays released because Mutex is still unavailable. No shared harness, parser or compiler fix is included.
