Built InferDataType, matchesDataType and isFamilyName in separate .a files: twelve dependency edges across four rules.
Commits: claim 252cb85a pushed before source; base c01907a; implementation SHA recorded after publication.
Commands: focused gate PASS 48.698s, 34,896 verdict/trace lines; complete owned gate PASS 393.189s, 3,026,152 comparison lines; shared helpers PASS 70.058s; vet/format clean; six uncached probes PASS 1.549s.
Mutants: twenty-nine new compiling variants caught; all one hundred eleven owned variants, four inherited variants and missing-consumer mutant caught.
Not covered: full repository gate, separate matcher/segment dependency wiring, whole-rule findings/fixes/suggestions or unavailable live Tailwind/corpora.

## Landing and ownership

Only codex/lint-helpers-03 is pushed by this worker. All thirty-six earlier helpers are complete, tested on c01907a and pushed at 2c023e53 before selection. Fetched main remains c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 and is contained in the branch; no rebase is needed at selection. All twenty origin/codex/lint-helpers* branches and every claims file are inspected. The comments bundle remains reserved in shared HELPERS.md; custom option decoding remains rule-local. These concrete helpers tie the highest available count at four consumers each. Claim 252cb85a is published before source; a refresh confirms all three occur only on slot 03. No fourth helper is claimed.

Harness commit ab70f38d4 is fetched from origin/lint-rules/harness, but is not an ancestor of the fetched main at selection. This helper unit creates no rules and needs no report API or Diagnostic model changes. No shared harness, registration, compiler, listener routing or AST-kind dispatch is edited. Allocator-test changes are accepted through main whenever they arrive; none is reverted here.

## Behavior and independent observations

InferDataType uses the exact lowercase var( prefix guard and returns the first successful type in the caller's order, otherwise empty. Empty lists and unknown types remain valid inputs. Its matcher dependency is explicit and follows the pure upstream predicates; callbacks mutating the type list are outside that contract.

matchesDataType ports the actual seventeen-way CSS type dispatcher, including ratio to fraction and integer to Go's nonnegative IsPositiveInteger predicate. It returns false for unknown names without invoking a predicate. A check callback receives the unchanged string and one fixed predicate index from Go's AllDataTypes order. README.md lists every index, type and Go function. This is CSS-type dispatch, not AST listener dispatch inside a rule.

isFamilyName obtains comma parts from its explicit CSS segment dependency. Each leading ASCII digit rejects the whole value. Exact lowercase var( parts are skipped. Every other part counts, including empty strings; at least one must count. Thus empty string and leading/doubled/trailing commas can qualify, while a list entirely of variable references does not. No trimming is introduced. Quote, escape and bracket behavior is supplied by Go-observed segmentation, not a hand-written segmenter delivered here.

Fixed string-prefix and byte-range checks become JS RegExp literals /^var\(/u and /^[0-9]/u. Both source Node and emitted JavaScript actually execute these literals; native compilation and comparison pass. Shared regex table 071fb012848ce0408428c61aba0857cca472236f contains 107 sites and no data_type.go row. No Go regexp compile site is ported in this batch, no hand-rolled matcher is introduced, and no option regex or finding position is constructed. Invalid UTF-8 bytes and unpaired UTF-16 are outside the Unicode adapter boundary.

The owned temporary Go overlay preserves actual helper control flow and adds top-level observations at real dispatcher returns and the family segment call. Expected verdicts still come from actual Go helper bodies. Actual Go supplies all seventeen private predicate verdicts and comma segments. Test-only Adamic callbacks return those observations and record the exact route, type, separator and unchanged value. This proves the compositions and family logic, not native implementations of all predicate/segment dependencies. Whole-rule consumers must wire those separately owned dependencies.

All four consuming rules supply 118 captured runtime sources. Complete sources, derived tokens, numeric/unit/math controls and family prefix/comma/quote/escape/nesting controls give 727 strings. Each runs 21 known/unknown type selectors, 26 caller type orders and the family predicate: 48 observations per string, 34,896 verdict/call-trace lines. Actual Go, source Node, emitted JavaScript and ASan/UBSan native with Linux leak checking agree byte for byte.

Capture's Go Tailwind package still exits 1 on the known unavailable external engine/corpus guards. The /Users/kirkouimet/Projects/ahra/app/_theme/styles installation and external class corpora are absent. Inputs are captured before skips and every selected consumer is observed. This is explicitly not a passing whole-rule package or complete native finding/fix/suggestion comparison. Source and coverage capture reproduce byte for byte; hashes are in evidence/reproducibility.log.

## Consumers and readiness

Each of collapse.InferDataType, collapse.matchesDataType and collapse.isFamilyName removes one dependency from each rule below:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Twelve edges across four unique rules, zero final blockers removed alone. readiness.json subtracts only this worker's thirty-nine delivered helpers, not other workers' claims or unintegrated implementations. No rule is marked implemented; the frozen inventory's common AST assumptions still apply.

## Commands and environment

All test output goes directly to logs. Source /workspace/adamic-tools/env.sh first:

- bash cloud/setup.sh > stage1/cohere/lint/helpers/slot03/batch13/evidence/setup.log 2>&1
- python3 stage1/cohere/lint/helpers/slot03/batch13/testdata/regenerate.py > stage1/cohere/lint/helpers/slot03/batch13/evidence/regenerate.log 2>&1
- go test ./stage1/cohere/lint/helpers/slot03 -run 'TestBatch13' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch13/evidence/focused.log 2>&1
- go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch13/evidence/final.log 2>&1
- go vet ./... > stage1/cohere/lint/helpers/slot03/batch13/evidence/vet.log 2>&1
- gofmt -l cmd internal stage1/cohere/lint/helpers/slot03 > stage1/cohere/lint/helpers/slot03/batch13/evidence/gofmt.log
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch13/evidence/oracle.log 2>&1

Setup: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 51s, done 51s on 5 processors; nproc=5, cgroup cpu.max 400000 100000, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup warms compilation, not tests. Cohere pin 715ba94f3608a6500086b1076ce5cb7e51b836db.

## New compiling semantic mutants

Every temporary variant compiles, exits 0 and writes no stderr before comparison. Refusals, crashes and sanitizer failures do not count as semantic catches. Some first witnesses are wrong dependency requests whose boolean result happens to agree; the trace independently names the actual Go route.

Nineteen dispatcher variants rotate each of the seventeen individual arms to the next predicate, invoke a predicate for an unknown type, or change the forwarded value. Four inference variants remove the var guard, fold its case, reverse caller order or change forwarded input. Six family variants admit a leading zero, skip uppercase VAR instead of lowercase var, stop counting, accept zero counted parts, change separator or change forwarded input. The identity guards are deliberately mutated in this first batch gate, not left unproven.

## Complete regression evidence

Complete owned package PASS 393.189s on c01907a: thirty-nine helpers, 3,026,152 comparison lines and all one hundred eleven compiling semantic variants plus missing-consumer mutant caught. This includes every prior forwarding variant in the full gate. Batches three through thirteen also compare emitted JavaScript; the first two original batches compare source Node and sanitized native. Inherited shared helpers PASS 70.058s, four variants and ten refusals. Vet and formatting logs are empty. Six uncached input probes PASS 1.549s, zero hits and six misses. Source and coverage regeneration are byte-identical.

Final fetch retains main c01907a, already contained in the branch. Every claims file on all twenty helper branches is inspected again; the selected symbols remain uniquely reserved on slot 03. Evidence is in claims-final.log. Only the own branch is published.

Every observed semantic witness follows. Each is runtime disagreement with actual Go after a successful native compile/run, with no stderr:

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
- TestBatch12ArgumentMutants/is_angle.a; at line 1: got "false/suffix:deg,rad,grad,turn;wrong-value;" Go "false/suffix:deg,rad,grad,turn;"
- TestBatch12ArgumentMutants/is_number.a; at line 2: got "false/scan;wrong-value;math;" Go "false/scan;math;"
- TestBatch12ArgumentMutants/is_number.a#01; at line 2: got "false/scan;math;wrong-value;" Go "false/scan;math;"
- TestBatch13Mutants/matches_data_type.a; at line 1: got "false/check:1;" Go "false/check:0;"
- TestBatch13Mutants/matches_data_type.a#01; at line 2: got "false/check:2;" Go "false/check:1;"
- TestBatch13Mutants/matches_data_type.a#02; at line 3: got "false/check:3;" Go "false/check:2;"
- TestBatch13Mutants/matches_data_type.a#03; at line 4: got "false/check:4;" Go "false/check:3;"
- TestBatch13Mutants/matches_data_type.a#04; at line 5: got "false/check:5;" Go "false/check:4;"
- TestBatch13Mutants/matches_data_type.a#05; at line 6: got "false/check:6;" Go "false/check:5;"
- TestBatch13Mutants/matches_data_type.a#06; at line 7: got "false/check:7;" Go "false/check:6;"
- TestBatch13Mutants/matches_data_type.a#07; at line 8: got "false/check:8;" Go "false/check:7;"
- TestBatch13Mutants/matches_data_type.a#08; at line 9: got "false/check:9;" Go "false/check:8;"
- TestBatch13Mutants/matches_data_type.a#09; at line 10: got "false/check:10;" Go "false/check:9;"
- TestBatch13Mutants/matches_data_type.a#10; at line 11: got "true/check:11;" Go "false/check:10;"
- TestBatch13Mutants/matches_data_type.a#11; at line 12: got "false/check:12;" Go "true/check:11;"
- TestBatch13Mutants/matches_data_type.a#12; at line 13: got "false/check:13;" Go "false/check:12;"
- TestBatch13Mutants/matches_data_type.a#13; at line 14: got "false/check:14;" Go "false/check:13;"
- TestBatch13Mutants/matches_data_type.a#14; at line 15: got "false/check:15;" Go "false/check:14;"
- TestBatch13Mutants/matches_data_type.a#15; at line 16: got "false/check:16;" Go "false/check:15;"
- TestBatch13Mutants/matches_data_type.a#16; at line 17: got "false/check:0;" Go "false/check:16;"
- TestBatch13Mutants/matches_data_type.a#17; at line 18: got "false/check:0;" Go "false/"
- TestBatch13Mutants/matches_data_type.a#18; at line 1: got "false/check:0;wrong-value;" Go "false/check:0;"
- TestBatch13Mutants/infer_data_type.a; at line 29879: got "/type:color;check:0;type:length;check:1;type:percentage;check:2;type:ratio;check:3;type:number;check:4;type:integer;check:5;type:url;check:6;type:position;check:7;type:bg-size;check:8;type:line-width;check:9;type:image;check:10;type:family-name;check:11;type:generic-name;check:12;type:absolute-size;check:13;type:relative-size;check:14;type:angle;check:15;type:vector;check:16;" Go "/"
- TestBatch13Mutants/infer_data_type.a#01; at line 14279: got "/" Go "length/type:color;check:0;type:length;check:1;"
- TestBatch13Mutants/infer_data_type.a#02; at line 23: got "family-name/type:vector;check:16;type:angle;check:15;type:relative-size;check:14;type:absolute-size;check:13;type:generic-name;check:12;type:family-name;check:11;" Go "family-name/type:color;check:0;type:length;check:1;type:percentage;check:2;type:ratio;check:3;type:number;check:4;type:integer;check:5;type:url;check:6;type:position;check:7;type:bg-size;check:8;type:line-width;check:9;type:image;check:10;type:family-name;check:11;"
- TestBatch13Mutants/infer_data_type.a#03; at line 23: got "family-name/type:color;wrong-value;check:0;wrong-value;type:length;wrong-value;check:1;wrong-value;type:percentage;wrong-value;check:2;wrong-value;type:ratio;wrong-value;check:3;wrong-value;type:number;wrong-value;check:4;wrong-value;type:integer;wrong-value;check:5;wrong-value;type:url;wrong-value;check:6;wrong-value;type:position;wrong-value;check:7;wrong-value;type:bg-size;wrong-value;check:8;wrong-value;type:line-width;wrong-value;check:9;wrong-value;type:image;wrong-value;check:10;wrong-value;type:family-name;wrong-value;check:11;wrong-value;" Go "family-name/type:color;check:0;type:length;check:1;type:percentage;check:2;type:ratio;check:3;type:number;check:4;type:integer;check:5;type:url;check:6;type:position;check:7;type:bg-size;check:8;type:line-width;check:9;type:image;check:10;type:family-name;check:11;"
- TestBatch13Mutants/is_family_name.a; at line 5328: got "true/segment:,;" Go "false/segment:,;"
- TestBatch13Mutants/is_family_name.a#01; at line 14304: got "false/segment:,;" Go "true/segment:,;"
- TestBatch13Mutants/is_family_name.a#02; at line 48: got "false/segment:,;" Go "true/segment:,;"
- TestBatch13Mutants/is_family_name.a#03; at line 29904: got "true/segment:,;" Go "false/segment:,;"
- TestBatch13Mutants/is_family_name.a#04; at line 48: got "true/segment:;;" Go "true/segment:,;"
- TestBatch13Mutants/is_family_name.a#05; at line 48: got "true/segment:,;wrong-value;" Go "true/segment:,;"
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

The coverage mutant removes no-unknown-classes; coverageVerdict names that exact missing consumer. No full repository or whole-rule gate is claimed.
