# Breakpoint helper continuation

Delivered two additional helpers in separate `.a` files: `collapse.resolveBreakpointWidth` and `collapse.compareBreakpointVariants`. Prior eighteen rule ports and both first-batch helper ports were tested and pushed before new ownership. Continue on the helper branch created from origin/codex/lint-helpers (`95100eb4`), codex/lint-helpers-from-lint-wave1-08. Read the base README and frozen readiness ledger, fetched all origin heads, and checked every helpers/claims document. The first scan covered 417 refs; the second covered 418. Both symbols were unclaimed and tied the largest available concrete-symbol count, six consumers. Higher-count dependencies were reserved; the comment bundle is delivered on the base branch. Claims `081992a9` and `ca3548cc` were pushed before implementation. Width was independently tested and pushed at `388ab583` before claiming comparison.

## Behavior

Width resolution preserves nil-theme precedence even for arbitrary widths, static breakpoint lookup, absent functional value, exact case-sensitive `var(` rejection anywhere in an arbitrary value, container selection for roots beginning `@`, and unresolved arbitrary/compound kinds. The theme lookup operation remains an explicit caller dependency. This helper includes the trivial root-to-namespace selection in its dispatch; it does not duplicate the separately owned theme resolver.

Comparison resolves left before right, treats both unresolved as a tie, places one unresolved value at the narrow end in either sort direction, and delegates two resolved widths to an explicit comparator. Both resolution and the actual breakpoint numeric/string comparator belong to separately owned helpers. The fixture driver uses the delivered width resolver and only compares equal arbitrary widths or the fixed distinct pair 40rem/24rem. Its comparison adapter refuses any unexpected distinct pair rather than guessing a comparator result.

## Consumers and readiness

Each helper removes one dependency from every rule below, twelve entries for this batch and twenty-four cumulatively with the prior constructor/clearing batch:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

**Zero additional rules become fully helper-ready from this batch alone.** Other CSS, candidate, theme, design-system and variant dependencies remain. These are bounded shared-helper comparisons, not complete rule finding/fix certifications. The base readiness.json and shared registration/harness files are unchanged.

## Evidence

The Go oracle is pinned cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`. Only oracle-only probe files are added through an overlay; upstream helper implementations are unchanged. All six consumer test files contribute every Go string literal, with per-consumer counts 264, 376, 112, 133, 255 and 244. Eleven controls give 1,395 samples. Width exercises four variant kinds, two value kinds, absent/present values, nil/present theme and media/container roots: 89,280 cases. Comparison exercises eight shape pairs, nil/present theme and both directions: 44,640 cases. Width text is serialized as UTF-16 numeric units so Unicode and control values remain directly comparable.

| New helper | Cases | Go-identical bytes on Node, emitted JavaScript and sanitized native |
| --- | ---: | ---: |
| resolveBreakpointWidth | 89,280 | 1,184,104 |
| compareBreakpointVariants | 44,640 | 103,198 |

Every baseline and mutant must compile and exit zero with empty stderr. Native uses ASan/UBSan and leak checking. A compile error, panic or sanitizer error cannot count as a killed semantic mutant.

- Width mutant disables the arbitrary-value `var(` refusal. Actual Go differs on the explicit variable controls; source Node, emitted JavaScript and native comparisons all caught it.
- Comparison mutant reverses the left-unresolved ordering. All three comparisons caught it on the one-unresolved direction cases.
- The constructor and namespace-clearing baselines and their prior mutants were rerun as part of this batch, all passing. Their corpus now includes the four added controls, so historical first-batch counts remain in the earlier report.

Final commands, after sourcing /workspace/adamic-tools/env.sh:

```
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/from_wave08 -count=1 -v -timeout=10m > /tmp/wave08-breakpoint-tests.log 2>&1
go vet ./... > /tmp/wave08-breakpoint-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/wave08-breakpoint-oracle.log 2>&1
```

Owned package PASS in 21.783s, four helper baselines and four mutants across all three port backends. Repository vet exit zero with empty log. Filtered external oracle PASS in 0.675s, one native and one Node miss, zero hits. git diff --check passed. The initial comparison fixture driver passed a number to the string-only console API and was rejected by type checking; it now formats the number explicitly. That failed attempt is not counted as a semantic mutant or passing test.

Setup passed: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 1s, build cache warm 29s, total 29s; nproc=5. Go 1.27.1, clang 20.1.8, Node 24.19.0. Logs are under evidence/breakpoints*.log.

## Limits

Every consumer supplies fixture strings and receives isolated helper shape probes; complete rule execution and every arbitrary theme configuration were not tested. Named lookups in the width fixture theme are present; missing-key semantics are delegated to the externally owned theme resolver. The comparator fixture adapter covers equal arbitrary widths and a fixed numeric pair, not the full Go CompareBreakpoints contract or integer-overflow behavior. There is no certification of arbitrary stateful callbacks, concurrency, complete CSS/design-system integration, the full inherited helper suite or the full repository test gate. No shared compiler or test harness file was edited.
