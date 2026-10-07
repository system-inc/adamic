Built: rebased wave 23 onto integrated lint area 7481e032 and ported react/jsx-no-undef in Adamic; sixteen native decisions now implemented.
Commits: prior remote 7f6e34b73da12e595814a57cae4c1d13efe5e60b; rebased foundation c364cbe054445e8141c8403c2ae877904ddac1f5; accompanying checkpoint commit.
Commands and outputs: existing six wave gates PASS 524.506s; new JSX rule gate PASS 55.760s; bridge PASS 69.367s, checker PASS 0.152s, filtered Node PASS 0.623s; vet/diff checks pass.
Mutants: sixteen rule and seven additional output mutations caught; eighteen internal listener mutations caught; five existing registry-handle mutations caught; checker guard intentionally fails; new rule rejects a released handle.
Not covered: react/jsx-fragments and react/jsx-no-constructed-context-values decision implementations remain unfinished; three analysis-dependent hooks rules remain parked; no new claims or full repository gate.

The user explicitly instructed a rebase onto origin/area/stage1-lint. Its fetched tip is 7481e0324e34a2537aafa9db7eeacda50405611b, including shared harness 41eb6eab2b6de45ede0a40250765be295ee25fbd and current main 39638d9e278d38bb5aeae887f46d55a70e47aaad. The rebase was clean and preserves integrated shared files. No shared registration generator, harness, parser or protected compiler file was edited here. Push targets only codex/typeaware-wave-23, using an exact lease on its prior remote tip.

The former JSX blocker is resolved: all three original capability inputs exit 0, print parsed SourceFile and have empty stderr. The claim status now says so; the remaining two rules are unfinished, not falsely marked blocked or parked. HIR/SSA/capture remains the parked hooks dependency (#dnv6f2c).

The new jsx_no_undef.a rule declares named kinds JsxOpeningElement and JsxSelfClosingElement and takes its handed ParseNode. The owned driver selects those kinds and fetches each node once. The rule classifies the tag's children, follows member receivers, excludes intrinsic/custom/namespaced/this tags, and uses the existing binding-declarations checker question to distinguish current-file declarations from globals. Default allowGlobals is false; .cjs allows globals, as Go does. No new checker question or bridge registration was needed. The owned rule.json declaration marks node: true. It remains declaration metadata under typeaware/listeners-wave23 rather than a new shared-registry integration.

A dedicated independent production-Go oracle calls only react.JsxNoUndef, with the allowGlobals option passed to the actual Go rule. The twelve generated TSX witness inputs exercise unresolved and declared names, block scope, lowercase intrinsic tags, uppercase dashed custom tags, underscores, dollar prefixes, Unicode, member receivers, this, globals and namespaced names. TSX files are generated witness inputs, not Adamic implementation modules. Both default and allowGlobals outputs match in native release and ASan/UBSan builds. The repeatable gate is TestWave23JsxNoUndefAgreementAndMutants; its default controls contain seven findings and zero fixes/suggestions. Full upstream JSX fixture coverage, import-equals witnesses and a dedicated .cjs witness are not claimed.

All sixteen native decisions match Go findings, fixes and suggestions on TypeScript compiler (77 frozen roots, 5318 bytes) and repository (287 frozen roots, 18485 bytes), with zero findings in those corpora and positive control findings checked separately. The existing fifteen also retain all their option, mutant and released-handle comparisons. The new uppercase/lowercase classification mutant compiles and exits normally with empty stderr; only the byte oracle catches it (byte 63 in the repeatable test; byte 57 in the preliminary manually generated paths). Releasing the checker before the new rule queries it causes exit 70 and exactly adamic: panic: invalid or released checker handle.

Timings are single observed runs, including loading. Native remains slower than Go. Complete gate timing and mutant observations:

```text
wave_23_behavior_test.go:183: throw mutant: exit 0, empty stderr, byte oracle catches byte 57708
wave_23_behavior_test.go:183: backreference mutant: exit 0, empty stderr, byte oracle catches byte 3962
wave_23_behavior_test.go:183: arrow mutant: exit 0, empty stderr, byte oracle catches byte 341
wave_23_behavior_test.go:193: callback symbol origin mutant: exit 0, empty stderr, Go bytes catch byte 40491
wave_23_behavior_test.go:223: repository timing: native 315.623697ms, Go 150.194742ms; native stderr tsgo: load_ns=83373541 query_ns=4511184 queries=25 first_query_ns=537390 run_ns=226121735
wave_23_behavior_test.go:223: compiler timing: native 2.212830778s, Go 377.616947ms; native stderr tsgo: load_ns=253725488 query_ns=190438832 queries=449 first_query_ns=10475515 run_ns=1922543722
wave_23_behavior_test.go:243: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_constructor_test.go:157: func mutant: exit 0, empty stderr, byte oracle catches byte 9065
wave_23_constructor_test.go:157: nonconstructor mutant: exit 0, empty stderr, byte oracle catches byte 1886
wave_23_constructor_test.go:157: wrappers mutant: exit 0, empty stderr, byte oracle catches byte 136
wave_23_constructor_test.go:189: repository timing: native 243.903465ms, Go 138.246298ms; native stderr tsgo: load_ns=78653050 query_ns=0 queries=0 first_query_ns=0 run_ns=160234726
wave_23_constructor_test.go:189: compiler timing: native 1.706409644s, Go 325.125277ms; native stderr tsgo: load_ns=270879312 query_ns=129951977 queries=2 first_query_ns=3563112 run_ns=1415483983
wave_23_constructor_test.go:209: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_core_test.go:174: eval mutant: exit 0, empty stderr, byte oracle catches byte 4142
wave_23_core_test.go:174: extend mutant: exit 0, empty stderr, byte oracle catches byte 8828
wave_23_core_test.go:174: func mutant: exit 0, empty stderr, byte oracle catches byte 72477
wave_23_core_test.go:203: eval-option mutant: exit 0, empty stderr, byte oracle catches byte 50652
wave_23_core_test.go:203: extend-option mutant: exit 0, empty stderr, byte oracle catches byte 3154
wave_23_core_test.go:222: repository timing: native 337.8951ms, Go 145.308691ms; native stderr tsgo: load_ns=73647895 query_ns=26642889 queries=644 first_query_ns=148504 run_ns=258648973
wave_23_core_test.go:222: compiler timing: native 4.051150654s, Go 390.39713ms; native stderr tsgo: load_ns=245752632 query_ns=210400988 queries=9190 first_query_ns=4252834 run_ns=3784652975
wave_23_core_test.go:242: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_listener_test.go:142: @typescript-eslint/only-throw-error listener mutant: exit 0, empty stderr, Go byte oracle catches byte 36
wave_23_listener_test.go:142: @typescript-eslint/prefer-promise-reject-errors listener mutant: exit 0, empty stderr, Go byte oracle catches byte 88
wave_23_listener_test.go:142: @typescript-eslint/prefer-reduce-type-parameter listener mutant: exit 0, empty stderr, Go byte oracle catches byte 144
wave_23_listener_test.go:142: nexus/correctness-no-collection-misuse listener mutant: exit 0, empty stderr, Go byte oracle catches byte 187
wave_23_listener_test.go:142: nexus/correctness-no-discarded-outcome listener mutant: exit 0, empty stderr, Go byte oracle catches byte 238
wave_23_listener_test.go:142: nexus/correctness-no-discarded-pure-result listener mutant: exit 0, empty stderr, Go byte oracle catches byte 285
wave_23_listener_test.go:142: no-eval listener mutant: exit 0, empty stderr, Go byte oracle catches byte 297
wave_23_listener_test.go:142: no-extend-native listener mutant: exit 0, empty stderr, Go byte oracle catches byte 329
wave_23_listener_test.go:142: no-func-assign listener mutant: exit 0, empty stderr, Go byte oracle catches byte 352
wave_23_listener_test.go:142: no-new-func listener mutant: exit 0, empty stderr, Go byte oracle catches byte 368
wave_23_listener_test.go:142: no-new-native-nonconstructor listener mutant: exit 0, empty stderr, Go byte oracle catches byte 405
wave_23_listener_test.go:142: no-new-wrappers listener mutant: exit 0, empty stderr, Go byte oracle catches byte 425
wave_23_listener_test.go:142: no-throw-literal listener mutant: exit 0, empty stderr, Go byte oracle catches byte 446
wave_23_listener_test.go:142: no-useless-backreference listener mutant: exit 0, empty stderr, Go byte oracle catches byte 475
wave_23_listener_test.go:142: prefer-arrow-callback listener mutant: exit 0, empty stderr, Go byte oracle catches byte 504
wave_23_listener_test.go:142: react-hooks/set-state-in-effect listener mutant: exit 0, empty stderr, Go byte oracle catches byte 540
wave_23_listener_test.go:142: react-hooks/set-state-in-render listener mutant: exit 0, empty stderr, Go byte oracle catches byte 576
wave_23_listener_test.go:142: react-hooks/static-components listener mutant: exit 0, empty stderr, Go byte oracle catches byte 610
wave_23_next_test.go:153: collection mutant: exit 0, empty stderr, byte oracle catches byte 1858
wave_23_next_test.go:153: outcome mutant: exit 0, empty stderr, byte oracle catches byte 29003
wave_23_next_test.go:153: pure mutant: exit 0, empty stderr, byte oracle catches byte 17200
wave_23_next_test.go:167: ancestry-count mutant: exit 0, empty stderr, Go byte oracle catches byte 29003
wave_23_next_test.go:167: awaited mutant: exit 0, empty stderr, Go byte oracle catches byte 32070
wave_23_next_test.go:185: repository timing: native 366.252242ms, Go 159.505277ms; native stderr tsgo: load_ns=78771233 query_ns=89084697 queries=8861 first_query_ns=291069 run_ns=281754801
wave_23_next_test.go:185: compiler timing: native 2.853768595s, Go 859.477072ms; native stderr tsgo: load_ns=250644266 query_ns=1005877570 queries=73411 first_query_ns=6072379 run_ns=2577451131
wave_23_next_test.go:205: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_test.go:123: throw mutant: exit 0, empty stderr, byte oracle catches byte 99
wave_23_test.go:123: reject mutant: exit 0, empty stderr, byte oracle catches byte 5471
wave_23_test.go:123: reduce mutant: exit 0, empty stderr, byte oracle catches byte 11161
wave_23_test.go:137: alias-argument mutant: exit 0, empty stderr, Go byte oracle catches byte 6623
wave_23_test.go:137: rest-callback mutant: exit 0, empty stderr, Go byte oracle catches byte 4599
wave_23_test.go:155: repository timing: native 353.172031ms, Go 167.504605ms; native stderr tsgo: load_ns=91360531 query_ns=31892296 queries=1213 first_query_ns=204989 run_ns=252062021
wave_23_test.go:155: compiler timing: native 2.086622708s, Go 319.202033ms; native stderr tsgo: load_ns=260753696 query_ns=231471423 queries=1694 first_query_ns=5136778 run_ns=1799944220
wave_23_test.go:175: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_jsx_undef_test.go:86: repository timing: native 233.26367ms, Go 144.175302ms
wave_23_jsx_undef_test.go:86: compiler timing: native 1.324757209s, Go 324.207465ms
wave_23_jsx_undef_test.go:114: jsx-no-undef mutant: exit 0, empty stderr, Go byte oracle catches byte 63
wave_23_jsx_undef_test.go:122: released handle rejected: exit 70 and required error
```

No Go regex exists in jsx-no-undef; its component-name predicate is the production Go byte predicate translated to the corresponding ASCII/Unicode string operations, not a replacement regex matcher. No existing regex implementation was edited.

Commands, with /workspace/adamic-tools/env.sh sourced and frozen manifests/per-batch artifact paths exported:

```sh
go test ./stage1/cohere/typeaware -run '^TestWave23((Next|Core|Constructor|Behavior)?AgreementAndMutants|NumericListeners)$' -count=1 -timeout=30m -v > /workspace/wave-23/landing-area/rules.log 2>&1
ADAMIC_WAVE23_JSX_UNDEF_ARTIFACTS=/workspace/wave-23/landing-area/undef-test go test ./stage1/cohere/typeaware -run '^TestWave23JsxNoUndefAgreementAndMutants$' -count=1 -timeout=15m -v > /workspace/wave-23/landing-area/undef-test.log 2>&1
ADAMIC_TSGO_CORPUS=/workspace/wave-23/typescript go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-23/landing-area/bridge.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|functions|closures|sorting|call_targets_.*|devirtualize|047cb0d_n_.*|library_map_set_iterator_.*|override_same_representation|inherited_static_field_read)\.a$' -count=1 -timeout=10m -v > /workspace/wave-23/landing-area/node.log 2>&1
go vet ./... > /workspace/wave-23/landing-area/vet-final.log 2>&1
go test -overlay /workspace/wave-23/callback-guard-overlay.json ./bridge/tsgo/checker -run '^TestCallbackSymbolFacts$' -count=1 -v > /workspace/wave-23/landing-area/callback-guard-mutant.log 2>&1
```

The checker guard intentionally fails with accepted malformed question. Setup was not rerun; initial timing is total 83s (cache warm 83s, Go/clang/Node/submodules 0s), nproc 5 (quota 4). Raw logs and byte-stream hashes are under validation-wave23-area. Tests wrote output to logs and were not piped. Historical reports describe their old snapshots and numeric manifests; current manifests use named kinds. Existing legacy typeaware dispatch has not been migrated to the shared registry.
