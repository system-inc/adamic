Built: source-only partial .a ports of cloneNode, cloneNodes and removeNodes; zero readiness credit, stopped at an ownership-lowering blocker.
Commits: original pushed claim 54013ad7; rebased claim fa8d93a5 on current main c01907a7; partial implementation b0134e282d258445958331deafe91087f07db271; final report SHA in handoff.
Commands and outputs: setup PASS 26s, nproc 5; full helper landing run PASS 595.735s with 76 passes and eight explicit skips; vet clean; uncached input oracle PASS 1.202s.
Mutants: eight new source-only mutants caught by Go/Node comparisons, zero new native mutant credit; all 79 prior compiled checks caught again, 77 semantic and two refusal checks.
Not covered: native/emitted-JavaScript tree-helper parity and native tree mutants blocked by adamic/cycle-capable; full repository gate and whole-rule integration not run; no shared files edited.

# Slot 01 eleventh helper batch: blocked

## Exact blocker and disposition

The source ports are reviewable but do not meet the native handoff bar. Lowering refuses the mutable node arena in gaps/slot01_wave11_main.a at its arena declaration and push:

> Adamic 0.1 refuses CollapseCopyNode[], an array whose elements can reach back to an array like it: a cycle reference counting can't free ... (adamic/cycle-capable)

The complete diagnostic, including file/line positions and suggested alternatives, is in evidence/slot01-wave11/landing.log. The earlier mutable-parameter design was refused as well. A read-only arena API and caller-owned allocator still trigger the refusal at the driver's mutable arena. That attempt is preserved in readonly-probe.log. No compiler or shared harness change is made to bypass it. The current driver resides under the existing gaps/ exclusion; no shared configuration changed.

The tests compare actual Go to source Node and run their source-only mutant before trying lowering. Only the specific cycle-capable refusal naming CollapseCopyNode[] causes a clearly labeled BLOCKED skip. Any unrelated error fails the test. If the ownership refusal closes, baseline native, emitted-JavaScript and sanitized-native mutant checks resume automatically. A refusal is never credited as a semantic mutant caught.

Per the unit instruction to state the blocker and stop, no further helper is claimed. All three remain reserved but blocked, and their partial sources are pushed for review. slot01_wave11_readiness.json deliberately has an empty completed helpers list, three blocked_helpers, and zero dependency_entries_removed. No rule worker should count these helpers as native-ready.

## Landing and ownership

Only codex/lint-helpers-01 was pushed in this thread. At reservation, the branch was landing-ready on f8013f0. All eighteen origin codex/lint-helpers* branches and their claims were fetched and checked, including dash-suffixed workers. The base comment bundle remains reserved in HELPERS.md. These three tied the largest unclaimed concrete count at four each. Claim 54013ad7 was pushed before source creation. The final wildcard audit finds only slot 01 mentions for these three symbols.

Main advanced to c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 during the final check. The 82 existing commits rebased without conflicts, incorporating the incoming stage3 landing. All previous completed helpers were rerun on that main before publication. The final exact worker lease was independently observed as 54013ad75c15525c8e03b943447f8f58fe11f9cb. Publication uses an exact force-with-lease on this own branch, authorized by the user's rebase-and-push instruction. No main or area/ branch is pushed and no pull request opened. Historic SHAs in earlier reports remain historical provenance.

No shared allocator-check diff was present in this main delta; no such diff was reverted or overridden. No shared Diagnostic landing SHA was supplied. This batch introduces no regexp matcher and does not port any rule-local Go regex. It creates no rule listener/dispatcher, so no relevance dispatch or rule.json is modified.

## Consumers and readiness

Each blocked helper is listed by the same four frozen rules:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

There are twelve potential prerequisite entries, but all twelve remain listed because native validation is blocked. Zero prerequisite or final-blocker removals are claimed. The original readiness inventory is unchanged; the separate ledger preserves every consumer's full remaining dependency list.

## What is ported and observed

cloneNode preserves nil nodes and every one of Go's ten Node fields, copies present context maps, and delegates recursive child-list copying through a caller-owned allocator with numeric handles. A reflection check fails if the Go field inventory changes. cloneNodes preserves nil-versus-empty slices and clones each occurrence independently, including shared-DAG nodes and nil entries. removeNodes returns the unchanged header/backing for an empty set, skips only true-valued map entries, recurses through nonempty children regardless of kind, mutates retained original nodes and keeps empty containers.

The slice adapter carries nil, length, capacity and backing values explicitly. This preserves spare-capacity nil slots, header-copy behavior and backing identity; a plain growing array would lose those distinctions. Context presence distinguishes a nil map from an empty map. Inputs are valid acyclic trees or DAGs with well-formed handles/headers; Go's own helpers do not terminate on cycles. See SLOT01_WAVE11_README.md for proposed APIs and preconditions.

A virtual Go overlay calls the actual private functions without changing cohere's worktree. The four fixture files contribute 460 complete string expressions: 155 class order, 67 shorthand, 109 conflicts, 129 unknown classes. Concatenations and local constants remain whole. Five fixed inputs bring the corpus to 465. Each contributes eleven tree variants: nil/empty lists, spare capacity, every kind including unknown, nil children, shared children/roots, context nil/empty/present, true/false removal entries, all-dropped children, and actual ParseCSS where successful. There are 5,115 cases per helper, including 345 actual parsed CSS trees. Parse failures fall back to explicit manual trees and are not treated as helper failures.

Observations include all fields through UTF-16 unit output, context contents/presence, list nil/length/capacity and backing slots, original-pointer identity, clone occurrence identity, source state after clone mutations, retained-source mutations after removal, source-context mutations and output backing behavior after changing the source root list. Source Node matches 783,668 Go lines for cloneNode, 1,159,052 for cloneNodes and 566,371 for removeNodes: 2,509,091 lines for one pass of each helper. Additional mutant tests repeat those observations. Native parity is not inferred from these matches.

## Commands and outputs

All test processes write directly to logs, never through pipes. source /workspace/adamic-tools/env.sh supplies Go 1.27.1, clang 20.1.8 and Node 24.19.0.

- bash cloud/setup.sh > /tmp/lint-helpers-01-wave11-setup.log 2>&1: go 0s, clang 0s, node 0s, submodules 0s, build cache warm 26s, done 26s; nproc 5. Copy: setup.log.
- Initial go test ./stage1/cohere/lint/helpers -run '^TestSlot01Wave11' -count=1 -v -timeout=20m: FAIL 23.051s at ownership lowering, preserved in mutable-arena-gap.log. No native mutant credit.
- Read-only allocator probe with -run '^TestSlot01Wave11Node$': FAIL 2.881s at the same driver's arena ownership, readonly-probe.log. No native mutant credit.
- Source-only/known-gap targeted rerun with -run '^TestSlot01Wave11': PASS 30.460s with eight explicit BLOCKED skips after Go/Node comparisons and all eight source-only mutants; final.log. This is not a native pass.
- git rebase origin/main > /tmp/lint-helpers-01-wave11-rebase.log 2>&1: 82 commits, no conflicts. Copy: rebase.log.
- go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/evidence/slot01-wave11/landing.log 2>&1: PASS 595.735s, 76 passing tests plus eight blocked skips. All 79 prior compiled mutant checks caught again on c01907a7. All eight new source-only checks rerun there; zero new compiled-mutant credit.
- go vet ./... > stage1/cohere/lint/helpers/evidence/slot01-wave11/vet.log 2>&1: current-main rerun exit zero, empty log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/evidence/slot01-wave11/oracle.log 2>&1: current-main PASS 1.202s, six uncached probes.
- gofmt -l over the three new Go files: empty format.log. git diff --check: empty diff-check.log.

The complete helper package, repository vet and filtered uncached external oracle were run. The full repository gate was not run. Native executions for prior delivered helpers remain sanitized and must finish with zero stderr. New tree checks stop before native construction at the explicit refusal, so there is no claim about their leaks or native behavior.

## Every new source-only mutant

These execute only as source on Node. All are caught by stdout differences from actual Go. They are not native semantic kills.

| Helper | Mutation | Witness in current-main landing.log |
|---|---|---|
| cloneNode | clear Important | line 43: mutant true:false:true, Go true:true:true |
| cloneNode | share source context map | line 143: mutant 3 context entries, Go 2 |
| cloneNode | share source child list | line 33: mutant source identity 1, Go fresh -2 |
| cloneNodes | return original handles after invoking callback | line 10: mutant original 0, Go fresh -2 |
| cloneNodes | turn nil list into empty non-nil | line 1: mutant false:0:0, Go true:0:0 |
| removeNodes | treat false-valued membership as removal | line 719: mutant false:0:2, Go false:2:2 |
| removeNodes | allocate for empty removal set | line 1: mutant false:0:0, Go true:0:0 |
| removeNodes | prune retained empty containers | line 453: mutant false:0:2, Go false:2:2 |

## Every prior compiled mutant witness after rebase

The following actual current-main log witnesses retain their prior credit: 77 semantic stdout differences and two domain-refusal checks. Mutation definitions remain in the earlier reports/tests. No new tree-helper mutation is included here.

- helpers_test.go:145: compiled semantic mutant caught at output line 15157: got "valid", Go "invalid"
- helpers_test.go:145: compiled semantic mutant caught at output line 15166: got "valid", Go "invalid"
- helpers_test.go:145: compiled semantic mutant caught at output line 22166: got "This throws a bare `{{constructor}}`, which names no declared failure. Raise it through the tier that declares it, `AccountModule.error(identifier, data, cause)`, `ApiWorker.error(...)` or `Base.error(...)`. A bare throw carries no identifier, so the board groups it by its message and one interpolated value mints one identity per value, and it normalizes to 500, so a refusal reads as our fault.", Go "This throws a bare `sentinel é😀`, which names no declared failure. Raise it through the tier that declares it, `AccountModule.error(identifier, data, cause)`, `ApiWorker.error(...)` or `Base.error(...)`. A bare throw carries no identifier, so the board groups it by its message and one interpolated value mints one identity per value, and it normalizes to 500, so a refusal reads as our fault."
- helpers_test.go:145: compiled semantic mutant caught at output line 15178: got "valid", Go "invalid"
- slot01_test.go:183: compiled structure_file_context.a mutant caught at output line 7: got "true false false false false false false false false", Go "true true false false false false true false false"
- slot01_test.go:183: compiled react_es6_component_class.a mutant caught at output line 13400: got "false", Go "true"
- slot01_test.go:183: compiled tailwind_default_class_literal_settings.a mutant caught at output line 1: got "className;mergeClassNames|createVariantClassNames;.*[Cc]lassName$|.*[Cc]lassNames$", Go "class|className;mergeClassNames|createVariantClassNames;.*[Cc]lassName$|.*[Cc]lassNames$"
- slot01_test.go:206: compiled shared attributeNames mutant caught at output line 3: got "mutated attribute|className;mergeClassNames|createVariantClassNames;.*[Cc]lassName$|.*[Cc]lassNames$", Go "class|className;mergeClassNames|createVariantClassNames;.*[Cc]lassName$|.*[Cc]lassNames$"
- slot01_test.go:206: compiled shared calleeNames mutant caught at output line 3: got "class|className;mutated callee|createVariantClassNames;.*[Cc]lassName$|.*[Cc]lassNames$", Go "class|className;mergeClassNames|createVariantClassNames;.*[Cc]lassName$|.*[Cc]lassNames$"
- slot01_test.go:206: compiled shared variablePatterns mutant caught at output line 3: got "class|className;mergeClassNames|createVariantClassNames;mutated pattern|.*[Cc]lassNames$", Go "class|className;mergeClassNames|createVariantClassNames;.*[Cc]lassName$|.*[Cc]lassNames$"
- slot01_wave10_test.go:155: compiled regexp_word_class_atoms.a mutant caught at output line 19: got "6", Go "4"
- slot01_wave10_test.go:160: compiled regexp_nonword_class_atoms.a mutant caught at output line 10: got "1:123:65535", Go "1:123:1114111"
- slot01_wave10_test.go:165: compiled regexp_word_boundary.a mutant caught at output line 1: got "(?:(?<=[0-9A-Z\\u005fa-z])(?=[0-9A-Z\\u005fa-z])|(?<![0-9A-Z\\u005fa-z])(?![0-9A-Z\\u005fa-z]))", Go "(?:(?<![0-9A-Z\\u005fa-z])(?=[0-9A-Z\\u005fa-z])|(?<=[0-9A-Z\\u005fa-z])(?![0-9A-Z\\u005fa-z]))"
- slot01_wave10_test.go:170: compiled regexp_word_class_atoms.a mutant caught at output line 11: got "1:99999:57", Go "1:48:57"
- slot01_wave10_test.go:175: compiled regexp_nonword_class_atoms.a mutant caught at output line 13: got "1:99999:47", Go "1:0:47"
- slot01_wave10_test.go:180: compiled regexp_word_boundary.a mutant caught at output line 9: got "0", Go "1"
- slot01_wave10_test.go:185: compiled regexp_word_boundary.a mutant caught at output line 2: got "2", Go "1"
- slot01_wave2_test.go:13: compiled tailwind_new_class_literal_reader.a mutant caught at output line 14: got "^custom$", Go ".*[Cc]lassName$"
- slot01_wave2_test.go:16: compiled tailwind_class_values_in.a mutant caught at output line 14: got "", Go "!"
- slot01_wave2_test.go:19: compiled tailwind_class_literal_from.a mutant caught at output line 2: got "225 235 false false", Go "226 235 false false"
- slot01_wave2_test.go:22: compiled tailwind_class_literal_from.a mutant caught at output line 522: got "1 0 false false", Go "0 1 false false"
- slot01_wave2_test.go:25: compiled tailwind_new_class_literal_reader.a mutant caught at output line 1: got "3 3 6 true", Go "3 3 5 true"
- slot01_wave2_test.go:28: compiled tailwind_new_class_literal_reader.a mutant caught at output line 2: got "false true", Go "true true"
- slot01_wave2_test.go:31: compiled tailwind_class_values_in.a mutant caught at output line 1: got "1 0 1", Go "1 0 0"
- slot01_wave2_test.go:34: compiled tailwind_class_values_in.a mutant caught at output line 19: got "0 0 2", Go "0 0 3"
- slot01_wave3_test.go:19: compiled tailwind_attribute_values.a mutant caught at output line 5: got "1 0", Go "0 0"
- slot01_wave3_test.go:161: compiled text_decode_entity.a mutant caught at output line 20: got "0", Go "38,35,49,49,49,52,49,49,50,59"
- slot01_wave3_test.go:164: compiled text_decode_entity.a mutant caught at output line 5384: got "55832", Go "65533"
- slot01_wave3_test.go:169: compiled react_likely_component_name.a mutant caught at output line 580: got "false", Go "true"
- slot01_wave3_test.go:191: compiled tailwind_attribute_values.a mutant caught at output line 5: got "1 0", Go "0 0"
- slot01_wave3_test.go:194: compiled text_decode_entity.a mutant caught at output line 50: got "38,97,109,112,59", Go "38"
- slot01_wave3_test.go:199: compiled react_likely_component_name.a mutant caught at output line 584: got "true", Go "false"
- slot01_wave4_test.go:15: compiled jsx_string_attribute_value.a mutant caught at output line 3: got "true", Go "false"
- slot01_wave4_test.go:146: compiled collapse_is_hex_digit.a mutant caught at output line 64: got "false", Go "true"
- slot01_wave4_test.go:151: compiled collapse_node_is_container.a mutant caught at output line 3: got "false", Go "true"
- slot01_wave4_test.go:156: compiled jsx_string_attribute_value.a mutant caught at output line 64: got "&#47;about", Go "/about"
- slot01_wave4_test.go:159: compiled jsx_string_attribute_value.a mutant caught at output line 75: got "true", Go "false"
- slot01_wave4_test.go:218: compiled byte-domain guard mutant accepted 256 and printed false; source/native/emitted-JS refusal checks catch it
- slot01_wave5_test.go:142: compiled collapse_valid_named_value.a mutant caught at output line 637: got "false", Go "true"
- slot01_wave5_test.go:145: compiled collapse_valid_named_value.a mutant caught at output line 1: got "true", Go "false"
- slot01_wave5_test.go:150: compiled collapse_valid_arbitrary.a mutant caught at output line 4: got "true", Go "false"
- slot01_wave5_test.go:153: compiled collapse_valid_arbitrary.a mutant caught at output line 648: got "false", Go "true"
- slot01_wave5_test.go:156: compiled collapse_valid_arbitrary.a mutant caught at output line 1021: got "true", Go "false"
- slot01_wave5_test.go:161: compiled collapse_parse_theme_options.a mutant caught at output line 1693: got "13 0:", Go "15 0:"
- slot01_wave5_test.go:164: compiled collapse_parse_theme_options.a mutant caught at output line 1695: got "0 5:102,105,114,115,116,", Go "0 6:115,101,99,111,110,100,"
- slot01_wave6_test.go:156: compiled collapse_normalize_utility_definition.a mutant caught at output line 4: got "43:99,97,108,99,40,45,45,109,111,100,105,102,105,101,114,40,45,45,116,101,120,116,45,42,45,45,108,105,110,101,45,104,101,105,103,104,116,41,32,42,32,50,41,", Go "42:99,97,108,99,40,45,45,109,111,100,105,102,105,101,114,40,45,45,116,101,120,116,32,45,45,108,105,110,101,45,104,101,105,103,104,116,41,32,42,32,50,41,"
- slot01_wave6_test.go:159: compiled collapse_normalize_utility_definition.a mutant caught at output line 10: got "42:99,97,108,99,40,45,45,109,111,100,105,102,105,101,114,40,45,45,116,101,120,116,32,45,45,108,105,110,101,45,104,101,105,103,104,116,41,32,42,32,50,41,", Go "43:99,97,108,99,40,45,45,109,111,100,105,102,105,101,114,40,45,45,116,101,120,116,45,42,45,45,108,105,110,101,45,104,101,105,103,104,116,41,32,42,32,50,41,"
- slot01_wave6_test.go:164: compiled collapse_normalize_value_function_arguments.a mutant caught at output line 11: got "42:99,97,108,99,40,45,45,109,111,100,105,102,105,101,114,40,45,45,116,101,120,116,32,45,45,108,105,110,101,45,104,101,105,103,104,116,41,32,42,32,50,41,", Go "43:99,97,108,99,40,45,45,109,111,100,105,102,105,101,114,40,45,45,116,101,120,116,45,42,45,45,108,105,110,101,45,104,101,105,103,104,116,41,32,42,32,50,41,"
- slot01_wave6_test.go:167: compiled collapse_normalize_value_function_arguments.a mutant caught at output line 7377: got "164:64,105,109,112,111,114,116,32,39,116,97,105,108,119,105,110,100,99,115,115,39,59,10,64,116,104,101,109,101,32,123,32,45,45,102,114,111,98,110,105,99,97,116,101,45,115,109,97,108,108,58,32,50,112,120,59,32,125,10,64,117,116,105,108,105,116,121,32,103,117,116,116,101,114,45,42,32,123,10,32,32,32,32,38,32,62,32,58,110,111,116,40,58,108,97,115,116,45,99,104,105,108,100,41,32,123,32,109,97,114,103,105,110,45,105,110,108,105,110,101,45,101,110,100,58,32,45,45,118,97,108,117,101,40,45,45,102,114,111,98,110,105,99,97,116,101,45,42,44,91,108,101,110,103,116,104,93,41,59,32,125,10,125,10,", Go "165:64,105,109,112,111,114,116,32,39,116,97,105,108,119,105,110,100,99,115,115,39,59,10,64,116,104,101,109,101,32,123,32,45,45,102,114,111,98,110,105,99,97,116,101,45,115,109,97,108,108,58,32,50,112,120,59,32,125,10,64,117,116,105,108,105,116,121,32,103,117,116,116,101,114,45,42,32,123,10,32,32,32,32,38,32,62,32,58,110,111,116,40,58,108,97,115,116,45,99,104,105,108,100,41,32,123,32,109,97,114,103,105,110,45,105,110,108,105,110,101,45,101,110,100,58,32,45,45,118,97,108,117,101,40,45,45,102,114,111,98,110,105,99,97,116,101,45,42,44,32,91,108,101,110,103,116,104,93,41,59,32,125,10,125,10,"
- slot01_wave6_test.go:170: compiled collapse_normalize_value_function_arguments.a mutant caught at output line 7373: got "164:64,105,109,112,111,114,116,32,39,116,97,105,108,119,105,110,100,99,115,115,39,59,10,64,116,104,101,109,101,32,123,32,45,45,102,114,111,98,110,105,99,97,116,101,45,115,109,97,108,108,58,32,50,112,120,59,32,125,10,64,117,116,105,108,105,116,121,32,103,117,116,116,101,114,45,42,32,123,10,32,32,32,32,38,32,62,32,58,110,111,116,40,58,108,97,115,116,45,99,104,105,108,100,41,32,123,32,109,97,114,103,105,110,45,105,110,108,105,110,101,45,101,110,100,58,32,45,45,118,97,108,117,101,40,45,45,102,114,111,98,110,105,99,97,116,101,45,42,44,91,108,101,110,103,116,104,93,41,59,32,125,10,125,10,", Go "165:64,105,109,112,111,114,116,32,39,116,97,105,108,119,105,110,100,99,115,115,39,59,10,64,116,104,101,109,101,32,123,32,45,45,102,114,111,98,110,105,99,97,116,101,45,115,109,97,108,108,58,32,50,112,120,59,32,125,10,64,117,116,105,108,105,116,121,32,103,117,116,116,101,114,45,42,32,123,10,32,32,32,32,38,32,62,32,58,110,111,116,40,58,108,97,115,116,45,99,104,105,108,100,41,32,123,32,109,97,114,103,105,110,45,105,110,108,105,110,101,45,101,110,100,58,32,45,45,118,97,108,117,101,40,45,45,102,114,111,98,110,105,99,97,116,101,45,42,44,32,91,108,101,110,103,116,104,93,41,59,32,125,10,125,10,"
- slot01_wave6_test.go:173: compiled collapse_normalize_value_function_arguments.a mutant caught at output line 29: got "trace:2", Go "trace:1"
- slot01_wave6_test.go:178: compiled collapse_normalize_value_function_arguments.a mutant caught at output line 7380: got "165:64,105,109,112,111,114,116,32,39,116,97,105,108,119,105,110,100,99,115,115,39,59,10,64,116,104,101,109,101,32,123,32,45,45,102,114,111,98,110,105,99,97,116,101,45,115,109,97,108,108,58,32,50,112,120,59,32,125,10,64,117,116,105,108,105,116,121,32,103,117,116,116,101,114,45,42,32,123,10,32,32,32,32,38,32,62,32,58,110,111,116,40,58,108,97,115,116,45,99,104,105,108,100,41,32,123,32,109,97,114,103,105,110,45,105,110,108,105,110,101,45,101,110,100,58,32,45,45,118,97,108,117,101,40,45,45,102,114,111,98,110,105,99,97,116,101,45,42,44,32,91,108,101,110,103,116,104,93,41,59,32,125,10,125,10,", Go "42:99,97,108,99,40,45,45,109,111,100,105,102,105,101,114,40,45,45,116,101,120,116,32,45,45,108,105,110,101,45,104,101,105,103,104,116,41,32,42,32,50,41,"
- slot01_wave6_test.go:187: compiled collapse_register_framework_variants.a mutant caught at output line 2: got "1000", Go "3"
- slot01_wave6_test.go:190: compiled collapse_register_framework_variants.a mutant caught at output line 1: got "3 5", Go "5 5"
- slot01_wave6_test.go:193: compiled collapse_register_framework_variants.a mutant caught at output line 26: got "8:116,97,109,112,101,114,101,100,", Go "6:115,116,97,116,105,99,"
- slot01_wave6_test.go:252: compiled exact-integer guard mutant accepted 9007199254740993 as rounded 9007199254740992; source/native/emitted-JS refusal checks catch it
- slot01_wave7_test.go:217: compiled collapse_ingest_theme_block.a mutant caught at output line 1: got "107:102,105,120,116,117,114,101,46,99,115,115,58,32,96,64,116,104,101,109,101,96,32,98,108,111,99,107,115,32,109,117,115,116,32,111,110,108,121,32,99,111,110,116,97,105,110,32,99,117,115,116,111,109,32,112,114,111,112,101,114,116,105,101,115,32,111,114,32,96,64,107,101,121,102,114,97,109,101,115,96,44,32,102,111,117,110,100,32,100,101,99,108,97,114,97,116,105,111,110,32,34,99,111,108,111,114,34,", Go "64:102,105,120,116,117,114,101,46,99,115,115,58,32,48,58,32,73,110,118,97,108,105,100,32,116,104,101,109,101,32,118,97,108,117,101,32,96,96,32,102,111,114,32,110,97,109,101,115,112,97,99,101,32,96,45,45,99,111,108,111,114,45,42,96,"
- slot01_wave7_test.go:220: compiled collapse_ingest_theme_block.a mutant caught at output line 3157: got "3:111,108,100,", Go "2:111,107,"
- slot01_wave7_test.go:223: compiled collapse_ingest_theme_block.a mutant caught at output line 3199: got "0:", Go "109:102,105,120,116,117,114,101,46,99,115,115,58,32,96,64,116,104,101,109,101,96,32,98,108,111,99,107,115,32,109,117,115,116,32,111,110,108,121,32,99,111,110,116,97,105,110,32,99,117,115,116,111,109,32,112,114,111,112,101,114,116,105,101,115,32,111,114,32,96,64,107,101,121,102,114,97,109,101,115,96,44,32,102,111,117,110,100,32,100,101,99,108,97,114,97,116,105,111,110,32,34,105,110,105,116,105,97,108,34,"
- slot01_wave7_test.go:227: compiled collapse_ingest_theme_block.a mutant caught at output line 4: got "11:45,45,99,111,108,111,114,45,92,54,49, 0: 0", Go "9:45,45,99,111,108,111,114,45,97, 0: 0"
- slot01_wave7_test.go:230: compiled collapse_ingest_theme_block.a mutant caught at output line 1: got "102:102,105,120,116,117,114,101,46,99,115,115,58,32,96,64,116,104,101,109,101,96,32,98,108,111,99,107,115,32,109,117,115,116,32,111,110,108,121,32,99,111,110,116,97,105,110,32,99,117,115,116,111,109,32,112,114,111,112,101,114,116,105,101,115,32,111,114,32,96,64,107,101,121,102,114,97,109,101,115,96,44,32,102,111,117,110,100,32,100,101,99,108,97,114,97,116,105,111,110,32,34,34,", Go "64:102,105,120,116,117,114,101,46,99,115,115,58,32,48,58,32,73,110,118,97,108,105,100,32,116,104,101,109,101,32,118,97,108,117,101,32,96,96,32,102,111,114,32,110,97,109,101,115,112,97,99,101,32,96,45,45,99,111,108,111,114,45,42,96,"
- slot01_wave7_test.go:233: compiled collapse_ingest_theme_block.a mutant caught at output line 3161: got "69:102,105,120,116,117,114,101,46,99,115,115,58,32,48,58,32,73,110,118,97,108,105,100,32,116,104,101,109,101,32,118,97,108,117,101,32,96,102,111,111,45,42,96,32,102,111,114,32,110,97,109,101,115,112,97,99,101,32,96,45,45,99,111,108,111,114,45,42,96,", Go "93:102,105,120,116,117,114,101,46,99,115,115,58,32,116,104,101,32,112,114,101,102,105,120,32,34,66,65,68,34,32,105,115,32,105,110,118,97,108,105,100,46,32,80,114,101,102,105,120,101,115,32,109,117,115,116,32,98,101,32,108,111,119,101,114,99,97,115,101,32,65,83,67,73,73,32,108,101,116,116,101,114,115,32,40,97,45,122,41,32,111,110,108,121,"
- slot01_wave7_test.go:238: compiled collapse_new_table.a mutant caught at output line 2: got "false 79 891 0 0", Go "false 0 0 0 0"
- slot01_wave7_test.go:241: compiled collapse_new_table.a mutant caught at output line 41: got "99999 0 0", Go "1 1 1 99999"
- slot01_wave7_test.go:244: compiled collapse_new_table.a mutant caught at output line 44: got "297", Go "99999"
- slot01_wave7_test.go:247: compiled collapse_new_table.a mutant caught at output line 25: got "table.addRepositoryStatics(system),table.addThemeNamespaces(system.theme),table.addRepositoryFunctionalRoots(system)", Go "table.addThemeNamespaces(system.theme),table.addRepositoryStatics(system),table.addRepositoryFunctionalRoots(system)"
- slot01_wave7_test.go:251: compiled collapse_new_table.a mutant caught at output line 35: got "true 78 891 359 7", Go "true 79 891 359 7"
- slot01_wave7_test.go:254: compiled collapse_new_table.a mutant caught at output line 24: got "true 79 1 359 7", Go "true 79 891 359 7"
- slot01_wave7_test.go:258: compiled collapse_new_table.a mutant caught at output line 41: got "1 0 0", Go "1 1 1 99999"
- slot01_wave7_test.go:261: compiled collapse_new_table.a mutant caught at output line 41: got "1 1 1 39", Go "1 1 1 99999"
- slot01_wave7_test.go:266: compiled tailwind_dissect_class.a mutant caught at output line 91: got "35:99,111,110,115,116,32,101,108,101,109,101,110,116,32,61,32,60,100,105,118,32,99,108,97,115,115,78,97,109,101,61,34,115,109,58,", Go "43:99,111,110,115,116,32,101,108,101,109,101,110,116,32,61,32,60,100,105,118,32,99,108,97,115,115,78,97,109,101,61,34,115,109,58,112,120,45,52,32,115,109,58,"
- slot01_wave7_test.go:269: compiled tailwind_dissect_class.a mutant caught at output line 332: got "5:112,120,45,52,33,", Go "4:112,120,45,52,"
- slot01_wave7_test.go:272: compiled tailwind_dissect_class.a mutant caught at output line 7: got "55:99,111,110,115,116,32,101,61,60,100,105,118,32,104,114,101,102,61,34,34,32,72,82,69,70,61,34,38,35,52,55,59,97,98,111,117,116,34,32,123,46,46,46,112,114,111,112,115,125,32,120,108,105,110,107,", Go "56:99,111,110,115,116,32,101,61,60,100,105,118,32,104,114,101,102,61,34,34,32,72,82,69,70,61,34,38,35,52,55,59,97,98,111,117,116,34,32,123,46,46,46,112,114,111,112,115,125,32,120,108,105,110,107,58,"
- slot01_wave8_test.go:142: compiled regexp_decimal_escape.a mutant caught at output line 51577: got "123456789", Go "1234567"
- slot01_wave8_test.go:147: compiled regexp_legacy_octal.a mutant caught at output line 49739: got "302", Go "37"
- slot01_wave8_test.go:152: compiled regexp_class_atom_covers.a mutant caught at output line 6: got "false", Go "true"
- slot01_wave9_test.go:142: compiled regexp_bounded_quantifier_width.a mutant caught at output line 22: got "2", Go "0"
- slot01_wave9_test.go:147: compiled regexp_quantifier_width.a mutant caught at output line 25795: got "1", Go "2"
- slot01_wave9_test.go:152: compiled regexp_group_kind.a mutant caught at output line 25936: got "0", Go "2"

## Landing refresh on b8fb957a

Rebased the worker branch onto current main b8fb957aa839a9e8cb0b54279dd9864fa317bd30 without conflicts. The incoming inherited static-field read fix is retained. No shared files were edited. The pre-report rebased head was e2a4cef34e4c6ab90d9bbc001ee8675244b45f71.

Ran `go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m`: PASS 590.238s, 76 top-level passes and eight explicit ownership-blocked skips. The log records all 77 compiled semantic mutant catches, the two expected refusal checks, and all eight source-only tree mutant catches. Native/emitted-JavaScript validation of the tree helpers is still blocked by the same CollapseCopyNode[] adamic/cycle-capable refusal; no readiness credit is added.

`go vet ./stage1/cohere/lint/helpers` exited zero. `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m` passed in 1.052s. Logs are evidence/slot01-wave11/{landing,vet,input}-b8fb.log. No full repository gate was run. No new helper was claimed.

## Landing on the shared harness integration base

At the user's explicit instruction, rebased onto origin/area/stage1-lint d65a8f931c98655936ae04c6899f38f14862b73e, which contains harness 41eb6eab2 and main 39638d9e. Rebase completed without conflicts, skipping five patches already applied upstream. All incoming shared changes are retained; no shared source was edited. Pre-report head: da319ca069c52dbc24ce6150e9ad49c75e265572. Publication is solely to codex/lint-helpers-01 with the exact prior-worker SHA lease.

Setup completed in 96s; Go/clang/Node/submodules were ready at 1s and cache warming completed at 96s. nproc=5. The full helper oracle command `go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m` passed in 608.885s: 76 top-level passes, eight explicit ownership-blocked skips, 77 compiled semantic mutant catches, two expected refusal checks and eight source-only tree mutant catches. `go vet ./stage1/cohere/lint/helpers` passed. `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m` passed in 6.512s. Logs: evidence/slot01-wave11/area-{oracle,vet,input,setup}.log.

The same CollapseCopyNode[] adamic/cycle-capable ownership refusal remains after integrating the shared harness. No tree native/emitted-JavaScript parity or compiled tree mutant is credited. No further helper is claimed, no readiness credit is added, and no full repository gate was run.
