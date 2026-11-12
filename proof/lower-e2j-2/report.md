**Compiler proof — task #e2jp7my, lower-e2j-2**

PASS at `a114e99ef3d55024743d8c49bf62324224780358`. The requested unmodified checks all passed. The single requested mutant was caught by three subtests, then restored. Admission delta passed with no newly admitted or newly refused programs, no classification-stderr differences, and no timeouts.

Base: `49d8529950ce6330d011a118eae4d2a80bc4d14d`. HEAD is exactly its one-commit child, fetched from `origin/compiler/tsc-version-zero-agrees-2` and checked out detached. This proof tests the inherited #39wtqdj empty-build fix as well as HEAD’s updated pins; it does not attribute that earlier production fix to HEAD. HEAD changes tests, mutant documentation, and the version-zero fixture, with no production lowering changes.

Executed on 2026-11-12 (UTC). Heavy commands ran sequentially. Admission delta ran alone with two workers. Only this report is committed, on `proof/lower-e2j-2-report`, based directly on the tested HEAD.

| Check | Exit | Wall seconds | Result |
|---|---:|---:|---|
| Uncached oracle pins | 0 | 177.468 | PASS, both fixtures |
| Lower pins | 0 | 28.422 | PASS, all 37 reported tests/subtests |
| Mutant oracle pins | 1 | 39.176 | Expected FAIL, version-zero refusal |
| Mutant lower pins | 1 | 15.928 | Expected FAIL, two admission controls |
| Full lower suite | 0 | 926.397 | PASS, no failures |
| Full IR suite | 0 | 77.509 | PASS, no failures |
| Admission delta | 0 | 776.949 | PASS, complete corpus |
| Formatting and vet | 0 | 132.631 | PASS, no output |

**Step 2: decisive backend observations**

Exact requested commands, run on the unmodified HEAD:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestStaticSelfConstructionAgreesWithNode' -count=1 -v -timeout 40m
go test ./internal/lower -run 'TestStaticSelfConstruction|TestStaticWalk|TestNeverCall' -count=1 -v -timeout 30m
```

`tsc_version_zero.ts`: each backend is compared to source Node. JSON string notation below preserves the complete stdout bytes, including the final newline.

| Observation | Native release | Sanitized native | JavaScript | Source Node |
|---|---|---|---|---|
| status | PASS | PASS | PASS | PASS |
| exit | 0 | 0 | 0 | 0 |
| stdout | `"0.0.0-0\n1.2.3-beta.1+build.5\n2.0.0\nDebug Failure. False expression: Invalid argument: major\n"` | `"0.0.0-0\n1.2.3-beta.1+build.5\n2.0.0\nDebug Failure. False expression: Invalid argument: major\n"` | `"0.0.0-0\n1.2.3-beta.1+build.5\n2.0.0\nDebug Failure. False expression: Invalid argument: major\n"` | `"0.0.0-0\n1.2.3-beta.1+build.5\n2.0.0\nDebug Failure. False expression: Invalid argument: major\n"` |
| stderr | `""` | `""` | `""` | `""` |

`earlier_static.a`: native, sanitized, JavaScript, and Node all PASS, exit 0, stdout `"6 4 p2\n10 4 p2\n"`, stderr `""`. There are no unmodified pin failures, so there are no failing got/want pairs to omit.

Every reported unmodified oracle and lower test/subtest outcome:

```text
--- PASS: TestStaticSelfConstructionAgreesWithNode (0.00s)
    --- PASS: TestStaticSelfConstructionAgreesWithNode/earlier_static.a (162.49s)
    --- PASS: TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts (162.49s)
```

```text
--- PASS: TestNeverCallMethodTwinsStayUnproven (0.00s)
    --- PASS: TestNeverCallMethodTwinsStayUnproven/an_object-literal_method (0.22s)
    --- PASS: TestNeverCallMethodTwinsStayUnproven/a_class_instance's_method (0.15s)
--- PASS: TestNeverCallNamespaceMember (0.00s)
    --- PASS: TestNeverCallNamespaceMember/through_the_namespace (0.47s)
    --- PASS: TestNeverCallNamespaceMember/a_write_through_the_namespace_stops_.a (0.04s)
    --- PASS: TestNeverCallNamespaceMember/a_write_through_the_namespace_stops_.ts (0.05s)
    --- PASS: TestNeverCallNamespaceMember/through_a_parameter (0.07s)
--- PASS: TestNeverCallTwinsKeepTheirCheck (0.00s)
    --- PASS: TestNeverCallTwinsKeepTheirCheck/a_never_helper_that_returns_on_one_path (0.08s)
    --- PASS: TestNeverCallTwinsKeepTheirCheck/an_object-literal_method (0.10s)
    --- PASS: TestNeverCallTwinsKeepTheirCheck/a_helper_typed_void (0.13s)
    --- PASS: TestNeverCallTwinsKeepTheirCheck/a_const_bound_to_an_arrow (0.09s)
--- PASS: TestStaticSelfConstructionStops (0.00s)
    --- PASS: TestStaticSelfConstructionStops/a_later_static_read_by_the_constructor_body (0.06s)
    --- PASS: TestStaticSelfConstructionStops/a_later_static_read_by_a_static_method_the_constructor_calls (0.05s)
    --- PASS: TestStaticSelfConstructionStops/a_class_with_a_base (0.04s)
    --- PASS: TestStaticSelfConstructionStops/a_call_the_walk_can't_resolve (0.05s)
    --- PASS: TestStaticSelfConstructionStops/a_later_static_read_by_an_instance_field_initializer (0.03s)
    --- PASS: TestStaticSelfConstructionStops/a_later_static_read_by_an_instance_getter_the_constructor_reads (0.06s)
    --- PASS: TestStaticSelfConstructionStops/a_later_static_read_by_a_parameter_default (0.05s)
--- PASS: TestStaticWalkPerEdgeWitness (0.00s)
    --- PASS: TestStaticWalkPerEdgeWitness/a_declared_function_reading_a_later_static_through_a_getter (0.04s)
    --- PASS: TestStaticWalkPerEdgeWitness/an_arrow_in_the_static_initializer_handed_to_the_constructor (0.04s)
    --- PASS: TestStaticWalkPerEdgeWitness/the_same_with_the_static_declared_first (0.09s)
--- PASS: TestStaticWalkReachAdmissions (0.00s)
    --- PASS: TestStaticWalkReachAdmissions/assert_reached_with_no_verboseDebugInfo (0.03s)
    --- PASS: TestStaticWalkReachAdmissions/assert_reached_with_a_closure_reading_a_later_static (0.08s)
    --- PASS: TestStaticWalkReachAdmissions/a_program_object's_captureStackTrace (0.04s)
    --- PASS: TestStaticWalkReachAdmissions/assert_reached_with_a_static_arrow_field_reading_a_later_static (0.04s)
    --- PASS: TestStaticWalkReachAdmissions/assert_reached_with_an_outer_function_reading_the_class (0.06s)
    --- PASS: TestStaticWalkReachAdmissions/Node's_Error.captureStackTrace (0.30s)
--- PASS: TestNeverCallVersionZeroCore (1.53s)
--- PASS: TestNeverCallProvesAssertIsDefined (0.00s)
    --- PASS: TestNeverCallProvesAssertIsDefined/proof (0.11s)
    --- PASS: TestNeverCallProvesAssertIsDefined/checkDefined_of_undefined (11.98s)
    --- PASS: TestNeverCallProvesAssertIsDefined/checkDefined (12.01s)
```

The oracle helper compares successful runs without printing their stdout. A temporary evidence-only test called the same `onNode`, `lowered`, `releasedUncached`, `nativelyUncached`, and `onJavaScriptBackend` helpers to record every observation above with `ADAMIC_GATE_UNCACHED=1`. It passed in 17.262 seconds and was removed. Its first attempt used a relative path and failed in the evidence harness, before backend comparison; the corrected attempt used `filepath.Abs`, as the repository helper requires. Complete first-attempt diagnostic:

```text
=== RUN   TestProofE2JBackendEvidence
    cache_test.go:284: Rel: can't make /workspace/adamic/internal/oracle/testdata/static_self_construction/earlier_static.a relative to ../../internal/oracle/testdata/static_self_construction
--- FAIL: TestProofE2JBackendEvidence (4.14s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=0
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	4.151s
FAIL
```

Corrected evidence-capture log:

```text
=== RUN   TestProofE2JBackendEvidence
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types b0da814891b1 hit 0.39
    proof_e2j_evidence_test.go:15: earlier_static.a native PASS exit=0 stdout="6 4 p2\n10 4 p2\n" stderr=""
    proof_e2j_evidence_test.go:15: earlier_static.a sanitized PASS exit=0 stdout="6 4 p2\n10 4 p2\n" stderr=""
    proof_e2j_evidence_test.go:15: earlier_static.a JavaScript PASS exit=0 stdout="6 4 p2\n10 4 p2\n" stderr=""
    proof_e2j_evidence_test.go:15: earlier_static.a Node PASS exit=0 stdout="6 4 p2\n10 4 p2\n" stderr=""
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types b0da814891b1 hit 0.02
    proof_e2j_evidence_test.go:15: tsc_version_zero.ts Node PASS exit=0 stdout="0.0.0-0\n1.2.3-beta.1+build.5\n2.0.0\nDebug Failure. False expression: Invalid argument: major\n" stderr=""
    proof_e2j_evidence_test.go:15: tsc_version_zero.ts native PASS exit=0 stdout="0.0.0-0\n1.2.3-beta.1+build.5\n2.0.0\nDebug Failure. False expression: Invalid argument: major\n" stderr=""
    proof_e2j_evidence_test.go:15: tsc_version_zero.ts sanitized PASS exit=0 stdout="0.0.0-0\n1.2.3-beta.1+build.5\n2.0.0\nDebug Failure. False expression: Invalid argument: major\n" stderr=""
    proof_e2j_evidence_test.go:15: tsc_version_zero.ts JavaScript PASS exit=0 stdout="0.0.0-0\n1.2.3-beta.1+build.5\n2.0.0\nDebug Failure. False expression: Invalid argument: major\n" stderr=""
--- PASS: TestProofE2JBackendEvidence (6.54s)
PASS
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=4
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	6.556s
```

**Step 3: only “Parameter calls never known”**

In `internal/lower/class_static.go`, asserted that the old text occurred exactly once, saved the original bytes outside the repository, and replaced it with the documented new text:

```text
old: if judged && target.Kind == ast.KindParameter
new: if false && target.Kind == ast.KindParameter
```

Reran both exact step-2 commands sequentially. The oracle command exited 1 in 39.176 seconds; the lower command exited 1 in 15.928 seconds. These are the complete red leaf subtests and their full got/want evidence. Parent test FAIL lines are aggregates of these leaves.

1. `TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts` — FAIL.

```text
got (lowering error):
/workspace/adamic/internal/oracle/testdata/static_self_construction/tsc_version_zero.ts:67:111: stage 0 can't lower an indirect call before static fields are initialized; move it after the field initializers yet

want: successful lowering and all backends agreeing with source Node
Node exit: 0
Node stdout: "0.0.0-0\n1.2.3-beta.1+build.5\n2.0.0\nDebug Failure. False expression: Invalid argument: major\n"
Node stderr: ""
Native/sanitized/JavaScript output: not produced; the subtest stopped at lowering.
```

2. `TestStaticWalkReachAdmissions/assert_reached_with_no_verboseDebugInfo` — FAIL.

```text
got:
/tmp/adamic-gate/TestStaticWalkReachAdmissionsassert_reached_with_no_verboseDebu664104279/001/main.a:8:70: stage 0 can't lower an indirect call before static fields are initialized; move it after the field initializers yet

want: nil lowering error (must lower)
```

3. `TestStaticWalkPerEdgeWitness/the_same_with_the_static_declared_first` — FAIL.

```text
got:
/tmp/adamic-gate/TestStaticWalkPerEdgeWitnessthe_same_with_the_static_declared_f2642386282/001/main.a:2:9: stage 0 can't lower an indirect call before static fields are initialized; move it after the field initializers yet

want: nil lowering error (must lower)
```

Every mutant subtest outcome (the oracle’s other fixture and all remaining lower subtests passed):

```text
--- FAIL: TestStaticSelfConstructionAgreesWithNode (0.00s)
    --- FAIL: TestStaticSelfConstructionAgreesWithNode/tsc_version_zero.ts (4.54s)
    --- PASS: TestStaticSelfConstructionAgreesWithNode/earlier_static.a (20.18s)
```

```text
--- PASS: TestNeverCallTwinsKeepTheirCheck (0.00s)
    --- PASS: TestNeverCallTwinsKeepTheirCheck/a_never_helper_that_returns_on_one_path (0.14s)
    --- PASS: TestNeverCallTwinsKeepTheirCheck/a_helper_typed_void (0.30s)
    --- PASS: TestNeverCallTwinsKeepTheirCheck/an_object-literal_method (0.31s)
    --- PASS: TestNeverCallTwinsKeepTheirCheck/a_const_bound_to_an_arrow (0.34s)
--- PASS: TestNeverCallMethodTwinsStayUnproven (0.00s)
    --- PASS: TestNeverCallMethodTwinsStayUnproven/an_object-literal_method (0.27s)
    --- PASS: TestNeverCallMethodTwinsStayUnproven/a_class_instance's_method (0.30s)
--- FAIL: TestStaticWalkPerEdgeWitness (0.00s)
    --- PASS: TestStaticWalkPerEdgeWitness/a_declared_function_reading_a_later_static_through_a_getter (0.07s)
    --- FAIL: TestStaticWalkPerEdgeWitness/the_same_with_the_static_declared_first (0.03s)
    --- PASS: TestStaticWalkPerEdgeWitness/an_arrow_in_the_static_initializer_handed_to_the_constructor (0.05s)
--- PASS: TestStaticSelfConstructionStops (0.00s)
    --- PASS: TestStaticSelfConstructionStops/a_later_static_read_by_the_constructor_body (0.07s)
    --- PASS: TestStaticSelfConstructionStops/a_class_with_a_base (0.06s)
    --- PASS: TestStaticSelfConstructionStops/a_call_the_walk_can't_resolve (0.04s)
    --- PASS: TestStaticSelfConstructionStops/a_later_static_read_by_a_static_method_the_constructor_calls (0.05s)
    --- PASS: TestStaticSelfConstructionStops/a_later_static_read_by_an_instance_getter_the_constructor_reads (0.04s)
    --- PASS: TestStaticSelfConstructionStops/a_later_static_read_by_an_instance_field_initializer (0.03s)
    --- PASS: TestStaticSelfConstructionStops/a_later_static_read_by_a_parameter_default (0.03s)
--- PASS: TestNeverCallNamespaceMember (0.00s)
    --- PASS: TestNeverCallNamespaceMember/through_the_namespace (0.38s)
    --- PASS: TestNeverCallNamespaceMember/a_write_through_the_namespace_stops_.ts (0.04s)
    --- PASS: TestNeverCallNamespaceMember/through_a_parameter (0.03s)
    --- PASS: TestNeverCallNamespaceMember/a_write_through_the_namespace_stops_.a (0.03s)
--- FAIL: TestStaticWalkReachAdmissions (0.00s)
    --- FAIL: TestStaticWalkReachAdmissions/assert_reached_with_no_verboseDebugInfo (0.05s)
    --- PASS: TestStaticWalkReachAdmissions/a_program_object's_captureStackTrace (0.05s)
    --- PASS: TestStaticWalkReachAdmissions/assert_reached_with_a_closure_reading_a_later_static (0.06s)
    --- PASS: TestStaticWalkReachAdmissions/assert_reached_with_a_static_arrow_field_reading_a_later_static (0.03s)
    --- PASS: TestStaticWalkReachAdmissions/assert_reached_with_an_outer_function_reading_the_class (0.04s)
    --- PASS: TestStaticWalkReachAdmissions/Node's_Error.captureStackTrace (0.29s)
--- PASS: TestNeverCallVersionZeroCore (1.57s)
--- PASS: TestNeverCallProvesAssertIsDefined (0.00s)
    --- PASS: TestNeverCallProvesAssertIsDefined/proof (0.16s)
    --- PASS: TestNeverCallProvesAssertIsDefined/checkDefined_of_undefined (1.91s)
    --- PASS: TestNeverCallProvesAssertIsDefined/checkDefined (2.01s)
```

Restored the exact original bytes immediately after both mutant commands. `git diff --exit-code` exited 0 and printed nothing; `git status --short` was empty. A second clean-tree check after the full suites and admission run also passed. No mutant or evidence harness is committed.

**Step 4: complete lower suite**

```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

Exit 0; wall time 926.397 seconds (15m 26.397s). Every failing test / first error line: none. Environment failures: none. Complete `/tmp/lower.log`:

```text
ok  	github.com/system-inc/adamic/internal/lower	924.468s
```

**Step 4b: IR source and field guards**

```sh
go test ./internal/ir -count=1 -timeout 30m
```

Exit 0; wall time 77.509 seconds. Every failure: none. The full package run includes `TestCallTargetReadersLower`, the other call-target reader guards, `TestEveryCalleeFieldReadGoesThroughMethodCallee`, and the IR serialization/field guards. No tests were selected out. Complete output:

```text
ok  	github.com/system-inc/adamic/internal/ir	60.062s
```

**Step 6: complete admission delta**

```sh
python3 cloud/admission-corpus/manifest.py --sha a114e99ef3d55024743d8c49bf62324224780358 > /tmp/lower-e2j-2/manifest.json
go build -o /tmp/lower-e2j-2/admission-delta ./cmd/adamic-admission-delta
/tmp/lower-e2j-2/admission-delta --base 49d8529950ce6330d011a118eae4d2a80bc4d14d --head a114e99ef3d55024743d8c49bf62324224780358 --manifest /tmp/lower-e2j-2/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --workers 2 --json > /tmp/lower-e2j-2/admission.json 2> /tmp/lower-e2j-2/admission.stderr
```

Built the tool outside the repository (exit 0, 2.345 seconds), then ran it alone with `--workers 2`. No corpus filter, shard, or runtime sampling budget was used.

| Measurement | Result |
|---|---|
| Exit | 0 |
| Verdict | pass |
| void | false |
| Void reasons | None at report or program level |
| Wall seconds | 776.949 |
| complete_corpus | true |
| Classified programs | 5358 |
| diff_count | 1 |
| admitted | 0 |
| Newly refused | 0 |
| Accepted by both | 3357 |
| Refused by both | 2001 |
| Classification stderr differs | 0 |
| Compile timeouts | 0 |
| Runtime timeouts | 0 |
| Retried runs | 0 |
| Still timed out | 0 |
| Runtime selections | 115 |
| Omitted | 0 |
| Node agreements | 13 |
| Accepted negative witnesses | 50 |
| Ruled runtime stops | 52 |

Changed program: `internal/oracle/testdata/static_self_construction/tsc_version_zero.ts`, blob `9de25c52b04c2bcdd37772cb6edc13f8e7f32059`. `diff_ts_unmatched` is empty. The changed fixture is already admitted by both compilers; the explicit native/sanitized/JavaScript/Node proof is recorded in step 2.

Compared `base_compile.stderr` and `compile.stderr` for all 5,358 program entries: zero differences. Consequently there are no differing program paths or stderr text pairs to print. Scanned both classification observations, every program’s void reason, and runtime timeout records: zero timeouts. No timeout reruns at either SHA were needed.

Manifest counts: witnesses=2 fixtures=4410 gaps=105 review=20 fuzz=0 stage3=821.

Gate stderr:

```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 1
```

Phase times in seconds:

```json
{
  "base_build": 16.555770497,
  "checkout": 10.035306533,
  "checkout_and_provision": 10.700197493,
  "classification": 447.968362979,
  "head_build": 10.734548961,
  "manifest_and_verification": 15.203940688,
  "provision": 0.664887975,
  "runtime": 275.406388203,
  "total": 776.575512245
}
```

Manifest Git blob: `d5a679ff5eedec5088b824be23999bb017baaa50`; generator revision: `a114e99ef3d55024743d8c49bf62324224780358`; generator blob: `09d061a03b1e016387e975d25ee0cd5b75ac6c1f`.

**Step 1 and environment reproducibility**

```sh
if [ -f /workspace/adamic-tools/env.sh ]; then source /workspace/adamic-tools/env.sh; fi
git fetch origin compiler/tsc-version-zero-agrees-2
git checkout --detach a114e99ef3d55024743d8c49bf62324224780358
git submodule update --init --recursive
gofmt -l internal/lower internal/oracle
go vet ./internal/lower ./internal/oracle
```

Fetch, checkout, and recursive submodule update exited 0. Cohere pin: `5360753e19e4417675a90e92a1823bea66b13918`; nested compiler pin: `d92d9bfee114c80be2c375d72edae966176e3a4f`. Formatting printed nothing; vet exited 0 and printed nothing (combined wall time 132.631 seconds). No AGENTS.md files were found under `/workspace`.

Loaded `/workspace/adamic-tools/env.sh` before work and sourced it in the environment used by subsequent commands. It sets `TMPDIR=/tmp/adamic-gate`; that missing directory was created outside the repository before Go ran. The proof setup, logs, npm install, and upstream checkout are outside the repository.

```sh
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/lower-e2j-2/TypeScript
export PATH=/tmp/lower-e2j-2/api/node_modules/.bin:$PATH
```

These certificate exports were present for npm and inherited by the admission gate and its children. TLS verification stayed enabled. Provisioned upstream with `git init`, remote `https://github.com/microsoft/TypeScript.git`, `git fetch --depth=1 origin 050880ce59e30b356b686bd3144efe24f875ebc8`, and detached checkout of FETCH_HEAD. Verified exact HEAD `050880ce59e30b356b686bd3144efe24f875ebc8`. Copied `stage3/api/package.json` and its lockfile to `/tmp/lower-e2j-2/api` and ran `npm ci` there (3 packages). The installed pins are TypeScript 6.0.3 and @types/node 25.3.3.

```text
tsc: Version 6.0.3
go: go version go1.27.1 linux/amd64
node: v24.19.0
clang: 20.1.8 (llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
```

**External evidence fingerprints**

Raw artifacts remain external to the repository; their SHA-256 values bind this report to the captured evidence. Only `proof/lower-e2j-2/report.md` is versioned.

| Artifact | SHA-256 |
|---|---|
| `/tmp/lower-e2j-2/manifest.json` | `2668c9fbbe40f8fa7b2ad1a5ca1ad806b5a5a26d2f7786c9df4096bc92eed24c` |
| `/tmp/lower-e2j-2/admission.json` | `ef683012ca59e67f7899cb6df6eeee04080b035994ecd7efc6dbd9863fa7dbfc` |
| `/tmp/lower-e2j-2/admission.stderr` | `dee4069ac9a7b3ac2d9f56fc26ddf53c881f98f3dcd3e91bffd533c5c3d925d2` |
| `/tmp/lower-e2j-2/backend-evidence.json` | `194a8101223f36d306163ac01bfcce8e11d1dd6ec60fc7393c68337dbb53faa9` |
| `/tmp/lower-e2j-2/oracle.log` | `dc4ca1963ba92e7edde533d778ffea1fdbbe6c33567786ab458730b990eed307` |
| `/tmp/lower-e2j-2/lower-pins.log` | `866d724587336242546b606a80d47d7dd35cf36db42a82d56ba64a0edf26f863` |
| `/tmp/lower-e2j-2/mutant-oracle.log` | `183116a062c7c1e56dda496de2346f6ef21e52688927b40d84ea0163010fbb8d` |
| `/tmp/lower-e2j-2/mutant-lower.log` | `8736f7632dea9b491567c90c8aa469b72f85317c32899135a488c59abf152bce` |
| `/tmp/lower-e2j-2/ir.log` | `418c5908c44fc72700670c8f7f9b9c6ba059bd1139c298428b99517130c73659` |
| `/tmp/lower-e2j-2/step1.log` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `/tmp/lower-e2j-2/backend-evidence-fixed.log` | `d876b34cba3e783c7c3ff71d0f595b43e4156736c86cf2f242e3bef18826c130` |
| `/tmp/lower.log` | `c3ba218eb0477ea7c4f291c70bad1e8d9bf37723171650d89ff6c690a4a6f109` |
