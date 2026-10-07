Built: addRepositoryFunctionalRoots, addThemeNamespaces and NewUtilityEvaluator, each in its own .a file.
Commits: claim e0b9515 pushed before code; implementation 05b1ced9e9b8583c88f5317a71e20153eab3ae9a; current main base e8ba3d5.
Checks: three-helper four-way Go oracle PASS 22.266s, 3,876 cases; vet/types/format pass; filtered uncached oracle PASS 0.830s.
Mutants: eleven compiling semantic mutants caught by actual Go output comparisons, four descriptor, five namespace and two constructor variants.
Not covered: full repository gate, full rule findings/fixes/suggestions, shared parser integration or standalone normalization/CSS engine implementations.

## Landing and ownership

Before any new claim, the only branch this unit has pushed, codex/lint-helpers-05, was already rebased onto current origin/main e8ba3d5d81de4d3773c723914fccd4c76248b965 and pushed at 474d088. Its complete six-package helper gate had passed again, including all forty-one previous semantic mutants. See ../LANDING_REPORT.md. A fresh fetch confirmed main had not advanced. No main or area branch was pushed.

Explicit wildcard fetch checked all eighteen origin codex/lint-helpers* branches and their claims files. The original comment bundle is also reserved in shared HELPERS.md. All three selected concrete helpers tie the highest unclaimed count at six consumers. The generic seven-consumer strict-option label is remaining rule-local schema/decoding work, not an additional public helper.

Claim e0b95155d759f07881ef6b6acd9ec584851e2ca4 was committed at 2026-10-07 03:12:38 UTC and pushed before writing code. A later refresh found slot 02 overlapping reservations for NewUtilityEvaluator and addRepositoryFunctionalRoots at e283cc2317a01b30dadd0831c5635a81c594e7d2, 03:13:26 UTC. Our claim precedes them by forty-eight seconds and retains ownership under the established earliest-claim rule. No competing namespace claim was found. evidence/ownership.json retains original immutable claim identifiers.

## Behavior and external oracle

The descriptor helper keeps the caller map, leaves static/false functional entries alone, overwrites a collision with a fresh root-specific declining descriptor and replaces that record again on a repeated call. Old records survive through existing aliases. Nil metadata is projected with presence bits; zero axis orders/counts and absent readings are checked against the Go descriptor.

The namespace helper consumes Theme.Entries keys and the separately owned pure Theme.KeysInNamespaces dependency. It considers all proper segment prefixes, handles empty segments, ignores non-custom-property keys, deduplicates candidates/member keys, drops empty namespaces and replaces the table containers. Namespace precedence uses descending UTF-8 byte length and ascending byte order. The real Go theme builder supplies ambiguous nested keys, font subnamespace exclusions, subvariable controls, Unicode keys and prefixed entry cases. Previous table map/list aliases remain unchanged after the new table is mutated.

The constructor normalizes every input definition in order before registration. All input definitions, including overwritten ones, remain observable; the final map holds the last duplicate's original identity. Theme and definition aliases survive, and each constructor map is independent. Normalization belongs to the explicit normalizeUtilityDefinition callback. The private adapter's body is an opaque handle for that dependency, not a CSS parser. Its oracle predictions are actual Go node values after the real constructor normalized them. Empty-input and absent-theme cases are checked as well.

Stage 0 refused EvaluatorTheme | null at new_utility_evaluator.a:6:37 in the first run, which failed after 2.558s. No mutant from that run is credited. The private representation uses explicit presence bits instead, preserving nil states without changing any compiler or shared harness file. Definitions are nonnil; Go itself faults on nil definitions. Callback mutation outside normalization, concurrent mutation and exact runtime crash prose remain outside the adapter contract. See README.md for the projection boundaries.

Every nonempty Go string literal from every inventory-listed test file of all six consumers contributes to the derived cases. The corpus includes descriptions and option strings, not just complete source fixtures. All six consumers have nonzero coverage: 259, 370, 110, 131, 247 and 239 literals respectively; deduplication yields 639 distinct strings. Seven additional empty/dash/Unicode/NUL controls and two prefix variants produce 1,292 cases per helper. Pin drift and missing consumer coverage fail the test. This is observed helper behavior on those derived inputs, not whole-rule parity or an exhaustive proof over arbitrary repositories.

Each baseline compares actual Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db, source Node, emitted JavaScript and native under ASan/UBSan. Native and emitted compilation use existing compiler APIs in the private test file. No shared registration, parser, rule harness, compiler implementation or cohere worktree was edited.

## Mutants

All eleven variants compiled and exited zero without stderr before their wrong output was compared with Go. Compiler failures, crashes and sanitizer errors do not count. Exact anchors and first witnesses are retained in evidence/mutants.json and evidence/helpers-final.log.

| Helper | Mutation | First independent witness |
|---|---|---|
| addRepositoryFunctionalRoots | accept false functional entries | verdict 24: mutant true, Go false |
| addRepositoryFunctionalRoots | clear PerDeclaration | verdict 6: mutant false, Go true |
| addRepositoryFunctionalRoots | preserve an existing collision | verdict 2: old-record identity true, Go false |
| addRepositoryFunctionalRoots | reuse a previously declining record | verdict 26: repeated-call identity true, Go false |
| addThemeNamespaces | shortest namespace first | verdict 2: --a instead of --font-weight |
| addThemeNamespaces | retain empty membership | verdict 1: 12 namespaces instead of 8 |
| addThemeNamespaces | reverse lexical precedence | verdict 12: --😀 instead of --aaaa |
| addThemeNamespaces | use UTF-16 lengths | verdict 18: --a instead of --é |
| addThemeNamespaces | reuse old table map | verdict 30: old map size 10 instead of 1 |
| NewUtilityEvaluator | omit normalization | verdict 1: --value(--spacing) instead of --value(--spacing-*) |
| NewUtilityEvaluator | retain first duplicate | verdict 7: winning-definition identity false instead of true |

## Commands and observations

All test output went directly to logs. Source /workspace/adamic-tools/env.sh before commands. nproc printed 5. Setup timing lines:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (80s)
setup: done in 80s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

```sh
bash cloud/setup.sh > /tmp/lint05-batch7-setup.log 2>&1
ADAMIC_SLOT05_BATCH7_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch7/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch7 -count=1 -v -timeout=20m > /tmp/lint05-batch7-final2.log 2>&1
go vet ./stage1/cohere/lint/helpers/slot05/batch7 > /tmp/lint05-batch7-vet.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch7/main.a > /tmp/lint05-batch7-types.log 2>&1
gofmt -l stage1/cohere/lint/helpers/slot05/batch7 > /tmp/lint05-batch7-format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/lint05-batch7-oracle.log 2>&1
```

Final package PASS 22.266s: descriptor 5.20s, namespaces 13.31s, evaluator 3.75s. Seven-mutant and ten-mutant intermediate gates passed at 16.426s and 21.398s respectively; their logs are retained without crediting them as additional mutants. The final gate supersedes both. Vet and formatting logs are empty; types prints ordinary inferred-type records. Filtered uncached oracle PASS 0.830s, six fixtures, zero cache hits and six probe misses. Committed log copies trim trailing whitespace; raw logs remain under /tmp.

## Rules unblocked and limits

Each helper removes one listed prerequisite from each of:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

This removes eighteen dependency occurrences across six rules and no final blocker. Cumulative slot 05 retains twenty helpers, removes 169 prerequisite occurrences across sixty-four unique rules, and removes four final blockers, raising conditional helper readiness from 46 to 50. CONSUMERS.md maps each helper explicitly; readiness.json records every residual blocker. No inventory rule status is rewritten.

The full repository gate and full-rule diagnostic/span/fix/suggestion parity were not run. Other workers' helper integration, dynamic/external fixtures, arbitrary invalid UTF-8, standalone CSS parsing/normalization and production parser/linter integration remain uncovered. The claimed helper comparisons have no remaining shared-harness blocker. No fourth helper is reserved.
