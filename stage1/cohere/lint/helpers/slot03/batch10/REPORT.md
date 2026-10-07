Built isURL, isAbsoluteSize and isRelativeSize in separate .a files: twelve dependency edges across four rules.
Commits: landing-ready prior tip 760155c on main e8ba3d5; URL claim 4f5c75f and replacement claim 5ae67b6 pushed before source; implementation follows the gate.
Commands: final focused four-way gate PASS 7.934s and 1,071 lines; final thirty-helper regression PASS 346.152s, 2,988,068 lines; vet/format clean; six uncached input fixtures PASS 1.473s; capture byte-identical.
Mutants: five new compiling semantic variants caught; all sixty-four owned variants and the missing-consumer mutant caught in the final regression.
Not covered: full repository gate, whole-rule integration, arbitrary malformed UTF-8/UTF-16 adapters and unavailable external live Tailwind/corpora.

## Landing and claim race

Only codex/lint-helpers-03 is pushed by this worker. Before selecting new helpers, all twenty-seven existing helpers were ported, green and pushed at 760155c, with fetched origin/main e8ba3d5 contained in the branch. No rebase was necessary. All eighteen origin codex/lint-helpers* branches and every claim were inspected; the original comments bundle remains reserved by shared HELPERS.md. Every higher-count concrete symbol was reserved; these helpers tie the highest available count at four each.

Initial claim 4f5c75f at 04:07:41 UTC reserved URL, digit and prefix predicates. A post-build all-branch refresh found slot 04 bd21674 at 04:07:40 UTC had reserved digit/prefix one second earlier. Those two implementations are withdrawn, not delivered and not counted. Slot 03 retains URL. Replacement claim 5ae67b6 reserved the exact absolute/relative-size predicates after another complete scan, and was pushed before their source. The subsequent complete scan in evidence/claims.log confirms the retained three symbols appear only on slot 03. No fourth helper is reserved.

## Behavior and external observations

isURL implements actual Go's lowercase prefix, exact final parenthesis, nonoverlap/minimum width and four excluded interior line terminators. It does not parse URI syntax or require balanced inner parentheses. Unicode, NUL, tabs, vertical tabs, form feeds, NEL, NBSP and BOM remain valid interior text. A terminator after the closing parenthesis is rejected by Go's suffix check, even where a JavaScript regexp end anchor might permit it; this port follows Go behavior, not an inference from the source comment's regexp.

The size helpers preserve the complete eight-keyword absolute-size set and two-keyword relative-size set, exact case-sensitive equality without trimming or normalization. They are pure leaf helpers with no dependency callbacks. Valid UTF-8-derived strings form the adapter contract; arbitrary invalid UTF-8 bytes and unpaired UTF-16 are outside it. No new rules, listener routing, rule.json, shared registration generator, test harness, Diagnostic model or compiler files change.

The owned Go export overlay calls the real pinned private functions. No Go implementation is copied into the expected side. All four rule corpora supply 118 captured runtime sources, which are additional predicate controls alongside derived tokens and synthetic boundaries. Every positive size keyword is present, together with uppercase, surrounding whitespace, newline/NUL and affix variants. URL controls include wrong case/prefix/suffix, empty body, punctuation, Unicode and every excluded line terminator plus accepted whitespace controls. The final corpus has 357 strings and 1,071 outcomes, compared against Go on source Node, emitted JavaScript and ASan/UBSan native.

Five new mutants compile and run with exit 0 and no stderr before their wrong semantic output is compared with Go. URL folding admits URL(x); removing U+2028 exclusion admits that terminator; removing xxx-large drops an accepted keyword; folding absolute sizes admits uppercase; dropping smaller loses the second relative keyword. Temporary variants never mutate production source. Native compiler failures, stderr, crashes and sanitizer findings never count as catches.

## Evidence and superseded runs

The initial unpublished digit/prefix focused run passed, but initial-unpublished-focused.log and initial-unpublished-regenerate.log prove only local comparison work on the withdrawn duplicates. They do not count as delivered helpers or owned mutant witnesses.

The first package regression had already compiled the original test table before the one-second collision was discovered. evidence/final.log is that superseded run, containing the withdrawn helper observations. The package is rerun with the final test table in evidence/final-current.log. Only the final uniquely owned set is credited in this report. evidence/replacements.log passes all three retained baselines and five semantic mutants in 7.934s.

Capture's upstream rule package still exits 1 on known installed-Tailwind and external corpus guards. The unavailable /Users/kirkouimet/Projects/ahra/app/_theme/styles and external class corpora prevent that live gate. Runtime inputs are captured before skips; every selected consumer is observed. This is not a passing upstream whole-rule package or complete native findings/fixes/suggestions comparison.

## Consumers and readiness

collapse.isURL removes one dependency from each of:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

collapse.isAbsoluteSize removes one dependency from each of:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

collapse.isRelativeSize removes one dependency from each of:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Twelve edges across four unique rules; zero final blockers removed alone. readiness.json subtracts only this worker's thirty delivered helpers, not claims or implementations on other branches. The inventory's common AST adapter assumption still applies and no rule is marked implemented.

## Commands and environment

Source /workspace/adamic-tools/env.sh; all test output goes directly to logs:

- bash cloud/setup.sh: evidence/setup.log.
- python3 stage1/cohere/lint/helpers/slot03/batch10/testdata/regenerate.py: evidence/regenerate.log, capture.log and repeated regenerate-repeat.log; reproducibility.log records identical hashes.
- go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch10' -count=1 -v -timeout=20m: evidence/replacements.log.
- go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m: evidence/final-current.log, final table.
- go vet ./...: evidence/vet.log.
- gofmt -l cmd internal stage1/cohere/lint/helpers/slot03: evidence/gofmt.log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m: evidence/oracle.log.

```
go version go1.27.1 linux/amd64
setup: go ready (0s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
v24.19.0
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (36s)
setup: done in 36s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Cohere pin 715ba94f3608a6500086b1076ce5cb7e51b836db; nproc=5, cgroup quota 400000 100000. Setup warms build caches without executing tests. The full repository gate, rule integration, malformed adapters and external live corpora are not covered.

## Every final mutant witness on e8ba3d5

- TestBatch10Mutants/is_url.a; at line 79: got "true" Go "false"
- TestBatch10Mutants/is_url.a#01; at line 916: got "true" Go "false"
- TestBatch10Mutants/is_absolute_size.a; at line 1046: got "false" Go "true"
- TestBatch10Mutants/is_absolute_size.a#01; at line 65: got "true" Go "false"
- TestBatch10Mutants/is_relative_size.a; at line 783: got "false" Go "true"
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

Main advanced to f8013f0 during the gate. The implementation is committed before rebasing; current-main landing validation is recorded separately after the rebase. No main or area branch is pushed.
