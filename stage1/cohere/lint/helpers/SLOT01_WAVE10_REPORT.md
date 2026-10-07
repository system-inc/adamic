Built: wordClassAtoms, nonWordClassAtoms and wordBoundary, one helper per .a file; twelve prerequisites removed across four rules, zero additional final blockers.
Commits: claim 74dfcf68 pushed before code, ownership clarification 64469738, implementation 683871983742a9f047fae24b33e9723953bd4177; final report SHA in handoff.
Commands and outputs: setup PASS 25s, nproc 5; targeted seven-test differential PASS 28.442s; vet exit zero; uncached input oracle PASS 1.017s; format and diff checks empty.
Mutants: seven compile and finish successfully, then actual Go catches flag conjunction, complement endpoint, boundary polarity, two cached-result aliases, dropped forwarded flag and duplicate callback.
Not covered: integrated regexp compilation, whole-rule findings, arbitrary dependency callbacks and full repository gate; shared harness, generator, rules and compiler untouched.

# Slot 01 tenth helper batch

## Landing and ownership

Only codex/lint-helpers-01 has been pushed in this thread. The preceding tip 8981d1af is already based on current main f8013f0baac41ddc340d76f83bddde38536a8f07, with all existing helpers tested and pushed. A fresh fetch before claiming and before final implementation confirms main is unchanged and an ancestor. The full earlier landing gate on this main passed 66 tests and 69 compiled mutant checks in 472.122s; the ninth batch then added three passing scanners. No rebase is necessary for this unit. No main or area/ branch is pushed and no pull request is opened.

The wildcard fetch covered all eighteen origin codex/lint-helpers* branches and every claims file. The shared HELPERS.md comment bundle remains reserved, and all larger concrete helpers are reserved. These three tie the largest unclaimed concrete count at four each. Group claim 74dfcf68 was pushed before source creation.

Refresh revealed slot 04 later reserving wordBoundary in f2d65497 at 04:49:03 UTC. Our 74dfcf68 at 04:48:20 UTC precedes it by forty-three seconds and retains ownership under the established earliest-claim rule. The clarification was published in 64469738. The two atom helpers remain uniquely mentioned. The evidence records both boundary reservations and timestamps rather than claiming a duplicate mention does not exist. No other worker files are edited. No fourth helper is claimed.

## Rules and readiness

Every helper lists the same four frozen consumers:

- @next/next/no-html-link-for-pages
- @typescript-eslint/no-empty-object-type
- no-restricted-exports
- no-restricted-imports

Three times four removes twelve prerequisite entries. All four rules still have other blockers, so zero additional rules lose their final listed helper blocker. This is conditional helper readiness, not implemented rules or findings parity. slot01_wave10_readiness.json lists every residual dependency after this batch and after all retained slot 01 helpers. The original readiness ledger is unchanged.

## Contracts and actual Go observations

Word atoms preserve ordered ASCII ranges and underscore. Only combined ignoreCase and unicode adds LONG S and KELVIN SIGN. Non-word atoms preserve every complement gap, singleton representation and inclusive Unicode MaxRune endpoint. Both atom helpers return fresh mutable lists and records, independent of previous output mutations. Boundaries call the separately owned wordCharacters dependency once with the original four flags, then construct the exact positive/negative Go lookaround string. The dependency returns Go-compatible class spelling; production does not replace its semantics with a guessed table. See SLOT01_WAVE10_README.md for APIs and representation.

The actual private Go helpers are exposed through virtual overlay files, with cohere's worktree untouched. A second observational overlay replaces the single boundary wordCharacters call with a wrapper that records flags and delegates to that actual function. This changes only observation, not returned behavior. Every boundary comparison includes returned text, callback count and forwarded flags.

All sixteen combinations of ignoreCase, unicode, multiline and dotAll are exercised, and both negations for boundaries. Every atom query records the first result, mutates its first atom to low 99999 and checks a later result against Go's fresh output. Numeric enum kind, low/high and text fields are all compared. The four fixture files contribute 918 complete literal expressions: 280 next, 47 TypeScript, 177 exports, 414 imports. These helpers have only flag/negation inputs: strings determine corpus repetitions, not invented string arguments. Five fixed corpus controls bring the total to 923 iterations over the full flag domain.

Distinct three-helper case counts are 14,768 word, 14,768 non-word and 29,536 boundary: 59,072 queries. Output counts are 295,360, 354,432 and 88,608, totaling 738,400 Go output lines for one pass of each helper. Additional mutant tests repeat those same observations. Go, source Node, sanitized native and emitted JavaScript match byte for byte. This is complete coverage of the finite flag/negation argument combinations under the wordCharacters dependency contract, not exhaustive arbitrary callback or integrated regexp coverage.

## Commands and observed output

All tests write directly to logs without pipes. source /workspace/adamic-tools/env.sh supplies Go 1.27.1, clang 20.1.8 and Node 24.19.0.

- bash cloud/setup.sh > /tmp/lint-helpers-01-wave10-setup.log 2>&1: go 0s, clang 0s, node 0s, submodules 0s, cache warm 25s, done 25s. nproc 5. Copy: evidence/slot01-wave10/setup.log.
- go test ./stage1/cohere/lint/helpers -run '^TestSlot01Wave10' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/evidence/slot01-wave10/final.log 2>&1: PASS 28.442s, seven tests and seven compiled semantic mutants. Source Node/native/emitted JavaScript all match actual Go.
- The initial three-test run PASS 12.521s is retained in initial.log. Four additional mutants then proved freshness and callback observation checks can fail. final.log is the final artifact.
- go vet ./... > stage1/cohere/lint/helpers/evidence/slot01-wave10/vet.log 2>&1: final rerun exit zero, empty log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/evidence/slot01-wave10/oracle.log 2>&1: PASS 1.017s, six uncached probes.
- gofmt -l over all new Go files: format.log empty. git diff --check: diff-check.log empty.

This unit runs bounded new helper tests, repository vet and a filtered uncached external oracle. Earlier helper sources/tests are unchanged and their successful same-main gates remain prior evidence. The full helper package and full repository gate were not rerun for this batch. Prior mutants remain documented in the landing and per-batch reports; only the seven new checks are credited here.

## Every new compiling mutant

Every credited mutant compiles and runs successfully with exit zero and empty stderr, then differs from Go stdout. A panic, compiler rejection or sanitizer error is not counted. Native builds use ASan/UBSan and Linux leak checking.

| Helper | Mutation | Go witness in final.log |
|---|---|---|
| wordClassAtoms | change ignoreCase AND unicode to OR | line 19: mutant 6 atoms, Go 4 |
| nonWordClassAtoms | truncate final complement at U+FFFF | line 10: mutant high 65535, Go 1114111 |
| wordBoundary | invert negated branch | line 1: mutant negative expression, Go positive |
| wordClassAtoms | return cached list and records | line 11: mutant low 99999, Go 48 |
| nonWordClassAtoms | return cached list and records | line 13: mutant low 99999, Go 0 |
| wordBoundary | clear forwarded ignoreCase | line 9: mutant flags 0, Go 1, with output text unchanged |
| wordBoundary | call dependency twice | line 2: mutant count 2, Go 1, with output text unchanged |

## Limits

Whole-rule findings and integration with separately owned regexp compilation/case-folding helpers are not claimed. The boundary dependency must follow the wordCharacters contract; arbitrary callbacks that mutate options or return unrelated strings are outside that contract. No shared-harness gap blocks these standalone APIs. No rule entry or dispatcher is created, and the shared harness, generator, compiler and rule directories are untouched. No shared Diagnostic landing SHA has been supplied; no speculative rebase onto that branch is performed.
