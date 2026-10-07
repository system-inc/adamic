Built: rebased codex/typeaware-wave-23 onto current origin/main e8ba3d5d and revalidated fifteen existing rule ports; no new claims.
Commits: previous pushed tip a4413db6; freshly tested rebased source 263db6c294ae7d0c66b25361f39404c401b80674; this evidence is committed separately.
Commands and outputs: five rule gates PASS in 477.943s; bridge PASS 85.698s, checker PASS 0.377s, expanded Node oracle PASS 2.193s; vet and formatting checks exit 0.
Mutants: fifteen rule mutants, seven additional output mutants, five released-registry mutants and the malformed-question guard mutant caught by their intended checks.
Not covered: full repository gate or native completion of three reserved React rules; shared JSX frontend integration still blocks those ports.

The fetched main is e8ba3d5d81de4d3773c723914fccd4c76248b965. It added compiler
call-target routing and devirtualization since the preceding landing pass.
All 22 worker-branch commits rebased without conflicts. This worker has only
pushed codex/typeaware-wave-23; no main or area branch was written. The previous
verified remote tip is a4413db676ca1c13e932a1409c35cc63efe34159; the push uses
an exact lease on that tip, under the explicit user rebase-and-push instruction.

Every existing batch gate includes raw findings, fixes and suggestions byte
comparison with an independent Go oracle, positive controls, ASan/UBSan with
leak checking, rule mutants, both frozen corpora and released-handle tests.
The 77-root compiler stream and 287-root repository stream each contain zero
findings for these rules and match at 5,318 and 18,485 bytes respectively.
Positive controls remain mandatory. Artifact paths are serialized, so moving
scratch directories changes total control bytes; no output is normalized.
Nondefault option combinations and custom declarations also pass.

Fresh observations below identify every output mutant's first differing byte,
all five released-registry kills, control counts and native-versus-Go timings.
Output mutants compile, exit 0 and have empty stderr; the exact byte comparison
catches them. The unmodified released-handle probes exit 70 while the registry
mutants exit 0. The malformed callback-question overlay exits 1 at the intended
accepted-malformed-question assertion; the normal checker test passed first.
Single-run timing includes loading and serialization; these are gate observations,
not best-of-five benchmark claims. Bridge validation overlapped the first batch.

```
wave_23_behavior_test.go:162: controls: 328 submitted roots; 326 Go-parse-valid roots
wave_23_behavior_test.go:164: controls: 78287 identical finding bytes; findings 200
wave_23_behavior_test.go:172: controls-asan: 78287 identical finding bytes; findings 200
wave_23_behavior_test.go:183: throw mutant: exit 0, empty stderr, byte oracle catches byte 58431
wave_23_behavior_test.go:183: backreference mutant: exit 0, empty stderr, byte oracle catches byte 4004
wave_23_behavior_test.go:183: arrow mutant: exit 0, empty stderr, byte oracle catches byte 344
wave_23_behavior_test.go:193: callback symbol origin mutant: exit 0, empty stderr, Go bytes catch byte 41025
wave_23_behavior_test.go:223: repository timing: native 291.539375ms, Go 152.058491ms; native stderr tsgo: load_ns=73529586 query_ns=3897907 queries=25 first_query_ns=606195 run_ns=210070576
wave_23_behavior_test.go:223: compiler timing: native 4.304861235s, Go 701.843416ms; native stderr tsgo: load_ns=431848438 query_ns=541029116 queries=449 first_query_ns=5232736 run_ns=3828857648
wave_23_behavior_test.go:243: released handle: normal exit 70; registry mutant caught by required released-handle error
--- PASS: TestWave23BehaviorAgreementAndMutants (125.37s)
wave_23_constructor_test.go:138: controls: 25869 identical finding bytes; findings 54
wave_23_constructor_test.go:146: controls-asan: 25869 identical finding bytes; findings 54
wave_23_constructor_test.go:157: func mutant: exit 0, empty stderr, byte oracle catches byte 9131
wave_23_constructor_test.go:157: nonconstructor mutant: exit 0, empty stderr, byte oracle catches byte 1907
wave_23_constructor_test.go:157: wrappers mutant: exit 0, empty stderr, byte oracle catches byte 142
wave_23_constructor_test.go:189: repository timing: native 235.311371ms, Go 127.426172ms; native stderr tsgo: load_ns=75478779 query_ns=0 queries=0 first_query_ns=0 run_ns=154360077
wave_23_constructor_test.go:189: compiler timing: native 1.515225012s, Go 319.26581ms; native stderr tsgo: load_ns=237608897 query_ns=93816639 queries=2 first_query_ns=3260595 run_ns=1259896843
wave_23_constructor_test.go:209: released handle: normal exit 70; registry mutant caught by required released-handle error
--- PASS: TestWave23ConstructorAgreementAndMutants (61.86s)
wave_23_core_test.go:144: controls: 126416 identical finding bytes; findings 208
wave_23_core_test.go:152: controls-asan: 126416 identical finding bytes; findings 208
wave_23_core_test.go:174: eval mutant: exit 0, empty stderr, byte oracle catches byte 4166
wave_23_core_test.go:174: extend mutant: exit 0, empty stderr, byte oracle catches byte 8873
wave_23_core_test.go:174: func mutant: exit 0, empty stderr, byte oracle catches byte 72957
wave_23_core_test.go:203: eval-option mutant: exit 0, empty stderr, byte oracle catches byte 51069
wave_23_core_test.go:203: extend-option mutant: exit 0, empty stderr, byte oracle catches byte 3172
wave_23_core_test.go:222: repository timing: native 311.558009ms, Go 131.2741ms; native stderr tsgo: load_ns=66234374 query_ns=23845530 queries=643 first_query_ns=155564 run_ns=240563483
wave_23_core_test.go:222: compiler timing: native 3.956789779s, Go 372.543304ms; native stderr tsgo: load_ns=234893428 query_ns=243020266 queries=9190 first_query_ns=3127647 run_ns=3698553073
wave_23_core_test.go:242: released handle: normal exit 70; registry mutant caught by required released-handle error
--- PASS: TestWave23CoreAgreementAndMutants (98.96s)
wave_23_next_test.go:134: controls: 56537 identical finding bytes; findings 120
wave_23_next_test.go:142: controls-asan: 56537 identical finding bytes; findings 120
wave_23_next_test.go:153: collection mutant: exit 0, empty stderr, byte oracle catches byte 1861
wave_23_next_test.go:153: outcome mutant: exit 0, empty stderr, byte oracle catches byte 29078
wave_23_next_test.go:153: pure mutant: exit 0, empty stderr, byte oracle catches byte 17251
wave_23_next_test.go:167: ancestry-count mutant: exit 0, empty stderr, Go byte oracle catches byte 29078
wave_23_next_test.go:167: awaited mutant: exit 0, empty stderr, Go byte oracle catches byte 32151
wave_23_next_test.go:185: repository timing: native 345.676385ms, Go 160.103978ms; native stderr tsgo: load_ns=73890998 query_ns=84347786 queries=8768 first_query_ns=194464 run_ns=264855030
wave_23_next_test.go:185: compiler timing: native 2.627985476s, Go 823.670909ms; native stderr tsgo: load_ns=249566334 query_ns=910445301 queries=73411 first_query_ns=5665504 run_ns=2354531478
wave_23_next_test.go:205: released handle: normal exit 70; registry mutant caught by required released-handle error
--- PASS: TestWave23NextAgreementAndMutants (95.43s)
wave_23_test.go:104: controls: 13718 identical finding bytes; findings 39
wave_23_test.go:112: controls-asan: 13718 identical finding bytes; findings 39
wave_23_test.go:123: throw mutant: exit 0, empty stderr, byte oracle catches byte 102
wave_23_test.go:123: reject mutant: exit 0, empty stderr, byte oracle catches byte 5507
wave_23_test.go:123: reduce mutant: exit 0, empty stderr, byte oracle catches byte 11221
wave_23_test.go:137: alias-argument mutant: exit 0, empty stderr, Go byte oracle catches byte 6662
wave_23_test.go:137: rest-callback mutant: exit 0, empty stderr, Go byte oracle catches byte 4626
wave_23_test.go:155: repository timing: native 321.540984ms, Go 140.371312ms; native stderr tsgo: load_ns=75110609 query_ns=28969459 queries=1204 first_query_ns=160021 run_ns=240445643
wave_23_test.go:155: compiler timing: native 2.136834524s, Go 319.501145ms; native stderr tsgo: load_ns=281565644 query_ns=223689772 queries=1694 first_query_ns=4264862 run_ns=1833036530
wave_23_test.go:175: released handle: normal exit 70; registry mutant caught by required released-handle error
--- PASS: TestWave23AgreementAndMutants (96.32s)
callback guard mutant: TestCallbackSymbolFacts FAIL, accepted malformed question, exit 1
```

Exact commands, after source /workspace/adamic-tools/env.sh:

```sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-23/typescript
export ADAMIC_WAVE23_COMPILER_MANIFEST=/workspace/wave-23/compiler.manifest
export ADAMIC_WAVE23_REPOSITORY_MANIFEST=/workspace/wave-23/repository.manifest
export ADAMIC_WAVE23_ARTIFACTS=/workspace/wave-23/landing-current/first
export ADAMIC_WAVE23_NEXT_ARTIFACTS=/workspace/wave-23/landing-current/next
export ADAMIC_WAVE23_CORE_ARTIFACTS=/workspace/wave-23/landing-current/core
export ADAMIC_WAVE23_CONSTRUCTOR_ARTIFACTS=/workspace/wave-23/landing-current/constructor
export ADAMIC_WAVE23_BEHAVIOR_ARTIFACTS=/workspace/wave-23/landing-current/behavior
go test ./stage1/cohere/typeaware -run '^TestWave23(Next|Core|Constructor|Behavior)?AgreementAndMutants$' -count=1 -timeout=30m -v > /workspace/wave-23/landing-current/rules.log 2>&1
ADAMIC_TSGO_CORPUS=/workspace/wave-23/typescript go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-23/landing-current/bridge.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|functions|closures|sorting|call_targets_.*|devirtualize)\.a$' -count=1 -timeout=10m -v > /workspace/wave-23/landing-current/node.log 2>&1
go vet ./... > /workspace/wave-23/landing-current/vet.log 2>&1
go test -overlay /workspace/wave-23/callback-guard-overlay.json ./bridge/tsgo/checker -run '^TestCallbackSymbolFacts$' -count=1 -v > /workspace/wave-23/landing-current/callback-guard-mutant.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware > /workspace/wave-23/landing-current/gofmt.log 2>&1
git diff --check > /workspace/wave-23/landing-current/diff.log 2>&1
```

Only the intentional guard mutant exits nonzero. The expanded filtered Node run
includes main's call-target and devirtualization fixtures; its independent
one-byte mutant also passes. It reports native hits 13 / misses 27 and Node
hits 0 / misses 27. This is not a full uncached repository oracle claim.

Setup was completed earlier: Go ready 0s, clang ready 0s, Node ready 0s,
submodules 0s, cache warm 83s, done 83s; nproc 5. No setup rerun is claimed.

A new native React parser probe was built with this pass's fresh stage0 and
checker archive. Ordinary TypeScript exits 0; the static-components,
set-state-in-effect and set-state-in-render JSX controls still exit 70 at
bytes 53, 101 and 83. All three rules remain reserved and unfinished. Main,
the bridge branch and the shared harness branch lack JSX grammar; the published
codex/stage1-jsx-lint implementation at e715ef4a remains outside this unit's
shared-file territory. No shared parser, scanner or harness changes were made.
This landing pass does not claim React decisions, lowering, parity, mutants or
timings. No additional rules were claimed and no PR was opened.

[validation-wave23-landing-current](validation-wave23-landing-current) preserves
all raw batch stdout/stderr streams and logs compressed losslessly, with SHA-256
fingerprints, plus the fresh React errors. Earlier reports describe historical
pre-rebase commits and checks; this report identifies the newly tested baseline.
