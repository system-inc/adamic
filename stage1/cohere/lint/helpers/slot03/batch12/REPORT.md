Built isAngle, isNumber and isPercentage in separate .a files: twelve dependency edges across four rules.
Commits: rebased implementation c4193801d1abc3347984519593147654d471775c on main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06; original claim bc15542 was pushed before source.
Commands: current-main owned gate PASS 362.530s, 2,991,256 verdict/trace lines; shared helpers PASS 69.690s; forwarding witnesses PASS 3.518s; vet/format clean; six uncached probes PASS 1.505s.
Mutants: eleven new; seventy-nine owned variants in the complete gate plus three focused forwarding variants, all caught; four inherited variants and missing-consumer mutant also caught.
Not covered: full repository gate, dependency matcher implementation/wiring, whole-rule findings/fixes/suggestions or unavailable external live Tailwind/corpora.

## Landing and claim

Only codex/lint-helpers-03 is pushed by this worker. All thirty-three earlier helpers are ported, tested and pushed at eda4850 on f8013f0baac41ddc340d76f83bddde38536a8f07. Fetch confirms main remains this base, already an ancestor of the published branch, so no rebase is needed before claims. All twenty origin/codex/lint-helpers* branches and every claims file are inspected. Comments remain reserved by shared HELPERS.md; strict custom options is rule-local decoding. These three tie the highest available concrete symbol count at four consumers each. Claim bc15542 is pushed before any source; refresh confirms unique ownership. No fourth helper claimed.

## Behavior and independent observations

isAngle delegates unchanged input and the exact ordered angle units deg, rad, grad, turn to its number-with-suffix dependency. It returns the dependency result, without a math fallback. isNumber scans once, accepts only a positive whole-input numeric spelling, otherwise invokes math detection. isPercentage passes the unchanged input and exact one-element percent suffix list to number-with-suffix; a true result skips math detection. No new Go regex is ported and no hand-rolled matcher is written. These are predicate compositions; the numeric scanner, suffix matcher and math detector are explicit dependencies still owned separately. Callers must wire their real implementations before whole-rule integration.

The scanner's Go contract consumes only an ASCII numeric prefix. Its count therefore equals the prefix's UTF-16 length. A value wholly consumed is ASCII; any non-ASCII remainder makes the whole-input comparison false in both Go bytes and JavaScript code units. The helper compares the consumed count with JavaScript value.length under that prerequisite. It does not convert or build finding offsets. Text inputs are valid UTF-8-derived Unicode; invalid UTF-8 and unpaired UTF-16, malformed callback implementations and unrelated parser adapters are outside this contract.

The owned temporary Go overlay retains actual predicate bodies unchanged. Only dependency definition names are changed to allow tracing wrappers; those wrappers execute the original Go implementations. Top-level call traces record scanner, exact suffix list and math calls. Nested scanning inside number-with-suffix is suppressed so the observation matches the explicit dependency boundary. Case generation invokes the real scanner, suffix matcher and math detector and serializes their observations. The Adamic test-only adapters validate input identity and suffix requests, then return those observed values. Native runs therefore prove these compositions and lazy dependency calls, not native implementations of the separate matchers.

All four consuming rules supply 118 captured runtime sources. Whole sources, derived tokens, numeric controls including empty input/signs/dots/exponents, every angle unit, percent and length distractors, every math-function name, uppercase, whitespace, NUL and Unicode tails produce 690 distinct strings. Three verdict/call traces per string give 2,070 lines compared byte for byte with actual Go on source Node, emitted JavaScript and ASan/UBSan native with Linux leak checking. No new rule, rule.json, listener routing, shared harness/registration, compiler or Diagnostic model changes.

The first focused native attempt refuses a possible ownership cycle in the test adapter's captured function-value chain. Replacing those closures with file-level functions and explicit adapter state resolves it inside this owned directory. evidence/focused.log retains the failed refusal; focused-fixed.log is the passing final adapter. Refusals are not credited as mutant catches. A later regeneration retry initially fails with `Go: Unknown option: test`, because that shell omitted source /workspace/adamic-tools/env.sh and invoked an unrelated go executable. capture-unsourced.log and regenerate-repeat.log retain that failure. The sourced retry passes and source/coverage hashes remain identical, recorded in reproducibility.log.

Capture's actual Go Tailwind package still exits 1 on the known installed-engine/external-corpus guards. The unavailable /Users/kirkouimet/Projects/ahra/app/_theme/styles and external class corpora block the live whole-rule gate. Every selected rule's runtime input is nevertheless captured before its skip. This is not a passing upstream rule package or a native whole-rule finding/fix/suggestion comparison.

## Consumers and readiness

Each of collapse.isAngle, collapse.isNumber and collapse.isPercentage removes its own dependency from each rule below:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Twelve dependency edges across four distinct rules, zero final blockers removed alone. readiness.json subtracts only this worker's thirty-six delivered helpers, not unintegrated work on other branches. Dependency adapters and inventory's common AST assumptions remain prerequisites. No rule is marked implemented.

## Commands and environment

All tests write output directly to logs. Source /workspace/adamic-tools/env.sh before builds and captures:

- bash cloud/setup.sh > stage1/cohere/lint/helpers/slot03/batch12/evidence/setup.log 2>&1
- python3 stage1/cohere/lint/helpers/slot03/batch12/testdata/regenerate.py > stage1/cohere/lint/helpers/slot03/batch12/evidence/regenerate.log 2>&1
- go test ./stage1/cohere/lint/helpers/slot03 -run 'TestBatch12' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch12/evidence/focused-fixed.log 2>&1
- go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch12/evidence/final.log 2>&1
- go vet ./... > stage1/cohere/lint/helpers/slot03/batch12/evidence/vet.log 2>&1
- gofmt -l cmd internal stage1/cohere/lint/helpers/slot03 > stage1/cohere/lint/helpers/slot03/batch12/evidence/gofmt.log
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch12/evidence/oracle.log 2>&1

Setup: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 39s, done 39s on 5 processors; nproc=5, cgroup cpu.max 400000 100000, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup warms compilation, not tests. Cohere pin 715ba94f3608a6500086b1076ce5cb7e51b836db.

## New compiling semantic mutants

Every temporary variant must compile, exit 0 and write no stderr before comparison. The comparator checks verdicts and actual-Go top-level call traces. Some first witnesses are changed call arguments or order before a later changed boolean verdict; these are dependency contract observations, not matcher implementation evidence.

| Change | First differing line | Actual Go witness |
|---|---:|---|
| Drop turn from angle suffix request | 1 | suffix:deg,rad,grad,turn; versus deg,rad,grad; |
| Negate angle result | 1 | false versus true |
| Allow zero-length number scan | 2 | false/scan;math; versus true/scan; |
| Allow partial number scan | 92 | false/scan;math; versus true/scan; |
| Eager math before number scan | 2 | scan;math; versus math;scan;math; |
| Replace percent suffix by percent text | 3 | suffix:%;math; versus suffix:percent;math; |
| Replace percentage OR by AND | 3 | suffix:%;math; versus suffix:%; |
| Eager math before percentage suffix | 3 | suffix:%;math; versus math;suffix:%;math; |

## Complete regression evidence

Pre-rebase owned gate PASS 365.301s on f8013f0: thirty-six helpers, 2,991,256 comparison lines and all seventy-nine compiling variants plus missing-consumer mutant caught. Inherited shared helpers PASS 80.484s, four variants and ten refusal cases. Batches three through twelve additionally compare emitted JavaScript; the first two original batches compare source Node and sanitized native. Vet/format clean; six uncached input probes PASS 1.841s. Sources and coverage reproduce byte for byte.

Main advanced to c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 during the gate. The completed implementation is committed before rebasing; current-main landing validation follows. The f8013f0 results are historical, not a current-main pass. Only the own branch is published.

Every pre-rebase semantic witness follows; the comparator catches actual-Go verdict, argument or call-trace disagreement after successful compilation and execution:

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
- TestBatch12Mutants/is_angle.a; at line 1: got "false/suffix:deg,rad,grad;" Go "false/suffix:deg,rad,grad,turn;"
- TestBatch12Mutants/is_angle.a#01; at line 1: got "true/suffix:deg,rad,grad,turn;" Go "false/suffix:deg,rad,grad,turn;"
- TestBatch12Mutants/is_number.a; at line 2: got "true/scan;" Go "false/scan;math;"
- TestBatch12Mutants/is_number.a#01; at line 92: got "true/scan;" Go "false/scan;math;"
- TestBatch12Mutants/is_number.a#02; at line 2: got "false/math;scan;math;" Go "false/scan;math;"
- TestBatch12Mutants/is_percentage.a; at line 3: got "false/suffix:percent;math;" Go "false/suffix:%;math;"
- TestBatch12Mutants/is_percentage.a#01; at line 3: got "false/suffix:%;" Go "false/suffix:%;math;"
- TestBatch12Mutants/is_percentage.a#02; at line 3: got "false/math;suffix:%;math;" Go "false/suffix:%;math;"
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

The missing-consumer mutant removes no-unknown-classes; coverageVerdict names that missing consumer. Compiler refusals and the unsourced capture invocation are separate failed attempts, not mutant catches.

## Additional forwarding coverage on current main

Landing review found the test adapters' input-identity checks had no deliberate changed-argument witness. Three additional temporary variants append x to the angle suffix input, the number scanner input and the number math input. All compile and exit 0 without stderr; runtime wrong-value trace witnesses are caught at lines 1, 2 and 2 respectively. TestBatch12ArgumentMutants passes in 3.518s on c01907a, evidence/landing-arguments.log. This adds three tests without changing any helper source.

The combined landing package run was already compiled with seventy-nine owned variants when these were added. Its count remains seventy-nine; the three focused variants are credited separately, bringing the distinct owned total to eighty-two, eleven new in this batch. Vet and formatting are rerun after the test addition in landing-vet-current.log and landing-gofmt-current.log, both empty. No unrun full-suite count of eighty-two is claimed.

## Current-main landing

All forty-eight branch commits rebased cleanly onto c01907a. Implementation c4193801d1abc3347984519593147654d471775c preserves all helper source from the completed pre-rebase implementation; no compiler or shared harness change was needed. Main's advance adds Stage 3 files and internal/oracle/stage3_hook_test.go, with no helper or compiler implementation changes. The previously passing f8013f0 gate remains historical evidence.

Current-main commands, with /workspace/adamic-tools/env.sh sourced:

- go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch12/evidence/landing-final.log 2>&1
- go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch12ArgumentMutants$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch12/evidence/landing-arguments.log 2>&1
- go vet ./... > stage1/cohere/lint/helpers/slot03/batch12/evidence/landing-vet-current.log 2>&1
- gofmt -l cmd internal stage1/cohere/lint/helpers/slot03 > stage1/cohere/lint/helpers/slot03/batch12/evidence/landing-gofmt-current.log
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch12/evidence/landing-oracle.log 2>&1

Owned helpers PASS 362.530s: 2,991,256 comparison lines, all seventy-nine original semantic variants and missing-consumer mutant caught. Inherited shared helpers PASS 69.690s, four variants and ten refusals. Three newly added forwarding variants PASS 3.518s separately, giving eighty-two distinct owned variants. Helper source is unchanged by those test additions. Latest vet/format logs are empty; uncached input oracle PASS 1.505s, zero hits and six misses. No full Stage 3 fixture or full repository gate is claimed.

The final publication targets only codex/lint-helpers-03, with an exact force-with-lease against the last published own claim bc1554253adf444f292943b629e1021f4d1d52bf after the user-required rebase. No main or area branch is pushed. Every current-main semantic witness follows:

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
- TestBatch12Mutants/is_angle.a; at line 1: got "false/suffix:deg,rad,grad;" Go "false/suffix:deg,rad,grad,turn;"
- TestBatch12Mutants/is_angle.a#01; at line 1: got "true/suffix:deg,rad,grad,turn;" Go "false/suffix:deg,rad,grad,turn;"
- TestBatch12Mutants/is_number.a; at line 2: got "true/scan;" Go "false/scan;math;"
- TestBatch12Mutants/is_number.a#01; at line 92: got "true/scan;" Go "false/scan;math;"
- TestBatch12Mutants/is_number.a#02; at line 2: got "false/math;scan;math;" Go "false/scan;math;"
- TestBatch12Mutants/is_percentage.a; at line 3: got "false/suffix:percent;math;" Go "false/suffix:%;math;"
- TestBatch12Mutants/is_percentage.a#01; at line 3: got "false/suffix:%;" Go "false/suffix:%;math;"
- TestBatch12Mutants/is_percentage.a#02; at line 3: got "false/math;suffix:%;math;" Go "false/suffix:%;math;"
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
- TestBatch12ArgumentMutants/is_angle.a; at line 1: got "false/suffix:deg,rad,grad,turn;wrong-value;" Go "false/suffix:deg,rad,grad,turn;"
- TestBatch12ArgumentMutants/is_number.a; at line 2: got "false/scan;wrong-value;math;" Go "false/scan;math;"
- TestBatch12ArgumentMutants/is_number.a#01; at line 2: got "false/scan;math;wrong-value;" Go "false/scan;math;"
