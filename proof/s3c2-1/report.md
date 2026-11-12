# S3c-2 proof — #qk8rztp

HEAD: `aef2e3c132c7ca0acc365932e88deec2b6033c48` (emit/generators-s3c2-1); base main: `730e7c26e590af0c57c545ba0cb905b09e4bee54`. HEAD is two commits above base. Run date: 2026-11-12 UTC.

## Decisive results


- **M1 SURVIVED**. M2, M3a, M3b and M4 were killed. Every edited package built; every restoration passed `git diff --exit-code`.
- Admission passed: exit 0, verdict `pass`, void `false`, diff_count 3, admitted 3, newly refused 0. The one native-build timeout passed its gate retry and the additional isolated rerun; main refused the same program without timing out.
- Step 1 formatting/vet and all four pin groups passed.
- Step 4 exited 1: 65 split tests lacked the oracle compiling-fixtures prerequisite. The other four packages passed.
- Step 5 exited 1 after 1884.196 seconds: lower passed; worker failed for missing products/WASI prerequisites.

## Step 6 — admission delta


```sh
go run ./cmd/adamic-admission-delta --base 730e7c26e590af0c57c545ba0cb905b09e4bee54 --head aef2e3c132c7ca0acc365932e88deec2b6033c48 --workers 2 --manifest /tmp/adamic-gate/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --json
```

Exit: 0; wall seconds: 1398.922. Run alone, workers=2, full generated manifest, no sampling budget or corpus filter.

```json
{
  "base": "730e7c26e590af0c57c545ba0cb905b09e4bee54",
  "head": "aef2e3c132c7ca0acc365932e88deec2b6033c48",
  "verdict": "pass",
  "void": false,
  "void_reason": null,
  "diff_count": 3,
  "admitted": 3,
  "classification": "lowering-only",
  "workers": 2,
  "complete_corpus": true,
  "corpus_filter": "",
  "sampling_size": 118,
  "omitted": 0,
  "timed_out_runs": 1,
  "retried_runs": 1,
  "still_timed_out_runs": 0,
  "phase_seconds": {
    "base_build": 55.422546068,
    "checkout": 14.859535378,
    "checkout_and_provision": 16.144249645,
    "classification": 781.383431095,
    "head_build": 38.934854864,
    "manifest_and_verification": 31.087487711,
    "provision": 1.284652623,
    "runtime": 474.979579507,
    "total": 1397.967363295
  },
  "newly_refused": 0,
  "program_classes": {
    "newly-accepted": 3,
    "accepted-by-both": 3319,
    "refused-by-both": 1982
  }
}
```

Admission stderr:

```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 1 runs timed out, 1 retried alone, 0 timed out again
changed programs: 3
```

### Every program whose classification or stderr differs (3)

#### `internal/oracle/testdata/generators_several/elaborate_four.a`

Gate class: `newly-accepted`; corpus: `diff`.
Base: `refused`, exit 1, error "". Complete stderr (278 characters):

```text
adamic: /tmp/adamic-gate/admission-delta-2490578997/head/internal/oracle/testdata/generators_several/elaborate_four.a:34:31: stage 0 can't lower a view that generators of more than one function feed (a step names the one function whose throws it may leave) (#qk8rztp S2c-1) yet
```

HEAD: `accepted`, exit 0, error "". Complete stderr (0 characters):

```text

```

agree:

```json
true
```

node:

```json
{
  "wall_seconds": 0.326520006,
  "exit": 0,
  "stdout": "t1a0;t1a1;=1 | t1o;=20 | t1c2;t1c1;=6 | =0\n=0 | z0o;=0 | =0 | =0\n",
  "stderr": ""
}
```

javascript:

```json
{
  "wall_seconds": 0.178061189,
  "exit": 0,
  "stdout": "t1a0;t1a1;=1 | t1o;=20 | t1c2;t1c1;=6 | =0\n=0 | z0o;=0 | =0 | =0\n",
  "stderr": ""
}
```

native:

```json
{
  "wall_seconds": 0.002763132,
  "exit": 0,
  "stdout": "t1a0;t1a1;=1 | t1o;=20 | t1c2;t1c1;=6 | =0\n=0 | z0o;=0 | =0 | =0\n",
  "stderr": ""
}
```

javascript_compile:

```json
{
  "wall_seconds": 0.13084104,
  "exit": 0,
  "stdout": "",
  "stderr": ""
}
```

native_compile:

```json
{
  "wall_seconds": 9.993915643,
  "exit": 0,
  "stdout": "",
  "stderr": ""
}
```

timed_out:

```json
[
  "native compile"
]
```

retried:

```json
true
```

#### `internal/oracle/testdata/generators_several/held_across_yield.a`

Gate class: `newly-accepted`; corpus: `diff`.
Base: `refused`, exit 1, error "". Complete stderr (281 characters):

```text
adamic: /tmp/adamic-gate/admission-delta-2490578997/head/internal/oracle/testdata/generators_several/held_across_yield.a:22:24: stage 0 can't lower a view that generators of more than one function feed (a step names the one function whose throws it may leave) (#qk8rztp S2c-1) yet
```

HEAD: `accepted`, exit 0, error "". Complete stderr (0 characters):

```text

```

agree:

```json
true
```

node:

```json
{
  "wall_seconds": 0.104893012,
  "exit": 0,
  "stdout": "after n1-0\nn1-0,\nrepeated done w2w2\nw2w2,w2w2!,\n\nafter m0-0\nafter m0-1\nm0-0,m0-1,\n",
  "stderr": ""
}
```

javascript:

```json
{
  "wall_seconds": 0.050166772,
  "exit": 0,
  "stdout": "after n1-0\nn1-0,\nrepeated done w2w2\nw2w2,w2w2!,\n\nafter m0-0\nafter m0-1\nm0-0,m0-1,\n",
  "stderr": ""
}
```

native:

```json
{
  "wall_seconds": 0.003447334,
  "exit": 0,
  "stdout": "after n1-0\nn1-0,\nrepeated done w2w2\nw2w2,w2w2!,\n\nafter m0-0\nafter m0-1\nm0-0,m0-1,\n",
  "stderr": ""
}
```

javascript_compile:

```json
{
  "wall_seconds": 0.075594971,
  "exit": 0,
  "stdout": "",
  "stderr": ""
}
```

native_compile:

```json
{
  "wall_seconds": 2.6351863030000002,
  "exit": 0,
  "stdout": "",
  "stderr": ""
}
```

#### `internal/oracle/testdata/generators_several/second_throws.a`

Gate class: `newly-accepted`; corpus: `diff`.
Base: `refused`, exit 1, error "". Complete stderr (277 characters):

```text
adamic: /tmp/adamic-gate/admission-delta-2490578997/head/internal/oracle/testdata/generators_several/second_throws.a:23:28: stage 0 can't lower a view that generators of more than one function feed (a step names the one function whose throws it may leave) (#qk8rztp S2c-1) yet
```

HEAD: `accepted`, exit 0, error "". Complete stderr (0 characters):

```text

```

agree:

```json
true
```

node:

```json
{
  "wall_seconds": 0.133223319,
  "exit": 0,
  "stdout": "quiet 0, quiet 1, \nloud 3, caught loud stopped after 3\n\nc1 once; c1 twice; \nloud 4; caught again loud stopped after 4\nc2 once; c2 twice; \n",
  "stderr": ""
}
```

javascript:

```json
{
  "wall_seconds": 0.08636974,
  "exit": 0,
  "stdout": "quiet 0, quiet 1, \nloud 3, caught loud stopped after 3\n\nc1 once; c1 twice; \nloud 4; caught again loud stopped after 4\nc2 once; c2 twice; \n",
  "stderr": ""
}
```

native:

```json
{
  "wall_seconds": 0.005111849,
  "exit": 0,
  "stdout": "quiet 0, quiet 1, \nloud 3, caught loud stopped after 3\n\nc1 once; c1 twice; \nloud 4; caught again loud stopped after 4\nc2 once; c2 twice; \n",
  "stderr": ""
}
```

javascript_compile:

```json
{
  "wall_seconds": 0.160792225,
  "exit": 0,
  "stdout": "",
  "stderr": ""
}
```

native_compile:

```json
{
  "wall_seconds": 3.673600477,
  "exit": 0,
  "stdout": "",
  "stderr": ""
}
```

### Timeout reruns alone at both SHAs


The gate recorded one native-compile timeout for `internal/oracle/testdata/generators_several/elaborate_four.a`; its automatic isolated HEAD retry passed. I additionally reran classification and native compilation sequentially at both exact cached compiler SHAs, with the gate's original 45-second classification / 10-second native limits. The cache `.inputs` files were checked to contain each exact SHA. Source was the unchanged HEAD fixture; base does not contain this newly added file, so both compilers were given that same HEAD program, as in the admission comparison. No other workload ran.

#### base classification


```json
{
  "label": "base classification",
  "command": [
    "/home/agent/.cache/adamic-build/local/027fcb9152937f26f93d97114259158a9a7989c7d0c18338f16520761f4d9981/adamic",
    "admission-lower",
    "internal/oracle/testdata/generators_several/elaborate_four.a"
  ],
  "timeout_seconds": 45,
  "wall_seconds": 0.1295260490005603,
  "exit": 1,
  "timed_out": false,
  "stdout": "",
  "stderr": "adamic: /workspace/adamic/internal/oracle/testdata/generators_several/elaborate_four.a:34:31: stage 0 can't lower a view that generators of more than one function feed (a step names the one function whose throws it may leave) (#qk8rztp S2c-1) yet\n"
}
```


#### base native compile


```json
{
  "label": "base native compile",
  "command": [
    "/home/agent/.cache/adamic-build/local/027fcb9152937f26f93d97114259158a9a7989c7d0c18338f16520761f4d9981/adamic",
    "build",
    "internal/oracle/testdata/generators_several/elaborate_four.a",
    "-o",
    "/tmp/adamic-gate/timeout-reruns/base-elaborate"
  ],
  "timeout_seconds": 10,
  "wall_seconds": 0.0876479520002249,
  "exit": 1,
  "timed_out": false,
  "stdout": "",
  "stderr": "adamic: /workspace/adamic/internal/oracle/testdata/generators_several/elaborate_four.a:34:31: stage 0 can't lower a view that generators of more than one function feed (a step names the one function whose throws it may leave) (#qk8rztp S2c-1) yet\n"
}
```


#### head classification


```json
{
  "label": "head classification",
  "command": [
    "/home/agent/.cache/adamic-build/local/17bf4215e46d35cdf9195f7161a89607625274fd4d9e1d1e391c2cfc64f4fffa/adamic",
    "admission-lower",
    "internal/oracle/testdata/generators_several/elaborate_four.a"
  ],
  "timeout_seconds": 45,
  "wall_seconds": 0.13114372600011848,
  "exit": 0,
  "timed_out": false,
  "stdout": "",
  "stderr": ""
}
```


#### head native compile


```json
{
  "label": "head native compile",
  "command": [
    "/home/agent/.cache/adamic-build/local/17bf4215e46d35cdf9195f7161a89607625274fd4d9e1d1e391c2cfc64f4fffa/adamic",
    "build",
    "internal/oracle/testdata/generators_several/elaborate_four.a",
    "-o",
    "/tmp/adamic-gate/timeout-reruns/head-elaborate"
  ],
  "timeout_seconds": 10,
  "wall_seconds": 2.3866361239997786,
  "exit": 0,
  "timed_out": false,
  "stdout": "",
  "stderr": ""
}
```


#### head native execution


```json
{
  "label": "head native execution",
  "command": [
    "/tmp/adamic-gate/timeout-reruns/head-elaborate"
  ],
  "timeout_seconds": 10,
  "wall_seconds": 0.0034992000000784174,
  "exit": 0,
  "timed_out": false,
  "stdout": "t1a0;t1a1;=1 | t1o;=20 | t1c2;t1c1;=6 | =0\n=0 | z0o;=0 | =0 | =0\n",
  "stderr": ""
}
```


#### source Node


```json
{
  "label": "source Node",
  "command": [
    "node",
    "--disable-warning=ExperimentalWarning",
    "/workspace/adamic/oracle/node.mjs",
    "/workspace/adamic/internal/oracle/testdata/generators_several/elaborate_four.a"
  ],
  "timeout_seconds": 10,
  "wall_seconds": 0.15230948500084196,
  "exit": 0,
  "timed_out": false,
  "stdout": "t1a0;t1a1;=1 | t1o;=20 | t1c2;t1c1;=6 | =0\n=0 | z0o;=0 | =0 | =0\n",
  "stderr": ""
}
```


Compiler input identities:

```text
base: /home/agent/.cache/adamic-build/local/027fcb9152937f26f93d97114259158a9a7989c7d0c18338f16520761f4d9981/adamic
name admission-compiler
flag 730e7c26e590af0c57c545ba0cb905b09e4bee54
flag adapter=fd17ea5ed37137cf8666c675e74699284f529212114e9958f9d7005569da6bfd
flag -trimpath
flag -buildvcs=false
flag GOFLAGS=
flag CGO_ENABLED=
flag GOEXPERIMENT=
tool go1.27.1
tool linux
tool amd64
```


```text
head: /home/agent/.cache/adamic-build/local/17bf4215e46d35cdf9195f7161a89607625274fd4d9e1d1e391c2cfc64f4fffa/adamic
name admission-compiler
flag aef2e3c132c7ca0acc365932e88deec2b6033c48
flag adapter=fd17ea5ed37137cf8666c675e74699284f529212114e9958f9d7005569da6bfd
flag -trimpath
flag -buildvcs=false
flag GOFLAGS=
flag CGO_ENABLED=
flag GOEXPERIMENT=
tool go1.27.1
tool linux
tool amd64
```


## Step 3 — exact mutants

The table has five edits (M3 is split into M3a and M3b). Each was applied alone after asserting the old text occurs exactly once in its named file; the file was restored byte for byte in a finally block before `git diff --exit-code`. No compiler fixes were made.

### M1: SURVIVOR

File: `internal/ir/generator.go`; old occurrence count: 1.
Old:

```text
for _, function := range next.Functions {
```

New:

```text
for _, function := range next.Functions[:1] {
```

Package build exit: 0; build wall: 0.809 seconds; restore diff exit: 0.

```sh
go test ./internal/oracle -timeout 60m -run '^(TestGeneratorsSeveralAgreeWithNode)$' -count=1 -v
```

Test exit: 0; wall: 82.500 seconds.
All three subtests stayed green with ADAMIC_GATE_UNCACHED=1. The native and Node outputs had no reported differing lines; there are no red subtests or got/want diagnostics to quote. The table predicted red, but this observed run does not establish that prediction.
Complete failure diagnostics (got and want retained verbatim):

```text
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 4e61731957d4 hit 31.89
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 4e61731957d4 hit 31.64
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 4e61731957d4 miss 32.18
```

Red tests/subtests:

```text
(none)
```

### M2: KILLED

File: `internal/lower/iterable_views.go`; old occurrence count: 1.
Old:

```text
if other := l.result.Functions[function].Yields; other != yields {
```

New:

```text
if other := l.result.Functions[function].Yields; false && other != yields {
```

Package build exit: 0; build wall: 15.662 seconds; restore diff exit: 0.

```sh
go test ./internal/lower -timeout 40m -run '^(TestGeneratorIntoAnIterableParameterLowers|TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad|TestNestedIterableCalleeLowers|TestNestedIterableCalleeRoadsAreNotYet|TestNestedIterableCalleeIsExactlyOneBody|TestHandDrivenNextOverAGeneratorStops|TestHandDrivenNextOverAGeneratorLowers|TestHandDrivenNextStopsOnEveryOtherRoad|TestGraphFlowFollowsEveryExpression|TestGraphFlowSteps|TestLocalFieldsClassifyEveryIntField|TestGeneratorViewDispatchesOnTheHeapKind)$' -count=1 -v
```

Test exit: 1; wall: 26.997 seconds.
Complete failure diagnostics (got and want retained verbatim):

```text
    graph_flow_classified_test.go:69: 144 expressions: 17 planned, 21 wrappers, 6 forward, 9 planned elsewhere, 82 can't hold, 9 stop
    iterable_generators_test.go:135: got /tmp/adamic-gate/TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoadtwo_ge3972001505/001/main.a:3:94: stage 0 can't lower a view over an iterable whose origins yield values held differently (a generator from counted() held as number; a generator from words() held as string, read as union of differently held members) yet, want a NotYet saying "a view that generators of more than one function feed, yielding"
```

Red tests/subtests:

```text
--- FAIL: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad (0.00s)
--- FAIL: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_string_(S3c-2) (0.13s)
```

### M3a: KILLED

File: `internal/worker/purity.go`; old occurrence count: 1.
Old:

```text
for _, function := range node.Functions {
```

New:

```text
for _, function := range node.Functions[:1] {
```

Package build exit: 0; build wall: 6.895 seconds; restore diff exit: 0.

```sh
go test ./internal/worker -timeout 30m -run '^(TestCheckPureRefusesAnOperationNoReviewNames|TestCheckPureInspectsAGeneratorStepsBody|TestCheckPureInspectsEveryBodyAGeneratorStepMayRun)$' -count=1 -v
```

Test exit: 1; wall: 9.982 seconds.
Complete failure diagnostics (got and want retained verbatim):

```text
    purity_inert_test.go:58: want the second stepped body's console write refused, got <nil>
```


```text
--- PASS: TestCheckPureRefusesAnOperationNoReviewNames (0.00s)
--- PASS: TestCheckPureInspectsAGeneratorStepsBody (0.00s)
--- FAIL: TestCheckPureInspectsEveryBodyAGeneratorStepMayRun (0.00s)
```

### M3b: KILLED

File: `internal/native/reuse.go`; old occurrence count: 1.
Old:

```text
for _, function := range expression.Functions {
```

New:

```text
for _, function := range expression.Functions[:1] {
```

Package build exit: 0; build wall: 2.679 seconds; restore diff exit: 0.

```sh
go test ./internal/native -timeout 30m -run '^(TestReuseTakesAnOperationNoReviewNamesAsTouchingTheGlobal|TestReuseFollowsAGeneratorStepIntoItsBody|TestReuseFollowsEveryBodyAGeneratorStepMayRun)$' -count=1 -v
```

Test exit: 1; wall: 21.832 seconds.
Got: `touches(program, 0, 0, map[int]bool{}) == false`. Want: `true`, because the second stepped generator body reads the global. The test reports this in words rather than printing a literal got/want pair.
Complete failure diagnostics (got and want retained verbatim):

```text
    reuse_inert_test.go:68: a step of generators the second of which reads the global is taken to touch nothing
```


```text
--- PASS: TestReuseTakesAnOperationNoReviewNamesAsTouchingTheGlobal (0.00s)
--- FAIL: TestReuseFollowsEveryBodyAGeneratorStepMayRun (0.00s)
--- PASS: TestReuseFollowsAGeneratorStepIntoItsBody (0.00s)
```

### M4: KILLED

File: `internal/lower/iterable_views.go`; old occurrence count: 1.
Old:

```text
function, err := l.oneGeneratorFunction(view, generators)
```

New:

```text
function, err := generators[0].generator-1, error(nil)
```

Package build exit: 0; build wall: 13.431 seconds; restore diff exit: 0.

```sh
go test ./internal/lower -timeout 40m -run '^(TestGeneratorIntoAnIterableParameterLowers|TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad|TestNestedIterableCalleeLowers|TestNestedIterableCalleeRoadsAreNotYet|TestNestedIterableCalleeIsExactlyOneBody|TestHandDrivenNextOverAGeneratorStops|TestHandDrivenNextOverAGeneratorLowers|TestHandDrivenNextStopsOnEveryOtherRoad|TestGraphFlowFollowsEveryExpression|TestGraphFlowSteps|TestLocalFieldsClassifyEveryIntField|TestGeneratorViewDispatchesOnTheHeapKind)$' -count=1 -v
```

Test exit: 1; wall: 23.426 seconds.
Complete failure diagnostics (got and want retained verbatim):

```text
    graph_flow_classified_test.go:69: 144 expressions: 17 planned, 21 wrappers, 6 forward, 9 planned elsewhere, 82 can't hold, 9 stop
    iterable_hand_next_test.go:221: got <nil>, want a NotYet saying "generators of more than one function"
```

Red tests/subtests:

```text
--- FAIL: TestHandDrivenNextOverAGeneratorStops (0.00s)
--- FAIL: TestHandDrivenNextOverAGeneratorStops/two_generator_functions (0.06s)
```

## Step 1 — checkout, formatting and vet

Fetched `origin emit/generators-s3c2-1`, checked out `aef2e3c132c7ca0acc365932e88deec2b6033c48` detached, and completed `git submodule update --init --recursive`. Submodule pins: cohere `5360753e19e4417675a90e92a1823bea66b13918`; cohere/TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. Initial working tree was clean; every mutant restoration was clean.

```sh
gofmt -l cmd internal stage3
```

Exit 0; output empty (0 bytes).

```sh
go vet ./internal/ir ./internal/lower ./internal/native ./internal/worker ./internal/oracle ./internal/flow ./internal/fresh ./internal/split ./internal/javascript ./stage3/census/latent/statecopy
```

Exit 0; output empty; wall_seconds=235.928
. No compile errors were emitted.
The first vet launch did not execute because the timing wrapper `/usr/bin/time` was absent (exit 127). It was immediately rerun with Bash timing; the successful vet above is that actual execution.

## Step 2 — pins, per-test and per-subtest results

Each command ran with `-count=1 -v`; oracle also had `ADAMIC_GATE_UNCACHED=1`. No failures or skips occurred in the baseline pins, so no failing got/want pairs exist. Logs were saved individually outside the repository.

### 2a: PASS


```sh
go test ./internal/lower -timeout 40m -run '^(TestGeneratorIntoAnIterableParameterLowers|TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad|TestNestedIterableCalleeLowers|TestNestedIterableCalleeRoadsAreNotYet|TestNestedIterableCalleeIsExactlyOneBody|TestHandDrivenNextOverAGeneratorStops|TestHandDrivenNextOverAGeneratorLowers|TestHandDrivenNextStopsOnEveryOtherRoad|TestGraphFlowFollowsEveryExpression|TestGraphFlowSteps|TestLocalFieldsClassifyEveryIntField|TestGeneratorViewDispatchesOnTheHeapKind)$' -count=1 -v
```

Exit 0; wall 31.231 seconds; log `/tmp/adamic-gate/step2a.log`.

```text
--- PASS: TestGraphFlowSteps (0.00s)
--- PASS: TestGraphFlowFollowsEveryExpression (0.03s)
--- PASS: TestLocalFieldsClassifyEveryIntField (0.03s)
--- PASS: TestGeneratorViewDispatchesOnTheHeapKind (0.31s)
--- PASS: TestHandDrivenNextOverAGeneratorStops (0.00s)
--- PASS: TestHandDrivenNextOverAGeneratorStops/Generator<boolean_|_undefined> (0.33s)
--- PASS: TestHandDrivenNextOverAGeneratorStops/two_generator_functions (0.30s)
--- PASS: TestHandDrivenNextOverAGeneratorStops/forward(iterator)_beside_a_next (0.31s)
--- PASS: TestHandDrivenNextOverAGeneratorStops/holder.it_narrowed,_then_next() (0.34s)
--- PASS: TestHandDrivenNextOverAGeneratorStops/Generator<number_|_undefined> (0.14s)
--- PASS: TestHandDrivenNextOverAGeneratorLowers (0.00s)
--- PASS: TestHandDrivenNextOverAGeneratorLowers/checker.ts:21636's_loop (0.11s)
--- PASS: TestHandDrivenNextOverAGeneratorLowers/a_number_generator,_one_next (0.14s)
--- PASS: TestHandDrivenNextOverAGeneratorLowers/a_generator_and_a_Map's_values()_at_two_call_sites (0.16s)
--- PASS: TestHandDrivenNextOverAGeneratorLowers/Generator<string_|_undefined> (0.17s)
--- PASS: TestHandDrivenNextOverAGeneratorLowers/the_|_undefined_form,_tested_first (0.13s)
--- PASS: TestNestedIterableCalleeIsExactlyOneBody (0.00s)
--- PASS: TestNestedIterableCalleeIsExactlyOneBody/a_sibling's_direct_call (0.00s)
--- PASS: TestNestedIterableCalleeIsExactlyOneBody/a_virtual_call (0.00s)
--- PASS: TestNestedIterableCalleeIsExactlyOneBody/a_module_function's_call (0.00s)
--- PASS: TestNestedIterableCalleeIsExactlyOneBody/an_optional_call (0.00s)
--- PASS: TestNestedIterableCalleeIsExactlyOneBody/a_bound_closure's_kept_call (0.00s)
--- PASS: TestNestedIterableCalleeIsExactlyOneBody/a_spread (0.00s)
--- PASS: TestNestedIterableCalleeIsExactlyOneBody/a_call_through_a_function_value (0.00s)
--- PASS: TestNestedIterableCalleeIsExactlyOneBody/a_call_through_a_closure_made_for_it (0.00s)
--- PASS: TestNestedIterableCalleeIsExactlyOneBody/the_parent's_call_through_its_binding (0.00s)
--- PASS: TestNestedIterableCalleeLowers (0.00s)
--- PASS: TestNestedIterableCalleeLowers/two_generator_functions_into_one_parameter_(S3c-2) (0.16s)
--- PASS: TestNestedIterableCalleeLowers/a_sibling_calling_the_callee (0.20s)
--- PASS: TestNestedIterableCalleeLowers/an_Iterable<T>_|_undefined_parameter (0.17s)
--- PASS: TestNestedIterableCalleeLowers/nested_callee_view.a (0.24s)
--- PASS: TestNestedIterableCalleeLowers/a_plain_function*_written_in_place_into_a_nested_callee (0.15s)
--- PASS: TestNestedIterableCalleeLowers/an_IterableIterator_parameter (0.21s)
--- PASS: TestNestedIterableCalleeLowers/written_in_place_into_a_nested_callee_(S3b's_row) (0.11s)
--- PASS: TestNestedIterableCalleeLowers/a_module_generator_into_a_nested_callee (0.10s)
--- PASS: TestNestedIterableCalleeLowers/a_plain_nested_function*_into_a_nested_callee (0.12s)
--- PASS: TestNestedIterableCalleeLowers/nested_callee_overloads.a (0.07s)
--- PASS: TestNestedIterableCalleeLowers/nested_callee_generic.a (0.11s)
--- PASS: TestNestedIterableCalleeLowers/nested_callee_grandchild.a (0.11s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers (0.00s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/a_module_function (0.07s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/an_IterableIterator<T>_|_undefined_parameter (0.09s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/a_nested_block's_own_name (0.08s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/a_nullable_parameter_tested_===_undefined,_returning_early (0.13s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/firstOrUndefinedIterator_beside_a_Map's_values() (0.14s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/tsc's_firstOrUndefinedIterator (0.12s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/a_type-only_typeof (0.11s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/an_outer_loop_over_the_for-of (0.11s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/a_generator_beside_a_collection's_iterator (0.10s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/an_Iterable<string>_parameter_beside_a_Set's_keys() (0.11s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/two_generator_functions,_a_TReturn_of_undefined_and_of_void_(S3c-2) (0.09s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/an_Iterable<string>_parameter (0.14s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/a_generic_instance (0.07s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/one_function_at_two_call_sites (0.10s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/two_generator_functions_(S3c-2) (0.11s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/a_void_TReturn (0.08s)
--- PASS: TestGeneratorIntoAnIterableParameterLowers/an_IterableIterator_parameter (0.09s)
--- PASS: TestNestedIterableCalleeRoadsAreNotYet (0.00s)
--- PASS: TestNestedIterableCalleeRoadsAreNotYet/road_1:_the_callee_also_passed_as_a_callback (0.11s)
--- PASS: TestNestedIterableCalleeRoadsAreNotYet/a_closure_capturing_the_parameter (0.06s)
--- PASS: TestNestedIterableCalleeRoadsAreNotYet/road_3:_a_local_of_the_callee's_name (0.10s)
--- PASS: TestNestedIterableCalleeRoadsAreNotYet/a_call_through_a_local_holding_the_callee (0.12s)
--- PASS: TestNestedIterableCalleeRoadsAreNotYet/road_2:_the_callee_recursing_with_its_parameter (0.09s)
--- PASS: TestNestedIterableCalleeRoadsAreNotYet/road_1:_the_callee_also_held_in_a_local (0.09s)
--- PASS: TestHandDrivenNextStopsOnEveryOtherRoad (0.01s)
--- PASS: TestHandDrivenNextStopsOnEveryOtherRoad/next(x) (0.10s)
--- PASS: TestHandDrivenNextStopsOnEveryOtherRoad/a_for-of_beside_a_next_in_.ts (0.08s)
--- PASS: TestHandDrivenNextStopsOnEveryOtherRoad/return() (0.08s)
--- PASS: TestHandDrivenNextStopsOnEveryOtherRoad/value_in_the_done_arm (0.09s)
--- PASS: TestHandDrivenNextStopsOnEveryOtherRoad/value_unnarrowed (0.09s)
--- PASS: TestHandDrivenNextStopsOnEveryOtherRoad/a_detached_next (0.08s)
--- PASS: TestHandDrivenNextStopsOnEveryOtherRoad/a_for-of_beside_a_next_in_.a (0.12s)
--- PASS: TestHandDrivenNextStopsOnEveryOtherRoad/an_optional_call_of_next (0.07s)
--- PASS: TestHandDrivenNextStopsOnEveryOtherRoad/an_arrow's_parameter,_its_callers_unseen (0.09s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad (0.01s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/typeof_it (0.07s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/spreading_it (0.06s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_=_g2()_before_the_loop (0.10s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_for-of_head_rebinding_it (0.08s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/destructuring_it (0.09s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_destructuring_assignment_rebinding_it (0.14s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_completion_read_through_two_generator_functions_whose_TReturn_differs_(S3c-2) (0.09s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_through_the_parameter (0.09s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/next_inside_the_loop (0.10s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/yields_held_otherwise (0.08s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_number_|_undefined_(S3c-2) (0.06s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_g2() (0.07s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/arguments (0.08s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/two_generator_functions_yielding_a_number_and_a_string_(S3c-2) (0.10s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_a_generator_into_Iterable<string> (0.09s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_string_beside_two_generator_functions_into_Iterable<string>_(S3c-2) (0.10s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_method (0.06s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/an_async_callee (0.09s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_call_through_a_function_value (0.09s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_stored_generator_passed_in (0.10s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_closure_capturing_it,_never_called (0.09s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/reassigning_it (0.11s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/passing_it_on (0.09s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_default_naming_it (0.08s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/it_===_other (0.09s)
--- PASS: TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad/a_read_after_a_labeled_break (0.09s)
```

### 2b: PASS


```sh
go test ./internal/native -timeout 30m -run '^(TestReuseTakesAnOperationNoReviewNamesAsTouchingTheGlobal|TestReuseFollowsAGeneratorStepIntoItsBody|TestReuseFollowsEveryBodyAGeneratorStepMayRun)$' -count=1 -v
```

Exit 0; wall 8.831 seconds; log `/tmp/adamic-gate/step2b.log`.

```text
--- PASS: TestReuseFollowsEveryBodyAGeneratorStepMayRun (0.00s)
--- PASS: TestReuseFollowsAGeneratorStepIntoItsBody (0.00s)
--- PASS: TestReuseTakesAnOperationNoReviewNamesAsTouchingTheGlobal (0.00s)
```

### 2c: PASS


```sh
go test ./internal/worker -timeout 30m -run '^(TestCheckPureRefusesAnOperationNoReviewNames|TestCheckPureInspectsAGeneratorStepsBody|TestCheckPureInspectsEveryBodyAGeneratorStepMayRun)$' -count=1 -v
```

Exit 0; wall 9.434 seconds; log `/tmp/adamic-gate/step2c.log`.

```text
--- PASS: TestCheckPureRefusesAnOperationNoReviewNames (0.00s)
--- PASS: TestCheckPureInspectsEveryBodyAGeneratorStepMayRun (0.00s)
--- PASS: TestCheckPureInspectsAGeneratorStepsBody (0.00s)
```

### 2d: PASS


```sh
go test ./internal/oracle -timeout 60m -run '^(TestGeneratorsSeveralAgreeWithNode|TestIterableViewsAgreeWithNode|TestGeneratorsAgreeWithNode)$' -count=1 -v
```

Exit 0; wall 335.881 seconds; log `/tmp/adamic-gate/step2d.log`.

```text
--- PASS: TestGeneratorsSeveralAgreeWithNode (0.00s)
--- PASS: TestGeneratorsSeveralAgreeWithNode/held_across_yield.a (267.07s)
--- PASS: TestGeneratorsSeveralAgreeWithNode/elaborate_four.a (267.07s)
--- PASS: TestGeneratorsSeveralAgreeWithNode/second_throws.a (267.09s)
--- PASS: TestIterableViewsAgreeWithNode (0.00s)
--- PASS: TestIterableViewsAgreeWithNode/hand_next.a (267.08s)
--- PASS: TestIterableViewsAgreeWithNode/view_values.a (267.09s)
--- PASS: TestIterableViewsAgreeWithNode/view_partial.a (1.97s)
--- PASS: TestIterableViewsAgreeWithNode/view_live.a (2.04s)
--- PASS: TestIterableViewsAgreeWithNode/view_delete.a (2.08s)
--- PASS: TestIterableViewsAgreeWithNode/view_entries.a (2.09s)
--- PASS: TestIterableViewsAgreeWithNode/view_union.a (2.13s)
--- PASS: TestIterableViewsAgreeWithNode/nullable_string_keys.a (1.23s)
--- PASS: TestIterableViewsAgreeWithNode/nullable_values.a (1.39s)
--- PASS: TestIterableViewsAgreeWithNode/hand_next_mixed.a (1.88s)
--- PASS: TestIterableViewsAgreeWithNode/hand_next_maybe_string.a (1.93s)
--- PASS: TestIterableViewsAgreeWithNode/hand_next_generator_throw.a (1.92s)
--- PASS: TestIterableViewsAgreeWithNode/hand_next_generator.a (2.92s)
--- PASS: TestGeneratorsAgreeWithNode (0.00s)
--- PASS: TestGeneratorsAgreeWithNode/array_from.a (1.75s)
--- PASS: TestGeneratorsAgreeWithNode/nested_hoisted.a (1.52s)
--- PASS: TestGeneratorsAgreeWithNode/nested_dead_zone.a (1.67s)
--- PASS: TestGeneratorsAgreeWithNode/nested_early_break.a (2.08s)
--- PASS: TestGeneratorsAgreeWithNode/generic_map.a (1.85s)
--- PASS: TestGeneratorsAgreeWithNode/first_or_undefined.a (1.55s)
--- PASS: TestGeneratorsAgreeWithNode/first_throw.a (1.66s)
--- PASS: TestGeneratorsAgreeWithNode/generic.a (2.30s)
--- PASS: TestGeneratorsAgreeWithNode/nested_checker_shapes.a (2.95s)
--- PASS: TestGeneratorsAgreeWithNode/elementwise_view.a (2.06s)
--- PASS: TestGeneratorsAgreeWithNode/drop_start.a (1.89s)
--- PASS: TestGeneratorsAgreeWithNode/consumer_throw.a (1.65s)
--- PASS: TestGeneratorsAgreeWithNode/early_return.a (2.03s)
--- PASS: TestGeneratorsAgreeWithNode/declared_iterators.a (2.97s)
--- PASS: TestGeneratorsAgreeWithNode/complete.a (2.17s)
--- PASS: TestGeneratorsAgreeWithNode/body_throw.a (1.93s)
--- PASS: TestGeneratorsAgreeWithNode/break_map.a (2.35s)
--- PASS: TestGeneratorsAgreeWithNode/builder_state_strings.a (2.48s)
--- PASS: TestGeneratorsAgreeWithNode/array_from_view.a (1.72s)
--- PASS: TestGeneratorsAgreeWithNode/map_iterator.a (1.97s)
--- PASS: TestGeneratorsAgreeWithNode/nested_carrier_temporary.a (2.11s)
--- PASS: TestGeneratorsAgreeWithNode/nested_callee_overloads.a (1.73s)
--- PASS: TestGeneratorsAgreeWithNode/nested_callee_view.a (2.45s)
--- PASS: TestGeneratorsAgreeWithNode/nested_callee_grandchild.a (1.90s)
--- PASS: TestGeneratorsAgreeWithNode/nested_callee_generic.a (1.93s)
--- PASS: TestGeneratorsAgreeWithNode/map_iterator_entries.a (1.68s)
--- PASS: TestGeneratorsAgreeWithNode/stored_for_of.a (1.72s)
--- PASS: TestGeneratorsAgreeWithNode/stored_exits_finally.a (1.70s)
--- PASS: TestGeneratorsAgreeWithNode/nested_iterable.a (2.37s)
--- PASS: TestGeneratorsAgreeWithNode/split.a (1.36s)
--- PASS: TestGeneratorsAgreeWithNode/spread.a (1.57s)
--- PASS: TestGeneratorsAgreeWithNode/stored_exits.a (1.87s)
--- PASS: TestGeneratorsAgreeWithNode/return_from_loop.a (1.49s)
--- PASS: TestGeneratorsAgreeWithNode/retained_param.a (1.25s)
--- PASS: TestGeneratorsAgreeWithNode/nullish_view_tests.a (1.41s)
--- PASS: TestGeneratorsAgreeWithNode/new_set_plain.a (1.44s)
--- PASS: TestGeneratorsAgreeWithNode/new_set.a (1.41s)
--- PASS: TestGeneratorsAgreeWithNode/new_map_throw.a (1.27s)
--- PASS: TestGeneratorsAgreeWithNode/new_map_plain.a (1.39s)
--- PASS: TestGeneratorsAgreeWithNode/generic_throw.a (1.45s)
--- PASS: TestGeneratorsAgreeWithNode/new_map_entries.a (1.62s)
--- PASS: TestGeneratorsAgreeWithNode/new_map.a (1.67s)
--- PASS: TestGeneratorsAgreeWithNode/map_defined_iterator.a (1.60s)
--- PASS: TestGeneratorsAgreeWithNode/nested_stack_parent.a (1.31s)
--- PASS: TestGeneratorsAgreeWithNode/labeled_exit.a (1.43s)
--- PASS: TestGeneratorsAgreeWithNode/in_place_loop.a (1.65s)
--- PASS: TestGeneratorsAgreeWithNode/in_place_environment.a (1.60s)
--- PASS: TestGeneratorsAgreeWithNode/in_place_forms.a (1.83s)
--- PASS: TestGeneratorsAgreeWithNode/view_body_throw.a (1.52s)
--- PASS: TestGeneratorsAgreeWithNode/view_throw.a (1.67s)
--- PASS: TestGeneratorsAgreeWithNode/view_early_return.a (1.73s)
--- PASS: TestGeneratorsAgreeWithNode/in_place_capture.a (2.23s)
--- PASS: TestGeneratorsAgreeWithNode/view_mixed.a (2.05s)
--- PASS: TestGeneratorsAgreeWithNode/view_outer_loop.a (1.34s)
--- PASS: TestGeneratorsAgreeWithNode/view_string_items.a (1.60s)
--- PASS: TestGeneratorsAgreeWithNode/view_counted_items.a (1.88s)
--- PASS: TestGeneratorsAgreeWithNode/uncaught.a (1.52s)
--- PASS: TestGeneratorsAgreeWithNode/unused_class.a (1.62s)
--- PASS: TestGeneratorsAgreeWithNode/uncaught_spread.a (1.75s)
--- PASS: TestGeneratorsAgreeWithNode/uncaught_from.a (1.70s)
--- PASS: TestGeneratorsAgreeWithNode/stored_later_throw.a (1.50s)
--- PASS: TestGeneratorsAgreeWithNode/stored_untaken.a (1.79s)
--- PASS: TestGeneratorsAgreeWithNode/stored_jsx_children.a (1.56s)
--- PASS: TestGeneratorsAgreeWithNode/nested_live_capture.a (34.33s)
```

## Step 4 — IR/flow/fresh/split/statecopy


```sh
go test ./internal/ir ./internal/flow ./internal/fresh ./internal/split ./stage3/census/latent/statecopy -count=1 -timeout 30m
```

Exit 1; wall 718.021 seconds. All 65 failures are missing fixture-product prerequisite failures, not compile diagnostics. Complete output, including every failure:

```text
ok  	github.com/system-inc/adamic/internal/ir	569.786s
ok  	github.com/system-inc/adamic/internal/flow	689.128s
ok  	github.com/system-inc/adamic/internal/fresh	237.673s
--- FAIL: TestProduct_SplitOracleFixtures (7.47s)
    oracle_fixtures_test.go:48: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_035 (0.21s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_063 (0.15s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_062 (0.10s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_061 (0.15s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_060 (0.15s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_059 (0.10s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_058 (0.06s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_057 (0.12s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_056 (0.06s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_055 (0.08s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_054 (0.09s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_053 (0.10s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_052 (0.04s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_051 (0.06s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_050 (0.03s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_049 (0.06s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_048 (0.06s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_047 (0.07s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_046 (0.06s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_045 (0.12s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_044 (0.12s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_043 (0.06s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_042 (0.04s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_041 (0.05s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_040 (0.03s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_039 (0.09s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_038 (0.06s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_037 (0.05s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_036 (0.07s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_017 (0.12s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_034 (0.07s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_033 (0.09s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_032 (0.08s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_031 (0.09s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_030 (0.05s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_029 (0.06s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_028 (0.04s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_027 (0.06s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_026 (0.07s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_025 (0.09s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_024 (0.05s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_023 (0.02s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_022 (0.04s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_021 (0.05s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_020 (0.02s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_019 (0.01s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_018 (0.02s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_008 (0.06s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_016 (0.04s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_015 (0.05s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_014 (0.05s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_013 (0.05s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_012 (0.06s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_011 (0.19s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_010 (0.07s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_009 (0.12s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_003 (0.08s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_007 (0.05s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_006 (0.10s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_005 (0.06s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_004 (0.03s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_001 (0.06s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_002 (0.08s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
--- FAIL: TestCodegenSnapshot_000 (0.09s)
    codegen_test.go:28: build oracle compiling fixtures: run internal/oracle TestProduct_OracleCompilingFixtures ahead
FAIL
FAIL	github.com/system-inc/adamic/internal/split	616.813s
ok  	github.com/system-inc/adamic/stage3/census/latent/statecopy	21.903s
FAIL
```

## Step 5 — complete lower/worker suite


```sh
go test ./internal/lower ./internal/worker -count=1 -timeout 90m
```

Combined stdout/stderr: `/tmp/lower.log`. Exit 1; wall 1884.196 seconds. Lower passed in 1877.255 seconds; worker failed in 649.508 seconds. Worker had 123 top-level failures and 54 failing subtests, plus two failing child tests captured by the planted-case checks. Every failure was environmental: missing worker products or missing WASI SDK / WASI_SYSROOT. The two planted-case checks failed because their child corpus products were absent; they did not reach the intended native disagreement.

### Every failing test and first error line

- **environment** `TestProduct_WorkerComputeCorpus`: `compute_corpus_test.go:80: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerComputeSieveMutant`: `compute_corpus_test.go:83: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_000`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_000/decode/json_decode_coverage_unions.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_001`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_001/decode/json_decode_coverage_tags.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_002`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_002/decode/json_decode_coverage_depth_extra.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_002/decode/json_decode_coverage_object_scalar.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_002/decode/json_decode_depth.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_003`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_003/decode/json_decode_coverage_calls.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_003/encode/json_encode_last_use.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_006`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_006/decode/json_decode_coverage_arrays.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_006/decode/json_decode_coverage_shapes.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_007`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_007/decode/json_decode_coverage_nested.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_007/encode/json_encode_differences.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_008`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_008/decode/json_decode_supplement.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_010`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_010/decode/json_decode_coverage_unicode_tag.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_010/decode/json_decode_grammar.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_011`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_011/decode/json_decode_empty_tuple.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_011/decode/json_decode_objects.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_014`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_014/encode/json_encode_depth.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_015`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_015/encode/json_encode_unions.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_016`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_016/decode/fast_risks.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_018`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_018/encode/json_encode_method.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_018/encode/json_encode_own_data.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_019`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_019/decode/json_decode_coverage_grammar.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_020`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_020/decode/corpus`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_022`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_022/encode/json_encode_scalars.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_023`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_023/encode/json_encode_objects.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_026`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_026/decode/json_decode_scalars.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_028`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_028/encode/corpus`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_029`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerJSON_029/decode/json_decode_coverage_layouts.a`: `json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first`
- **environment** `TestProduct_WorkerWasm_000`: `products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasm_001`: `products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasm_002`: `products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasm_003`: `products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasm_004`: `products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasm_005`: `products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasm_006`: `products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasm_007`: `products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasmMutants_000`: `products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasmMutants_001`: `products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasmMutants_002`: `products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasmMutants_003`: `products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasmMutants_004`: `products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasmMutants_005`: `products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasmMutants_006`: `products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasmMutants_007`: `products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasmMutants_008`: `products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestProduct_WorkerWasmTransform`: `products_test.go:170: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWorkerComputeCorpus`: `compute_corpus_test.go:100: Worker product worker compute corpus sieve-mutant=false not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_016`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_grammar.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_016/json_decode_coverage_grammar.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_grammar.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_003`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_depth.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_003/json_encode_depth.a`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_depth.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_002`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_000`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_017`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_object_scalar.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_017/json_decode_coverage_object_scalar.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_object_scalar.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_017/json_decode_coverage_unicode_tag.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_unicode_tag.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_030`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_029`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_031`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_supplement.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_031/json_decode_supplement.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_supplement.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_028`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_026`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_objects.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_026/json_decode_objects.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_objects.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_027`: `json_decode_test.go:96: Worker product worker-json-decode/fast_risks.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_027/fast_risks.a`: `json_decode_test.go:96: Worker product worker-json-decode/fast_risks.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_021`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_023`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_depth.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_023/json_decode_depth.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_depth.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_023/json_decode_scalars.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_scalars.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_019`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_024`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_025`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_depth_extra.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_025/json_decode_coverage_depth_extra.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_depth_extra.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_018`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_031`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_029`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_030`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_028`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_027`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_026`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_024`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_023`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_013`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_025`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_021`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_017`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_018`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_015`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_019`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_scalars.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_019/json_encode_scalars.a`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_scalars.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_014`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_differences.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_014/json_encode_differences.a`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_differences.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_014`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_015`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_013`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_012`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWasmWorkers_003`: `wasm_test.go:78: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWorkerDecodeJSON_010`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_009`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_unions.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_009/json_decode_coverage_unions.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_unions.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_011`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_arrays.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_011/json_decode_coverage_arrays.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_arrays.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_011/json_decode_coverage_shapes.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_shapes.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_008`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_calls.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_008/json_decode_coverage_calls.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_calls.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWasmPurity`: `wasm_test.go:150: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmMultipleFunctions_000`: `wasm_test.go:195: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMemory_000`: `wasm_test.go:132: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkers_004`: `wasm_test.go:78: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWorkerEncodeJSON_008`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_011`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_010`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWasmWorkerMutants_002`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkers_002`: `wasm_test.go:78: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkers_001`: `wasm_test.go:78: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkers_000`: `wasm_test.go:78: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMutants_015`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMutants_014`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMutants_013`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMutants_012`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMutants_011`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMutants_010`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMutants_009`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMutants_008`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMutants_007`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMutants_006`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMutants_005`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMutants_004`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMutants_003`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWorkerEncodeJSON_012`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_unions.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_012/json_encode_unions.a`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_unions.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_009`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_006`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_objects.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_006/json_encode_objects.a`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_objects.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_005`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_method.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_005/json_encode_method.a`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_method.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_005/json_encode_own_data.a`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_own_data.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWasmWorkerMutants_001`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWasmWorkerMutants_000`: `wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT`
- **environment** `TestWorkerEncodeJSON_004`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_last_use.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSON_004/json_encode_last_use.a`: `json_encode_test.go:94: Worker product worker-json-encode/json_encode_last_use.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerEncodeJSONFamiliesCatchAPlantedCase`: `json_families_test.go:116: planted wrong case must fail exactly TestWorkerEncodeJSON_014/corpus/00000 with native disagreement: exit status 1`
- **environment** `captured child TestWorkerEncodeJSON_014`: `json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSONFamiliesCatchAPlantedCase`: `json_families_test.go:112: planted wrong case must fail exactly TestWorkerDecodeJSON_014/corpus/00000 with native disagreement: exit status 1`
- **environment** `captured child TestWorkerDecodeJSON_014`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_003`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_grammar.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_003/json_decode_grammar.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_grammar.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_006`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_nested.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_006/json_decode_coverage_nested.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_nested.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_004`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_001`: `json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_002`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_layouts.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_002/json_decode_coverage_layouts.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_layouts.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_002/json_decode_empty_tuple.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_empty_tuple.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerComputeSieveMutant`: `compute_corpus_test.go:101: Worker product worker compute corpus sieve-mutant=true not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_000`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_tags.a not built ahead; run TestProduct_Worker tests first`
- **environment** `TestWorkerDecodeJSON_000/json_decode_coverage_tags.a`: `json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_tags.a not built ahead; run TestProduct_Worker tests first`

### Complete suite output


```text
ok  	github.com/system-inc/adamic/internal/lower	1877.255s
--- FAIL: TestProduct_WorkerComputeCorpus (3.08s)
    compute_corpus_test.go:80: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerComputeSieveMutant (0.03s)
    compute_corpus_test.go:83: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_000 (0.05s)
    --- FAIL: TestProduct_WorkerJSON_000/decode/json_decode_coverage_unions.a (0.05s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_001 (0.03s)
    --- FAIL: TestProduct_WorkerJSON_001/decode/json_decode_coverage_tags.a (0.03s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_002 (0.10s)
    --- FAIL: TestProduct_WorkerJSON_002/decode/json_decode_coverage_depth_extra.a (0.03s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
    --- FAIL: TestProduct_WorkerJSON_002/decode/json_decode_coverage_object_scalar.a (0.03s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
    --- FAIL: TestProduct_WorkerJSON_002/decode/json_decode_depth.a (0.03s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_003 (0.07s)
    --- FAIL: TestProduct_WorkerJSON_003/decode/json_decode_coverage_calls.a (0.04s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
    --- FAIL: TestProduct_WorkerJSON_003/encode/json_encode_last_use.a (0.03s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_006 (0.07s)
    --- FAIL: TestProduct_WorkerJSON_006/decode/json_decode_coverage_arrays.a (0.03s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
    --- FAIL: TestProduct_WorkerJSON_006/decode/json_decode_coverage_shapes.a (0.04s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_007 (0.06s)
    --- FAIL: TestProduct_WorkerJSON_007/decode/json_decode_coverage_nested.a (0.03s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
    --- FAIL: TestProduct_WorkerJSON_007/encode/json_encode_differences.a (0.03s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_008 (0.04s)
    --- FAIL: TestProduct_WorkerJSON_008/decode/json_decode_supplement.a (0.03s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_010 (0.07s)
    --- FAIL: TestProduct_WorkerJSON_010/decode/json_decode_coverage_unicode_tag.a (0.03s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
    --- FAIL: TestProduct_WorkerJSON_010/decode/json_decode_grammar.a (0.04s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_011 (0.07s)
    --- FAIL: TestProduct_WorkerJSON_011/decode/json_decode_empty_tuple.a (0.04s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
    --- FAIL: TestProduct_WorkerJSON_011/decode/json_decode_objects.a (0.03s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_014 (0.04s)
    --- FAIL: TestProduct_WorkerJSON_014/encode/json_encode_depth.a (0.04s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_015 (0.03s)
    --- FAIL: TestProduct_WorkerJSON_015/encode/json_encode_unions.a (0.03s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_016 (0.03s)
    --- FAIL: TestProduct_WorkerJSON_016/decode/fast_risks.a (0.03s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_018 (0.09s)
    --- FAIL: TestProduct_WorkerJSON_018/encode/json_encode_method.a (0.05s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
    --- FAIL: TestProduct_WorkerJSON_018/encode/json_encode_own_data.a (0.04s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_019 (0.06s)
    --- FAIL: TestProduct_WorkerJSON_019/decode/json_decode_coverage_grammar.a (0.05s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_020 (0.04s)
    --- FAIL: TestProduct_WorkerJSON_020/decode/corpus (0.04s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_022 (0.17s)
    --- FAIL: TestProduct_WorkerJSON_022/encode/json_encode_scalars.a (0.17s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_023 (0.12s)
    --- FAIL: TestProduct_WorkerJSON_023/encode/json_encode_objects.a (0.11s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_026 (0.12s)
    --- FAIL: TestProduct_WorkerJSON_026/decode/json_decode_scalars.a (0.10s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_028 (0.10s)
    --- FAIL: TestProduct_WorkerJSON_028/encode/corpus (0.09s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerJSON_029 (0.19s)
    --- FAIL: TestProduct_WorkerJSON_029/decode/json_decode_coverage_layouts.a (0.19s)
        json_products_test.go:109: Worker product go build ./cmd/adamic adamic not built ahead; run TestProduct_Worker tests first
--- FAIL: TestProduct_WorkerWasm_000 (0.00s)
    products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasm_001 (0.00s)
    products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasm_002 (0.00s)
    products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasm_003 (0.00s)
    products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasm_004 (0.00s)
    products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasm_005 (0.00s)
    products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasm_006 (0.00s)
    products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasm_007 (0.00s)
    products_test.go:175: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasmMutants_000 (0.00s)
    products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasmMutants_001 (0.00s)
    products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasmMutants_002 (0.00s)
    products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasmMutants_003 (0.00s)
    products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasmMutants_004 (0.00s)
    products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasmMutants_005 (0.00s)
    products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasmMutants_006 (0.00s)
    products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasmMutants_007 (0.00s)
    products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasmMutants_008 (0.00s)
    products_test.go:211: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestProduct_WorkerWasmTransform (0.00s)
    products_test.go:170: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWorkerComputeCorpus (0.36s)
    compute_corpus_test.go:100: Worker product worker compute corpus sieve-mutant=false not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_016 (0.57s)
    --- FAIL: TestWorkerDecodeJSON_016/json_decode_coverage_grammar.a (0.18s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_grammar.a not built ahead; run TestProduct_Worker tests first
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_003 (0.65s)
    --- FAIL: TestWorkerEncodeJSON_003/json_encode_depth.a (0.24s)
        json_encode_test.go:94: Worker product worker-json-encode/json_encode_depth.a not built ahead; run TestProduct_Worker tests first
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_002 (0.35s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_000 (0.55s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_017 (1.22s)
    --- FAIL: TestWorkerDecodeJSON_017/json_decode_coverage_object_scalar.a (0.21s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_object_scalar.a not built ahead; run TestProduct_Worker tests first
    --- FAIL: TestWorkerDecodeJSON_017/json_decode_coverage_unicode_tag.a (0.24s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_unicode_tag.a not built ahead; run TestProduct_Worker tests first
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_030 (0.61s)
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_029 (0.66s)
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_031 (0.93s)
    --- FAIL: TestWorkerDecodeJSON_031/json_decode_supplement.a (0.28s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_supplement.a not built ahead; run TestProduct_Worker tests first
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_028 (0.55s)
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_026 (0.71s)
    --- FAIL: TestWorkerDecodeJSON_026/json_decode_objects.a (0.27s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_objects.a not built ahead; run TestProduct_Worker tests first
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_027 (0.85s)
    --- FAIL: TestWorkerDecodeJSON_027/fast_risks.a (0.24s)
        json_decode_test.go:96: Worker product worker-json-decode/fast_risks.a not built ahead; run TestProduct_Worker tests first
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_021 (0.40s)
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_023 (0.82s)
    --- FAIL: TestWorkerDecodeJSON_023/json_decode_depth.a (0.23s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_depth.a not built ahead; run TestProduct_Worker tests first
    --- FAIL: TestWorkerDecodeJSON_023/json_decode_scalars.a (0.21s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_scalars.a not built ahead; run TestProduct_Worker tests first
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_019 (0.42s)
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_024 (0.99s)
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_025 (1.24s)
    --- FAIL: TestWorkerDecodeJSON_025/json_decode_coverage_depth_extra.a (0.21s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_depth_extra.a not built ahead; run TestProduct_Worker tests first
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_018 (0.29s)
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_031 (0.33s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_029 (0.41s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_030 (0.42s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_028 (0.48s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_027 (0.43s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_026 (0.47s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_024 (0.80s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_023 (0.64s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_013 (0.63s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_025 (0.82s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_021 (0.49s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_017 (0.48s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_018 (0.50s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_015 (0.51s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_019 (0.75s)
    --- FAIL: TestWorkerEncodeJSON_019/json_encode_scalars.a (0.34s)
        json_encode_test.go:94: Worker product worker-json-encode/json_encode_scalars.a not built ahead; run TestProduct_Worker tests first
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_014 (0.73s)
    --- FAIL: TestWorkerEncodeJSON_014/json_encode_differences.a (0.35s)
        json_encode_test.go:94: Worker product worker-json-encode/json_encode_differences.a not built ahead; run TestProduct_Worker tests first
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_014 (0.31s)
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_015 (0.36s)
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_013 (0.35s)
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_012 (0.36s)
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWasmWorkers_003 (0.00s)
    wasm_test.go:78: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWorkerDecodeJSON_010 (0.36s)
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_009 (0.51s)
    --- FAIL: TestWorkerDecodeJSON_009/json_decode_coverage_unions.a (0.16s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_unions.a not built ahead; run TestProduct_Worker tests first
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_011 (0.68s)
    --- FAIL: TestWorkerDecodeJSON_011/json_decode_coverage_arrays.a (0.24s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_arrays.a not built ahead; run TestProduct_Worker tests first
    --- FAIL: TestWorkerDecodeJSON_011/json_decode_coverage_shapes.a (0.15s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_shapes.a not built ahead; run TestProduct_Worker tests first
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_008 (0.86s)
    --- FAIL: TestWorkerDecodeJSON_008/json_decode_coverage_calls.a (0.23s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_calls.a not built ahead; run TestProduct_Worker tests first
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWasmPurity (0.00s)
    wasm_test.go:150: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmMultipleFunctions_000 (0.00s)
    wasm_test.go:195: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMemory_000 (0.00s)
    wasm_test.go:132: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkers_004 (0.00s)
    wasm_test.go:78: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWorkerEncodeJSON_008 (0.43s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_011 (0.48s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_010 (0.49s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWasmWorkerMutants_002 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkers_002 (0.00s)
    wasm_test.go:78: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkers_001 (0.00s)
    wasm_test.go:78: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkers_000 (0.00s)
    wasm_test.go:78: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMutants_015 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMutants_014 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMutants_013 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMutants_012 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMutants_011 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMutants_010 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMutants_009 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMutants_008 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMutants_007 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMutants_006 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMutants_005 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMutants_004 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMutants_003 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWorkerEncodeJSON_012 (0.77s)
    --- FAIL: TestWorkerEncodeJSON_012/json_encode_unions.a (0.39s)
        json_encode_test.go:94: Worker product worker-json-encode/json_encode_unions.a not built ahead; run TestProduct_Worker tests first
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_009 (0.77s)
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_006 (0.81s)
    --- FAIL: TestWorkerEncodeJSON_006/json_encode_objects.a (0.28s)
        json_encode_test.go:94: Worker product worker-json-encode/json_encode_objects.a not built ahead; run TestProduct_Worker tests first
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSON_005 (1.31s)
    --- FAIL: TestWorkerEncodeJSON_005/json_encode_method.a (0.50s)
        json_encode_test.go:94: Worker product worker-json-encode/json_encode_method.a not built ahead; run TestProduct_Worker tests first
    --- FAIL: TestWorkerEncodeJSON_005/json_encode_own_data.a (0.39s)
        json_encode_test.go:94: Worker product worker-json-encode/json_encode_own_data.a not built ahead; run TestProduct_Worker tests first
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWasmWorkerMutants_001 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWasmWorkerMutants_000 (0.00s)
    wasm_mutants_test.go:14: Workers require setup.sh --wasi-sdk: native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestWorkerEncodeJSON_004 (0.68s)
    --- FAIL: TestWorkerEncodeJSON_004/json_encode_last_use.a (0.22s)
        json_encode_test.go:94: Worker product worker-json-encode/json_encode_last_use.a not built ahead; run TestProduct_Worker tests first
    json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerEncodeJSONFamiliesCatchAPlantedCase (11.10s)
    json_families_test.go:116: planted wrong case must fail exactly TestWorkerEncodeJSON_014/corpus/00000 with native disagreement: exit status 1
        === RUN   TestWorkerEncodeJSON_000
        === PAUSE TestWorkerEncodeJSON_000
        === RUN   TestWorkerEncodeJSON_001
        === PAUSE TestWorkerEncodeJSON_001
        === RUN   TestWorkerEncodeJSON_002
        === PAUSE TestWorkerEncodeJSON_002
        === RUN   TestWorkerEncodeJSON_003
        === PAUSE TestWorkerEncodeJSON_003
        === RUN   TestWorkerEncodeJSON_004
        === PAUSE TestWorkerEncodeJSON_004
        === RUN   TestWorkerEncodeJSON_005
        === PAUSE TestWorkerEncodeJSON_005
        === RUN   TestWorkerEncodeJSON_006
        === PAUSE TestWorkerEncodeJSON_006
        === RUN   TestWorkerEncodeJSON_007
        === PAUSE TestWorkerEncodeJSON_007
        === RUN   TestWorkerEncodeJSON_008
        === PAUSE TestWorkerEncodeJSON_008
        === RUN   TestWorkerEncodeJSON_009
        === PAUSE TestWorkerEncodeJSON_009
        === RUN   TestWorkerEncodeJSON_010
        === PAUSE TestWorkerEncodeJSON_010
        === RUN   TestWorkerEncodeJSON_011
        === PAUSE TestWorkerEncodeJSON_011
        === RUN   TestWorkerEncodeJSON_012
        === PAUSE TestWorkerEncodeJSON_012
        === RUN   TestWorkerEncodeJSON_013
        === PAUSE TestWorkerEncodeJSON_013
        === RUN   TestWorkerEncodeJSON_014
        === PAUSE TestWorkerEncodeJSON_014
        === RUN   TestWorkerEncodeJSON_015
        === PAUSE TestWorkerEncodeJSON_015
        === RUN   TestWorkerEncodeJSON_016
        === PAUSE TestWorkerEncodeJSON_016
        === RUN   TestWorkerEncodeJSON_017
        === PAUSE TestWorkerEncodeJSON_017
        === RUN   TestWorkerEncodeJSON_018
        === PAUSE TestWorkerEncodeJSON_018
        === RUN   TestWorkerEncodeJSON_019
        === PAUSE TestWorkerEncodeJSON_019
        === RUN   TestWorkerEncodeJSON_020
        === PAUSE TestWorkerEncodeJSON_020
        === RUN   TestWorkerEncodeJSON_021
        === PAUSE TestWorkerEncodeJSON_021
        === RUN   TestWorkerEncodeJSON_022
        === PAUSE TestWorkerEncodeJSON_022
        === RUN   TestWorkerEncodeJSON_023
        === PAUSE TestWorkerEncodeJSON_023
        === RUN   TestWorkerEncodeJSON_024
        === PAUSE TestWorkerEncodeJSON_024
        === RUN   TestWorkerEncodeJSON_025
        === PAUSE TestWorkerEncodeJSON_025
        === RUN   TestWorkerEncodeJSON_026
        === PAUSE TestWorkerEncodeJSON_026
        === RUN   TestWorkerEncodeJSON_027
        === PAUSE TestWorkerEncodeJSON_027
        === RUN   TestWorkerEncodeJSON_028
        === PAUSE TestWorkerEncodeJSON_028
        === RUN   TestWorkerEncodeJSON_029
        === PAUSE TestWorkerEncodeJSON_029
        === RUN   TestWorkerEncodeJSON_030
        === PAUSE TestWorkerEncodeJSON_030
        === RUN   TestWorkerEncodeJSON_031
        === PAUSE TestWorkerEncodeJSON_031
        === CONT  TestWorkerEncodeJSON_000
        === CONT  TestWorkerEncodeJSON_016
        --- PASS: TestWorkerEncodeJSON_000 (0.00s)
        === CONT  TestWorkerEncodeJSON_015
        === CONT  TestWorkerEncodeJSON_008
        === CONT  TestWorkerEncodeJSON_004
        --- PASS: TestWorkerEncodeJSON_016 (0.01s)
        === CONT  TestWorkerEncodeJSON_006
        --- PASS: TestWorkerEncodeJSON_006 (0.01s)
        === CONT  TestWorkerEncodeJSON_007
        --- PASS: TestWorkerEncodeJSON_007 (0.00s)
        === CONT  TestWorkerEncodeJSON_005
        --- PASS: TestWorkerEncodeJSON_015 (0.03s)
        === CONT  TestWorkerEncodeJSON_012
        --- PASS: TestWorkerEncodeJSON_008 (0.03s)
        === CONT  TestWorkerEncodeJSON_014
        --- PASS: TestWorkerEncodeJSON_005 (0.01s)
        === CONT  TestWorkerEncodeJSON_013
        --- PASS: TestWorkerEncodeJSON_013 (0.00s)
        === CONT  TestWorkerEncodeJSON_010
        --- PASS: TestWorkerEncodeJSON_004 (0.07s)
        === CONT  TestWorkerEncodeJSON_011
        --- PASS: TestWorkerEncodeJSON_011 (0.00s)
        === CONT  TestWorkerEncodeJSON_002
        --- PASS: TestWorkerEncodeJSON_010 (0.05s)
        === CONT  TestWorkerEncodeJSON_003
        --- PASS: TestWorkerEncodeJSON_002 (0.03s)
        === CONT  TestWorkerEncodeJSON_009
        --- PASS: TestWorkerEncodeJSON_012 (0.09s)
        === CONT  TestWorkerEncodeJSON_001
        --- PASS: TestWorkerEncodeJSON_003 (0.03s)
        === CONT  TestWorkerEncodeJSON_024
        --- PASS: TestWorkerEncodeJSON_009 (0.03s)
        === CONT  TestWorkerEncodeJSON_031
        --- PASS: TestWorkerEncodeJSON_024 (0.01s)
        === CONT  TestWorkerEncodeJSON_030
        --- PASS: TestWorkerEncodeJSON_001 (0.02s)
        === CONT  TestWorkerEncodeJSON_029
        --- PASS: TestWorkerEncodeJSON_031 (0.01s)
        === CONT  TestWorkerEncodeJSON_028
        --- PASS: TestWorkerEncodeJSON_028 (0.00s)
        === CONT  TestWorkerEncodeJSON_027
        --- PASS: TestWorkerEncodeJSON_030 (0.05s)
        === CONT  TestWorkerEncodeJSON_026
        --- PASS: TestWorkerEncodeJSON_029 (0.05s)
        === CONT  TestWorkerEncodeJSON_025
        --- PASS: TestWorkerEncodeJSON_026 (0.00s)
        === CONT  TestWorkerEncodeJSON_020
        --- PASS: TestWorkerEncodeJSON_020 (0.00s)
        === CONT  TestWorkerEncodeJSON_023
        --- PASS: TestWorkerEncodeJSON_025 (0.00s)
        === CONT  TestWorkerEncodeJSON_022
        --- PASS: TestWorkerEncodeJSON_022 (0.01s)
        === CONT  TestWorkerEncodeJSON_021
        --- PASS: TestWorkerEncodeJSON_021 (0.00s)
        === CONT  TestWorkerEncodeJSON_018
        --- PASS: TestWorkerEncodeJSON_018 (0.00s)
        === CONT  TestWorkerEncodeJSON_019
        --- PASS: TestWorkerEncodeJSON_027 (0.06s)
        === CONT  TestWorkerEncodeJSON_017
        --- PASS: TestWorkerEncodeJSON_017 (0.00s)
        --- PASS: TestWorkerEncodeJSON_019 (0.01s)
        --- PASS: TestWorkerEncodeJSON_023 (0.03s)
        === NAME  TestWorkerEncodeJSON_014
            json_encode_test.go:111: Worker product worker-json-encode/corpus not built ahead; run TestProduct_Worker tests first
        --- FAIL: TestWorkerEncodeJSON_014 (10.91s)
        FAIL
--- FAIL: TestWorkerDecodeJSONFamiliesCatchAPlantedCase (12.07s)
    json_families_test.go:112: planted wrong case must fail exactly TestWorkerDecodeJSON_014/corpus/00000 with native disagreement: exit status 1
        === RUN   TestWorkerDecodeJSON_000
        === PAUSE TestWorkerDecodeJSON_000
        === RUN   TestWorkerDecodeJSON_001
        === PAUSE TestWorkerDecodeJSON_001
        === RUN   TestWorkerDecodeJSON_002
        === PAUSE TestWorkerDecodeJSON_002
        === RUN   TestWorkerDecodeJSON_003
        === PAUSE TestWorkerDecodeJSON_003
        === RUN   TestWorkerDecodeJSON_004
        === PAUSE TestWorkerDecodeJSON_004
        === RUN   TestWorkerDecodeJSON_005
        === PAUSE TestWorkerDecodeJSON_005
        === RUN   TestWorkerDecodeJSON_006
        === PAUSE TestWorkerDecodeJSON_006
        === RUN   TestWorkerDecodeJSON_007
        === PAUSE TestWorkerDecodeJSON_007
        === RUN   TestWorkerDecodeJSON_008
        === PAUSE TestWorkerDecodeJSON_008
        === RUN   TestWorkerDecodeJSON_009
        === PAUSE TestWorkerDecodeJSON_009
        === RUN   TestWorkerDecodeJSON_010
        === PAUSE TestWorkerDecodeJSON_010
        === RUN   TestWorkerDecodeJSON_011
        === PAUSE TestWorkerDecodeJSON_011
        === RUN   TestWorkerDecodeJSON_012
        === PAUSE TestWorkerDecodeJSON_012
        === RUN   TestWorkerDecodeJSON_013
        === PAUSE TestWorkerDecodeJSON_013
        === RUN   TestWorkerDecodeJSON_014
        === PAUSE TestWorkerDecodeJSON_014
        === RUN   TestWorkerDecodeJSON_015
        === PAUSE TestWorkerDecodeJSON_015
        === RUN   TestWorkerDecodeJSON_016
        === PAUSE TestWorkerDecodeJSON_016
        === RUN   TestWorkerDecodeJSON_017
        === PAUSE TestWorkerDecodeJSON_017
        === RUN   TestWorkerDecodeJSON_018
        === PAUSE TestWorkerDecodeJSON_018
        === RUN   TestWorkerDecodeJSON_019
        === PAUSE TestWorkerDecodeJSON_019
        === RUN   TestWorkerDecodeJSON_020
        === PAUSE TestWorkerDecodeJSON_020
        === RUN   TestWorkerDecodeJSON_021
        === PAUSE TestWorkerDecodeJSON_021
        === RUN   TestWorkerDecodeJSON_022
        === PAUSE TestWorkerDecodeJSON_022
        === RUN   TestWorkerDecodeJSON_023
        === PAUSE TestWorkerDecodeJSON_023
        === RUN   TestWorkerDecodeJSON_024
        === PAUSE TestWorkerDecodeJSON_024
        === RUN   TestWorkerDecodeJSON_025
        === PAUSE TestWorkerDecodeJSON_025
        === RUN   TestWorkerDecodeJSON_026
        === PAUSE TestWorkerDecodeJSON_026
        === RUN   TestWorkerDecodeJSON_027
        === PAUSE TestWorkerDecodeJSON_027
        === RUN   TestWorkerDecodeJSON_028
        === PAUSE TestWorkerDecodeJSON_028
        === RUN   TestWorkerDecodeJSON_029
        === PAUSE TestWorkerDecodeJSON_029
        === RUN   TestWorkerDecodeJSON_030
        === PAUSE TestWorkerDecodeJSON_030
        === RUN   TestWorkerDecodeJSON_031
        === PAUSE TestWorkerDecodeJSON_031
        === CONT  TestWorkerDecodeJSON_000
        === CONT  TestWorkerDecodeJSON_016
        === CONT  TestWorkerDecodeJSON_008
        --- PASS: TestWorkerDecodeJSON_016 (0.00s)
        === CONT  TestWorkerDecodeJSON_031
        === CONT  TestWorkerDecodeJSON_007
        --- PASS: TestWorkerDecodeJSON_007 (0.01s)
        === CONT  TestWorkerDecodeJSON_023
        --- PASS: TestWorkerDecodeJSON_023 (0.01s)
        === CONT  TestWorkerDecodeJSON_022
        --- PASS: TestWorkerDecodeJSON_031 (0.03s)
        === CONT  TestWorkerDecodeJSON_021
        --- PASS: TestWorkerDecodeJSON_008 (0.04s)
        === CONT  TestWorkerDecodeJSON_020
        --- PASS: TestWorkerDecodeJSON_020 (0.00s)
        === CONT  TestWorkerDecodeJSON_019
        --- PASS: TestWorkerDecodeJSON_022 (0.02s)
        === CONT  TestWorkerDecodeJSON_018
        --- PASS: TestWorkerDecodeJSON_019 (0.03s)
        === CONT  TestWorkerDecodeJSON_017
        --- PASS: TestWorkerDecodeJSON_018 (0.03s)
        === CONT  TestWorkerDecodeJSON_024
        --- PASS: TestWorkerDecodeJSON_017 (0.00s)
        === CONT  TestWorkerDecodeJSON_030
        --- PASS: TestWorkerDecodeJSON_000 (0.07s)
        === CONT  TestWorkerDecodeJSON_029
        --- PASS: TestWorkerDecodeJSON_021 (0.05s)
        === CONT  TestWorkerDecodeJSON_028
        --- PASS: TestWorkerDecodeJSON_030 (0.01s)
        === CONT  TestWorkerDecodeJSON_027
        --- PASS: TestWorkerDecodeJSON_024 (0.02s)
        === CONT  TestWorkerDecodeJSON_026
        --- PASS: TestWorkerDecodeJSON_027 (0.02s)
        === CONT  TestWorkerDecodeJSON_025
        --- PASS: TestWorkerDecodeJSON_028 (0.04s)
        === CONT  TestWorkerDecodeJSON_004
        --- PASS: TestWorkerDecodeJSON_025 (0.02s)
        === CONT  TestWorkerDecodeJSON_005
        === CONT  TestWorkerDecodeJSON_006
        --- PASS: TestWorkerDecodeJSON_026 (0.03s)
        --- PASS: TestWorkerDecodeJSON_006 (0.02s)
        === CONT  TestWorkerDecodeJSON_002
        --- PASS: TestWorkerDecodeJSON_005 (0.03s)
        === CONT  TestWorkerDecodeJSON_003
        --- PASS: TestWorkerDecodeJSON_002 (0.01s)
        === CONT  TestWorkerDecodeJSON_014
        --- PASS: TestWorkerDecodeJSON_004 (0.04s)
        --- PASS: TestWorkerDecodeJSON_029 (0.08s)
        === CONT  TestWorkerDecodeJSON_012
        === CONT  TestWorkerDecodeJSON_015
        --- PASS: TestWorkerDecodeJSON_003 (0.04s)
        === CONT  TestWorkerDecodeJSON_013
        --- PASS: TestWorkerDecodeJSON_013 (0.00s)
        === CONT  TestWorkerDecodeJSON_001
        --- PASS: TestWorkerDecodeJSON_001 (0.00s)
        === CONT  TestWorkerDecodeJSON_010
        --- PASS: TestWorkerDecodeJSON_010 (0.00s)
        === CONT  TestWorkerDecodeJSON_011
        --- PASS: TestWorkerDecodeJSON_011 (0.02s)
        === CONT  TestWorkerDecodeJSON_009
        --- PASS: TestWorkerDecodeJSON_015 (0.05s)
        --- PASS: TestWorkerDecodeJSON_012 (0.06s)
        --- PASS: TestWorkerDecodeJSON_009 (0.02s)
        === NAME  TestWorkerDecodeJSON_014
            json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
        --- FAIL: TestWorkerDecodeJSON_014 (11.85s)
        FAIL
--- FAIL: TestWorkerDecodeJSON_003 (0.71s)
    --- FAIL: TestWorkerDecodeJSON_003/json_decode_grammar.a (0.33s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_grammar.a not built ahead; run TestProduct_Worker tests first
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_006 (0.65s)
    --- FAIL: TestWorkerDecodeJSON_006/json_decode_coverage_nested.a (0.20s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_nested.a not built ahead; run TestProduct_Worker tests first
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_004 (0.46s)
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_001 (0.52s)
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_002 (1.07s)
    --- FAIL: TestWorkerDecodeJSON_002/json_decode_coverage_layouts.a (0.43s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_layouts.a not built ahead; run TestProduct_Worker tests first
    --- FAIL: TestWorkerDecodeJSON_002/json_decode_empty_tuple.a (0.19s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_empty_tuple.a not built ahead; run TestProduct_Worker tests first
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerComputeSieveMutant (0.24s)
    compute_corpus_test.go:101: Worker product worker compute corpus sieve-mutant=true not built ahead; run TestProduct_Worker tests first
--- FAIL: TestWorkerDecodeJSON_000 (0.68s)
    --- FAIL: TestWorkerDecodeJSON_000/json_decode_coverage_tags.a (0.17s)
        json_decode_test.go:96: Worker product worker-json-decode/json_decode_coverage_tags.a not built ahead; run TestProduct_Worker tests first
    json_decode_test.go:131: Worker product worker-json-decode/corpus not built ahead; run TestProduct_Worker tests first
FAIL
FAIL	github.com/system-inc/adamic/internal/worker	649.508s
FAIL
```

## Environment and execution controls

Loaded `/workspace/adamic-tools/env.sh` first. Created `/tmp/adamic-gate` using `mkdir -p` outside the repository. Exported `GOFLAGS=-buildvcs=false`. Upstream TypeScript was fetched into `/tmp/adamic-gate/TypeScript`, checked out detached at `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported as `ADAMIC_TYPESCRIPT_SOURCE`. Copied stage3/api package.json and package-lock.json to `/tmp/adamic-gate/api` and installed with `npm ci`; its `.bin` directory was on PATH and `tsc --version` returned `Version 6.0.3`. No packages were installed into the invoking repository.
Exported `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true`. The admission gate sanitizes tool environments, dropping these variables, so external wrappers under `/tmp/adamic-gate/admission-tools/bin` re-exported all three before executing the existing git/npm/go/node/python/clang tools. The gate used that external ADAMIC_TOOLS/PATH; no source change was made for environment propagation.
One heavy command ran at a time. Admission ran with no other build or test command running, with `--workers 2`. All helper scripts, logs, package installations, and upstream checkouts were outside the invoking repository. Only this report is intended for the proof branch commit.
