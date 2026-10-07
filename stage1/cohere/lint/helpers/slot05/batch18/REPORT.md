Built patternReads, patternWrites and isEs5ComponentCallStrict in separate .a files, serving eight rules and removing twelve prerequisites.
SHAs: published claim e6b36884 before source; preceding landing 2b45a51a; final rebased implementation/publication SHA reported in final response.
Commands and outputs on final area: all eighteen retained packages PASS; new trio 34,828 rows PASS 75.188s; vet/format PASS; input oracle PASS 12.049s; final setup 151s, nproc 5.
Mutants: all 179 retained compiling semantic variants caught again, including eighteen new ones; every exact Go divergence in evidence/landing-mutants.json.
Not covered: complete rule diagnostics, external dependency implementation parity, invalid adapter/AST representations, full repository gate and its seventeen required external-input correctness checks.

# Ownership and readiness

All 47 previous helpers were complete, tested and pushed before this claim. At reservation main 39638d9e and area d65a8f93 were current and ancestors; all twenty helper branches were refreshed and every claim inspected. Higher fan-out comments leaves remain owned by the shared bundle. The three selected concrete symbols tie the highest unclaimed count at four consumers. Claim e6b36884 was pushed before any batch18 source was written. No fourth helper is claimed.

Both CFG helpers serve array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. The strict React helper serves react/no-direct-mutation-state, react/no-this-in-sfc, react/prefer-es6-class and react/require-render-return. Twelve prerequisite occurrences removed across eight consumers and no final blocker removed. readiness.json calculates the cumulative owned inventory of fifty helpers under the frozen cohort assumptions; it is not rule finding parity.

# Observations before rebase

Pinned real Go methods and dependencies are executed through temporary overlays, with only dependency call names renamed. Node values and wrapper children have stable arena IDs so callbacks reveal identity and evaluation order. Callback classification mirrors each actual Go switch; nil inputs must not query it. Expression dependencies execute normally but their nested instrumentation is excluded from the caller's trace. README.md describes the exact seam. No compiler, production cohere, shared harness or rule registry edits were made.

For each CFG helper the corpus inspects 166 array-callback-return literals, 272 consistent-return, 71 no-unreachable-loop and 555 rules-of-hooks; 801 distinct strings after controls and 12,508 nodes including nil. It includes eight as expressions, two non-null, thirty-five parenthesized, two satisfies and one type assertion. There are 3,570 actual throwable-fork cases. Strict factory corpus: 160, 215, 163 and 223 occurrences respectively in its four consumer rules, 630 distinct strings and 9,812 nodes including nil, with 51 positive verdicts. Its wrapper counts are also archived. All five kinds are required by corpus assertions.

The first complete gate passed fifteen variants in 71.752s. The strengthened gate added direct fork omission, reversed throwability and skipped parentheses: all eighteen compiling variants caught, total 34,828 rows across Go, source Node, emitted JavaScript and sanitized native, PASS 73.445s. Sanitized runs must exit zero with empty stderr; compile failures, refusals, panics or sanitizer findings do not count as semantic catches. No test in this bounded run skipped.

Mutants: reads ignore nil, omit the read, omit the fork, invert throwability, omit its query, test/fork before reading, stop at wrappers or send default expressions to read; writes ignore nil, omit writing, stop at wrappers or write non-identifiers; strict factory accepts rejected initial inputs, omits parentheses skipping, ignores direct identifiers, ignores properties, or substitutes createClass in either final name test. The fork omission is caught at output line 20 against a real Go fork call. Each exact first mismatch is archived.

# Commands

```
bash cloud/setup.sh > /tmp/lint05-batch18-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH18_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch18/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch18 -count=1 -v -timeout=20m > /tmp/lint05-batch18-final.log 2>&1
go vet ./... > /tmp/lint05-batch18-final-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05/batch18 > /tmp/lint05-batch18-format.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch18-oracle.log 2>&1
```

Initial setup timings: Go, clang, Node and submodules ready 0s; cache warm 38s; done 38s on five processors, cpu.max 400000 100000, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Vet and format logs are empty; input probes report zero cache hits and six misses. Logs were written directly, never piped.

# Landing first after upstream advanced

Main advanced to c7991b90 and the area merged it at b84a9d93 while the reserved trio was under test. New lowering/proof and native record runtime changes are genuine test inputs, so earlier retained results cannot be carried by source identity. Complete this trio, commit locally, rebase onto the requested current area (which contains main), then rerun the full retained helper selection and every semantic mutant before publishing the owned branch. No further helper will be claimed in this landing unit. The final observations below replace the provisional base for publication; earlier logs remain named before-rebase evidence.

# Final landing observations

Rebase onto origin/area/stage1-lint b84a9d9314b65d3d0261ee017e233287b4f071da completed cleanly, replaying all 79 branch commits. This area contains current main c7991b900362796aefd111474e65eb5398e91953. New lowering/proof and native record changes were accepted without edits or reversions. Active trio source commit is 9774dc8dc2196a8a4dc7f56a86aad97fd6d52e0f, active rebased claim f04dc3df (original claim e6b36884 was pushed before source). Git object IDs show the entire retained slot05 tree is unchanged by rebase. Upstream compiler/runtime inputs changed, so no earlier observation was carried forward without rerunning.

Complete uncached retained gate: all eighteen actual packages PASS on this tested head. All fifty owned helpers and the inherited shared helpers were reverified. Every one of the 179 retained compiling semantic mutants was caught again. Every witness is preserved in landing-mutants.json, and package-results records all eighteen timings. The new trio's final package PASS 75.188s on the new compiler/runtime. New trio row count remains 34,828; every wrapper and 3,570 throwable cases are covered. All consumer and corpus metadata is unchanged. This proves bounded helper parity, not full rule execution.

Final setup: Go, clang, Node and submodules ready 0s; cache warm 151s; done 151s on five processors, cpu.max 400000 100000, 17.6 GB. Repository-wide vet and formatting logs are empty. Filtered uncached Node input oracle PASS 12.049s, six misses and zero hits. No test in the bounded gate skipped. The full repository gate and its seventeen required external-input correctness checks were not run; none was skipped, relaxed, deleted or credited.

All twenty helper branches and every claim were scanned again after the landing gate. No competing claim names any of the trio. Main and area were refreshed and remained c7991b90 and b84a9d93. All existing claims are complete. Only codex/lint-helpers-05 is published, with an exact force-with-lease against the original pushed claim e6b3688400156b2af9ded8ffdc4d99a78432b896; no main or area branch is pushed.

```
git rebase origin/area/stage1-lint > /tmp/lint05-batch18-rebase.log 2>&1
bash cloud/setup.sh > /tmp/lint05-batch18-landing-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test -p 2 ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot05/... -count=1 -v -timeout=30m > /tmp/lint05-batch18-landing-helpers.log 2>&1
go vet ./... > /tmp/lint05-batch18-landing-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05/batch18 > /tmp/lint05-batch18-landing-format.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch18-landing-oracle.log 2>&1
```
