Landed-ready: all eighteen retained slot 01 helpers rebased without conflicts onto origin/main e8ba3d5.\
Commits: tested rebased implementation/report tip 8ae8f23a973092115d59b4adca11fe62db7a2d0c; publication evidence is this commit.\
Commands and outputs: full helper package PASS 285.624s (46 tests), vet exit 0, uncached input oracle PASS 29.982s; setup 120s, nproc 5.\
Mutants: all 47 compiling semantic mutants and two explicit domain-refusal mutants caught again; every witness is recorded below.\
Not covered: full repository test gate, whole-rule parity and cross-slot callback integration; no new helpers claimed during this landing unit.

# Slot 01 landing evidence

The only branch pushed by this unit is codex/lint-helpers-01. Before this operation it was not on main and did not contain current main. The original local tip was 5992275e855dbe3a932dcde1f0350fcb33b1c44c, while the fetched remote tip was 5fd641ddb4fd5818c40c7d67eea8e77bdd926dd4. All local work, including the sixth batch report, was preserved by rebasing onto e8ba3d5d81de4d3773c723914fccd4c76248b965. All 57 commits replayed without conflicts. No implementation changes were required. Main was fetched again after validation and remained unchanged.

The user's explicit landing instruction authorizes this rebase and the corresponding history update of this unit's own branch. The first push lease against stale tracking tip 5fd641d was rejected. The origin fetch refspec fetches only main; direct git ls-remote confirmed the actual remote tip was 5992275e855dbe3a932dcde1f0350fcb33b1c44c, the original local tip already preserved by this rebase. The retry uses an exact force-with-lease against that verified tip, protecting concurrent branch changes. Main and area branches are never pushed. This unit finishes the required landing work before any new claims.

## Commands and observed outputs

All commands run from repository root with /workspace/adamic-tools/env.sh sourced. Output goes directly to evidence/slot01-landing logs under this helper directory.

- `git rebase origin/main`: successful, no conflicts; rebase.log.
- `bash cloud/setup.sh`: successful; go, clang, Node and submodules ready at 0s, build cache warm at 120s, done in 120s; five processors; setup.log.
- `go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m`: PASS 285.624s, 46 top-level tests; helpers.log.
- `go vet ./...`: exit 0, empty vet.log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m`: PASS 29.982s, six probe misses; oracle.log.
- `gofmt -l stage1/cohere/lint/helpers/slot01*test.go`: empty format.log.
- `git diff --check`: empty diff-check.log.

Existing batch reports describe all eighteen helpers, each consumer, remaining dependencies, callback seams and individual mutant implementations. This revalidation uses the entire touched package, including inherited shared-helper checks. Go decides helper behavior; source Node, sanitized native and emitted JavaScript comparisons are exercised as recorded by each test. Semantic mutants must compile and run successfully before stdout differences are credited. The two domain mutants instead prove explicit byte and exact-integer adapter refusals. The new main's compiler and runtime are used by this rerun.

## Every mutant rerun and its catch

1. `TestHelperMutants/options_json.ts`

   helpers_test.go:145: compiled semantic mutant caught at output line 15157: got "valid", Go "invalid"

2. `TestHelperMutants/option_schema.ts`

   helpers_test.go:145: compiled semantic mutant caught at output line 15166: got "valid", Go "invalid"

3. `TestHelperMutants/policy_message.ts`

   helpers_test.go:145: compiled semantic mutant caught at output line 22166: got "This throws a bare `{{constructor}}`, which names no declared failure. Raise it through the tier that declares it, `AccountModule.error(identifier, data, cause)`, `ApiWorker.error(...)` or `Base.error(...)`. A bare throw carries no identifier, so the board groups it by its message and one interpolated value mints one identity per value, and it normalizes to 500, so a refusal reads as our fault.", Go "This throws a bare `sentinel é😀`, which names no declared failure. Raise it through the tier that declares it, `AccountModule.error(identifier, data, cause)`, `ApiWorker.error(...)` or `Base.error(...)`. A bare throw carries no identifier, so the board groups it by its message and one interpolated value mints one identity per value, and it normalizes to 500, so a refusal reads as our fault."

4. `TestHelperMutants/strict_options.ts`

   helpers_test.go:145: compiled semantic mutant caught at output line 15178: got "valid", Go "invalid"

5. `TestSlot01FileContextMatchesCohere`

   slot01_test.go:183: compiled structure_file_context.a mutant caught at output line 7: got "true false false false false false false false false", Go "true true false false false false true false false"

6. `TestSlot01Es6ComponentClassMatchesCohere`

   slot01_test.go:183: compiled react_es6_component_class.a mutant caught at output line 13400: got "false", Go "true"

7. `TestSlot01TailwindDefaultsMatchCohere`

   slot01_test.go:183: compiled tailwind_default_class_literal_settings.a mutant caught at output line 1: got "className;mergeClassNames|createVariantClassNames;.*[Cc]lassName$|.*[Cc]lassNames$", Go "class|className;mergeClassNames|createVariantClassNames;.*[Cc]lassName$|.*[Cc]lassNames$"

8. `TestSlot01TailwindDefaultsMatchCohere`

   slot01_test.go:206: compiled shared attributeNames mutant caught at output line 3: got "mutated attribute|className;mergeClassNames|createVariantClassNames;.*[Cc]lassName$|.*[Cc]lassNames$", Go "class|className;mergeClassNames|createVariantClassNames;.*[Cc]lassName$|.*[Cc]lassNames$"

9. `TestSlot01TailwindDefaultsMatchCohere`

   slot01_test.go:206: compiled shared calleeNames mutant caught at output line 3: got "class|className;mutated callee|createVariantClassNames;.*[Cc]lassName$|.*[Cc]lassNames$", Go "class|className;mergeClassNames|createVariantClassNames;.*[Cc]lassName$|.*[Cc]lassNames$"

10. `TestSlot01TailwindDefaultsMatchCohere`

   slot01_test.go:206: compiled shared variablePatterns mutant caught at output line 3: got "class|className;mergeClassNames|createVariantClassNames;mutated pattern|.*[Cc]lassNames$", Go "class|className;mergeClassNames|createVariantClassNames;.*[Cc]lassName$|.*[Cc]lassNames$"

11. `TestSlot01Wave2FactoryMatchesCohere`

   slot01_wave2_test.go:13: compiled tailwind_new_class_literal_reader.a mutant caught at output line 14: got "^custom$", Go ".*[Cc]lassName$"

12. `TestSlot01Wave2MemoMatchesCohere`

   slot01_wave2_test.go:16: compiled tailwind_class_values_in.a mutant caught at output line 14: got "", Go "!"

13. `TestSlot01Wave2LiteralMatchesCohere`

   slot01_wave2_test.go:19: compiled tailwind_class_literal_from.a mutant caught at output line 2: got "225 235 false false", Go "226 235 false false"

14. `TestSlot01Wave2LiteralShortRangeMutant`

   slot01_wave2_test.go:22: compiled tailwind_class_literal_from.a mutant caught at output line 522: got "1 0 false false", Go "0 1 false false"

15. `TestSlot01Wave2FactoryInvalidPatternMutant`

   slot01_wave2_test.go:25: compiled tailwind_new_class_literal_reader.a mutant caught at output line 1: got "3 3 6 true", Go "3 3 5 true"

16. `TestSlot01Wave2FactoryAttributeMutant`

   slot01_wave2_test.go:28: compiled tailwind_new_class_literal_reader.a mutant caught at output line 2: got "false true", Go "true true"

17. `TestSlot01Wave2MemoUnboundMutant`

   slot01_wave2_test.go:31: compiled tailwind_class_values_in.a mutant caught at output line 1: got "1 0 1", Go "1 0 0"

18. `TestSlot01Wave2MemoEmptyMutant`

   slot01_wave2_test.go:34: compiled tailwind_class_values_in.a mutant caught at output line 19: got "0 0 2", Go "0 0 3"

19. `TestSlot01Wave3AttributeMatchesCohere`

   slot01_wave3_test.go:19: compiled tailwind_attribute_values.a mutant caught at output line 5: got "1 0", Go "0 0"

20. `TestSlot01Wave3EntityMatchesCohere`

   slot01_wave3_test.go:161: compiled text_decode_entity.a mutant caught at output line 20: got "0", Go "38,35,49,49,49,52,49,49,50,59"

21. `TestSlot01Wave3EntitySurrogateMutant`

   slot01_wave3_test.go:164: compiled text_decode_entity.a mutant caught at output line 5384: got "55832", Go "65533"

22. `TestSlot01Wave3ComponentMatchesCohere`

   slot01_wave3_test.go:169: compiled react_likely_component_name.a mutant caught at output line 580: got "false", Go "true"

23. `TestSlot01Wave3AttributeFalseValueMutant`

   slot01_wave3_test.go:191: compiled tailwind_attribute_values.a mutant caught at output line 5: got "1 0", Go "0 0"

24. `TestSlot01Wave3EntityNamedTableMutant`

   slot01_wave3_test.go:194: compiled text_decode_entity.a mutant caught at output line 50: got "38,97,109,112,59", Go "38"

25. `TestSlot01Wave3ComponentStrideMutant`

   slot01_wave3_test.go:199: compiled react_likely_component_name.a mutant caught at output line 584: got "true", Go "false"

26. `TestSlot01Wave4StringValueMatchesCohere`

   slot01_wave4_test.go:15: compiled jsx_string_attribute_value.a mutant caught at output line 3: got "true", Go "false"

27. `TestSlot01Wave4HexMatchesCohere`

   slot01_wave4_test.go:146: compiled collapse_is_hex_digit.a mutant caught at output line 64: got "false", Go "true"

28. `TestSlot01Wave4ContainerMatchesCohere`

   slot01_wave4_test.go:151: compiled collapse_node_is_container.a mutant caught at output line 3: got "false", Go "true"

29. `TestSlot01Wave4StringDecoderMutant`

   slot01_wave4_test.go:156: compiled jsx_string_attribute_value.a mutant caught at output line 64: got "&#47;about", Go "/about"

30. `TestSlot01Wave4StringNilFirstMutant`

   slot01_wave4_test.go:159: compiled jsx_string_attribute_value.a mutant caught at output line 75: got "true", Go "false"

31. `TestSlot01Wave4HexDomainRefusal`

   slot01_wave4_test.go:218: compiled byte-domain guard mutant accepted 256 and printed false; source/native/emitted-JS refusal checks catch it

32. `TestSlot01Wave5NamedMatchesCohere`

   slot01_wave5_test.go:142: compiled collapse_valid_named_value.a mutant caught at output line 637: got "false", Go "true"

33. `TestSlot01Wave5NamedEmptyMutant`

   slot01_wave5_test.go:145: compiled collapse_valid_named_value.a mutant caught at output line 1: got "true", Go "false"

34. `TestSlot01Wave5ArbitraryMatchesCohere`

   slot01_wave5_test.go:150: compiled collapse_valid_arbitrary.a mutant caught at output line 4: got "true", Go "false"

35. `TestSlot01Wave5ArbitraryFinalMutant`

   slot01_wave5_test.go:153: compiled collapse_valid_arbitrary.a mutant caught at output line 648: got "false", Go "true"

36. `TestSlot01Wave5ArbitraryEscapeMutant`

   slot01_wave5_test.go:156: compiled collapse_valid_arbitrary.a mutant caught at output line 1021: got "true", Go "false"

37. `TestSlot01Wave5ThemeMatchesCohere`

   slot01_wave5_test.go:161: compiled collapse_parse_theme_options.a mutant caught at output line 1693: got "13 0:", Go "15 0:"

38. `TestSlot01Wave5ThemeLastPrefixMutant`

   slot01_wave5_test.go:164: compiled collapse_parse_theme_options.a mutant caught at output line 1695: got "0 5:102,105,114,115,116,", Go "0 6:115,101,99,111,110,100,"

39. `TestSlot01Wave6DefinitionMatchesCohere`

   slot01_wave6_test.go:156: compiled collapse_normalize_utility_definition.a mutant caught at output line 4: got "43:99,97,108,99,40,45,45,109,111,100,105,102,105,101,114,40,45,45,116,101,120,116,45,42,45,45,108,105,110,101,45,104,101,105,103,104,116,41,32,42,32,50,41,", Go "42:99,97,108,99,40,45,45,109,111,100,105,102,105,101,114,40,45,45,116,101,120,116,32,45,45,108,105,110,101,45,104,101,105,103,104,116,41,32,42,32,50,41,"

40. `TestSlot01Wave6DefinitionForwardMutant`

   slot01_wave6_test.go:159: compiled collapse_normalize_utility_definition.a mutant caught at output line 10: got "42:99,97,108,99,40,45,45,109,111,100,105,102,105,101,114,40,45,45,116,101,120,116,32,45,45,108,105,110,101,45,104,101,105,103,104,116,41,32,42,32,50,41,", Go "43:99,97,108,99,40,45,45,109,111,100,105,102,105,101,114,40,45,45,116,101,120,116,45,42,45,45,108,105,110,101,45,104,101,105,103,104,116,41,32,42,32,50,41,"

41. `TestSlot01Wave6WalkerMatchesCohere`

   slot01_wave6_test.go:164: compiled collapse_normalize_value_function_arguments.a mutant caught at output line 11: got "42:99,97,108,99,40,45,45,109,111,100,105,102,105,101,114,40,45,45,116,101,120,116,32,45,45,108,105,110,101,45,104,101,105,103,104,116,41,32,42,32,50,41,", Go "43:99,97,108,99,40,45,45,109,111,100,105,102,105,101,114,40,45,45,116,101,120,116,45,42,45,45,108,105,110,101,45,104,101,105,103,104,116,41,32,42,32,50,41,"

42. `TestSlot01Wave6WalkerPresentMutant`

   slot01_wave6_test.go:167: compiled collapse_normalize_value_function_arguments.a mutant caught at output line 7377: got "164:64,105,109,112,111,114,116,32,39,116,97,105,108,119,105,110,100,99,115,115,39,59,10,64,116,104,101,109,101,32,123,32,45,45,102,114,111,98,110,105,99,97,116,101,45,115,109,97,108,108,58,32,50,112,120,59,32,125,10,64,117,116,105,108,105,116,121,32,103,117,116,116,101,114,45,42,32,123,10,32,32,32,32,38,32,62,32,58,110,111,116,40,58,108,97,115,116,45,99,104,105,108,100,41,32,123,32,109,97,114,103,105,110,45,105,110,108,105,110,101,45,101,110,100,58,32,45,45,118,97,108,117,101,40,45,45,102,114,111,98,110,105,99,97,116,101,45,42,44,91,108,101,110,103,116,104,93,41,59,32,125,10,125,10,", Go "165:64,105,109,112,111,114,116,32,39,116,97,105,108,119,105,110,100,99,115,115,39,59,10,64,116,104,101,109,101,32,123,32,45,45,102,114,111,98,110,105,99,97,116,101,45,115,109,97,108,108,58,32,50,112,120,59,32,125,10,64,117,116,105,108,105,116,121,32,103,117,116,116,101,114,45,42,32,123,10,32,32,32,32,38,32,62,32,58,110,111,116,40,58,108,97,115,116,45,99,104,105,108,100,41,32,123,32,109,97,114,103,105,110,45,105,110,108,105,110,101,45,101,110,100,58,32,45,45,118,97,108,117,101,40,45,45,102,114,111,98,110,105,99,97,116,101,45,42,44,32,91,108,101,110,103,116,104,93,41,59,32,125,10,125,10,"

43. `TestSlot01Wave6WalkerKindMutant`

   slot01_wave6_test.go:170: compiled collapse_normalize_value_function_arguments.a mutant caught at output line 7373: got "164:64,105,109,112,111,114,116,32,39,116,97,105,108,119,105,110,100,99,115,115,39,59,10,64,116,104,101,109,101,32,123,32,45,45,102,114,111,98,110,105,99,97,116,101,45,115,109,97,108,108,58,32,50,112,120,59,32,125,10,64,117,116,105,108,105,116,121,32,103,117,116,116,101,114,45,42,32,123,10,32,32,32,32,38,32,62,32,58,110,111,116,40,58,108,97,115,116,45,99,104,105,108,100,41,32,123,32,109,97,114,103,105,110,45,105,110,108,105,110,101,45,101,110,100,58,32,45,45,118,97,108,117,101,40,45,45,102,114,111,98,110,105,99,97,116,101,45,42,44,91,108,101,110,103,116,104,93,41,59,32,125,10,125,10,", Go "165:64,105,109,112,111,114,116,32,39,116,97,105,108,119,105,110,100,99,115,115,39,59,10,64,116,104,101,109,101,32,123,32,45,45,102,114,111,98,110,105,99,97,116,101,45,115,109,97,108,108,58,32,50,112,120,59,32,125,10,64,117,116,105,108,105,116,121,32,103,117,116,116,101,114,45,42,32,123,10,32,32,32,32,38,32,62,32,58,110,111,116,40,58,108,97,115,116,45,99,104,105,108,100,41,32,123,32,109,97,114,103,105,110,45,105,110,108,105,110,101,45,101,110,100,58,32,45,45,118,97,108,117,101,40,45,45,102,114,111,98,110,105,99,97,116,101,45,42,44,32,91,108,101,110,103,116,104,93,41,59,32,125,10,125,10,"

44. `TestSlot01Wave6WalkerBailMutant`

   slot01_wave6_test.go:173: compiled collapse_normalize_value_function_arguments.a mutant caught at output line 29: got "trace:2", Go "trace:1"

45. `TestSlot01Wave6WalkerOrderMutant`

   slot01_wave6_test.go:178: compiled collapse_normalize_value_function_arguments.a mutant caught at output line 7380: got "165:64,105,109,112,111,114,116,32,39,116,97,105,108,119,105,110,100,99,115,115,39,59,10,64,116,104,101,109,101,32,123,32,45,45,102,114,111,98,110,105,99,97,116,101,45,115,109,97,108,108,58,32,50,112,120,59,32,125,10,64,117,116,105,108,105,116,121,32,103,117,116,116,101,114,45,42,32,123,10,32,32,32,32,38,32,62,32,58,110,111,116,40,58,108,97,115,116,45,99,104,105,108,100,41,32,123,32,109,97,114,103,105,110,45,105,110,108,105,110,101,45,101,110,100,58,32,45,45,118,97,108,117,101,40,45,45,102,114,111,98,110,105,99,97,116,101,45,42,44,32,91,108,101,110,103,116,104,93,41,59,32,125,10,125,10,", Go "42:99,97,108,99,40,45,45,109,111,100,105,102,105,101,114,40,45,45,116,101,120,116,32,45,45,108,105,110,101,45,104,101,105,103,104,116,41,32,42,32,50,41,"

46. `TestSlot01Wave6FrameworkMatchesCohere`

   slot01_wave6_test.go:187: compiled collapse_register_framework_variants.a mutant caught at output line 2: got "1000", Go "3"

47. `TestSlot01Wave6FrameworkLastOrderMutant`

   slot01_wave6_test.go:190: compiled collapse_register_framework_variants.a mutant caught at output line 1: got "3 5", Go "5 5"

48. `TestSlot01Wave6FrameworkCopyMutant`

   slot01_wave6_test.go:193: compiled collapse_register_framework_variants.a mutant caught at output line 26: got "8:116,97,109,112,101,114,101,100,", Go "6:115,116,97,116,105,99,"

49. `TestSlot01Wave6FrameworkDomainRefusal`

   slot01_wave6_test.go:252: compiled exact-integer guard mutant accepted 9007199254740993 as rounded 9007199254740992; source/native/emitted-JS refusal checks catch it

## Limits

No full repository test gate or whole-rule findings/fixes comparison was run. Prior callback integration and numeric-domain limits remain documented in the batch reports. No shared generator, harness or compiler file was changed. No new helper was claimed or built during this landing unit. Integration remains responsible for merging this branch.
