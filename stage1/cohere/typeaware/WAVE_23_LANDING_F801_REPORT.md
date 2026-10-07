Built: rebased all worker work onto main f8013f0b and revalidated fifteen rule ports plus eighteen numeric listener/rule.json declarations; no new claims.
Commits: previous remote f1996689; tested rebased source 99cc6808b2338542f715b21f9bbcd71aab7848bb; this evidence is committed separately.
Commands and outputs: six wave gates PASS in 484.601s; bridge PASS 87.949s, checker PASS 0.153s; expanded Node oracle PASS 14.907s; vet, gofmt and diff checks pass.
Mutants: fifteen rule, seven additional output, eighteen listener, five released-registry, one checker-question guard and one JSON declaration mutant caught by their intended checks.
Not covered: full repository gate, shared numeric dispatch/Diagnostic adoption, speedup or native completion of the three reserved React rules.

The freshly fetched main is f8013f0baac41ddc340d76f83bddde38536a8f07. Since
previous e8ba3d5d it adds checker-narrowed element handling, override checks,
checking-pragma refusals and Map/Set iterator, compaction and hashing fixes.
All 25 branch commits rebase without conflicts. This is the only branch the
worker pushed. User-requested rebase and push uses an exact lease on the verified
old remote f199668992b4363333525fcd9ae961f78303f05d. No main or area branch is
written; no PR is opened. No implementation changes were needed for this rebase.

Five rule gates compare complete findings, automatic fixes and suggestions with
independent Go production rules, both normally and under ASan/UBSan with leak
checking. Each includes positive controls, compiled semantic mutants, normal
and sanitized frozen compiler/repository corpora and released-handle probes.
The 77 compiler roots and 287 frozen repository roots have zero findings for
these rules, with identical 5,318-byte and 18,485-byte streams respectively.
No output is normalized. Moving artifact directories changes control bytes
because paths are serialized. Documented nondefault options and custom
declarations are also held by these gates.

The sixth gate compares all eighteen rule.json declarations with Go registrations
and matches the 614-byte numeric listener stream in native, sanitized native
and emitted JavaScript under Node. It retains one compiled, clean-running kind
mutation per declaration. All mutants and single-run native/Go timing observations
from the new baseline follow. Timings include loading and serialization; these
are gate observations, not best-of-five speed claims, and early runs overlap
bridge/Node validation.

```
wave_23_behavior_test.go:162: controls: 328 submitted roots; 326 Go-parse-valid roots
wave_23_behavior_test.go:164: controls: 77309 identical finding bytes; findings 200
wave_23_behavior_test.go:172: controls-asan: 77309 identical finding bytes; findings 200
wave_23_behavior_test.go:183: throw mutant: exit 0, empty stderr, byte oracle catches byte 57708
wave_23_behavior_test.go:183: backreference mutant: exit 0, empty stderr, byte oracle catches byte 3962
wave_23_behavior_test.go:183: arrow mutant: exit 0, empty stderr, byte oracle catches byte 341
wave_23_behavior_test.go:193: callback symbol origin mutant: exit 0, empty stderr, Go bytes catch byte 40491
wave_23_behavior_test.go:223: repository timing: native 305.011078ms, Go 134.547588ms; native stderr tsgo: load_ns=83658462 query_ns=4422862 queries=25 first_query_ns=655038 run_ns=215328910
wave_23_behavior_test.go:223: compiler timing: native 2.319433114s, Go 351.103804ms; native stderr tsgo: load_ns=236504062 query_ns=201595639 queries=449 first_query_ns=5145110 run_ns=2065318355
wave_23_behavior_test.go:243: released handle: normal exit 70; registry mutant caught by required released-handle error
--- PASS: TestWave23BehaviorAgreementAndMutants (129.68s)
wave_23_constructor_test.go:138: controls: 25602 identical finding bytes; findings 54
wave_23_constructor_test.go:146: controls-asan: 25602 identical finding bytes; findings 54
wave_23_constructor_test.go:157: func mutant: exit 0, empty stderr, byte oracle catches byte 9065
wave_23_constructor_test.go:157: nonconstructor mutant: exit 0, empty stderr, byte oracle catches byte 1886
wave_23_constructor_test.go:157: wrappers mutant: exit 0, empty stderr, byte oracle catches byte 136
wave_23_constructor_test.go:189: repository timing: native 244.28404ms, Go 128.099991ms; native stderr tsgo: load_ns=75682553 query_ns=0 queries=0 first_query_ns=0 run_ns=163674950
wave_23_constructor_test.go:189: compiler timing: native 1.512641171s, Go 300.607166ms; native stderr tsgo: load_ns=242858830 query_ns=89743757 queries=2 first_query_ns=3266541 run_ns=1253892786
wave_23_constructor_test.go:209: released handle: normal exit 70; registry mutant caught by required released-handle error
--- PASS: TestWave23ConstructorAgreementAndMutants (64.93s)
wave_23_core_test.go:144: controls: 125642 identical finding bytes; findings 208
wave_23_core_test.go:152: controls-asan: 125642 identical finding bytes; findings 208
wave_23_core_test.go:174: eval mutant: exit 0, empty stderr, byte oracle catches byte 4142
wave_23_core_test.go:174: extend mutant: exit 0, empty stderr, byte oracle catches byte 8828
wave_23_core_test.go:174: func mutant: exit 0, empty stderr, byte oracle catches byte 72477
wave_23_core_test.go:203: eval-option mutant: exit 0, empty stderr, byte oracle catches byte 50652
wave_23_core_test.go:203: extend-option mutant: exit 0, empty stderr, byte oracle catches byte 3154
wave_23_core_test.go:222: repository timing: native 313.901284ms, Go 135.793726ms; native stderr tsgo: load_ns=69058496 query_ns=25691455 queries=643 first_query_ns=164147 run_ns=238293595
wave_23_core_test.go:222: compiler timing: native 3.989296452s, Go 431.101266ms; native stderr tsgo: load_ns=262944638 query_ns=230304989 queries=9190 first_query_ns=3326616 run_ns=3700489068
wave_23_core_test.go:242: released handle: normal exit 70; registry mutant caught by required released-handle error
--- PASS: TestWave23CoreAgreementAndMutants (98.28s)
wave_23_listener_test.go:65: eighteen rule.json name/kinds declarations match the independent Go registry
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
wave_23_listener_test.go:142: eighteen production registrations: 614 identical Go/native/sanitized/Node bytes
--- PASS: TestWave23NumericListeners (6.65s)
wave_23_next_test.go:134: controls: 56411 identical finding bytes; findings 120
wave_23_next_test.go:142: controls-asan: 56411 identical finding bytes; findings 120
wave_23_next_test.go:153: collection mutant: exit 0, empty stderr, byte oracle catches byte 1858
wave_23_next_test.go:153: outcome mutant: exit 0, empty stderr, byte oracle catches byte 29003
wave_23_next_test.go:153: pure mutant: exit 0, empty stderr, byte oracle catches byte 17200
wave_23_next_test.go:167: ancestry-count mutant: exit 0, empty stderr, Go byte oracle catches byte 29003
wave_23_next_test.go:167: awaited mutant: exit 0, empty stderr, Go byte oracle catches byte 32070
wave_23_next_test.go:185: repository timing: native 343.228039ms, Go 172.146533ms; native stderr tsgo: load_ns=68298907 query_ns=84970487 queries=8768 first_query_ns=234243 run_ns=269999890
wave_23_next_test.go:185: compiler timing: native 2.851418868s, Go 770.506143ms; native stderr tsgo: load_ns=236236184 query_ns=1008345318 queries=73411 first_query_ns=6045986 run_ns=2586356335
wave_23_next_test.go:205: released handle: normal exit 70; registry mutant caught by required released-handle error
--- PASS: TestWave23NextAgreementAndMutants (93.20s)
wave_23_test.go:104: controls: 13640 identical finding bytes; findings 39
wave_23_test.go:112: controls-asan: 13640 identical finding bytes; findings 39
wave_23_test.go:123: throw mutant: exit 0, empty stderr, byte oracle catches byte 99
wave_23_test.go:123: reject mutant: exit 0, empty stderr, byte oracle catches byte 5471
wave_23_test.go:123: reduce mutant: exit 0, empty stderr, byte oracle catches byte 11161
wave_23_test.go:137: alias-argument mutant: exit 0, empty stderr, Go byte oracle catches byte 6623
wave_23_test.go:137: rest-callback mutant: exit 0, empty stderr, Go byte oracle catches byte 4599
wave_23_test.go:155: repository timing: native 306.04003ms, Go 138.467021ms; native stderr tsgo: load_ns=76970920 query_ns=26261495 queries=1204 first_query_ns=173001 run_ns=223304107
wave_23_test.go:155: compiler timing: native 1.994752455s, Go 311.870529ms; native stderr tsgo: load_ns=250614350 query_ns=190733682 queries=1694 first_query_ns=3511616 run_ns=1725245923
wave_23_test.go:175: released handle: normal exit 70; registry mutant caught by required released-handle error
--- PASS: TestWave23AgreementAndMutants (91.85s)
```

Output mutants compile and exit 0 with empty stderr; only exact Go bytes catch
them. Normal released handles exit 70 while retained-registry mutants exit 0.
The question-guard overlay and JSON manifest mutation are separately rechecked:

```
=== RUN   TestCallbackSymbolFacts
    callback_symbol_facts_test.go:92: accepted malformed question
        invalid
--- FAIL: TestCallbackSymbolFacts (0.02s)
FAIL
FAIL	github.com/system-inc/adamic/bridge/tsgo/checker	0.020s
FAIL

=== RUN   TestWave23NumericListeners
    wave_23_listener_test.go:62: rule.json kinds differ from Go: got @typescript-eslint/only-throw-error	0, want @typescript-eslint/only-throw-error	258
--- FAIL: TestWave23NumericListeners (7.40s)
FAIL
FAIL	github.com/system-inc/adamic/stage1/cohere/typeaware	7.400s
FAIL

```

Both expected mutant commands exit 1 at their intended assertions, after their
normal controls pass. The JSON mutation changes [258] to [0] and is restored
to its exact original bytes in a finally block; git is clean after restoration.
There is no accidental production mutation counted as a check or a compiler
failure counted as a semantic mutant kill.

Exact commands, after source /workspace/adamic-tools/env.sh:

```sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-23/typescript
export ADAMIC_WAVE23_COMPILER_MANIFEST=/workspace/wave-23/compiler.manifest
export ADAMIC_WAVE23_REPOSITORY_MANIFEST=/workspace/wave-23/repository.manifest
export ADAMIC_WAVE23_ARTIFACTS=/workspace/wave-23/landing-f801/first
export ADAMIC_WAVE23_NEXT_ARTIFACTS=/workspace/wave-23/landing-f801/next
export ADAMIC_WAVE23_CORE_ARTIFACTS=/workspace/wave-23/landing-f801/core
export ADAMIC_WAVE23_CONSTRUCTOR_ARTIFACTS=/workspace/wave-23/landing-f801/constructor
export ADAMIC_WAVE23_BEHAVIOR_ARTIFACTS=/workspace/wave-23/landing-f801/behavior
export ADAMIC_WAVE23_LISTENER_ARTIFACTS=/workspace/wave-23/landing-f801/listeners
go test ./stage1/cohere/typeaware -run '^TestWave23((Next|Core|Constructor|Behavior)?AgreementAndMutants|NumericListeners)$' -count=1 -timeout=30m -v > /workspace/wave-23/landing-f801/rules.log 2>&1
ADAMIC_TSGO_CORPUS=/workspace/wave-23/typescript go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-23/landing-f801/bridge.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|functions|closures|sorting|call_targets_.*|devirtualize|047cb0d_n_.*|library_map_set_iterator_.*|override_same_representation)\.a$' -count=1 -timeout=10m -v > /workspace/wave-23/landing-f801/node.log 2>&1
go vet ./... > /workspace/wave-23/landing-f801/vet.log 2>&1
go test -overlay /workspace/wave-23/callback-guard-overlay.json ./bridge/tsgo/checker -run '^TestCallbackSymbolFacts$' -count=1 -v > /workspace/wave-23/landing-f801/callback-guard-mutant.log 2>&1
# With only-throw-error/rule.json temporarily changed to kinds [0], then restored:
ADAMIC_WAVE23_LISTENER_ARTIFACTS=/workspace/wave-23/landing-f801/json-mutant go test ./stage1/cohere/typeaware -run '^TestWave23NumericListeners$' -count=1 -timeout=10m -v > /workspace/wave-23/landing-f801/json-mutant.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware > /workspace/wave-23/landing-f801/gofmt.log 2>&1
git diff --check > /workspace/wave-23/landing-f801/diff.log 2>&1
```

The expanded Node run includes the freshly merged narrowing, Map/Set iterator
and override fixtures along with earlier call-target/devirtualization controls.
It reports native hits 24 / misses 57 and Node hits 0 / misses 57. This is a
filtered oracle and not a claimed full uncached repository gate.

The filesystem had only 1.4 GB free. To complete fresh builds, 146 rebuildable
ELF binaries and C archives were removed only from the named old scratch roots
/workspace/wave-23/landing and /workspace/wave-23/landing-current. Magic bytes
identified these generated artifacts, freeing 5,989,008,939 bytes. Source files,
input fixtures, old logs and committed evidence remain intact. Setup was already
complete: Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 83s, done 83s;
nproc 5. No setup rerun is claimed.

A fresh React capability probe built with this baseline's compiler and checker
still exits 0 on ordinary TypeScript and 70 on all three JSX controls, at bytes
53, 101 and 83. The shared parser still lacks JSX grammar and exposes string
kinds. The published JSX worker's implementation remains outside this unit's
shared-file scope. React static-components, set-state-in-effect and
set-state-in-render remain reserved and unfinished. Their numeric declarations
do not complete their decisions or lowering. No shared parser, driver, harness,
registration generator or protected compiler file was edited by this worker.

The user has not named the shared batch-8 Diagnostic landing sha. No numeric
callback dispatch, legacy string-kind/refetch migration, new finding-model
adoption or speedup is claimed here. No new rules were claimed.

[validation-wave23-landing-f801](validation-wave23-landing-f801) preserves complete
compressed raw streams/logs, SHA-256 fingerprints and fresh React probe errors.
Historical reports retain their old baseline IDs; this report records the new
rebased source and fresh observations.
