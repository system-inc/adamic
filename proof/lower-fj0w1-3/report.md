**Compiler agreement proof — #fj0vqtf**

The targeted agreement check passed all 24 subtests, both main s8 checks passed, and both mutants were killed by exactly their specified rows. Lower and IR regression suites passed. **The admission-delta gate is VOID** because 20 native compilations exceeded its 10-second limit, including all six newly admitted programs, and all 20 timed out again after isolated retries. This is not a complete admission-delta agreement proof. No compiler code was fixed.

**Step 2: agreement or exact refusal.** This measures exact pinned lowering diagnostics for refused programs, and source Node agreement with release native, sanitized native, and JavaScript backends for admitted programs. Exit 0; wall time 713.62 seconds. No failures.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestHostMemberWriteAgreesOrStops$' -count=1 -v -timeout 40m
```

The following table names all 24 subtests and also records both independent mutant runs. Each program name is under `TestHostMemberWriteAgreesOrStops/`.

| Program | HEAD .a | HEAD .ts | Mutant (a) .a | Mutant (a) .ts | Mutant (b) .a | Mutant (b) .ts |
|---|---|---|---|---|---|---|
| a1_devirtualized | PASS | PASS | PASS | PASS | PASS | PASS |
| a2_function_expression | PASS | PASS | PASS | PASS | PASS | PASS |
| f1_frozen | PASS | PASS | PASS | PASS | PASS | PASS |
| s1_narrow_wide | PASS | PASS | PASS | PASS | FAIL | FAIL |
| s2_narrow_wide_object | PASS | PASS | PASS | PASS | FAIL | FAIL |
| s3_class_instance | PASS | PASS | FAIL | FAIL | PASS | PASS |
| s4_exported_parameter | PASS | PASS | FAIL | FAIL | PASS | PASS |
| s5_arrow_parameter | PASS | PASS | FAIL | FAIL | PASS | PASS |
| s6_accessor_pair | PASS | PASS | PASS | PASS | PASS | PASS |
| s6b_getter_only | PASS | PASS | PASS | PASS | PASS | PASS |
| s7_spread | PASS | PASS | PASS | PASS | PASS | PASS |
| s8_earlier_refusal | PASS | PASS | PASS | PASS | PASS | PASS |

Both HEAD s8 subtests PASS. Whole contents of `internal/oracle/testdata/host_member_write/s8_earlier_refusal.a.stop` (including its terminating newline):

```text
internal/oracle/testdata/host_member_write/s8_earlier_refusal.a:12:11: Adamic 0.1 refuses ==; use ===, which doesn't coerce
```

Whole contents of `internal/oracle/testdata/host_member_write/s8_earlier_refusal.ts.stop` (including its terminating newline):

```text
internal/oracle/testdata/host_member_write/s8_earlier_refusal.ts:12:11: Adamic 0.1 refuses ==; use ===, which doesn't coerce
```

**Step 2m: the same earlier refusal on main.** This measures whether the parent revision returns exactly those pins when given HEAD’s test and programs. Exit 0; wall time 381.69 seconds. A separate detached worktree at `/tmp/adamic-main` was created at `d43437079a7e7ca723168e67c5efb5bf7ebee216`; recursive submodules were initialized there; only HEAD’s requested test and fixture directory were taken into that worktree.

```sh
git worktree add --detach /tmp/adamic-main d43437079a7e7ca723168e67c5efb5bf7ebee216
git -C /tmp/adamic-main submodule update --init --recursive
git -C /tmp/adamic-main checkout 6548afd035da564452863075a74e4167e9db50e2 -- internal/oracle/host_member_write_test.go internal/oracle/testdata/host_member_write
cd /tmp/adamic-main
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestHostMemberWriteAgreesOrStops$/^s8_earlier_refusal\.' -count=1 -v -timeout 40m
```

`s8_earlier_refusal.a`: PASS. Whole got and want (the passing test establishes equality):

```text
got:  internal/oracle/testdata/host_member_write/s8_earlier_refusal.a:12:11: Adamic 0.1 refuses ==; use ===, which doesn't coerce
want: internal/oracle/testdata/host_member_write/s8_earlier_refusal.a:12:11: Adamic 0.1 refuses ==; use ===, which doesn't coerce
```

`s8_earlier_refusal.ts`: PASS. Whole got and want (the passing test establishes equality):

```text
got:  internal/oracle/testdata/host_member_write/s8_earlier_refusal.ts:12:11: Adamic 0.1 refuses ==; use ===, which doesn't coerce
want: internal/oracle/testdata/host_member_write/s8_earlier_refusal.ts:12:11: Adamic 0.1 refuses ==; use ===, which doesn't coerce
```

Removed the worktree with `git worktree remove --force /tmp/adamic-main`. The primary checkout’s `git diff --exit-code` and `git diff --cached --exit-code` both exited 0 with no output.

**Step 3: mutation sensitivity.** Each exact edit from `internal/lower/host-member-write.mutants.md` was applied alone to `internal/lower/host_member_write.go`, after asserting its old text occurred exactly once. Each used the same uncached 24-subtest command as step 2. The table above reports every verdict. After each run, the file was restored from HEAD; both unstaged and staged diffs exited 0 with no output.

Mutant (a), Dropping the reach rule: old-text count **1**; test exit 1; wall time 152.91 seconds.

Exactly s3_class_instance, s4_exported_parameter, and s5_arrow_parameter, each .a and .ts, turned red from green in step 2. **Exactly the specified rows turned red.** s6 and s6b stayed green.

Old:

```go
		if err := l.hostMemberReach(write, reaches, graph.projectionIndex()); err != nil {
```

New:

```go
		if err := l.hostMemberReach(write, reaches, graph.projectionIndex()); false && err != nil {
```

Every failure’s whole message follows, untruncated. In each case got means successful lowering with no error; the quoted text is the whole expected refusal.

```text
host_member_write_test.go:68: lowered, want "internal/oracle/testdata/host_member_write/s3_class_instance.a:12:1: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)"
host_member_write_test.go:68: lowered, want "internal/oracle/testdata/host_member_write/s3_class_instance.ts:12:1: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)"
host_member_write_test.go:68: lowered, want "internal/oracle/testdata/host_member_write/s4_exported_parameter.a:6:2: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)"
host_member_write_test.go:68: lowered, want "internal/oracle/testdata/host_member_write/s4_exported_parameter.ts:6:2: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)"
host_member_write_test.go:68: lowered, want "internal/oracle/testdata/host_member_write/s5_arrow_parameter.a:6:2: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)"
host_member_write_test.go:68: lowered, want "internal/oracle/testdata/host_member_write/s5_arrow_parameter.ts:6:2: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)"
```

Mutant (b), Judging holders using only the write site’s type: old-text count **1**; test exit 1; wall time 131.70 seconds.

Exactly s1_narrow_wide and s2_narrow_wide_object, each .a and .ts, turned red from green in step 2. **Exactly the specified rows turned red.**

Old:

```go
		holders := l.entriesRecordHandles(node)
```

New:

```go
		holders := []*checker.Type{l.checker.GetTypeAtLocation(write.target.AsPropertyAccessExpression().Expression)}
```

Every failure’s whole message follows, untruncated. In each case got means successful lowering with no error; the quoted text is the whole expected refusal.

```text
host_member_write_test.go:68: lowered, want "internal/oracle/testdata/host_member_write/s1_narrow_wide.a:5:1: stage 0 can't lower a write to method read of wide of a value of type () => string | undefined, which the object, also held as Narrow, can't hold: its read there is () => string yet"
host_member_write_test.go:68: lowered, want "internal/oracle/testdata/host_member_write/s1_narrow_wide.ts:5:1: stage 0 can't lower a write to method read of wide of a value of type () => string | undefined, which the object, also held as Narrow, can't hold: its read there is () => string yet"
host_member_write_test.go:68: lowered, want "internal/oracle/testdata/host_member_write/s2_narrow_wide_object.a:21:1: stage 0 can't lower a write to method read of wide of a value of type () => Animal, which the object, also held as Narrow, can't hold: its read there is () => Dog yet"
host_member_write_test.go:68: lowered, want "internal/oracle/testdata/host_member_write/s2_narrow_wide_object.ts:21:1: stage 0 can't lower a write to method read of wide of a value of type () => Animal, which the object, also held as Narrow, can't hold: its read there is () => Dog yet"
```

**Step 5: admission delta against the parent.** This measures lowering acceptance over the complete head corpus with base and head compilers, then runtime agreement for newly admitted programs and required witness/stop coverage. It ran alone.

```sh
python3 cloud/admission-corpus/manifest.py --sha HEAD > /tmp/manifest.json
go run ./cmd/adamic-admission-delta --base d43437079a7e7ca723168e67c5efb5bf7ebee216 --head HEAD --manifest /tmp/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --json > /tmp/delta.json
```

Manifest generation succeeded. The requested `go run` command exited **1**; its gate binary exited **3** (`exit status 3`), with JSON `verdict: "void"` and `void: true`. Wall time: **1934.14 seconds**; gate-internal total: 1922.24 seconds. No altered timeout or follow-up rerun was used.

Whole stderr, including the exact summary line and complete void reasons:

```text
negative witnesses: 50
ruled run-time stops: 47
ruled backend divergences: 0
runtime timeouts: 20 runs timed out, 20 retried alone, 20 timed out again
changed programs: 24
admission delta is void: internal/oracle/testdata/host_member_write/a1_devirtualized.a: native compile timeout; internal/oracle/testdata/host_member_write/a1_devirtualized.ts: native compile timeout; internal/oracle/testdata/host_member_write/a2_function_expression.a: native compile timeout; internal/oracle/testdata/host_member_write/a2_function_expression.ts: native compile timeout; internal/oracle/testdata/host_member_write/f1_frozen.a: native compile timeout; internal/oracle/testdata/host_member_write/f1_frozen.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_boolean.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_literal.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_token.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_write.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_write_repaired.ts: native compile timeout; internal/oracle/testdata/intl_collator/in_compare.a: native compile timeout; internal/oracle/testdata/intl_collator/locale_en.a: native compile timeout; internal/oracle/testdata/intl_collator/locale_malformed.a: native compile timeout; internal/oracle/testdata/intl_collator/locale_variable.a: native compile timeout; internal/oracle/testdata/placeholder_nonnull_null_saved_leak.a: native compile timeout; internal/oracle/testdata/reflect_get/stop_literal_getter.a: native compile timeout
exit status 3
```

The complete corpus classified 5257 unique programs: 3282 accepted by both, 1969 refused by both, and 6 newly accepted; no newly refused program. The manifest contained witnesses=2, fixtures=4309, gaps=105, review=20, fuzz=0, stage3=821; the revision diff covered 24 programs. `diff_ts_unmatched` is empty. Five classification workers, 45-second classification limit, 10-second runtime-command limit, no filter/shard, zero budget limit, 121 runtime selections, and zero omissions were recorded.

Every class change is listed below. Base “refused” means lowering exited 1; HEAD “accepted” means lowering exited 0. For every listed program Node and JavaScript exited 0 with identical stdout and empty stderr. Native compilation timed out, including its isolated retry, so native execution did not occur and three-way agreement is **unproven in this gate** (`agree: false`, `void_reason: "native compile timeout"`).

| Path | Base | HEAD | Node / JavaScript / native |
|---|---|---|---|
| `internal/oracle/testdata/host_member_write/a1_devirtualized.a` | Refused (1) | Accepted (0), newly-accepted | Node = JavaScript; native compile timeout; three-way agreement unproven |
| `internal/oracle/testdata/host_member_write/a1_devirtualized.ts` | Refused (1) | Accepted (0), newly-accepted | Node = JavaScript; native compile timeout; three-way agreement unproven |
| `internal/oracle/testdata/host_member_write/a2_function_expression.a` | Refused (1) | Accepted (0), newly-accepted | Node = JavaScript; native compile timeout; three-way agreement unproven |
| `internal/oracle/testdata/host_member_write/a2_function_expression.ts` | Refused (1) | Accepted (0), newly-accepted | Node = JavaScript; native compile timeout; three-way agreement unproven |
| `internal/oracle/testdata/host_member_write/f1_frozen.a` | Refused (1) | Accepted (0), newly-accepted | Node = JavaScript; native compile timeout; three-way agreement unproven |
| `internal/oracle/testdata/host_member_write/f1_frozen.ts` | Refused (1) | Accepted (0), newly-accepted | Node = JavaScript; native compile timeout; three-way agreement unproven |

Whole base diagnostics for those six programs (HEAD lowering stdout and stderr are empty in each case):

```text
adamic: /tmp/adamic-gate/admission-delta-591581047/head/internal/oracle/testdata/host_member_write/a1_devirtualized.a:17:1: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)
adamic: /tmp/adamic-gate/admission-delta-591581047/head/internal/oracle/testdata/host_member_write/a1_devirtualized.ts:17:1: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)
adamic: /tmp/adamic-gate/admission-delta-591581047/head/internal/oracle/testdata/host_member_write/a2_function_expression.a:11:1: Adamic 0.1 refuses a method read as a value (step would lose its object, and this with it); call it in an arrow that keeps the object: (by) => counter.step(by) (unbound-method)
adamic: /tmp/adamic-gate/admission-delta-591581047/head/internal/oracle/testdata/host_member_write/a2_function_expression.ts:11:1: Adamic 0.1 refuses a method read as a value (step would lose its object, and this with it); call it in an arrow that keeps the object: (by) => counter.step(by) (unbound-method)
adamic: /tmp/adamic-gate/admission-delta-591581047/head/internal/oracle/testdata/host_member_write/f1_frozen.a:12:2: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)
adamic: /tmp/adamic-gate/admission-delta-591581047/head/internal/oracle/testdata/host_member_write/f1_frozen.ts:12:2: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)
```

The matching Node and JavaScript stdout for each twin pair is:

`a1_devirtualized.a` and `a1_devirtualized.ts`:

```text
original a
original b
patched c
patched d
```

`a2_function_expression.a` and `a2_function_expression.ts`:

```text
2
20
```

`f1_frozen.a` and `f1_frozen.ts`:

```text
TypeError
Cannot assign to read only property 'read' of object '#<Object>'
original
```

**No program outside `internal/oracle/testdata/host_member_write` changes outcome. The list of outside changes is empty.** This also holds when comparing whole classification exit codes, stdout, stderr, and errors rather than just admission class. Within the directory, s1 and s2 in both extensions remain refused by both but change diagnostic; the other refused fixtures retain their diagnostics. These classification findings are separate from the runtime timeouts that make the overall gate void.

**Step 4: lower regression suite.** This measures lower-package regressions, scanning-reference pipeline parity, host-origin method presence, and indexed/scanning cycle equality.

```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

Exit **0**; wall time **3216.97 seconds** (53m 36.97s). Every failing test and whole first error: **none**. Whole output:

```text
ok  	github.com/system-inc/adamic/internal/lower	3158.484s
```

`TestScanningReferenceCallsEveryPass`: PASS. `TestMethodPresenceThroughHostOrigin`: PASS. `TestCycleFinderEqualityPartition`: PASS. Every equality shard passed; all names and results follow. These results come from the complete non-verbose suite; pass lines below enumerate its shard inventory, not a separate rerun.

```text
TestCycleFinderEquality_000: PASS
TestCycleFinderEquality_001: PASS
TestCycleFinderEquality_002: PASS
TestCycleFinderEquality_003: PASS
TestCycleFinderEquality_004: PASS
TestCycleFinderEquality_005: PASS
TestCycleFinderEquality_006: PASS
TestCycleFinderEquality_007: PASS
TestCycleFinderEquality_008: PASS
TestCycleFinderEquality_009: PASS
TestCycleFinderEquality_010: PASS
TestCycleFinderEquality_011: PASS
TestCycleFinderEquality_012: PASS
TestCycleFinderEquality_013: PASS
TestCycleFinderEquality_014: PASS
TestCycleFinderEquality_015: PASS
TestCycleFinderEquality_016: PASS
TestCycleFinderEquality_017: PASS
TestCycleFinderEquality_018: PASS
TestCycleFinderEquality_019: PASS
TestCycleFinderEquality_020: PASS
TestCycleFinderEquality_021: PASS
TestCycleFinderEquality_022: PASS
TestCycleFinderEquality_023: PASS
TestCycleFinderEquality_024: PASS
TestCycleFinderEquality_025: PASS
TestCycleFinderEquality_026: PASS
TestCycleFinderEquality_027: PASS
TestCycleFinderEquality_028: PASS
TestCycleFinderEquality_029: PASS
TestCycleFinderEquality_030: PASS
TestCycleFinderEquality_031: PASS
TestCycleFinderEquality_032: PASS
TestCycleFinderEquality_033: PASS
TestCycleFinderEquality_034: PASS
TestCycleFinderEquality_035: PASS
TestCycleFinderEquality_036: PASS
TestCycleFinderEquality_037: PASS
TestCycleFinderEquality_038: PASS
TestCycleFinderEquality_039: PASS
TestCycleFinderEquality_040: PASS
TestCycleFinderEquality_041: PASS
TestCycleFinderEquality_042: PASS
TestCycleFinderEquality_043: PASS
TestCycleFinderEquality_044: PASS
TestCycleFinderEquality_045: PASS
TestCycleFinderEquality_046: PASS
TestCycleFinderEquality_047: PASS
TestCycleFinderEquality_048: PASS
TestCycleFinderEquality_049: PASS
TestCycleFinderEquality_050: PASS
TestCycleFinderEquality_051: PASS
TestCycleFinderEquality_052: PASS
TestCycleFinderEquality_053: PASS
TestCycleFinderEquality_054: PASS
TestCycleFinderEquality_055: PASS
TestCycleFinderEquality_056: PASS
TestCycleFinderEquality_057: PASS
TestCycleFinderEquality_058: PASS
TestCycleFinderEquality_059: PASS
TestCycleFinderEquality_060: PASS
TestCycleFinderEquality_061: PASS
TestCycleFinderEquality_062: PASS
TestCycleFinderEquality_063: PASS
TestCycleFinderEquality_064: PASS
TestCycleFinderEquality_065: PASS
TestCycleFinderEquality_066: PASS
TestCycleFinderEquality_067: PASS
TestCycleFinderEquality_068: PASS
TestCycleFinderEquality_069: PASS
TestCycleFinderEquality_070: PASS
TestCycleFinderEquality_071: PASS
TestCycleFinderEquality_072: PASS
TestCycleFinderEquality_073: PASS
TestCycleFinderEquality_074: PASS
TestCycleFinderEquality_075: PASS
TestCycleFinderEquality_076: PASS
TestCycleFinderEquality_077: PASS
TestCycleFinderEquality_078: PASS
TestCycleFinderEquality_079: PASS
TestCycleFinderEquality_080: PASS
TestCycleFinderEquality_081: PASS
TestCycleFinderEquality_082: PASS
TestCycleFinderEquality_083: PASS
TestCycleFinderEquality_084: PASS
TestCycleFinderEquality_085: PASS
TestCycleFinderEquality_086: PASS
TestCycleFinderEquality_087: PASS
TestCycleFinderEquality_088: PASS
TestCycleFinderEquality_089: PASS
TestCycleFinderEquality_090: PASS
TestCycleFinderEquality_091: PASS
TestCycleFinderEquality_092: PASS
TestCycleFinderEquality_093: PASS
TestCycleFinderEquality_094: PASS
TestCycleFinderEquality_095: PASS
TestCycleFinderEquality_096: PASS
TestCycleFinderEquality_097: PASS
TestCycleFinderEquality_098: PASS
TestCycleFinderEquality_099: PASS
TestCycleFinderEquality_100: PASS
TestCycleFinderEquality_101: PASS
TestCycleFinderEquality_102: PASS
TestCycleFinderEquality_103: PASS
TestCycleFinderEquality_104: PASS
TestCycleFinderEquality_105: PASS
TestCycleFinderEquality_106: PASS
TestCycleFinderEquality_107: PASS
TestCycleFinderEquality_108: PASS
TestCycleFinderEquality_109: PASS
TestCycleFinderEquality_110: PASS
TestCycleFinderEquality_111: PASS
TestCycleFinderEquality_112: PASS
TestCycleFinderEquality_113: PASS
TestCycleFinderEquality_114: PASS
TestCycleFinderEquality_115: PASS
TestCycleFinderEquality_116: PASS
TestCycleFinderEquality_117: PASS
TestCycleFinderEquality_118: PASS
TestCycleFinderEquality_119: PASS
TestCycleFinderEquality_120: PASS
TestCycleFinderEquality_121: PASS
TestCycleFinderEquality_122: PASS
TestCycleFinderEquality_123: PASS
TestCycleFinderEquality_124: PASS
TestCycleFinderEquality_125: PASS
TestCycleFinderEquality_126: PASS
TestCycleFinderEquality_127: PASS
```

**Step 4b: IR regression suite and source guards.** This measures IR behavior and protected access to call targets and fields, including guards over lower’s source.

```sh
go test ./internal/ir -count=1 -timeout 30m > /tmp/ir.log 2>&1
```

Exit **0**; wall time **253.72 seconds**. Every failing test and whole first error: **none**. `TestCallTargetReadersLower`, `TestEveryCalleeFieldReadGoesThroughMethodCallee`, `TestMethodCalleeSeesTheCheckedField`, `TestEverySwitchOverGeneratorNextNamesIteratorNext`, and `TestMixedOptionalFixedLayoutPreservesMaybeNumber` passed with the suite. Whole output:

```text
ok  	github.com/system-inc/adamic/internal/ir	192.446s
```

**Step 1: formatting and static checks.** This measures gofmt compliance and vet diagnostics in the changed packages.

```sh
gofmt -l internal/lower internal/oracle
go vet ./internal/lower ./internal/oracle
```

Both exited **0**. `gofmt -l` printed nothing, as required. `go vet` printed nothing. There is no output to quote. Combined wall time: **503.36 seconds**.

**Checkout, environment, and preservation.** Loaded `/workspace/adamic-tools/env.sh` first. Ran `git fetch origin compiler/host-member-write-3 main`, checked out `6548afd035da564452863075a74e4167e9db50e2` detached, and ran `git submodule update --init --recursive`. Confirmed HEAD’s parent is exactly `d43437079a7e7ca723168e67c5efb5bf7ebee216`. Submodules: cohere `5360753e19e4417675a90e92a1823bea66b13918`, cohere/TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. No AGENTS.md was found.

Created the missing `/tmp/adamic-gate` outside the repository, matching the tools file’s TMPDIR. Provisioned the upstream checkout outside the repository at `/tmp/adamic-typescript`, detached at `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported `ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-typescript`. Copied stage3/api’s pinned package.json and package-lock.json to `/tmp/adamic-api` and ran `npm ci --ignore-scripts` there. `/tmp/adamic-api/node_modules/.bin` was placed on PATH; `tsc --version` reported 6.0.3. Go was 1.27.1 linux/amd64; Node was v24.19.0.

For npm, all proof commands, and gate children, exported `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true`. One heavy command ran at a time. No fixes were made. After all measurements, both unstaged and staged diffs were empty and the only worktree was the primary detached checkout.

Report publication: `proof/lower-fj0w1-3/report.md` is the sole file committed on new branch `proof/lower-fj0w1-3-report`, based on the measured HEAD, and that branch is pushed to origin. No other branch was pushed.
