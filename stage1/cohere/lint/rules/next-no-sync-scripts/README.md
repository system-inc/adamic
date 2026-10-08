# no-sync-scripts and JSX helper landing

The rule uses the shared JSX helpers through a structural projection of the stage 1 parser's node indices. It preserves exact attribute matching, both opening forms, namespaced/member tag rejection, spread rejection and the presence-based async/defer exemptions, including `async={false}`. Its oracle returns unchanged Go cohere `next.NoSyncScripts`.

The frozen helper ledger marks eight rules ready after the six base JSX helpers: `@next/next/no-img-element`, `@next/next/no-sync-scripts`, `react/forbid-dom-props`, `react/jsx-no-duplicate-props`, `react/jsx-no-script-url`, `react/jsx-no-target-blank`, `react/no-danger`, and `react/no-unknown-property`. This is helper readiness; this JSX unit ports only no-sync-scripts.

| Go JSX symbol | File under `stage1/cohere/lint/helpers/jsx/` | Selected source branch |
| --- | --- | --- |
| AttributeName | `attribute_name.a` | `codex/lint-helpers-02` |
| HasAttributeNamed | `has_attribute_named.a` | `codex/lint-helpers-02`, slot02/batch7 |
| MatchIgnoringCase | `match_ignoring_case.a` | `codex/lint-helpers-02`, slot02/batch7 |
| IsIntrinsicElementNamed | `intrinsic_element_named.a` | `codex/lint-helpers-03`, slot03/batch2 |
| ElementParts | `element_parts.a` | `codex/lint-helpers-04` |
| MatchExactly | `match_exactly.a` | `codex/lint-helpers-05`, slot05/batch3 |
| StringAttributeValue | `string_attribute_value.a` | `codex/lint-helpers-01` |

The triage's retained owners are preserved: AttributeName comes from slot02 rather than withdrawn slot05; MatchExactly comes from slot05 rather than withdrawn slot03. StringAttributeValue keeps slot01 over withdrawn slot02 and now calls the landed `ecmascript/text.UnescapeStringLiteralText` directly, matching `cohere/internal/lint/ecmascript/jsx/attributes.go:119`. It stops at the first matching nonliteral initializer rather than reading a later duplicate.

`helpers/jsx_test.go` freshly captures 1,284 inputs from all 27 JSX consumers at cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`. It compares 23,942 AttributeName/ElementParts node observations, 31,880 exact/folded matcher pairs, and 459,941 intrinsic/attribute/string-value queries: 515,763 output rows, byte-identical on source Node, emitted JavaScript and sanitized native. It runs one output-changing mutant per helper on all three backends. Complete Unicode simple-fold table regeneration checks all scalar values and 1,512 noncanonical mappings.

The registered rule captures 11 upstream cases and adds `testdata/blocking.tsx.txt`; `mutant.json` inverts src-presence handling. The existing lint discovery harness owns its Go/source Node/emitted JavaScript/native comparisons and semantic mutant checking.

The complete helper package passes. The full lint run exercised upstream fixtures, compiler corpus, owned witnesses, all registered mutants, sharding, throughput and profiles. The frozen count-guard edit was undone, and `e7c196a9` supplies JSX inventory discovery. The separate no-progress driver fix preserves Go edit-engine refusals. The existing `TestCheckerBridgeRefusalPending` still skips because the compiler prelude lacks `TSGoError`; this compiler dependency is outside the JSX landing.
