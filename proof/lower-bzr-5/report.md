# Adamic record-spread proof — #bzrdem0

Proof completed 2026-11-11 (UTC) against detached HEAD `9041aa257ee1a8da0e7df5f0e2820b5ef3dbbb93`, the single commit above `d43437079a7e7ca723168e67c5efb5bf7ebee216` on `compiler/record-spread-undefined-5`.

## Decisive results

- Baseline: all seven oracle pins and both fixed-object lowering pins PASS.
- Both documented mutants are caught (oracle command exit 1). Each old-text occurrence assertion was exactly 1. Each restoration passed `git diff --exit-code` (exit 0).
- Mutant 1 differs from the mutant document's predicted null behavior: both null fixtures fail native C compilation; sanitized and JavaScript are not reached for those fixtures. Both undefined fixtures fail native and sanitized execution while JavaScript stays green.
- Mutant 2 makes all seven fixtures fail lowering, including both overwritten-key fixtures; no subtest stays green.
- `go test ./internal/ir -count=1 -timeout 30m` PASS, exit 0, with no failures. This includes `TestCallTargetReadersLower` and the IR field guards over lower's source.
- Formatting and vet PASS. No whole lower suite, whole oracle suite, or admission delta was run.

## Setup and step 1

Loaded `/workspace/adamic-tools/env.sh` first, and sourced it for subsequent proof commands. Created its configured `/tmp/adamic-gate` work directory outside the repository before Go ran.

Executed successfully, in order:

```sh
git fetch origin compiler/record-spread-undefined-5
git checkout --detach 9041aa257ee1a8da0e7df5f0e2820b5ef3dbbb93
git submodule update --init --recursive
gofmt -l internal/lower internal/oracle
go vet ./internal/lower ./internal/oracle
```

`gofmt -l` printed nothing (empty captured log); vet exited 0 and printed nothing. Recursive submodules:

```text
5360753e19e4417675a90e92a1823bea66b13918 cohere
d92d9bfee114c80be2c375d72edae966176e3a4f cohere/TypeScript
```

Provisioned upstream `https://github.com/microsoft/TypeScript.git` outside the repository at `/workspace/scratch/bzr5/typescript`, detached at verified commit `050880ce59e30b356b686bd3144efe24f875ebc8`. Exported `ADAMIC_TYPESCRIPT_SOURCE` to that directory.

Copied stage3/api's package.json and package-lock.json to `/workspace/scratch/bzr5/api` and ran `npm ci --prefix /workspace/scratch/bzr5/api` successfully (3 packages). This installed the pinned TypeScript 6.0.3 and @types/node 25.3.3. Exposed the external packages with links inside the real, gitignored directory `stage3/api/node_modules/`; exported `NODE_PATH=/workspace/adamic/stage3/api/node_modules` and put the external installation's `.bin` on PATH. Verified both `tsc --version` and `require("typescript").version` give 6.0.3. An initial top-level node_modules symlink was replaced with the real ignored directory because Git's directory-only ignore rule does not ignore that symlink; the working tree was clean before checks.

Exported these for npm, every test command, and inherited child processes:

```sh
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
```

Toolchain: Go 1.27.1 linux/amd64; Node v24.19.0; tsc 6.0.3; clang 20.1.8 (LLVM commit 87f0227cb60147a26a1eeb4fb06e3b505e9c7261). All heavy commands ran serially, including provisioning, vet, test runs, and the diagnostic replay. Logs and backups were kept outside the repository at `/workspace/scratch/bzr5`.

## Step 2 — baseline pins

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestRecordSpreadFromMissingAgreesWithNode' -count=1 -v -timeout 40m
go test ./internal/lower -run 'TestRecordSpreadFromMissingFixedObjectStops' -count=1 -v
```

Both commands exited 0. Oracle package time: 265.525s; lower package time: 0.096s.

| Oracle subtest | Result | Evidence checked |
|---|---|---|
| null_args.a | PASS | native, sanitized, JavaScript agree with Node |
| null_args.ts | PASS | native, sanitized, JavaScript agree with Node |
| overwritten_key.a | PASS | native, sanitized, JavaScript agree with Node |
| overwritten_key.ts | PASS | native, sanitized, JavaScript agree with Node |
| undefined_args.a | PASS | native, sanitized, JavaScript agree with Node |
| undefined_args.ts | PASS | native, sanitized, JavaScript agree with Node |
| tracing_write_stack_event.ts | PASS | Node succeeds; exact later lowering stop |

The tracing pin checks this exact suffix:

```text
tracing_write_stack_event.ts:28:37: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet
```

Both `TestRecordSpreadFromMissingFixedObjectStops/.a` and `/.ts` PASS. They require the exact suffix `main.a:6:11: stage 0 can't lower a spread into record storage from a fixed object yet` or its `main.ts` equivalent. There were no baseline failures requiring got/want dumps.

## Step 3 — mutants

Each mutant below was applied alone to `internal/lower/records.go`, after asserting its old text occurred exactly once, and tested with the same uncached oracle command from step 2. The original source bytes were restored after every run and `git diff --exit-code` returned 0 with no output.

### Mutant 1 — the missing arm dropped

```go
// Old (exactly one occurrence):
result.Spread = ir.Coalesce{Value: result.Spread, Fallback: ir.RecordLiteral{Element: of}, Of: ir.Record}
// New:
result.Spread = ir.Coalesce{Value: result.Spread, Fallback: ir.RecordLiteral{Element: of}, Of: ir.Record}.Value
```

Oracle command FAIL, exit 1, package time 32.978s.

| Subtest | Result | Native | Sanitized | JavaScript |
|---|---|---|---|---|
| null_args.a | FAIL | C compilation fails | not reached | not reached |
| null_args.ts | FAIL | C compilation fails | not reached | not reached |
| undefined_args.a | FAIL | exit -1 | exit 1, UBSan | PASS |
| undefined_args.ts | FAIL | exit -1 | exit 1, UBSan | PASS |
| overwritten_key.a | PASS | PASS | PASS | PASS |
| overwritten_key.ts | PASS | PASS | PASS | PASS |
| tracing_write_stack_event.ts | PASS | exact lowering-stop pin; no compiled backend | — | — |

The original harness only prints a mismatch category and stderr. To capture the whole got for executions, replayed this same mutant with temporary logging in `loomViewsDAgreesBackend`: logged Node's stdout/stderr/exit_code and each returned backend's stdout/stderr/exit_code, with no behavioral change. Replay also exited 1 (8.388s), with identical red/green subtests. Both records.go and the logging helper were then restored byte-for-byte, and the tracked diff was empty (exit 0).

For **each** of undefined_args.a and undefined_args.ts, complete outcomes from that replay are:

```text
Node want: stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit_code=0
native got: stdout="" stderr="" exit_code=-1
sanitized got: stdout="" stderr="record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n" exit_code=1
javascript got: stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit_code=0
```

Exit -1 is the harness's recorded process exit code; no more specific signal is asserted. The full C build diagnostics for **each** null fixture appear in the complete original log below. Compilation fails before a native execution result exists, and the helper's fatal build failure prevents sanitized and JavaScript runs for those subtests. These are C compilation failures, not lowering stops or measured runtime failures.

### Mutant 2 — the source judged as its own type

```go
// Old (exactly one occurrence):
present = l.checker.GetNonNullableType(own)
// New:
present = own
```

Oracle command FAIL, exit 1, package time 6.078s. All seven subtests FAIL; none stay green. Node runs before lowering; native, sanitized and JavaScript backend runs are not reached. Complete got error text for every red subtest:

```text
null_args.a:
/workspace/adamic/internal/oracle/testdata/record_spread_missing/null_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet

null_args.ts:
/workspace/adamic/internal/oracle/testdata/record_spread_missing/null_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet

overwritten_key.a:
/workspace/adamic/internal/oracle/testdata/record_spread_missing/overwritten_key.a:8:11: stage 0 can't lower a spread into record storage from a fixed object yet

overwritten_key.ts:
/workspace/adamic/internal/oracle/testdata/record_spread_missing/overwritten_key.ts:8:11: stage 0 can't lower a spread into record storage from a fixed object yet

undefined_args.a:
/workspace/adamic/internal/oracle/testdata/record_spread_missing/undefined_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet

undefined_args.ts:
/workspace/adamic/internal/oracle/testdata/record_spread_missing/undefined_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet

tracing_write_stack_event.ts got:
/workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:21:34: stage 0 can't lower a spread into record storage from a fixed object yet

tracing_write_stack_event.ts want:
the spread lowered and JSON.stringify of object references to stop at tracing_write_stack_event.ts:28:37: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet
```

## Step 4b — internal/ir

Ran against fully restored HEAD:

```sh
go test ./internal/ir -count=1 -timeout 30m
```

PASS, exit 0, package time 106.924s. No failures. The complete suite includes `TestCallTargetReadersLower`, the callee-field reader guard, and the other IR field/source guards; no filter was used.

## Report isolation

After all proof runs, both `git diff --exit-code` and `git status --porcelain=v1` were empty. Created the new branch `proof/lower-bzr-5-report` from the tested HEAD for this report. Only `proof/lower-bzr-5/report.md` is to be committed and pushed to that branch on origin. No compiler, test, package, submodule, or mutant change is retained in the report commit.

## Complete captured test output

The following logs are verbatim, including the complete got diagnostics. Escaped newlines in quoted stdout/stderr preserve complete strings, including trailing newlines.

### Baseline oracle pins — exit 0

```text
=== RUN   TestRecordSpreadFromMissingAgreesWithNode
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode
=== CONT  TestRecordSpreadFromMissingAgreesWithNode
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.a
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/null_args.a
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 334c4a0880b3 hit 222.87
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 334c4a0880b3 hit 222.87
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 334c4a0880b3 hit 222.87
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 334c4a0880b3 hit 222.87
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 334c4a0880b3 miss 222.87
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 334c4a0880b3 hit 0.03
--- PASS: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (39.41s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (263.95s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (263.96s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (263.96s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (224.56s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (263.98s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (1.55s)
PASS
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=13
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	265.525s
```

### Baseline lower pins — exit 0

```text
=== RUN   TestRecordSpreadFromMissingFixedObjectStops
=== PAUSE TestRecordSpreadFromMissingFixedObjectStops
=== CONT  TestRecordSpreadFromMissingFixedObjectStops
=== RUN   TestRecordSpreadFromMissingFixedObjectStops/.a
=== PAUSE TestRecordSpreadFromMissingFixedObjectStops/.a
=== RUN   TestRecordSpreadFromMissingFixedObjectStops/.ts
=== PAUSE TestRecordSpreadFromMissingFixedObjectStops/.ts
=== CONT  TestRecordSpreadFromMissingFixedObjectStops/.a
=== CONT  TestRecordSpreadFromMissingFixedObjectStops/.ts
--- PASS: TestRecordSpreadFromMissingFixedObjectStops (0.00s)
    --- PASS: TestRecordSpreadFromMissingFixedObjectStops/.ts (0.08s)
    --- PASS: TestRecordSpreadFromMissingFixedObjectStops/.a (0.08s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.096s
```

### Mutant 1 original run — exit 1

```text
=== RUN   TestRecordSpreadFromMissingAgreesWithNode
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode
=== CONT  TestRecordSpreadFromMissingAgreesWithNode
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.a
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/null_args.a
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    record_spread_missing_test.go:40: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__531ab0e390c0a6eb20d29abc43e652d8_2 = adamic_record_iterate(adamic_local_115732354606622155727914530510908772032_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    record_spread_missing_test.go:40: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__9566534b6660ba7451f22b2adb2dc5e2_2 = adamic_record_iterate(adamic_local_74212821344202605624165413379636541401_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:40: native: exit codes differ, stderr ""
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: native: exit codes differ, stderr ""
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:40: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types fb2a908a27a2 hit 25.55
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types fb2a908a27a2 hit 25.55
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types fb2a908a27a2 miss 25.55
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types fb2a908a27a2 hit 25.55
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (6.22s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (6.27s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (6.35s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (26.70s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (32.92s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (32.94s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (26.68s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=11
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	32.978s
FAIL
```

### Mutant 1 complete-outcome diagnostic replay — exit 1

```text
=== RUN   TestRecordSpreadFromMissingAgreesWithNode
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode
=== CONT  TestRecordSpreadFromMissingAgreesWithNode
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.a
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/null_args.a
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    record_spread_missing_test.go:40: FULL Node want: stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit_code=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    record_spread_missing_test.go:40: FULL Node want: stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit_code=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    record_spread_missing_test.go:40: FULL Node want: stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit_code=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    record_spread_missing_test.go:40: FULL Node want: stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit_code=0
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    record_spread_missing_test.go:40: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__9566534b6660ba7451f22b2adb2dc5e2_2 = adamic_record_iterate(adamic_local_74212821344202605624165413379636541401_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    record_spread_missing_test.go:40: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__531ab0e390c0a6eb20d29abc43e652d8_2 = adamic_record_iterate(adamic_local_115732354606622155727914530510908772032_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:40: FULL Node want: stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit_code=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    record_spread_missing_test.go:40: FULL native got: stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit_code=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: FULL Node want: stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit_code=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    record_spread_missing_test.go:40: FULL native got: stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit_code=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:40: FULL native got: stdout="" stderr="" exit_code=-1
    record_spread_missing_test.go:40: native: exit codes differ, stderr ""
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: FULL native got: stdout="" stderr="" exit_code=-1
    record_spread_missing_test.go:40: native: exit codes differ, stderr ""
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    record_spread_missing_test.go:40: FULL sanitized got: stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit_code=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    record_spread_missing_test.go:40: FULL sanitized got: stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit_code=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:40: FULL sanitized got: stdout="" stderr="record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n" exit_code=1
    record_spread_missing_test.go:40: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: FULL sanitized got: stdout="" stderr="record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n" exit_code=1
    record_spread_missing_test.go:40: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types fb2a908a27a2 hit 0.53
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types fb2a908a27a2 hit 0.55
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types fb2a908a27a2 hit 0.57
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types fb2a908a27a2 hit 0.59
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    record_spread_missing_test.go:40: FULL javascript got: stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit_code=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:40: FULL javascript got: stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit_code=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    record_spread_missing_test.go:40: FULL javascript got: stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit_code=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: FULL javascript got: stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit_code=0
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (6.61s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (6.76s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (6.80s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (8.30s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (1.71s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (8.33s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (1.60s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=11
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	8.388s
FAIL
```

### Mutant 2 original run — exit 1

```text
=== RUN   TestRecordSpreadFromMissingAgreesWithNode
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode
=== CONT  TestRecordSpreadFromMissingAgreesWithNode
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.a
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/null_args.a
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/undefined_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/null_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/overwritten_key.a:8:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
    record_spread_missing_test.go:37: got /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:21:34: stage 0 can't lower a spread into record storage from a fixed object yet, want the spread lowered and JSON.stringify of object references to stop at tracing_write_stack_event.ts:28:37: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/undefined_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/overwritten_key.ts:8:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/null_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (5.75s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (5.77s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (5.87s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (5.92s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (5.92s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (0.30s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (0.29s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=7
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	6.078s
FAIL
```

### IR suite — exit 0

```text
ok  	github.com/system-inc/adamic/internal/ir	106.924s
```
