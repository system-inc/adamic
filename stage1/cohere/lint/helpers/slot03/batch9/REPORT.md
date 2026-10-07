Built ParseCSS, DesignSystemDeclineMessage and LoadedDesignSystem.Theme in separate .a files: sixteen dependency edges across six rules.
Commits: prior landing-ready 8185ae1 on main e8ba3d5; claim 962fd49 pushed before source; implementation/evidence follow the final gate.
Commands: twenty-seven-helper gate PASS 314.916s, 2,986,997 lines; focused four-way gate PASS 31.279s; vet/format clean; six uncached input fixtures PASS 1.104s; capture byte-identical.
Mutants: all ten new compiling semantic variants caught; fifty-nine owned variants and the missing-consumer coverage mutant caught in the complete regression.
Not covered: full repository gate, whole-rule integration, independently ported dependency wiring, arbitrary malformed byte/arena adapters, nil system receivers and external Tailwind/corpora.

## Landing and ownership

The only branch this worker pushes is codex/lint-helpers-03. Before selection, HEAD 8185ae1 was verified based on fetched origin/main e8ba3d5 and had passing oracle evidence on that base. No rebase was needed. All eighteen origin codex/lint-helpers* branches were fetched and every claim checked; comments remain reserved by shared HELPERS.md. ParseCSS was the highest unclaimed concrete symbol at six consumers. The other two tied the next count at five. Claim 962fd49 was pushed before source. A subsequent complete fetch/scan confirms unique ownership in evidence/claims.log. No main, area branch, shared harness, registration generator, compiler or other worker directory is changed.

## Behavior and oracle

ParseCSS ports the actual byte-indexed Go state machine, preserving custom-property raw whitespace, independent bracket scanning, lazy whitespace peeks, comment/string routing, declaration boundaries, at-rule placement, nested tree order, license hoisting, nil/present child lists, value presence, importance and exact errors. Errors discard partial output. A surviving open parent produces Go's missing-closing-brace diagnostic; an unclosed parenthesis without a parent can still return an empty successful tree, matching Go. Unterminated string handling delegates Go's actual EOF behavior rather than inventing rejection.

UTF-8 is represented as a byte string so every index and escape consumption matches Go. Go replaces a three-byte BOM with one ASCII space; the port follows observed code and tests, despite the source comment's length-preservation claim. Relevant text fields are decoded through an explicit callback; context is always nil in actual parser output and is omitted. The result arena uses numeric child/root edges and a readonly node-array view.

Separately owned at-rule, declaration and string parsers, constructors, whitespace, trim and encoding/decoding are explicit dependencies. Owned temporary overlays wrap the real Go functions and retain exact argument/result observations, without changing the results. The test adapter matches exact buffer/colon and start/quote arguments, checks the normalized stylesheet passed to the string helper, and returns distinguishing results on an unexpected argument. The production parser has no lookup table or fixture-specific branch. This proves its state machine under the dependency adapter contracts; independent implementations and integration wiring of those dependencies remain outside this unit.

The message helper reproduces the complete shared wording, optional entry point, nil error and present empty error. Rule names and file names from every consumer's captured source corpus feed the controls, including Unicode/NUL/newline names and paths. Go's actual public function decides the expected text. Error formatting itself is an explicit already-rendered dependency.

The theme getter preserves the same nullable pointer handle and leaves system state unchanged. An owned Go export calls the actual method with nil and two independent themes, mutates the returned theme to prove the original observes it, and checks the stored field. Adamic uses stable numeric handles into the caller's theme arena; the same alias mutation and field checks run on Node and native. Invalid/nil system receivers are outside the initialized-system prerequisite, not silently treated as nil themes.

All 157 consuming-rule runtime source fixtures are additional CSS controls. The parser also reads every input in the checked-in 91-positive/11-negative engine-generated CSS corpus, eight real stylesheets among them, and systematic malformed/raw bracket/quote/whitespace/BOM/Unicode cases. The baseline compares 420 CSS outcomes, 312 message outcomes and three theme states: 735 lines on real Go, source Node, emitted JavaScript and sanitized native. This does not claim these JSX/TypeScript source controls are the CSS graph the rule loads, or that the rules ran on native.

## Compiler findings and validation

The first build refused nested function declarations. Local arrow functions compiled farther, but capturing the writable node arena prevented the cycle checker from proving allocations were safe. The final implementation keeps all node writes in their owning scope, publishes a readonly array view and moves the capture-free error constructor to file scope. No compiler change was needed. focused.log, closures.log, arena.log, readonly.log and ownership.log record superseded refused builds; none is a pass or a semantic mutant catch. functions.log is the passing baseline after those changes.

The initial focused gate passes in 31.279s with nine variants. All variants compile and exit 0 without stderr before comparison with Go; a refusal, panic, sanitizer failure or stderr never counts as a catch. The final regression additionally checks that a getter returning the correct theme cannot mutate the system's stored handle. Production source is never mutated; variants live in temporary directories.

The complete owned gate passes in 314.916s over 2,986,997 Go/Node/sanitized-native comparison lines across all twenty-seven helpers. Batches three through nine additionally compare emitted JavaScript; the first two retain their original source-Node/native checks. All fifty-nine compiling semantic variants and the missing-consumer mutant are caught in this invocation. Vet and formatting logs are empty with exit 0. Six uncached input fixtures pass in 1.104s, zero cache hits and six probe misses. Capture/coverage regeneration is byte-identical; exact hashes are in reproducibility.log.

The separate upstream command runs actual Go's TestParseCSS tests without an overlay. PASS 0.018s against the recorded Tailwind engine corpus, including eleven rejected malformed stylesheets and eight real stylesheets with 338 theme entries, 51 utility blocks and one custom variant. Its positive-test log prints zero nodes because that upstream test logs the total before its parallel subtests execute; this report does not credit that printed zero as a measured node count. All 91 positive subtests do run and pass.

Capture generation's whole upstream Tailwind rule package exits 1 on the known external installed-engine/corpus guards. The unavailable /Users/kirkouimet/Projects/ahra/app/_theme/styles and external class corpora remain blockers for that package, not for the checked-in CSS tests. Capture occurs before those skips and observes every selected consumer; it is not a green whole-rule package.

## Consumers and readiness

collapse.ParseCSS removes one dependency from each of:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

tailwind.DesignSystemDeclineMessage removes one dependency from each of:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

collapse.*LoadedDesignSystem.Theme removes one dependency from each of:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Sixteen edges across six unique rules, zero final blockers removed alone. readiness.json subtracts only this slot's twenty-seven delivered helpers from the original frozen ledger, never assuming another worker's claim is integrated. Dependency readiness under the common AST adapter assumption is not rule completion.

## Commands and environment

Source /workspace/adamic-tools/env.sh. Every test writes directly to a log:

- bash cloud/setup.sh: evidence/setup.log.
- python3 stage1/cohere/lint/helpers/slot03/batch9/testdata/regenerate.py: evidence/regenerate.log and capture.log; repeated capture in regenerate-repeat.log and reproducibility.log.
- go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch9' -count=1 -v -timeout=20m: evidence/mutants.log.
- go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m: evidence/final.log.
- go vet ./...: evidence/vet.log.
- gofmt -l cmd internal stage1/cohere/lint/helpers/slot03: evidence/gofmt.log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m: evidence/oracle.log.
- In cohere, go test ./internal/lint/rules/tailwind/collapse -run '^TestParseCSS' -count=1 -v -timeout=20m: evidence/upstream-css.log.

```
go version go1.27.1 linux/amd64
setup: go ready (1s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
v24.19.0
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (36s)
setup: done in 36s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Cohere pin 715ba94f3608a6500086b1076ce5cb7e51b836db; nproc=5. Setup warms build caches without running tests and is not a full test gate. The full repository gate, whole-rule integration, independent dependency ports, malformed adapters and external live corpora were not covered.

## Every final semantic mutant witness

- drop license hoisting: TestBatch9Mutants/parse_css.a; at line 31: got "ok:[rule:46,97,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]" Go "ok:[comment:::::33,32,108,105,99,101,110,115,101,32,:false:false:false:[]|rule:46,97,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]"
- skip BOM replacement: TestBatch9Mutants/parse_css.a#01; at line 418: got "ok:[rule:117,110,101,120,112,101,99,116,101,100,32,116,114,105,109,58,65279,46,97,32,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]" Go "ok:[rule:46,97,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]"
- route custom properties through ordinary whitespace collapsing: TestBatch9Mutants/parse_css.a#02; at line 7: got "ok:[]" Go "ok:[declaration::::45,45,120,::false:true:false:[]|]"
- split semicolons inside parentheses: TestBatch9Mutants/parse_css.a#03; at line 21: got "ok:[rule:46,97,:::::false:false:true:[unexpected declaration:::::98,97,99,107,103,114,111,117,110,100,58,32,117,114,108,40,97,:false:false:false:[]|unexpected declaration:::::98,46,112,110,103,41,:false:false:false:[]|]|]" Go "ok:[rule:46,97,:::::false:false:true:[declaration::::98,97,99,107,103,114,111,117,110,100,:117,114,108,40,97,59,98,46,112,110,103,41,:false:true:false:[]|]|]"
- ignore unterminated blocks: TestBatch9Mutants/parse_css.a#04; at line 27: got "ok:[]" Go "error:5:77,105,115,115,105,110,103,32,99,108,111,115,105,110,103,32,125,32,97,116,32,46,97,"
- shift every parser error offset: TestBatch9Mutants/parse_css.a#05; at line 5: got "error:1:77,105,115,115,105,110,103,32,111,112,101,110,105,110,103,32,40," Go "error:0:77,105,115,115,105,110,103,32,111,112,101,110,105,110,103,32,40,"
- omit the entry-point clause: TestBatch9Mutants/design_system_decline_message.a; at line 433: got "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32,60,110,105,108,62," Go "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,32,102,114,111,109,32,116,104,101,109,101,46,99,115,115,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32,60,110,105,108,62,"
- render nil error as empty text: TestBatch9Mutants/design_system_decline_message.a#01; at line 421: got "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32," Go "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32,60,110,105,108,62,"
- return nil instead of the stored theme: TestBatch9Mutants/loaded_theme.a; at line 734: got "-1:false:true" Go "0:true:true"
- mutate the stored theme while returning its original value: TestBatch9Mutants/loaded_theme.a#01; at line 734: got "0:true:false" Go "0:true:true"

All prior forty-nine variants were rerun in the same gate. Exact independent Go witnesses:

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
- TestSlot03HelperMutants/component_base_name.a/value; at line 159: got "false" Go "true"
- TestSlot03HelperMutants/tailwind_space.a/value; at line 847: got "false" Go "true"
- TestSlot03HelperMutants/listener_kinds.a/value; at line 1114948: got "JsxAttribute,CallExpression,StringLiteral" Go "JsxAttribute,CallExpression,VariableDeclaration"
- TestSlot03HelperMutants/listener_kinds.a/shared-list; at line 1114949: got "StringLiteral,CallExpression,VariableDeclaration" Go "JsxAttribute,CallExpression,VariableDeclaration"

Coverage mutant: consumer capture mismatch: missing [better-tailwindcss/no-unknown-classes] extra []
