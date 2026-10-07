Built: rebased all fifteen completed ports and six reserved React declarations onto origin/main c01907a7; no new claims.
Commits: tested source 13d09cc75f794da0e27adc3161c5b130a499b76c; landing evidence is the accompanying commit on codex/typeaware-wave-23.
Commands and outputs: six wave gates PASS 476.138s; bridge PASS 79.949s, checker PASS 0.221s; filtered Node oracle PASS 4.416s; vet and diff checks pass.
Mutants: fifteen rule, seven additional output, twenty-one numeric listener, five released-registry, one checker guard and one JSON declaration mutant caught.
Not covered: three HIR/SSA/capture React rules parked; three new JSX rules lack native decisions because the shared parser rejects JSX; no full repository gate or shared dispatch migration.

The branch is rebased onto c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. This main advance adds Stage3 work, including internal/oracle/stage3_hook_test.go; those changes are retained. No shared harness, registration generator, parser, or protected compiler file was edited for this landing. The previous remote tip used for the explicit branch lease is f87766da70ea6d66599c8aecff4e5993afebc0aa. This report and compressed evidence are the only changes after the tested source.

All fifteen implemented rules retain byte-for-byte Go findings, fixes and suggestions on positive controls and frozen TypeScript compiler (77 roots) and repository (287 roots) corpora, in release and sanitized native builds. The corpus streams contain zero findings and respectively 5318 and 18485 bytes; positive controls exercise the actual decisions. No speed improvement is claimed. The measurements below are single runs, not statistical benchmarks.

```text
wave_23_behavior_test.go:223: repository timing: native 298.446898ms, Go 166.799536ms; native stderr tsgo: load_ns=79990270 query_ns=3751357 queries=25 first_query_ns=662738 run_ns=213742755
wave_23_behavior_test.go:223: compiler timing: native 2.366142079s, Go 360.562227ms; native stderr tsgo: load_ns=336603424 query_ns=188529237 queries=449 first_query_ns=6992107 run_ns=2009741967
wave_23_constructor_test.go:189: repository timing: native 257.489117ms, Go 138.699571ms; native stderr tsgo: load_ns=80697060 query_ns=0 queries=0 first_query_ns=0 run_ns=170770056
wave_23_constructor_test.go:189: compiler timing: native 1.574117788s, Go 316.536183ms; native stderr tsgo: load_ns=257601490 query_ns=118189780 queries=2 first_query_ns=4562178 run_ns=1297681334
wave_23_core_test.go:222: repository timing: native 314.555704ms, Go 135.260431ms; native stderr tsgo: load_ns=68512969 query_ns=27481249 queries=643 first_query_ns=130516 run_ns=239766908
wave_23_core_test.go:222: compiler timing: native 3.953769412s, Go 382.662207ms; native stderr tsgo: load_ns=239416249 query_ns=223602727 queries=9190 first_query_ns=4027982 run_ns=3691385402
wave_23_next_test.go:185: repository timing: native 357.742401ms, Go 187.987609ms; native stderr tsgo: load_ns=75925693 query_ns=87873040 queries=8768 first_query_ns=336297 run_ns=276537024
wave_23_next_test.go:185: compiler timing: native 2.937071166s, Go 829.206824ms; native stderr tsgo: load_ns=339855482 query_ns=1009314042 queries=73411 first_query_ns=7133934 run_ns=2566564371
wave_23_test.go:155: repository timing: native 320.098693ms, Go 141.341644ms; native stderr tsgo: load_ns=72961747 query_ns=30523674 queries=1204 first_query_ns=184088 run_ns=240921710
wave_23_test.go:155: compiler timing: native 1.981846459s, Go 317.415228ms; native stderr tsgo: load_ns=245911702 query_ns=203708660 queries=1694 first_query_ns=3865680 run_ns=1714625845
```

Every rule and additional decision/output mutant, plus the existing eighteen listener mutants and five released-handle registry mutations, is reported by the gates:

```text
wave_23_behavior_test.go:183: throw mutant: exit 0, empty stderr, byte oracle catches byte 57708
wave_23_behavior_test.go:183: backreference mutant: exit 0, empty stderr, byte oracle catches byte 3962
wave_23_behavior_test.go:183: arrow mutant: exit 0, empty stderr, byte oracle catches byte 341
wave_23_behavior_test.go:193: callback symbol origin mutant: exit 0, empty stderr, Go bytes catch byte 40491
wave_23_behavior_test.go:243: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_constructor_test.go:157: func mutant: exit 0, empty stderr, byte oracle catches byte 9065
wave_23_constructor_test.go:157: nonconstructor mutant: exit 0, empty stderr, byte oracle catches byte 1886
wave_23_constructor_test.go:157: wrappers mutant: exit 0, empty stderr, byte oracle catches byte 136
wave_23_constructor_test.go:209: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_core_test.go:174: eval mutant: exit 0, empty stderr, byte oracle catches byte 4142
wave_23_core_test.go:174: extend mutant: exit 0, empty stderr, byte oracle catches byte 8828
wave_23_core_test.go:174: func mutant: exit 0, empty stderr, byte oracle catches byte 72477
wave_23_core_test.go:203: eval-option mutant: exit 0, empty stderr, byte oracle catches byte 50652
wave_23_core_test.go:203: extend-option mutant: exit 0, empty stderr, byte oracle catches byte 3154
wave_23_core_test.go:242: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_listener_test.go:140: @typescript-eslint/only-throw-error listener mutant: exit 0, empty stderr, Go byte oracle catches byte 36
wave_23_listener_test.go:140: @typescript-eslint/prefer-promise-reject-errors listener mutant: exit 0, empty stderr, Go byte oracle catches byte 88
wave_23_listener_test.go:140: @typescript-eslint/prefer-reduce-type-parameter listener mutant: exit 0, empty stderr, Go byte oracle catches byte 144
wave_23_listener_test.go:140: nexus/correctness-no-collection-misuse listener mutant: exit 0, empty stderr, Go byte oracle catches byte 187
wave_23_listener_test.go:140: nexus/correctness-no-discarded-outcome listener mutant: exit 0, empty stderr, Go byte oracle catches byte 238
wave_23_listener_test.go:140: nexus/correctness-no-discarded-pure-result listener mutant: exit 0, empty stderr, Go byte oracle catches byte 285
wave_23_listener_test.go:140: no-eval listener mutant: exit 0, empty stderr, Go byte oracle catches byte 297
wave_23_listener_test.go:140: no-extend-native listener mutant: exit 0, empty stderr, Go byte oracle catches byte 329
wave_23_listener_test.go:140: no-func-assign listener mutant: exit 0, empty stderr, Go byte oracle catches byte 352
wave_23_listener_test.go:140: no-new-func listener mutant: exit 0, empty stderr, Go byte oracle catches byte 368
wave_23_listener_test.go:140: no-new-native-nonconstructor listener mutant: exit 0, empty stderr, Go byte oracle catches byte 405
wave_23_listener_test.go:140: no-new-wrappers listener mutant: exit 0, empty stderr, Go byte oracle catches byte 425
wave_23_listener_test.go:140: no-throw-literal listener mutant: exit 0, empty stderr, Go byte oracle catches byte 446
wave_23_listener_test.go:140: no-useless-backreference listener mutant: exit 0, empty stderr, Go byte oracle catches byte 475
wave_23_listener_test.go:140: prefer-arrow-callback listener mutant: exit 0, empty stderr, Go byte oracle catches byte 504
wave_23_listener_test.go:140: react-hooks/set-state-in-effect listener mutant: exit 0, empty stderr, Go byte oracle catches byte 540
wave_23_listener_test.go:140: react-hooks/set-state-in-render listener mutant: exit 0, empty stderr, Go byte oracle catches byte 576
wave_23_listener_test.go:140: react-hooks/static-components listener mutant: exit 0, empty stderr, Go byte oracle catches byte 610
wave_23_next_test.go:153: collection mutant: exit 0, empty stderr, byte oracle catches byte 1858
wave_23_next_test.go:153: outcome mutant: exit 0, empty stderr, byte oracle catches byte 29003
wave_23_next_test.go:153: pure mutant: exit 0, empty stderr, byte oracle catches byte 17200
wave_23_next_test.go:167: ancestry-count mutant: exit 0, empty stderr, Go byte oracle catches byte 29003
wave_23_next_test.go:167: awaited mutant: exit 0, empty stderr, Go byte oracle catches byte 32070
wave_23_next_test.go:205: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_test.go:123: throw mutant: exit 0, empty stderr, byte oracle catches byte 99
wave_23_test.go:123: reject mutant: exit 0, empty stderr, byte oracle catches byte 5471
wave_23_test.go:123: reduce mutant: exit 0, empty stderr, byte oracle catches byte 11161
wave_23_test.go:137: alias-argument mutant: exit 0, empty stderr, Go byte oracle catches byte 6623
wave_23_test.go:137: rest-callback mutant: exit 0, empty stderr, Go byte oracle catches byte 4599
wave_23_test.go:175: released handle: normal exit 70; registry mutant caught by required released-handle error
```

The checker question guard overlay failed with `accepted malformed question`; the JSON declaration mutant changed only-throw-error's kind 258 to 0 and failed with `rule.json kinds differ from Go`. The manifest was restored byte-for-byte. A verification script initially looked for uppercase JSON in the latter log and failed its own assertion; inspecting the diagnostic and verifying its exact lowercase text confirmed the intended test failure.

The three newly reserved rules remain react/jsx-fragments, react/jsx-no-constructed-context-values and react/jsx-no-undef. Their independent production-Go controls still match the previously recorded Go findings. Their numeric declarations match 107 Go bytes in release and sanitized native builds. Three successfully compiled first-kind-to-zero metadata mutants are caught at bytes 20, 72 and 99; these are declaration mutants, not rule-decision mutants. Exact current observations:

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

The first Go registration invocation omitted its required config and manifest and panicked with usage; it was corrected and rerun. These observations establish the shared frontend blocker, not native rule parity. The earlier react-hooks/set-state-in-effect, set-state-in-render and static-components claims remain parked on native HIR, SSA and capture analysis (#dnv6f2c). No further rules were claimed. Numeric `rule.json` declarations do not change the existing shared driver's behavior; legacy runtime dispatch has not been migrated here. No new regex matching implementation was added.

Commands were run with the existing toolchain environment sourced from /workspace/adamic-tools/env.sh. Initial setup measurements remain Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 83s, total 83s; nproc 5 (quota 4). Setup was not rerun.

```sh
source /workspace/adamic-tools/env.sh
# Frozen manifests and per-batch artifact paths exported under /workspace/wave-23/landing-c019.
go test ./stage1/cohere/typeaware -run '^TestWave23((Next|Core|Constructor|Behavior)?AgreementAndMutants|NumericListeners)$' -count=1 -timeout=30m -v > /workspace/wave-23/landing-c019/rules.log 2>&1
ADAMIC_TSGO_CORPUS=/workspace/wave-23/typescript go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-23/landing-c019/bridge.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|functions|closures|sorting|call_targets_.*|devirtualize|047cb0d_n_.*|library_map_set_iterator_.*|override_same_representation)\.a$' -count=1 -timeout=10m -v > /workspace/wave-23/landing-c019/node.log 2>&1
go vet ./... > /workspace/wave-23/landing-c019/vet.log 2>&1
go test -overlay /workspace/wave-23/callback-guard-overlay.json ./bridge/tsgo/checker -run '^TestCallbackSymbolFacts$' -count=1 -v > /workspace/wave-23/landing-c019/callback-guard-mutant.log 2>&1
# JSON mutation test expects failure, then restores original bytes.
# The JSX replay builds an independent Go oracle, release/sanitized declarations,
# three compiled metadata mutants and the existing shared-parser capability probe.
python3 /workspace/wave-23/landing-c019/recheck-jsx.py > /workspace/wave-23/landing-c019/jsx-recheck.log 2>&1
```

Compressed raw logs and stdout/stderr streams are in validation-wave23-landing-c019, with original-byte SHA-256 hashes. Test output was written to files, never piped. No full gate, Stage3 validation, native JSX findings/fixes/suggestions, or shared Diagnostic migration is claimed.
