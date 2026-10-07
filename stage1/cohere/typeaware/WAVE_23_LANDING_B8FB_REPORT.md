Built: fifteen completed ports and six reserved React declarations rebased onto origin/main b8fb957a; no new claims.
Commits: tested source c3ee08eedd06e3354bdb2cf5e644c95a6d632839; accompanying evidence commit on codex/typeaware-wave-23.
Commands and outputs: six wave gates PASS 506.259s; bridge PASS 81.094s, checker PASS 0.203s; filtered Node oracle PASS 4.911s; vet and diff checks pass.
Mutants: fifteen rule, seven additional output, twenty-one internal listener, five released-registry, one checker guard and one named JSON declaration mutation caught.
Not covered: three HIR/SSA/capture rules parked; three JSX rules remain frontend-blocked; no full repository gate or shared driver migration.

Main b8fb957aa839a9e8cb0b54279dd9864fa317bd30 adds inherited static-field emission. Its emission, oracle and count changes were retained through a clean rebase. No protected compiler, shared parser, harness or registration generator was edited. The explicit lease for pushing this rebase is the previous remote tip d7bd05fe324763d95c46a6cf6f574b9f1cdb3df8. The report and compressed evidence are the only changes after the tested source.

All fifteen completed rules match production Go byte-for-byte on findings, fixes and suggestions in release and sanitized builds, over positive controls and both frozen corpora: TypeScript compiler 77 roots (5318 bytes, zero findings), repository 287 roots (18485 bytes, zero findings). Positive control findings by batch: 39, 120, 208, 54, 200. Native handle rejection and sanitizer checks pass. The existing named rule.json declarations agree with production Go; internal numeric probe output is retained separately and is not the manifest contract. Eighteen probes match Go/native/sanitized/emitted-JavaScript output. Three JSX probes match 107 Go/native/sanitized bytes, with named manifests separately verified.

The complete measured timings and mutants are recorded below. Timings are individual runs including loading, not statistical benchmarks; native is still slower than Go.

```text
wave_23_behavior_test.go:183: throw mutant: exit 0, empty stderr, byte oracle catches byte 57708
wave_23_behavior_test.go:183: backreference mutant: exit 0, empty stderr, byte oracle catches byte 3962
wave_23_behavior_test.go:183: arrow mutant: exit 0, empty stderr, byte oracle catches byte 341
wave_23_behavior_test.go:193: callback symbol origin mutant: exit 0, empty stderr, Go bytes catch byte 40491
wave_23_behavior_test.go:223: repository timing: native 758.526535ms, Go 387.072512ms; native stderr tsgo: load_ns=290651636 query_ns=3705288 queries=25 first_query_ns=532854 run_ns=459186061
wave_23_behavior_test.go:223: compiler timing: native 3.305266623s, Go 392.158635ms; native stderr tsgo: load_ns=321028583 query_ns=358659676 queries=449 first_query_ns=6591496 run_ns=2912418343
wave_23_behavior_test.go:243: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_constructor_test.go:157: func mutant: exit 0, empty stderr, byte oracle catches byte 9065
wave_23_constructor_test.go:157: nonconstructor mutant: exit 0, empty stderr, byte oracle catches byte 1886
wave_23_constructor_test.go:157: wrappers mutant: exit 0, empty stderr, byte oracle catches byte 136
wave_23_constructor_test.go:189: repository timing: native 249.87197ms, Go 136.120997ms; native stderr tsgo: load_ns=78983762 query_ns=0 queries=0 first_query_ns=0 run_ns=164616738
wave_23_constructor_test.go:189: compiler timing: native 1.631085243s, Go 319.312322ms; native stderr tsgo: load_ns=272536736 query_ns=106299305 queries=2 first_query_ns=3511456 run_ns=1339809523
wave_23_constructor_test.go:209: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_core_test.go:174: eval mutant: exit 0, empty stderr, byte oracle catches byte 4142
wave_23_core_test.go:174: extend mutant: exit 0, empty stderr, byte oracle catches byte 8828
wave_23_core_test.go:174: func mutant: exit 0, empty stderr, byte oracle catches byte 72477
wave_23_core_test.go:203: eval-option mutant: exit 0, empty stderr, byte oracle catches byte 50652
wave_23_core_test.go:203: extend-option mutant: exit 0, empty stderr, byte oracle catches byte 3154
wave_23_core_test.go:222: repository timing: native 348.132003ms, Go 135.358396ms; native stderr tsgo: load_ns=77285204 query_ns=27578013 queries=643 first_query_ns=143706 run_ns=263116202
wave_23_core_test.go:222: compiler timing: native 3.900854778s, Go 384.801204ms; native stderr tsgo: load_ns=240814200 query_ns=217379820 queries=9190 first_query_ns=3320047 run_ns=3637024446
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
wave_23_next_test.go:185: repository timing: native 401.140754ms, Go 178.119529ms; native stderr tsgo: load_ns=78387799 query_ns=91010552 queries=8768 first_query_ns=164118 run_ns=311461483
wave_23_next_test.go:185: compiler timing: native 3.00327673s, Go 920.636564ms; native stderr tsgo: load_ns=246798528 query_ns=1108852530 queries=73411 first_query_ns=6598332 run_ns=2728393912
wave_23_next_test.go:205: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_test.go:123: throw mutant: exit 0, empty stderr, byte oracle catches byte 99
wave_23_test.go:123: reject mutant: exit 0, empty stderr, byte oracle catches byte 5471
wave_23_test.go:123: reduce mutant: exit 0, empty stderr, byte oracle catches byte 11161
wave_23_test.go:137: alias-argument mutant: exit 0, empty stderr, Go byte oracle catches byte 6623
wave_23_test.go:137: rest-callback mutant: exit 0, empty stderr, Go byte oracle catches byte 4599
wave_23_test.go:155: repository timing: native 309.863494ms, Go 152.007275ms; native stderr tsgo: load_ns=77642531 query_ns=27205084 queries=1204 first_query_ns=159060 run_ns=226329208
wave_23_test.go:155: compiler timing: native 2.070655561s, Go 364.31849ms; native stderr tsgo: load_ns=240243388 query_ns=214108118 queries=1694 first_query_ns=3759953 run_ns=1807277274
wave_23_test.go:175: released handle: normal exit 70; registry mutant caught by required released-handle error
```

Checker guard overlay intentionally fails with accepted malformed question. Named JSON mutation Identifier instead of ThrowStatement intentionally fails with rule.json kinds differ from Go; the exact original manifest bytes were restored. Three additional compiled JSX metadata mutations are caught at bytes 20, 72 and 99, not rule-decision mutations. Current JSX observations:

```json
{
  "native": {
    "exit": 0,
    "bytes": 107
  },
  "native-asan": {
    "exit": 0,
    "bytes": 107
  },
  "jsx_fragments_listener.a metadata mutant": {
    "exit": 0,
    "first_difference": 20
  },
  "jsx_no_constructed_context_values_listener.a metadata mutant": {
    "exit": 0,
    "first_difference": 72
  },
  "jsx_no_undef_listener.a metadata mutant": {
    "exit": 0,
    "first_difference": 99
  },
  "fragment JSX probe": {
    "exit": 70,
    "stderr": "adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 76 in /workspace/wave-23/jsx-next/fragment.tsx\n"
  },
  "context JSX probe": {
    "exit": 70,
    "stderr": "adamic: panic: parser slice expected GreaterThanToken, got Identifier at 61 in /workspace/wave-23/jsx-next/context.tsx\n"
  },
  "undef JSX probe": {
    "exit": 70,
    "stderr": "adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 34 in /workspace/wave-23/jsx-next/undef.tsx\n"
  }
}
```

react/jsx-fragments, react/jsx-no-constructed-context-values and react/jsx-no-undef still fail in the current shared frontend before a rule runs. ab70f38d4 is available on origin/lint-rules/harness but remains outside both main and area/stage1-lint at the fetched snapshot. react-hooks/set-state-in-effect, set-state-in-render and static-components remain parked on native HIR, SSA and capture analysis (#dnv6f2c). No new rule decision parity is claimed for these six reservations. No further rules were claimed. No regex matching implementation was added.

Commands, with /workspace/adamic-tools/env.sh sourced and frozen manifest/per-batch artifact exports under /workspace/wave-23/landing-b8fb:

```sh
go test ./stage1/cohere/typeaware -run '^TestWave23((Next|Core|Constructor|Behavior)?AgreementAndMutants|NumericListeners)$' -count=1 -timeout=30m -v > /workspace/wave-23/landing-b8fb/rules.log 2>&1
ADAMIC_TSGO_CORPUS=/workspace/wave-23/typescript go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-23/landing-b8fb/bridge.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|functions|closures|sorting|call_targets_.*|devirtualize|047cb0d_n_.*|library_map_set_iterator_.*|override_same_representation|inherited_static_field_read)\.a$' -count=1 -timeout=10m -v > /workspace/wave-23/landing-b8fb/node.log 2>&1
go vet ./... > /workspace/wave-23/landing-b8fb/vet.log 2>&1
go test -overlay /workspace/wave-23/callback-guard-overlay.json ./bridge/tsgo/checker -run '^TestCallbackSymbolFacts$' -count=1 -v > /workspace/wave-23/landing-b8fb/callback-guard-mutant.log 2>&1
# Named JSON mutation repeats TestWave23NumericListeners, expects failure and restores original bytes.
python3 /workspace/wave-23/landing-b8fb/recheck-jsx.py > /workspace/wave-23/landing-b8fb/jsx-recheck.log 2>&1
```

The filtered Node oracle includes inherited_static_field_read.a and proves the new main behavior against Node. No full repository gate was run. Toolchain setup was not rerun; initial measurements remain total 83s (cache warm 83s; Go, clang, Node and submodules 0s), nproc 5 (quota 4). Raw logs and stdout/stderr streams are compressed under validation-wave23-landing-b8fb, with original byte hashes. Test output was written to logs, never piped. Existing native runtime dispatch was not migrated to the unintegrated shared harness.
