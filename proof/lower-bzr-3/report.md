# Compiler proof: #bzrdem0

Date: 2026-11-11 (UTC). Requested HEAD: `a38bdd1290858572341e397d51331c662f52ef0c`; main: `efe887763f26e96653546fae904d57539139f7ae`. HEAD is one commit above main. Report branch: `proof/lower-bzr-3-report`, based on the requested HEAD.

## Decisive results

- **Admission delta PASS**: exit 0, wall 1,663 s (27m43s), full corpus, two workers, `void=false`, no void reasons, `diff_count=7`, admitted 6, newly refused 0. All six newly admitted programs agree with Node on native and JavaScript. Exactly six classification stderr changes, reproduced in full below. Two initial repair-native-compile timeouts passed automatic retries and all four additional isolated checks at the two SHAs.
- **Targeted oracle pin FAIL**: six subtests pass; `tracing_write_stack_event.ts` fails lowering at 14:37 on JSON.stringify object references. **Main fails the same fixture with the identical diagnostic**. This fixture never reaches backend execution. Its entire Node output and diagnostic are included below.
- **Documented mutant detected**: dropping the missing-arm coalesce makes both undefined fixtures and both null fixtures red. Both overwritten-key fixtures stay green. Tracing was already red and is not credited as a mutant kill. Restored source passed `git diff --exit-code`.
- **Formatting, vet, fixed-object pin, full lower, and full IR suites PASS**.
- **Full oracle FAIL**: three failing tests: the same tracing refusal; an environmental missing `WASI_SYSROOT`; and a 90-second nested count-probe timeout. The probe passes alone at HEAD and main.

## Setup and execution discipline

Loaded `/workspace/adamic-tools/env.sh` before repository commands. Created `/tmp/adamic-gate` outside the repository for its configured `TMPDIR`. All heavy commands ran sequentially, including the admission delta and its isolated reruns. No compiler changes were committed or pushed.

```sh
git fetch origin compiler/record-spread-undefined-4
git checkout --detach a38bdd1290858572341e397d51331c662f52ef0c
git submodule update --init --recursive
```

These commands exited 0. Recursive submodules: cohere `5360753e19e4417675a90e92a1823bea66b13918`; cohere/TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`.

Provisioned a separate clean upstream checkout at `/tmp/adamic-typescript`, exactly `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported `ADAMIC_TYPESCRIPT_SOURCE` to it. Installed stage3/api package.json and package-lock.json pins using `npm ci --prefix /tmp/adamic-api`: TypeScript 6.0.3, @types/node 25.3.3 and its pinned transitive package. Exposed that external installation through file links inside the actual ignored `stage3/api/node_modules` directory; assignment_proofs_test.go explicitly overrides NODE_PATH to that directory. Verified both `tsc --version` and `require("typescript").version` report 6.0.3. The worktree stayed clean.

```sh
source /workspace/adamic-tools/env.sh
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-typescript
export PATH=/tmp/adamic-api/node_modules/.bin:$PATH
export NODE_PATH=/workspace/adamic/stage3/api/node_modules
```

These CA and strict-SSL exports were present for npm and inherited by the admission gate and its children. Versions: Go 1.27.1 linux/amd64, Node v24.19.0, tsc 6.0.3. No TLS verification was disabled.

## Step 1: formatting and vet

```sh
gofmt -l internal/lower internal/oracle
go vet ./internal/lower ./internal/oracle
```

gofmt exited 0 and printed **nothing**. Vet: exit=0 wall_seconds=235; output empty.

## Step 2: targeted pins

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestRecordSpreadFromMissingAgreesWithNode' -count=1 -v -timeout 40m
go test ./internal/lower -run 'TestRecordSpreadFromMissingFixedObjectStops' -count=1 -v
```

Oracle: exit=1 wall_seconds=305. Lower: exit=0 wall_seconds=34.

| Oracle subtest | Result |
|---|---|
| null_args.a | PASS |
| null_args.ts | PASS |
| overwritten_key.a | PASS |
| overwritten_key.ts | PASS |
| tracing_write_stack_event.ts | FAIL |
| undefined_args.a | PASS |
| undefined_args.ts | PASS |

Fixed-object lower subtests `.a` and `.ts`: both PASS.

Tracing got (entire lowering failure):

```text
/workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:14:37: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet
```

Want: lowering succeeds, then each backend agrees with the complete Node run. Node exits 0 with empty stderr; its complete stdout is logged in the supplemental evidence below. For tracing, native, sanitized, and JavaScript backend execution were **not reached**: there is no backend process stdout, stderr, or exit status (N/A, not an observed exit 0). Reporting empty successful runs here would be incorrect.

To obtain full got/want bytes without modifying repository tests, ran a supplemental external Go overlay (`GOFLAGS=-overlay=/tmp/adamic-proof/overlay.json`) that adds `%q` logs of Node and each backend stdout, stderr and exit, and gives each backend its own subtest so a native compile failure cannot hide the other observations. `%q` logs are complete escaped strings, not truncations. Ordinary required commands above ran without this overlay. Supplemental baseline: exit=1 wall_seconds=33.

### Original oracle pin

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
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
    record_spread_missing_test.go:31: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:14:37: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 334c4a0880b3 hit 215.85
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 334c4a0880b3 hit 215.85
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 334c4a0880b3 hit 215.85
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 334c4a0880b3 miss 215.85
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 334c4a0880b3 hit 215.86
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 334c4a0880b3 hit 0.02
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (41.19s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (258.75s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (258.79s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (258.79s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (258.80s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (217.68s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (1.04s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=13
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	259.812s
FAIL
```

### Original fixed-object lower pin

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
    --- PASS: TestRecordSpreadFromMissingFixedObjectStops/.ts (0.14s)
    --- PASS: TestRecordSpreadFromMissingFixedObjectStops/.a (0.15s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.269s
```

### Full baseline Node/backend observations

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
    record_spread_missing_test.go:31: PROOF Node stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:31: PROOF Node stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
    record_spread_missing_test.go:31: PROOF Node stdout="{\"pid\":1,\"tid\":1,\"ph\":\"X\",\"cat\":\"check\",\"ts\":0,\"name\":\"checkSourceFile\",\"dur\":20000,\"args\":{\"results\":{\"count\":2,\"nested\":{\"deep\":true},\"list\":[\"a\",1,null]}} keys results count 1}\n{\"pid\":1,\"tid\":1,\"ph\":\"X\",\"cat\":\"emit\",\"ts\":5,\"name\":\"emitJsFileOrBundle\",\"dur\":19995,\"args\":{\"path\":\"x.ts\"} keys path,results count 2}\n{\"pid\":1,\"tid\":1,\"ph\":\"X\",\"cat\":\"bind\",\"ts\":7,\"name\":\"bindSourceFile\",\"dur\":19993,\"args\":{\"path\":\"y.ts\",\"results\":{\"kind\":1}} keys path,results count 2}\n{\"pid\":1,\"tid\":1,\"ph\":\"X\",\"cat\":\"parse\",\"ts\":9,\"name\":\"createSourceFile\",\"dur\":19991,\"args\":{} keys results count 1}\n{\"pid\":1,\"tid\":1,\"ph\":\"E\",\"cat\":\"program\",\"ts\":20000,\"name\":\"createProgram\",\"args\":{\"configFilePath\":\"tsconfig.json\"} keys configFilePath count 1}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    record_spread_missing_test.go:31: PROOF Node stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    record_spread_missing_test.go:31: PROOF Node stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/native
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/native
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/native
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
    record_spread_missing_test.go:31: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:14:37: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.a/native
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/native
    loom_views_d_test.go:76: PROOF native stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/sanitized
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a/native
    loom_views_d_test.go:76: PROOF native stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.a/sanitized
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/native
    loom_views_d_test.go:76: PROOF native stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/sanitized
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/native
    loom_views_d_test.go:76: PROOF native stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/sanitized
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    record_spread_missing_test.go:31: PROOF Node stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/native
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/sanitized
    loom_views_d_test.go:76: PROOF sanitized stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/javascript
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a/sanitized
    loom_views_d_test.go:76: PROOF sanitized stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.a/javascript
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/sanitized
    loom_views_d_test.go:76: PROOF sanitized stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/javascript
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/native
    loom_views_d_test.go:76: PROOF native stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/sanitized
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/sanitized
    loom_views_d_test.go:76: PROOF sanitized stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/javascript
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/sanitized
    loom_views_d_test.go:76: PROOF sanitized stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/javascript
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types c759e4ddb97f hit 6.76
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types c759e4ddb97f hit 6.35
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types c759e4ddb97f hit 6.76
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types c759e4ddb97f hit 6.76
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types c759e4ddb97f miss 6.76
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/javascript
    loom_views_d_test.go:76: PROOF javascript stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/javascript
    loom_views_d_test.go:76: PROOF javascript stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/javascript
    loom_views_d_test.go:76: PROOF javascript stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a/javascript
    loom_views_d_test.go:76: PROOF javascript stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/javascript
    loom_views_d_test.go:76: PROOF javascript stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    record_spread_missing_test.go:31: PROOF Node stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/native
    loom_views_d_test.go:76: PROOF native stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/sanitized
    loom_views_d_test.go:76: PROOF sanitized stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types c759e4ddb97f hit 0.02
    loom_views_d_test.go:76: PROOF javascript stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (6.53s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (8.07s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/native (0.36s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/sanitized (0.55s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/javascript (6.43s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (14.62s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/native (0.62s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/sanitized (0.55s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/javascript (6.98s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (14.62s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/native (0.30s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/sanitized (0.71s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/javascript (7.11s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (14.65s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.a/native (0.22s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.a/sanitized (0.68s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.a/javascript (7.19s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (14.65s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/native (0.29s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/sanitized (0.63s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/javascript (7.27s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (0.93s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/native (0.16s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/sanitized (0.40s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/javascript (0.10s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=13
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	15.558s
FAIL
```

## Step 3: every documented mutant

The document contains exactly one mutant, **The missing arm dropped**. Asserted the old text occurs exactly once in `internal/lower/records.go`, saved the original bytes outside the repository, and replaced it once.

```text
OLD: result.Spread = ir.Coalesce{Value: result.Spread, Fallback: ir.RecordLiteral{Element: of}, Of: ir.Record}
NEW: result.Spread = ir.Coalesce{Value: result.Spread, Fallback: ir.RecordLiteral{Element: of}, Of: ir.Record}.Value
```

Reran the exact step-2 oracle command: exit=1 wall_seconds=108. Then reran with only the external observation overlay: exit=1 wall_seconds=34.

| Fixture(s) | Native got | Sanitized got | JavaScript got |
|---|---|---|---|
| undefined_args.a, undefined_args.ts | stdout/stderr empty; run helper exit -1 (signaled process) | stdout empty; exit 1 with full UBSan diagnostic below | PASS, stdout equals Node, stderr empty, exit 0 |
| null_args.a, null_args.ts | C compile fails, exit status 1; no execution | C compile fails, exit status 1; no execution | PASS, stdout equals Node, stderr empty, exit 0 |
| tracing_write_stack_event.ts | Not reached: same baseline lowering refusal | Not reached | Not reached |
| overwritten_key.a, overwritten_key.ts | PASS | PASS | PASS |

Every red fixture, its whole assertion, Node want, native/sanitized/JavaScript got, and complete compiler/sanitizer stderr appear in the logs below. In the required uninstrumented run, null native compilation aborts its parent subtest before other backends; the supplemental run measures sanitized and JavaScript independently. Restored original `records.go` bytes and ran `git diff --exit-code`: exit 0, empty output.

### Original mutant oracle run

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
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
    record_spread_missing_test.go:31: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:14:37: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:31: native: exit codes differ, stderr ""
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    record_spread_missing_test.go:31: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__531ab0e390c0a6eb20d29abc43e652d8_2 = adamic_record_iterate(adamic_local_115732354606622155727914530510908772032_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:31: native: exit codes differ, stderr ""
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:31: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
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
    record_spread_missing_test.go:31: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types fb2a908a27a2 hit 38.26
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types fb2a908a27a2 hit 38.26
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types fb2a908a27a2 hit 37.89
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types fb2a908a27a2 miss 38.26
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (22.48s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (22.94s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (0.55s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (39.61s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (62.11s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (62.12s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (62.13s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=11
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	62.163s
FAIL
```

### Full mutant Node/backend observations

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
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a
    record_spread_missing_test.go:31: PROOF Node stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts
    record_spread_missing_test.go:31: PROOF Node stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts
    record_spread_missing_test.go:31: PROOF Node stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a
    record_spread_missing_test.go:31: PROOF Node stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
    record_spread_missing_test.go:31: PROOF Node stdout="{\"pid\":1,\"tid\":1,\"ph\":\"X\",\"cat\":\"check\",\"ts\":0,\"name\":\"checkSourceFile\",\"dur\":20000,\"args\":{\"results\":{\"count\":2,\"nested\":{\"deep\":true},\"list\":[\"a\",1,null]}} keys results count 1}\n{\"pid\":1,\"tid\":1,\"ph\":\"X\",\"cat\":\"emit\",\"ts\":5,\"name\":\"emitJsFileOrBundle\",\"dur\":19995,\"args\":{\"path\":\"x.ts\"} keys path,results count 2}\n{\"pid\":1,\"tid\":1,\"ph\":\"X\",\"cat\":\"bind\",\"ts\":7,\"name\":\"bindSourceFile\",\"dur\":19993,\"args\":{\"path\":\"y.ts\",\"results\":{\"kind\":1}} keys path,results count 2}\n{\"pid\":1,\"tid\":1,\"ph\":\"X\",\"cat\":\"parse\",\"ts\":9,\"name\":\"createSourceFile\",\"dur\":19991,\"args\":{} keys results count 1}\n{\"pid\":1,\"tid\":1,\"ph\":\"E\",\"cat\":\"program\",\"ts\":20000,\"name\":\"createProgram\",\"args\":{\"configFilePath\":\"tsconfig.json\"} keys configFilePath count 1}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.a/native
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/native
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/native
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/native
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a/native
    loom_views_d_test.go:70: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__531ab0e390c0a6eb20d29abc43e652d8_2 = adamic_record_iterate(adamic_local_115732354606622155727914530510908772032_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.a/sanitized
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
    record_spread_missing_test.go:31: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:14:37: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/native
    loom_views_d_test.go:70: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__9566534b6660ba7451f22b2adb2dc5e2_2 = adamic_record_iterate(adamic_local_74212821344202605624165413379636541401_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/sanitized
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a/sanitized
    loom_views_d_test.go:72: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__531ab0e390c0a6eb20d29abc43e652d8_2 = adamic_record_iterate(adamic_local_115732354606622155727914530510908772032_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.a/javascript
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/sanitized
    loom_views_d_test.go:72: native: clang failed: exit status 1
        main.c:32:108: error: incompatible pointer types passing 'adamic_heap *' (aka 'struct adamic_heap *') to parameter of type 'adamic_record *' (aka 'struct adamic_object *') [-Werror,-Wincompatible-pointer-types]
           32 |         adamic_record_iterator *adamic_temporary_body__9566534b6660ba7451f22b2adb2dc5e2_2 = adamic_record_iterate(adamic_local_74212821344202605624165413379636541401_args);
              |                                                                                                                   ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
        runtime/adamic.h:1149:62: note: passing argument to parameter 'record' here
         1149 | adamic_record_iterator *adamic_record_iterate(adamic_record *record);
              |                                                              ^
        1 error generated.
        
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/javascript
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/native
    loom_views_d_test.go:76: PROOF native stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/sanitized
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/native
    loom_views_d_test.go:76: PROOF native stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/sanitized
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts
    record_spread_missing_test.go:31: PROOF Node stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/native
    loom_views_d_test.go:76: PROOF native stdout="" stderr="" exit=-1
    loom_views_d_test.go:78: native: exit codes differ, stderr ""
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/sanitized
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/sanitized
    loom_views_d_test.go:76: PROOF sanitized stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/javascript
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/sanitized
    loom_views_d_test.go:76: PROOF sanitized stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/javascript
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/sanitized
    loom_views_d_test.go:76: PROOF sanitized stdout="" stderr="record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n" exit=1
    loom_views_d_test.go:78: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/javascript
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 1e6aafc733c7 miss 7.73
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 1e6aafc733c7 hit 7.03
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 1e6aafc733c7 hit 7.44
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 1e6aafc733c7 hit 7.73
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 1e6aafc733c7 hit 7.58
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/javascript
    loom_views_d_test.go:76: PROOF javascript stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/javascript
    loom_views_d_test.go:76: PROOF javascript stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/null_args.a/javascript
    loom_views_d_test.go:76: PROOF javascript stdout="results\n1\n{\"results\":5}\nfirst,results,last\n3\n{\"first\":1,\"results\":6,\"last\":2}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/javascript
    loom_views_d_test.go:76: PROOF javascript stdout="first,second,results\n3\n3\nresults,last\n2\n{\"results\":4,\"last\":2}\n{\"first\":1,\"second\":2,\"results\":9}\n{\"results\":9,\"last\":2}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/javascript
    loom_views_d_test.go:76: PROOF javascript stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
=== NAME  TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a
    record_spread_missing_test.go:31: PROOF Node stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/native
    loom_views_d_test.go:76: PROOF native stdout="" stderr="" exit=-1
    loom_views_d_test.go:78: native: exit codes differ, stderr ""
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/sanitized
    loom_views_d_test.go:76: PROOF sanitized stdout="" stderr="record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n" exit=1
    loom_views_d_test.go:78: sanitized: exit codes differ, stderr "record.c:34:17: runtime error: member access within null pointer of type 'const adamic_record' (aka 'const struct adamic_object')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior record.c:34:17 \n"
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 1e6aafc733c7 hit 0.05
    loom_views_d_test.go:76: PROOF javascript stdout="results\n1\n3\nresults\n1\n{}\n" stderr="" exit=0
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (6.38s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts (14.76s)
        --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/native (0.13s)
        --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/sanitized (0.11s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.ts/javascript (8.23s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a (14.79s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/native (0.29s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/sanitized (0.59s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.a/javascript (7.62s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.a (14.80s)
        --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.a/native (0.15s)
        --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/null_args.a/sanitized (0.11s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/null_args.a/javascript (8.33s)
    --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts (14.83s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/native (0.28s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/sanitized (0.46s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/overwritten_key.ts/javascript (7.79s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts (8.46s)
        --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/native (0.24s)
        --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/sanitized (0.64s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.ts/javascript (7.23s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a (2.43s)
        --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/native (0.26s)
        --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/sanitized (1.63s)
        --- PASS: TestRecordSpreadFromMissingAgreesWithNode/undefined_args.a/javascript (0.15s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=13
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	17.225s
FAIL
```

## Step 4: full lower

```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

exit=0 wall_seconds=1576 (26m16s). No failing tests; no environmental failures.

```text
ok  	github.com/system-inc/adamic/internal/lower	1571.555s
```

## Step 4b: IR guards

```sh
go test ./internal/ir -count=1 -timeout 30m
```

exit=0 wall_seconds=149. Full IR suite passed, including `TestCallTargetReadersLower` and the IR field guards. Every failure: none.

```text
ok  	github.com/system-inc/adamic/internal/ir	116.398s
```

## Step 5: full oracle and main comparisons

```sh
go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1
```

exit=1 wall_seconds=4071 (67m51s). Every failing test and its first error line:

| Failing test | First error line | Classification / isolated result |
|---|---|---|
| TestStdoutColumnsRefusesWASI | `library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT` | **Environmental**: missing WASI_SYSROOT. No main rerun required for this plainly environmental failure. |
| TestFamiliesCatchAPlantedCase_CountsAreRecorded | `families_generated_test.go:5190: the probe of TestCountsAreRecorded was killed at 90 s:` | Full-suite internal timeout; not conclusively assigned to environment. PASS alone on HEAD and main. |
| TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts | `record_spread_missing_test.go:31: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:14:37: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet` | FAIL identically on main. Parent TestRecordSpreadFromMissingAgreesWithNode fails because of this child. The preceding gate-cache miss is informational. |

Isolated commands and outcomes:

```sh
# HEAD
go test ./internal/oracle -run '^TestFamiliesCatchAPlantedCase_CountsAreRecorded$' -count=1 -v -timeout 10m
# Main efe887763f26e96653546fae904d57539139f7ae
go test ./internal/oracle -run '^TestFamiliesCatchAPlantedCase_CountsAreRecorded$' -count=1 -v -timeout 10m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestRecordSpreadFromMissingAgreesWithNode$/^tracing_write_stack_event.ts$' -count=1 -v -timeout 40m
```

HEAD count probe: exit=0 wall_seconds=26. Main count probe: exit=0 wall_seconds=50. Main tracing: exit=1 wall_seconds=10. Main was checked out detached in the same repository, preserving its exact compiler source. The new oracle test and seven fixtures do not exist on main, so they were copied from HEAD as untracked test evidence only. No compiler source was carried. They were removed before restoring HEAD. The tracing diagnostic is identical on both revisions. The count-probe timeout reproduces on neither in isolation.

### HEAD isolated count probe

```text
=== RUN   TestFamiliesCatchAPlantedCase_CountsAreRecorded
=== PAUSE TestFamiliesCatchAPlantedCase_CountsAreRecorded
=== CONT  TestFamiliesCatchAPlantedCase_CountsAreRecorded
    families_generated_test.go:5190: the mutant planted in dedication/dedication.a failed exactly TestCountsAreRecorded_003, with "the counts changed"
--- PASS: TestFamiliesCatchAPlantedCase_CountsAreRecorded (17.94s)
PASS
ok  	github.com/system-inc/adamic/internal/oracle	17.959s
```

### Main isolated count probe

```text
=== RUN   TestFamiliesCatchAPlantedCase_CountsAreRecorded
=== PAUSE TestFamiliesCatchAPlantedCase_CountsAreRecorded
=== CONT  TestFamiliesCatchAPlantedCase_CountsAreRecorded
    families_generated_test.go:5190: the mutant planted in dedication/dedication.a failed exactly TestCountsAreRecorded_003, with "the counts changed"
--- PASS: TestFamiliesCatchAPlantedCase_CountsAreRecorded (16.96s)
PASS
ok  	github.com/system-inc/adamic/internal/oracle	16.977s
```

### Main isolated tracing fixture

```text
=== RUN   TestRecordSpreadFromMissingAgreesWithNode
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode
=== CONT  TestRecordSpreadFromMissingAgreesWithNode
=== RUN   TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
=== PAUSE TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
=== CONT  TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts
    record_spread_missing_test.go:31: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:14:37: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (5.89s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=1
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	5.905s
FAIL
```

### Full oracle log, unabridged

```text
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
--- FAIL: TestStdoutColumnsRefusesWASI (0.67s)
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestFamiliesCatchAPlantedCase_CountsAreRecorded (90.02s)
    families_generated_test.go:5190: the probe of TestCountsAreRecorded was killed at 90 s:
        === RUN   TestCountsAreRecorded_000
        === PAUSE TestCountsAreRecorded_000
        === RUN   TestCountsAreRecorded_001
        === PAUSE TestCountsAreRecorded_001
        === RUN   TestCountsAreRecorded_002
        === PAUSE TestCountsAreRecorded_002
        === RUN   TestCountsAreRecorded_003
        === PAUSE TestCountsAreRecorded_003
        === RUN   TestCountsAreRecorded_004
        === PAUSE TestCountsAreRecorded_004
        === RUN   TestCountsAreRecorded_005
        === PAUSE TestCountsAreRecorded_005
        === RUN   TestCountsAreRecorded_006
        === PAUSE TestCountsAreRecorded_006
        === RUN   TestCountsAreRecorded_007
        === PAUSE TestCountsAreRecorded_007
        === RUN   TestCountsAreRecorded_008
        === PAUSE TestCountsAreRecorded_008
        === RUN   TestCountsAreRecorded_009
        === PAUSE TestCountsAreRecorded_009
        === RUN   TestCountsAreRecorded_010
        === PAUSE TestCountsAreRecorded_010
        === RUN   TestCountsAreRecorded_011
        === PAUSE TestCountsAreRecorded_011
        === RUN   TestCountsAreRecorded_012
        === PAUSE TestCountsAreRecorded_012
        === RUN   TestCountsAreRecorded_013
        === PAUSE TestCountsAreRecorded_013
        === RUN   TestCountsAreRecorded_014
        === PAUSE TestCountsAreRecorded_014
        === RUN   TestCountsAreRecorded_015
        === PAUSE TestCountsAreRecorded_015
        === RUN   TestCountsAreRecorded_016
        === PAUSE TestCountsAreRecorded_016
        === RUN   TestCountsAreRecorded_017
        === PAUSE TestCountsAreRecorded_017
        === RUN   TestCountsAreRecorded_018
        === PAUSE TestCountsAreRecorded_018
        === RUN   TestCountsAreRecorded_019
        === PAUSE TestCountsAreRecorded_019
        === RUN   TestCountsAreRecorded_020
        === PAUSE TestCountsAreRecorded_020
        === RUN   TestCountsAreRecorded_021
        === PAUSE TestCountsAreRecorded_021
        === RUN   TestCountsAreRecorded_022
        === PAUSE TestCountsAreRecorded_022
        === RUN   TestCountsAreRecorded_023
        === PAUSE TestCountsAreRecorded_023
        === RUN   TestCountsAreRecorded_024
        === PAUSE TestCountsAreRecorded_024
        === RUN   TestCountsAreRecorded_025
        === PAUSE TestCountsAreRecorded_025
        === RUN   TestCountsAreRecorded_026
        === PAUSE TestCountsAreRecorded_026
        === RUN   TestCountsAreRecorded_027
        === PAUSE TestCountsAreRecorded_027
        === RUN   TestCountsAreRecorded_028
        === PAUSE TestCountsAreRecorded_028
        === RUN   TestCountsAreRecorded_029
        === PAUSE TestCountsAreRecorded_029
        === RUN   TestCountsAreRecorded_030
        === PAUSE TestCountsAreRecorded_030
        === RUN   TestCountsAreRecorded_031
        === PAUSE TestCountsAreRecorded_031
        === RUN   TestCountsAreRecorded_little_streams
        === PAUSE TestCountsAreRecorded_little_streams
        === RUN   TestCountsAreRecorded_lots_streams
        === PAUSE TestCountsAreRecorded_lots_streams
        === RUN   TestCountsAreRecorded_stderr_only
        === PAUSE TestCountsAreRecorded_stderr_only
        === RUN   TestCountsAreRecorded_panic_edges
        === PAUSE TestCountsAreRecorded_panic_edges
        === RUN   TestCountsAreRecorded_read_after_print
        === PAUSE TestCountsAreRecorded_read_after_print
        === RUN   TestCountsAreRecorded_async_counts_payload
        === PAUSE TestCountsAreRecorded_async_counts_payload
        === RUN   TestCountsAreRecorded_async_counts_stable
        === PAUSE TestCountsAreRecorded_async_counts_stable
        === RUN   TestCountsAreRecorded_async_counts_reassigned
        === PAUSE TestCountsAreRecorded_async_counts_reassigned
        === RUN   TestCountsAreRecorded_async_counts_captured
        === PAUSE TestCountsAreRecorded_async_counts_captured
        === RUN   TestCountsAreRecorded_async_counts_expression_write
        === PAUSE TestCountsAreRecorded_async_counts_expression_write
        === RUN   TestCountsAreRecorded_async_counts_parameter
        === PAUSE TestCountsAreRecorded_async_counts_parameter
        === RUN   TestCountsAreRecorded_async_counts_thrown
        === PAUSE TestCountsAreRecorded_async_counts_thrown
        === RUN   TestCountsAreRecorded_async_counts_root_error
        === PAUSE TestCountsAreRecorded_async_counts_root_error
        === RUN   TestCountsAreRecorded_async_catch_return
        === PAUSE TestCountsAreRecorded_async_catch_return
        === RUN   TestCountsAreRecorded_async_catches
        === PAUSE TestCountsAreRecorded_async_catches
        === RUN   TestCountsAreRecorded_async_closures
        === PAUSE TestCountsAreRecorded_async_closures
        === RUN   TestCountsAreRecorded_async_conditions
        === PAUSE TestCountsAreRecorded_async_conditions
        === RUN   TestCountsAreRecorded_async_control
        === PAUSE TestCountsAreRecorded_async_control
        === RUN   TestCountsAreRecorded_async_escaped_capture
        === PAUSE TestCountsAreRecorded_async_escaped_capture
        === RUN   TestCountsAreRecorded_async_frame_capture_safe
        === PAUSE TestCountsAreRecorded_async_frame_capture_safe
        === RUN   TestCountsAreRecorded_async_fresh_holder
        === PAUSE TestCountsAreRecorded_async_fresh_holder
        === RUN   TestCountsAreRecorded_async_live_values
        === PAUSE TestCountsAreRecorded_async_live_values
        === RUN   TestCountsAreRecorded_async_methods
        === PAUSE TestCountsAreRecorded_async_methods
        === RUN   TestCountsAreRecorded_async_mutation
        === PAUSE TestCountsAreRecorded_async_mutation
        === RUN   TestCountsAreRecorded_async_operand_order
        === PAUSE TestCountsAreRecorded_async_operand_order
        === RUN   TestCountsAreRecorded_async_promise_expression
        === PAUSE TestCountsAreRecorded_async_promise_expression
        === RUN   TestCountsAreRecorded_async_receiver
        === PAUSE TestCountsAreRecorded_async_receiver
        === RUN   TestCountsAreRecorded_async_snapshots
        === PAUSE TestCountsAreRecorded_async_snapshots
        === RUN   TestCountsAreRecorded_async_void_value
        === PAUSE TestCountsAreRecorded_async_void_value
        === RUN   TestCountsAreRecorded_typed_arrays_integer
        === PAUSE TestCountsAreRecorded_typed_arrays_integer
        === RUN   TestCountsAreRecorded_typed_arrays_integer_stop
        === PAUSE TestCountsAreRecorded_typed_arrays_integer_stop
        === RUN   TestCountsAreRecorded_typed_arrays_sort_uint8array
        === PAUSE TestCountsAreRecorded_typed_arrays_sort_uint8array
        === RUN   TestCountsAreRecorded_typed_arrays_sort_uint16array
        === PAUSE TestCountsAreRecorded_typed_arrays_sort_uint16array
        === RUN   TestCountsAreRecorded_typed_arrays_sort_int32array
        === PAUSE TestCountsAreRecorded_typed_arrays_sort_int32array
        === RUN   TestCountsAreRecorded_typed_arrays_sort_float64array
        === PAUSE TestCountsAreRecorded_typed_arrays_sort_float64array
        === RUN   TestCountsAreRecorded_counters_shapes
        === PAUSE TestCountsAreRecorded_counters_shapes
        === RUN   TestCountsAreRecorded_counters_square_jump
        === PAUSE TestCountsAreRecorded_counters_square_jump
        === RUN   TestCountsAreRecorded_release_fast_graph
        === PAUSE TestCountsAreRecorded_release_fast_graph
        === RUN   TestCountsAreRecorded_release_fast_shared
        === PAUSE TestCountsAreRecorded_release_fast_shared
        === RUN   TestCountsAreRecorded_flat
        === PAUSE TestCountsAreRecorded_flat
        === RUN   TestCountsAreRecorded_array
        === PAUSE TestCountsAreRecorded_array
        === RUN   TestCountsAreRecorded_async_loop_body_capture
        === PAUSE TestCountsAreRecorded_async_loop_body_capture
        === RUN   TestCountsAreRecorded_async_finally_completion
        === PAUSE TestCountsAreRecorded_async_finally_completion
        === RUN   TestCountsAreRecorded_async_labels
        === PAUSE TestCountsAreRecorded_async_labels
        === RUN   TestCountsAreRecorded_async_labels_finally
        === PAUSE TestCountsAreRecorded_async_labels_finally
        === RUN   TestCountsAreRecorded_async_labels_block_finally
        === PAUSE TestCountsAreRecorded_async_labels_block_finally
        === RUN   TestCountsAreRecorded_label_block_edges
        === PAUSE TestCountsAreRecorded_label_block_edges
        === RUN   TestCountsAreRecorded_async_iterables_cells_labels
        === PAUSE TestCountsAreRecorded_async_iterables_cells_labels
        === RUN   TestCountsAreRecorded_label_block_break_sync
        === PAUSE TestCountsAreRecorded_label_block_break_sync
        === RUN   TestCountsAreRecorded_label_block_continue_sync
        === PAUSE TestCountsAreRecorded_label_block_continue_sync
        === RUN   TestCountsAreRecorded_label_block_switch_break_sync
        === PAUSE TestCountsAreRecorded_label_block_switch_break_sync
        === RUN   TestCountsAreRecorded_label_block_outer_continue_sync
        === PAUSE TestCountsAreRecorded_label_block_outer_continue_sync
        === RUN   TestCountsAreRecorded_label_block_break_async
        === PAUSE TestCountsAreRecorded_label_block_break_async
        === RUN   TestCountsAreRecorded_label_block_continue_async
        === PAUSE TestCountsAreRecorded_label_block_continue_async
        === RUN   TestCountsAreRecorded_label_block_switch_break_async
        === PAUSE TestCountsAreRecorded_label_block_switch_break_async
        === RUN   TestCountsAreRecorded_label_block_outer_continue_async
        === PAUSE TestCountsAreRecorded_label_block_outer_continue_async
        === RUN   TestCountsAreRecorded_async_finally_return
        === PAUSE TestCountsAreRecorded_async_finally_return
        === RUN   TestCountsAreRecorded_async_finally_throw
        === PAUSE TestCountsAreRecorded_async_finally_throw
        === RUN   TestCountsAreRecorded_async_finally_break
        === PAUSE TestCountsAreRecorded_async_finally_break
        === RUN   TestCountsAreRecorded_async_finally_continue
        === PAUSE TestCountsAreRecorded_async_finally_continue
        === RUN   TestCountsAreRecorded_async_finally_void_return
        === PAUSE TestCountsAreRecorded_async_finally_void_return
        === RUN   TestCountsAreRecorded_async_finally_snapshot
        === PAUSE TestCountsAreRecorded_async_finally_snapshot
        === RUN   TestCountsAreRecorded_async_gap_label_admission
        === PAUSE TestCountsAreRecorded_async_gap_label_admission
        === RUN   TestCountsAreRecorded_async_gap_catch_admission
        === PAUSE TestCountsAreRecorded_async_gap_catch_admission
        === RUN   TestCountsAreRecorded_async_gap_finally_admission
        === PAUSE TestCountsAreRecorded_async_gap_finally_admission
        === RUN   TestCountsAreRecorded_async_gap_finally_return_admission
        === PAUSE TestCountsAreRecorded_async_gap_finally_return_admission
        === RUN   TestCountsAreRecorded_async_iterable_string_ascii_function
        === PAUSE TestCountsAreRecorded_async_iterable_string_ascii_function
        === RUN   TestCountsAreRecorded_async_for_of_string
        === PAUSE TestCountsAreRecorded_async_for_of_string
        === RUN   TestCountsAreRecorded_async_for_of_map
        === PAUSE TestCountsAreRecorded_async_for_of_map
        === RUN   TestCountsAreRecorded_async_for_of_collections_mutation
        === PAUSE TestCountsAreRecorded_async_for_of_collections_mutation
        === RUN   TestCountsAreRecorded_async_iterable_string_ascii_module
        === PAUSE TestCountsAreRecorded_async_iterable_string_ascii_module
        === RUN   TestCountsAreRecorded_async_iterable_string_emoji_function
        === PAUSE TestCountsAreRecorded_async_iterable_string_emoji_function
        === RUN   TestCountsAreRecorded_async_iterable_string_emoji_module
        === PAUSE TestCountsAreRecorded_async_iterable_string_emoji_module
        === RUN   TestCountsAreRecorded_async_iterable_string_surrogates_function
        === PAUSE TestCountsAreRecorded_async_iterable_string_surrogates_function
        === RUN   TestCountsAreRecorded_async_iterable_string_surrogates_module
        === PAUSE TestCountsAreRecorded_async_iterable_string_surrogates_module
        === RUN   TestCountsAreRecorded_async_iterable_map_entries_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_entries_function
        === RUN   TestCountsAreRecorded_async_iterable_map_entries_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_entries_module
        === RUN   TestCountsAreRecorded_async_iterable_map_keys_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_keys_function
        === RUN   TestCountsAreRecorded_async_iterable_map_keys_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_keys_module
        === RUN   TestCountsAreRecorded_async_iterable_map_values_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_values_function
        === RUN   TestCountsAreRecorded_async_iterable_map_values_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_values_module
        === RUN   TestCountsAreRecorded_async_iterable_set_values_function
        === PAUSE TestCountsAreRecorded_async_iterable_set_values_function
        === RUN   TestCountsAreRecorded_async_iterable_set_values_module
        === PAUSE TestCountsAreRecorded_async_iterable_set_values_module
        === RUN   TestCountsAreRecorded_async_iterable_set_entries_function
        === PAUSE TestCountsAreRecorded_async_iterable_set_entries_function
        === RUN   TestCountsAreRecorded_async_iterable_set_entries_module
        === PAUSE TestCountsAreRecorded_async_iterable_set_entries_module
        === RUN   TestCountsAreRecorded_async_iterable_map_clear_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_clear_function
        === RUN   TestCountsAreRecorded_async_iterable_map_clear_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_clear_module
        === RUN   TestCountsAreRecorded_async_iterable_map_readd_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_readd_function
        === RUN   TestCountsAreRecorded_async_iterable_map_readd_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_readd_module
        === RUN   TestCountsAreRecorded_async_iterable_map_rehash_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_rehash_function
        === RUN   TestCountsAreRecorded_async_iterable_map_rehash_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_rehash_module
        === RUN   TestCountsAreRecorded_async_iterable_nested_maps_function
        === PAUSE TestCountsAreRecorded_async_iterable_nested_maps_function
        === RUN   TestCountsAreRecorded_async_iterable_nested_maps_module
        === PAUSE TestCountsAreRecorded_async_iterable_nested_maps_module
        === RUN   TestCountsAreRecorded_async_iterable_map_break_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_break_function
        === RUN   TestCountsAreRecorded_async_iterable_map_break_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_break_module
        === RUN   TestCountsAreRecorded_async_iterable_map_return_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_return_function
        === RUN   TestCountsAreRecorded_async_iterable_map_return_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_return_module
        === RUN   TestCountsAreRecorded_async_iterable_string_break_function
        === PAUSE TestCountsAreRecorded_async_iterable_string_break_function
        === RUN   TestCountsAreRecorded_async_iterable_string_break_module
        === PAUSE TestCountsAreRecorded_async_iterable_string_break_module
        === RUN   TestCountsAreRecorded_async_iterable_set_return_function
        === PAUSE TestCountsAreRecorded_async_iterable_set_return_function
        === RUN   TestCountsAreRecorded_async_iterable_set_return_module
        === PAUSE TestCountsAreRecorded_async_iterable_set_return_module
        === RUN   TestCountsAreRecorded_async_for_of_string_admission
        === PAUSE TestCountsAreRecorded_async_for_of_string_admission
        === RUN   TestCountsAreRecorded_async_top_catch_binding
        === PAUSE TestCountsAreRecorded_async_top_catch_binding
        === RUN   TestCountsAreRecorded_async_top_catch_after_await
        === PAUSE TestCountsAreRecorded_async_top_catch_after_await
        === RUN   TestCountsAreRecorded_async_top_loop
        === PAUSE TestCountsAreRecorded_async_top_loop
        === RUN   TestCountsAreRecorded_async_for_of_body
        === PAUSE TestCountsAreRecorded_async_for_of_body
        === RUN   TestCountsAreRecorded_async_for_of_mutation
        === PAUSE TestCountsAreRecorded_async_for_of_mutation
        === RUN   TestCountsAreRecorded_async_for_of_return
        === PAUSE TestCountsAreRecorded_async_for_of_return
        === RUN   TestCountsAreRecorded_async_switch_body
        === PAUSE TestCountsAreRecorded_async_switch_body
        === RUN   TestCountsAreRecorded_async_loop_cells
        === PAUSE TestCountsAreRecorded_async_loop_cells
        === RUN   TestCountsAreRecorded_async_for_of_capture
        === PAUSE TestCountsAreRecorded_async_for_of_capture
        === RUN   TestCountsAreRecorded_async_for_of_cells
        === PAUSE TestCountsAreRecorded_async_for_of_cells
        === RUN   TestCountsAreRecorded_async_for_of_body_cells
        === PAUSE TestCountsAreRecorded_async_for_of_body_cells
        === RUN   TestCountsAreRecorded_async_for_of_break_cells
        === PAUSE TestCountsAreRecorded_async_for_of_break_cells
        === RUN   TestCountsAreRecorded_async_coverage_discard
        === PAUSE TestCountsAreRecorded_async_coverage_discard
        === RUN   TestCountsAreRecorded_async_coverage_parameters
        === PAUSE TestCountsAreRecorded_async_coverage_parameters
        === RUN   TestCountsAreRecorded_async_coverage_reject_eager
        === PAUSE TestCountsAreRecorded_async_coverage_reject_eager
        === RUN   TestCountsAreRecorded_async_coverage_reject_empty
        === PAUSE TestCountsAreRecorded_async_coverage_reject_empty
        === RUN   TestCountsAreRecorded_async_coverage_reject_nested
        === PAUSE TestCountsAreRecorded_async_coverage_reject_nested
        === RUN   TestCountsAreRecorded_async_coverage_typeof
        === PAUSE TestCountsAreRecorded_async_coverage_typeof
        === RUN   TestCountsAreRecorded_async_coverage_unions
        === PAUSE TestCountsAreRecorded_async_coverage_unions
        === RUN   TestCountsAreRecorded_async_coverage_values
        === PAUSE TestCountsAreRecorded_async_coverage_values
        === RUN   TestCountsAreRecorded_async_nested
        === PAUSE TestCountsAreRecorded_async_nested
        === RUN   TestCountsAreRecorded_async_plain
        === PAUSE TestCountsAreRecorded_async_plain
        === RUN   TestCountsAreRecorded_async_three
        === PAUSE TestCountsAreRecorded_async_three
        === RUN   TestCountsAreRecorded_async_throw
        === PAUSE TestCountsAreRecorded_async_throw
        === RUN   TestCountsAreRecorded_async_typeof
        === PAUSE TestCountsAreRecorded_async_typeof
        === RUN   TestCountsAreRecorded_slab_sizes
        === PAUSE TestCountsAreRecorded_slab_sizes
        === RUN   TestCountsAreRecorded_reuse_handover_fresh
        === PAUSE TestCountsAreRecorded_reuse_handover_fresh
        === RUN   TestCountsAreRecorded_async_iterable_map_boolean_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_boolean_function
        === RUN   TestCountsAreRecorded_async_iterable_set_boolean_function
        === PAUSE TestCountsAreRecorded_async_iterable_set_boolean_function
        === CONT  TestCountsAreRecorded_label_block_switch_break_sync
        === CONT  TestCountsAreRecorded_000
        === CONT  TestCountsAreRecorded_label_block_continue_sync
        === CONT  TestCountsAreRecorded_label_block_break_sync
        --- PASS: TestCountsAreRecorded_000 (0.01s)
        === CONT  TestCountsAreRecorded_async_iterables_cells_labels
        --- PASS: TestCountsAreRecorded_label_block_break_sync (0.01s)
        === CONT  TestCountsAreRecorded_label_block_edges
        --- PASS: TestCountsAreRecorded_async_iterables_cells_labels (0.01s)
        === CONT  TestCountsAreRecorded_async_labels_block_finally
        --- PASS: TestCountsAreRecorded_label_block_continue_sync (0.02s)
        === CONT  TestCountsAreRecorded_async_labels_finally
        --- PASS: TestCountsAreRecorded_async_labels_block_finally (0.00s)
        === CONT  TestCountsAreRecorded_async_labels
        --- PASS: TestCountsAreRecorded_label_block_switch_break_sync (0.03s)
        === CONT  TestCountsAreRecorded_async_finally_completion
        --- PASS: TestCountsAreRecorded_async_labels_finally (0.01s)
        === CONT  TestCountsAreRecorded_async_loop_body_capture
        --- PASS: TestCountsAreRecorded_async_finally_completion (0.01s)
        === CONT  TestCountsAreRecorded_array
        --- PASS: TestCountsAreRecorded_async_labels (0.01s)
        === CONT  TestCountsAreRecorded_flat
        --- PASS: TestCountsAreRecorded_array (0.01s)
        === CONT  TestCountsAreRecorded_release_fast_shared
        --- PASS: TestCountsAreRecorded_flat (0.01s)
        === CONT  TestCountsAreRecorded_release_fast_graph
        --- PASS: TestCountsAreRecorded_label_block_edges (0.03s)
        === CONT  TestCountsAreRecorded_counters_square_jump
        --- PASS: TestCountsAreRecorded_async_loop_body_capture (0.01s)
        === CONT  TestCountsAreRecorded_counters_shapes
        === CONT  TestCountsAreRecorded_typed_arrays_sort_float64array
        --- PASS: TestCountsAreRecorded_release_fast_shared (0.02s)
        --- PASS: TestCountsAreRecorded_release_fast_graph (0.03s)
        === CONT  TestCountsAreRecorded_typed_arrays_sort_int32array
        --- PASS: TestCountsAreRecorded_typed_arrays_sort_int32array (0.00s)
        === CONT  TestCountsAreRecorded_typed_arrays_sort_uint16array
        --- PASS: TestCountsAreRecorded_counters_square_jump (0.04s)
        === CONT  TestCountsAreRecorded_typed_arrays_sort_uint8array
        --- PASS: TestCountsAreRecorded_typed_arrays_sort_float64array (0.01s)
        === CONT  TestCountsAreRecorded_typed_arrays_integer_stop
        --- PASS: TestCountsAreRecorded_counters_shapes (0.04s)
        --- PASS: TestCountsAreRecorded_typed_arrays_sort_uint8array (0.01s)
        === CONT  TestCountsAreRecorded_typed_arrays_integer
        === CONT  TestCountsAreRecorded_async_void_value
        --- PASS: TestCountsAreRecorded_typed_arrays_sort_uint16array (0.01s)
        === CONT  TestCountsAreRecorded_async_snapshots
        --- PASS: TestCountsAreRecorded_typed_arrays_integer_stop (0.01s)
        --- PASS: TestCountsAreRecorded_typed_arrays_integer (0.01s)
        === CONT  TestCountsAreRecorded_async_receiver
        === CONT  TestCountsAreRecorded_async_promise_expression
        --- PASS: TestCountsAreRecorded_async_void_value (0.02s)
        --- PASS: TestCountsAreRecorded_async_snapshots (0.01s)
        === CONT  TestCountsAreRecorded_async_operand_order
        === CONT  TestCountsAreRecorded_async_mutation
        --- PASS: TestCountsAreRecorded_async_mutation (0.00s)
        === CONT  TestCountsAreRecorded_async_methods
        === CONT  TestCountsAreRecorded_async_live_values
        --- PASS: TestCountsAreRecorded_async_promise_expression (0.02s)
        --- PASS: TestCountsAreRecorded_async_operand_order (0.02s)
        === CONT  TestCountsAreRecorded_async_fresh_holder
        --- PASS: TestCountsAreRecorded_async_receiver (0.04s)
        === CONT  TestCountsAreRecorded_async_frame_capture_safe
        === CONT  TestCountsAreRecorded_async_escaped_capture
        --- PASS: TestCountsAreRecorded_async_frame_capture_safe (0.05s)
        --- PASS: TestCountsAreRecorded_async_escaped_capture (0.01s)
        --- PASS: TestCountsAreRecorded_async_methods (0.06s)
        === CONT  TestCountsAreRecorded_async_conditions
        --- PASS: TestCountsAreRecorded_async_live_values (0.07s)
        === CONT  TestCountsAreRecorded_async_closures
        --- PASS: TestCountsAreRecorded_async_closures (0.01s)
        === CONT  TestCountsAreRecorded_async_control
        === CONT  TestCountsAreRecorded_async_catch_return
        --- PASS: TestCountsAreRecorded_async_fresh_holder (0.07s)
        === CONT  TestCountsAreRecorded_async_catches
        --- PASS: TestCountsAreRecorded_async_conditions (0.02s)
        === CONT  TestCountsAreRecorded_async_counts_root_error
        --- PASS: TestCountsAreRecorded_async_counts_root_error (0.00s)
        --- PASS: TestCountsAreRecorded_async_catches (0.01s)
        --- PASS: TestCountsAreRecorded_async_catch_return (0.02s)
        === CONT  TestCountsAreRecorded_async_counts_thrown
        === CONT  TestCountsAreRecorded_async_counts_parameter
        === CONT  TestCountsAreRecorded_async_counts_expression_write
        === CONT  TestCountsAreRecorded_async_counts_captured
        === CONT  TestCountsAreRecorded_async_counts_reassigned
        --- PASS: TestCountsAreRecorded_async_counts_thrown (0.00s)
        --- PASS: TestCountsAreRecorded_async_counts_parameter (0.05s)
        --- PASS: TestCountsAreRecorded_async_counts_captured (0.06s)
        === CONT  TestCountsAreRecorded_async_counts_stable
        --- PASS: TestCountsAreRecorded_async_counts_reassigned (0.01s)
        === CONT  TestCountsAreRecorded_async_counts_payload
        --- PASS: TestCountsAreRecorded_async_counts_payload (0.01s)
        === CONT  TestCountsAreRecorded_read_after_print
        --- PASS: TestCountsAreRecorded_async_control (0.10s)
        === CONT  TestCountsAreRecorded_panic_edges
        --- PASS: TestCountsAreRecorded_panic_edges (0.01s)
        === CONT  TestCountsAreRecorded_stderr_only
        --- PASS: TestCountsAreRecorded_stderr_only (0.00s)
        === CONT  TestCountsAreRecorded_lots_streams
        --- PASS: TestCountsAreRecorded_lots_streams (0.00s)
        === CONT  TestCountsAreRecorded_little_streams
        --- PASS: TestCountsAreRecorded_read_after_print (0.02s)
        === CONT  TestCountsAreRecorded_031
        --- PASS: TestCountsAreRecorded_031 (0.00s)
        === CONT  TestCountsAreRecorded_030
        --- PASS: TestCountsAreRecorded_030 (0.00s)
        === CONT  TestCountsAreRecorded_029
        --- PASS: TestCountsAreRecorded_async_counts_expression_write (0.10s)
        === CONT  TestCountsAreRecorded_028
        === CONT  TestCountsAreRecorded_027
        --- PASS: TestCountsAreRecorded_little_streams (0.02s)
        === CONT  TestCountsAreRecorded_026
        --- PASS: TestCountsAreRecorded_async_counts_stable (0.05s)
        --- PASS: TestCountsAreRecorded_026 (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_set_boolean_function
        --- PASS: TestCountsAreRecorded_async_iterable_set_boolean_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_boolean_function
        --- PASS: TestCountsAreRecorded_028 (0.01s)
        === CONT  TestCountsAreRecorded_reuse_handover_fresh
        --- PASS: TestCountsAreRecorded_reuse_handover_fresh (0.00s)
        === CONT  TestCountsAreRecorded_slab_sizes
        === CONT  TestCountsAreRecorded_async_typeof
        === CONT  TestCountsAreRecorded_async_three
        --- PASS: TestCountsAreRecorded_029 (0.02s)
        --- PASS: TestCountsAreRecorded_027 (0.03s)
        --- PASS: TestCountsAreRecorded_async_iterable_map_boolean_function (0.03s)
        === CONT  TestCountsAreRecorded_async_throw
        --- PASS: TestCountsAreRecorded_async_typeof (0.01s)
        === CONT  TestCountsAreRecorded_async_plain
        --- PASS: TestCountsAreRecorded_async_plain (0.00s)
        === CONT  TestCountsAreRecorded_async_nested
        --- PASS: TestCountsAreRecorded_async_three (0.01s)
        === CONT  TestCountsAreRecorded_async_coverage_values
        --- PASS: TestCountsAreRecorded_async_nested (0.01s)
        === CONT  TestCountsAreRecorded_async_coverage_unions
        --- PASS: TestCountsAreRecorded_async_coverage_values (0.01s)
        === CONT  TestCountsAreRecorded_async_coverage_typeof
        --- PASS: TestCountsAreRecorded_async_throw (0.02s)
        === CONT  TestCountsAreRecorded_async_coverage_reject_nested
        --- PASS: TestCountsAreRecorded_async_coverage_typeof (0.01s)
        === CONT  TestCountsAreRecorded_async_coverage_reject_empty
        --- PASS: TestCountsAreRecorded_slab_sizes (0.05s)
        === CONT  TestCountsAreRecorded_async_coverage_reject_eager
        --- PASS: TestCountsAreRecorded_async_coverage_reject_empty (0.00s)
        === CONT  TestCountsAreRecorded_async_coverage_parameters
        --- PASS: TestCountsAreRecorded_async_coverage_unions (0.02s)
        === CONT  TestCountsAreRecorded_async_coverage_discard
        --- PASS: TestCountsAreRecorded_async_coverage_parameters (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_break_cells
        --- PASS: TestCountsAreRecorded_async_for_of_break_cells (0.01s)
        === CONT  TestCountsAreRecorded_async_for_of_body_cells
        --- PASS: TestCountsAreRecorded_async_coverage_discard (0.01s)
        === CONT  TestCountsAreRecorded_async_for_of_cells
        --- PASS: TestCountsAreRecorded_async_coverage_reject_nested (0.03s)
        === CONT  TestCountsAreRecorded_async_for_of_capture
        --- PASS: TestCountsAreRecorded_async_for_of_body_cells (0.01s)
        --- PASS: TestCountsAreRecorded_async_coverage_reject_eager (0.02s)
        === CONT  TestCountsAreRecorded_async_switch_body
        === CONT  TestCountsAreRecorded_async_loop_cells
        --- PASS: TestCountsAreRecorded_async_loop_cells (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_return
        --- PASS: TestCountsAreRecorded_async_for_of_cells (0.01s)
        === CONT  TestCountsAreRecorded_async_for_of_mutation
        --- PASS: TestCountsAreRecorded_async_for_of_capture (0.01s)
        === CONT  TestCountsAreRecorded_async_for_of_body
        --- PASS: TestCountsAreRecorded_async_switch_body (0.02s)
        === CONT  TestCountsAreRecorded_async_top_catch_after_await
        --- PASS: TestCountsAreRecorded_async_for_of_return (0.02s)
        === CONT  TestCountsAreRecorded_async_top_catch_binding
        --- PASS: TestCountsAreRecorded_async_top_catch_after_await (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_string_admission
        --- PASS: TestCountsAreRecorded_async_for_of_string_admission (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_set_return_module
        --- PASS: TestCountsAreRecorded_async_top_catch_binding (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_set_return_function
        --- PASS: TestCountsAreRecorded_async_for_of_mutation (0.02s)
        === CONT  TestCountsAreRecorded_async_top_loop
        --- PASS: TestCountsAreRecorded_async_iterable_set_return_function (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_string_break_module
        --- PASS: TestCountsAreRecorded_async_for_of_body (0.04s)
        --- PASS: TestCountsAreRecorded_async_iterable_set_return_module (0.03s)
        --- PASS: TestCountsAreRecorded_async_top_loop (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_string_break_function
        === CONT  TestCountsAreRecorded_async_iterable_map_return_module
        === CONT  TestCountsAreRecorded_async_iterable_map_return_function
        --- PASS: TestCountsAreRecorded_async_iterable_string_break_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_break_module
        --- PASS: TestCountsAreRecorded_async_iterable_string_break_function (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_break_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_return_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_nested_maps_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_break_module (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_nested_maps_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_return_function (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_map_rehash_module
        === CONT  TestCountsAreRecorded_async_iterable_map_rehash_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_break_function (0.02s)
        --- PASS: TestCountsAreRecorded_async_iterable_map_rehash_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_readd_module
        --- PASS: TestCountsAreRecorded_async_iterable_nested_maps_module (0.05s)
        === CONT  TestCountsAreRecorded_async_iterable_map_readd_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_rehash_function (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_map_clear_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_clear_module (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_clear_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_clear_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_set_entries_module
        --- PASS: TestCountsAreRecorded_async_iterable_nested_maps_function (0.04s)
        === CONT  TestCountsAreRecorded_async_iterable_set_entries_function
        --- PASS: TestCountsAreRecorded_async_iterable_set_entries_module (0.03s)
        === CONT  TestCountsAreRecorded_async_iterable_set_values_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_readd_module (0.05s)
        === CONT  TestCountsAreRecorded_async_iterable_set_values_function
        --- PASS: TestCountsAreRecorded_async_iterable_set_entries_function (0.03s)
        === CONT  TestCountsAreRecorded_async_iterable_map_values_module
        --- PASS: TestCountsAreRecorded_async_iterable_set_values_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_values_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_values_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_keys_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_keys_module (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_keys_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_keys_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_entries_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_values_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_entries_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_readd_function (0.06s)
        === CONT  TestCountsAreRecorded_async_iterable_string_surrogates_module
        --- PASS: TestCountsAreRecorded_async_iterable_set_values_function (0.03s)
        === CONT  TestCountsAreRecorded_async_iterable_string_surrogates_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_entries_function (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_string_emoji_module
        --- PASS: TestCountsAreRecorded_async_iterable_string_emoji_module (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_string_emoji_function
        --- PASS: TestCountsAreRecorded_async_iterable_string_surrogates_module (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_string_ascii_module
        --- PASS: TestCountsAreRecorded_async_iterable_string_ascii_module (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_collections_mutation
        --- PASS: TestCountsAreRecorded_async_iterable_string_surrogates_function (0.02s)
        === CONT  TestCountsAreRecorded_async_for_of_map
        --- PASS: TestCountsAreRecorded_async_iterable_map_entries_module (0.04s)
        === CONT  TestCountsAreRecorded_async_for_of_string
        --- PASS: TestCountsAreRecorded_async_for_of_collections_mutation (0.08s)
        === CONT  TestCountsAreRecorded_async_iterable_string_ascii_function
        === CONT  TestCountsAreRecorded_async_gap_finally_return_admission
        --- PASS: TestCountsAreRecorded_async_for_of_map (0.08s)
        --- PASS: TestCountsAreRecorded_async_gap_finally_return_admission (0.02s)
        === CONT  TestCountsAreRecorded_async_gap_catch_admission
        --- PASS: TestCountsAreRecorded_async_iterable_string_emoji_function (0.11s)
        === CONT  TestCountsAreRecorded_async_gap_finally_admission
        --- PASS: TestCountsAreRecorded_async_iterable_string_ascii_function (0.02s)
        === CONT  TestCountsAreRecorded_async_gap_label_admission
        --- PASS: TestCountsAreRecorded_async_gap_finally_admission (0.00s)
        === CONT  TestCountsAreRecorded_async_finally_snapshot
        --- PASS: TestCountsAreRecorded_async_finally_snapshot (0.00s)
        === CONT  TestCountsAreRecorded_async_finally_void_return
        --- PASS: TestCountsAreRecorded_async_gap_label_admission (0.00s)
        === CONT  TestCountsAreRecorded_async_finally_continue
        === CONT  TestCountsAreRecorded_async_finally_break
        --- PASS: TestCountsAreRecorded_async_finally_void_return (0.00s)
        === CONT  TestCountsAreRecorded_async_finally_return
        --- PASS: TestCountsAreRecorded_async_finally_break (0.00s)
        === CONT  TestCountsAreRecorded_label_block_outer_continue_async
        --- PASS: TestCountsAreRecorded_async_finally_return (0.00s)
        --- PASS: TestCountsAreRecorded_label_block_outer_continue_async (0.00s)
        === CONT  TestCountsAreRecorded_label_block_continue_async
        --- PASS: TestCountsAreRecorded_async_gap_catch_admission (0.01s)
        --- PASS: TestCountsAreRecorded_async_for_of_string (0.10s)
        === CONT  TestCountsAreRecorded_async_finally_throw
        === CONT  TestCountsAreRecorded_label_block_switch_break_async
        --- PASS: TestCountsAreRecorded_async_finally_throw (0.04s)
        === CONT  TestCountsAreRecorded_label_block_break_async
        --- PASS: TestCountsAreRecorded_async_finally_continue (0.04s)
        === CONT  TestCountsAreRecorded_label_block_outer_continue_sync
        --- PASS: TestCountsAreRecorded_label_block_switch_break_async (0.01s)
        === CONT  TestCountsAreRecorded_013
        --- PASS: TestCountsAreRecorded_013 (0.00s)
        === CONT  TestCountsAreRecorded_025
        --- PASS: TestCountsAreRecorded_label_block_outer_continue_sync (0.01s)
        === CONT  TestCountsAreRecorded_024
        --- PASS: TestCountsAreRecorded_024 (0.00s)
        === CONT  TestCountsAreRecorded_023
        --- PASS: TestCountsAreRecorded_023 (0.01s)
        === CONT  TestCountsAreRecorded_022
        --- PASS: TestCountsAreRecorded_label_block_continue_async (0.06s)
        === CONT  TestCountsAreRecorded_021
        --- PASS: TestCountsAreRecorded_025 (0.02s)
        === CONT  TestCountsAreRecorded_020
        --- PASS: TestCountsAreRecorded_020 (0.00s)
        === CONT  TestCountsAreRecorded_019
        --- PASS: TestCountsAreRecorded_label_block_break_async (0.04s)
        === CONT  TestCountsAreRecorded_018
        --- PASS: TestCountsAreRecorded_021 (0.02s)
        === CONT  TestCountsAreRecorded_017
        --- PASS: TestCountsAreRecorded_022 (0.02s)
        === CONT  TestCountsAreRecorded_016
        === CONT  TestCountsAreRecorded_015
        --- PASS: TestCountsAreRecorded_019 (0.01s)
        --- PASS: TestCountsAreRecorded_017 (0.01s)
        === CONT  TestCountsAreRecorded_014
        --- PASS: TestCountsAreRecorded_016 (0.01s)
        === CONT  TestCountsAreRecorded_012
        --- PASS: TestCountsAreRecorded_014 (0.01s)
        === CONT  TestCountsAreRecorded_011
        --- PASS: TestCountsAreRecorded_018 (0.01s)
        === CONT  TestCountsAreRecorded_007
        --- PASS: TestCountsAreRecorded_015 (0.01s)
        === CONT  TestCountsAreRecorded_010
        --- PASS: TestCountsAreRecorded_012 (0.01s)
        === CONT  TestCountsAreRecorded_009
        --- PASS: TestCountsAreRecorded_011 (0.00s)
        === CONT  TestCountsAreRecorded_008
        --- PASS: TestCountsAreRecorded_008 (0.00s)
        === CONT  TestCountsAreRecorded_004
        --- PASS: TestCountsAreRecorded_004 (0.00s)
        === CONT  TestCountsAreRecorded_006
        --- PASS: TestCountsAreRecorded_007 (0.01s)
        === CONT  TestCountsAreRecorded_005
        --- PASS: TestCountsAreRecorded_005 (0.00s)
        === CONT  TestCountsAreRecorded_002
        --- PASS: TestCountsAreRecorded_002 (0.00s)
        === CONT  TestCountsAreRecorded_003
        --- PASS: TestCountsAreRecorded_010 (0.04s)
        === CONT  TestCountsAreRecorded_001
        --- PASS: TestCountsAreRecorded_001 (0.00s)
        --- PASS: TestCountsAreRecorded_009 (0.05s)
        --- PASS: TestCountsAreRecorded_006 (0.07s)
        === NAME  TestCountsAreRecorded_003
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
--- FAIL: TestRecordSpreadFromMissingAgreesWithNode (0.00s)
    --- FAIL: TestRecordSpreadFromMissingAgreesWithNode/tracing_write_stack_event.ts (0.50s)
        record_spread_missing_test.go:31: gate cache node miss
        record_spread_missing_test.go:31: /workspace/adamic/internal/oracle/testdata/record_spread_missing/tracing_write_stack_event.ts:14:37: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet
FAIL
gate cache: native hits=3085 misses=6612
gate cache: node hits=1745 misses=6552
gate cache: probe hits=0 misses=40
FAIL	github.com/system-inc/adamic/internal/oracle	4063.564s
FAIL
```

## Step 6: full admission delta

```sh
python3 cloud/admission-corpus/manifest.py --sha a38bdd1290858572341e397d51331c662f52ef0c > /tmp/adamic-proof/manifest.json
go build -o /tmp/adamic-proof/admission-delta ./cmd/adamic-admission-delta
/tmp/adamic-proof/admission-delta \
  --base efe887763f26e96653546fae904d57539139f7ae \
  --head a38bdd1290858572341e397d51331c662f52ef0c \
  --manifest /tmp/adamic-proof/manifest.json \
  --manifest-generator cloud/admission-corpus/manifest.py \
  --workers 2 --json > /tmp/adamic-proof/admission.json 2> /tmp/adamic-proof/admission.stderr
```

Gate build: exit=0 wall_seconds=4. Gate: exit=0 wall_seconds=1663. Nothing else heavy was running. No corpus filter, shard, or runtime budget was used for this full run.

```json
{
  "base": "efe887763f26e96653546fae904d57539139f7ae",
  "head": "a38bdd1290858572341e397d51331c662f52ef0c",
  "verdict": "pass",
  "void": false,
  "diff_count": 7,
  "admitted": 6,
  "workers": 2,
  "classification": "lowering-only",
  "complete_corpus": true,
  "compile_timeout_seconds": 45,
  "runtime_timeout_seconds": 10,
  "sampling_size": 121,
  "omitted": 0,
  "budget_seconds": 0,
  "negative_witness_count": 50,
  "accepted_witnesses": 50,
  "node_agreements": 19,
  "runtime_notyet_count": 52,
  "ruled_runtime_stops": 52,
  "stale_runtime_stops": [],
  "pending_runtime_stops": [],
  "timed_out_runs": 2,
  "retried_runs": 2,
  "still_timed_out_runs": 0,
  "generator_revision": "a38bdd1290858572341e397d51331c662f52ef0c",
  "generator_blob": "09d061a03b1e016387e975d25ee0cd5b75ac6c1f",
  "manifest_blob": "4e42e0747ced72f1873d002f4d02ba373e64c895",
  "phase_seconds": {
    "base_build": 57.075336533,
    "checkout": 17.423563374,
    "checkout_and_provision": 19.457155409,
    "classification": 891.840475634,
    "head_build": 32.256310595,
    "manifest_and_verification": 27.436043329,
    "provision": 2.033506874,
    "runtime": 634.051469148,
    "total": 1662.126091936
  },
  "void_reasons": [],
  "newly_refused": 0,
  "classes": {
    "newly-accepted": 6,
    "refused-by-both": 1952,
    "accepted-by-both": 3282
  },
  "corpus_counts": {
    "diff": 7,
    "witnesses": 2,
    "fixtures": 4292,
    "gaps": 105,
    "review": 20,
    "fuzz": 0,
    "stage3": 821
  }
}
```

Full-run stderr:

```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 2 runs timed out, 2 retried alone, 0 timed out again
changed programs: 7
```

All seven changed programs (paths and exact blobs):

```json
[
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
    "blob": "dd106e17996fb1ff71854cef950c2c2d4bd42493"
  },
  {
    "path": "internal/oracle/testdata/record_spread_missing/undefined_args.a",
    "blob": "16fc8a95544b18e3e1876e61ebfd86d25456a6a2"
  },
  {
    "path": "internal/oracle/testdata/record_spread_missing/undefined_args.ts",
    "blob": "16fc8a95544b18e3e1876e61ebfd86d25456a6a2"
  }
]
```

### Every classification stderr difference

Exactly 6 programs differ. The following strings are the complete, unnormalized stderr texts, including their original temporary checkout prefix and trailing newline. HEAD stderr is empty for all six. No other program has a classification stderr difference.

#### internal/oracle/testdata/record_spread_missing/null_args.a

```json
{
  "path": "internal/oracle/testdata/record_spread_missing/null_args.a",
  "class": "newly-accepted",
  "main": {
    "wall_seconds": 0.135039901,
    "exit": 1,
    "stdout": "",
    "stderr": "adamic: /tmp/adamic-gate/admission-delta-2703529198/head/internal/oracle/testdata/record_spread_missing/null_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet\n"
  },
  "HEAD": {
    "wall_seconds": 0.087506163,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "runtime_agree": true
}
```

#### internal/oracle/testdata/record_spread_missing/null_args.ts

```json
{
  "path": "internal/oracle/testdata/record_spread_missing/null_args.ts",
  "class": "newly-accepted",
  "main": {
    "wall_seconds": 0.145274058,
    "exit": 1,
    "stdout": "",
    "stderr": "adamic: /tmp/adamic-gate/admission-delta-2703529198/head/internal/oracle/testdata/record_spread_missing/null_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet\n"
  },
  "HEAD": {
    "wall_seconds": 0.064301928,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "runtime_agree": true
}
```

#### internal/oracle/testdata/record_spread_missing/overwritten_key.a

```json
{
  "path": "internal/oracle/testdata/record_spread_missing/overwritten_key.a",
  "class": "newly-accepted",
  "main": {
    "wall_seconds": 0.1250702,
    "exit": 1,
    "stdout": "",
    "stderr": "adamic: /tmp/adamic-gate/admission-delta-2703529198/head/internal/oracle/testdata/record_spread_missing/overwritten_key.a:8:11: stage 0 can't lower a spread into record storage from a fixed object yet\n"
  },
  "HEAD": {
    "wall_seconds": 0.134388572,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "runtime_agree": true
}
```

#### internal/oracle/testdata/record_spread_missing/overwritten_key.ts

```json
{
  "path": "internal/oracle/testdata/record_spread_missing/overwritten_key.ts",
  "class": "newly-accepted",
  "main": {
    "wall_seconds": 0.147993993,
    "exit": 1,
    "stdout": "",
    "stderr": "adamic: /tmp/adamic-gate/admission-delta-2703529198/head/internal/oracle/testdata/record_spread_missing/overwritten_key.ts:8:11: stage 0 can't lower a spread into record storage from a fixed object yet\n"
  },
  "HEAD": {
    "wall_seconds": 0.110769589,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "runtime_agree": true
}
```

#### internal/oracle/testdata/record_spread_missing/undefined_args.a

```json
{
  "path": "internal/oracle/testdata/record_spread_missing/undefined_args.a",
  "class": "newly-accepted",
  "main": {
    "wall_seconds": 0.066988755,
    "exit": 1,
    "stdout": "",
    "stderr": "adamic: /tmp/adamic-gate/admission-delta-2703529198/head/internal/oracle/testdata/record_spread_missing/undefined_args.a:7:11: stage 0 can't lower a spread into record storage from a fixed object yet\n"
  },
  "HEAD": {
    "wall_seconds": 0.066646169,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "runtime_agree": true
}
```

#### internal/oracle/testdata/record_spread_missing/undefined_args.ts

```json
{
  "path": "internal/oracle/testdata/record_spread_missing/undefined_args.ts",
  "class": "newly-accepted",
  "main": {
    "wall_seconds": 0.066870538,
    "exit": 1,
    "stdout": "",
    "stderr": "adamic: /tmp/adamic-gate/admission-delta-2703529198/head/internal/oracle/testdata/record_spread_missing/undefined_args.ts:7:11: stage 0 can't lower a spread into record storage from a fixed object yet\n"
  },
  "HEAD": {
    "wall_seconds": 0.059435787,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "runtime_agree": true
}
```

### Every timeout, rerun alone at both SHAs

The full run reports two initial timeouts, each `repair native compile`: `internal/oracle/testdata/views_v3/for-of-lie.ts` and `internal/oracle/testdata/views_v3/generic-object-lie.ts`. Both were retried alone automatically, neither timed out again, and both have `repair_agree=true`. There were no classification timeouts or void programs. In addition, each witness and its repair was rerun alone at both exact SHAs through a one-program manifest with base=head at that SHA. These four independent runs ran sequentially, with the original 45 s classification / 10 s runtime command limits and two workers. Their verdict is `partial` because the corpus is deliberately restricted, not because anything failed. All exit 0, with zero timeouts, `void=false`, accepted witness and agreeing repair.

```json
{
  "path": "internal/oracle/testdata/views_v3/for-of-lie.ts",
  "initial_timed_out": [
    "repair native compile"
  ],
  "automatically_retried": true,
  "repair_agree_after_retry": true
}
```

#### internal/oracle/testdata/views_v3/for-of-lie.ts — head

```json
{
  "case": "internal/oracle/testdata/views_v3/for-of-lie.ts",
  "revision": "head",
  "sha": "a38bdd1290858572341e397d51331c662f52ef0c",
  "command": [
    "/tmp/adamic-proof/admission-delta",
    "--base",
    "a38bdd1290858572341e397d51331c662f52ef0c",
    "--head",
    "a38bdd1290858572341e397d51331c662f52ef0c",
    "--manifest",
    "/tmp/adamic-proof/timeout-for-of-lie-head-manifest.json",
    "--corpus-filter",
    "single",
    "--workers",
    "2",
    "--json"
  ],
  "exit": 0,
  "wall_seconds": 24.113311508999686
}
```

```json
{
  "verdict": "partial",
  "void": false,
  "diff_count": 0,
  "admitted": 0,
  "timed_out_runs": 0,
  "retried_runs": 0,
  "still_timed_out_runs": 0,
  "programs": [
    {
      "path": "internal/oracle/testdata/views_v3/for-of-lie.ts",
      "blob": "664e24e42ea334d7c184a8b028a9fa4533406fe8",
      "corpus": "single",
      "class": "accepted-by-both",
      "base_compile": {
        "wall_seconds": 0.184615922,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "compile": {
        "wall_seconds": 0.184615922,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "node": {
        "wall_seconds": 0.126168256,
        "exit": 0,
        "stdout": "before\nbody\n",
        "stderr": ""
      },
      "javascript": {
        "wall_seconds": 0.059209607,
        "exit": 70,
        "stdout": "before\n",
        "stderr": "adamic: panic: element read failed: view._children's element, in for...of is not a Node; expected Node, found number\n"
      },
      "native": {
        "wall_seconds": 0.004298987,
        "exit": 70,
        "stdout": "before\n",
        "stderr": "adamic: panic: element read failed: view._children's element, in for...of is not a Node; expected Node, found number\n"
      },
      "javascript_compile": {
        "wall_seconds": 0.079053786,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "native_compile": {
        "wall_seconds": 2.463136584,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "agree": false,
      "negative_related": true,
      "negative_witness": true,
      "notyet_related": false,
      "ruled_runtime_stop": false,
      "ruled_backend_divergence": false,
      "repair_agree": true,
      "sampled": true
    }
  ]
}
```

```text
negative witnesses: 50
ruled run-time stops: 0
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 0
```

#### internal/oracle/testdata/views_v3/for-of-lie.ts — main

```json
{
  "case": "internal/oracle/testdata/views_v3/for-of-lie.ts",
  "revision": "main",
  "sha": "efe887763f26e96653546fae904d57539139f7ae",
  "command": [
    "/tmp/adamic-proof/admission-delta",
    "--base",
    "efe887763f26e96653546fae904d57539139f7ae",
    "--head",
    "efe887763f26e96653546fae904d57539139f7ae",
    "--manifest",
    "/tmp/adamic-proof/timeout-for-of-lie-main-manifest.json",
    "--corpus-filter",
    "single",
    "--workers",
    "2",
    "--json"
  ],
  "exit": 0,
  "wall_seconds": 17.722110290000273
}
```

```json
{
  "verdict": "partial",
  "void": false,
  "diff_count": 0,
  "admitted": 0,
  "timed_out_runs": 0,
  "retried_runs": 0,
  "still_timed_out_runs": 0,
  "programs": [
    {
      "path": "internal/oracle/testdata/views_v3/for-of-lie.ts",
      "blob": "664e24e42ea334d7c184a8b028a9fa4533406fe8",
      "corpus": "single",
      "class": "accepted-by-both",
      "base_compile": {
        "wall_seconds": 0.102374328,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "compile": {
        "wall_seconds": 0.102374328,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "node": {
        "wall_seconds": 0.230931166,
        "exit": 0,
        "stdout": "before\nbody\n",
        "stderr": ""
      },
      "javascript": {
        "wall_seconds": 0.060953487,
        "exit": 70,
        "stdout": "before\n",
        "stderr": "adamic: panic: element read failed: view._children's element, in for...of is not a Node; expected Node, found number\n"
      },
      "native": {
        "wall_seconds": 0.003671892,
        "exit": 70,
        "stdout": "before\n",
        "stderr": "adamic: panic: element read failed: view._children's element, in for...of is not a Node; expected Node, found number\n"
      },
      "javascript_compile": {
        "wall_seconds": 0.091707254,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "native_compile": {
        "wall_seconds": 3.190979168,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "agree": false,
      "negative_related": true,
      "negative_witness": true,
      "notyet_related": false,
      "ruled_runtime_stop": false,
      "ruled_backend_divergence": false,
      "repair_agree": true,
      "sampled": true
    }
  ]
}
```

```text
negative witnesses: 50
ruled run-time stops: 0
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 0
```

```json
{
  "path": "internal/oracle/testdata/views_v3/generic-object-lie.ts",
  "initial_timed_out": [
    "repair native compile"
  ],
  "automatically_retried": true,
  "repair_agree_after_retry": true
}
```

#### internal/oracle/testdata/views_v3/generic-object-lie.ts — head

```json
{
  "case": "internal/oracle/testdata/views_v3/generic-object-lie.ts",
  "revision": "head",
  "sha": "a38bdd1290858572341e397d51331c662f52ef0c",
  "command": [
    "/tmp/adamic-proof/admission-delta",
    "--base",
    "a38bdd1290858572341e397d51331c662f52ef0c",
    "--head",
    "a38bdd1290858572341e397d51331c662f52ef0c",
    "--manifest",
    "/tmp/adamic-proof/timeout-generic-object-lie-head-manifest.json",
    "--corpus-filter",
    "single",
    "--workers",
    "2",
    "--json"
  ],
  "exit": 0,
  "wall_seconds": 18.207532377999087
}
```

```json
{
  "verdict": "partial",
  "void": false,
  "diff_count": 0,
  "admitted": 0,
  "timed_out_runs": 0,
  "retried_runs": 0,
  "still_timed_out_runs": 0,
  "programs": [
    {
      "path": "internal/oracle/testdata/views_v3/generic-object-lie.ts",
      "blob": "211fbfbd03e643888c37a2e6fc8c37fd96c72b9c",
      "corpus": "single",
      "class": "accepted-by-both",
      "base_compile": {
        "wall_seconds": 0.074521161,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "compile": {
        "wall_seconds": 0.074521161,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "node": {
        "wall_seconds": 0.137963436,
        "exit": 0,
        "stdout": "before\nafter\n",
        "stderr": ""
      },
      "javascript": {
        "wall_seconds": 0.109051914,
        "exit": 70,
        "stdout": "before\n",
        "stderr": "adamic: panic: element read failed: xs[0] is not a Node; expected Node, found object\n"
      },
      "native": {
        "wall_seconds": 0.002259936,
        "exit": 70,
        "stdout": "before\n",
        "stderr": "adamic: panic: element read failed: xs[0] is not a Node; expected Node, found object\n"
      },
      "javascript_compile": {
        "wall_seconds": 0.170353414,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "native_compile": {
        "wall_seconds": 2.5159759790000003,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "agree": false,
      "negative_related": true,
      "negative_witness": true,
      "notyet_related": false,
      "ruled_runtime_stop": false,
      "ruled_backend_divergence": false,
      "repair_agree": true,
      "sampled": true
    }
  ]
}
```

```text
negative witnesses: 50
ruled run-time stops: 0
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 0
```

#### internal/oracle/testdata/views_v3/generic-object-lie.ts — main

```json
{
  "case": "internal/oracle/testdata/views_v3/generic-object-lie.ts",
  "revision": "main",
  "sha": "efe887763f26e96653546fae904d57539139f7ae",
  "command": [
    "/tmp/adamic-proof/admission-delta",
    "--base",
    "efe887763f26e96653546fae904d57539139f7ae",
    "--head",
    "efe887763f26e96653546fae904d57539139f7ae",
    "--manifest",
    "/tmp/adamic-proof/timeout-generic-object-lie-main-manifest.json",
    "--corpus-filter",
    "single",
    "--workers",
    "2",
    "--json"
  ],
  "exit": 0,
  "wall_seconds": 16.439963785000145
}
```

```json
{
  "verdict": "partial",
  "void": false,
  "diff_count": 0,
  "admitted": 0,
  "timed_out_runs": 0,
  "retried_runs": 0,
  "still_timed_out_runs": 0,
  "programs": [
    {
      "path": "internal/oracle/testdata/views_v3/generic-object-lie.ts",
      "blob": "211fbfbd03e643888c37a2e6fc8c37fd96c72b9c",
      "corpus": "single",
      "class": "accepted-by-both",
      "base_compile": {
        "wall_seconds": 0.088077786,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "compile": {
        "wall_seconds": 0.088077786,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "node": {
        "wall_seconds": 0.136798925,
        "exit": 0,
        "stdout": "before\nafter\n",
        "stderr": ""
      },
      "javascript": {
        "wall_seconds": 0.059947307,
        "exit": 70,
        "stdout": "before\n",
        "stderr": "adamic: panic: element read failed: xs[0] is not a Node; expected Node, found object\n"
      },
      "native": {
        "wall_seconds": 0.003891184,
        "exit": 70,
        "stdout": "before\n",
        "stderr": "adamic: panic: element read failed: xs[0] is not a Node; expected Node, found object\n"
      },
      "javascript_compile": {
        "wall_seconds": 0.072312178,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "native_compile": {
        "wall_seconds": 2.529306555,
        "exit": 0,
        "stdout": "",
        "stderr": ""
      },
      "agree": false,
      "negative_related": true,
      "negative_witness": true,
      "notyet_related": false,
      "ruled_runtime_stop": false,
      "ruled_backend_divergence": false,
      "repair_agree": true,
      "sampled": true
    }
  ]
}
```

```text
negative witnesses: 50
ruled run-time stops: 0
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 0
```

## Integrity and delivery

After all tests and temporary main comparisons: HEAD was again exactly `a38bdd1290858572341e397d51331c662f52ef0c`; `git status --short`, `git diff --exit-code`, and `git diff --cached --exit-code` were clean. The pinned upstream checkout also had clean status. The only final repository artifact is this report. The report is committed on the new `proof/lower-bzr-3-report` branch based on HEAD; only this branch is authorized for push.

External full admission JSON (not committed): `/tmp/adamic-proof/admission.json`; SHA-256 `01cefb755e0d5062d36f97c4d9b2bf630dd43c3369e49f513ffeaaa3efe00d6b`. It contains all 5,240 classifications. The requested verdict, counters, complete stderr differences and timeout rerun evidence are embedded above so those results do not depend on access to temporary files.

