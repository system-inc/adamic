**Adamic compiler proof — task #fj0vqtf, 2026-11-11 UTC**

**Agreement result: FAIL.** The unmodified baseline passes 16 of 20 subtests. Both frozen-object cases stop instead of lowering; both accessor cases produce an earlier stop than expected. Mutant (a) introduces six additional failures, demonstrating the reach rule for class instances and unknown parameter reach; the accessor mutant effect is masked by the existing earlier stop. Mutant (b) introduces four additional failures, demonstrating the holder rule for narrow/wide aliases. The full lower suite also fails. This is not a clean agreement proof.

**Step 2 — baseline agreement.** Measures byte-for-byte execution agreement with Node for accepted programs across release, sanitized, and JavaScript backends, and exact repository-relative lowering-stop text for rejected programs. The test also asserts that all 20 fixtures are present.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestHostMemberWriteAgreesOrStops$' -count=1 -v -timeout 40m
```

Result: FAIL; exit 1; wall time 462.348 seconds; 16 PASS, 4 FAIL.

**Step 3 — mutation sensitivity.** Runs the same uncached gate once for each exact edit from `internal/lower/host-member-write.mutants.md`, with the other edit absent. Each old-text occurrence count was asserted to equal 1 before editing. The original bytes were restored after each run, and each subsequent `git diff --exit-code` returned 0 with no output.

Mutant (a) measures whether dropping reach-rule enforcement admits unsafe receiver allocations:

```go
// Old (count = 1)
if err := l.hostMemberReach(write, store, reaches, graph.projectionIndex()); err != nil {
// New
if err := l.hostMemberReach(write, store, reaches, graph.projectionIndex()); false && err != nil {
```

Result: FAIL; exit 1; wall time 106.638 seconds; 10 PASS, 10 FAIL. New failures versus baseline: `s3_class_instance.a`, `s3_class_instance.ts`, `s4_exported_parameter.a`, `s4_exported_parameter.ts`, `s5_arrow_parameter.a`, `s5_arrow_parameter.ts`. All six lower when they should stop. The documented `s6_accessor_pair` effect is not demonstrated: both twins retain the baseline getter/method stop mismatch.

Mutant (b) measures whether checking only the write-site type loses compatibility obligations from other types holding the same literal:

```go
// Old (count = 1)
holders := l.entriesRecordHandles(node)
// New
holders := []*checker.Type{l.checker.GetTypeAtLocation(write.target.AsPropertyAccessExpression().Expression)}
```

Result: FAIL; exit 1; wall time 113.769 seconds; 12 PASS, 8 FAIL. New failures versus baseline: `s1_narrow_wide.a`, `s1_narrow_wide.ts`, `s2_narrow_wide_object.a`, `s2_narrow_wide_object.ts`. All four lower when they should stop.

Every subtest result below is under `TestHostMemberWriteAgreesOrStops/`:

| Subtest | Baseline | Mutant (a) | Mutant (b) |
|---|---|---|---|
| a1_devirtualized.a | PASS | PASS | PASS |
| a1_devirtualized.ts | PASS | PASS | PASS |
| a2_function_expression.a | PASS | PASS | PASS |
| a2_function_expression.ts | PASS | PASS | PASS |
| f1_frozen.a | FAIL | FAIL | FAIL |
| f1_frozen.ts | FAIL | FAIL | FAIL |
| s1_narrow_wide.a | PASS | PASS | FAIL |
| s1_narrow_wide.ts | PASS | PASS | FAIL |
| s2_narrow_wide_object.a | PASS | PASS | FAIL |
| s2_narrow_wide_object.ts | PASS | PASS | FAIL |
| s3_class_instance.a | PASS | FAIL | PASS |
| s3_class_instance.ts | PASS | FAIL | PASS |
| s4_exported_parameter.a | PASS | FAIL | PASS |
| s4_exported_parameter.ts | PASS | FAIL | PASS |
| s5_arrow_parameter.a | PASS | FAIL | PASS |
| s5_arrow_parameter.ts | PASS | FAIL | PASS |
| s6_accessor_pair.a | FAIL | FAIL | FAIL |
| s6_accessor_pair.ts | FAIL | FAIL | FAIL |
| s7_spread.a | PASS | PASS | PASS |
| s7_spread.ts | PASS | PASS | PASS |

**Baseline — every failure message, untruncated.**

`f1_frozen.a`:

```text
    host_member_write_test.go:67: Lower: /workspace/adamic/internal/oracle/testdata/host_member_write/f1_frozen.a:10:15: stage 0 can't lower Object.freeze on a shape not proven by a plain literal or its const binding yet
```

`f1_frozen.ts`:

```text
    host_member_write_test.go:67: Lower: /workspace/adamic/internal/oracle/testdata/host_member_write/f1_frozen.ts:10:15: stage 0 can't lower Object.freeze on a shape not proven by a plain literal or its const binding yet
```

`s6_accessor_pair.a`:

```text
    host_member_write_test.go:62: stop: got "internal/oracle/testdata/host_member_write/s6_accessor_pair.a:14:13: stage 0 can't lower a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation yet", want "internal/oracle/testdata/host_member_write/s6_accessor_pair.a:15:1: stage 0 can't lower a write to method read of host, which may hold an object literal with an accessor, whose setter a write may run yet"
```

`s6_accessor_pair.ts`:

```text
    host_member_write_test.go:62: stop: got "internal/oracle/testdata/host_member_write/s6_accessor_pair.ts:14:13: stage 0 can't lower a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation yet", want "internal/oracle/testdata/host_member_write/s6_accessor_pair.ts:15:1: stage 0 can't lower a write to method read of host, which may hold an object literal with an accessor, whose setter a write may run yet"
```

**Mutant (a) — every failure message, untruncated.**

`f1_frozen.a`:

```text
    host_member_write_test.go:67: Lower: /workspace/adamic/internal/oracle/testdata/host_member_write/f1_frozen.a:10:15: stage 0 can't lower Object.freeze on a shape not proven by a plain literal or its const binding yet
```

`f1_frozen.ts`:

```text
    host_member_write_test.go:67: Lower: /workspace/adamic/internal/oracle/testdata/host_member_write/f1_frozen.ts:10:15: stage 0 can't lower Object.freeze on a shape not proven by a plain literal or its const binding yet
```

`s3_class_instance.a`:

```text
    host_member_write_test.go:59: lowered, want "internal/oracle/testdata/host_member_write/s3_class_instance.a:12:1: stage 0 can't lower a write to method read of host, which may hold an instance of a class, whose methods aren't fields of its own yet"
```

`s3_class_instance.ts`:

```text
    host_member_write_test.go:59: lowered, want "internal/oracle/testdata/host_member_write/s3_class_instance.ts:12:1: stage 0 can't lower a write to method read of host, which may hold an instance of a class, whose methods aren't fields of its own yet"
```

`s4_exported_parameter.a`:

```text
    host_member_write_test.go:59: lowered, want "internal/oracle/testdata/host_member_write/s4_exported_parameter.a:6:2: stage 0 can't lower a write to method read of host, which may hold an object not known to be one of the program's object literals yet"
```

`s4_exported_parameter.ts`:

```text
    host_member_write_test.go:59: lowered, want "internal/oracle/testdata/host_member_write/s4_exported_parameter.ts:6:2: stage 0 can't lower a write to method read of host, which may hold an object not known to be one of the program's object literals yet"
```

`s5_arrow_parameter.a`:

```text
    host_member_write_test.go:59: lowered, want "internal/oracle/testdata/host_member_write/s5_arrow_parameter.a:6:2: stage 0 can't lower a write to method read of host, which may hold an object not known to be one of the program's object literals yet"
```

`s5_arrow_parameter.ts`:

```text
    host_member_write_test.go:59: lowered, want "internal/oracle/testdata/host_member_write/s5_arrow_parameter.ts:6:2: stage 0 can't lower a write to method read of host, which may hold an object not known to be one of the program's object literals yet"
```

`s6_accessor_pair.a`:

```text
    host_member_write_test.go:62: stop: got "internal/oracle/testdata/host_member_write/s6_accessor_pair.a:14:13: stage 0 can't lower a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation yet", want "internal/oracle/testdata/host_member_write/s6_accessor_pair.a:15:1: stage 0 can't lower a write to method read of host, which may hold an object literal with an accessor, whose setter a write may run yet"
```

`s6_accessor_pair.ts`:

```text
    host_member_write_test.go:62: stop: got "internal/oracle/testdata/host_member_write/s6_accessor_pair.ts:14:13: stage 0 can't lower a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation yet", want "internal/oracle/testdata/host_member_write/s6_accessor_pair.ts:15:1: stage 0 can't lower a write to method read of host, which may hold an object literal with an accessor, whose setter a write may run yet"
```

**Mutant (b) — every failure message, untruncated.**

`f1_frozen.a`:

```text
    host_member_write_test.go:67: Lower: /workspace/adamic/internal/oracle/testdata/host_member_write/f1_frozen.a:10:15: stage 0 can't lower Object.freeze on a shape not proven by a plain literal or its const binding yet
```

`f1_frozen.ts`:

```text
    host_member_write_test.go:67: Lower: /workspace/adamic/internal/oracle/testdata/host_member_write/f1_frozen.ts:10:15: stage 0 can't lower Object.freeze on a shape not proven by a plain literal or its const binding yet
```

`s1_narrow_wide.a`:

```text
    host_member_write_test.go:59: lowered, want "internal/oracle/testdata/host_member_write/s1_narrow_wide.a:5:1: stage 0 can't lower a write to method read of wide of a value of type () => string | undefined, which the object, also held as Narrow, can't hold: its read there is () => string yet"
```

`s1_narrow_wide.ts`:

```text
    host_member_write_test.go:59: lowered, want "internal/oracle/testdata/host_member_write/s1_narrow_wide.ts:5:1: stage 0 can't lower a write to method read of wide of a value of type () => string | undefined, which the object, also held as Narrow, can't hold: its read there is () => string yet"
```

`s2_narrow_wide_object.a`:

```text
    host_member_write_test.go:59: lowered, want "internal/oracle/testdata/host_member_write/s2_narrow_wide_object.a:21:1: stage 0 can't lower a write to method read of wide of a value of type () => Animal, which the object, also held as Narrow, can't hold: its read there is () => Dog yet"
```

`s2_narrow_wide_object.ts`:

```text
    host_member_write_test.go:59: lowered, want "internal/oracle/testdata/host_member_write/s2_narrow_wide_object.ts:21:1: stage 0 can't lower a write to method read of wide of a value of type () => Animal, which the object, also held as Narrow, can't hold: its read there is () => Dog yet"
```

`s6_accessor_pair.a`:

```text
    host_member_write_test.go:62: stop: got "internal/oracle/testdata/host_member_write/s6_accessor_pair.a:14:13: stage 0 can't lower a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation yet", want "internal/oracle/testdata/host_member_write/s6_accessor_pair.a:15:1: stage 0 can't lower a write to method read of host, which may hold an object literal with an accessor, whose setter a write may run yet"
```

`s6_accessor_pair.ts`:

```text
    host_member_write_test.go:62: stop: got "internal/oracle/testdata/host_member_write/s6_accessor_pair.ts:14:13: stage 0 can't lower a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation yet", want "internal/oracle/testdata/host_member_write/s6_accessor_pair.ts:15:1: stage 0 can't lower a write to method read of host, which may hold an object literal with an accessor, whose setter a write may run yet"
```

**Step 4 — full lower regression suite.** Measures regression coverage across the lower package, including host-origin diagnostics and equality between the scanning reference and indexed lowering pipeline. Executed on the restored, unmodified source.

```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

Result: FAIL; exit 1; wall time 2387.580 seconds (39m 47.580s), including Go build/startup. The test package reports 2353.775 seconds. The run completed before timeout. There are 13 failing top-level tests: one host-origin test, one scanning-reference pass-list test, and 11 cycle-equality shards. Failing child tests are identified below. Every failing test's complete first error is preserved in this entire, untruncated 9384-byte log:

```text
--- FAIL: TestMethodPresenceThroughHostOrigin (9.57s)
    --- FAIL: TestMethodPresenceThroughHostOrigin/a_write_to_now_through_the_local (0.58s)
        method_presence_host_origin_test.go:269: got /tmp/adamic-gate/TestMethodPresenceThroughHostOrigina_write_to_now_through_the_l1300661283/001/performanceCore.ts:2:1: stage 0 can't lower writing performance host slots through a structural view yet, want a stop on the write's line, performanceCore.ts:40:, saying Adamic 0.1 refuses a method read as a value (now would lose its object, and this with it)
--- FAIL: TestScanningReferenceCallsEveryPass (0.02s)
    cycles_scanning_pipeline_test.go:65: the scanning reference's passes differ from the real pipeline's:
        real (exemptions taken out): [lowering.moduleOrder lowering.collectViewedElements lowering.noteInheritance lowering.enumInitialization lowering.noteAccessorNames lowering.noteOmittedOptionals lowering.namespaceInitialization lowering.refuse lowering.declareModule lowering.recordExports lowering.moduleStatements lowering.finishClassCalls lowering.finishAccessors lowering.finishToPrimitive lowering.checkIterableViews lowering.exceptions lowering.checkLibrarySpreads lowering.checkAccessorSpreads lowering.checkDynamicKeyedGetters checkGeneratorRoads flow.NormalizeAsync lowering.modulePath lowering.findCycles lowering.checkBoundArguments lowering.checkViewCopies lowering.checkViewWidenedReads lowering.checkHostMemberWrites lowering.checkPromiseConsumption lowering.checkViewLogicalWrites lowering.applyProgramMetadata fresh.QueryOwnershipTransfers lowering.moveRefused readiness borrow counters lowering.voidDiscards]
        reference: [lowering.moduleOrder lowering.collectViewedElements lowering.noteInheritance lowering.enumInitialization lowering.noteAccessorNames lowering.noteOmittedOptionals lowering.namespaceInitialization lowering.refuse lowering.declareModule lowering.recordExports lowering.moduleStatements lowering.finishClassCalls lowering.finishAccessors lowering.finishToPrimitive lowering.checkIterableViews lowering.exceptions lowering.checkLibrarySpreads lowering.checkAccessorSpreads lowering.checkDynamicKeyedGetters checkGeneratorRoads flow.NormalizeAsync lowering.modulePath lowering.findCycles lowering.checkBoundArguments lowering.checkViewCopies lowering.checkViewWidenedReads lowering.checkPromiseConsumption lowering.checkViewLogicalWrites lowering.applyProgramMetadata fresh.QueryOwnershipTransfers lowering.moveRefused readiness borrow counters lowering.voidDiscards]
--- FAIL: TestCycleFinderEquality_125 (16.40s)
    --- FAIL: TestCycleFinderEquality_125/internal/oracle/testdata/host_member_write/s3_class_instance.a (0.95s)
        cycles_equality_test.go:402: cycle equality difference: internal/oracle/testdata/host_member_write/s3_class_instance.a ProgramRegion=false
            scanning:  (cycle=false)
            indexed: *lower.NotYet: /workspace/adamic/internal/oracle/testdata/host_member_write/s3_class_instance.a:12:1: stage 0 can't lower a write to method read of host, which may hold an instance of a class, whose methods aren't fields of its own yet (cycle=false)
--- FAIL: TestCycleFinderEquality_116 (24.46s)
    --- FAIL: TestCycleFinderEquality_116/internal/oracle/testdata/host_member_write/s7_spread.ts (0.48s)
        cycles_equality_test.go:402: cycle equality difference: internal/oracle/testdata/host_member_write/s7_spread.ts ProgramRegion=false
            scanning:  (cycle=false)
            indexed: *lower.NotYet: /workspace/adamic/internal/oracle/testdata/host_member_write/s7_spread.ts:12:1: stage 0 can't lower a write to method read of host, which may hold an object literal made with a spread yet (cycle=false)
--- FAIL: TestCycleFinderEquality_107 (3.83s)
    --- FAIL: TestCycleFinderEquality_107/internal/oracle/testdata/host_member_write/s2_narrow_wide_object.ts (0.19s)
        cycles_equality_test.go:402: cycle equality difference: internal/oracle/testdata/host_member_write/s2_narrow_wide_object.ts ProgramRegion=false
            scanning:  (cycle=false)
            indexed: *lower.NotYet: /workspace/adamic/internal/oracle/testdata/host_member_write/s2_narrow_wide_object.ts:21:1: stage 0 can't lower a write to method read of wide of a value of type () => Animal, which the object, also held as Narrow, can't hold: its read there is () => Dog yet (cycle=false)
--- FAIL: TestCycleFinderEquality_086 (21.01s)
    --- FAIL: TestCycleFinderEquality_086/internal/oracle/testdata/host_member_write/s1_narrow_wide.ts (0.19s)
        cycles_equality_test.go:402: cycle equality difference: internal/oracle/testdata/host_member_write/s1_narrow_wide.ts ProgramRegion=false
            scanning:  (cycle=false)
            indexed: *lower.NotYet: /workspace/adamic/internal/oracle/testdata/host_member_write/s1_narrow_wide.ts:5:1: stage 0 can't lower a write to method read of wide of a value of type () => string | undefined, which the object, also held as Narrow, can't hold: its read there is () => string yet (cycle=false)
--- FAIL: TestCycleFinderEquality_081 (12.93s)
    --- FAIL: TestCycleFinderEquality_081/internal/oracle/testdata/host_member_write/s2_narrow_wide_object.a (0.25s)
        cycles_equality_test.go:402: cycle equality difference: internal/oracle/testdata/host_member_write/s2_narrow_wide_object.a ProgramRegion=false
            scanning:  (cycle=false)
            indexed: *lower.NotYet: /workspace/adamic/internal/oracle/testdata/host_member_write/s2_narrow_wide_object.a:21:1: stage 0 can't lower a write to method read of wide of a value of type () => Animal, which the object, also held as Narrow, can't hold: its read there is () => Dog yet (cycle=false)
--- FAIL: TestCycleFinderEquality_080 (11.19s)
    --- FAIL: TestCycleFinderEquality_080/internal/oracle/testdata/host_member_write/s5_arrow_parameter.a (0.22s)
        cycles_equality_test.go:402: cycle equality difference: internal/oracle/testdata/host_member_write/s5_arrow_parameter.a ProgramRegion=false
            scanning:  (cycle=false)
            indexed: *lower.NotYet: /workspace/adamic/internal/oracle/testdata/host_member_write/s5_arrow_parameter.a:6:2: stage 0 can't lower a write to method read of host, which may hold an object not known to be one of the program's object literals yet (cycle=false)
--- FAIL: TestCycleFinderEquality_032 (16.82s)
    --- FAIL: TestCycleFinderEquality_032/internal/oracle/testdata/host_member_write/s7_spread.a (0.22s)
        cycles_equality_test.go:402: cycle equality difference: internal/oracle/testdata/host_member_write/s7_spread.a ProgramRegion=false
            scanning:  (cycle=false)
            indexed: *lower.NotYet: /workspace/adamic/internal/oracle/testdata/host_member_write/s7_spread.a:12:1: stage 0 can't lower a write to method read of host, which may hold an object literal made with a spread yet (cycle=false)
--- FAIL: TestCycleFinderEquality_017 (11.77s)
    --- FAIL: TestCycleFinderEquality_017/internal/oracle/testdata/host_member_write/s4_exported_parameter.ts (0.17s)
        cycles_equality_test.go:402: cycle equality difference: internal/oracle/testdata/host_member_write/s4_exported_parameter.ts ProgramRegion=false
            scanning:  (cycle=false)
            indexed: *lower.NotYet: /workspace/adamic/internal/oracle/testdata/host_member_write/s4_exported_parameter.ts:6:2: stage 0 can't lower a write to method read of host, which may hold an object not known to be one of the program's object literals yet (cycle=false)
--- FAIL: TestCycleFinderEquality_010 (6.57s)
    --- FAIL: TestCycleFinderEquality_010/internal/oracle/testdata/bind_refused/method_write.a (0.26s)
        cycles_equality_test.go:402: cycle equality difference: internal/oracle/testdata/bind_refused/method_write.a ProgramRegion=false
            scanning:  (cycle=false)
            indexed: *lower.NotYet: /workspace/adamic/internal/oracle/testdata/bind_refused/method_write.a:5:1: stage 0 can't lower a write to method now of h, which may hold an instance of a class, whose methods aren't fields of its own yet (cycle=false)
--- FAIL: TestCycleFinderEquality_007 (14.33s)
    --- FAIL: TestCycleFinderEquality_007/internal/oracle/testdata/host_member_write/s3_class_instance.ts (0.22s)
        cycles_equality_test.go:402: cycle equality difference: internal/oracle/testdata/host_member_write/s3_class_instance.ts ProgramRegion=false
            scanning:  (cycle=false)
            indexed: *lower.NotYet: /workspace/adamic/internal/oracle/testdata/host_member_write/s3_class_instance.ts:12:1: stage 0 can't lower a write to method read of host, which may hold an instance of a class, whose methods aren't fields of its own yet (cycle=false)
--- FAIL: TestCycleFinderEquality_004 (16.32s)
    --- FAIL: TestCycleFinderEquality_004/internal/oracle/testdata/host_member_write/s5_arrow_parameter.ts (0.13s)
        cycles_equality_test.go:402: cycle equality difference: internal/oracle/testdata/host_member_write/s5_arrow_parameter.ts ProgramRegion=false
            scanning:  (cycle=false)
            indexed: *lower.NotYet: /workspace/adamic/internal/oracle/testdata/host_member_write/s5_arrow_parameter.ts:6:2: stage 0 can't lower a write to method read of host, which may hold an object not known to be one of the program's object literals yet (cycle=false)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	2353.775s
FAIL
```

**Step 1 — formatting and static analysis.** `gofmt -l internal/lower internal/oracle` measures formatting conformance; exit 0, wall time 0.550 seconds, no output. `go vet ./internal/lower ./internal/oracle` measures static issues detected by Go vet; exit 0, wall time 350.817 seconds, no output. Both PASS. There is no diagnostic output to reproduce.

**Checkout and environment.** Loaded `/workspace/adamic-tools/env.sh` before setup and commands. Fetched `origin compiler/host-member-write-1`, checked out `6cdc2d4b3bb42786ed569b5476affa9e047aa80d` detached, and ran `git submodule update --init --recursive` successfully. Its immediate parent is `1bb9663d4a9cbf1466a758c41c2293c9c869e6ab`. Submodules: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`, `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`.

Created `/tmp/adamic-gate` outside the repository for `TMPDIR` before Go commands. Provisioned an external upstream checkout from `https://github.com/microsoft/TypeScript.git` at `/tmp/adamic-proof-typescript`, verified HEAD `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported `ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-proof-typescript`. Copied `stage3/api/package.json` and its lockfile to `/tmp/adamic-proof-api` and successfully ran `npm ci --prefix /tmp/adamic-proof-api` (3 packages installed). Prepended `/tmp/adamic-proof-api/node_modules/.bin` to PATH and verified `tsc --version` as `Version 6.0.3`.

Exported these values before npm and for every gate process and its children:

```sh
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
```

Tools: Go `go1.27.1 linux/amd64`; Node `v24.19.0`; clang `20.1.8`. Heavy commands ran one at a time. Wall times use a monotonic clock around each complete command. All requested results were reported as their steps finished. No compiler or fixture fixes were made.

**Report publication.** Created branch `proof/lower-fj0w1-1-report` from the tested HEAD. Committed only `proof/lower-fj0w1-1/report.md` and pushed that branch to `origin`. No other branch was pushed.
