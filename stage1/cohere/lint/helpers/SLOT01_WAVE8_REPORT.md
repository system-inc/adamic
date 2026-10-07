Built: decimalEscape, decodeLegacyOctal and classAtom.covers, one helper per .a file; 12 prerequisites removed across four rules, zero new final blockers removed.
Commits: claim 54e7be2 pushed before code; implementation 18f0d1ade6b43e7231f089558a1a3e70ae66cd8a; final report SHA is in the handoff.
Commands and outputs: setup PASS 28s, nproc 5; final targeted Go differential PASS 45.047s; vet exit 0; uncached input oracle PASS 1.096s; format and diff checks empty.
Mutants: all three compile and finish successfully, then Go stdout catches missing decimal saturation, widened octal digit budget and excluded range upper bound.
Not covered: whole-rule findings, shared regexp integration, arbitrary invalid adapter numeric/byte arguments and the full repository gate; no shared compiler, harness, generator or rule entries changed.

# Slot 01 eighth helper batch

## Landing and ownership

Only codex/lint-helpers-01 was pushed. Previous delivery 73adc26 was already green against its full helper package and uncached external oracle on main e8ba3d5d81de4d3773c723914fccd4c76248b965. A fresh main fetch confirms that same main remains an ancestor of this branch. No rebase is necessary. No main or area/ branch is pushed; no pull request is opened.

Wildcard fetch included every origin codex/lint-helpers* branch: 18 branches and 17 claim files, including dash-suffixed wave workers. All larger concrete dependencies were reserved. The five 24-consumer comment symbols are reserved in the base HELPERS.md parser-comments claim. The three selected regexp helpers tie the highest unclaimed concrete count at four each. Claim 54e7be2 was pushed before source creation. A second wildcard refresh confirms only slot 01 claims these symbols; see evidence/slot01-wave8/claims-final.json. No further helper is claimed.

## Rules and readiness

Each of the three helpers is listed by the same four frozen consumers:

- @next/next/no-html-link-for-pages
- @typescript-eslint/no-empty-object-type
- no-restricted-exports
- no-restricted-imports

Three times four yields 12 dependency entries removed, four distinct rules. All four still need other helpers. This is prerequisite coverage, conditional on the inventory's common AST adapter, not implemented rules or findings parity. slot01_wave8_readiness.json lists every remaining dependency both after this batch and after all retained slot 01 helpers. The original inventory is unchanged.

## Contracts and external observations

See SLOT01_WAVE8_README.md for APIs and caller preconditions. Decimal processing preserves the preceding-value threshold rather than clamping the new result: a run can retain 1234567, then count further digits without changing that retained value. Octal leading 0 through 3 permits three digits; 4 through 7 permits only two. Class coverage uses Go numeric enum discriminants, equality for rune/dash, inclusive range endpoints, and false for sets or unknown kinds.

A private Go overlay exports the actual three defining helpers. It changes no cohere worktree file. Consumer fixtures contribute 918 complete string inputs: 280 next, 47 TypeScript, 177 exports, 414 imports. The extractor preserves complete concatenations and resolves local constants. Every fixture byte offset is queried for decimal and octal; every fixture rune contributes kind/range queries for coverage. Identical coverage tuples are deduplicated without removing any distinct observation. Controls exercise all 256 leading bytes, saturation and full-width digit runs, octal suffix boundaries, Unicode bytes/runes, reversed ranges and int32 extremes.

Final query counts: 27,441 decimal, 26,508 octal, 2,388 distinct coverage. Outputs are respectively 54,882, 53,016 and 2,388 lines, totaling 110,286 Go output lines. Source Node, sanitized native and emitted JavaScript match byte for byte. Successful native builds use existing ASan/UBSan/leak-checking helpers; successful runs must exit zero with empty stderr. No new rule or registration is created, so there is no relevance dispatch to alter.

## Commands and outputs

All test commands write output directly to logs; no test process is piped. Environment is sourced from /workspace/adamic-tools/env.sh. Go 1.27.1, clang 20.1.8, Node 24.19.0.

- bash cloud/setup.sh > /tmp/lint-helpers-01-wave8-setup.log 2>&1: go 0s, clang 0s, node 0s, submodules 1s, build cache warm 28s, done 28s; nproc 5. Copy preserved in evidence/slot01-wave8/setup.log.
- go test ./stage1/cohere/lint/helpers -run '^TestSlot01Wave8' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/evidence/slot01-wave8/final.log 2>&1: PASS 45.047s, three tests, all three semantic mutants caught. This is the final generator after query deduplication.
- Initial same targeted command to targeted.log: PASS 92.907s with 598,884 repeated coverage queries before deduplication, the same three mutants caught. This supports the reduction; final.log is the final artifact.
- go vet ./... > stage1/cohere/lint/helpers/evidence/slot01-wave8/vet.log 2>&1: exit 0, empty log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/evidence/slot01-wave8/oracle.log 2>&1: PASS 1.096s, six uncached probes.
- gofmt -l over the three new Go files: empty format.log. git diff --check: empty diff-check.log.

This unit uses the bounded newly touched helper tests, repository vet and filtered uncached external oracle. The complete helper package was already green at 73adc26 in 420.872s with 63 tests and 66 mutant checks on the same main; those earlier sources and tests are unchanged. It was not rerun for this batch, and neither was the full repository gate. Historical mutant evidence remains in the prior reports; only the three new mutants are credited here.

## Compiling mutants and independent witnesses

Every credited mutant compiles to sanitized native, runs with exit zero and empty stderr, then differs from Go. Compiler failure, panic or sanitizer stderr is not credited.

| Helper | Mutation | Go witness in final.log |
|---|---|---|
| regexpDecimalEscape | replace value < 1048576 with true | line 51577: mutant 123456789, Go 1234567 |
| regexpLegacyOctal | allow three digits for leading 4 through 7 | line 49739: mutant 302, Go 37 |
| regexpClassAtomCovers | replace rune <= high with rune < high | line 6: mutant false, Go true |

## Limits

Fixture extraction is bounded, not an exhaustive regexp fuzzer. Full regexp compilation, case folding and whole-rule findings are separately owned and not exercised as integrated Adamic consumers here. The adapter must provide valid Go-shaped byte arenas and numeric inputs: nonnegative integral offsets, a present byte for octal, uint8 kinds, signed int32 runes. Arbitrary invalid JavaScript numeric inputs are outside these contracts. No shared-harness gap blocks these standalone helper APIs.
