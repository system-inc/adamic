Built: nodesFromStaticDeclarations, one .a helper, removing a listed prerequisite from six Tailwind rules.
Commits: claim 4d352341 pushed before code; source, comparator and evidence in this report commit.
Checks: owned package PASS 15.276s, 3,461 cases match actual Go/source Node/emitted JavaScript/ASan+UBSan native; vet/types/format pass; filtered uncached input oracle PASS 7.056s.
Mutants: value presence, important, kind, order and fresh allocation all compile, exit 0 without stderr, and fail only output comparison on all three Adamic execution paths.
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
