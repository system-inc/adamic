# Compiler proof: lower-bzr-2 (#bzrdem0)

Measured 2026-11-11 (UTC). HEAD `b7fbcd0abebad8c985aee9ebacfdb99a308641eb`; base/main `16429a30452d027809f0ea85bc395479423e493e`. HEAD has exactly this one parent. No compiler changes are delivered with this report.

## Decisive results

- **Proof is not wholly green.** The requested uncached oracle pin fails only `tracing_write_stack_event.ts`: five TS2339 diagnostics at its array-element destructuring prevent loading, before the record spread is lowered. Main has the same loader diagnostics. The six other fixtures pass on native release, ASan/UBSan, and JavaScript.
- The fixed-object stop pin passes for both `.a` and `.ts` at HEAD and main, with exactly the same stop text.
- The sole specified mutant is killed by both `null_args` variants (native and sanitized C compile failures) and both `undefined_args` variants (native signal termination, UBSan failure). Both `overwritten_key` variants pass. Tracing is already red without the mutant and is not evidence of killing it.
- Full lower suite: **PASS**, exit 0; wall 1467.69s. Step 4b full IR suite: **PASS**, exit 0; wall 140.75s; no failing tests. Full oracle suite: **FAIL**, exit 1; wall 3776.67s; all failing names and first errors are listed below.
- Full admission delta: **void**, `diff_count=7`, `admitted=6`, newly refused **0**. Six classification stderr texts change from main’s refusal to empty HEAD stderr. Three native-build timeouts were retried by the gate; one persisted. All three affected programs were then rerun alone at **both SHAs**, with no timeout under the same 10-second bound. These isolated results do not rewrite the original full gate’s void verdict.

## Setup and step 1

Loaded `/workspace/adamic-tools/env.sh` before repository commands. It sets `TMPDIR=/tmp/adamic-gate`; that absent directory was created outside the repository before Go needed it. No missing-work-directory failure occurred.

```sh
git fetch origin compiler/record-spread-undefined-2
git checkout --detach b7fbcd0abebad8c985aee9ebacfdb99a308641eb
git submodule update --init --recursive
gofmt -l internal/lower internal/oracle
go vet ./internal/lower ./internal/oracle
```

All exited 0. `gofmt -l` printed nothing; vet printed nothing. Recursive submodule pins: cohere `5360753e19e4417675a90e92a1823bea66b13918`; its TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. No AGENTS.md instructions were found in the workspace scan.
The separate upstream checkout is `/tmp/bzrdem0-proof/TypeScript`, fetched from `https://github.com/microsoft/TypeScript.git`, detached at `050880ce59e30b356b686bd3144efe24f875ebc8`. Exported `ADAMIC_TYPESCRIPT_SOURCE` to that path. This is distinct from the compiler’s pinned checker submodule.
Copied only stage3/api’s package.json and package-lock.json into `/tmp/bzrdem0-proof/api`, then ran `npm ci --prefix /tmp/bzrdem0-proof/api`. Installed packages remain outside the repository: TypeScript 6.0.3, @types/node 25.3.3, and the lockfile’s undici-types dependency. Prepended the external node_modules/.bin to PATH; `tsc --version` returned `Version 6.0.3`. Go was go1.27.1; Node was v24.19.0.

```sh
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
```

These were exported for npm and proof commands. The admission tool filters child environments through internal/toolchain.Env, which drops these names. For admission and its isolated reruns, external pass-through wrappers for go, git, node, npm, clang, clang++, python3, and tsc re-exported them before execing the original tools. `ADAMIC_TOOLS=/tmp/bzrdem0-proof/tools` and that tools/bin led PATH for those runs. No repository code was changed for this. TLS verification stayed enabled.
Only one top-level heavy command ran at a time. The full lower suite completed before oracle started; oracle completed before admission began. Admission ran alone with `--workers 2`, no filter, no shard, and no budget limit. The gate’s own worker/backend processes and Go tests’ own children are part of their respective commands.

## Step 2: HEAD pins

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestRecordSpreadFromMissingAgreesWithNode' -count=1 -v -timeout 40m
go test ./internal/lower -run 'TestRecordSpreadFromMissingFixedObjectStops' -count=1 -v
```

Oracle: exit 1; wall 281.09s. Lower: exit 0; wall 27.42s.

| Oracle subtest | Result |
|---|---|
| `null_args.a` | PASS |
| `null_args.ts` | PASS |
| `overwritten_key.a` | PASS |
| `overwritten_key.ts` | PASS |
| `tracing_write_stack_event.ts` | FAIL |
| `undefined_args.a` | PASS |
| `undefined_args.ts` | PASS |

Lower subtests `.a` and `.ts`: **PASS**. Expected suffixes are `main.a:6:11: stage 0 can't lower a spread into record storage from a fixed object yet` and its `.ts` twin.

### Whole got and want for the oracle failure

Got: the five loader diagnostics in the complete pin log below. Want: source Node exit 0, successful lowering, and all three backends’ stdout, compared stderr, and exit matching Node. The failure occurred in `cache_test.go:297` during onNode’s import-discovery load. The pin launched neither Node nor any backend for this subtest. Consequently there are no observed runtime stdout/stderr streams for native, sanitized, or JavaScript; they are **not run**, not observed empty streams.
A separate direct source-Node invocation, without changing the fixture, succeeded:

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts
```

Exit 0; whole stdout:

```text
{"pid":1,"tid":1,"ph":"X","cat":"check","ts":0,"name":"checkSourceFile","dur":20000,"args":{"results":{"count":2,"nested":{"deep":true},"list":["a",1,null]}} keys results count 1}
{"pid":1,"tid":1,"ph":"X","cat":"emit","ts":5,"name":"emitJsFileOrBundle","dur":19995,"args":{"path":"x.ts"} keys path,results count 2}
{"pid":1,"tid":1,"ph":"X","cat":"bind","ts":7,"name":"bindSourceFile","dur":19993,"args":{"path":"y.ts","results":{"kind":1}} keys path,results count 2}
{"pid":1,"tid":1,"ph":"X","cat":"parse","ts":9,"name":"createSourceFile","dur":19991,"args":{} keys results count 1}
{"pid":1,"tid":1,"ph":"E","cat":"program","ts":20000,"name":"createProgram","args":{"configFilePath":"tsconfig.json"} keys configFilePath count 1}
```

Whole stderr (empty):

```text

```

Complete original oracle pin log, including the whole got and every subtest result:

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
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
    cache_test.go:297: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:10: error TS2339: Property 'phase' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:17: error TS2339: Property 'name' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:23: error TS2339: Property 'args' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:29: error TS2339: Property 'time' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:35: error TS2339: Property 'separateBeginAndEnd' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 5bfaa813e537 hit 212.51
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 5bfaa813e537 hit 212.51
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 5bfaa813e537 miss 212.51
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 5bfaa813e537 hit 212.51
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 5bfaa813e537 hit 212.51
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 5bfaa813e537 hit 0.02
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (39.39s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (214.35s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (253.74s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (253.77s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (253.80s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (253.83s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (1.14s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=12
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	254.906s
FAIL
```

Complete lower pin log:

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
    --- PASS: TestRecordSpreadFromMissingFixedObjectStops/.a (0.09s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.108s
```

## Step 2b: main control

Created `/tmp/bzrdem0-proof/main` as a detached worktree at the exact base SHA. Copied only `internal/lower/record_spread_missing_test.go` and the seven requested oracle fixture programs from HEAD. The worktree’s cohere path shares the already initialized submodule files through an external-worktree symlink; both revisions pin the same gitlink. No HEAD lowering implementation was copied.

```sh
go test ./internal/lower -run 'TestRecordSpreadFromMissingFixedObjectStops' -count=1 -v
```

exit 0; wall 227.22s. Both `.a` and `.ts`: **PASS**.
To obtain the full errors even on passing assertions, a second run added only t.Logf of the returned error to the copied test, then restored it byte-for-byte. It changed no assertions or compiler behavior. Complete diagnostic run:

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
=== NAME  TestRecordSpreadFromMissingFixedObjectStops/.a
    record_spread_missing_test.go:20: PROOF got: /tmp/adamic-gate/TestRecordSpreadFromMissingFixedObjectStops.a1354731247/001/main.a:6:11: stage 0 can't lower a spread into record storage from a fixed object yet
=== NAME  TestRecordSpreadFromMissingFixedObjectStops/.ts
    record_spread_missing_test.go:20: PROOF got: /tmp/adamic-gate/TestRecordSpreadFromMissingFixedObjectStops.ts2842854487/001/main.ts:6:11: stage 0 can't lower a spread into record storage from a fixed object yet
--- PASS: TestRecordSpreadFromMissingFixedObjectStops (0.00s)
    --- PASS: TestRecordSpreadFromMissingFixedObjectStops/.a (0.08s)
    --- PASS: TestRecordSpreadFromMissingFixedObjectStops/.ts (0.08s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.096s
```

For each fixture, used main’s **`go run ./cmd/adamic c <repository-relative path>`**. Every command exited 1 and stdout was empty. Six stop during lowering; tracing stops at type checking before lowering. Whole stderr for each:

### null_args.a

Command: `go run ./cmd/adamic c internal/oracle/testdata/record_spread_missing/null_args.a`; exit 1; wall 30.105s.

```text
adamic: /tmp/bzrdem0-proof/main/internal/oracle/testdata/record_spread_missing/null_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
exit status 1
```

### null_args.ts

Command: `go run ./cmd/adamic c internal/oracle/testdata/record_spread_missing/null_args.ts`; exit 1; wall 0.220s.

```text
adamic: /tmp/bzrdem0-proof/main/internal/oracle/testdata/record_spread_missing/null_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
exit status 1
```

### overwritten_key.a

Command: `go run ./cmd/adamic c internal/oracle/testdata/record_spread_missing/overwritten_key.a`; exit 1; wall 0.275s.

```text
adamic: /tmp/bzrdem0-proof/main/internal/oracle/testdata/record_spread_missing/overwritten_key.a:8:11: stage 0 can't lower a spread into record storage from a fixed object yet
exit status 1
```

### overwritten_key.ts

Command: `go run ./cmd/adamic c internal/oracle/testdata/record_spread_missing/overwritten_key.ts`; exit 1; wall 0.208s.

```text
adamic: /tmp/bzrdem0-proof/main/internal/oracle/testdata/record_spread_missing/overwritten_key.ts:8:11: stage 0 can't lower a spread into record storage from a fixed object yet
exit status 1
```

### tracing_write_stack_event.ts

Command: `go run ./cmd/adamic c internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts`; exit 1; wall 0.179s.

```text
internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:10: error TS2339: Property 'phase' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:17: error TS2339: Property 'name' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:23: error TS2339: Property 'args' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:29: error TS2339: Property 'time' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:35: error TS2339: Property 'separateBeginAndEnd' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
exit status 1
```

### undefined_args.a

Command: `go run ./cmd/adamic c internal/oracle/testdata/record_spread_missing/undefined_args.a`; exit 1; wall 0.184s.

```text
adamic: /tmp/bzrdem0-proof/main/internal/oracle/testdata/record_spread_missing/undefined_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
exit status 1
```

### undefined_args.ts

Command: `go run ./cmd/adamic c internal/oracle/testdata/record_spread_missing/undefined_args.ts`; exit 1; wall 0.176s.

```text
adamic: /tmp/bzrdem0-proof/main/internal/oracle/testdata/record_spread_missing/undefined_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
exit status 1
```

## Step 3: sole specified mutant

Specification: `internal/lower/record-spread-missing.mutants.md`, “The missing arm dropped”. Asserted the exact old text occurs **once** in `internal/lower/records.go` before replacement:

```go
result.Spread = ir.Coalesce{Value: result.Spread, Fallback: ir.RecordLiteral{Element: of}, Of: ir.Record}
```

Replaced with:

```go
result.Spread = ir.Coalesce{Value: result.Spread, Fallback: ir.RecordLiteral{Element: of}, Of: ir.Record}.Value
```

Ran exactly step 2’s uncached oracle command. exit 1; wall 63.28s.

| Subtest | Result | Whole-got location / meaning |
|---|---|---|
| `null_args.a` | FAIL | Native clang pointer-type error; sanitized same compile failure in isolated backend capture. |
| `null_args.ts` | FAIL | Native clang pointer-type error; sanitized same compile failure in isolated backend capture. |
| `overwritten_key.a` | PASS | All three backends agree with Node. |
| `overwritten_key.ts` | PASS | All three backends agree with Node. |
| `tracing_write_stack_event.ts` | FAIL | Existing five TS2339 diagnostics, before execution. |
| `undefined_args.a` | FAIL | Native stdout/stderr empty, exit -1; sanitized stdout empty, exit 1 with full UBSan stderr below; JavaScript agrees. |
| `undefined_args.ts` | FAIL | Native stdout/stderr empty, exit -1; sanitized stdout empty, exit 1 with full UBSan stderr below; JavaScript agrees. |

For all runnable fixtures, want is the full `PROOF Node stdout=… stderr=… exit=0` observation below. `PROOF <backend>` records the whole corresponding got. Strings use Go `%q` escaping (`\n` represents each newline); none are truncated. For compile failures the complete clang diagnostic is present; runtime stdout/stderr are unavailable because no executable was produced. Tracing’s wanted source-Node streams are the direct run in step 2; every compiler backend is blocked before execution.
The original test stops at a fatal native compile error, so supplemental logging-only reruns captured missing observations. The first logged the existing helper’s streams; the second isolated all 21 program/backend combinations in separate subtests, allowing sanitized and JavaScript checks after native failure. These supplementary tests changed no compiler behavior beyond the specified records.go mutant. The old-text count was again asserted once on the second application. Original sources were restored after each application; temporary test files were removed. Each restoration passed `git diff --exit-code` (exit 0).
Original mutant run, whole got diagnostics:

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
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
    cache_test.go:297: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:10: error TS2339: Property 'phase' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:17: error TS2339: Property 'name' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:23: error TS2339: Property 'args' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:29: error TS2339: Property 'time' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:35: error TS2339: Property 'separateBeginAndEnd' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    record_spread_missing_test.go:31: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__531ab0e390c0a6eb20d29abc43e652d8_2 = adamic_record_iterate(adamic_local_115732354606622155727914530510908772032_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    record_spread_missing_test.go:31: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__9566534b6660ba7451f22b2adb2dc5e2_2 = adamic_record_iterate(adamic_local_74212821344202605624165413379636541401_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:31: native: exit codes differ, stderr ""
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:31: native: exit codes differ, stderr ""
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:31: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:31: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types dc5ee5f7d6af miss 25.49
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types dc5ee5f7d6af hit 25.53
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types dc5ee5f7d6af hit 25.53
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types dc5ee5f7d6af hit 25.53
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (5.62s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (6.08s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (0.54s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (26.52s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (32.61s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (32.61s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (32.63s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=10
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	32.657s
FAIL
```

Supplemental independent-backend run, complete Node and backend observations and errors:

```text
=== RUN   TestProofRecordSpreadStreams
=== RUN   TestProofRecordSpreadStreams/null_args.a/native
    proof_streams_test.go:6: gate cache node miss
    proof_streams_test.go:6: PROOF Node stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
    proof_streams_test.go:6: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__531ab0e390c0a6eb20d29abc43e652d8_2 = adamic_record_iterate(adamic_local_115732354606622155727914530510908772032_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== RUN   TestProofRecordSpreadStreams/null_args.a/sanitized
    proof_streams_test.go:6: gate cache node hit
    proof_streams_test.go:6: PROOF Node stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
    proof_streams_test.go:6: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__531ab0e390c0a6eb20d29abc43e652d8_2 = adamic_record_iterate(adamic_local_115732354606622155727914530510908772032_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== RUN   TestProofRecordSpreadStreams/null_args.a/javascript
    proof_streams_test.go:6: gate cache node hit
    proof_streams_test.go:6: PROOF Node stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
    proof_streams_test.go:6: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types dc5ee5f7d6af hit 1.11
    proof_streams_test.go:6: PROOF javascript stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
=== RUN   TestProofRecordSpreadStreams/null_args.ts/native
    proof_streams_test.go:6: gate cache node miss
    proof_streams_test.go:6: PROOF Node stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
    proof_streams_test.go:6: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__9566534b6660ba7451f22b2adb2dc5e2_2 = adamic_record_iterate(adamic_local_74212821344202605624165413379636541401_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== RUN   TestProofRecordSpreadStreams/null_args.ts/sanitized
    proof_streams_test.go:6: gate cache node hit
    proof_streams_test.go:6: PROOF Node stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
    proof_streams_test.go:6: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__9566534b6660ba7451f22b2adb2dc5e2_2 = adamic_record_iterate(adamic_local_74212821344202605624165413379636541401_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== RUN   TestProofRecordSpreadStreams/null_args.ts/javascript
    proof_streams_test.go:6: gate cache node hit
    proof_streams_test.go:6: PROOF Node stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
    proof_streams_test.go:6: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types dc5ee5f7d6af hit 0.03
    proof_streams_test.go:6: PROOF javascript stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
=== RUN   TestProofRecordSpreadStreams/overwritten_key.a/native
    proof_streams_test.go:6: gate cache node miss
    proof_streams_test.go:6: PROOF Node stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
    proof_streams_test.go:6: PROOF native stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestProofRecordSpreadStreams/overwritten_key.a/sanitized
    proof_streams_test.go:6: gate cache node hit
    proof_streams_test.go:6: PROOF Node stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
    proof_streams_test.go:6: PROOF sanitized stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestProofRecordSpreadStreams/overwritten_key.a/javascript
    proof_streams_test.go:6: gate cache node hit
    proof_streams_test.go:6: PROOF Node stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
    proof_streams_test.go:6: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types dc5ee5f7d6af hit 0.01
    proof_streams_test.go:6: PROOF javascript stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestProofRecordSpreadStreams/overwritten_key.ts/native
    proof_streams_test.go:6: gate cache node miss
    proof_streams_test.go:6: PROOF Node stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
    proof_streams_test.go:6: PROOF native stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestProofRecordSpreadStreams/overwritten_key.ts/sanitized
    proof_streams_test.go:6: gate cache node hit
    proof_streams_test.go:6: PROOF Node stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
    proof_streams_test.go:6: PROOF sanitized stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestProofRecordSpreadStreams/overwritten_key.ts/javascript
    proof_streams_test.go:6: gate cache node hit
    proof_streams_test.go:6: PROOF Node stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
    proof_streams_test.go:6: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types dc5ee5f7d6af hit 0.01
    proof_streams_test.go:6: PROOF javascript stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestProofRecordSpreadStreams/tracing_write_stack_event.ts/native
    cache_test.go:297: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:10: error TS2339: Property 'phase' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:17: error TS2339: Property 'name' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:23: error TS2339: Property 'args' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:29: error TS2339: Property 'time' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:35: error TS2339: Property 'separateBeginAndEnd' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
=== RUN   TestProofRecordSpreadStreams/tracing_write_stack_event.ts/sanitized
    cache_test.go:297: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:10: error TS2339: Property 'phase' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:17: error TS2339: Property 'name' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:23: error TS2339: Property 'args' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:29: error TS2339: Property 'time' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:35: error TS2339: Property 'separateBeginAndEnd' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
=== RUN   TestProofRecordSpreadStreams/tracing_write_stack_event.ts/javascript
    cache_test.go:297: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:10: error TS2339: Property 'phase' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:17: error TS2339: Property 'name' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:23: error TS2339: Property 'args' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:29: error TS2339: Property 'time' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
        /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:35: error TS2339: Property 'separateBeginAndEnd' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } | undefined'.
=== RUN   TestProofRecordSpreadStreams/undefined_args.a/native
    proof_streams_test.go:6: gate cache node miss
    proof_streams_test.go:6: PROOF Node stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
    proof_streams_test.go:6: PROOF native stdout="" stderr="" exit=-1
    proof_streams_test.go:6: native: exit codes differ, stderr ""
=== RUN   TestProofRecordSpreadStreams/undefined_args.a/sanitized
    proof_streams_test.go:6: gate cache node hit
    proof_streams_test.go:6: PROOF Node stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
    proof_streams_test.go:6: PROOF sanitized stdout="" stderr="record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n" exit=1
    proof_streams_test.go:6: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
=== RUN   TestProofRecordSpreadStreams/undefined_args.a/javascript
    proof_streams_test.go:6: gate cache node hit
    proof_streams_test.go:6: PROOF Node stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
    proof_streams_test.go:6: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types dc5ee5f7d6af hit 0.01
    proof_streams_test.go:6: PROOF javascript stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
=== RUN   TestProofRecordSpreadStreams/undefined_args.ts/native
    proof_streams_test.go:6: gate cache node miss
    proof_streams_test.go:6: PROOF Node stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
    proof_streams_test.go:6: PROOF native stdout="" stderr="" exit=-1
    proof_streams_test.go:6: native: exit codes differ, stderr ""
=== RUN   TestProofRecordSpreadStreams/undefined_args.ts/sanitized
    proof_streams_test.go:6: gate cache node hit
    proof_streams_test.go:6: PROOF Node stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
    proof_streams_test.go:6: PROOF sanitized stdout="" stderr="record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n" exit=1
    proof_streams_test.go:6: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
=== RUN   TestProofRecordSpreadStreams/undefined_args.ts/javascript
    proof_streams_test.go:6: gate cache node hit
    proof_streams_test.go:6: PROOF Node stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
    proof_streams_test.go:6: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types dc5ee5f7d6af hit 0.01
    proof_streams_test.go:6: PROOF javascript stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
--- FAIL: TestProofRecordSpreadStreams (24.20s)
    --- FAIL: TestProofRecordSpreadStreams/null_args.a/native (14.47s)
    --- FAIL: TestProofRecordSpreadStreams/null_args.a/sanitized (0.12s)
    --- PASS: TestProofRecordSpreadStreams/null_args.a/javascript (1.75s)
    --- FAIL: TestProofRecordSpreadStreams/null_args.ts/native (1.39s)
    --- FAIL: TestProofRecordSpreadStreams/null_args.ts/sanitized (0.43s)
    --- PASS: TestProofRecordSpreadStreams/null_args.ts/javascript (0.34s)
    --- PASS: TestProofRecordSpreadStreams/overwritten_key.a/native (0.74s)
    --- PASS: TestProofRecordSpreadStreams/overwritten_key.a/sanitized (0.45s)
    --- PASS: TestProofRecordSpreadStreams/overwritten_key.a/javascript (0.10s)
    --- PASS: TestProofRecordSpreadStreams/overwritten_key.ts/native (0.37s)
    --- PASS: TestProofRecordSpreadStreams/overwritten_key.ts/sanitized (0.31s)
    --- PASS: TestProofRecordSpreadStreams/overwritten_key.ts/javascript (0.10s)
    --- FAIL: TestProofRecordSpreadStreams/tracing_write_stack_event.ts/native (0.03s)
    --- FAIL: TestProofRecordSpreadStreams/tracing_write_stack_event.ts/sanitized (0.03s)
    --- FAIL: TestProofRecordSpreadStreams/tracing_write_stack_event.ts/javascript (0.03s)
    --- FAIL: TestProofRecordSpreadStreams/undefined_args.a/native (0.34s)
    --- FAIL: TestProofRecordSpreadStreams/undefined_args.a/sanitized (0.51s)
    --- PASS: TestProofRecordSpreadStreams/undefined_args.a/javascript (0.19s)
    --- FAIL: TestProofRecordSpreadStreams/undefined_args.ts/native (1.37s)
    --- FAIL: TestProofRecordSpreadStreams/undefined_args.ts/sanitized (0.99s)
    --- PASS: TestProofRecordSpreadStreams/undefined_args.ts/javascript (0.14s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=12 misses=12
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	24.222s
FAIL
```

Driver restoration evidence:

```text
Old text occurrence count: 1
{"name": "mutant-oracle", "cmd": ["go", "test", "./internal/oracle", "-run", "TestRecordSpreadFromMissingAgreesWithNode", "-count=1", "-v", "-timeout", "40m"], "cwd": "/workspace/adamic", "exit": 1, "wall_seconds": 63.28225255012512}
{"name": "mutant-streams", "cmd": ["go", "test", "./internal/oracle", "-run", "TestRecordSpreadFromMissingAgreesWithNode", "-count=1", "-v", "-timeout", "40m"], "cwd": "/workspace/adamic", "exit": 1, "wall_seconds": 27.661745071411133}
Restored records.go and diagnostic logging; git diff --exit-code: 0
```

## Step 4: complete lowering suite

```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

All suite commands ran without another heavy command alongside them. Timings include go build/link overhead and are from an external wall-clock wrapper.
Lower: **PASS**, exit 0; wall 1467.69s. No failing tests, environmental or otherwise. Whole log:

```text
ok  	github.com/system-inc/adamic/internal/lower	1459.631s
```

## Step 4b: complete IR suite

```sh
go test ./internal/ir -count=1 -timeout 30m > /tmp/ir.log 2>&1
```

**PASS**, exit code **0**; wall time **140.75 seconds** (2m 20.75s). **No failing tests**, so there are no first error lines to list. The package-reported test time was 106.884s; the wall time includes Go build/link overhead.

This additional step was requested after the original proof had completed and is placed here after step 4 as requested. It ran alone on report commit `a1913c7d4ec221a167cd59f5380cc800186bca6e`, whose sole difference from the pinned compiler HEAD `b7fbcd0abebad8c985aee9ebacfdb99a308641eb` is this report. The compiler and IR sources were unchanged. Loaded `/workspace/adamic-tools/env.sh` first and reused the external proof environment and `/tmp/adamic-gate` work directory.

Whole `/tmp/ir.log`:

```text
ok  	github.com/system-inc/adamic/internal/ir	106.884s
```

## Step 5: complete oracle suite

```sh
go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1
```

Oracle: **FAIL**, exit 1; wall 3776.67s. 16 failing top-level tests; 19 failing names when failed parent/subtest rows are all counted. No package-level 120-minute timeout occurred.
ENV means an established setup issue. TIMING/UNRESOLVED marks a timeout or abnormal exit whose compiler-versus-environment cause was not established; it is not claimed as a compiler regression or a confirmed environment failure. The twelve assignment-proof tests overwrite NODE_PATH with `<repository>/stage3/api/node_modules`; their fixed path cannot find the requested external install even though tsc 6.0.3 is on PATH. WASI_SYSROOT was not provisioned. The fixture’s TS2339 failure is a real compiler/fixture gate failure, not an environment failure.

| Failing test / subtest | Mark | First error line |
|---|---|---|
| `TestAssignmentProofWiderTargetMutant` | ENV: module lookup | assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 |
| `TestAssignmentProofWiderTarget` | ENV: module lookup | assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 |
| `TestAssignmentProofCompoundMutant` | ENV: module lookup | assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 |
| `TestAssignmentProofCompound` | ENV: module lookup | assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 |
| `TestAssignmentProofChained` | ENV: module lookup | assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 |
| `TestAssignmentProofChainedMutant` | ENV: module lookup | assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 |
| `TestAssignmentProofReturnDefined` | ENV: module lookup | assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 |
| `TestAssignmentProofReturnDefinedMutant` | ENV: module lookup | assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 |
| `TestAssignmentProofIfNarrowingMutant` | ENV: module lookup | assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 |
| `TestAssignmentProofIfNarrowing` | ENV: module lookup | assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 |
| `TestAssignmentProofScannerKeywordMutant` | ENV: module lookup | assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 |
| `TestAssignmentProofScannerKeyword` | ENV: module lookup | assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 |
| `TestParallelBenchmark_tsan_One_2` | TIMING/UNRESOLVED: 60s run | parallel_test.go:119: exit codes differ: Node exit 0 stdout "4096 files, 20282626 UTF-16 units, 2621440 tokens, 556992 distinct per file, ordered digest 823772603\n" stderr ""; native exit -1 stdout "" stderr "" |
| `TestParallelBenchmark_tsan_One_2/tsan` | TIMING/UNRESOLVED: 60s run | parallel_test.go:119: exit codes differ: Node exit 0 stdout "4096 files, 20282626 UTF-16 units, 2621440 tokens, 556992 distinct per file, ordered digest 823772603\n" stderr ""; native exit -1 stdout "" stderr "" |
| `TestParallelBenchmark_tsan_One_2/tsan/1` | TIMING/UNRESOLVED: 60s run | parallel_test.go:119: exit codes differ: Node exit 0 stdout "4096 files, 20282626 UTF-16 units, 2621440 tokens, 556992 distinct per file, ordered digest 823772603\n" stderr ""; native exit -1 stdout "" stderr "" |
| `TestStdoutColumnsRefusesWASI` | ENV: missing WASI_SYSROOT | library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT |
| `TestFamiliesCatchAPlantedCase_CountsAreRecorded` | TIMING/UNRESOLVED: 90s probe | families_generated_test.go:5190: the probe of TestCountsAreRecorded was killed at 90 s: |
| `TestRecordSpreadFromMissingAgreesWithNode` | COMPILER/FIXTURE | cache_test.go:297: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:10: error TS2339: Property 'phase' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } \| undefined'. |
| `TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts` | COMPILER/FIXTURE | cache_test.go:297: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:19:10: error TS2339: Property 'phase' does not exist on type '{ phase: string; name: string; args?: Args; time: number; separateBeginAndEnd: boolean; } \| undefined'. |

The parallel benchmark row propagates the same failing `/tsan/1` error through its parents; cache-hit messages preceding it are logs, not errors. The test helper has a 60-second execution deadline. The probe explicitly reports being killed at its own 90-second bound. Those suite-local failures are distinct from admission timeouts below.

## Step 6: full admission delta

Manifest generation:

```sh
python3 cloud/admission-corpus/manifest.py --sha b7fbcd0abebad8c985aee9ebacfdb99a308641eb > /tmp/bzrdem0-proof/manifest.json
```

Manifest counts: witnesses 2, fixtures 4,273, gaps 105, review 20, fuzz 0, stage3 821; total **5,221**. All 5,221 unique paths were classified. Generator blob `09d061a03b1e016387e975d25ee0cd5b75ac6c1f`; manifest blob `028df5f47097c0e9731482881920e50e994fa101`. No unmatched changed .ts programs.

```sh
go run ./cmd/adamic-admission-delta \
  --base 16429a30452d027809f0ea85bc395479423e493e \
  --head b7fbcd0abebad8c985aee9ebacfdb99a308641eb \
  --manifest /tmp/bzrdem0-proof/manifest.json \
  --manifest-generator cloud/admission-corpus/manifest.py \
  --workers 2 --json
```

Command exit 1; wall 1499.61s; internal gate exit **3**, surfaced by go run as shell exit **1**. Verdict **void**; `void=true`; whole void reason: `internal/oracle/testdata/record_spread_missing/null_args.a: native compile timeout`.
`diff_count=7`, `admitted=6`, newly refused **0**, accepted by both **3,270**, refused by both **1,945**. Complete corpus=true, sampling omitted=0, budget=0; 121 runtime candidates selected. Gate records 18 Node agreements, 50 accepted negative witnesses, and 52 ruled runtime stops. Seven diff programs are precisely the seven record-spread fixtures; tracing is refused by both with the same five TS2339 messages.
Whole gate stderr:

```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 3 runs timed out, 3 retried alone, 1 timed out again
changed programs: 7
admission delta is void: internal/oracle/testdata/record_spread_missing/null_args.a: native compile timeout
exit status 3
```

Phase timings (seconds):

```json
{
  "base_build": 62.046275116,
  "checkout": 16.160026845,
  "checkout_and_provision": 17.257555101,
  "classification": 799.87734358,
  "head_build": 27.885014876,
  "manifest_and_verification": 33.606915882,
  "provision": 1.097422354,
  "runtime": 550.43317958,
  "total": 1491.117574884
}
```

### Every classification-stderr difference

Compared exact `base_compile.stderr` and `compile.stderr` over all 5,221 paths: **six differences**, all newly accepted. Both texts below are complete. Empty HEAD text is represented by an empty fenced block. The temporary head-tree prefix in main’s diagnostic is expected: both compilers classify the same head-corpus file.

#### internal/oracle/testdata/record_spread_missing/null_args.a

Main stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-4016340344/head/internal/oracle/testdata/record_spread_missing/null_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
```

HEAD stderr:

```text

```

#### internal/oracle/testdata/record_spread_missing/null_args.ts

Main stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-4016340344/head/internal/oracle/testdata/record_spread_missing/null_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
```

HEAD stderr:

```text

```

#### internal/oracle/testdata/record_spread_missing/overwritten_key.a

Main stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-4016340344/head/internal/oracle/testdata/record_spread_missing/overwritten_key.a:8:11: stage 0 can't lower a spread into record storage from a fixed object yet
```

HEAD stderr:

```text

```

#### internal/oracle/testdata/record_spread_missing/overwritten_key.ts

Main stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-4016340344/head/internal/oracle/testdata/record_spread_missing/overwritten_key.ts:8:11: stage 0 can't lower a spread into record storage from a fixed object yet
```

HEAD stderr:

```text

```

#### internal/oracle/testdata/record_spread_missing/undefined_args.a

Main stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-4016340344/head/internal/oracle/testdata/record_spread_missing/undefined_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
```

HEAD stderr:

```text

```

#### internal/oracle/testdata/record_spread_missing/undefined_args.ts

Main stderr:

```text
adamic: /tmp/adamic-gate/admission-delta-4016340344/head/internal/oracle/testdata/record_spread_missing/undefined_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet
```

HEAD stderr:

```text

```

### Every timeout and isolated rerun at both SHAs

No classification timeout occurred at either SHA. Runtime native-compilation timeouts were on `null_args.a`, `null_args.ts`, and `reflect_get/stop_literal_getter.a`. The gate internally retried all three alone; null_args.ts and the getter recovered, null_args.a did not. The gate records timed_out_runs=3, retried_runs=3, still_timed_out_runs=1.
After the complete gate exited, reran **each affected native build sequentially, alone, at both pinned SHAs**, using the gate’s cached compiler binaries. Compiler cache `.inputs` records map the binaries to the exact SHAs and the gate’s adapter digest `fd17ea5ed37137cf8666c675e74699284f529212114e9958f9d7005569da6bfd`. The source is the same HEAD corpus, now at `/workspace/adamic`, as in the gate contract. Used the original **10-second** runtime-command limit, process-group cleanup, and unchanged build arguments `adamic build <path> -o <external-output>`; then ran successful binaries under the same 10-second bound. No extended-limit retry was needed.

| Program | SHA | Build exit | Build wall | Runtime result |
|---|---|---:|---:|---|
| `internal/oracle/testdata/record_spread_missing/null_args.a` | `16429a30452d027809f0ea85bc395479423e493e` | 1 | 0.120s | not run: main refuses lowering |
| `internal/oracle/testdata/record_spread_missing/null_args.a` | `b7fbcd0abebad8c985aee9ebacfdb99a308641eb` | 0 | 2.719s | exit 0; matches Node |
| `internal/oracle/testdata/record_spread_missing/null_args.ts` | `16429a30452d027809f0ea85bc395479423e493e` | 1 | 0.068s | not run: main refuses lowering |
| `internal/oracle/testdata/record_spread_missing/null_args.ts` | `b7fbcd0abebad8c985aee9ebacfdb99a308641eb` | 0 | 2.556s | exit 0; matches Node |
| `internal/oracle/testdata/reflect_get/stop_literal_getter.a` | `16429a30452d027809f0ea85bc395479423e493e` | 0 | 2.438s | exit 70; exact ruled getter stop |
| `internal/oracle/testdata/reflect_get/stop_literal_getter.a` | `b7fbcd0abebad8c985aee9ebacfdb99a308641eb` | 0 | 2.184s | exit 70; exact ruled getter stop |

Full commands, compiler identities, got stdout/stderr, exit status, and wall time for all six reruns:

```json
[
  {
    "path": "internal/oracle/testdata/record_spread_missing/null_args.a",
    "revision": "base",
    "sha": "16429a30452d027809f0ea85bc395479423e493e",
    "compiler": "/home/agent/.cache/adamic-build/local/3cd7a846a4ad548be4278b46475010063c2765075ea2f5611ce00dd66935105f/adamic",
    "native_compile": {
      "cmd": [
        "/home/agent/.cache/adamic-build/local/3cd7a846a4ad548be4278b46475010063c2765075ea2f5611ce00dd66935105f/adamic",
        "build",
        "internal/oracle/testdata/record_spread_missing/null_args.a",
        "-o",
        "/tmp/bzrdem0-proof/retry-0-base"
      ],
      "limit_seconds": 10,
      "exit": 1,
      "timeout": false,
      "wall_seconds": 0.11977720260620117,
      "stdout": "",
      "stderr": "adamic: /workspace/adamic/internal/oracle/testdata/record_spread_missing/null_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet\n"
    }
  },
  {
    "path": "internal/oracle/testdata/record_spread_missing/null_args.a",
    "revision": "head",
    "sha": "b7fbcd0abebad8c985aee9ebacfdb99a308641eb",
    "compiler": "/home/agent/.cache/adamic-build/local/53dac23a6ae4e74ec66576d830dd7e39da0b0c68675a14ec5759399b8e585a81/adamic",
    "native_compile": {
      "cmd": [
        "/home/agent/.cache/adamic-build/local/53dac23a6ae4e74ec66576d830dd7e39da0b0c68675a14ec5759399b8e585a81/adamic",
        "build",
        "internal/oracle/testdata/record_spread_missing/null_args.a",
        "-o",
        "/tmp/bzrdem0-proof/retry-0-head"
      ],
      "limit_seconds": 10,
      "exit": 0,
      "timeout": false,
      "wall_seconds": 2.719425678253174,
      "stdout": "",
      "stderr": ""
    },
    "native": {
      "cmd": [
        "/tmp/bzrdem0-proof/retry-0-head"
      ],
      "limit_seconds": 10,
      "exit": 0,
      "timeout": false,
      "wall_seconds": 0.0036385059356689453,
      "stdout": "results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n",
      "stderr": ""
    }
  },
  {
    "path": "internal/oracle/testdata/record_spread_missing/null_args.ts",
    "revision": "base",
    "sha": "16429a30452d027809f0ea85bc395479423e493e",
    "compiler": "/home/agent/.cache/adamic-build/local/3cd7a846a4ad548be4278b46475010063c2765075ea2f5611ce00dd66935105f/adamic",
    "native_compile": {
      "cmd": [
        "/home/agent/.cache/adamic-build/local/3cd7a846a4ad548be4278b46475010063c2765075ea2f5611ce00dd66935105f/adamic",
        "build",
        "internal/oracle/testdata/record_spread_missing/null_args.ts",
        "-o",
        "/tmp/bzrdem0-proof/retry-1-base"
      ],
      "limit_seconds": 10,
      "exit": 1,
      "timeout": false,
      "wall_seconds": 0.06806015968322754,
      "stdout": "",
      "stderr": "adamic: /workspace/adamic/internal/oracle/testdata/record_spread_missing/null_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet\n"
    }
  },
  {
    "path": "internal/oracle/testdata/record_spread_missing/null_args.ts",
    "revision": "head",
    "sha": "b7fbcd0abebad8c985aee9ebacfdb99a308641eb",
    "compiler": "/home/agent/.cache/adamic-build/local/53dac23a6ae4e74ec66576d830dd7e39da0b0c68675a14ec5759399b8e585a81/adamic",
    "native_compile": {
      "cmd": [
        "/home/agent/.cache/adamic-build/local/53dac23a6ae4e74ec66576d830dd7e39da0b0c68675a14ec5759399b8e585a81/adamic",
        "build",
        "internal/oracle/testdata/record_spread_missing/null_args.ts",
        "-o",
        "/tmp/bzrdem0-proof/retry-1-head"
      ],
      "limit_seconds": 10,
      "exit": 0,
      "timeout": false,
      "wall_seconds": 2.5559778213500977,
      "stdout": "",
      "stderr": ""
    },
    "native": {
      "cmd": [
        "/tmp/bzrdem0-proof/retry-1-head"
      ],
      "limit_seconds": 10,
      "exit": 0,
      "timeout": false,
      "wall_seconds": 0.003157377243041992,
      "stdout": "results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n",
      "stderr": ""
    }
  },
  {
    "path": "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
    "revision": "base",
    "sha": "16429a30452d027809f0ea85bc395479423e493e",
    "compiler": "/home/agent/.cache/adamic-build/local/3cd7a846a4ad548be4278b46475010063c2765075ea2f5611ce00dd66935105f/adamic",
    "native_compile": {
      "cmd": [
        "/home/agent/.cache/adamic-build/local/3cd7a846a4ad548be4278b46475010063c2765075ea2f5611ce00dd66935105f/adamic",
        "build",
        "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
        "-o",
        "/tmp/bzrdem0-proof/retry-2-base"
      ],
      "limit_seconds": 10,
      "exit": 0,
      "timeout": false,
      "wall_seconds": 2.4378862380981445,
      "stdout": "",
      "stderr": ""
    },
    "native": {
      "cmd": [
        "/tmp/bzrdem0-proof/retry-2-base"
      ],
      "limit_seconds": 10,
      "exit": 70,
      "timeout": false,
      "wall_seconds": 0.0023403167724609375,
      "stdout": "got\n",
      "stderr": "adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')\n"
    }
  },
  {
    "path": "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
    "revision": "head",
    "sha": "b7fbcd0abebad8c985aee9ebacfdb99a308641eb",
    "compiler": "/home/agent/.cache/adamic-build/local/53dac23a6ae4e74ec66576d830dd7e39da0b0c68675a14ec5759399b8e585a81/adamic",
    "native_compile": {
      "cmd": [
        "/home/agent/.cache/adamic-build/local/53dac23a6ae4e74ec66576d830dd7e39da0b0c68675a14ec5759399b8e585a81/adamic",
        "build",
        "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
        "-o",
        "/tmp/bzrdem0-proof/retry-2-head"
      ],
      "limit_seconds": 10,
      "exit": 0,
      "timeout": false,
      "wall_seconds": 2.1837568283081055,
      "stdout": "",
      "stderr": ""
    },
    "native": {
      "cmd": [
        "/tmp/bzrdem0-proof/retry-2-head"
      ],
      "limit_seconds": 10,
      "exit": 70,
      "timeout": false,
      "wall_seconds": 0.001863241195678711,
      "stdout": "got\n",
      "stderr": "adamic: panic: an object literal's getter isn't read by a run-time key yet (reading 'value')\n"
    }
  }
]
```

The two HEAD null_args outputs equal the original gate’s Node stdout byte-for-byte, exit 0, stderr empty. Main’s two null_args refusals are expected and are fully retained above. The getter runs return exit 70, stdout `got\n`, and exactly the gate’s ruled stderr at both SHAs. All observed admission timeouts therefore cleared on these subsequent isolated runs. **The original full admission verdict is still void; no second full-corpus gate was run.**

## Evidence integrity and delivery scope

Raw evidence remained outside the repository. SHA-256 digests for the principal evidence files (the relevant full diagnostics and stream observations are embedded above):

| Evidence | SHA-256 |
|---|---|
| `/tmp/bzrdem0-proof/pin-oracle.log` | `fd55e7e72446e51e7dce05a6d4534772a220f7b3905b70a681b66b40f469a7f0` |
| `/tmp/bzrdem0-proof/pin-lower.log` | `3d6d74e8043f52c06f34bf398dc6ed671732340a8c2f074ed01ae3fc51396dc2` |
| `/tmp/bzrdem0-proof/main-pin.log` | `38035e670ea275c88acbe500ec8e9b8559374c08dbd94eecac60ed7e23f5bf6d` |
| `/tmp/bzrdem0-proof/main-pin-errors.log` | `e4479c09643adad3223423bb9aa59f26faddf40f05bc601aec0722e1e7814b21` |
| `/tmp/bzrdem0-proof/main-programs.json` | `230bfac973d0bec9ef10ead655ab6c4bf89f398afde258db58355a1caf72e2de` |
| `/tmp/bzrdem0-proof/mutant-oracle.log` | `f4cb312c3147ed3b828a356a478b7cab830893c6c29095c49e9a95c4ad749f3d` |
| `/tmp/bzrdem0-proof/mutant-streams.log` | `200e8669e6b990b431a73fdd29973b26914596d8e7c744885540c6b9c4ee4943` |
| `/tmp/bzrdem0-proof/mutant-all-backends.log` | `095600b860d59b7cd865af942783dbe18882da77754ce68dc22ef2ee5fe2e0b5` |
| `/tmp/lower.log` | `f49c023aa6d239f61bbb76c09fd80e3ecdcd67743a70ec80ed1a9ef66db88de3` |
| `/tmp/ir.log` | `e932c3b4d6e450a2bcefc1ff160101621360d11cdb72cc1a54301b874af515b1` |
| `/tmp/oracle.log` | `799be632ea68445875c533d72368c1f96ba431f647f4131c5dcd34f06eff31f8` |
| `/tmp/bzrdem0-proof/admission.json` | `679b18ee6cfa739d6c6b72910b5c4fa71027cf3ee07559080abb9006c9e5d0fd` |
| `/tmp/bzrdem0-proof/admission.stderr` | `a6dea952c8386c73f43a6c78e8795c00b784c1ef58512c8f3626f1ade1a0f973` |
| `/tmp/bzrdem0-proof/timeout-reruns.json` | `db873fe9b4de2c54695a76e75ea8b3e1e0c80fcb3d43796f247c845d57a08c56` |

Before report creation, `git status --short` printed nothing and `git diff --exit-code` exited 0. All mutant and logging changes were restored; no compiler code is part of this delivery. The only intended committed/pushed path is `proof/lower-bzr-2/report.md`, on new branch `proof/lower-bzr-2-report` created from the pinned detached HEAD; origin’s compiler and main branches are untouched.
