Built: ingestThemeBlock, NewTable and dissectClass, one .a helper per file; 17 prerequisites across six rules.\
Commits: claims 2d3df20, 8226408, 7b6baac; retained implementations 78dabb9, e487592, 009cf0f, all pushed on codex/lint-helpers-01.\
Commands and outputs: full helper package PASS 420.872s (63 tests), vet exit 0, uncached input oracle PASS 0.905s; setup 25s, nproc 5; all test output goes directly to retained logs.\
Mutants: 17 new compiling semantic mutants; all 49 prior/inherited checks caught again (64 semantic plus two refusal checks total). Every final witness follows below.\
Not covered: whole-rule parity, dynamic fixture generation, arbitrary cross-slot dependency integration or the full repository test gate; utility and variant reservations withdrawn.

# Slot 01 seventh retained helper batch

## Behavior and dependency contracts

Theme ingestion validates and updates only nonempty prefixes before walking. Keyframes return Skip, comments Continue, custom declarations call unescape then Add, and containers Continue. A store error or disallowed leaf stops traversal, preserving earlier changes. Options parsing, prefix validity, Go quoting, Walk, IsContainer, identifier unescaping and Theme.Add are explicit externally owned dependencies. Flat arena indices preserve node identity; an invalid index refuses instead of returning clean output.

NewTable returns an explicit missing-system projection for a nil Go system. A present system receives fresh descriptor and static maps, sharing descriptor handles, theme identity and PropertyOrder. Reading structs copy count and length/capacity headers while sharing backing values. Composition calls run theme namespaces, repository statics and repository functional roots in that order. Base descriptors, framework registrations/reading factory and those composition calls remain explicit dependencies. Descriptor handles refer to complete external objects; no descriptor fields are erased by this helper.

Class dissection splits on the last colon and retains that colon with the variants. It removes exactly one trailing importance marker from the base. Leading markers, brackets and escaped colons receive Go's literal behavior. Empty bases and supplementary Unicode text are preserved. The additional no-deprecated-classes fixture exercises the defining rule even though it lies outside the frozen blocked cohort.

Each helper occupies one production .a file. Existing rule entries, shared registration/test harness and protected compiler files are untouched. New slot-owned runners are .a. The inherited options_json.ts module remains unchanged; temporary mutant copies rename it to .a. No PR is opened and no main or area branch is pushed.

## Landing, selection and ownership

The previous branch was landing-ready at 1d748576fcbeef2d5dbe949b8dcfb7ff44be7646 on current main e8ba3d5, with all prior oracles freshly green. This turn verified both main ancestry and the live remote SHA before new work. Every origin codex/lint-helpers* branch was explicitly wildcard-fetched, since the default origin refspec fetches main only. All claims on eighteen matching branches were inspected, with the original comments bundle in HELPERS.md also excluded.

Theme ingestion and NewTable tied the highest available concrete count at six. Before the final replacement selection, the other six-consumer helpers had been reserved; dissectClass tied the highest available count at five. Claim commits were pushed successfully before code. The final snapshot, evidence/slot01-wave7/claims-final.json, finds only slot 01 claims for all three retained symbols.

The first utility-ingestion reservation lost to slot 04 bc590dc at 03:21:16 UTC, twenty-three seconds before slot 01 3475c2a at 03:21:39. Its duplicate source and tests are removed in 009cf0f and receive no readiness or mutant credit. Historical passing utility and header logs are explicitly withdrawn. The initial full regression was terminated and superseded after that collision appeared. ParseVariant was already held by slot 05; a failed availability assertion did not stop the shell and 672cadc mistakenly published a reservation. It was released immediately in 7b6baac; no variant code was written. Later reservation checks use failure-stopping shell semantics. These withdrawals leave exactly three retained new helpers.

## Commands and observations

Build shells source /workspace/adamic-tools/env.sh. All test stdout/stderr is directed to a log file without piping. Paths below are relative to this helper directory; commands run from repository root with that full prefix.

- `bash cloud/setup.sh > /tmp/lint-helpers-01-wave7-setup.log 2>&1`: PASS. Go, clang, Node and submodules ready at 0s; build cache warm 25s, done 25s; nproc 5, cgroup cpu.max 400000 100000, 17.6 GB. Retained as evidence/slot01-wave7/setup.log.
- `go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > evidence/slot01-wave7/final.log 2>&1`: PASS 420.872s, 63 top-level tests, final retained code at 009cf0f.
- `go vet ./... > evidence/slot01-wave7/vet.log 2>&1`: exit 0, empty log, final retained files.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > evidence/slot01-wave7/oracle.log 2>&1`: PASS 0.905s, six probe misses.
- Retained three-helper targeted run before the additional defining-rule fixture: PASS 147.071s, seventeen compiling semantic mutants; evidence/slot01-wave7/final-targeted.log.
- Final dissection gate with the defining-rule fixture: PASS 15.309s, three semantic mutants; evidence/slot01-wave7/dissect.log.
- Initial isolated theme gate: PASS 25.635s, six mutants. Initial table gate: PASS 67.552s, eight mutants. theme.log and table.log are bounded intermediate observations; final.log is the final combined evidence.
- `gofmt -l` over the four slot-owned Go files: empty format.log; `git diff --check`: empty diff-check.log.

Go cohere 715ba94f decides helper behavior. Source Node, sanitized native and emitted JavaScript match the captured observations. Native mutants must compile and finish with exit 0 and empty stderr before a stdout difference is credited. A compiler error, panic or sanitizer failure is not a semantic mutant catch.

The theme overlay records actual Go visitor return actions and actual Add arguments without changing branches or dependency behavior. The table overlay records its three actual composition calls and retains actual FrameworkStaticReading results for subsequent count/header/backing probes. No cohere worktree source is edited. Both functions run over 625 complete statically extractable string expressions from all six consuming files plus controls. Theme observations include manual node shapes and any successfully parsed CSS roots. They are not a claim that every JSX string is a stylesheet.

The table's metadata comes from actual Go: 78 base descriptor handles, 890 framework static registrations and 359 property positions. Nil/present controls produce 1,304 batches. It observes map sizes and three descriptor/static queries, callback ordering, namespaces, version and theme identity, plus source-map deletion, property-map writes and mutations of a returned reading's count/header/backing storage. Every framework reading is supplied by Go, but all arbitrary descriptor fields and all namespace-key bucket contents are not independently compared here.

Dissection covers 579 fixture strings from the five blocked consumers and 100 from no-deprecated-classes, plus controls: 706 batches and 2,118 UTF-16 observation lines. Class strings with multiple colons, escaped/bracketed colons, leading/trailing/repeated markers, empty bases and supplementary Unicode characters are included.

Observations while validating: native rejected a recursive function-valued local in the private theme runner; the runner now uses a named recursive function. A property-guard mutant initially survived because an earlier invalid clear stopped traversal; valid-clear controls now reach the guard and kill it. A missing-framework mutant initially made a probe panic; the probe now observes a missing capture without panicking, and the mutant runs successfully before its wrong table output is caught. Those early attempts are not credited as passing gates or semantic catches.

## Every consumer

Theme ingestion and NewTable each remove a listed dependency from all six:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

DissectClass removes a dependency from the same list except enforce-consistent-variant-order, five rules. The additional defining no-deprecated-classes fixture is coverage and is not counted as frozen readiness. Thus this batch removes seventeen prerequisite entries across six rules, with zero additional rules becoming fully helper-ready. Cumulative slot-only readiness remains the three structure rules from the first batch. slot01_wave7_readiness.json lists every residual helper after this batch and after all twenty-one retained slot 01 helpers. Other workers' unmerged helpers are not credited.

## Every new mutant and its independent witness

1. `TestSlot01Wave7ThemeMatchesCohere`

   slot01_wave7_test.go:217: compiled collapse_ingest_theme_block.a mutant caught at output line 1: got "107:102,105,120,116,117,114,101,46,99,115,115,58,32,96,64,116,104,101,109,101,96,32,98,108,111,99,107,115,32,109,117,115,116,32,111,110,108,121,32,99,111,110,116,97,105,110,32,99,117,115,116,111,109,32,112,114,111,112,101,114,116,105,101,115,32,111,114,32,96,64,107,101,121,102,114,97,109,101,115,96,44,32,102,111,117,110,100,32,100,101,99,108,97,114,97,116,105,111,110,32,34,99,111,108,111,114,34,", Go "64:102,105,120,116,117,114,101,46,99,115,115,58,32,48,58,32,73,110,118,97,108,105,100,32,116,104,101,109,101,32,118,97,108,117,101,32,96,96,32,102,111,114,32,110,97,109,101,115,112,97,99,101,32,96,45,45,99,111,108,111,114,45,42,96,"

2. `TestSlot01Wave7ThemePrefixMutant`

   slot01_wave7_test.go:220: compiled collapse_ingest_theme_block.a mutant caught at output line 3157: got "3:111,108,100,", Go "2:111,107,"

3. `TestSlot01Wave7ThemePropertyMutant`

   slot01_wave7_test.go:223: compiled collapse_ingest_theme_block.a mutant caught at output line 3199: got "0:", Go "109:102,105,120,116,117,114,101,46,99,115,115,58,32,96,64,116,104,101,109,101,96,32,98,108,111,99,107,115,32,109,117,115,116,32,111,110,108,121,32,99,111,110,116,97,105,110,32,99,117,115,116,111,109,32,112,114,111,112,101,114,116,105,101,115,32,111,114,32,96,64,107,101,121,102,114,97,109,101,115,96,44,32,102,111,117,110,100,32,100,101,99,108,97,114,97,116,105,111,110,32,34,105,110,105,116,105,97,108,34,"

4. `TestSlot01Wave7ThemeUnescapeMutant`

   slot01_wave7_test.go:227: compiled collapse_ingest_theme_block.a mutant caught at output line 4: got "11:45,45,99,111,108,111,114,45,92,54,49, 0: 0", Go "9:45,45,99,111,108,111,114,45,97, 0: 0"

5. `TestSlot01Wave7ThemeErrorStopMutant`

   slot01_wave7_test.go:230: compiled collapse_ingest_theme_block.a mutant caught at output line 1: got "102:102,105,120,116,117,114,101,46,99,115,115,58,32,96,64,116,104,101,109,101,96,32,98,108,111,99,107,115,32,109,117,115,116,32,111,110,108,121,32,99,111,110,116,97,105,110,32,99,117,115,116,111,109,32,112,114,111,112,101,114,116,105,101,115,32,111,114,32,96,64,107,101,121,102,114,97,109,101,115,96,44,32,102,111,117,110,100,32,100,101,99,108,97,114,97,116,105,111,110,32,34,34,", Go "64:102,105,120,116,117,114,101,46,99,115,115,58,32,48,58,32,73,110,118,97,108,105,100,32,116,104,101,109,101,32,118,97,108,117,101,32,96,96,32,102,111,114,32,110,97,109,101,115,112,97,99,101,32,96,45,45,99,111,108,111,114,45,42,96,"

6. `TestSlot01Wave7ThemeInvalidPrefixMutant`

   slot01_wave7_test.go:233: compiled collapse_ingest_theme_block.a mutant caught at output line 3161: got "69:102,105,120,116,117,114,101,46,99,115,115,58,32,48,58,32,73,110,118,97,108,105,100,32,116,104,101,109,101,32,118,97,108,117,101,32,96,102,111,111,45,42,96,32,102,111,114,32,110,97,109,101,115,112,97,99,101,32,96,45,45,99,111,108,111,114,45,42,96,", Go "93:102,105,120,116,117,114,101,46,99,115,115,58,32,116,104,101,32,112,114,101,102,105,120,32,34,66,65,68,34,32,105,115,32,105,110,118,97,108,105,100,46,32,80,114,101,102,105,120,101,115,32,109,117,115,116,32,98,101,32,108,111,119,101,114,99,97,115,101,32,65,83,67,73,73,32,108,101,116,116,101,114,115,32,40,97,45,122,41,32,111,110,108,121,"

7. `TestSlot01Wave7TableMatchesCohere`

   slot01_wave7_test.go:238: compiled collapse_new_table.a mutant caught at output line 2: got "false 79 891 0 0", Go "false 0 0 0 0"

8. `TestSlot01Wave7TableReadingCopyMutant`

   slot01_wave7_test.go:241: compiled collapse_new_table.a mutant caught at output line 41: got "99999 0 0", Go "1 1 1 99999"

9. `TestSlot01Wave7TablePropertyAliasMutant`

   slot01_wave7_test.go:244: compiled collapse_new_table.a mutant caught at output line 44: got "297", Go "99999"

10. `TestSlot01Wave7TableOrderMutant`

   slot01_wave7_test.go:247: compiled collapse_new_table.a mutant caught at output line 25: got "table.addRepositoryStatics(system),table.addThemeNamespaces(system.theme),table.addRepositoryFunctionalRoots(system)", Go "table.addThemeNamespaces(system.theme),table.addRepositoryStatics(system),table.addRepositoryFunctionalRoots(system)"

11. `TestSlot01Wave7TableDescriptorMapMutant`

   slot01_wave7_test.go:251: compiled collapse_new_table.a mutant caught at output line 35: got "true 78 891 359 7", Go "true 79 891 359 7"

12. `TestSlot01Wave7TableFrameworkMutant`

   slot01_wave7_test.go:254: compiled collapse_new_table.a mutant caught at output line 24: got "true 79 1 359 7", Go "true 79 891 359 7"

13. `TestSlot01Wave7TableHeaderCopyMutant`

   slot01_wave7_test.go:258: compiled collapse_new_table.a mutant caught at output line 41: got "1 0 0", Go "1 1 1 99999"

14. `TestSlot01Wave7TableBackingShareMutant`

   slot01_wave7_test.go:261: compiled collapse_new_table.a mutant caught at output line 41: got "1 1 1 39", Go "1 1 1 99999"

15. `TestSlot01Wave7DissectMatchesCohere`

   slot01_wave7_test.go:266: compiled tailwind_dissect_class.a mutant caught at output line 91: got "35:99,111,110,115,116,32,101,108,101,109,101,110,116,32,61,32,60,100,105,118,32,99,108,97,115,115,78,97,109,101,61,34,115,109,58,", Go "43:99,111,110,115,116,32,101,108,101,109,101,110,116,32,61,32,60,100,105,118,32,99,108,97,115,115,78,97,109,101,61,34,115,109,58,112,120,45,52,32,115,109,58,"

16. `TestSlot01Wave7DissectSuffixMutant`

   slot01_wave7_test.go:269: compiled tailwind_dissect_class.a mutant caught at output line 332: got "5:112,120,45,52,33,", Go "4:112,120,45,52,"

17. `TestSlot01Wave7DissectColonMutant`

   slot01_wave7_test.go:272: compiled tailwind_dissect_class.a mutant caught at output line 7: got "55:99,111,110,115,116,32,101,61,60,100,105,118,32,104,114,101,102,61,34,34,32,72,82,69,70,61,34,38,35,52,55,59,97,98,111,117,116,34,32,123,46,46,46,112,114,111,112,115,125,32,120,108,105,110,107,", Go "56:99,111,110,115,116,32,101,61,60,100,105,118,32,104,114,101,102,61,34,34,32,72,82,69,70,61,34,38,35,52,55,59,97,98,111,117,116,34,32,123,46,46,46,112,114,111,112,115,125,32,120,108,105,110,107,58,"



The final log also reruns every inherited/prior check. The final regression catches 64 compiling semantic mutants and two explicit domain-refusal mutants, 66 checks in all. Existing SLOT01_REPORT.md and SLOT01_WAVE2_REPORT.md through SLOT01_WAVE6_REPORT.md describe their implementations. All current witnesses are retained in final.log.

## Limits

No whole-rule findings, spans, fixes or suggestions comparison, dynamic fixture generation, shared-profile compilation or registration, or complete dependency integration is claimed. Callback behavior and complete external arena storage remain separate work. Table handles describe complete external objects but this gate checks a bounded projection and does not inspect every descriptor field or every namespace-key value. Reading headers are observable; arbitrary future append/reslice operations on backing storage are outside this helper. Malformed raw UTF-8, isolated UTF-16 surrogates, invalid/cyclic arenas, concurrent mutation and arbitrary adapter-domain misuse are not covered. The full repository go test ./... gate was not run; the entire touched helper package, repository-wide vet and a filtered uncached external oracle are the bounded gate.

## Every prior and inherited witness rerun

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
