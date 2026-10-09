TestNodeAgreement is defended by D1, a production parser-check deletion.
Only TestNodeAgreement fails; all 17 other current top-level tests pass.
Full-package uniqueness established in this session, with a clean baseline.

CODE UNDER TEST: Adamic internal/regexp parser Parse and its syntax routines, specifically parser.modifiers. ORACLE: live Node RegExp constructor success versus SyntaxError. Tests, Node script, corpus and harness stayed unchanged. No compiler/native product was mutated.

Starting origin/main is recorded in origin.txt. Current go test -list inventory has 18 top-level tests; TestUnicodeDecimalEscape was added since the audit. It was read and included in the full-package mutant matrix. No test vanished. Audit report-output.txt, functions.txt, plan.json, mutants.json and matrix.json were read and preserved as audit-*.

Coverage commands:
timeout 120 go test -count=1 -timeout 90s -run '^TestNodeAgreement$' -coverpkg=./internal/regexp -coverprofile=/tmp/defend-regexp/node.cover ./internal/regexp/ > node-coverage.log 2>&1
timeout 120 go test -count=1 -timeout 90s -run '^TestFlags$' -coverpkg=./internal/regexp -coverprofile=/tmp/defend-regexp/flags.cover ./internal/regexp/ > flags-coverage.log 2>&1
Node coverage 37.3%, flags 1.5%. Node-only covered blocks 353, flags-only zero, shared 14. Exact block spans retained in coverage-diff.json. TestFlags feeds only an empty pattern and the flags uu, uv and z. TestNodeAgreement exercises the full grammar through 5746 test262 and 630 generated cases, including invalid inline modifiers. The exclusive modifier validation branch is an actual grammar distinction, not a claim inferred from percentage alone.

D1 internal/regexp/parser.go:271 deletes the whole conditional statement rejecting c outside i/m/s or already present in seen. Modifier scanning and valid modifier application remain in place. This is the permitted drop-statement menu; no statement was inserted. Keeping writes to seen is harmless and Go vet accepts the standalone source. The diff applies to starting origin/main and the changed package compiled and executed.

Command:
ADAMIC_BUILD_CACHE_DIR=/tmp/defend-regexp/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./internal/regexp/ -run . > D1.log 2>&1

Failure:
oracle_test.go:79: 52 disagreements among 5746 test262 and 630 generated cases:
test/language/literals/regexp/early-err-arithmetic-modifiers-add-remove-i.js: pattern="(?i-i:a)" flags="" parser=true node=false error=<nil>

Node still rejects the pattern, while the mutant parser silently accepts it. Other examples include illegal letters and repeated modifiers. All 52 disagreement lines are preserved. D1.json and rows.json list the 17 passing rows: TestFlags, TestParse, the new decimal escape assertion, quantifier checks, canonicalization, all execution comparisons, provider/snapshot checks, step-limit tests and the built-in matcher witness. No skip or package panic occurred. Their production-precondition results are not used as proof of witness quality. The uniquely failing row is not part of the execution family, because it compares syntax acceptance rather than match output.

Only one aimed mutant was needed to prove a unique catch. No additional faults were planted after this proof. This is not a cost row and not a Node/native executor twin pair.

Costs and environment:
Warm env.sh worked; setup skipped, nproc=5. npm ci in stage3/api completed before baseline; oracle uses Node built-in fs. Clean whole-package baseline PASS, package line 0.854s (JSON event rounded 0.855). Coverage runs 0.176s and 0.003s. D1 full matrix 0.879 binary seconds, 1.298 wall seconds. Separate ADAMIC_BUILD_CACHE_DIR supplied, although a Go parser mutant needs no native product rebuild. No run approached 90s. Production source restored; no tests deleted, rewritten or weakened.

Disk check before all other work: /tmp 8.6GB free, /workspace 17GB. Removed only prior-unit /tmp/defend-unicode scratch and rechecked, still 8.6GB/17GB. /tmp total capacity is 8.8GB, so the requested 15GB threshold cannot be met there. No full-disk baseline occurred. Repository and tools were preserved.

Brief friction and limits:
No scope or oracle ambiguity blocked the work. The disk threshold is impossible on this /tmp filesystem. The audit's lone shared flag mutant did not exercise inline syntax checks, leaving an easy defense outside its sampled faults. TestNodeAgreement promises Node agreement for parser acceptance and its assertions do check that. It does not compare matching/captures, which are covered by distinct execution rows. All current package tests were run; repository-wide replay and the complete repository gate were not run. No main push or pull request.
