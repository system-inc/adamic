Built isGenericName, hasMathFunction and LoadedDesignSystem.Utilities in separate .a files: twelve dependency edges across four rules.
Commits: claim 413ad5a published before source; current main f8013f0; implementation 8b8d1de696776fac87ff7dd6554b514b01a67469 pushed only to codex/lint-helpers-03.
Commands: focused gate PASS 9.291s, 1,118 comparison lines; complete owned gate PASS 358.002s, 2,989,186 lines; inherited helpers PASS 71.629s; vet/format clean; six uncached input probes PASS 1.546s.
Mutants: seven new compiling variants caught; all seventy-one owned variants, four inherited variants and missing-consumer mutant caught.
Not covered: full repository gate, whole-rule findings/fixes/suggestions integration, invalid Unicode adapters or unavailable external live Tailwind/corpora.

## Landing and ownership

Only codex/lint-helpers-03 is pushed by this worker. At selection, fetched origin/main remains f8013f0baac41ddc340d76f83bddde38536a8f07, already contained in published tip 9033ef8. All thirty previous helpers are complete and oracle-green on that base (batch10/evidence/landing-final.log). No rebase is needed before these claims. Every claims file on all eighteen fetched origin/codex/lint-helpers* branches is inspected. The comments bundle is separately reserved in shared HELPERS.md and rule-local strict option decoders are not additional named helpers. These three tie the highest available concrete count at four each. Claim 413ad5a is pushed before source; a refresh confirms only slot 03 claims them. See evidence/claims.log.

## Behavior and independent observations

isGenericName reproduces the exact thirteen-keyword Go map. It does not trim, fold case or normalize Unicode. hasMathFunction reproduces the exact nineteen math-name substrings followed immediately by (, searched anywhere in the string. It accepts malformed or quoted input and prefixes before the function name. The Go preliminary check for any opening parenthesis is redundant with the exact substring search, so the leaf helper expresses the complete predicate directly. Full source strings are additional controls, not evidence of executing the whole utility pipeline.

LoadedDesignSystem.Utilities returns the identical nullable evaluator pointer. The initialized system is represented by an owned {utility: number} record and existing evaluator records by stable arena indices, with -1 for nil. The test adapter supplies nil and two distinct pointers, proves aliasing by writing a theme marker through the returned object, checks the original stored pointer, and repeats the getter. This helper does not instantiate or evaluate utilities. A null receiver and invalid arena handles are outside the initialized-system prerequisite, not silently accepted equivalents.

The owned temporary Go overlay exports the real private predicates and calls the actual getter. Expected output is not obtained from copied Go implementations. Source Node, emitted JavaScript on Node and native under ASan/UBSan and Linux leak checking compare byte for byte. Capture covers all four consumer rules and 118 runtime sources before external engine skips. Full sources and extracted tokens, all thirteen accepted generic-family names, every nineteen math-function spelling and uppercase, whitespace, NUL, Unicode, prefix and quotation variations yield 556 strings. Two predicates per string plus six pointer/state observations give 1,118 lines.

Text inputs are valid UTF-8-derived Unicode strings. Invalid UTF-8 bytes and unpaired UTF-16 are outside this adapter boundary. No new rules, rule dispatch, rule.json, shared registration generator, shared test harness, compiler or Diagnostic model changes.

The first focused run fails because the newly written Go export adapter guessed a lowercase private theme field on UtilityEvaluator. The actual field is exported Theme. Corrected only the owned adapter; evidence/focused.log retains the failure and focused-fixed.log records the passing rerun. No compiler or production helper fix was needed for this mistake.

Capture's actual upstream Tailwind package exits 1 on the known external installed-engine/corpus guards. The unavailable /Users/kirkouimet/Projects/ahra/app/_theme/styles and external class corpora block the whole-rule live gate. Every selected consumer source is still captured before its skip. This is explicitly not a passing upstream rule package or complete native finding/fix/suggestion comparison.

## Consumers and readiness

Each of collapse.isGenericName, collapse.hasMathFunction and collapse.*LoadedDesignSystem.Utilities removes its own dependency from every rule below:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Twelve dependency edges across four distinct rules, zero final blockers removed by this batch alone. readiness.json subtracts only this worker's thirty-three delivered helpers, not other workers' claims or unintegrated implementations. The frozen ledger's common AST adapter assumption still applies; no rule is marked implemented.

## Commands and environment

All test output goes directly to logs. Source /workspace/adamic-tools/env.sh before Go builds:

- bash cloud/setup.sh > stage1/cohere/lint/helpers/slot03/batch11/evidence/setup.log 2>&1
- python3 stage1/cohere/lint/helpers/slot03/batch11/testdata/regenerate.py > stage1/cohere/lint/helpers/slot03/batch11/evidence/regenerate.log 2>&1
- go test ./stage1/cohere/lint/helpers/slot03 -run 'TestBatch11' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch11/evidence/focused-fixed.log 2>&1
- go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch11/evidence/final.log 2>&1
- go vet ./... > stage1/cohere/lint/helpers/slot03/batch11/evidence/vet.log 2>&1
- gofmt -l cmd internal stage1/cohere/lint/helpers/slot03 > stage1/cohere/lint/helpers/slot03/batch11/evidence/gofmt.log
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch11/evidence/oracle.log 2>&1

Setup: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 115s, done 115s on 5 processors; nproc=5, cgroup cpu.max 400000 100000, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup warms compilation, not tests. Cohere pin 715ba94f3608a6500086b1076ce5cb7e51b836db.

## New compiling semantic mutants

Each temporary variant must compile, exit 0 and emit no stderr before comparison with actual Go. Refusals, crashes, compiler warnings and sanitizer failures never count as semantic catches.

| Change | First differing line | Actual Go witness |
|---|---:|---|
| Drop fangsong | 549 | true versus mutant false |
| Fold generic-family case | 101 | false versus mutant true |
| Anchor calc( at the start | 38 | true versus mutant false |
| Drop round( | 60 | true versus mutant false |
| Fold math-function case | 88 | false versus mutant true |
| Return nil for every utility getter | 1115 | 0/true/true versus -1/false/true |
| Clear the stored handle while returning its old value | 1115 | 0/true/true versus 0/true/false |

## Complete regression evidence

Complete owned package PASS 358.002s on f8013f0: thirty-three helpers and 2,989,186 comparison lines. All seventy-one owned compiling semantic variants and the missing-consumer mutant are caught. Batches three through eleven additionally compare emitted JavaScript; the first two original batches compare source Node and sanitized native. Inherited shared helpers PASS 71.629s, 22,412 output lines, four separate semantic mutants and ten message refusals. Vet and formatting logs are empty; uncached input oracle PASS 1.546s, zero hits and six probe misses. Capture reproduces byte for byte; hashes are in evidence/reproducibility.log. Final eighteen-branch ownership and current-main check in evidence/claims-final.log.

Every observed semantic witness follows, from the final combined gate. The helper-to-change definitions live in the owned mutant tables and previous batch reports; stdout disagreement with actual Go catches each after successful compilation and execution.

- TestHelperMutants/options_json.ts; at output line 15157: got "valid", Go "invalid"
- TestHelperMutants/option_schema.ts; at output line 15166: got "valid", Go "invalid"
- TestHelperMutants/policy_message.ts; at output line 22166: got "This throws a bare `{{constructor}}`, which names no declared failure. Raise it through the tier that declares it, `AccountModule.error(identifier, data, cause)`, `ApiWorker.error(...)` or `Base.error(...)`. A bare throw carries no identifier, so the board groups it by its message and one interpolated value mints one identity per value, and it normalizes to 500, so a refusal reads as our fault.", Go "This throws a bare `sentinel é😀`, which names no declared failure. Raise it through the tier that declares it, `AccountModule.error(identifier, data, cause)`, `ApiWorker.error(...)` or `Base.error(...)`. A bare throw carries no identifier, so the board groups it by its message and one interpolated value mints one identity per value, and it normalizes to 500, so a refusal reads as our fault."
- TestHelperMutants/strict_options.ts; at output line 15178: got "valid", Go "invalid"
- TestBatch10Mutants/is_url.a; at line 79: got "true" Go "false"
- TestBatch10Mutants/is_url.a#01; at line 916: got "true" Go "false"
- TestBatch10Mutants/is_absolute_size.a; at line 1046: got "false" Go "true"
- TestBatch10Mutants/is_absolute_size.a#01; at line 65: got "true" Go "false"
- TestBatch10Mutants/is_relative_size.a; at line 783: got "false" Go "true"
- TestBatch11Mutants/is_generic_name.a; at line 549: got "false" Go "true"
- TestBatch11Mutants/is_generic_name.a#01; at line 101: got "true" Go "false"
- TestBatch11Mutants/has_math_function.a; at line 38: got "false" Go "true"
- TestBatch11Mutants/has_math_function.a#01; at line 60: got "false" Go "true"
- TestBatch11Mutants/has_math_function.a#02; at line 88: got "true" Go "false"
- TestBatch11Mutants/loaded_utilities.a; at line 1115: got "-1/false/true" Go "0/true/true"
- TestBatch11Mutants/loaded_utilities.a#01; at line 1115: got "0/true/false" Go "0/true/true"
- TestBatch2Mutants/intrinsic_element_named.a; at line 24: got "true" Go "false"
- TestBatch2Mutants/hole_edges.a; at line 179230: got "false,false" Go "true,false"
- TestBatch2Mutants/read_class_values.a; at line 185959: got "Attribute::" Go "Callee::"
- TestBatch3Mutants/hex_value.a; at line 43: got "-1" Go "15"
- TestBatch3Mutants/unescape_string_literal_text.a; at line 1114212: got "0,99,111,112,121,10,:99,111,112,121," Go "0,169,10,:99,111,112,121,"
- TestBatch3Mutants/parameter_nodes.a; at line 1115385: got "2,-1,3,2" Go "-1,-1,3,2"
- TestBatch3Mutants/parameter_nodes.a#01; at line 1115375: got "0,1" Go ""
- TestBatch4Mutants/escape_terminator.a; at line 13: got "false" Go "true"
- TestBatch4Mutants/followed_by_whitespace.a; at line 3585: got "true:10" Go "false:10,11"
- TestBatch4Mutants/ignored_theme_key.a; at line 75243: got "true" Go "false"
- TestBatch5Mutants/split_theme_key.a; at line 283: got "1:" Go "0:"
- TestBatch5Mutants/split_theme_key.a#01; at line 286: got "0:" Go "1:"
- TestBatch5Mutants/split_theme_key.a#02; at line 288: got "1:99,104,97,110,103,101,100," Go "1:"
- TestBatch5Mutants/join_segments.a; at line 290: got "0," Go "0,45,"
- TestBatch5Mutants/breakpoint_group_order.a; at line 9639: got "2:true" Go "0:false"
- TestBatch5Mutants/breakpoint_group_order.a#01; at line 9641: got "23:true" Go "-17:true"
- TestBatch6Mutants/at_rule.a; at line 1: got "rule|||112,97,114,97,109,115,58,|||false|false|false|false|0|0|" Go "at-rule|||112,97,114,97,109,115,58,|||false|false|false|false|0|0|"
- TestBatch6Mutants/at_rule.a#01; at line 14: got "1:1:true:0" Go "1:1:true:-1"
- TestBatch6Mutants/at_rule.a#02; at line 14: got "0:0:true:" Go "1:1:true:-1"
- TestBatch6Mutants/at_rule.a#03; at line 3: got "at-rule|99,104,97,110,103,101,100,|99,104,97,110,103,101,100,|112,97,114,97,109,115,58,|||false|false|false|false|0|0|" Go "at-rule|||112,97,114,97,109,115,58,|||false|false|false|false|0|0|"
- TestBatch6Mutants/style_rule.a; at line 4: got "at-rule||||||false|false|false|false|0|0|" Go "rule||||||false|false|false|false|0|0|"
- TestBatch6Mutants/style_rule.a#01; at line 17: got "1:1:true:0" Go "1:1:true:-1"
- TestBatch6Mutants/style_rule.a#02; at line 17: got "0:0:true:" Go "1:1:true:-1"
- TestBatch6Mutants/style_rule.a#03; at line 6: got "rule|99,104,97,110,103,101,100,|99,104,97,110,103,101,100,||||false|false|false|false|0|0|" Go "rule||||||false|false|false|false|0|0|"
- TestBatch6Mutants/variant_next_order.a; at line 29488: got "-9007199254740989" Go "-9007199254740991"
- TestBatch6Mutants/variant_next_order.a#01; at line 29485: got "-9007199254740988" Go "-9007199254740989"
- TestBatch6Mutants/variant_next_order.a#02; at line 29486: got "false:0:-9007199254740989" Go "false:0:-9007199254740990"
- TestBatch7Mutants/new_variant_registry.a; at line 1: got "1:false:0:0:0" Go "0:false:0:0:0"
- TestBatch7Mutants/new_variant_registry.a#01; at line 25: got "84:false:0:2:2" Go "0:false:0:0:0"
- TestBatch7Mutants/register.a; at line 8: got ":84:replacement" Go ":83:replacement"
- TestBatch7Mutants/register.a#01; at line 680: got ":83:static" Go ":-17:static"
- TestBatch7Mutants/register.a#02; at line 679: got "-17:true:-17:1:0" Go "82:true:-17:1:0"
- TestBatch7Mutants/attach_comparison.a; at line 16: got "84:false:0:2:0" Go "84:false:0:2:1"
- TestBatch7Mutants/attach_comparison.a#01; at line 21: got "-8" Go "15"
- TestBatch7Mutants/attach_comparison.a#02; at line 15: got "missing" Go "-8"
- TestBatch8Mutants/recursively_decode_arbitrary_values.a; at line 413: got "function:102,110,|function:118,97,114,|word:45,45,97,95,98,|separator:44,|word:32,99,32,100,|separator:44,|function:117,114,108,|word:97,32,98,|" Go "function:102,110,|function:118,97,114,|word:45,45,97,95,98,|separator:44,|word:32,99,32,100,|separator:44,|function:117,114,108,|word:97,95,98,|"
- TestBatch8Mutants/recursively_decode_arbitrary_values.a#01; at line 71: got "function:99,97,108,99,|function:118,97,114,|word:45,45,97,32,98,|word:43,49,112,120,|" Go "function:99,97,108,99,|function:118,97,114,|word:45,45,97,95,98,|word:43,49,112,120,|"
- TestBatch8Mutants/recursively_decode_arbitrary_values.a#02; at line 714: got "unknown:97,32,117,114,108,|word:45,45,97,95,98,|word:99,92,95,100,95,101,|word:95,102,95,103,|function:102,111,111,95,98,97,114,|word:104,95,105,|" Go "unknown:97,95,117,114,108,|word:45,45,97,95,98,|word:99,92,95,100,95,101,|word:95,102,95,103,|function:102,111,111,95,98,97,114,|word:104,95,105,|"
- TestBatch8Mutants/decode_arbitrary_value.a; at line 70: got "99,97,108,99,40,49,112,120,43,50,112,120,41," Go "99,97,108,99,40,49,112,120,32,43,32,50,112,120,41,"
- TestBatch8Mutants/decode_arbitrary_value.a#01; at line 70: got "99,97,108,99,40,49,112,120,43,50,112,120,41," Go "99,97,108,99,40,49,112,120,32,43,32,50,112,120,41,"
- TestBatch8Mutants/decode_arbitrary_value.a#02; at line 30: got "117,110,101,120,112,101,99,116,101,100,32,109,97,116,104,32,105,110,112,117,116,58,117,110,101,120,112,101,99,116,101,100,32,112,97,114,115,101,32,105,110,112,117,116,58,85,82,76,40,97,32,98,41,32," Go "85,82,76,40,97,32,98,41,"
- TestBatch8Mutants/register_theme_breakpoint_variants.a; at line 851: got "49,48,48,:-17:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|" Go "101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
- TestBatch8Mutants/register_theme_breakpoint_variants.a#01; at line 855: got "49,48,48,:-17:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|" Go "101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
- TestBatch8Mutants/register_theme_breakpoint_variants.a#02; at line 747: got "101,120,105,115,116,105,110,103,:1:static|115,109,:2:compound|" Go "101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
- TestBatch8Mutants/register_theme_breakpoint_variants.a#03; at line 859: got "49,48,48,:-16:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|" Go "49,48,48,:-17:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
- TestBatch9Mutants/parse_css.a; at line 31: got "ok:[rule:46,97,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]" Go "ok:[comment:::::33,32,108,105,99,101,110,115,101,32,:false:false:false:[]|rule:46,97,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]"
- TestBatch9Mutants/parse_css.a#01; at line 418: got "ok:[rule:117,110,101,120,112,101,99,116,101,100,32,116,114,105,109,58,65279,46,97,32,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]" Go "ok:[rule:46,97,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]"
- TestBatch9Mutants/parse_css.a#02; at line 7: got "ok:[]" Go "ok:[declaration::::45,45,120,::false:true:false:[]|]"
- TestBatch9Mutants/parse_css.a#03; at line 21: got "ok:[rule:46,97,:::::false:false:true:[unexpected declaration:::::98,97,99,107,103,114,111,117,110,100,58,32,117,114,108,40,97,:false:false:false:[]|unexpected declaration:::::98,46,112,110,103,41,:false:false:false:[]|]|]" Go "ok:[rule:46,97,:::::false:false:true:[declaration::::98,97,99,107,103,114,111,117,110,100,:117,114,108,40,97,59,98,46,112,110,103,41,:false:true:false:[]|]|]"
- TestBatch9Mutants/parse_css.a#04; at line 27: got "ok:[]" Go "error:5:77,105,115,115,105,110,103,32,99,108,111,115,105,110,103,32,125,32,97,116,32,46,97,"
- TestBatch9Mutants/parse_css.a#05; at line 5: got "error:1:77,105,115,115,105,110,103,32,111,112,101,110,105,110,103,32,40," Go "error:0:77,105,115,115,105,110,103,32,111,112,101,110,105,110,103,32,40,"
- TestBatch9Mutants/design_system_decline_message.a; at line 433: got "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32,60,110,105,108,62," Go "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,32,102,114,111,109,32,116,104,101,109,101,46,99,115,115,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32,60,110,105,108,62,"
- TestBatch9Mutants/design_system_decline_message.a#01; at line 421: got "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32," Go "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32,60,110,105,108,62,"
- TestBatch9Mutants/loaded_theme.a; at line 734: got "-1:false:true" Go "0:true:true"
- TestBatch9Mutants/loaded_theme.a#01; at line 734: got "0:true:false" Go "0:true:true"
- TestSlot03HelperMutants/component_base_name.a/value; at line 159: got "false" Go "true"
- TestSlot03HelperMutants/tailwind_space.a/value; at line 847: got "false" Go "true"
- TestSlot03HelperMutants/listener_kinds.a/value; at line 1114948: got "JsxAttribute,CallExpression,StringLiteral" Go "JsxAttribute,CallExpression,VariableDeclaration"
- TestSlot03HelperMutants/listener_kinds.a/shared-list; at line 1114949: got "StringLiteral,CallExpression,VariableDeclaration" Go "JsxAttribute,CallExpression,VariableDeclaration"

The coverage mutant removes no-unknown-classes; coverageVerdict reports that exact missing consumer. No silent acceptance is credited. The initial failed Go adapter run is retained separately and is not counted as a semantic mutant. Only the owned branch is published.
