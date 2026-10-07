Built ImportedNameOf, isDecimalDigit and countGroups in .a files, removing 12 dependencies across eight rules.
Commits: retained import port 33e3346, replacement claim 69bdfe7, final implementation 79450869f7172776861559992221366c573395f3; main base e8ba3d5 remains an ancestor.
Commands: helper gate PASS 54.468s, 3,370,439 Go queries; vet/types/format PASS; uncached input oracle PASS 1.451s; setup 27s, nproc 5.
Mutants: eight compiled and exited zero without stderr, then differed from real Go; every witness is listed below.
Not covered: full repository test gate, complete rule findings/fixes/suggestions, invalid UTF-8 Go strings, corrupted AST arenas or exact panic wording.

# Ninth helper batch report

## Ownership and landing

Branch: codex/lint-helpers-05. The only branch owned and pushed by this unit contains current origin/main e8ba3d5d81de4d3773c723914fccd4c76248b965. The prior 23 helpers were already complete, green against that unchanged main, and pushed at 28f93fd. No rebase was needed. No main or area branch is pushed and no pull request is opened.

The first claim f846de1f640356fda6f11d263dfe3cf4c8ec922f was pushed before code. A later refresh discovered slot 02 d9b239af40dcf9632c3d55b2bbfb4b3474454ecb at 03:54:07 UTC, two seconds before our 03:54:09 claim, covering CallExpressionSource and HasAttributeNamed. Both duplicates were withdrawn and their .a files removed. Their local passing comparisons are historical evidence only and are not counted as delivery or readiness.

ImportedNameOf was finished, tested and pushed at 33e3346 before claiming replacements. Refetched all 18 origin codex/lint-helpers* branches and checked their claims. isDecimalDigit and countGroups tied the highest unclaimed concrete count at four consumers each. Claim 69bdfe7 preceded replacement code. The final refreshed scan finds all three retained symbols only on slot 05. See evidence/final-claims.json for current-main ancestry, all-branch ownership and the race timestamps. The shared comments bundle remains reserved and the generic strict-options ledger label remains rule-local work.

## Delivered contracts

- imported_name_of.a: nil and wrong-kind decline; a present PropertyName wins over Name, including empty and decoded string aliases; missing names answer empty. Numeric AST edges preserve identity without reference cycles.
- regexp_is_decimal_digit.a: ASCII 0 through 9 at the exact Go byte offset; EOF and overshoot answer false. Negative offsets refuse like Go indexing. UTF-8 length and byte access preserve supplementary-character offsets.
- regexp_count_groups.a: skip one escaped rune, bracket class state, ordinary capture count, named (?< captures and both lookbehind exclusions. Terminal escapes and unfinished syntax reproduce Go's scanner without adding syntax validation.

Actual cohere 715ba94f3608a6500086b1076ce5cb7e51b836db provides every expected answer. The test-only Go overlay adds exports for the two private regexp helpers; it does not modify the submodule. Every consumer's nonempty Go test-file string literal is included, with per-consumer counts in evidence/*-coverage.json. Literal counts include source, configuration, expected messages and other test strings; they are not a claim of whole-rule diagnostic replay.

| Helper | Consumer samples and controls | Actual Go queries |
|---|---:|---:|
| ImportedNameOf | 287 AST samples, including one factory control sample | 4,093 |
| isDecimalDigit | 1,165 strings at every byte offset through EOF + 2, plus all Unicode scalars at byte zero and EOF | 2,253,117 |
| countGroups | 1,165 strings plus all Unicode scalars escaped before capture controls | 1,113,229 |

The scalar sweeps cover all 1,112,064 scalar values, excluding surrogate code points. AST tests query every node and nil, include alias/string-name controls, and preserve present empty property names separately from absent ones. Invalid UTF-8 in the Go corpus is explicitly refused before JSON serialization can replace it. Corpus hashes are recorded in evidence.

## Rules and readiness

ImportedNameOf removes a prerequisite for react-hooks/config, react-hooks/incompatible-library, structure/network-no-forbidden-import and structure/react-component-no-forward-ref.

Both regexp helpers remove a prerequisite for @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports.

Observed dependency ledger: 12 occurrences across eight distinct rules, zero new final blockers removed by this batch alone. Cumulative slot 05 delivery: 26 uniquely retained helpers, 199 dependency occurrences across 66 rules. The four previously removed final blockers remain the same; baseline helper readiness 46 plus those four gives 50. CONSUMERS.md lists each helper's rules; readiness.json lists every affected rule's remaining dependencies. These are helper-readiness calculations conditional on the common AST adapter, not observations of complete rule parity. The withdrawn source helper's final-blocker effect is not counted here.

## Commands and outputs

All test commands wrote directly to log files; no test output was piped. Source /workspace/adamic-tools/env.sh before Go commands. Toolchain: Go 1.27.1, clang 20.1.8 and Node 24.19.0.

Setup command: bash cloud/setup.sh > /tmp/lint05-batch9-setup.log 2>&1. Exact timing lines in evidence/setup.log:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (27s)
setup: done in 27s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

nproc printed 5. The cgroup quota is four cores.

- ADAMIC_SLOT05_BATCH9_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch9/evidence" go test ./stage1/cohere/lint/helpers/slot05/batch9 -count=1 -v -timeout=20m > /tmp/lint05-batch9-helpers-green.log 2>&1: PASS 54.468s. Real Go, source Node, emitted JavaScript and sanitized native agree. All eight semantic mutants caught. Retained per-test times 6.86s, 12.25s and 35.35s; evidence/helpers.log.
- go vet ./... > /tmp/lint05-batch9-all-vet.log 2>&1: exit 0, empty log; evidence/vet.log.
- go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch9/main.a > /tmp/lint05-batch9-green-types.log 2>&1: exit 0; evidence/types.log.
- gofmt -l stage1/cohere/lint/helpers/slot05/batch9 > /tmp/lint05-batch9-green-format.log: exit 0, empty log; evidence/format.log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > /tmp/lint05-batch9-oracle.log 2>&1: PASS 1.451s, six fixtures, zero probe cache hits and six misses; evidence/oracle.log.

The full repository test gate was not run. The touched package, repository-wide vet and filtered uncached external oracle form the bounded gate. Native comparisons use ASan/UBSan and default Linux leak checking; every successful run must exit zero with empty stderr.

## Every delivered mutant

Temporary copies only. A compile failure, panic, nonzero exit or stderr fails the harness and is not credited. After successful native execution, byte comparisons against real Go catch these eight semantic changes. The harness's reported verdict indexes are output-line indexes, not query numbers.

| Helper | Mutation | First output line | Mutant / Go |
|---|---|---:|---|
| ImportedNameOf | ignore PropertyName | 1654 | wrap / forwardRef |
| ImportedNameOf | treat present empty PropertyName as absent | 1720 | Empty / empty string |
| isDecimalDigit | exclude 9 | 2504 | false / true |
| isDecimalDigit | use UTF-16 length for a byte-offset boundary | 28909 | false / true |
| countGroups | count parentheses inside character classes | 727 | 2 / 1 |
| countGroups | count lookbehind as named capture | 147 | 1 / 0 |
| countGroups | skip only the backslash | 5 | 1 / 0 |
| countGroups | skip one byte instead of one escaped rune | 1027 | 0 / 1 |

## Superseded attempts and limits

An initial oracle call passed the mode instead of the full symbol and failed with no consumers: 6.174s, evidence/superseded-mode-lookup.log. Corrected in owned test code. The first mixed-corpus run required an optional mode field on AST rows and failed on source Node with missing option value: 43.145s, evidence/superseded-driver.log. The two regexp comparisons passed in that run, but the package failure is not credited as a green gate. The final complete rerun above passed after optional-field handling and removal of withdrawn oracle paths.

The discarded CallExpressionSource and HasAttributeNamed ports had local passing four-way comparisons and mutants. Historical coverage and logs carry the withdrawn prefix and do not count toward this batch's three retained ports. ImportedNameOf's earlier 4,109-query corpus became 4,093 after removing unrelated factory nodes and adding a present empty string property-name control. No shared compiler, registration generator or test harness was edited.

Not covered: full linter integration, complete rule defaults/listeners, diagnostic spans, fixes or suggestions; the regexp engine and remaining rewrite helpers; invalid UTF-8 Go byte strings, arbitrary noninteger numeric offsets, corrupted AST arenas or exact Go panic prose. Negative offsets refuse, but their panic text is not compared in this gate. Common adapter decoding and valid named-field projections remain integration dependencies. No fourth helper is reserved.
