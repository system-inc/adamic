**Adamic compiler proof — #fj0vqtf — 2026-11-11 UTC**

The targeted agreement/refusal check passes all 22 rows. Both documented mutants are caught by exactly their specified rows. The complete admission delta passes: six newly admitted programs all agree with Node, and no program outside `internal/oracle/testdata/host_member_write` changes outcome. The full lower suite passes. The full oracle suite fails 13 tests with missing WASI-sysroot or missing TypeScript-module diagnostics, reproduced in full below. No fixes were made.

HEAD: `f58958522c6dbce1edd48d190c47d15ee6736ec9`. Verified parent: `16429a30452d027809f0ea85bc395479423e493e`. Tests ran with HEAD detached. Results were reported as each step finished; heavy commands ran one at a time.

**Step 2 — targeted agreement and exact refusal check**

Measures Node agreement on release native, sanitized native, and JavaScript backends for admitted rows, and exact lowering-refusal strings for stopped rows.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestHostMemberWriteAgreesOrStops$' -count=1 -v -timeout 40m
```


Exit **0**, wall **174.570s**. All 22 subtests PASS; no failure messages. The following table also gives every subtest result for the two step-3 mutant runs. Each name is beneath `TestHostMemberWriteAgreesOrStops/`.

| Subtest | Step 2 | Mutant (a) | Mutant (b) |
|---|---|---|---|
| `a1_devirtualized.a` | PASS | PASS | PASS |
| `a1_devirtualized.ts` | PASS | PASS | PASS |
| `a2_function_expression.a` | PASS | PASS | PASS |
| `a2_function_expression.ts` | PASS | PASS | PASS |
| `f1_frozen.a` | PASS | PASS | PASS |
| `f1_frozen.ts` | PASS | PASS | PASS |
| `s1_narrow_wide.a` | PASS | PASS | FAIL |
| `s1_narrow_wide.ts` | PASS | PASS | FAIL |
| `s2_narrow_wide_object.a` | PASS | PASS | FAIL |
| `s2_narrow_wide_object.ts` | PASS | PASS | FAIL |
| `s3_class_instance.a` | PASS | FAIL | PASS |
| `s3_class_instance.ts` | PASS | FAIL | PASS |
| `s4_exported_parameter.a` | PASS | FAIL | PASS |
| `s4_exported_parameter.ts` | PASS | FAIL | PASS |
| `s5_arrow_parameter.a` | PASS | FAIL | PASS |
| `s5_arrow_parameter.ts` | PASS | FAIL | PASS |
| `s6_accessor_pair.a` | PASS | PASS | PASS |
| `s6_accessor_pair.ts` | PASS | PASS | PASS |
| `s6b_getter_only.a` | PASS | PASS | PASS |
| `s6b_getter_only.ts` | PASS | PASS | PASS |
| `s7_spread.a` | PASS | PASS | PASS |
| `s7_spread.ts` | PASS | PASS | PASS |

**Step 3 — mutation sensitivity**

Each documented old text was asserted to occur exactly once in `internal/lower/host_member_write.go` before replacement. Each mutant ran alone with the exact step-2 command. After each run the file was restored and `git diff --exit-code` exited **0** with no output.

**Mutant (a): reach rule dropped.** Measures whether the gate detects this lost obligation. Exit **1**, wall **40.769s**.

Old: `if err := l.hostMemberReach(write, reaches, graph.projectionIndex()); err != nil {`

New: `if err := l.hostMemberReach(write, reaches, graph.projectionIndex()); false && err != nil {`

Newly red relative to step 2: `s3_class_instance.a`, `s3_class_instance.ts`, `s4_exported_parameter.a`, `s4_exported_parameter.ts`, `s5_arrow_parameter.a`, `s5_arrow_parameter.ts`. **Exactly the specified rows turned red; all other rows stayed green.**

Complete failure messages (the observed result is successful lowering, so these errors say `lowered, want`):

`s3_class_instance.a`:

```text
    host_member_write_test.go:65: lowered, want "internal/oracle/testdata/host_member_write/s3_class_instance.a:12:1: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)"
```


`s3_class_instance.ts`:

```text
    host_member_write_test.go:65: lowered, want "internal/oracle/testdata/host_member_write/s3_class_instance.ts:12:1: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)"
```


`s4_exported_parameter.a`:

```text
    host_member_write_test.go:65: lowered, want "internal/oracle/testdata/host_member_write/s4_exported_parameter.a:6:2: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)"
```


`s4_exported_parameter.ts`:

```text
    host_member_write_test.go:65: lowered, want "internal/oracle/testdata/host_member_write/s4_exported_parameter.ts:6:2: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)"
```


`s5_arrow_parameter.a`:

```text
    host_member_write_test.go:65: lowered, want "internal/oracle/testdata/host_member_write/s5_arrow_parameter.a:6:2: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)"
```


`s5_arrow_parameter.ts`:

```text
    host_member_write_test.go:65: lowered, want "internal/oracle/testdata/host_member_write/s5_arrow_parameter.ts:6:2: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)"
```


**Mutant (b): only the write-site type judges the holder rule.** Measures whether the gate detects this lost obligation. Exit **1**, wall **38.697s**.

Old: `holders := l.entriesRecordHandles(node)`

New: `holders := []*checker.Type{l.checker.GetTypeAtLocation(write.target.AsPropertyAccessExpression().Expression)}`

Newly red relative to step 2: `s1_narrow_wide.a`, `s1_narrow_wide.ts`, `s2_narrow_wide_object.a`, `s2_narrow_wide_object.ts`. **Exactly the specified rows turned red; all other rows stayed green.**

Complete failure messages (the observed result is successful lowering, so these errors say `lowered, want`):

`s1_narrow_wide.a`:

```text
    host_member_write_test.go:65: lowered, want "internal/oracle/testdata/host_member_write/s1_narrow_wide.a:5:1: stage 0 can't lower a write to method read of wide of a value of type () => string | undefined, which the object, also held as Narrow, can't hold: its read there is () => string yet"
```


`s1_narrow_wide.ts`:

```text
    host_member_write_test.go:65: lowered, want "internal/oracle/testdata/host_member_write/s1_narrow_wide.ts:5:1: stage 0 can't lower a write to method read of wide of a value of type () => string | undefined, which the object, also held as Narrow, can't hold: its read there is () => string yet"
```


`s2_narrow_wide_object.a`:

```text
    host_member_write_test.go:65: lowered, want "internal/oracle/testdata/host_member_write/s2_narrow_wide_object.a:21:1: stage 0 can't lower a write to method read of wide of a value of type () => Animal, which the object, also held as Narrow, can't hold: its read there is () => Dog yet"
```


`s2_narrow_wide_object.ts`:

```text
    host_member_write_test.go:65: lowered, want "internal/oracle/testdata/host_member_write/s2_narrow_wide_object.ts:21:1: stage 0 can't lower a write to method read of wide of a value of type () => Animal, which the object, also held as Narrow, can't hold: its read there is () => Dog yet"
```


**Step 6 — complete parent-to-HEAD admission delta**

Measures base/head lowering admission over the head corpus, then compares stdout bytes and exit codes on source Node, the JavaScript backend, and native release for newly admitted programs and mandatory witness coverage. Ruled witnesses and runtime stops are counted separately from Node agreements.

```sh
python3 cloud/admission-corpus/manifest.py --sha HEAD > /tmp/manifest.json
go run ./cmd/adamic-admission-delta --base 16429a30452d027809f0ea85bc395479423e493e --head HEAD --manifest /tmp/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --json > /tmp/delta.json
```


Manifest exit **0**, wall **3.295s**. Manifest counts: witnesses=2, fixtures=4288, gaps=105, review=20, fuzz=0, stage3=821. Delta exit **0**, wall **597.451s**.

Result summary: **verdict=pass; complete_corpus=true; programs=5236; newly-accepted=6; newly-refused=0; accepted-by-both=3270; refused-by-both=1960; runtime checks=121; omitted=0; void=false.**

Complete emitted stderr summary:

```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 22
```


`changed programs: 22` counts programs added or changed in the source diff, not admission changes. No unmatched changed TypeScript programs were reported. Runtime results include 19 Node agreements, 50 accepted negative witnesses, and 52 ruled runtime stops; there were no stale or pending runtime stops and no runtime timeouts.

Every program whose admission class differs between base and HEAD:

| Path | Base outcome | HEAD outcome | Node / JavaScript / native agree? |
|---|---|---|---|
| `internal/oracle/testdata/host_member_write/a1_devirtualized.a` | refused (`unbound-method`), exit 1 | accepted, exit 0 (`newly-accepted`) | YES; all exit 0 |
| `internal/oracle/testdata/host_member_write/a1_devirtualized.ts` | refused (`unbound-method`), exit 1 | accepted, exit 0 (`newly-accepted`) | YES; all exit 0 |
| `internal/oracle/testdata/host_member_write/a2_function_expression.a` | refused (`unbound-method`), exit 1 | accepted, exit 0 (`newly-accepted`) | YES; all exit 0 |
| `internal/oracle/testdata/host_member_write/a2_function_expression.ts` | refused (`unbound-method`), exit 1 | accepted, exit 0 (`newly-accepted`) | YES; all exit 0 |
| `internal/oracle/testdata/host_member_write/f1_frozen.a` | refused (`unbound-method`), exit 1 | accepted, exit 0 (`newly-accepted`) | YES; all exit 0 |
| `internal/oracle/testdata/host_member_write/f1_frozen.ts` | refused (`unbound-method`), exit 1 | accepted, exit 0 (`newly-accepted`) | YES; all exit 0 |

All six runtime observations have empty stderr. Exact common stdout for each `.a`/`.ts` pair:

`a1_devirtualized`: "original a\noriginal b\npatched c\npatched d\n"

`a2_function_expression`: "2\n20\n"

`f1_frozen`: "TypeError\nCannot assign to read only property 'read' of object '#<Object>'\noriginal\n"

**No program outside `internal/oracle/testdata/host_member_write` changes outcome. Outside-change list: empty.** This holds for admission class and for compile exit code, stdout, stderr, and error, excluding wall time. The only additional compile-observation differences are the `.a` and `.ts` twins of `s1_narrow_wide` and `s2_narrow_wide_object`, whose refusal diagnostics change while they remain refused by both.

`internal/oracle/testdata/bind_refused/method_write.a` is `refused-by-both`, exit 1 on both, with identical diagnostics; the delta does not run it on the backends because it remains refused.

**Step 4 — full lower regression suite**

Measures the entire lower package, including scanning/indexed equivalence, host-origin method presence, and refusal behavior. `GOFLAGS=-v` was set solely to record explicit test results.

```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```


Exit **0**, wall **793.612s** (13m 13.612s). **No failing tests.** The log contains 10,369 passing test/subtest results. Intentionally caught child-mutant failures are output of passing mutation tests, not suite failures.

`TestScanningReferenceCallsEveryPass`: **PASS**.

`TestMethodPresenceThroughHostOrigin`: **PASS**. All 17 subtests also PASS.

`TestCycleFinderEqualityPartition`: **PASS**.

Every equality shard result:

```text
TestCycleFinderEquality_000 PASS
TestCycleFinderEquality_001 PASS
TestCycleFinderEquality_002 PASS
TestCycleFinderEquality_003 PASS
TestCycleFinderEquality_004 PASS
TestCycleFinderEquality_005 PASS
TestCycleFinderEquality_006 PASS
TestCycleFinderEquality_007 PASS
TestCycleFinderEquality_008 PASS
TestCycleFinderEquality_009 PASS
TestCycleFinderEquality_010 PASS
TestCycleFinderEquality_011 PASS
TestCycleFinderEquality_012 PASS
TestCycleFinderEquality_013 PASS
TestCycleFinderEquality_014 PASS
TestCycleFinderEquality_015 PASS
TestCycleFinderEquality_016 PASS
TestCycleFinderEquality_017 PASS
TestCycleFinderEquality_018 PASS
TestCycleFinderEquality_019 PASS
TestCycleFinderEquality_020 PASS
TestCycleFinderEquality_021 PASS
TestCycleFinderEquality_022 PASS
TestCycleFinderEquality_023 PASS
TestCycleFinderEquality_024 PASS
TestCycleFinderEquality_025 PASS
TestCycleFinderEquality_026 PASS
TestCycleFinderEquality_027 PASS
TestCycleFinderEquality_028 PASS
TestCycleFinderEquality_029 PASS
TestCycleFinderEquality_030 PASS
TestCycleFinderEquality_031 PASS
TestCycleFinderEquality_032 PASS
TestCycleFinderEquality_033 PASS
TestCycleFinderEquality_034 PASS
TestCycleFinderEquality_035 PASS
TestCycleFinderEquality_036 PASS
TestCycleFinderEquality_037 PASS
TestCycleFinderEquality_038 PASS
TestCycleFinderEquality_039 PASS
TestCycleFinderEquality_040 PASS
TestCycleFinderEquality_041 PASS
TestCycleFinderEquality_042 PASS
TestCycleFinderEquality_043 PASS
TestCycleFinderEquality_044 PASS
TestCycleFinderEquality_045 PASS
TestCycleFinderEquality_046 PASS
TestCycleFinderEquality_047 PASS
TestCycleFinderEquality_048 PASS
TestCycleFinderEquality_049 PASS
TestCycleFinderEquality_050 PASS
TestCycleFinderEquality_051 PASS
TestCycleFinderEquality_052 PASS
TestCycleFinderEquality_053 PASS
TestCycleFinderEquality_054 PASS
TestCycleFinderEquality_055 PASS
TestCycleFinderEquality_056 PASS
TestCycleFinderEquality_057 PASS
TestCycleFinderEquality_058 PASS
TestCycleFinderEquality_059 PASS
TestCycleFinderEquality_060 PASS
TestCycleFinderEquality_061 PASS
TestCycleFinderEquality_062 PASS
TestCycleFinderEquality_063 PASS
TestCycleFinderEquality_064 PASS
TestCycleFinderEquality_065 PASS
TestCycleFinderEquality_066 PASS
TestCycleFinderEquality_067 PASS
TestCycleFinderEquality_068 PASS
TestCycleFinderEquality_069 PASS
TestCycleFinderEquality_070 PASS
TestCycleFinderEquality_071 PASS
TestCycleFinderEquality_072 PASS
TestCycleFinderEquality_073 PASS
TestCycleFinderEquality_074 PASS
TestCycleFinderEquality_075 PASS
TestCycleFinderEquality_076 PASS
TestCycleFinderEquality_077 PASS
TestCycleFinderEquality_078 PASS
TestCycleFinderEquality_079 PASS
TestCycleFinderEquality_080 PASS
TestCycleFinderEquality_081 PASS
TestCycleFinderEquality_082 PASS
TestCycleFinderEquality_083 PASS
TestCycleFinderEquality_084 PASS
TestCycleFinderEquality_085 PASS
TestCycleFinderEquality_086 PASS
TestCycleFinderEquality_087 PASS
TestCycleFinderEquality_088 PASS
TestCycleFinderEquality_089 PASS
TestCycleFinderEquality_090 PASS
TestCycleFinderEquality_091 PASS
TestCycleFinderEquality_092 PASS
TestCycleFinderEquality_093 PASS
TestCycleFinderEquality_094 PASS
TestCycleFinderEquality_095 PASS
TestCycleFinderEquality_096 PASS
TestCycleFinderEquality_097 PASS
TestCycleFinderEquality_098 PASS
TestCycleFinderEquality_099 PASS
TestCycleFinderEquality_100 PASS
TestCycleFinderEquality_101 PASS
TestCycleFinderEquality_102 PASS
TestCycleFinderEquality_103 PASS
TestCycleFinderEquality_104 PASS
TestCycleFinderEquality_105 PASS
TestCycleFinderEquality_106 PASS
TestCycleFinderEquality_107 PASS
TestCycleFinderEquality_108 PASS
TestCycleFinderEquality_109 PASS
TestCycleFinderEquality_110 PASS
TestCycleFinderEquality_111 PASS
TestCycleFinderEquality_112 PASS
TestCycleFinderEquality_113 PASS
TestCycleFinderEquality_114 PASS
TestCycleFinderEquality_115 PASS
TestCycleFinderEquality_116 PASS
TestCycleFinderEquality_117 PASS
TestCycleFinderEquality_118 PASS
TestCycleFinderEquality_119 PASS
TestCycleFinderEquality_120 PASS
TestCycleFinderEquality_121 PASS
TestCycleFinderEquality_122 PASS
TestCycleFinderEquality_123 PASS
TestCycleFinderEquality_124 PASS
TestCycleFinderEquality_125 PASS
TestCycleFinderEquality_126 PASS
TestCycleFinderEquality_127 PASS
```


**Step 5 — full oracle regression and agreement suite**

Measures the complete oracle package: execution agreement, exact refusals, runtime instrumentation, and mutation checks. `GOFLAGS=-v` recorded explicit test and fixture results; `ADAMIC_GATE_UNCACHED` was not set for this full-suite run.

```sh
go test ./internal/oracle -count=1 -timeout 240m > /tmp/oracle.log 2>&1
```


Exit **1**, wall **2038.001s** (33m 58.001s). Results: **19,346 PASS, 782 SKIP, 13 FAIL**, counting tests and subtests. This was a completed failing run, not a timeout.

`internal/oracle/testdata/bind_refused/method_write.a`: **PASS** in `TestMethodBindRefusals/method_write` (0.06s), the only oracle-suite row that names/runs this fixture. It produced the expected `unbound-method` refusal. Separately, lower-suite `TestCycleFinderEquality_010/internal/oracle/testdata/bind_refused/method_write.a` passed in both scanning/indexed modes.

Every failing test and its complete first error follow. The assignment-proof tests force `NODE_PATH` to the repository’s `stage3/api/node_modules` (`internal/oracle/assignment_proofs_test.go:73`); the requested pinned install was outside the repository. No repository-local install, symlink, WASI setup, or source fix was added.

**TestStdoutColumnsRefusesWASI — FAIL**

```text
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
```


**TestAssignmentProofWiderTargetMutant — FAIL**

```text
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
```


**TestAssignmentProofWiderTarget — FAIL**

```text
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
```


**TestAssignmentProofCompoundMutant — FAIL**

```text
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
```


**TestAssignmentProofCompound — FAIL**

```text
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
```


**TestAssignmentProofChainedMutant — FAIL**

```text
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
```


**TestAssignmentProofChained — FAIL**

```text
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
```


**TestAssignmentProofReturnDefinedMutant — FAIL**

```text
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
```


**TestAssignmentProofIfNarrowingMutant — FAIL**

```text
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
```


**TestAssignmentProofReturnDefined — FAIL**

```text
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
```


**TestAssignmentProofIfNarrowing — FAIL**

```text
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
```


**TestAssignmentProofScannerKeywordMutant — FAIL**

```text
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
```


**TestAssignmentProofScannerKeyword — FAIL**

```text
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
```


**Step 1 — formatting and static analysis**

Measures Go formatting and `go vet` diagnostics in the changed compiler and oracle packages.

`gofmt -l internal/lower internal/oracle`: **exit 0**, wall **0.231s**. Complete output: empty.

`go vet ./internal/lower ./internal/oracle`: **exit 0**, wall **121.679s**. Complete output: empty.

**Checkout, environment, and report publication**

```sh
git fetch origin compiler/host-member-write-2 main
git checkout --detach f58958522c6dbce1edd48d190c47d15ee6736ec9
git submodule update --init --recursive
```


All three checkout commands succeeded. Parent confirmation was exact, not an abbreviated match. Submodules: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`.

Loaded `/workspace/adamic-tools/env.sh` first. Created `/tmp/adamic-gate` outside the repository before Go needed it. Provisioned upstream `https://github.com/microsoft/TypeScript.git` at `050880ce59e30b356b686bd3144efe24f875ebc8` in `/tmp/adamic-proof/TypeScript`, exported as `ADAMIC_TYPESCRIPT_SOURCE`. Copied `stage3/api/package.json` and its lockfile to `/tmp/adamic-proof/api` and ran `npm ci --ignore-scripts` there; pinned packages include TypeScript 6.0.3 and `@types/node` 25.3.3. Put the external `node_modules/.bin` first on PATH; verified `tsc --version` reports `Version 6.0.3`. Go: `go1.27.1 linux/amd64`; Node: `v24.19.0`.

Every proof/gate command inherited these exports, including npm and gate children:

```sh
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
```


Raw logs and metadata remain outside the repository in `/tmp`: `gofmt.log`, `vet.log`, `baseline.log`, `mutant-a.log`, `mutant-b.log`, `lower.log`, `oracle.log`, `manifest.json`, `delta.json`, and `delta.stderr`. No issue found was fixed. Both requested mutants were restored; the repository diff was clean before the report was added.

Report: `proof/lower-fj0w1-2/report.md`. Created new branch `proof/lower-fj0w1-2-report` from the checked-out HEAD, committed only this report file, and pushed that branch to `origin`. No other branch was pushed.
