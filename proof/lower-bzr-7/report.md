# Compiler proof: record spread from a missing source (#bzrdem0)

All requested baseline checks passed. Both specified mutants were caught. The complete admission delta passed: 6 newly accepted, 0 newly refused, 7 changed programs, no void reasons and no timeouts. The six new admissions agree with Node. Tracing advances past the spread to the later JSON.stringify stop. This is finite-corpus evidence, not a claim about all TypeScript programs.

- Tested HEAD: `529d9da9ec7d488a3ca0cc88861656834c7ffee9` (detached), fetched from `origin/compiler/record-spread-undefined-7`.
- Base/main: `0c7052759671347775615af5b4a8028dd8688af8`; verified as HEAD’s sole parent.
- Report destination: `proof/lower-bzr-7/report.md`; publication branch: `proof/lower-bzr-7-report`, created from tested HEAD. Only this report is to be committed and pushed.
- Date: 2026-11-12 UTC.

## Decisive results

| Check | Exit | Wall time | Result |
|---|---:|---:|---|
| gofmt | 0 | — | Empty output |
| go vet | 0 | — | No diagnostics |
| HEAD uncached oracle pin | 0 | 307 s | All 7 subtests pass |
| HEAD lower pin | 0 | 30 s | Both fixed-object subtests pass |
| Mutant 1 oracle / lower | 1 / 0 | 73.152 / 27.897 s | Oracle catches 4 fixtures |
| Mutant 2 oracle / lower | 1 / 0 | 57.850 / 29.004 s | Oracle catches all 7 fixtures |
| Full lower | 0 | 1790 s (29m 50s) | No failures |
| Full IR | 0 | 149 s | No failures |
| Admission delta, workers=2 | 0 | 1354 s (22m 34s) | pass, not void |

## Setup and step 1

Loaded `/workspace/adamic-tools/env.sh` before repository work. Its TMPDIR was `/tmp/adamic-gate`; created that directory outside the repository before Go commands. No AGENTS.md files were found under /workspace. Ran the requested fetch, detached checkout, and recursive submodule initialization successfully. Working tree was clean before checks and after each restoration.

```sh
git fetch origin compiler/record-spread-undefined-7
git checkout --detach 529d9da9ec7d488a3ca0cc88861656834c7ffee9
git submodule update --init --recursive
gofmt -l internal/lower internal/oracle
go vet ./internal/lower ./internal/oracle
```

Submodule pins: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`.

Provisioned a separate upstream checkout from https://github.com/microsoft/TypeScript.git at `050880ce59e30b356b686bd3144efe24f875ebc8` in `/tmp/adamic-proof/typescript`. Copied stage3/api’s package.json and package-lock.json to `/tmp/adamic-proof/api` and ran `npm ci --prefix /tmp/adamic-proof/api`: 3 packages installed. No npm dependencies were installed in the tested repository. The gate provisioned its own temporary checkout outside it.

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-proof/typescript
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
export PATH=/tmp/adamic-proof/api/node_modules/.bin:$PATH
```

These exports were inherited by npm, all proof commands, the admission gate, and its children. TLS verification stayed enabled. `tsc --version`: Version 6.0.3; @types/node pin: 25.3.3. Toolchain: Go go1.27.1 linux/amd64; Node v24.19.0; clang 20.1.8. One heavy command ran at a time. Admission ran alone; no tests, builds, installs, or other independent workloads overlapped it.

`gofmt -l` printed nothing; vet exited 0 and printed nothing.

## Step 2: baseline pins and exact tracing stop

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestRecordSpreadFromMissing' -count=1 -v -timeout 40m
go test ./internal/lower -run 'TestRecordSpreadFromMissing' -count=1 -v -timeout 30m
```

| Subtest (oracle parent: TestRecordSpreadFromMissingAgreesWithNode) | HEAD | Mutant 1 | Mutant 2 |
|---|---|---|---|
| null_args.a | PASS | FAIL | FAIL |
| null_args.ts | PASS | FAIL | FAIL |
| overwritten_key.a | PASS | PASS | FAIL |
| overwritten_key.ts | PASS | PASS | FAIL |
| tracing_write_stack_event.ts | PASS | PASS | FAIL |
| undefined_args.a | PASS | FAIL | FAIL |
| undefined_args.ts | PASS | FAIL | FAIL |

Lower parent `TestRecordSpreadFromMissingFixedObjectStops`: `.a` PASS and `.ts` PASS on HEAD and on each mutant. The parent oracle fails on both mutants; the parent lower passes throughout.

Tracing on HEAD: PASS. The Go test pins this exact substring (including its trailing space):
```text
"tracing_write_stack_event.ts:28:37: stage 0 can't lower JSON.stringify "
```

The complete observed CLI diagnostic is:
```text
adamic: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:28:37: stage 0 can't lower JSON.stringify recursive array types yet
exit status 1
```

The CLI intentionally exits 1 for this refusal (`go run` also prints `exit status 1`); this is the expected stop, not a failed proof. The baseline oracle test itself exits 0. No baseline FAIL exists, so no baseline got/want failure is omitted.

## Step 3: mutants, one at a time

Each edit was applied only after asserting its old text occurred exactly once in internal/lower/records.go. Both step-2 commands were rerun exactly, with oracle uncached. Source was restored and `git diff --exit-code` passed (exit 0, no output) after each mutant. Supplemental diagnostic runs were also restored and checked.

### Mutant 1

```text
Old text occurs exactly once.
result.Spread = ir.Coalesce{Value: result.Spread, Fallback: ir.RecordLiteral{Element: of}, Of: ir.Record}
=>
result.Spread = ir.Coalesce{Value: result.Spread, Fallback: ir.RecordLiteral{Element: of}, Of: ir.Record}.Value
```

Oracle: exit=1 wall_seconds=73.152. Lower: exit=0 wall_seconds=27.897. Restoration: exit=0 wall_seconds=0.418.

Red subtests: null_args.a, null_args.ts, undefined_args.a, undefined_args.ts. Both null fixtures fail native C compilation with the whole clang diagnostic reproduced in the verbatim log below; the fatal build error prevents their later sanitizer/JavaScript checks from running. Both undefined fixtures fail native execution and UBSan; their JavaScript comparisons complete without errors. Both overwritten-key fixtures pass. Tracing passes at the original JSON.stringify stop.

For **each** of undefined_args.a and undefined_args.ts, complete got/want process values (decoded from the supplemental diagnostic):
```text
want Node: exitCode=0, stdout="results\n1\n3\nresults\n1\n{}\n", stderr=""
got native: exitCode=-1, stdout="", stderr=""
got sanitized: exitCode=1, stdout="", stderr="record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
```

For **each** of null_args.a and null_args.ts, want successful lowering, native compilation and exit 0 matching Node; Node’s whole stdout is:
```text
results
1
{"results":5}
first,results,last
3
{"first":1,"results":6,"last":2}
```

Node stderr is empty. Got is the complete clang compilation refusal in the verbatim log below (no native process result exists). The actual red differs from the mutants.md prediction for null: it stops at C compilation, before execution.

Exact tracing CLI stop:
```text
adamic: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:28:37: stage 0 can't lower JSON.stringify recursive array types yet
exit status 1
```

Verbatim required oracle run (every red subtest and whole emitted diagnostic):
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
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    record_spread_missing_test.go:40: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__531ab0e390c0a6eb20d29abc43e652d8_2 = adamic_record_iterate(adamic_local_115732354606622155727914530510908772032_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1157:62: note: passing argument to parameter 'record' here
         1157 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: native: exit codes differ, stderr ""
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    record_spread_missing_test.go:40: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__9566534b6660ba7451f22b2adb2dc5e2_2 = adamic_record_iterate(adamic_local_74212821344202605624165413379636541401_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1157:62: note: passing argument to parameter 'record' here
         1157 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:40: native: exit codes differ, stderr ""
    record_spread_missing_test.go:40: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f24df64aec70 miss 30.68
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f24df64aec70 hit 30.68
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f24df64aec70 hit 30.68
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f24df64aec70 hit 30.57
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (7.52s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (7.55s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (0.47s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (39.39s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (39.40s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (39.40s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (31.86s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=11
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	39.440s
FAIL
```

Verbatim required lower run:
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
    --- PASS: TestRecordSpreadFromMissingFixedObjectStops/.a (0.09s)
    --- PASS: TestRecordSpreadFromMissingFixedObjectStops/.ts (0.09s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.103s
```

Supplemental full got/want capture:
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
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: FULL WANT Node: oracle.run{stdout:[]uint8{0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x33, 0xa, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x7b, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:40: FULL WANT Node: oracle.run{stdout:[]uint8{0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x33, 0xa, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x7b, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
    record_spread_missing_test.go:37: EXACT STOP got: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:28:37: stage 0 can't lower JSON.stringify recursive array types yet
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    record_spread_missing_test.go:40: FULL WANT Node: oracle.run{stdout:[]uint8{0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x35, 0x7d, 0xa, 0x66, 0x69, 0x72, 0x73, 0x74, 0x2c, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x2c, 0x6c, 0x61, 0x73, 0x74, 0xa, 0x33, 0xa, 0x7b, 0x22, 0x66, 0x69, 0x72, 0x73, 0x74, 0x22, 0x3a, 0x31, 0x2c, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x36, 0x2c, 0x22, 0x6c, 0x61, 0x73, 0x74, 0x22, 0x3a, 0x32, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    record_spread_missing_test.go:40: FULL WANT Node: oracle.run{stdout:[]uint8{0x66, 0x69, 0x72, 0x73, 0x74, 0x2c, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x2c, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x33, 0xa, 0x33, 0xa, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x2c, 0x6c, 0x61, 0x73, 0x74, 0xa, 0x32, 0xa, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x34, 0x2c, 0x22, 0x6c, 0x61, 0x73, 0x74, 0x22, 0x3a, 0x32, 0x7d, 0xa, 0x7b, 0x22, 0x66, 0x69, 0x72, 0x73, 0x74, 0x22, 0x3a, 0x31, 0x2c, 0x22, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x39, 0x7d, 0xa, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x39, 0x2c, 0x22, 0x6c, 0x61, 0x73, 0x74, 0x22, 0x3a, 0x32, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    record_spread_missing_test.go:40: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__531ab0e390c0a6eb20d29abc43e652d8_2 = adamic_record_iterate(adamic_local_115732354606622155727914530510908772032_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1157:62: note: passing argument to parameter 'record' here
         1157 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    record_spread_missing_test.go:40: FULL WANT Node: oracle.run{stdout:[]uint8{0x66, 0x69, 0x72, 0x73, 0x74, 0x2c, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x2c, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x33, 0xa, 0x33, 0xa, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x2c, 0x6c, 0x61, 0x73, 0x74, 0xa, 0x32, 0xa, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x34, 0x2c, 0x22, 0x6c, 0x61, 0x73, 0x74, 0x22, 0x3a, 0x32, 0x7d, 0xa, 0x7b, 0x22, 0x66, 0x69, 0x72, 0x73, 0x74, 0x22, 0x3a, 0x31, 0x2c, 0x22, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x39, 0x7d, 0xa, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x39, 0x2c, 0x22, 0x6c, 0x61, 0x73, 0x74, 0x22, 0x3a, 0x32, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:40: FULL GOT: oracle.run{stdout:[]uint8{}, stderr:[]uint8{}, exitCode:-1}; FULL WANT Node: oracle.run{stdout:[]uint8{0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x33, 0xa, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x7b, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
    record_spread_missing_test.go:40: native: exit codes differ, stderr ""
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: FULL GOT: oracle.run{stdout:[]uint8{}, stderr:[]uint8{}, exitCode:-1}; FULL WANT Node: oracle.run{stdout:[]uint8{0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x33, 0xa, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x7b, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
    record_spread_missing_test.go:40: native: exit codes differ, stderr ""
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    record_spread_missing_test.go:40: FULL WANT Node: oracle.run{stdout:[]uint8{0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x35, 0x7d, 0xa, 0x66, 0x69, 0x72, 0x73, 0x74, 0x2c, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x2c, 0x6c, 0x61, 0x73, 0x74, 0xa, 0x33, 0xa, 0x7b, 0x22, 0x66, 0x69, 0x72, 0x73, 0x74, 0x22, 0x3a, 0x31, 0x2c, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x36, 0x2c, 0x22, 0x6c, 0x61, 0x73, 0x74, 0x22, 0x3a, 0x32, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
    record_spread_missing_test.go:40: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__9566534b6660ba7451f22b2adb2dc5e2_2 = adamic_record_iterate(adamic_local_74212821344202605624165413379636541401_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1157:62: note: passing argument to parameter 'record' here
         1157 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:40: FULL GOT: oracle.run{stdout:[]uint8{}, stderr:[]uint8{0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x2e, 0x63, 0x3a, 0x33, 0x34, 0x3a, 0x31, 0x37, 0x3a, 0x20, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x3a, 0x20, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x20, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x20, 0x77, 0x69, 0x74, 0x68, 0x69, 0x6e, 0x20, 0x6e, 0x75, 0x6c, 0x6c, 0x20, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x20, 0x6f, 0x66, 0x20, 0x74, 0x79, 0x70, 0x65, 0x20, 0x27, 0x63, 0x6f, 0x6e, 0x73, 0x74, 0x20, 0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x5f, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x27, 0x20, 0x28, 0x61, 0x6b, 0x61, 0x20, 0x27, 0x63, 0x6f, 0x6e, 0x73, 0x74, 0x20, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x20, 0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x5f, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x27, 0x29, 0xa, 0x53, 0x55, 0x4d, 0x4d, 0x41, 0x52, 0x59, 0x3a, 0x20, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x42, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x53, 0x61, 0x6e, 0x69, 0x74, 0x69, 0x7a, 0x65, 0x72, 0x3a, 0x20, 0x75, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x2d, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x20, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x2e, 0x63, 0x3a, 0x33, 0x34, 0x3a, 0x31, 0x37, 0x20, 0xa}, exitCode:1}; FULL WANT Node: oracle.run{stdout:[]uint8{0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x33, 0xa, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x7b, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
    record_spread_missing_test.go:40: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: FULL GOT: oracle.run{stdout:[]uint8{}, stderr:[]uint8{0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x2e, 0x63, 0x3a, 0x33, 0x34, 0x3a, 0x31, 0x37, 0x3a, 0x20, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x20, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x3a, 0x20, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x20, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x20, 0x77, 0x69, 0x74, 0x68, 0x69, 0x6e, 0x20, 0x6e, 0x75, 0x6c, 0x6c, 0x20, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x20, 0x6f, 0x66, 0x20, 0x74, 0x79, 0x70, 0x65, 0x20, 0x27, 0x63, 0x6f, 0x6e, 0x73, 0x74, 0x20, 0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x5f, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x27, 0x20, 0x28, 0x61, 0x6b, 0x61, 0x20, 0x27, 0x63, 0x6f, 0x6e, 0x73, 0x74, 0x20, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x20, 0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x5f, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x27, 0x29, 0xa, 0x53, 0x55, 0x4d, 0x4d, 0x41, 0x52, 0x59, 0x3a, 0x20, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x42, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x53, 0x61, 0x6e, 0x69, 0x74, 0x69, 0x7a, 0x65, 0x72, 0x3a, 0x20, 0x75, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x2d, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x20, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x2e, 0x63, 0x3a, 0x33, 0x34, 0x3a, 0x31, 0x37, 0x20, 0xa}, exitCode:1}; FULL WANT Node: oracle.run{stdout:[]uint8{0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x33, 0xa, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x7b, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
    record_spread_missing_test.go:40: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f24df64aec70 hit 0.91
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f24df64aec70 hit 0.91
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f24df64aec70 hit 0.93
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types f24df64aec70 hit 1.01
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (10.11s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (10.46s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (0.83s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (2.53s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (12.65s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (12.72s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (12.77s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=11
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	12.810s
FAIL
```

### Mutant 2

```text
Old text occurs exactly once.
present = l.checker.GetNonNullableType(own)
=>
present = own
```

Oracle: exit=1 wall_seconds=57.850. Lower: exit=0 wall_seconds=29.004. Restoration: exit=0 wall_seconds=0.486.

All seven oracle subtests are red. For null_args.a/.ts and undefined_args.a/.ts, got a located spread refusal at 7:11; for overwritten_key.a/.ts at 8:11. The whole text for each appears below. Want: no lowering error and native, sanitized and JavaScript execution matching the complete Node outputs in the supplemental log. No backend process is started after the lowering refusal.

Tracing FAIL, whole got and want (the want ends with a space):
```text
got /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:21:34: stage 0 can't lower a spread into record storage from a fixed object yet, want the spread lowered and JSON.stringify to stop at tracing_write_stack_event.ts:28:37: stage 0 can't lower JSON.stringify 
```

Exact tracing CLI stop:
```text
adamic: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:21:34: stage 0 can't lower a spread into record storage from a fixed object yet
exit status 1
```

Verbatim required oracle run (every red subtest and whole emitted diagnostic):
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
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/undefined_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/undefined_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/null_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/overwritten_key.a:8:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
    record_spread_missing_test.go:37: got /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:21:34: stage 0 can't lower a spread into record storage from a fixed object yet, want the spread lowered and JSON.stringify to stop at tracing_write_stack_event.ts:28:37: stage 0 can't lower JSON.stringify 
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/overwritten_key.ts:8:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/null_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (9.20s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (9.31s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (9.36s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (9.38s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (9.41s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (0.35s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (0.32s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=7
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	9.665s
FAIL
```

Verbatim required lower run:
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
    --- PASS: TestRecordSpreadFromMissingFixedObjectStops/.ts (0.25s)
    --- PASS: TestRecordSpreadFromMissingFixedObjectStops/.a (0.25s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.355s
```

Supplemental full got/want capture:
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
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: FULL WANT Node: oracle.run{stdout:[]uint8{0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x33, 0xa, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x7b, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    record_spread_missing_test.go:40: FULL WANT Node: oracle.run{stdout:[]uint8{0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x35, 0x7d, 0xa, 0x66, 0x69, 0x72, 0x73, 0x74, 0x2c, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x2c, 0x6c, 0x61, 0x73, 0x74, 0xa, 0x33, 0xa, 0x7b, 0x22, 0x66, 0x69, 0x72, 0x73, 0x74, 0x22, 0x3a, 0x31, 0x2c, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x36, 0x2c, 0x22, 0x6c, 0x61, 0x73, 0x74, 0x22, 0x3a, 0x32, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    record_spread_missing_test.go:40: FULL WANT Node: oracle.run{stdout:[]uint8{0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x35, 0x7d, 0xa, 0x66, 0x69, 0x72, 0x73, 0x74, 0x2c, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x2c, 0x6c, 0x61, 0x73, 0x74, 0xa, 0x33, 0xa, 0x7b, 0x22, 0x66, 0x69, 0x72, 0x73, 0x74, 0x22, 0x3a, 0x31, 0x2c, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x36, 0x2c, 0x22, 0x6c, 0x61, 0x73, 0x74, 0x22, 0x3a, 0x32, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:40: FULL WANT Node: oracle.run{stdout:[]uint8{0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x33, 0xa, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x31, 0xa, 0x7b, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/undefined_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/null_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
    record_spread_missing_test.go:37: EXACT STOP got: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:21:34: stage 0 can't lower a spread into record storage from a fixed object yet
    record_spread_missing_test.go:37: got /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:21:34: stage 0 can't lower a spread into record storage from a fixed object yet, want the spread lowered and JSON.stringify to stop at tracing_write_stack_event.ts:28:37: stage 0 can't lower JSON.stringify 
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/null_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/undefined_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    record_spread_missing_test.go:40: FULL WANT Node: oracle.run{stdout:[]uint8{0x66, 0x69, 0x72, 0x73, 0x74, 0x2c, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x2c, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x33, 0xa, 0x33, 0xa, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x2c, 0x6c, 0x61, 0x73, 0x74, 0xa, 0x32, 0xa, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x34, 0x2c, 0x22, 0x6c, 0x61, 0x73, 0x74, 0x22, 0x3a, 0x32, 0x7d, 0xa, 0x7b, 0x22, 0x66, 0x69, 0x72, 0x73, 0x74, 0x22, 0x3a, 0x31, 0x2c, 0x22, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x39, 0x7d, 0xa, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x39, 0x2c, 0x22, 0x6c, 0x61, 0x73, 0x74, 0x22, 0x3a, 0x32, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    record_spread_missing_test.go:40: FULL WANT Node: oracle.run{stdout:[]uint8{0x66, 0x69, 0x72, 0x73, 0x74, 0x2c, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x2c, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0xa, 0x33, 0xa, 0x33, 0xa, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x2c, 0x6c, 0x61, 0x73, 0x74, 0xa, 0x32, 0xa, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x34, 0x2c, 0x22, 0x6c, 0x61, 0x73, 0x74, 0x22, 0x3a, 0x32, 0x7d, 0xa, 0x7b, 0x22, 0x66, 0x69, 0x72, 0x73, 0x74, 0x22, 0x3a, 0x31, 0x2c, 0x22, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x39, 0x7d, 0xa, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x22, 0x3a, 0x39, 0x2c, 0x22, 0x6c, 0x61, 0x73, 0x74, 0x22, 0x3a, 0x32, 0x7d, 0xa}, stderr:[]uint8{}, exitCode:0}
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/overwritten_key.a:8:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    record_spread_missing_test.go:40: /workspace/adamic/internal/oracle/testdata/record_spread_missing/overwritten_key.ts:8:11: stage 0 can't lower a spread into record storage from a fixed object yet
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (7.59s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (7.62s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (7.73s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (7.74s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (7.75s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (0.66s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (0.66s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=7
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	8.303s
FAIL
```

Supplemental diagnostics used `go test -overlay=/tmp/adamic-proof/overlay.json ./internal/oracle -run TestRecordSpreadFromMissing -count=1 -v -timeout 40m` with ADAMIC_GATE_UNCACHED=1. The external overlay added logs for full oracle.run values and the tracing stop; repository test files were not edited. The first supplemental attempt failed to build because the logging overlay used an out-of-scope variable (`undefined: candidate`); corrected the overlay, restored the mutant, and repeated successfully. This was a proof-harness error, not a compiler result. The exact required runs above were unaffected. Supplemental reruns confirmed the same red subtests.

### Readable full expected process values for mutant 2

All six non-tracing red fixtures want no lowering error and backend process results equal to Node. Each got the complete spread refusal shown below and therefore produced no backend process result. The Node values are the same fixture outputs also captured byte-for-byte by the diagnostic overlay.
**internal/oracle/testdata/record_spread_missing/null_args.a**

```json
{
  "got_lowering_error": "/workspace/adamic/internal/oracle/testdata/record_spread_missing/null_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet",
  "want_lowering_error": null,
  "want_process": {
    "exitCode": 0,
    "stdout": "results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n",
    "stderr": ""
  }
}
```

**internal/oracle/testdata/record_spread_missing/null_args.ts**

```json
{
  "got_lowering_error": "/workspace/adamic/internal/oracle/testdata/record_spread_missing/null_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet",
  "want_lowering_error": null,
  "want_process": {
    "exitCode": 0,
    "stdout": "results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n",
    "stderr": ""
  }
}
```

**internal/oracle/testdata/record_spread_missing/overwritten_key.a**

```json
{
  "got_lowering_error": "/workspace/adamic/internal/oracle/testdata/record_spread_missing/overwritten_key.a:8:11: stage 0 can't lower a spread into record storage from a fixed object yet",
  "want_lowering_error": null,
  "want_process": {
    "exitCode": 0,
    "stdout": "first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n",
    "stderr": ""
  }
}
```

**internal/oracle/testdata/record_spread_missing/overwritten_key.ts**

```json
{
  "got_lowering_error": "/workspace/adamic/internal/oracle/testdata/record_spread_missing/overwritten_key.ts:8:11: stage 0 can't lower a spread into record storage from a fixed object yet",
  "want_lowering_error": null,
  "want_process": {
    "exitCode": 0,
    "stdout": "first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n",
    "stderr": ""
  }
}
```

**internal/oracle/testdata/record_spread_missing/undefined_args.a**

```json
{
  "got_lowering_error": "/workspace/adamic/internal/oracle/testdata/record_spread_missing/undefined_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet",
  "want_lowering_error": null,
  "want_process": {
    "exitCode": 0,
    "stdout": "results\n1\n3\nresults\n1\n{}\n",
    "stderr": ""
  }
}
```

**internal/oracle/testdata/record_spread_missing/undefined_args.ts**

```json
{
  "got_lowering_error": "/workspace/adamic/internal/oracle/testdata/record_spread_missing/undefined_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet",
  "want_lowering_error": null,
  "want_process": {
    "exitCode": 0,
    "stdout": "results\n1\n3\nresults\n1\n{}\n",
    "stderr": ""
  }
}
```

## Step 4: full lower suite

```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

Exit 0. Wall time 1790 seconds. Every failing test: **none**. Environment failures: **none**. Whole log:
```text
ok  	github.com/system-inc/adamic/internal/lower	1786.164s
```

## Step 4b: full IR suite

```sh
go test ./internal/ir -count=1 -timeout 30m
```

Exit 0. Wall time 149 seconds. Every failure: **none**. This full unfiltered run includes TestCallTargetReadersLower, its lower export/archive inputs, the IR field readers and iterator-next source guards. Whole log:
```text
ok  	github.com/system-inc/adamic/internal/ir	110.204s
```

## Step 6: admission delta

```sh
python3 cloud/admission-corpus/manifest.py --sha 529d9da9ec7d488a3ca0cc88861656834c7ffee9 > /tmp/adamic-proof/manifest.json
go run ./cmd/adamic-admission-delta --base 0c7052759671347775615af5b4a8028dd8688af8 --head 529d9da9ec7d488a3ca0cc88861656834c7ffee9 --manifest /tmp/adamic-proof/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --workers 2 --json > /tmp/adamic-proof/admission.json 2> /tmp/adamic-proof/admission.stderr
```

Ran alone, with no concurrent independent workload; no corpus filter, shard or budget. Default compile timeout 45 seconds and runtime timeout 10 seconds. Manifest: witnesses=2, fixtures=4360, gaps=105, review=20, fuzz=0, stage3=821. 5,308 distinct classified programs.

- Exit: 0.
- Wall time: 1354 seconds (22m 34s); gate internal total: 1342.987206198 seconds.
- Verdict: **pass**.
- Void: **false**; global and program void reasons: **none**.
- diff_count: **7**.
- Admitted: **6**, all agree with Node.
- Newly refused: **0**.
- Accepted by both: **3319**; refused by both: **1983**.
- Classification errors/timeouts: **none**. Runtime timeouts: **0**; retries: **0**; still timed out: **0**. No timeout exists to rerun at either SHA.

Gate stderr:
```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 7
```

Phase seconds:
```json
{
  "base_build": 55.906286112,
  "checkout": 17.31422395,
  "checkout_and_provision": 18.388257453,
  "classification": 746.966782116,
  "head_build": 29.877133023,
  "manifest_and_verification": 22.669076763,
  "provision": 1.073993596,
  "runtime": 469.16471905,
  "total": 1342.987206198
}
```

### Every program with different classification stderr

The following are the literal complete stderr strings, JSON-quoted to preserve empty strings and final newlines. The temporary path is part of the observed text.

**internal/oracle/testdata/record_spread_missing/null_args.a** — newly-accepted
```text
base: "adamic: /tmp/adamic-gate/admission-delta-2149236228/head/internal/oracle/testdata/record_spread_missing/null_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet\n"
head: ""
```

Runtime agreement: true.
```json
{
  "node": {
    "wall_seconds": 0.232606941,
    "exit": 0,
    "stdout": "results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.057649056,
    "exit": 0,
    "stdout": "results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n",
    "stderr": ""
  },
  "native": {
    "wall_seconds": 0.002135732,
    "exit": 0,
    "stdout": "results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n",
    "stderr": ""
  },
  "agree": true
}
```

**internal/oracle/testdata/record_spread_missing/null_args.ts** — newly-accepted
```text
base: "adamic: /tmp/adamic-gate/admission-delta-2149236228/head/internal/oracle/testdata/record_spread_missing/null_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet\n"
head: ""
```

Runtime agreement: true.
```json
{
  "node": {
    "wall_seconds": 0.15699902,
    "exit": 0,
    "stdout": "results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.05966051,
    "exit": 0,
    "stdout": "results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n",
    "stderr": ""
  },
  "native": {
    "wall_seconds": 0.002558007,
    "exit": 0,
    "stdout": "results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n",
    "stderr": ""
  },
  "agree": true
}
```

**internal/oracle/testdata/record_spread_missing/overwritten_key.a** — newly-accepted
```text
base: "adamic: /tmp/adamic-gate/admission-delta-2149236228/head/internal/oracle/testdata/record_spread_missing/overwritten_key.a:8:11: stage 0 can't lower a spread into record storage from a fixed object yet\n"
head: ""
```

Runtime agreement: true.
```json
{
  "node": {
    "wall_seconds": 0.119087819,
    "exit": 0,
    "stdout": "first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.054631213,
    "exit": 0,
    "stdout": "first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n",
    "stderr": ""
  },
  "native": {
    "wall_seconds": 0.004241057,
    "exit": 0,
    "stdout": "first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n",
    "stderr": ""
  },
  "agree": true
}
```

**internal/oracle/testdata/record_spread_missing/overwritten_key.ts** — newly-accepted
```text
base: "adamic: /tmp/adamic-gate/admission-delta-2149236228/head/internal/oracle/testdata/record_spread_missing/overwritten_key.ts:8:11: stage 0 can't lower a spread into record storage from a fixed object yet\n"
head: ""
```

Runtime agreement: true.
```json
{
  "node": {
    "wall_seconds": 0.163826433,
    "exit": 0,
    "stdout": "first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.142052502,
    "exit": 0,
    "stdout": "first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n",
    "stderr": ""
  },
  "native": {
    "wall_seconds": 0.003151879,
    "exit": 0,
    "stdout": "first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n",
    "stderr": ""
  },
  "agree": true
}
```

**internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts** — refused-by-both
```text
base: "adamic: /tmp/adamic-gate/admission-delta-2149236228/head/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:21:34: stage 0 can't lower a spread into record storage from a fixed object yet\n"
head: "adamic: /tmp/adamic-gate/admission-delta-2149236228/head/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:28:37: stage 0 can't lower JSON.stringify recursive array types yet\n"
```

**internal/oracle/testdata/record_spread_missing/undefined_args.a** — newly-accepted
```text
base: "adamic: /tmp/adamic-gate/admission-delta-2149236228/head/internal/oracle/testdata/record_spread_missing/undefined_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet\n"
head: ""
```

Runtime agreement: true.
```json
{
  "node": {
    "wall_seconds": 0.187969152,
    "exit": 0,
    "stdout": "results\n1\n3\nresults\n1\n{}\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.074503663,
    "exit": 0,
    "stdout": "results\n1\n3\nresults\n1\n{}\n",
    "stderr": ""
  },
  "native": {
    "wall_seconds": 0.001982663,
    "exit": 0,
    "stdout": "results\n1\n3\nresults\n1\n{}\n",
    "stderr": ""
  },
  "agree": true
}
```

**internal/oracle/testdata/record_spread_missing/undefined_args.ts** — newly-accepted
```text
base: "adamic: /tmp/adamic-gate/admission-delta-2149236228/head/internal/oracle/testdata/record_spread_missing/undefined_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet\n"
head: ""
```

Runtime agreement: true.
```json
{
  "node": {
    "wall_seconds": 0.136052799,
    "exit": 0,
    "stdout": "results\n1\n3\nresults\n1\n{}\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.070665846,
    "exit": 0,
    "stdout": "results\n1\n3\nresults\n1\n{}\n",
    "stderr": ""
  },
  "native": {
    "wall_seconds": 0.002273054,
    "exit": 0,
    "stdout": "results\n1\n3\nresults\n1\n{}\n",
    "stderr": ""
  },
  "agree": true
}
```

### Provenance

```json
{
  "base": "0c7052759671347775615af5b4a8028dd8688af8",
  "head": "529d9da9ec7d488a3ca0cc88861656834c7ffee9",
  "generator_revision": "529d9da9ec7d488a3ca0cc88861656834c7ffee9",
  "generator_blob": "09d061a03b1e016387e975d25ee0cd5b75ac6c1f",
  "manifest_blob": "b0e6b5511969872f266a686e342fd8ae7af27d20",
  "complete_corpus": true,
  "workers": 2,
  "classification": "lowering-only",
  "diff": [
    {
      "path": "internal/oracle/testdata/record_spread_missing/null_args.a",
      "blob": "8501cc2f2e7fd882ff5de3d6a032a328256ee1eb"
    },
    {
      "path": "internal/oracle/testdata/record_spread_missing/null_args.ts",
      "blob": "8501cc2f2e7fd882ff5de3d6a032a328256ee1eb"
    },
    {
      "path": "internal/oracle/testdata/record_spread_missing/overwritten_key.a",
      "blob": "f3c6d5a3b5077249c5734dbfe8d89be3f8a9f55a"
    },
    {
      "path": "internal/oracle/testdata/record_spread_missing/overwritten_key.ts",
      "blob": "f3c6d5a3b5077249c5734dbfe8d89be3f8a9f55a"
    },
    {
      "path": "internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts",
      "blob": "49db094a11507875067930312da64dbf44f6b1c0"
    },
    {
      "path": "internal/oracle/testdata/record_spread_missing/undefined_args.a",
      "blob": "16fc8a95544b18e3e1876e61ebfd86d25456a6a2"
    },
    {
      "path": "internal/oracle/testdata/record_spread_missing/undefined_args.ts",
      "blob": "16fc8a95544b18e3e1876e61ebfd86d25456a6a2"
    }
  ],
  "diff_ts_unmatched": [],
  "sampling_seed": "4359e4846b77bc33f532e5e8d63a79b26bc79ffb85dd80b5eb38b399d260b668",
  "sampling_size": 121,
  "omitted": 0,
  "budget_seconds": 0
}
```

External admission JSON SHA-256: `5400fb8aa3934969c9d41b305f323bd14c276c3a80ef699d6a8a953ecffe6ae4`.

## Baseline verbose evidence

Oracle:
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
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 57a029ded0ab hit 240.77
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 57a029ded0ab hit 240.77
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 57a029ded0ab miss 240.77
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 57a029ded0ab hit 240.78
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 57a029ded0ab hit 240.77
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 57a029ded0ab hit 0.03
--- PASS: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (40.10s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (282.38s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (282.46s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (242.39s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (282.51s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (282.56s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (2.14s)
PASS
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=13
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	284.555s
```

Lower:
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
    --- PASS: TestRecordSpreadFromMissingFixedObjectStops/.a (0.15s)
    --- PASS: TestRecordSpreadFromMissingFixedObjectStops/.ts (0.15s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.163s
```

## Repository integrity and publication

All compiler mutations were restored. `git diff --exit-code` and `git status --short` were clean after the proof and before writing this report. No compiler, fixture, manifest, package, or submodule changes are included. The report branch is based directly on tested HEAD; only proof/lower-bzr-7/report.md is committed, and only refs/heads/proof/lower-bzr-7-report is pushed to origin.

