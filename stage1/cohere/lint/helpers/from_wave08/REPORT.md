# Helpers from lint wave 1 slot 08

Delivered two helpers, one per `.a` file: `collapse.NewTheme` and `collapse.(*Theme).ClearNamespace`. The prior eighteen rule ports and evidence were already tested and pushed through `6a66e2d3` before this unit. This branch starts at `origin/codex/lint-helpers` (`95100eb4`); it does not import the rule branch.

`NewTheme` was reserved at `e069e11a`, pushed before implementation, then ported/tested/pushed at `63c3111c`. The second refresh found slot 03 already reserved NewVariantRegistry; an erroneously submitted duplicate claim `60f424a3` was explicitly withdrawn at `38281ddc`. No variant-registry code was written. The same correction reserved ClearNamespace, pushed before implementation. The first scan inspected 389 origin refs; the second inspected 403 and all helper claim paths. All higher-count concrete helpers were reserved, and the 24-consumer comment bundle is already implemented on the base branch and reserved in HELPERS.md. Each chosen symbol ties the highest remaining unclaimed concrete count, six consumers. No shared files were edited.

## Contracts

`newTheme()` creates independently writable value maps and order arrays, empty prefix and zero dead-key count. Go's nil order slice is represented by an empty array; nil slice identity is not exposed by this adapter. Generic mutation of the returned store has the same ordered-store shape used by adjacent workers.

`clearNamespace(theme, namespace, clearOptions, ignored, remove)` preserves Go's bare prefix matching, including `--colorful` matching `--color`. Every required option bit must be present. The ignored prefixes also use bare prefix matching; collection finishes in insertion order before deletion callbacks run. Ignored namespace data and the deletion operation are explicit separately owned dependencies. The fixture driver uses the exact Go ignored table for three exercised namespaces and a test-only deletion/compaction adapter. The production helper does not duplicate those dependencies.

## Consumers and readiness

Both helpers remove one prerequisite from each of these six rules (twelve dependency entries total):

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

**Zero additional rules become fully helper-ready from these two helpers alone.** Their remaining design-system, candidate, CSS parser, theme and variant dependencies are unchanged. These are dependency removals, not complete rule ports or finding/fix parity certifications. The frozen base readiness.json is unchanged. consumers.json identifies every consumer test file.

## Observations

The test reads every Go string literal from all six consumers' test files: 264, 376, 112, 133, 255 and 244 respectively. Seven independent controls give 1,391 inputs. NewTheme has no arguments; fixture strings are mutation keys used to probe construction independence, fresh storage, writable maps and subsequent allocations.

ClearNamespace builds explicit ordered stores from the same inputs plus prefix/ignored/duplicate/empty controls, then exercises five namespace selections and eight option masks. There are 55,640 clearing cases. Store setup deliberately bypasses Add's clear-directive validation to isolate the helper. Values contain option bits 0 through 7; required masks include static/used bits and correctly reject entries without them. Arbitrary option combinations and every namespace table entry are not exhaustive.

Final four-way baseline results, actual Go helper versus source Node, emitted JavaScript and native under ASan/UBSan:

| Helper | Cases | Identical output bytes |
| --- | ---: | ---: |
| NewTheme | 1,391 | 59,813 |
| ClearNamespace | 55,640 | 5,277,288 |

Every successful backend must exit zero with empty stderr, including mutants. Native sanitizer/leak failures and compile failures never count as killed semantic mutants. The Go overlay only adds probes exposing private store fields; actual Go NewTheme and ClearNamespace implementations are unmodified. Clearing compares final membership, exact surviving order, tombstones and dead-key counts without sorting.

| Helper | Compiling mutant | What caught it |
| --- | --- | --- |
| NewTheme | deadKeys initially one | Go output comparison on Node, emitted JavaScript and sanitized native |
| ClearNamespace | require namespace plus hyphen instead of bare prefix | Same three comparisons; --colorful boundary witness |

## Commands and outputs

Source `/workspace/adamic-tools/env.sh` before running:

```
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/from_wave08 -count=1 -v -timeout=10m > /tmp/wave08-helper-tests.log 2>&1
go vet ./... > /tmp/wave08-helper-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/wave08-helper-oracle.log 2>&1
```

Owned package PASS in 12.686s, including both baselines and both three-backend mutant runs. Repository-wide vet exit zero, empty log. Filtered external oracle PASS in 7.758s, one native and one Node miss, zero hits. `git diff --check` passed.

`bash cloud/setup.sh` passed: Go ready 1s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 139s, total 139s. nproc=5. Go 1.27.1, clang 20.1.8, Node 24.19.0. This helper base has a working setup gate; the prior rule branch's shared profiler failure is not carried forward.

Earlier uncredited failed attempts: the constructor driver initially assumed string-valued readTextFile; it now handles the current Ok/Error result. The clearing fixture setup initially used Add and rejected directive-like keys; direct test stores isolate clearing. An empty ternary-array branch was refused by stage 0; the driver now uses an explicitly typed array and assignments. No compiler changes were made.

## Limits

No full repository test gate, full inherited helper suite or complete rule finding/fix corpus was run. All six consumers supply fixture strings, not entire rule executions. Thread safety and concurrent constructor use are not certified; these stores are mutable, as in Go. Arbitrary hostile deletion callbacks, nil Go theme receiver, unknown ignored-table data and complete CSS/theme integration are outside this bounded helper contract. The deleting callback must implement the externally owned theme operation. The report does not claim every unrelated shared helper is complete.
