Built: landing-ready rebase of all 23 retained slot 04 helpers; no new claims.
Commits: old tip 6877b840930cb8a3924cd7d1acf54a94196ef126; tested rebased tip bb1288067f6f8061e0dff45d07f998b35ea4f039; base e8ba3d5.
Checks: all nine helper packages and six uncached Node oracle probes passed; vet and format clean.
Mutants: 39 slot-owned and 10 inherited semantic/guard mutants rerun and caught by their existing oracle or refusal checks; exact test names below.
Not covered: full repository gate, broader input domains excluded by prior reports, new helper claims or rule ports.

The user required every pushed branch to be rebased onto current main and green before claiming more work. This report records that landing unit. Only codex/lint-helpers-04 is updated; integration owns main and area branches. The rebase completed without conflicts or implementation edits. Historical reports retain their original commit identities and input-domain limitations.

## Verification

With /workspace/adamic-tools/env.sh sourced:

```
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/... -count=1 -p 2 -v -timeout=30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m
go vet ./stage1/cohere/lint/helpers/...
gofmt -l stage1/cohere/lint/helpers
git diff --check
```

All output was redirected to logs. The helper packages passed in 189.267, 260.194, 27.344, 25.045, 43.332, 34.740, 39.805, 31.323 and 36.007 seconds respectively. The filtered oracle passed in 47.643 seconds with zero probe cache hits and six misses. Vet, formatting and whitespace checks produced no errors. Evidence logs are in evidence/.

The first setup already passed: Go 0s, clang 1s, Node 1s, submodules 2s, warm 165s, total 165s; nproc reported 5. Setup was not repeated for the rebase.

## Mutant evidence

These are existing compiling semantic mutants and explicit guard mutants, all rerun after rebasing. Each semantic mutant is rejected by the corresponding Go/Node/native differential checks; the short-key and JSX guards exercise explicit refusal behavior. Coverage omission checks also passed. No new mutants were introduced for this documentation-only landing record.

```text
--- PASS: TestHelperMutants (66.39s)
    --- PASS: TestHelperMutants/options_json.ts (10.19s)
    --- PASS: TestHelperMutants/option_schema.ts (8.20s)
    --- PASS: TestHelperMutants/policy_message.ts (10.62s)
    --- PASS: TestHelperMutants/strict_options.ts (21.65s)
--- PASS: TestSlot04ElementPartsMutant (4.83s)
--- PASS: TestSlot04SettingsKeyMutant (5.79s)
--- PASS: TestSlot04CompiledReaderMutant (3.62s)
--- PASS: TestCommentMutants (116.54s)
    --- PASS: TestCommentMutants/can_begin_at.ts (21.10s)
    --- PASS: TestCommentMutants/collect_list_interiors.ts (22.79s)
    --- PASS: TestCommentMutants/sort_by_position.ts (21.29s)
    --- PASS: TestCommentMutants/all.ts (14.23s)
    --- PASS: TestCommentMutants/for_file.ts (33.91s)
--- PASS: TestJsxAdapterGuardMutant (18.94s)
--- PASS: TestWave2CompilingMutants (12.67s)
    --- PASS: TestWave2CompilingMutants/class_values_under.aleading:_false,_trailing:_false (3.54s)
    --- PASS: TestWave2CompilingMutants/collect_class_values.acollectClassValues(arena,_item.whenFalse,_origin,_edges,_values,_dependencies); (2.19s)
    --- PASS: TestWave2CompilingMutants/collect_class_values.acollectClassValues(arena,_item.right,_origin,_edges,_values,_dependencies); (3.09s)
--- PASS: TestSpaceCompilingMutants (9.57s)
    --- PASS: TestSpaceCompilingMutants/javascript_space.acase_0xfeff: (1.58s)
    --- PASS: TestSpaceCompilingMutants/blank.areturn_true; (1.19s)
    --- PASS: TestSpaceCompilingMutants/blank.aindex++ (2.09s)
    --- PASS: TestSpaceCompilingMutants/value_separator.acase_58:_ (1.26s)
    --- PASS: TestSpaceCompilingMutants/value_separator.acase_60:_case_10: (2.09s)
--- PASS: TestBytesCompilingMutants (9.07s)
    --- PASS: TestBytesCompilingMutants/top_of_stack.astack[stack.length_-_1] (2.39s)
    --- PASS: TestBytesCompilingMutants/peek_byte.areturn_input[index] (2.68s)
    --- PASS: TestBytesCompilingMutants/peek_byte.areturn_0; (2.70s)
--- PASS: TestSortedKeysCompilingMutants (10.32s)
    --- PASS: TestSortedKeysCompilingMutants/omit_false_values (4.66s)
    --- PASS: TestSortedKeysCompilingMutants/reverse_UTF8 (2.15s)
    --- PASS: TestSortedKeysCompilingMutants/UTF16_comparator (2.96s)
--- PASS: TestStringsCompilingMutants (21.12s)
    --- PASS: TestStringsCompilingMutants/comment.akind:_'comment' (2.28s)
    --- PASS: TestStringsCompilingMutants/comment.avaluePresent:_false (2.77s)
    --- PASS: TestStringsCompilingMutants/declaration.avaluePresent:_true (5.91s)
    --- PASS: TestStringsCompilingMutants/declaration.aparams:_[],_property, (1.63s)
    --- PASS: TestStringsCompilingMutants/breakpoint_bucket.abyte_===_46 (3.94s)
    --- PASS: TestStringsCompilingMutants/breakpoint_bucket.abefore_<_index (3.17s)
--- PASS: TestStringsCompilingMutants (20.63s)
    --- PASS: TestStringsCompilingMutants/theme_prefix_key.alet_index_=_2 (5.22s)
    --- PASS: TestStringsCompilingMutants/theme_prefix_key.aresult.push(45); (5.29s)
    --- PASS: TestStringsCompilingMutants/variant_registry_has.aregistrations.has(root) (4.07s)
    --- PASS: TestStringsCompilingMutants/design_system_prefix.areturn_system.theme.prefix; (4.31s)
--- PASS: TestPrefixKeyShortKeyRefusalAndMutant (6.61s)
--- PASS: TestStringsCompilingMutants (15.27s)
    --- PASS: TestStringsCompilingMutants/has_variant.asystem.variants.has(root) (3.24s)
    --- PASS: TestStringsCompilingMutants/variant_kind.a??_'static' (1.96s)
    --- PASS: TestStringsCompilingMutants/variant_kind.areturn_system.variants.get(root)_??_'static'; (3.03s)
    --- PASS: TestStringsCompilingMutants/variant_compounds_with.a===_'compound' (3.33s)
    --- PASS: TestStringsCompilingMutants/variant_compounds_with.areturn_system.variants.get(parent) (2.37s)
--- PASS: TestStringsCompilingMutants (15.99s)
    --- PASS: TestStringsCompilingMutants/walk.awalkNodes(arena,roots,visit) (2.57s)
    --- PASS: TestStringsCompilingMutants/walk_nodes.aif(action_===_2)_{_return_false;_} (2.64s)
    --- PASS: TestStringsCompilingMutants/walk_nodes.aif(action_===_0) (1.99s)
    --- PASS: TestStringsCompilingMutants/walk_nodes.aif(action_===_0)#01 (2.05s)
    --- PASS: TestStringsCompilingMutants/write_value_css.anode.tag_===_'word'_||_node.tag_===_'separator' (2.33s)
    --- PASS: TestStringsCompilingMutants/write_value_css.abuilder.push(41) (3.21s)
```

## Implementation identities after rebase

| Batch | Original implementation | Rebased implementation |
| --- | --- | --- |
| JSX parts, settings key, compiled reader | 27ffb6a | a3d43a9 |
| Class traversal (retained two) | 7904fda | c394765 |
| Space predicates and removal of yielded scanner | 9192bd4 | be92f24 |
| Stack, byte peek, sorted keys | 076e912 | fcf1d7d |
| Bucket and CSS constructors | b18ecc9 | a3582e5 |
| Prefix and registry accessors | 56b0cd3 | d686e2d |
| Variant queries | e5bd2f4 | a605705 |
| Walkers and CSS value writer | 27af0b0 | cef94f4 |

Consumer lists and conservative readiness accounting remain in the eight existing batch reports. JSXParts completes the final shared-helper dependency of @next/next/no-img-element; the other retained helpers remove prerequisites without claiming those rules are fully ported. The frozen original readiness ledger remains unchanged. Inherited helper suites are verified here without claiming authorship.
