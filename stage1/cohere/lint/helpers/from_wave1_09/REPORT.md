Built: nodesFromStaticDeclarations and private propertySort, one .a file each, removing twelve listed prerequisites across six Tailwind rules.
Commits: first claim 4d352341, first delivery f1c61f9e and logs 24416e8e; second claim 87bcfd7e pushed before second-helper code.
Checks: two-helper package PASS 49.224s; 6,924 cases match actual Go/source Node/emitted JavaScript/ASan+UBSan native; vet/types/format pass; filtered uncached input oracle PASS 7.056s.
Mutants: twelve compiling semantic mutations, five declaration-node and seven property-sort, all exit 0 without stderr and fail only comparison on all three execution paths.
Not covered: whole-rule findings/fixes/suggestions, production integration, raw invalid UTF-8, dynamic/external fixture execution or the full repository gate.

## Observed behavior

The oracle-only Go overlay exports the unchanged private function in static_utility.go. Go cohere is pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db. All six consumers' inventory-listed test files contribute nonempty Go string literals: 639 distinct strings, with per-consumer counts 259, 370, 110, 131, 247 and 239 in the order below. These include descriptions and options as well as source literals; they are derivative helper inputs, not observed full-rule executions or instrumented live helper calls.

The corpus adds every entry of Go's actual framework static declaration table (890 utilities), empty lists, absent and present empty values, independent important flags, duplicate declarations, Unicode, NUL and line breaks. 3,461 lists produce 2,019,129 output bytes. Every declaration field and the zero fields of other node kinds match Go, including nil context/children represented by explicit absence flags. Repeated calls return independent nodes. Mutating an output node leaves its sibling, another call and input unchanged; subsequent input mutation leaves existing outputs unchanged. Native runs use sanitizer/leak checks; baseline and mutants finish with empty stderr.

## Mutants

| Mutation | First differing output line | Mutant / actual Go |
|---|---:|---|
| Infer presence from nonempty value | 29 | false / true |
| Drop important | 30 | false / true |
| Return comment kind | 3 | comment / declaration |
| Reverse declaration order | 113736 | margin-inline-end / --tw-sort |
| Share first node with second | 44 | true / false |

All five mutations are independently applied to temporary copies. Each compiles to native and JavaScript and is caught on source Node, emitted JavaScript and sanitized native. The fresh-node witness begins with two identical declarations, so it specifically detects aliasing even when ordinary field output matches. Compiler refusals, sanitizer failures and crashes do not count as killed semantic mutants.

## Rules and readiness

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Six dependency occurrences are removed. No rule loses its final helper blocker from this helper alone. This is an inventory inference, conditional on common AST integration; no rule is marked ported. The local readiness ledger retains all other prerequisites.

## Commands and outputs

```sh
bash cloud/setup.sh > /tmp/lint-helpers-wave109-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE109_HELPER_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/from_wave1_09/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/from_wave1_09 -count=1 -v -timeout=15m > /tmp/lint-helpers-wave109-test-final.log 2>&1
go vet ./stage1/cohere/lint/helpers/from_wave1_09 > /tmp/lint-helpers-wave109-vet.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/helpers/from_wave1_09/main.a > /tmp/lint-helpers-wave109-types.log 2>&1
gofmt -l stage1/cohere/lint/helpers/from_wave1_09 > /tmp/lint-helpers-wave109-format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/lint-helpers-wave109-oracle.log 2>&1
nproc
```

Setup: Go 0s, clang 0s, Node 0s, submodules 1s, cache warm 77s, total 77s; nproc 5. Toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0. Vet and formatting logs are empty, types prints checked declarations. Filtered oracle covers six input fixtures, zero hits and six probe misses. Test output goes directly to log files. Only the touched helper package plus filtered oracle gate was run; unchanged foundation packages and full repository gate were not rerun.

## Ownership and limits

Base origin/codex/lint-helpers at 95100eb440b47f3f18e960c6f5b49cadf0dc1d9d. The claim audit examined 403 origin references and 16 distinct helper claim blobs, excluding already delivered comments and the generic per-rule strict-options label. This symbol tied the highest unclaimed concrete count, six. All new Adamic source is .a; all changes stay inside the owned helper directory and claim. Shared harness, registry and compiler files are untouched. Prior rule candidates/evidence remain pushed at a7ff8948 on codex/lint-wave1-09, with the documented dynamic-RegExp native and shared repair integration gaps.

This bounded derivative corpus is not an exhaustive proof over arbitrary strings or full rule findings. It does not reconstruct dynamic concatenations or external fixtures. Raw invalid UTF-8 is outside the valid-text boundary, and no Go allocation capacity is claimed. Production consumers must adapt their shared declaration-node representation explicitly; this leaf supplies no general CSS tree or registered rule.

## Second delivery: private propertySort

The second claim was pushed at 87bcfd7e after the first helper and its evidence were pushed. The fresh all-origin audit saw 417 references and 17 distinct claim blobs; this helper again tied the largest unclaimed concrete consumer count, six. The same six rules listed above each lose one further dependency occurrence. Zero lose their last blocker. The shared public PropertySort wrapper remains another worker's territory.

property_sort.a preserves the actual private Go function's sort result and declaration visit sequence. Container references are arena indices, so the usable CSS tree has no ownership cycles. The actual Go PropertyOrder map is an explicit input, exported by the independent oracle rather than guessed or copied from a generated answer table. Go pointer trees are independently restored from the arena for oracle execution. A repeat call observes clean state and untouched roots/child indices. No shared parser, AST adapter, property-order generator or other worker's helper is changed.

The corpus has 3,463 cases, including all 639 distinct consumer test-file strings and all 890 framework static declaration bodies. Additional cases contain nested rule/at-rule containers, context/at-root/comment wrappers, unknown and competing known --tw-sort values, absent versus present empty declarations, repeated roots and duplicate property positions. Unicode/NUL source text is retained. 486,278 output bytes agree on all four execution paths. Consumer literal contribution counts match the first helper. These are derivative helper inputs; full consuming rules are not executed.

| Property-sort mutation | First differing line | Mutant / actual Go |
|---|---:|---|
| Skip by empty text instead of presence | 9 | 2 / 1 |
| Stop counting after latch | 8 | 1 / 3 |
| Latch unknown --tw-sort | 38 | 0 / 1 |
| Descend into context | 100 | 5 / 4 |
| Change queue to depth-first insertion | 10 | 341 / 340 |
| Admit duplicate positions | 101 | 3 / 2 |
| Sort positions descending | 69 | 341 / 340 |

All seven independently compile and exit 0 with empty stderr on source Node, emitted JavaScript and sanitized native, before the comparator kills them. The visit trace is held as well as the final reading, so traversal disagreement cannot be hidden by deduplicating and sorting positions.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE109_HELPER_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/from_wave1_09/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/from_wave1_09 -count=1 -v -timeout=15m > /tmp/lint-helpers-wave109-two-final.log 2>&1
go vet ./stage1/cohere/lint/helpers/from_wave1_09 > /tmp/lint-helpers-wave109-two-vet.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/helpers/from_wave1_09/sort_main.a > /tmp/lint-helpers-wave109-sort-types.log 2>&1
gofmt -l stage1/cohere/lint/helpers/from_wave1_09 > /tmp/lint-helpers-wave109-two-format.log
```

The complete two-helper package passes in 49.224s, after a focused sort run passed in 34.040s. Vet and formatting output are empty; types prints checked declarations. Prior filtered uncached input oracle remains PASS 7.056s; unchanged compiler/input behavior was not retested. Evidence includes complete two-helper gate output, coverage and input/output hashes. The representation covers finite valid-text acyclic trees and actual integer property positions; raw invalid UTF-8, nil Go node pointers, cyclic graphs, native slice capacity and nil-versus-empty sort-list allocation are outside this consumer boundary.
