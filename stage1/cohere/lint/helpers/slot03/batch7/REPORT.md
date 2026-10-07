Built NewVariantRegistry, VariantRegistry.Register and VariantRegistry.AttachComparison in three .a files, eighteen dependency edges across six rules.
Commits: c9c0b38 claims all three before code; implementation and evidence follow the final gate.
Commands: twenty-one-helper gate PASS 281.505s, 2,980,424 comparison lines; vet/format clean; six uncached input fixtures PASS 1.412s; reproducible capture.
Mutants: wrong initial order, reused registry, moved replacement order, ignored group order, advanced lastOrder inside group, erased comparison on nil, refused callback replacement and shifted callback key.
Not covered: full repository test gate, complete rule integration, arbitrary invalid UTF-8, full int64/unsafe numeric states, external Tailwind installations/corpora.

## Ownership and implementation

All eighteen previous helpers were tested and pushed through c06493f before selection. Wildcard fetch found thirteen helper branches, every claims document was inspected and readiness.json was read/ranked. Comment symbols remain reserved and delivered by the shared worker's HELPERS.md; generic strict-option configuration is rule-local. These three tie at the highest available concrete-helper count, six. Claim c9c0b38 was pushed before code. A subsequent fetch found a later duplicate constructor claim by wave1-08; its owner already withdrew it after discovering this earlier claim, without writing an implementation. No shared harness, registration generator, compiler or other worker directory changed.

newVariantRegistry creates a fresh record and two independently writable, nonnil empty maps, lastOrder zero and absent group order. register preserves an existing registration's name and order while replacing kind with a fresh value record, matching Go map-value copying. A new registration uses the group's exact order when present, including zero/negative orders, otherwise lastOrder+1 and advances lastOrder only outside a group. attachComparison accepts an unregistered order, preserves an existing callback on nil, and overwrites it on nonnil attachment. Its store operation never examines variant values.

The registry.a data model is type-only, not an additional helper. Callback arguments are opaque numeric handles to the caller's ParsedVariant adapter; the helper stores function values without inspecting/invoking them. Oracle callbacks receive actual Go ParsedVariant values and source callbacks receive handles; both compute independent expected outputs from those arguments. Full variant parsing/Compare are separately blocked and are not reimplemented here. Safe-integer orders, initialized registries, sequential mutation and valid UTF-8-derived strings define the bounded adapter contract. Go nil group pointers use groupPresent; inactive groupOrder is irrelevant. Go nil callbacks use undefined.

## Observed parity and capture

Capture contains 157 actual runtime sources from all six consumers, plus derived lexical tokens and empty/NUL/supplementary/prototype-like name controls. 308 names, five kind strings, five order states and present/absent groups yield 15,400 cases. Each observes initialization, distinct instances, two registrations of one name, a new name, callback attachment, nil attachment, replacement and another callback key. Snapshot observes every registry state field, map sizes, complete name UTF-16 units, registration order/kind and invoked callback results. Actual Go constructors/methods decide the answers, with a temporary owned overlay exposing private group fields and snapshots. Baseline produces 415,800 identical lines across all four execution paths, with successful exit and empty stderr; native uses ASan/UBSan and leak checks.

The first run exposed a test-adapter error: absent groupOrder retained a supplied numeric value while Go's nil snapshot uses zero. The adapter was corrected before the credited runs. evidence/focused.log is the initial failed run and is not counted as a pass; its early reused-registry witness was the adapter mismatch and is not credited. The corrected focused/final runs must catch each semantic mutant independently after a passing baseline.

The upstream capture package exits 1 on its installed-Tailwind and corpus-population guards: /Users/kirkouimet/Projects/ahra/app/_theme/styles is unavailable and the external 11,314-class corpus is absent. Capturing occurs before these skips, and every selected consumer is recorded. This is explicitly not a passing upstream rule-package gate or whole-rule findings/fixes/suggestions parity. Repeated capture and metadata hashes are byte-identical.

## Consumers and readiness

Each of the three helpers removes one dependency from each of:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Eighteen dependency edges across six unique rules; zero additional final blockers removed alone. readiness.json subtracts only this slot's twenty-one delivered helpers, without assuming other branches are integrated. Readiness is an inventory calculation under the common AST adapter assumption, not observed rule completion.

## Mutants and gate

Every mutant is copied to a temporary directory, compiles, exits 0 and emits no stderr before its wrong output is compared with Go. Compilation errors, runtime errors and sanitizer findings are not credited as semantic catches. Production helper source is never mutated. All eight checks are rerun with the corrected baseline. Exact final witnesses and gate timing are appended below.

Commands, with output redirected directly to evidence files:

- python3 stage1/cohere/lint/helpers/slot03/batch7/testdata/regenerate.py: regenerate.log, capture.log; repeated source/coverage SHA-256 equality in reproducibility.log.
- go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch7' -count=1 -v -timeout=20m: corrected-focused.log.
- go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m: final.log.
- go vet ./...: vet.log, exit 0 and empty output.
- gofmt -l cmd internal stage1/cohere/lint/helpers/slot03: gofmt.log, empty output.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m: oracle.log, PASS 1.412s, six probe misses, zero cache hits.

Existing setup remains valid: go ready 0s, clang ready 1s, node ready 1s, submodules ready 1s, cache warm 106s, done 106s; nproc=5, cgroup cpu.max=400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0, source /workspace/adamic-tools/env.sh. Cohere pin 715ba94f3608a6500086b1076ce5cb7e51b836db. Full repository testing and integration remain unrun.

Corrected focused gate PASS 77.577s. All eight semantic mutants compiled and exited successfully before comparison: wrong initial order line 1; reused registry line 25; moved replacement order line 8; ignored group line 680; advanced lastOrder inside group line 679; nil erased comparison line 16; refused replacement line 21; shifted callback key line 15. The independent Go outputs distinguish missing callback, original callback -8 and replacement callback 15. Final package regression evidence follows.

## Final regression observations

Complete owned slot package PASS 281.505s: seven baseline batches total 2,980,424 matching actual Go, source Node and sanitized native lines; batches three through seven additionally compare emitted JavaScript. New batch baseline is 415,800 lines and all eight new semantic mutants are caught again. All thirty-nine compiling native mutants pass their expected semantic rejection, plus the missing-consumer coverage mutant. Source/capture gates and full-rule integration retain the stated limits. Final wildcard fetch checked eighteen helper branches; claims.log confirms unique active ownership, including the other worker's explicit withdrawal of its later duplicate constructor claim.

Every semantic mutant's final independent witness:

- TestBatch2Mutants/intrinsic_element_named.a: at line 24: got "true" Go "false"
- TestBatch2Mutants/hole_edges.a: at line 179230: got "false,false" Go "true,false"
- TestBatch2Mutants/read_class_values.a: at line 185959: got "Attribute::" Go "Callee::"
- TestBatch3Mutants/hex_value.a: at line 43: got "-1" Go "15"
- TestBatch3Mutants/unescape_string_literal_text.a: at line 1114212: got "0,99,111,112,121,10,:99,111,112,121," Go "0,169,10,:99,111,112,121,"
- TestBatch3Mutants/parameter_nodes.a: at line 1115385: got "2,-1,3,2" Go "-1,-1,3,2"
- TestBatch3Mutants/parameter_nodes.a#01: at line 1115375: got "0,1" Go ""
- TestBatch4Mutants/escape_terminator.a: at line 13: got "false" Go "true"
- TestBatch4Mutants/followed_by_whitespace.a: at line 3585: got "true:10" Go "false:10,11"
- TestBatch4Mutants/ignored_theme_key.a: at line 75243: got "true" Go "false"
- TestBatch5Mutants/split_theme_key.a: at line 283: got "1:" Go "0:"
- TestBatch5Mutants/split_theme_key.a#01: at line 286: got "0:" Go "1:"
- TestBatch5Mutants/split_theme_key.a#02: at line 288: got "1:99,104,97,110,103,101,100," Go "1:"
- TestBatch5Mutants/join_segments.a: at line 290: got "0," Go "0,45,"
- TestBatch5Mutants/breakpoint_group_order.a: at line 9639: got "2:true" Go "0:false"
- TestBatch5Mutants/breakpoint_group_order.a#01: at line 9641: got "23:true" Go "-17:true"
- TestBatch6Mutants/at_rule.a: at line 1: got "rule|||112,97,114,97,109,115,58,|||false|false|false|false|0|0|" Go "at-rule|||112,97,114,97,109,115,58,|||false|false|false|false|0|0|"
- TestBatch6Mutants/at_rule.a#01: at line 14: got "1:1:true:0" Go "1:1:true:-1"
- TestBatch6Mutants/at_rule.a#02: at line 14: got "0:0:true:" Go "1:1:true:-1"
- TestBatch6Mutants/at_rule.a#03: at line 3: got "at-rule|99,104,97,110,103,101,100,|99,104,97,110,103,101,100,|112,97,114,97,109,115,58,|||false|false|false|false|0|0|" Go "at-rule|||112,97,114,97,109,115,58,|||false|false|false|false|0|0|"
- TestBatch6Mutants/style_rule.a: at line 4: got "at-rule||||||false|false|false|false|0|0|" Go "rule||||||false|false|false|false|0|0|"
- TestBatch6Mutants/style_rule.a#01: at line 17: got "1:1:true:0" Go "1:1:true:-1"
- TestBatch6Mutants/style_rule.a#02: at line 17: got "0:0:true:" Go "1:1:true:-1"
- TestBatch6Mutants/style_rule.a#03: at line 6: got "rule|99,104,97,110,103,101,100,|99,104,97,110,103,101,100,||||false|false|false|false|0|0|" Go "rule||||||false|false|false|false|0|0|"
- TestBatch6Mutants/variant_next_order.a: at line 29488: got "-9007199254740989" Go "-9007199254740991"
- TestBatch6Mutants/variant_next_order.a#01: at line 29485: got "-9007199254740988" Go "-9007199254740989"
- TestBatch6Mutants/variant_next_order.a#02: at line 29486: got "false:0:-9007199254740989" Go "false:0:-9007199254740990"
- TestBatch7Mutants/new_variant_registry.a: at line 1: got "1:false:0:0:0" Go "0:false:0:0:0"
- TestBatch7Mutants/new_variant_registry.a#01: at line 25: got "84:false:0:2:2" Go "0:false:0:0:0"
- TestBatch7Mutants/register.a: at line 8: got ":84:replacement" Go ":83:replacement"
- TestBatch7Mutants/register.a#01: at line 680: got ":83:static" Go ":-17:static"
- TestBatch7Mutants/register.a#02: at line 679: got "-17:true:-17:1:0" Go "82:true:-17:1:0"
- TestBatch7Mutants/attach_comparison.a: at line 16: got "84:false:0:2:0" Go "84:false:0:2:1"
- TestBatch7Mutants/attach_comparison.a#01: at line 21: got "-8" Go "15"
- TestBatch7Mutants/attach_comparison.a#02: at line 15: got "missing" Go "-8"
- TestSlot03HelperMutants/component_base_name.a/value: at line 159: got "false" Go "true"
- TestSlot03HelperMutants/tailwind_space.a/value: at line 847: got "false" Go "true"
- TestSlot03HelperMutants/listener_kinds.a/value: at line 1114948: got "JsxAttribute,CallExpression,StringLiteral" Go "JsxAttribute,CallExpression,VariableDeclaration"
- TestSlot03HelperMutants/listener_kinds.a/shared-list: at line 1114949: got "StringLiteral,CallExpression,VariableDeclaration" Go "JsxAttribute,CallExpression,VariableDeclaration"

The coverage mutant removes no-unknown-classes from captured inputs; the coverage ledger rejects the missing consumer before parity is claimed. No compilation failure or sanitizer error is counted as a semantic catch.
