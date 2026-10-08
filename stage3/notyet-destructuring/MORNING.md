# Destructuring topic-only continuation

Base: newest fetched origin/main 6998ebc24ae353193cb1495d3d51308131a4b5c7. Branch: codex/notyet-destructuring-topic. Only own non-merge commits were carried: a4d733d6 -> d2cf13cf, 2fc6f0ee -> 1f3c1894, 903ba7b2 -> b988a590. No replay, checked-non-null, area or other worker branch was merged. Old REPORT.md and its replay files are historical evidence from the previous branch, not this branch's ancestry or validation.

Counts conflicted during transplant. Current main's baseline was kept, then all fixture counts were regenerated. The object-literal conflict retained only this unit's explicit union packing and main's unrelated functions. The missing censusFieldSlotless dependency was replaced by a unit-local destructuringFieldSlotless helper, permitting optional boolean field slots and leaving explicit union packing at the construction sites.

## Morning lesson

The initial topic fixture run exposed an uncarried native dependency: optional booleans could not be stored or read in a field slot, so clang rejected six generated assignments/reads in notyet_destructuring_slots.a. This was a build failure, not a credited mutant kill. The topic implementation now stores optional booleans in non-reference numeric slots as undefined=2, false=0, true=1. A generated C helper takes the present/value pair once and restores it on read. Local/function pair representations remain unchanged. No runtime C file or header changed.

The new notyet_destructuring_option_flags.a fixture covers two optional flags, true versus false versus undefined, an actually absent property, a lazy default, initializer evaluation once, and a side-effecting optional flag producer packed exactly once. Its own registry is notyet_destructuring_option_flags_test.go. The first reductions exposed exactOptionalPropertyTypes and the separate shorthand-field gap. The final fixture explicitly permits undefined and uses ordinary property assignments. Neither failed reduction was claimed green.

Readonly field widening is refused at destructureFrom and tupleField when the source field representation differs from the view. The unit-local destructuring_slot_views.go guard checks contextual object/tuple views, including nested and callable types, and never creates a checked view. Three direct boundary probes cover boolean-to-optional, number-to-union and readonly tuple assignment. A guard-bypass mutant must fail the boundary test.

Supporting files outside the owned lowering functions: internal/lower/destructuring_bindings.go, destructuring_slot_views.go and destructuring_tuple_targets.go, internal/lower/object.go objectLiteral/tupleLiteral, internal/native/destructuring_boolean_slots.go, emit.go cProgram's helper declarations, emit_values.go member, slots.go slotted/unslotted, maybe.go maybeSlot. Native tests run for these hooks. No other worker's lower function was edited beyond the already-authorized literal hooks carried by the own commits.

## Morning census and checked views

Input: codex/stage3-notyet-table e8c283b5, stage3/notyet-table/rerun-0730/after/roots.csv. All 81 adapted source byte counts and SHA-256 values match its source-manifest.json. Replay tool 9a1f14c5 is used only in excluded scratch files; its branch is not merged or carried. The production compiler never receives the census overlay's permissive loader.

Twenty-one roots belong to these nine kinds. Four for...of object-binding roots and the destructured-parameter/default root belong to other functions and were excluded. Each owned kind was inspected largest first. All 21 exact target signatures are absent from the topic replay. This alone does not prove removal by this topic: the morning compiler included unlanded statics/binary/non-null work, while this topic starts at main. Earlier failures prevent some original statements or branches from being attempted. morning-replay.json preserves all findings; they are measurements on a checker-rejected program, not an executable compiler build.

| Topic commit | Kind | Morning roots | Result |
| --- | --- | ---: | --- |
| d2cf13cf | computed field name | 4 | Carried lesson held by three fixtures; scanner/sys clear; parser reaches generic function; visitor structural-method stop |
| 1f3c1894 | non-plain destructured name | 3 | Carried nested/default lesson; current checker/semver stop earlier; utilities reaches function returning undefined |
| morning commit + 1f3c1894 | field representation differs | 3 | Carried union lesson and new optional-flag fixture/native codec; declarations 310 now blocked at reading options after structural method calls, so not claimed lowered as an original site |
| 1f3c1894 | non-tuple array bindings | 3 | Two semver roots covered by the carried array lesson; utilities 9545 skipped for checked views |
| b988a590 | destructuring a value | 2 | Carried length lesson; checker 11491 has no findings; tsbuildPublic 2295 was already unreproduced, not claimed newly fixed |
| b988a590 | iterating a value | 2 | Skipped for checked views: NodeArray is a structural ReadonlyArray view with metadata, requiring a proven runtime representation before iteration |
| 1f3c1894 | watch-like tuple union | 2 | Carried exact union-slot fixture; replay reaches earlier structural call/poison reads |
| b988a590 | Map from non-array pairs | 1 | Carried custom-iterator Map lesson; next stop is core for...of over an object |
| b988a590 | destructuring a string | 1 | Carried UTF-16 length lesson; main replay now also sees earlier scanner blockers |

Checked-view skips: utilities.ts:9545:15 destructures a tuple-or-empty-array union with distinct runtime representations; expressionToTypeNode.ts:1242:52 and factory/nodeFactory.ts:6948:91 iterate NodeArray views. On main they currently stop even earlier at binary/prefix expressions or structural calls. These three roots are not implemented by inventing an unchecked cast, erasing the tuple/array distinction, or pretending a structural array view is a concrete native array. The remaining 18 roots' reduced lessons are covered by the carried fixtures and new flag fixture; no claim that all 18 original compiler sites execute or fully lower is made.

## Validation

Setup: GOPROXY=https://proxy.golang.org|direct, bash cloud/setup.sh, source /workspace/adamic-tools/env.sh. nproc=5, cgroup cpu.max=400000 100000. Timing lines: Go ready 3.002s; Node ready 3.025s; submodules ready 3.288s; markdown ready 3.518s (validated install skipped, step .076s); clang ready 4.131s; build cache warm 582.987s; done 584.373s. Complete log /tmp/destructuring-morning-setup.log. Go 1.27.1, Node v24.19.0, clang 20.1.8. No setup failure.

- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/notyet_' -count=1 -timeout 10m: all 13 topic fixtures passed 64.584s, source Node/native ASan+UBSan+leaks/emitted JavaScript Node. Log /tmp/destructuring-morning-oracle-final4.log.
- go test ./internal/native -run '^(TestMaybeNumbersPackIntoOneDouble|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked)$' -count=1 -timeout 10m: passed 26.614s. Log /tmp/destructuring-morning-native.log.
- go test ./internal/lower ./internal/javascript -run '^(TestComputed.*|TestDestructuringSlotViewsStayExplicit|TestObjectRefusalsExplainSoundness|TestObjectUnprovenShapesStayNotYet)$' -count=1: passed 10.404s; final log /tmp/destructuring-morning-lower-final4.log; JavaScript has no package test files.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 20m -args -update-counts: passed 206.311s with all 14 topic fixtures; final log /tmp/destructuring-hidden-counts-final.log.
- Replay: python3 /tmp/destructuring-morning-replay-all.py invokes the guarded census-replay worker once for each exact where/kind/reason against the complete tsc entry project. A second replay of declarations.ts:310:13 after the native codec reaches the same earlier reading-options stop. Logs /tmp/destructuring-morning-replay-before.log and /tmp/destructuring-morning-options-after.log.

The 26 original mutants are rerun on this topic using source overlays; the computed runner was changed to overlays so it never edits production source. New behavioral mutants are optional-boolean-presence, optional-undefined-pack, optional-value-unpack, optional-pack-once, and scanner-constructor-key. The checked-storage-view mutant separately proves all three refusal probes. Each must be caught specifically by stdout differs, not a type-check, lowering, Go or clang build failure. Final mutant outcomes are recorded below before the single completed-unit push. No whole-package test run or full gate was invoked.

## Scanner confirmation

The requested witness at codex/stage3-scanner-native-3 a3a8c599, stage3/drivers/scanner/evidence/land-area-next/witnesses/11-computed-field.a, is `const keywords = { ["" + "constructor"]: 1 }; console.log("ok");`. It is already supported by the carried constant-string concatenation rule. The dedicated notyet_computed_scanner_constructor.a fixture contains that witness and additionally checks the own key. It has its own registry, runs against Node in both backends with native sanitizers, and has a separate wrong-key mutant held to that exact fixture. No scanner-worker commit was merged.

The older scanner coordinate 113:5 is in the Scanner interface in the morning manifest. The same keyword expression is scanner.ts:150:5 in this manifest; the replay at that mapped coordinate is the relevant source check. The original requested witness is checked directly, without adapting its semantics.

The first strengthened scanner reduction tried to read keywords.constructor directly; TypeScript resolves that access as Function, which this console declaration does not accept. That fixture load failure was not a credited mutant kill. The final scanner fixture observes its own key with Object.keys instead.

Scanner replay after the final storage guard: selected scanner.ts:135:1 containing the key at 150:5; the computed-field signature did not reproduce (72.472s). See scanner-replay.json. Exit 1 is the replay tool's explicit absent-signature result, not a claimed successful complete compiler build.

The 13-fixture run passed 64.584s before the optional-producer observation was appended. That final fixture is rerun separately in /tmp/destructuring-morning-option-final8.log. Its codec mutations run from run-notyet-boolean-codec-mutants.py in /tmp/destructuring-morning-codec-mutants-final8.log. The count refresh is repeated after that fixture change. The changed library_string_raw counts also receive a targeted Node/native/emitted-JavaScript check in /tmp/destructuring-morning-string-raw.log.

The first duplicate-pack mutant survived because the field emitter already holds evaluated values in temporaries; duplicating those C names cannot repeat the producer. A closure reduction reaches main's separate refusal of function values returning boolean | undefined, so that reduction was withdrawn and that boundary remains unchanged. The final optional-pack-once mutant duplicates the MaybeBoolean field's evaluation before its temporary is held; it must change the producer-call observation. No surviving mutant or lowering failure is credited as a caught behavioral mutant.

## Hidden-region priority

Source input: codex/stage3-hidden-source 388096e6; the parser.ts and visitorPublic.ts SHA-256 values and byte counts match the adapted source exactly. Both heads were reproduced with a controlled scratch-only census overlay disabling this topic's const-enum key acceptance: visitorPublic.ts:624:5 (79.580s) and parser.ts:506:5 (27.011s). This recreates the former boundary; it is explicitly not a claim that the completed topic still refuses those keys or that another worker's branch was merged.

Final topic replays clear both exact computed-field signatures. The visitor replay takes 35.650s, the parser 23.387s. The unit-local concrete callback-table fixture notyet_computed_function_tables.a is registered by its own _test.go, uses both const-enum keys with function-expression values and enum-keyed destructuring, and checks keys and callback results against Node/native ASan+UBSan+leaks/emitted JavaScript. It passes 30.331s. The first wrong-number mutant caused a destructuring lower refusal instead of an output mismatch and was rejected as a behavioral proof. The final hidden-table-function-slot mutant swaps the two closure slot destinations and is caught by stdout differs. Existing enum-key and binding-key mutants separately prove name spelling and keyed lookup.

| Region | Previously hidden | Revealed | Still hidden | First next stop |
| --- | ---: | ---: | ---: | --- |
| visitorPublic.ts:619-1798 | 60,674 | 21,711 | 38,963 | 625:16, a method call through a structural signature in a program with statics; use typeof the declaring class |
| parser.ts:504-1136 | 49,243 | 0 | 49,243 | 506:33, a generic function expression |

These are regional replay ledger measurements, not a fresh global hidden-source census or proof that tsc executes. The visitor now reaches all 137 smaller callback-body boundaries instead of retaining the outer-table boundary. The parser retains one boundary over the entire variable statement because its first generic function expression remains refused. Generic function expressions and structural/static method calls are other workers' lowering territory; no unchecked lowering or foreign commit was added for them.

hidden/ contains trimmed raw before/after records, the two baseline metadata entries and RESULT.json. Reproduce the measurement with `python3 stage3/notyet-destructuring/measure-hidden-regions.py /tmp/destructuring-adapted/src/compiler stage3/notyet-destructuring/hidden`. The script checks source hashes, exact old boundary extents, attempted owner/status, absent checker/dependency skips and cleared computed signatures. It unions remaining half-open byte spans and independently checks that union against a per-byte mask. The parser's zero reveal is retained explicitly.

## Final validation and mutant outcomes

There are 14 registered topic fixtures. The 13-fixture run passed 64.584s; the final extended optional-producer fixture then passed 3.455s and the new hidden callback-table fixture passed 30.331s. The scanner's dedicated wrong-key mutant passes its catcher. The targeted moved-count library_string_raw oracle passed 6.682s. Final lower tests passed 10.404s, native hooks 26.614s, and the 14-fixture counts refresh 206.311s. JavaScript's package has no test files; its emitted programs are executed by the oracle. All test output is in the named /tmp log files. The superseded 13-fixture count refresh was cancelled after the hidden-table registry was added; the final 14-fixture refresh is the credited run. No full gate or whole-package test run was used.

All 33 intended mutants are caught: 28 by output mismatch and five by explicit refusal assertions. The source overlays never write production files.

| Mutants | Catcher | Log |
| --- | --- | --- |
| string-key, enum-key, binding-key, runtime-key-evaluation, own-proto | stdout differs | /tmp/destructuring-morning-computed-mutants.log |
| singleton-call, numeric-concatenation | want the computed-field gap | /tmp/destructuring-morning-computed-mutants.log |
| evaluate-once, holes, rest-copy, sticky-exhaustion, lazy-undefined-default, nested-field, object-union-box, tuple-union-box, tuple-field | stdout differs | /tmp/destructuring-morning-binding-mutants.log |
| array-length, utf16-length, map-interleaving, map-next-close, absent-array, absent-string, tuple-target-index, tuple-target-before-read | stdout differs | /tmp/destructuring-morning-collection-mutants.log |
| duplicate-field | want the repeated-field gap | /tmp/destructuring-morning-duplicate-mutant.log |
| storage-field | want the field storage gap | /tmp/destructuring-morning-storage-mutant.log |
| optional-boolean-presence | stdout differs | /tmp/destructuring-morning-option-mutant-final4.log |
| optional-undefined-pack, optional-value-unpack, optional-pack-once | stdout differs | /tmp/destructuring-morning-codec-mutants-final8.log |
| checked-storage-view | want the checked-slot-view gap in optional-boolean, boxed-union and tuple-assignment subtests | /tmp/destructuring-morning-view-mutant-final.log |
| scanner-constructor-key | stdout differs | /tmp/destructuring-morning-scanner-mutant-final.log |
| hidden-table-function-slot | stdout differs | /tmp/destructuring-hidden-table-mutant-final.log |

No runtime C file or header changed. No foreign non-merge commit or merge commit was carried. The remaining three morning checked-view roots, the parser's generic function expression and the visitor's structural/static method calls remain explicit dependencies.
