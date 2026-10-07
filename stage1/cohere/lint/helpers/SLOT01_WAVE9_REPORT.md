Built: boundedQuantifierWidth, quantifierWidth and groupKindOf, one helper per .a file; twelve prerequisites removed across four rules, zero additional final blockers.
Commits: claim 9867cb69 pushed before code, ownership clarification 2f44fd1f, implementation 2029d7c39a234f220d7467c824ca79c5bd69492d; final report SHA in handoff.
Commands and outputs: setup PASS 26s, nproc 5; targeted differential PASS 64.543s; vet exit zero; uncached input oracle PASS 0.939s; format and diff logs empty.
Mutants: all three compile and finish successfully; actual Go catches missing required digits, omitted lazy suffix and omitted lookbehind classification.
Not covered: integrated regexp compilation, whole-rule findings, arbitrary invalid adapter arguments and full repository gate; shared rule harness, generator, rule entries and compiler untouched.

# Slot 01 ninth helper batch

## Landing and selection

Only codex/lint-helpers-01 has been pushed in this thread. Previous tip 32dbfd0 was already rebased on current main f8013f0baac41ddc340d76f83bddde38536a8f07 and green against the complete helper package: 66 tests, all 69 compiled mutant checks, 472.122s. A fresh main fetch before claiming and another before implementation commit confirm that main is unchanged and remains an ancestor. No rebase is required this batch. No main or area/ branch is pushed and no pull request is opened.

The wildcard branch fetch includes all eighteen origin codex/lint-helpers* branches and all seventeen claims files, including dash-suffixed wave workers. All larger concrete helpers were reserved; the five 24-consumer comment helpers remain reserved by the base HELPERS.md bundle. These three helpers tie the highest unclaimed concrete count at four consumers each. The group claim was pushed before source creation. No fourth helper is claimed.

An ownership refresh found a concurrent reservation by slot 05: boundedQuantifierWidth and groupKindOf in 94e0c374 at 04:37:54 UTC. Our original 9867cb69 at 04:37:46 UTC precedes it by eight seconds. Slot 01 retains the earlier ownership under the established earliest-claim rule, and publishes that clarification in 2f44fd1f. The ownership evidence records both claims rather than claiming that no duplicate mention exists. QuantifierWidth remains unique. See evidence/slot01-wave9/claims-final.json. No other worker files are edited.

## Consumers and readiness

Every helper serves the same four frozen remaining rules:

- @next/next/no-html-link-for-pages
- @typescript-eslint/no-empty-object-type
- no-restricted-exports
- no-restricted-imports

Three helpers times four rules removes twelve dependency entries. All four retain other blockers, so zero additional rules have their final helper blocker removed. This is prerequisite readiness under the common AST adapter assumption, not implemented rules or findings parity. slot01_wave9_readiness.json lists every residual dependency after this batch and after all retained slot 01 helpers. The original readiness inventory is unchanged.

## Go behavior and observation

The bounded scanner starts at slice byte one without checking for an opening brace. It requires lower digits, accepts an optional comma with zero or more upper digits, and requires a closing brace. It does not enforce bound magnitude/order. The quantifier scanner checks its opener, delegates braces to the bounded helper, and consumes exactly one lazy question mark. Group kinds use Go numeric enum values plain 0, lookahead 1 and lookbehind 2, with exact positive/negative anchored prefixes.

The adapters accept validated dense byte arenas with nonnegative integral slice offsets within or at their lengths. This preserves Go UTF-8 byte indexing, including continuation bytes. Arbitrary invalid number/offset inputs are outside the adapter contract. These are helpers, not rule listeners, and do not introduce rule relevance dispatch.

The oracle adds a virtual export file to Go cohere using an overlay, without editing its worktree. All four consumer test files contribute 918 complete literal inputs: 280 next, 47 TypeScript, 177 exports, 414 imports. Concatenations and local constant references are retained whole. Every source suffix including the empty suffix is queried. Additional controls cover all 256 first bytes, malformed braces, lazy suffixes, non-ASCII digits, huge bounds and positive/negative/named/incomplete group prefixes. Each helper has 27,783 queries, totaling 83,349 output lines. Actual private Go, source Node, sanitized native and emitted JavaScript match byte for byte.

## Commands and outputs

All tests write directly to logs; no test process is piped. The environment is sourced from /workspace/adamic-tools/env.sh: Go 1.27.1, clang 20.1.8, Node 24.19.0.

- bash cloud/setup.sh > /tmp/lint-helpers-01-wave9-setup.log 2>&1: go 0s, clang 0s, node 0s, submodules 0s, cache warm 26s, done 26s. nproc 5. Preserved in evidence/slot01-wave9/setup.log.
- go test ./stage1/cohere/lint/helpers -run '^TestSlot01Wave9' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/evidence/slot01-wave9/final.log 2>&1: PASS 64.543s; three tests and three successful compiled semantic mutants.
- go vet ./... > stage1/cohere/lint/helpers/evidence/slot01-wave9/vet.log 2>&1: exit zero, empty log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/evidence/slot01-wave9/oracle.log 2>&1: PASS 0.939s, six uncached probes.
- gofmt -l over all three new Go files: format.log empty. git diff --check: diff-check.log empty.

This batch uses the bounded new helper tests, repository vet and filtered uncached external oracle. The complete prior helper package was green immediately before this batch on the same main, and its sources/tests are unchanged. It was not rerun here; neither was the full repository gate. Prior mutant evidence remains in SLOT01_LANDING2_REPORT.md. Only the three new semantic mutants are credited in this batch. Successful native builds use ASan/UBSan and Linux leak checking; semantic-mutant output is considered only after compilation and exit zero with empty stderr.

## Every new compiling mutant

| Helper | Mutation | Independent actual-Go witness |
|---|---|---|
| boundedQuantifierWidth | remove required-lower-digit guard | output line 22: mutant 2, Go 0 |
| quantifierWidth | omit lazy-suffix consumption | output line 25795: mutant 1, Go 2 |
| groupKindOf | omit lookbehind-prefix branch | output line 25936: mutant plain 0, Go lookbehind 2 |

All three appear in final.log and were caught by stdout differences after successful native execution. No compiler failure, panic or sanitizer stderr is credited.

## Limits

This is bounded fixture/byte-control coverage, not exhaustive regexp fuzzing. Full regexp parsing, case folding and integrated rule findings are separately owned. Numeric adapter validation for arbitrary hostile JavaScript values is outside these helper contracts. No shared-harness gap blocks the standalone helpers. No shared registration, generator, test harness, rule directory or compiler source changed. The user has not yet named a shared Diagnostic landing SHA; no speculative rebase onto that branch is performed.
