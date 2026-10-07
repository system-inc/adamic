Built: attributeValues, decodeEntity and IsLikelyComponentName in three .a helper files plus two generated data files.\
Commits: e31de1d, 1d106df and 0ca32ea pushed; report/evidence 3d69400 pushed; all naming claims preceded code.\
Commands/results: complete helper package PASS, 228.362s; targeted PASS 107.976s; vet clean; filtered uncached oracle PASS 1.106s.\
Mutants: seven new compiled semantic mutants caught by Go comparisons, plus all 18 prior/inherited mutants rerun.\
Not covered: whole-rule findings/fixes, dynamic fixtures, cross-slot integration, invalid raw-byte strings and full repository gate; initial GitHub authentication failure resolved by the final retry.

# Slot 01 third helper batch report

Branch codex/lint-helpers-01, same base and frozen cohort as the previous two reports. All six previously delivered helpers were tested and pushed before this batch. No rule implementation was claimed by this helper unit. The three highest available helpers were selected one at a time after wildcard fetches and inspection of every origin codex/lint-helpers* claims file; no further helpers were claimed.

## Delivered behavior and readiness

| Helper | Consumers | Final helper blockers removed alone |
|---|---:|---:|
| *ClassLiteralReader.attributeValues | 11 | 0 |
| text.decodeEntity | 9 | 0 |
| react.IsLikelyComponentName | 8 | 0 |

This batch removes 28 dependency entries for 28 distinct rules. It makes zero additional rules fully helper-ready. Across all nine slot 01 helpers, the same three structure rules from the first batch remain the only final listed helper blockers removed; no additional whole rule is marked implemented. slot01_wave3_readiness.json lists all consumers and every residual dependency, conservatively removing only this slot's helpers, with the inventory's common AST adapter assumption.

Attribute values preserves exact raw-name text matching by true map value, nil-name emptiness, nil-initializer delegation and Attribute origin. Go's private AST assertions refuse invalid preconditions; the adapter explicitly panics on nil/wrong-kind attributes or UnsupportedText name access. The separately owned classValuesUnder callback supplies literal/template payloads, untouched by this surface reader.

Entity decoding preserves exact decimal/hex validation, lowercase-x recognition, numeric saturation above U+10FFFF, literal passthrough for oversized numeric and unknown alphanumeric references, and U+FFFD for surrogate numeric values. Nonempty bodies are required, as Go indexes item[0]. The separately owned hexValue callback receives decoded runes. All 253 named XHTML entries are included in text_xhtml_entities.a, generated from pinned cohere's typescript-estree 8.65.0 table; no browser HTML decoder or guessed table is substituted.

Component naming checks only the first decoded rune, with Go unicode.IsUpper's category Lu semantics. Binary search over 152 Go unicode.Upper R16/R32 triples preserves strides, non-ASCII and astral capitals, titlecase exclusions and empty-name false. The data is Unicode 17.0.0 under Go 1.27.1, not a JavaScript casing heuristic. The data generator is retained in testdata/slot01_upper_ranges.go. One helper is defined per .a helper file; data files define only constants. SLOT01_WAVE3_README.md describes all APIs and adapter contracts.

## Commands and observations

No shared registration generator, shared rule harness, compiler ownership file, existing rule entry point, frozen readiness ledger or cohere source was edited. All third-batch test files and overlays are slot-owned additions. New emitted-JavaScript comparison lives in slot01_wave3_test.go rather than changing another worker's shared harness. No PR was opened.

All test stdout/stderr went directly to logs. Semantic mutants must compile and exit 0 with empty stderr before an actual output difference is credited; compiler rejection, panic or ASan/UBSan findings do not count. Baselines compare actual Go, Node source via oracle/node.mjs, native with ASan/UBSan and emitted JavaScript on Node. The latter is generated with internal/javascript from the same lowered IR; Go and source Node remain independent behavior oracles.

- source /workspace/adamic-tools/env.sh in every build shell. nproc printed 5.
- go test ./stage1/cohere/lint/helpers -run '^TestSlot01Wave3' -count=1 -v -timeout=20m > evidence/slot01-wave3/targeted.log 2>&1: PASS, 107.976s, all seven new tests.
- go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > evidence/slot01-wave3/final.log 2>&1: PASS, 228.362s, all 22 top-level tests.
- go vet ./... > evidence/slot01-wave3/vet.log 2>&1: exit 0, empty log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > evidence/slot01-wave3/oracle.log 2>&1: PASS, 1.106s, six fixtures, zero probe hits and six misses.
- gofmt -l stage1/cohere/lint/helpers and git diff --check: empty output.

Evidence paths in commands above are relative to stage1/cohere/lint/helpers; commands were run from the repository root with that prefix. Initial per-helper successful captures are also retained as attribute-initial.log, entity-initial.log and component-initial.log. No setup rerun was needed in this continuation; the same unit's settled-tree setup completed in 75s on five processors, with full initial failure/retry evidence in SLOT01_REPORT.md and evidence/slot01/setup.log. Timing lines: go ready (0s), clang ready (1s), node ready (1s), submodules ready (1s), build cache warm (75s), done in 75s, cgroup cpu.max 400000 100000, 17.6 GB.

Attribute corpus: all 11 consumer files, 956 extracted complete string expressions plus controls, 962 source batches and 720 helper queries. The pinned Go parser visits every JSX attribute; three settings maps test defaults, false entries, custom case-sensitive names and empty settings. Actual classValuesUnder supplies complete initializer payloads; actual attributeValues supplies expected outputs. Node, native and emitted JavaScript match 1,432 output lines.

Entity corpus: all nine consumer files, 863 extracted strings plus controls, 8,044 nonempty body queries. Every extracted string is tried as a body, embedded ampersand/semicolon references are captured, every actual named table entry and case variants are tested, numeric values sample the full Unicode range, all surrogate numeric values are covered in decimal/hex, and malformed/long-overflow cases are included. Node, native and emitted JavaScript match 24,134 output lines, observing success, UTF-16 units and raw replacement text. Go's empty-body panic is also asserted during capture.

Component corpus: all eight consumer files, 568 extracted complete strings, parsed identifier names, targeted name controls and every valid Unicode scalar value with a suffix. 1,115,624 verdicts match actual Go, Node, native and emitted JavaScript. Unicode version is printed by the actual Go oracle: 17.0.0. Across the three primary corpora, 1,141,190 Go output lines matched each of the three Adamic execution paths.

These corpora statically evaluate Go string expressions, including complete concatenations and labels/messages. They do not run whole consuming-rule harnesses or dynamic fixture generation. Attribute callbacks and hexValue are dependency seams, not duplicate implementations of other slots' helpers.

## Every new compiling semantic mutant

| Helper | Mutation | Actual Go observation that caught it |
|---|---|---|
| attributeValues | remove configured-name membership guard | line 5: mutant one literal, Go zero |
| attributeValues | treat false map entries as accepted via Map.has | line 5: mutant one literal, Go zero |
| decodeEntity | reset oversized numeric accumulation to zero | line 20: mutant UTF-16 unit 0, Go literal &#1114112; units |
| decodeEntity | retain numeric surrogates instead of U+FFFD | line 5384: mutant unit 55832, Go 65533 |
| decodeEntity | omit named-table lookup | line 50: mutant &amp; units, Go ampersand unit 38 |
| IsLikelyComponentName | restrict matched ranges to ASCII capitals | line 580: mutant false, Go true |
| IsLikelyComponentName | accept holes between uppercase range strides | line 584: mutant true, Go false |

The first surrogate mutant survived a console-only observation: Node/native console output renders a lone surrogate and U+FFFD identically. That green comparison was not accepted as proof. Adding actual UTF-16 code-unit observations exposes the difference; the rerun catches it without a compiler refusal, panic or sanitizer error. The successful retained logs include the strengthened check.

The full package reruns all 18 previous/inherited mutants: path normalization, class expressions, default class entry and the three shared default arrays; factory pattern reversal, invalid-pattern acceptance and false attribute values; memo cache bypass, unbound caching and skipped empty entries; literal opening-quote strip and short-range guard; inherited raw-control acceptance, overlapping oneOf acceptance, missing interpolation and unknown-field acceptance. Each successful compiled mutant is caught by its Go comparison, with exact witnesses in final.log and earlier reports.

## Claim ownership and publication

- attributeValues claim ad07122 at 00:59:17 UTC preceded slot 03 5142004 and slot 04 0582ef8. Both later duplicates withdrew. Implementation e31de1d was tested and pushed before the next claim.
- decodeEntity claim 892d365 was pushed before code. All remaining 11- and 10-consumer helpers were reserved by then; it tied UnescapeStringLiteralText at nine. Implementation 1d106df was tested and pushed before the next claim.
- IsLikelyComponentName claim b3923a0 was pushed before code. UnescapeStringLiteralText was already reserved by slots 03/04, making this eight-consumer helper the sole highest remaining count. The later wording clarification a61db53 corrected byte/rune terminology; implementation 0ca32ea is completed and pushed. A final wildcard fetch confirmed no other remote claim for this helper or decodeEntity.

GitHub publication initially failed after b3923a0: repeated git push origin codex/lint-helpers-01 attempts printed exactly `fatal: could not read Username for 'https://github.com': No such device or address`. Public wildcard fetch still succeeded. The cloud runtime status at diagnosis reported current observations, enforced networking, no secrets and no outbound identities. No interactive login, guessed identity, secret printing, shared configuration change or force push was attempted. After validation and the report commit, the final authorized retry succeeded, updating origin from b3923a0 to 3d69400. git ls-remote independently confirmed 3d69400 on codex/lint-helpers-01. All three implementations, clarification, final report and evidence are pushed; no format-patch fallback was required or created.

## Every consumer and remaining scope

### tailwind.*ClassLiteralReader.attributeValues

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-important-position`
- `better-tailwindcss/enforce-consistent-variable-syntax`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-deprecated-classes`
- `better-tailwindcss/no-duplicate-classes`
- `better-tailwindcss/no-unknown-classes`
- `better-tailwindcss/no-unnecessary-whitespace`

### text.decodeEntity

- `@next/next/google-font-display`
- `@next/next/google-font-preconnect`
- `@next/next/next-script-for-ga`
- `@next/next/no-before-interactive-script-outside-document`
- `@next/next/no-css-tags`
- `@next/next/no-html-link-for-pages`
- `@next/next/no-page-custom-font`
- `@next/next/no-unwanted-polyfillio`
- `structure/boundary-no-project-theme-value`

### react.IsLikelyComponentName

- `structure/consistency-require-matching-file-name`
- `structure/react-component-no-const-assignment`
- `structure/react-component-no-destructuring`
- `structure/react-component-no-display-name`
- `structure/react-component-no-separate-named-export`
- `structure/react-component-require-named-export`
- `structure/react-component-require-properties-type-suffix`
- `structure/react-hook-no-properties-in-dependencies`

No whole-rule findings/spans/fixes, shared registration, suggestion serialization, arbitrary cross-slot adapters or native rule integration is claimed. The attribute callback remains an integration dependency; the hexValue test provider implements the independently owned Go leaf contract rather than importing an unmerged worker file. Invalid raw UTF-8 byte strings and malformed manually assembled AST payloads beyond the documented precondition refusals are not covered. Unicode category data is pinned to the printed toolchain version and requires revalidation if Go changes it. The full repository go test ./... gate was not run; bounded verification runs the entire touched helper package, repository-wide vet and a filtered uncached external oracle.
