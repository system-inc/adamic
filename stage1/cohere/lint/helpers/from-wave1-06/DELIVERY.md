Built two uniquely owned .a helpers: Table.addRepositoryStatics and attachFrameworkVariantComparisons, removing twelve prerequisites across six rules, zero final blockers alone.
Commits: 2001f9b4 and 8702dbc6 claims pushed before implementation; repository-static delivery 1b93ad70; final second-helper delivery follows this report.
Commands: combined helper package PASS 35.664s; 2,147 repository cases and 1,253 comparison-name inputs; Go, source Node, emitted JavaScript and sanitized native match; vet clean.
Mutants: four per delivered helper, eight total; all compile and run cleanly and fail only Go byte comparison on all three Adamic execution paths.
Not covered: full consuming-rule integration, independently compiled dependency helpers, full repository gate, external Tailwind/corpora and resizing of documented read-only order storage.

## Delivered helpers

`repository_statics.a` owns Table.addRepositoryStatics. It iterates every repository static body, uses the separately owned PropertySort callback, overwrites framework collisions, keeps unrelated entries, shares sorted order storage and copies the reading record. The exact private Go method supplies answers. 2,147 cases compare every pinned framework declaration body, nil/empty/value-absent bodies and all six consumer fixture name sets, with overwrite/duplicate/Unicode/NUL controls: 169,528 identical bytes. See REPOSITORY_REPORT.md for the representation and observational Go hook.

`attach_variant_comparisons.a` owns attachFrameworkVariantComparisons. It registers every generated comparison group in order, captures each group's direction by value, captures the original theme reference, and delegates through explicit AttachComparison and compareBreakpointVariants dependencies. The generated group data comes from actual Go. Original Go comparison callbacks determine every answer; their dependency results are supplied to the isolated Adamic helper rather than guessing any breakpoint logic.

Comparison controls include all four pinned groups, empty groups, mixed directions, duplicate and negative/zero orders, nil themes, live theme mutation, changing group objects after registration, and caller theme-variable reassignment. Every one of 157 actual captured fixture sources across all six consumers contributes source-derived names. Those names plus controls make 1,253 name inputs; each becomes two real Go ParsedVariant pairs and is exercised over all groups, two initial themes and three phases. The complete ordered registration trace and callback results yield 1,212,743 identical bytes on actual Go, original .a Node, emitted JavaScript and ASan/UBSan native. Go's AttachComparison is instrumented only to record attachment order; its behavior and callback logic are unchanged. Theme and variant identities plus real Go dependency answers are the adapter boundary, not a port of those other helpers.

All stdout, stderr and exit-status checks use files. Successful native executions have no stderr and sanitizer/leak checks enabled. There are no production mutations and no garbage collector.

## Mutants

| Helper | Mutation | Catch |
|---|---|---|
| repository statics | skip an existing root | framework collision retains -99/-1 instead of the real reading |
| repository statics | increment sorted count | complete count snapshot differs |
| repository statics | clone sorted order array | returned-element mutation no longer reaches original sort storage |
| repository statics | reuse sorted record | returned-count mutation incorrectly changes original sort record |
| attach comparisons | invert direction | actual Go callback result differs |
| attach comparisons | capture live group object | post-registration group mutation changes direction |
| attach comparisons | increment group order | ordered registration trace and map keys differ |
| attach comparisons | clone captured theme | live theme update fails to reach registered callbacks |

Every row was caught independently on source Node, emitted JavaScript and sanitized native after clean execution. The final combined run repeats all eight, twenty-four successful mutant/backend executions before the expected byte differences. An earlier repository observer indexed an empty original array when testing the wrong-overwrite variant; that runtime failure was not credited and was corrected before the final clean-executing mutant. Historical NewTheme and FrameworkStaticReading duplicates passed separate gates but were withdrawn to earlier claims and are not counted as delivered helpers or readiness improvements.

## Consumers and readiness

Both delivered helpers remove one dependency from each of:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Twelve edges across six distinct consumers; zero additional rules become fully helper-ready. readiness.json subtracts only these two delivered symbols. Callback integration, remaining helpers and full rule findings/fixes/suggestion parity are separate work. The frozen 152-rule ledger was used for ranking; the delivered comments bundle and earlier option helpers were excluded as already owned. Every origin branch's helper claims was inspected, with immediate post-push refresh for both final claims. Earliest-commit ownership was respected for two delayed concurrent collisions.

## Validation commands

```bash
source /workspace/adamic-tools/env.sh
go test ./stage1/cohere/lint/helpers/from-wave1-06 -count=1 -v -timeout=20m > /tmp/helpers06-final.log 2>&1
go vet ./stage1/cohere/lint/helpers/from-wave1-06 > /tmp/helpers06-final-vet.log 2>&1
python3 stage1/cohere/lint/helpers/from-wave1-06/testdata/regenerate.py > /tmp/helpers06-capture-final.log 2>&1
go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > /tmp/helpers06-oracle.log 2>&1
```

Observed combined package PASS 35.664s, repository static gate PASS 12.65s and comparison gate PASS 23.01s. Vet exits zero. The filtered external input oracle passes all six probes in 10.132s with zero hits and six misses. The 157-source capture reproduced byte for byte across helper selection changes. Full repository tests were not run. Consumer-package capture still fails eight external installation/live/corpus checks because the hardcoded /Users/kirkouimet/Projects/ahra/app/_theme/styles Tailwind installation and repository corpora are absent. They remain recorded failures, not green rule gates.

## Environment and ownership

Worktree /workspace/adamic-helpers06; branch codex/lint-helpers-from-codex/lint-wave1-06 based on origin/codex/lint-helpers 95100eb4. bash cloud/setup.sh passed: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 20s, cache warm/done 490s, nproc 5. Go 1.27.1, clang 20.1.8, Node 24.19.0; cohere pin 715ba94f3608a6500086b1076ce5cb7e51b836db. The older runtime APIs and nullable-array compiler refusal were accommodated only in owned adapters. No shared compiler, shared registration generator, shared harness, other worker directory or cohere worktree source was edited.
