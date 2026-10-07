Built patternReads, patternWrites and isEs5ComponentCallStrict in separate .a files, serving eight rules and removing twelve prerequisites.
SHAs: published claim e6b36884 before source; preceding landing 2b45a51a; final rebased implementation/publication SHA reported in final response.
Commands and outputs before rebase: 34,828 four-way case rows PASS 73.445s; repository vet/format PASS; six-fixture input oracle PASS 1.036s; setup 38s, nproc 5.
Mutants: all eighteen compiling semantic variants caught; every exact Go divergence in evidence/mutants.json; final landing result recorded below.
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
