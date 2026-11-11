# Captured Union proof — #n7sx9gz

HEAD: `9d31ad142e9fc1cf22fe0565951e86ccf4af8679`; base main: `d43437079a7e7ca723168e67c5efb5bf7ebee216`. Run date: 2026-11-11 UTC. Commands ran sequentially, one heavy command at a time. No compiler changes were retained.

## Decisive results

HEAD passes all 16 lowering pin/dialect subtests and the maybe-boolean stop guard, and all 64 oracle backend subtests. All HEAD counted runs balance. M1 and M2 stop every Union pin during lowering. M3 is caught by sanitized and counted runs of by_value, released and alternating, in both dialects. M4 is caught by sanitized and counted runs of class, object, nested and graph, in both dialects; cycle stays green.

The full lowering suite passes (exit 0, 3012 s), as does the full IR guard suite (exit 0, 250 s). The admission delta is **void**, gate exit 3 (`go run` exit 1): 8 newly admitted, zero newly refused, diff_count 8. Its only persistent timeout is alternating.ts native compilation. Explicit isolated reruns do not time out, but alternating measures 9.166890 s and 14.594753 s on separate runs across the 10 s gate cutoff. The original gate remains void.

## Environment and step 1

Loaded `/workspace/adamic-tools/env.sh` first; created `/tmp/adamic-gate` outside the repository. Fetched `origin compiler/captured-union-4`, checked out the requested HEAD detached, and completed `git submodule update --init --recursive` (exit 0). Cohere is `5360753e19e4417675a90e92a1823bea66b13918`; its TypeScript submodule is `d92d9bfee114c80be2c375d72edae966176e3a4f`.

Provisioned upstream TypeScript outside the repository at `/tmp/n7proof/typescript`, checked out detached at `050880ce59e30b356b686bd3144efe24f875ebc8`, with `ADAMIC_TYPESCRIPT_SOURCE` exported. Copied stage3/api’s package.json and package-lock.json to `/tmp/n7proof/api` and ran npm ci there (exit 0). Its node_modules/.bin precedes PATH; `tsc --version` is `Version 6.0.3`.

Every proof command sourced the shared environment, exporting `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true`, inherited by npm and gate children. Logs and tool products were kept outside the repository. Wall times below include command startup/build time and are seconds unless stated otherwise.

```text
gofmt exit=0 wall=1
stdout/stderr: empty (0 bytes)
vet exit=0 wall=413
stdout/stderr: empty
```


## Step 2a: HEAD lowering pins

Command: `go test ./internal/lower -run 'TestCapturedUnionLowersToOneCell|TestCapturedMaybeBooleanStillStops' -count=1 -v -timeout 30m`.

```text
head-a exit=0 wall=61
```


Complete verbatim test log, including each pin’s PASS, each graph-cell line and each closure-carry line:

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== CONT  TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== CONT  TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== RUN   TestCapturedUnionLowersToOneCell/cycle.ts
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.ts
=== RUN   TestCapturedUnionLowersToOneCell/cycle.a
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
--- PASS: TestCapturedMaybeBooleanStillStops (1.13s)
=== CONT  TestCapturedUnionLowersToOneCell/cycle.a
=== NAME  TestCapturedUnionLowersToOneCell/class.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/cycle.ts
=== NAME  TestCapturedUnionLowersToOneCell/alternating.ts
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
=== NAME  TestCapturedUnionLowersToOneCell/nested.ts
    captured_union_test.go:56: extended: graph cell true
    captured_union_test.go:66: extended: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== NAME  TestCapturedUnionLowersToOneCell/graph.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
=== NAME  TestCapturedUnionLowersToOneCell/cycle.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 1 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== NAME  TestCapturedUnionLowersToOneCell/cycle.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 1 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== NAME  TestCapturedUnionLowersToOneCell/nested.a
    captured_union_test.go:56: extended: graph cell true
    captured_union_test.go:66: extended: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
=== NAME  TestCapturedUnionLowersToOneCell/alternating.a
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
=== NAME  TestCapturedUnionLowersToOneCell/released.a
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/object.a
=== NAME  TestCapturedUnionLowersToOneCell/graph.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
=== NAME  TestCapturedUnionLowersToOneCell/released.ts
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== NAME  TestCapturedUnionLowersToOneCell/by_value.a
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 3 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/object.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/object.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/class.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/by_value.ts
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 3 closures carry its cell
--- PASS: TestCapturedUnionLowersToOneCell (0.00s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.ts (1.12s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.ts (1.14s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.ts (1.15s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.ts (1.40s)
    --- PASS: TestCapturedUnionLowersToOneCell/cycle.ts (0.64s)
    --- PASS: TestCapturedUnionLowersToOneCell/cycle.a (0.64s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.a (0.67s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.a (0.52s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.a (0.18s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.a (0.87s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.ts (0.31s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.a (0.26s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.a (0.22s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.ts (0.15s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.a (0.10s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.ts (0.34s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	2.392s
```


## Step 2b: HEAD oracle pins

Command: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestCapturedUnionAgrees' -count=1 -v -timeout 40m`.

```text
head-b exit=0 wall=616
```


Every backend result and each counts line follows in the complete verbatim log. No HEAD pin failed, so there is no failing got/want to quote.

```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== CONT  TestCapturedUnionAgrees
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/cycle.ts
=== PAUSE TestCapturedUnionAgrees/cycle.ts
=== RUN   TestCapturedUnionAgrees/cycle.a
=== PAUSE TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/cycle.ts
=== CONT  TestCapturedUnionAgrees/graph.a
=== CONT  TestCapturedUnionAgrees/released.a
=== CONT  TestCapturedUnionAgrees/cycle.a
=== RUN   TestCapturedUnionAgrees/graph.a/native
=== RUN   TestCapturedUnionAgrees/released.a/native
=== RUN   TestCapturedUnionAgrees/cycle.ts/native
=== RUN   TestCapturedUnionAgrees/class.ts/native
=== RUN   TestCapturedUnionAgrees/cycle.a/native
=== RUN   TestCapturedUnionAgrees/released.a/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.a/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.ts/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.a/counted
=== RUN   TestCapturedUnionAgrees/class.ts/sanitized
=== RUN   TestCapturedUnionAgrees/released.a/counted
=== RUN   TestCapturedUnionAgrees/cycle.ts/counted
=== NAME  TestCapturedUnionAgrees/cycle.a/counted
    captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.a/javascript
=== NAME  TestCapturedUnionAgrees/released.a/counted
    captured_union_test.go:116: counts: allocations 96, frees 96, in regions 0
=== RUN   TestCapturedUnionAgrees/released.a/javascript
=== NAME  TestCapturedUnionAgrees/cycle.ts/counted
    captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.ts/javascript
=== RUN   TestCapturedUnionAgrees/class.ts/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.ts/javascript
=== RUN   TestCapturedUnionAgrees/graph.a/sanitized
=== RUN   TestCapturedUnionAgrees/graph.a/counted
    captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.a/javascript
=== NAME  TestCapturedUnionAgrees/cycle.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 456.65
=== NAME  TestCapturedUnionAgrees/cycle.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a miss 456.65
=== NAME  TestCapturedUnionAgrees/graph.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 195.51
=== NAME  TestCapturedUnionAgrees/released.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 456.66
=== NAME  TestCapturedUnionAgrees/class.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 456.66
=== CONT  TestCapturedUnionAgrees/released.ts
=== CONT  TestCapturedUnionAgrees/by_value.a
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== CONT  TestCapturedUnionAgrees/object.a
=== CONT  TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.ts/native
=== RUN   TestCapturedUnionAgrees/by_value.ts/native
=== RUN   TestCapturedUnionAgrees/released.ts/native
=== RUN   TestCapturedUnionAgrees/object.a/native
=== RUN   TestCapturedUnionAgrees/by_value.a/native
=== RUN   TestCapturedUnionAgrees/by_value.a/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.ts/sanitized
=== RUN   TestCapturedUnionAgrees/released.ts/sanitized
=== RUN   TestCapturedUnionAgrees/object.a/sanitized
=== RUN   TestCapturedUnionAgrees/object.ts/sanitized
=== RUN   TestCapturedUnionAgrees/released.ts/counted
=== RUN   TestCapturedUnionAgrees/by_value.a/counted
=== RUN   TestCapturedUnionAgrees/by_value.ts/counted
=== NAME  TestCapturedUnionAgrees/released.ts/counted
    captured_union_test.go:116: counts: allocations 96, frees 96, in regions 0
=== RUN   TestCapturedUnionAgrees/released.ts/javascript
=== RUN   TestCapturedUnionAgrees/object.a/counted
=== NAME  TestCapturedUnionAgrees/released.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.04
=== NAME  TestCapturedUnionAgrees/by_value.ts/counted
    captured_union_test.go:116: counts: allocations 35, frees 35, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== CONT  TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts/counted
=== NAME  TestCapturedUnionAgrees/by_value.a/counted
    captured_union_test.go:116: counts: allocations 35, frees 35, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.03
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== CONT  TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/class.a/native
=== RUN   TestCapturedUnionAgrees/graph.ts/native
=== RUN   TestCapturedUnionAgrees/alternating.ts/native
=== NAME  TestCapturedUnionAgrees/object.a/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== NAME  TestCapturedUnionAgrees/object.ts/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.ts/javascript
=== CONT  TestCapturedUnionAgrees/alternating.a
=== NAME  TestCapturedUnionAgrees/object.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== RUN   TestCapturedUnionAgrees/class.a/sanitized
=== CONT  TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/graph.ts/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.a/native
=== RUN   TestCapturedUnionAgrees/alternating.ts/sanitized
=== RUN   TestCapturedUnionAgrees/nested.a/native
=== RUN   TestCapturedUnionAgrees/alternating.a/sanitized
=== RUN   TestCapturedUnionAgrees/class.a/counted
=== RUN   TestCapturedUnionAgrees/nested.a/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.ts/counted
=== RUN   TestCapturedUnionAgrees/graph.ts/counted
=== RUN   TestCapturedUnionAgrees/alternating.a/counted
=== NAME  TestCapturedUnionAgrees/class.a/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.05
=== CONT  TestCapturedUnionAgrees/nested.ts
=== NAME  TestCapturedUnionAgrees/graph.ts/counted
    captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.ts/javascript
=== NAME  TestCapturedUnionAgrees/alternating.ts/counted
    captured_union_test.go:116: counts: allocations 54, frees 54, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.ts/javascript
=== NAME  TestCapturedUnionAgrees/graph.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== NAME  TestCapturedUnionAgrees/alternating.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.04
=== RUN   TestCapturedUnionAgrees/nested.a/counted
=== NAME  TestCapturedUnionAgrees/alternating.a/counted
    captured_union_test.go:116: counts: allocations 54, frees 54, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== RUN   TestCapturedUnionAgrees/nested.ts/native
=== NAME  TestCapturedUnionAgrees/nested.a/counted
    captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== RUN   TestCapturedUnionAgrees/nested.ts/sanitized
=== RUN   TestCapturedUnionAgrees/nested.ts/counted
    captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
--- PASS: TestCapturedUnionAgrees (0.00s)
    --- PASS: TestCapturedUnionAgrees/released.a (580.72s)
        --- PASS: TestCapturedUnionAgrees/released.a/native (0.58s)
        --- PASS: TestCapturedUnionAgrees/released.a/sanitized (1.20s)
        --- PASS: TestCapturedUnionAgrees/released.a/counted (0.94s)
        --- PASS: TestCapturedUnionAgrees/released.a/javascript (457.91s)
    --- PASS: TestCapturedUnionAgrees/graph.a (580.75s)
        --- PASS: TestCapturedUnionAgrees/graph.a/native (100.91s)
        --- PASS: TestCapturedUnionAgrees/graph.a/sanitized (98.81s)
        --- PASS: TestCapturedUnionAgrees/graph.a/counted (65.13s)
        --- PASS: TestCapturedUnionAgrees/graph.a/javascript (195.84s)
    --- PASS: TestCapturedUnionAgrees/cycle.ts (580.76s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/native (0.75s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/sanitized (1.54s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/counted (0.61s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/javascript (457.76s)
    --- PASS: TestCapturedUnionAgrees/cycle.a (580.79s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/native (0.43s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/sanitized (0.80s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/counted (1.00s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/javascript (458.31s)
    --- PASS: TestCapturedUnionAgrees/class.ts (580.80s)
        --- PASS: TestCapturedUnionAgrees/class.ts/native (1.50s)
        --- PASS: TestCapturedUnionAgrees/class.ts/sanitized (1.58s)
        --- PASS: TestCapturedUnionAgrees/class.ts/counted (0.51s)
        --- PASS: TestCapturedUnionAgrees/class.ts/javascript (457.06s)
    --- PASS: TestCapturedUnionAgrees/released.ts (2.65s)
        --- PASS: TestCapturedUnionAgrees/released.ts/native (0.49s)
        --- PASS: TestCapturedUnionAgrees/released.ts/sanitized (0.91s)
        --- PASS: TestCapturedUnionAgrees/released.ts/counted (0.42s)
        --- PASS: TestCapturedUnionAgrees/released.ts/javascript (0.19s)
    --- PASS: TestCapturedUnionAgrees/by_value.ts (2.71s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/native (0.47s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/sanitized (1.19s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/counted (0.33s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/javascript (0.13s)
    --- PASS: TestCapturedUnionAgrees/by_value.a (2.80s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/native (0.43s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/sanitized (1.09s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/counted (0.50s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/javascript (0.16s)
    --- PASS: TestCapturedUnionAgrees/object.a (3.30s)
        --- PASS: TestCapturedUnionAgrees/object.a/native (0.49s)
        --- PASS: TestCapturedUnionAgrees/object.a/sanitized (1.36s)
        --- PASS: TestCapturedUnionAgrees/object.a/counted (0.74s)
        --- PASS: TestCapturedUnionAgrees/object.a/javascript (0.13s)
    --- PASS: TestCapturedUnionAgrees/object.ts (3.46s)
        --- PASS: TestCapturedUnionAgrees/object.ts/native (0.75s)
        --- PASS: TestCapturedUnionAgrees/object.ts/sanitized (1.37s)
        --- PASS: TestCapturedUnionAgrees/object.ts/counted (0.68s)
        --- PASS: TestCapturedUnionAgrees/object.ts/javascript (0.19s)
    --- PASS: TestCapturedUnionAgrees/class.a (2.79s)
        --- PASS: TestCapturedUnionAgrees/class.a/native (0.46s)
        --- PASS: TestCapturedUnionAgrees/class.a/sanitized (1.08s)
        --- PASS: TestCapturedUnionAgrees/class.a/counted (0.60s)
        --- PASS: TestCapturedUnionAgrees/class.a/javascript (0.27s)
    --- PASS: TestCapturedUnionAgrees/graph.ts (2.83s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/native (0.42s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/sanitized (1.29s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/counted (0.62s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/javascript (0.11s)
    --- PASS: TestCapturedUnionAgrees/alternating.ts (3.02s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/native (0.50s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/sanitized (1.10s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/counted (0.73s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/javascript (0.22s)
    --- PASS: TestCapturedUnionAgrees/alternating.a (2.54s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/native (0.65s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/sanitized (0.83s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/counted (0.60s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/javascript (0.14s)
    --- PASS: TestCapturedUnionAgrees/nested.a (2.61s)
        --- PASS: TestCapturedUnionAgrees/nested.a/native (0.47s)
        --- PASS: TestCapturedUnionAgrees/nested.a/sanitized (0.94s)
        --- PASS: TestCapturedUnionAgrees/nested.a/counted (0.38s)
        --- PASS: TestCapturedUnionAgrees/nested.a/javascript (0.09s)
    --- PASS: TestCapturedUnionAgrees/nested.ts (1.67s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/native (0.36s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/sanitized (0.60s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/counted (0.25s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/javascript (0.10s)
PASS
gate cache: native hits=0 misses=32
gate cache: node hits=0 misses=32
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	587.864s
```


## Step 2c: emitted captured-cell stores

Built with `go build -o /tmp/n7proof/adamic-head ./cmd/adamic`, then ran `/tmp/n7proof/adamic-head c internal/oracle/testdata/captured_union/<pin>.ts` for each pin below. All exit 0. Every reference-load-to-temporary store block is quoted with the next three emitted lines, including other captured cells in the fixtures for completeness. These five fixtures use graph hold/drop for their Union cells; incoming owned values are also released. There is no plain-store Union cell in these five fixtures.

```text
build-head exit=0 wall=29
```


### class.ts

```text
c-class exit=0 wall=0
```


```c
// emitted C line 60
	void *adamic_temporary_body__980fb021557929b016b68b5d799ee48f_1 = adamic_local_70979869531032296000523767796741462618_active_cell->value.reference;
	adamic_local_70979869531032296000523767796741462618_active_cell->value.reference = adamic_graph_hold(adamic_local_70979869531032296000523767796741462618_active_cell, ((adamic_heap *)(NULL)));
	adamic_graph_drop(adamic_local_70979869531032296000523767796741462618_active_cell, adamic_temporary_body__980fb021557929b016b68b5d799ee48f_1);
	adamic_closure * adamic_temporary_body__980fb021557929b016b68b5d799ee48f_2 = adamic_closure_new(adamic_function_class_ts_profiler_see_anonymous_1_bb59c8dbb4492099a108dec52fa3b8f1, 1);

// emitted C line 69
	void *adamic_temporary_body__980fb021557929b016b68b5d799ee48f_3 = adamic_local_163040444718802598107488890449795746608_see_cell->value.reference;
	adamic_local_163040444718802598107488890449795746608_see_cell->value.reference = adamic_graph_hold(adamic_local_163040444718802598107488890449795746608_see_cell, adamic_temporary_body__980fb021557929b016b68b5d799ee48f_2);
	adamic_graph_drop(adamic_local_163040444718802598107488890449795746608_see_cell, adamic_temporary_body__980fb021557929b016b68b5d799ee48f_3);
	adamic_release(adamic_temporary_body__980fb021557929b016b68b5d799ee48f_2);

// emitted C line 380
			void *adamic_temporary_body__334690135d9f8836a0e5f22a757f7348_2 = self->cells[0]->value.reference;
			self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(adamic_temporary_body__334690135d9f8836a0e5f22a757f7348_1)));
			adamic_graph_drop(self->cells[0], adamic_temporary_body__334690135d9f8836a0e5f22a757f7348_2);
			adamic_release(adamic_temporary_body__334690135d9f8836a0e5f22a757f7348_1);

// emitted C line 388
				void *adamic_temporary_body__334690135d9f8836a0e5f22a757f7348_3 = self->cells[0]->value.reference;
				self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(&adamic_string_literal_9ea90488b5c8219acc0b2eb36f115242)));
				adamic_graph_drop(self->cells[0], adamic_temporary_body__334690135d9f8836a0e5f22a757f7348_3);
			}

// emitted C line 394
				void *adamic_temporary_body__334690135d9f8836a0e5f22a757f7348_4 = self->cells[0]->value.reference;
				self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(NULL)));
				adamic_graph_drop(self->cells[0], adamic_temporary_body__334690135d9f8836a0e5f22a757f7348_4);
			}

// emitted C line 539
	void *adamic_temporary_body__8059ae9409cee4b07ec736f121ec6c56_13 = self->cells[0]->value.reference;
	self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(adamic_temporary_body__8059ae9409cee4b07ec736f121ec6c56_12)));
	adamic_graph_drop(self->cells[0], adamic_temporary_body__8059ae9409cee4b07ec736f121ec6c56_13);
	adamic_release(adamic_temporary_body__8059ae9409cee4b07ec736f121ec6c56_12);

// emitted C line 587
			void *adamic_temporary_body__fd5d883fe39918f66c29a4dc247817b0_12 = self->cells[0]->value.reference;
			self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(&adamic_string_literal_9ea90488b5c8219acc0b2eb36f115242)));
			adamic_graph_drop(self->cells[0], adamic_temporary_body__fd5d883fe39918f66c29a4dc247817b0_12);
			adamic_closure * adamic_temporary_body__fd5d883fe39918f66c29a4dc247817b0_13 = adamic_retain((adamic_closure *)self->cells[1]->value.reference);

// emitted C line 600
			void *adamic_temporary_body__fd5d883fe39918f66c29a4dc247817b0_16 = self->cells[0]->value.reference;
			self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(NULL)));
			adamic_graph_drop(self->cells[0], adamic_temporary_body__fd5d883fe39918f66c29a4dc247817b0_16);
			adamic_string * adamic_temporary_body__fd5d883fe39918f66c29a4dc247817b0_17 = ((adamic_string *)(&ADAMIC_UNIFORM_OBJECT(adamic_local_243637366819159875032982188819332018862_held, "name")->slots[0])->reference);
```


### object.ts

```text
c-object exit=0 wall=0
```


```c
// emitted C line 47
	void *adamic_temporary_body__5b9093c2cdab8b3402885a20c351c148_1 = adamic_local_10762166693295184825812035896398178931_active_cell->value.reference;
	adamic_local_10762166693295184825812035896398178931_active_cell->value.reference = adamic_graph_hold(adamic_local_10762166693295184825812035896398178931_active_cell, ((adamic_heap *)(NULL)));
	adamic_graph_drop(adamic_local_10762166693295184825812035896398178931_active_cell, adamic_temporary_body__5b9093c2cdab8b3402885a20c351c148_1);
	adamic_closure * adamic_temporary_body__5b9093c2cdab8b3402885a20c351c148_2 = adamic_closure_new(adamic_function_object_ts_profiler_see_anonymous_1_5e2ac398842c7663a0903f41f6c9b7dd, 1);

// emitted C line 56
	void *adamic_temporary_body__5b9093c2cdab8b3402885a20c351c148_3 = adamic_local_126562659516442909445654413960754246739_see_cell->value.reference;
	adamic_local_126562659516442909445654413960754246739_see_cell->value.reference = adamic_graph_hold(adamic_local_126562659516442909445654413960754246739_see_cell, adamic_temporary_body__5b9093c2cdab8b3402885a20c351c148_2);
	adamic_graph_drop(adamic_local_126562659516442909445654413960754246739_see_cell, adamic_temporary_body__5b9093c2cdab8b3402885a20c351c148_3);
	adamic_release(adamic_temporary_body__5b9093c2cdab8b3402885a20c351c148_2);

// emitted C line 368
			void *adamic_temporary_body__aadebafd4a0d616e7929dd737f874ed0_2 = self->cells[0]->value.reference;
			self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(adamic_temporary_body__aadebafd4a0d616e7929dd737f874ed0_1)));
			adamic_graph_drop(self->cells[0], adamic_temporary_body__aadebafd4a0d616e7929dd737f874ed0_2);
			adamic_release(adamic_temporary_body__aadebafd4a0d616e7929dd737f874ed0_1);

// emitted C line 376
				void *adamic_temporary_body__aadebafd4a0d616e7929dd737f874ed0_3 = self->cells[0]->value.reference;
				self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(&adamic_string_literal_9ea90488b5c8219acc0b2eb36f115242)));
				adamic_graph_drop(self->cells[0], adamic_temporary_body__aadebafd4a0d616e7929dd737f874ed0_3);
			}

// emitted C line 382
				void *adamic_temporary_body__aadebafd4a0d616e7929dd737f874ed0_4 = self->cells[0]->value.reference;
				self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(NULL)));
				adamic_graph_drop(self->cells[0], adamic_temporary_body__aadebafd4a0d616e7929dd737f874ed0_4);
			}

// emitted C line 461
	void *adamic_temporary_body__123d9bff77d6d7ccb28090675f48b047_13 = self->cells[0]->value.reference;
	self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(adamic_temporary_body__123d9bff77d6d7ccb28090675f48b047_12)));
	adamic_graph_drop(self->cells[0], adamic_temporary_body__123d9bff77d6d7ccb28090675f48b047_13);
	adamic_release(adamic_temporary_body__123d9bff77d6d7ccb28090675f48b047_12);

// emitted C line 509
			void *adamic_temporary_body__4e9e87c1d214789799f801041d32096e_12 = self->cells[0]->value.reference;
			self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(&adamic_string_literal_9ea90488b5c8219acc0b2eb36f115242)));
			adamic_graph_drop(self->cells[0], adamic_temporary_body__4e9e87c1d214789799f801041d32096e_12);
			adamic_closure * adamic_temporary_body__4e9e87c1d214789799f801041d32096e_13 = adamic_retain((adamic_closure *)self->cells[1]->value.reference);

// emitted C line 522
			void *adamic_temporary_body__4e9e87c1d214789799f801041d32096e_16 = self->cells[0]->value.reference;
			self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(NULL)));
			adamic_graph_drop(self->cells[0], adamic_temporary_body__4e9e87c1d214789799f801041d32096e_16);
			adamic_string * adamic_temporary_body__4e9e87c1d214789799f801041d32096e_17 = ((adamic_string *)(&ADAMIC_UNIFORM_OBJECT(adamic_local_181289819388308358458421959642363494959_held, "name")->slots[0])->reference);
```


### graph.ts

```text
c-graph exit=0 wall=0
```


```c
// emitted C line 61
	void *adamic_temporary_body__c464e229e29eb0e94a6a089d2f97cbb7_1 = adamic_local_340005101388834193637987220327987819882_active_cell->value.reference;
	adamic_local_340005101388834193637987220327987819882_active_cell->value.reference = adamic_graph_hold(adamic_local_340005101388834193637987220327987819882_active_cell, ((adamic_heap *)(NULL)));
	adamic_graph_drop(adamic_local_340005101388834193637987220327987819882_active_cell, adamic_temporary_body__c464e229e29eb0e94a6a089d2f97cbb7_1);
	adamic_cell *adamic_local_46770150113887620091200810217711929278_holder_cell = adamic_cell_new((adamic_value){.number = 0}, true);

// emitted C line 66
	void *adamic_temporary_body__c464e229e29eb0e94a6a089d2f97cbb7_2 = adamic_local_46770150113887620091200810217711929278_holder_cell->value.reference;
	adamic_local_46770150113887620091200810217711929278_holder_cell->value.reference = adamic_graph_hold(adamic_local_46770150113887620091200810217711929278_holder_cell, NULL);
	adamic_graph_drop(adamic_local_46770150113887620091200810217711929278_holder_cell, adamic_temporary_body__c464e229e29eb0e94a6a089d2f97cbb7_2);
	adamic_closure * adamic_temporary_body__c464e229e29eb0e94a6a089d2f97cbb7_3 = adamic_closure_new(adamic_function_graph_ts_run_controls_write_f53f58b166562adca36aabea00dd916d, 1);

// emitted C line 88
	void *adamic_temporary_body__c464e229e29eb0e94a6a089d2f97cbb7_6 = adamic_local_46770150113887620091200810217711929278_holder_cell->value.reference;
	adamic_local_46770150113887620091200810217711929278_holder_cell->value.reference = adamic_graph_hold(adamic_local_46770150113887620091200810217711929278_holder_cell, adamic_local_262830241147927147583861215273874901821_controls);
	adamic_graph_drop(adamic_local_46770150113887620091200810217711929278_holder_cell, adamic_temporary_body__c464e229e29eb0e94a6a089d2f97cbb7_6);
	adamic_method adamic_temporary_body__c464e229e29eb0e94a6a089d2f97cbb7_7 = NULL;

// emitted C line 224
			void *adamic_temporary_body__f02af8056dd6c777a6ed911f4a5921aa_2 = self->cells[0]->value.reference;
			self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(adamic_temporary_body__f02af8056dd6c777a6ed911f4a5921aa_1)));
			adamic_graph_drop(self->cells[0], adamic_temporary_body__f02af8056dd6c777a6ed911f4a5921aa_2);
			adamic_release(adamic_temporary_body__f02af8056dd6c777a6ed911f4a5921aa_1);

// emitted C line 232
				void *adamic_temporary_body__f02af8056dd6c777a6ed911f4a5921aa_3 = self->cells[0]->value.reference;
				self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(&adamic_string_literal_9ea90488b5c8219acc0b2eb36f115242)));
				adamic_graph_drop(self->cells[0], adamic_temporary_body__f02af8056dd6c777a6ed911f4a5921aa_3);
			}

// emitted C line 239
					void *adamic_temporary_body__f02af8056dd6c777a6ed911f4a5921aa_4 = self->cells[0]->value.reference;
					self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(NULL)));
					adamic_graph_drop(self->cells[0], adamic_temporary_body__f02af8056dd6c777a6ed911f4a5921aa_4);
				}

// emitted C line 254
					void *adamic_temporary_body__f02af8056dd6c777a6ed911f4a5921aa_8 = self->cells[0]->value.reference;
					self->cells[0]->value.reference = adamic_graph_hold(self->cells[0], ((adamic_heap *)(adamic_temporary_body__f02af8056dd6c777a6ed911f4a5921aa_7)));
					adamic_graph_drop(self->cells[0], adamic_temporary_body__f02af8056dd6c777a6ed911f4a5921aa_8);
					adamic_release(adamic_temporary_body__f02af8056dd6c777a6ed911f4a5921aa_7);
```


### cycle.ts

```text
c-cycle exit=0 wall=0
```


```c
// emitted C line 49
	void *adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_1 = adamic_local_187927938248785647813563295002714990240_active_cell->value.reference;
	adamic_local_187927938248785647813563295002714990240_active_cell->value.reference = adamic_graph_hold(adamic_local_187927938248785647813563295002714990240_active_cell, ((adamic_heap *)(NULL)));
	adamic_graph_drop(adamic_local_187927938248785647813563295002714990240_active_cell, adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_1);
	adamic_closure * adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_2 = adamic_closure_new(adamic_function_cycle_ts_run_report_anonymous_1_5fb10d01f3528426c81e47781ffc41f0, 1);

// emitted C line 79
	void *adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_9 = adamic_local_187927938248785647813563295002714990240_active_cell->value.reference;
	adamic_local_187927938248785647813563295002714990240_active_cell->value.reference = adamic_graph_hold(adamic_local_187927938248785647813563295002714990240_active_cell, ((adamic_heap *)(adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_8)));
	adamic_graph_drop(adamic_local_187927938248785647813563295002714990240_active_cell, adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_9);
	adamic_release(adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_8);

// emitted C line 119
	void *adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_18 = adamic_local_187927938248785647813563295002714990240_active_cell->value.reference;
	adamic_local_187927938248785647813563295002714990240_active_cell->value.reference = adamic_graph_hold(adamic_local_187927938248785647813563295002714990240_active_cell, ((adamic_heap *)(&adamic_string_literal_9ea90488b5c8219acc0b2eb36f115242)));
	adamic_graph_drop(adamic_local_187927938248785647813563295002714990240_active_cell, adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_18);
	adamic_value adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_19 = adamic_local_63061723651032981089421994811472650453_report->code(adamic_local_63061723651032981089421994811472650453_report, NULL);

// emitted C line 145
	void *adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_25 = adamic_local_187927938248785647813563295002714990240_active_cell->value.reference;
	adamic_local_187927938248785647813563295002714990240_active_cell->value.reference = adamic_graph_hold(adamic_local_187927938248785647813563295002714990240_active_cell, ((adamic_heap *)(adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_24)));
	adamic_graph_drop(adamic_local_187927938248785647813563295002714990240_active_cell, adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_25);
	adamic_release(adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_24);

// emitted C line 185
	void *adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_34 = adamic_local_187927938248785647813563295002714990240_active_cell->value.reference;
	adamic_local_187927938248785647813563295002714990240_active_cell->value.reference = adamic_graph_hold(adamic_local_187927938248785647813563295002714990240_active_cell, ((adamic_heap *)(NULL)));
	adamic_graph_drop(adamic_local_187927938248785647813563295002714990240_active_cell, adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_34);
	adamic_value adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_35 = adamic_local_63061723651032981089421994811472650453_report->code(adamic_local_63061723651032981089421994811472650453_report, NULL);

// emitted C line 211
	void *adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_41 = adamic_local_187927938248785647813563295002714990240_active_cell->value.reference;
	adamic_local_187927938248785647813563295002714990240_active_cell->value.reference = adamic_graph_hold(adamic_local_187927938248785647813563295002714990240_active_cell, ((adamic_heap *)(adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_40)));
	adamic_graph_drop(adamic_local_187927938248785647813563295002714990240_active_cell, adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_41);
	adamic_release(adamic_temporary_body__496ad70f1c31faf508c809d05fa5ac29_40);
```


### nested.ts

```text
c-nested exit=0 wall=1
```


```c
// emitted C line 40
	void *adamic_temporary_body__a4ab7ae01bb867d7b69bb420553fd611_1 = adamic_local_110087480204372633436063825260439818806_extended_cell->value.reference;
	adamic_local_110087480204372633436063825260439818806_extended_cell->value.reference = adamic_graph_hold(adamic_local_110087480204372633436063825260439818806_extended_cell, ((adamic_heap *)(NULL)));
	adamic_graph_drop(adamic_local_110087480204372633436063825260439818806_extended_cell, adamic_temporary_body__a4ab7ae01bb867d7b69bb420553fd611_1);
	adamic_closure * adamic_temporary_body__a4ab7ae01bb867d7b69bb420553fd611_2 = adamic_closure_new(adamic_function_nested_ts_parse_onPropertySet_anonymous_1_7744a35bb04b4244c60dae235020891d, 1);

// emitted C line 49
	void *adamic_temporary_body__a4ab7ae01bb867d7b69bb420553fd611_3 = adamic_local_45235049387504643564994886099546440834_onPropertySet_cell->value.reference;
	adamic_local_45235049387504643564994886099546440834_onPropertySet_cell->value.reference = adamic_graph_hold(adamic_local_45235049387504643564994886099546440834_onPropertySet_cell, adamic_temporary_body__a4ab7ae01bb867d7b69bb420553fd611_2);
	adamic_graph_drop(adamic_local_45235049387504643564994886099546440834_onPropertySet_cell, adamic_temporary_body__a4ab7ae01bb867d7b69bb420553fd611_3);
	adamic_release(adamic_temporary_body__a4ab7ae01bb867d7b69bb420553fd611_2);

// emitted C line 173
	void *adamic_temporary_body__74ef40d949bf33c4810fdac8802f0431_1 = adamic_local_102552935997465435882939107326857724488_key_cell->value.reference;
	adamic_local_102552935997465435882939107326857724488_key_cell->value.reference = adamic_graph_hold(adamic_local_102552935997465435882939107326857724488_key_cell, adamic_local_102552935997465435882939107326857724488_key);
	adamic_graph_drop(adamic_local_102552935997465435882939107326857724488_key_cell, adamic_temporary_body__74ef40d949bf33c4810fdac8802f0431_1);
	adamic_cell *adamic_local_5550805993488708397420551089415515465_value_cell = adamic_cell_new((adamic_value){.number = 0}, true);

// emitted C line 178
	void *adamic_temporary_body__74ef40d949bf33c4810fdac8802f0431_2 = adamic_local_5550805993488708397420551089415515465_value_cell->value.reference;
	adamic_local_5550805993488708397420551089415515465_value_cell->value.reference = adamic_graph_hold(adamic_local_5550805993488708397420551089415515465_value_cell, adamic_local_5550805993488708397420551089415515465_value);
	adamic_graph_drop(adamic_local_5550805993488708397420551089415515465_value_cell, adamic_temporary_body__74ef40d949bf33c4810fdac8802f0431_2);
	adamic_closure * adamic_temporary_body__74ef40d949bf33c4810fdac8802f0431_3 = adamic_closure_new(adamic_function_nested_ts_parse_onPropertySet_anonymous_1_assign_anonymous_1_fc410c0e63900fe30d96e4ff3787a959, 3);

// emitted C line 210
			void *adamic_temporary_body__dcc59e156a4fbf4ea321d352b66b6ea1_4 = self->cells[1]->value.reference;
			self->cells[1]->value.reference = adamic_graph_hold(self->cells[1], ((adamic_heap *)(adamic_temporary_body__dcc59e156a4fbf4ea321d352b66b6ea1_3)));
			adamic_graph_drop(self->cells[1], adamic_temporary_body__dcc59e156a4fbf4ea321d352b66b6ea1_4);
			adamic_release(adamic_temporary_body__dcc59e156a4fbf4ea321d352b66b6ea1_3);

// emitted C line 222
				void *adamic_temporary_body__dcc59e156a4fbf4ea321d352b66b6ea1_9 = self->cells[1]->value.reference;
				self->cells[1]->value.reference = adamic_graph_hold(self->cells[1], ((adamic_heap *)(adamic_temporary_body__dcc59e156a4fbf4ea321d352b66b6ea1_8)));
				adamic_graph_drop(self->cells[1], adamic_temporary_body__dcc59e156a4fbf4ea321d352b66b6ea1_9);
				adamic_release(adamic_temporary_body__dcc59e156a4fbf4ea321d352b66b6ea1_8);

// emitted C line 232
					void *adamic_temporary_body__dcc59e156a4fbf4ea321d352b66b6ea1_12 = self->cells[1]->value.reference;
					self->cells[1]->value.reference = adamic_graph_hold(self->cells[1], ((adamic_heap *)(NULL)));
					adamic_graph_drop(self->cells[1], adamic_temporary_body__dcc59e156a4fbf4ea321d352b66b6ea1_12);
				}
```


## Step 3: isolated mutants

Each old-text match was asserted to occur exactly once before applying it. After each mutant, its exact original file bytes were restored and `git diff --exit-code` returned 0 with no output. M1/M2 changed internal/lower/locals.go; M3/M4 changed internal/native/emit_locals.go. Both step-2 commands above were rerun for each mutant; oracle runs retained ADAMIC_GATE_UNCACHED=1.

### M1: The guard restored

```text
M1 old text occurs exactly once
M1-a exit=1 wall=47
M1-b exit=1 wall=60
restore: git diff --exit-code = 0; output empty
```


All 16 Union pin/dialect cases fail during lowering in both suites. No backend subtest is reached: sanitized and counted are **not run**, and no counts line exists. TestCapturedMaybeBooleanStillStops passes.

Every red subtest and its first diagnostic line (parent failures inherit their child failures):

- `TestCapturedUnionLowersToOneCell` (a): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionLowersToOneCell/nested.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.ts3674549461/001/main.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/class.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.ts4008280696/001/main.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/graph.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.ts1762978926/001/main.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/cycle.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.a2219487233/001/main.a:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/cycle.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.ts2996869690/001/main.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/graph.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.a2404657976/001/main.a:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/released.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.a41746553/001/main.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/released.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.ts2812158124/001/main.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/by_value.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.ts3961235673/001/main.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/nested.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.a2703018572/001/main.a:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/by_value.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.a395434953/001/main.a:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/alternating.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.ts1032142941/001/main.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/alternating.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.a620197534/001/main.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/object.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.a1940353551/001/main.a:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/class.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.a2443202770/001/main.a:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionLowersToOneCell/object.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.ts1925304594/001/main.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/class.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.ts1779285515/001/class.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/graph.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.ts1146524001/001/graph.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/by_value.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.ts335739176/001/by_value.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/nested.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.ts3874390252/001/nested.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/object.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.a1598300908/001/object.a:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/released.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.a970898110/001/released.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/by_value.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.a3185133142/001/by_value.a:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/released.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.ts3493482084/001/released.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/object.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.ts4244009490/001/object.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/cycle.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.ts991103518/001/cycle.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/cycle.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.a428476426/001/cycle.a:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/alternating.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.a70069944/001/alternating.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/graph.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.a3434884735/001/graph.a:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/alternating.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.ts3724297979/001/alternating.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/class.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.a2705439248/001/class.a:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet

- `TestCapturedUnionAgrees/nested.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.a1711318900/001/nested.a:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet



Complete verbatim logs retain every first error and, for non-lowering failures, the whole observed leak/count diagnostic. These tests expect no leaks and balanced counts; their pinned stdout expectations are unchanged. No stdout disagreement was reported.

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== CONT  TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== CONT  TestCapturedMaybeBooleanStillStops
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== RUN   TestCapturedUnionLowersToOneCell/cycle.ts
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.ts
=== RUN   TestCapturedUnionLowersToOneCell/cycle.a
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
=== CONT  TestCapturedUnionLowersToOneCell/cycle.a
--- PASS: TestCapturedMaybeBooleanStillStops (0.37s)
=== CONT  TestCapturedUnionLowersToOneCell/cycle.ts
=== NAME  TestCapturedUnionLowersToOneCell/nested.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.ts3674549461/001/main.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
=== NAME  TestCapturedUnionLowersToOneCell/class.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.ts4008280696/001/main.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
=== NAME  TestCapturedUnionLowersToOneCell/graph.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.ts1762978926/001/main.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== NAME  TestCapturedUnionLowersToOneCell/cycle.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.a2219487233/001/main.a:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== NAME  TestCapturedUnionLowersToOneCell/cycle.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.ts2996869690/001/main.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
=== NAME  TestCapturedUnionLowersToOneCell/graph.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.a2404657976/001/main.a:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
=== NAME  TestCapturedUnionLowersToOneCell/released.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.a41746553/001/main.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
=== NAME  TestCapturedUnionLowersToOneCell/released.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.ts2812158124/001/main.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== NAME  TestCapturedUnionLowersToOneCell/by_value.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.ts3961235673/001/main.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
=== NAME  TestCapturedUnionLowersToOneCell/nested.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.a2703018572/001/main.a:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/object.a
=== NAME  TestCapturedUnionLowersToOneCell/by_value.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.a395434953/001/main.a:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== NAME  TestCapturedUnionLowersToOneCell/alternating.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.ts1032142941/001/main.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionLowersToOneCell/alternating.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.a620197534/001/main.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionLowersToOneCell/object.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.a1940353551/001/main.a:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionLowersToOneCell/class.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.a2443202770/001/main.a:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionLowersToOneCell/object.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.ts1925304594/001/main.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
--- FAIL: TestCapturedUnionLowersToOneCell (0.00s)
    --- FAIL: TestCapturedUnionLowersToOneCell/nested.ts (0.57s)
    --- FAIL: TestCapturedUnionLowersToOneCell/class.ts (0.57s)
    --- FAIL: TestCapturedUnionLowersToOneCell/graph.ts (0.59s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.a (0.61s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.ts (0.44s)
    --- FAIL: TestCapturedUnionLowersToOneCell/graph.a (0.39s)
    --- FAIL: TestCapturedUnionLowersToOneCell/released.a (0.61s)
    --- FAIL: TestCapturedUnionLowersToOneCell/released.ts (0.63s)
    --- FAIL: TestCapturedUnionLowersToOneCell/by_value.ts (0.68s)
    --- FAIL: TestCapturedUnionLowersToOneCell/nested.a (0.14s)
    --- FAIL: TestCapturedUnionLowersToOneCell/by_value.a (0.58s)
    --- FAIL: TestCapturedUnionLowersToOneCell/alternating.ts (0.47s)
    --- FAIL: TestCapturedUnionLowersToOneCell/alternating.a (0.29s)
    --- FAIL: TestCapturedUnionLowersToOneCell/object.a (0.26s)
    --- FAIL: TestCapturedUnionLowersToOneCell/class.a (0.26s)
    --- FAIL: TestCapturedUnionLowersToOneCell/object.ts (0.42s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	1.738s
FAIL
```


```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== CONT  TestCapturedUnionAgrees
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/cycle.ts
=== PAUSE TestCapturedUnionAgrees/cycle.ts
=== RUN   TestCapturedUnionAgrees/cycle.a
=== PAUSE TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/nested.ts
=== CONT  TestCapturedUnionAgrees/graph.ts
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== CONT  TestCapturedUnionAgrees/object.a
=== NAME  TestCapturedUnionAgrees/class.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.ts1779285515/001/class.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/released.a
=== NAME  TestCapturedUnionAgrees/graph.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.ts1146524001/001/graph.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/released.ts
=== NAME  TestCapturedUnionAgrees/by_value.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.ts335739176/001/by_value.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/by_value.a
=== NAME  TestCapturedUnionAgrees/nested.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.ts3874390252/001/nested.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/cycle.ts
=== NAME  TestCapturedUnionAgrees/object.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.a1598300908/001/object.a:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/cycle.a
=== NAME  TestCapturedUnionAgrees/released.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.a970898110/001/released.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/object.ts
=== NAME  TestCapturedUnionAgrees/by_value.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.a3185133142/001/by_value.a:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== NAME  TestCapturedUnionAgrees/released.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.ts3493482084/001/released.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/alternating.a
=== NAME  TestCapturedUnionAgrees/object.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.ts4244009490/001/object.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/graph.a
=== NAME  TestCapturedUnionAgrees/cycle.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.ts991103518/001/cycle.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/class.a
=== NAME  TestCapturedUnionAgrees/cycle.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.a428476426/001/cycle.a:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/nested.a
=== NAME  TestCapturedUnionAgrees/alternating.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.a70069944/001/alternating.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/graph.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.a3434884735/001/graph.a:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/alternating.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.ts3724297979/001/alternating.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/class.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.a2705439248/001/class.a:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/nested.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.a1711318900/001/nested.a:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
--- FAIL: TestCapturedUnionAgrees (0.00s)
    --- FAIL: TestCapturedUnionAgrees/class.ts (0.61s)
    --- FAIL: TestCapturedUnionAgrees/graph.ts (0.66s)
    --- FAIL: TestCapturedUnionAgrees/by_value.ts (0.67s)
    --- FAIL: TestCapturedUnionAgrees/nested.ts (0.69s)
    --- FAIL: TestCapturedUnionAgrees/object.a (0.76s)
    --- FAIL: TestCapturedUnionAgrees/released.a (0.26s)
    --- FAIL: TestCapturedUnionAgrees/by_value.a (0.34s)
    --- FAIL: TestCapturedUnionAgrees/released.ts (0.36s)
    --- FAIL: TestCapturedUnionAgrees/object.ts (0.23s)
    --- FAIL: TestCapturedUnionAgrees/cycle.ts (0.47s)
    --- FAIL: TestCapturedUnionAgrees/cycle.a (0.43s)
    --- FAIL: TestCapturedUnionAgrees/alternating.a (0.26s)
    --- FAIL: TestCapturedUnionAgrees/graph.a (0.20s)
    --- FAIL: TestCapturedUnionAgrees/alternating.ts (0.32s)
    --- FAIL: TestCapturedUnionAgrees/class.a (0.17s)
    --- FAIL: TestCapturedUnionAgrees/nested.a (0.16s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	1.393s
FAIL
```


### M2: The cell held as a plain reference

```text
M2 old text occurs exactly once
M2-a exit=1 wall=66
M2-b exit=1 wall=61
restore: git diff --exit-code = 0; output empty
```


All 16 Union pin/dialect cases fail during lowering in both suites. No backend subtest is reached: sanitized and counted are **not run**, and no counts line exists. TestCapturedMaybeBooleanStillStops passes.

Every red subtest and its first diagnostic line (parent failures inherit their child failures):

- `TestCapturedUnionLowersToOneCell` (a): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionLowersToOneCell/nested.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.ts3731655023/001/main.ts:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet

- `TestCapturedUnionLowersToOneCell/graph.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.ts4136597464/001/main.ts:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionLowersToOneCell/class.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.ts2328549592/001/main.ts:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionLowersToOneCell/by_value.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.ts2690487410/001/main.ts:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionLowersToOneCell/released.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.a4096265858/001/main.a:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionLowersToOneCell/released.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.ts1252887316/001/main.ts:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionLowersToOneCell/object.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.a1577060680/001/main.a:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionLowersToOneCell/alternating.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.ts1202949578/001/main.ts:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionLowersToOneCell/by_value.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.a3439589282/001/main.a:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionLowersToOneCell/object.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.ts3158015702/001/main.ts:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionLowersToOneCell/class.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.a2880403916/001/main.a:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionLowersToOneCell/alternating.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.a3748449423/001/main.a:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionLowersToOneCell/graph.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.a3881575516/001/main.a:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionLowersToOneCell/cycle.ts` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.ts730917601/001/main.ts:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionLowersToOneCell/cycle.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.a185077768/001/main.a:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionLowersToOneCell/nested.a` (a): captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.a3126333676/001/main.a:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet

- `TestCapturedUnionAgrees` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/by_value.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.ts3981086676/001/by_value.ts:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionAgrees/object.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.ts1931391235/001/object.ts:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionAgrees/nested.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.ts2102006538/001/nested.ts:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet

- `TestCapturedUnionAgrees/released.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.ts2345513928/001/released.ts:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionAgrees/class.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.ts718217399/001/class.ts:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionAgrees/cycle.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.ts1702155502/001/cycle.ts:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionAgrees/graph.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.a2117519174/001/graph.a:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionAgrees/cycle.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.a3422581501/001/cycle.a:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionAgrees/alternating.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.a3542789937/001/alternating.a:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionAgrees/graph.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.ts2648604565/001/graph.ts:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionAgrees/nested.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.a1828490831/001/nested.a:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet

- `TestCapturedUnionAgrees/class.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.a3223198771/001/class.a:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionAgrees/by_value.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.a91948282/001/by_value.a:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionAgrees/object.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.a2326650267/001/object.a:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionAgrees/released.a` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.a789452185/001/released.a:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet

- `TestCapturedUnionAgrees/alternating.ts` (b): captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.ts2143845773/001/alternating.ts:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet



Complete verbatim logs retain every first error and, for non-lowering failures, the whole observed leak/count diagnostic. These tests expect no leaks and balanced counts; their pinned stdout expectations are unchanged. No stdout disagreement was reported.

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== CONT  TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== RUN   TestCapturedUnionLowersToOneCell/cycle.ts
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.ts
=== RUN   TestCapturedUnionLowersToOneCell/cycle.a
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedMaybeBooleanStillStops
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
=== NAME  TestCapturedUnionLowersToOneCell/nested.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.ts3731655023/001/main.ts:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== NAME  TestCapturedUnionLowersToOneCell/graph.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.ts4136597464/001/main.ts:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== NAME  TestCapturedUnionLowersToOneCell/class.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.ts2328549592/001/main.ts:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
=== NAME  TestCapturedUnionLowersToOneCell/by_value.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.ts2690487410/001/main.ts:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
--- PASS: TestCapturedMaybeBooleanStillStops (0.74s)
=== CONT  TestCapturedUnionLowersToOneCell/object.a
=== NAME  TestCapturedUnionLowersToOneCell/released.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.a4096265858/001/main.a:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
=== NAME  TestCapturedUnionLowersToOneCell/released.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.ts1252887316/001/main.ts:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
=== NAME  TestCapturedUnionLowersToOneCell/object.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.a1577060680/001/main.a:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== NAME  TestCapturedUnionLowersToOneCell/alternating.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.ts1202949578/001/main.ts:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== NAME  TestCapturedUnionLowersToOneCell/by_value.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.a3439589282/001/main.a:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/cycle.ts
=== NAME  TestCapturedUnionLowersToOneCell/object.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.ts3158015702/001/main.ts:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/cycle.a
=== NAME  TestCapturedUnionLowersToOneCell/class.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.a2880403916/001/main.a:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
=== NAME  TestCapturedUnionLowersToOneCell/alternating.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.a3748449423/001/main.a:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionLowersToOneCell/graph.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.a3881575516/001/main.a:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionLowersToOneCell/cycle.ts
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.ts730917601/001/main.ts:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionLowersToOneCell/cycle.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.a185077768/001/main.a:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionLowersToOneCell/nested.a
    captured_union_test.go:45: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.a3126333676/001/main.a:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet
--- FAIL: TestCapturedUnionLowersToOneCell (0.00s)
    --- FAIL: TestCapturedUnionLowersToOneCell/nested.ts (0.48s)
    --- FAIL: TestCapturedUnionLowersToOneCell/graph.ts (0.52s)
    --- FAIL: TestCapturedUnionLowersToOneCell/class.ts (0.52s)
    --- FAIL: TestCapturedUnionLowersToOneCell/by_value.ts (0.74s)
    --- FAIL: TestCapturedUnionLowersToOneCell/released.a (0.44s)
    --- FAIL: TestCapturedUnionLowersToOneCell/released.ts (0.77s)
    --- FAIL: TestCapturedUnionLowersToOneCell/object.a (0.55s)
    --- FAIL: TestCapturedUnionLowersToOneCell/alternating.ts (0.38s)
    --- FAIL: TestCapturedUnionLowersToOneCell/by_value.a (0.88s)
    --- FAIL: TestCapturedUnionLowersToOneCell/object.ts (0.69s)
    --- FAIL: TestCapturedUnionLowersToOneCell/class.a (0.24s)
    --- FAIL: TestCapturedUnionLowersToOneCell/alternating.a (0.36s)
    --- FAIL: TestCapturedUnionLowersToOneCell/graph.a (0.14s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.ts (0.34s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.a (0.32s)
    --- FAIL: TestCapturedUnionLowersToOneCell/nested.a (0.46s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	1.839s
FAIL
```


```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== CONT  TestCapturedUnionAgrees
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/cycle.ts
=== PAUSE TestCapturedUnionAgrees/cycle.ts
=== RUN   TestCapturedUnionAgrees/cycle.a
=== PAUSE TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/nested.ts
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== CONT  TestCapturedUnionAgrees/released.ts
=== CONT  TestCapturedUnionAgrees/object.ts
=== NAME  TestCapturedUnionAgrees/by_value.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.ts3981086676/001/by_value.ts:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/cycle.a
=== NAME  TestCapturedUnionAgrees/object.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.ts1931391235/001/object.ts:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/cycle.ts
=== NAME  TestCapturedUnionAgrees/nested.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.ts2102006538/001/nested.ts:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet
=== CONT  TestCapturedUnionAgrees/graph.a
=== NAME  TestCapturedUnionAgrees/released.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.ts2345513928/001/released.ts:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/graph.ts
=== NAME  TestCapturedUnionAgrees/class.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.ts718217399/001/class.ts:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/alternating.a
=== NAME  TestCapturedUnionAgrees/cycle.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.ts1702155502/001/cycle.ts:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== NAME  TestCapturedUnionAgrees/graph.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.a2117519174/001/graph.a:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/nested.a
=== NAME  TestCapturedUnionAgrees/cycle.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.a3422581501/001/cycle.a:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/class.a
=== NAME  TestCapturedUnionAgrees/alternating.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.a3542789937/001/alternating.a:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/by_value.a
=== NAME  TestCapturedUnionAgrees/graph.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.ts2648604565/001/graph.ts:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/released.a
=== NAME  TestCapturedUnionAgrees/nested.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.a1828490831/001/nested.a:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet
=== CONT  TestCapturedUnionAgrees/object.a
=== NAME  TestCapturedUnionAgrees/class.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.a3223198771/001/class.a:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionAgrees/by_value.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.a91948282/001/by_value.a:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionAgrees/object.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.a2326650267/001/object.a:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionAgrees/released.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.a789452185/001/released.a:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionAgrees/alternating.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.ts2143845773/001/alternating.ts:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet
--- FAIL: TestCapturedUnionAgrees (0.00s)
    --- FAIL: TestCapturedUnionAgrees/by_value.ts (0.70s)
    --- FAIL: TestCapturedUnionAgrees/object.ts (0.71s)
    --- FAIL: TestCapturedUnionAgrees/nested.ts (0.76s)
    --- FAIL: TestCapturedUnionAgrees/released.ts (0.77s)
    --- FAIL: TestCapturedUnionAgrees/class.ts (0.89s)
    --- FAIL: TestCapturedUnionAgrees/cycle.ts (0.29s)
    --- FAIL: TestCapturedUnionAgrees/graph.a (0.24s)
    --- FAIL: TestCapturedUnionAgrees/cycle.a (0.30s)
    --- FAIL: TestCapturedUnionAgrees/alternating.a (0.22s)
    --- FAIL: TestCapturedUnionAgrees/graph.ts (0.40s)
    --- FAIL: TestCapturedUnionAgrees/nested.a (0.36s)
    --- FAIL: TestCapturedUnionAgrees/class.a (0.37s)
    --- FAIL: TestCapturedUnionAgrees/by_value.a (0.40s)
    --- FAIL: TestCapturedUnionAgrees/object.a (0.23s)
    --- FAIL: TestCapturedUnionAgrees/released.a (0.49s)
    --- FAIL: TestCapturedUnionAgrees/alternating.ts (0.66s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	1.684s
FAIL
```


### M3: The overwritten value kept

```text
M3 old text occurs exactly once
M3-a exit=0 wall=49
M3-b exit=1 wall=65
restore: git diff --exit-code = 0; output empty
```


All lowering pins and the maybe-boolean guard pass. Per pin and dialect, sanitized and counted results and the exact logged counts line:

| Pin | Sanitized | Counted | Counts line |
| --- | --- | --- | --- |
| class.ts | PASS | PASS | captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0 |
| class.a | PASS | PASS | captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0 |
| object.ts | PASS | PASS | captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0 |
| object.a | PASS | PASS | captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0 |
| by_value.ts | FAIL | FAIL | captured_union_test.go:116: counts: allocations 35, frees 32, in regions 0 |
| by_value.a | FAIL | FAIL | captured_union_test.go:116: counts: allocations 35, frees 32, in regions 0 |
| released.ts | FAIL | FAIL | captured_union_test.go:116: counts: allocations 96, frees 66, in regions 0 |
| released.a | FAIL | FAIL | captured_union_test.go:116: counts: allocations 96, frees 66, in regions 0 |
| nested.ts | PASS | PASS | captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0 |
| nested.a | PASS | PASS | captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0 |
| alternating.ts | FAIL | FAIL | captured_union_test.go:116: counts: allocations 54, frees 36, in regions 0 |
| alternating.a | FAIL | FAIL | captured_union_test.go:116: counts: allocations 54, frees 36, in regions 0 |
| graph.ts | PASS | PASS | captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0 |
| graph.a | PASS | PASS | captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0 |
| cycle.ts | PASS | PASS | captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0 |
| cycle.a | PASS | PASS | captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0 |


Every red subtest and its first diagnostic line (parent failures inherit their child failures):

- `TestCapturedUnionAgrees` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/by_value.ts` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/by_value.ts/sanitized` (b): captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1

- `TestCapturedUnionAgrees/by_value.ts/counted` (b): captured_union_test.go:110: counts: heap values leaked: 3 (allocations 35, frees 32, in regions 0)

- `TestCapturedUnionAgrees/released.a` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/released.a/sanitized` (b): captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1

- `TestCapturedUnionAgrees/released.a/counted` (b): captured_union_test.go:110: counts: heap values leaked: 30 (allocations 96, frees 66, in regions 0)

- `TestCapturedUnionAgrees/released.ts` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/released.ts/sanitized` (b): captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1

- `TestCapturedUnionAgrees/released.ts/counted` (b): captured_union_test.go:110: counts: heap values leaked: 30 (allocations 96, frees 66, in regions 0)

- `TestCapturedUnionAgrees/by_value.a` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/by_value.a/sanitized` (b): captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1

- `TestCapturedUnionAgrees/by_value.a/counted` (b): captured_union_test.go:110: counts: heap values leaked: 3 (allocations 35, frees 32, in regions 0)

- `TestCapturedUnionAgrees/alternating.ts` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/alternating.ts/sanitized` (b): captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1

- `TestCapturedUnionAgrees/alternating.ts/counted` (b): captured_union_test.go:110: counts: heap values leaked: 18 (allocations 54, frees 36, in regions 0)

- `TestCapturedUnionAgrees/alternating.a` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/alternating.a/sanitized` (b): captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1

- `TestCapturedUnionAgrees/alternating.a/counted` (b): captured_union_test.go:110: counts: heap values leaked: 18 (allocations 54, frees 36, in regions 0)



Complete verbatim logs retain every first error and, for non-lowering failures, the whole observed leak/count diagnostic. These tests expect no leaks and balanced counts; their pinned stdout expectations are unchanged. No stdout disagreement was reported.

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== CONT  TestCapturedUnionLowersToOneCell
=== CONT  TestCapturedMaybeBooleanStillStops
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== RUN   TestCapturedUnionLowersToOneCell/cycle.ts
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.ts
=== RUN   TestCapturedUnionLowersToOneCell/cycle.a
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
--- PASS: TestCapturedMaybeBooleanStillStops (0.40s)
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
=== NAME  TestCapturedUnionLowersToOneCell/by_value.ts
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 3 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
=== NAME  TestCapturedUnionLowersToOneCell/graph.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== NAME  TestCapturedUnionLowersToOneCell/class.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
=== NAME  TestCapturedUnionLowersToOneCell/nested.ts
    captured_union_test.go:56: extended: graph cell true
    captured_union_test.go:66: extended: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/object.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/cycle.ts
=== NAME  TestCapturedUnionLowersToOneCell/alternating.ts
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/cycle.a
=== NAME  TestCapturedUnionLowersToOneCell/object.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== NAME  TestCapturedUnionLowersToOneCell/nested.a
    captured_union_test.go:56: extended: graph cell true
    captured_union_test.go:66: extended: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== NAME  TestCapturedUnionLowersToOneCell/alternating.a
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
=== NAME  TestCapturedUnionLowersToOneCell/cycle.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 1 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
=== NAME  TestCapturedUnionLowersToOneCell/released.a
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/released.ts
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== NAME  TestCapturedUnionLowersToOneCell/cycle.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 1 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/graph.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 2 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/by_value.a
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 3 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/class.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
--- PASS: TestCapturedUnionLowersToOneCell (0.00s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.ts (0.41s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.ts (0.41s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.ts (0.43s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.ts (0.46s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.a (0.30s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.ts (0.40s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.ts (0.40s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.a (0.42s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.a (0.52s)
    --- PASS: TestCapturedUnionLowersToOneCell/cycle.ts (0.64s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.a (0.63s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.ts (0.63s)
    --- PASS: TestCapturedUnionLowersToOneCell/cycle.a (0.65s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.a (0.55s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.a (0.14s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.a (0.09s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	1.618s
```


```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== CONT  TestCapturedUnionAgrees
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/cycle.ts
=== PAUSE TestCapturedUnionAgrees/cycle.ts
=== RUN   TestCapturedUnionAgrees/cycle.a
=== PAUSE TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/nested.ts
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== CONT  TestCapturedUnionAgrees/object.ts
=== CONT  TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts/native
=== RUN   TestCapturedUnionAgrees/object.a/native
=== RUN   TestCapturedUnionAgrees/nested.ts/native
=== RUN   TestCapturedUnionAgrees/object.ts/native
=== RUN   TestCapturedUnionAgrees/class.ts/native
=== RUN   TestCapturedUnionAgrees/by_value.ts/sanitized
=== RUN   TestCapturedUnionAgrees/nested.ts/sanitized
=== RUN   TestCapturedUnionAgrees/object.a/sanitized
=== RUN   TestCapturedUnionAgrees/object.ts/sanitized
=== RUN   TestCapturedUnionAgrees/class.ts/sanitized
=== RUN   TestCapturedUnionAgrees/nested.ts/counted
=== NAME  TestCapturedUnionAgrees/by_value.ts/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==10922==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 72 byte(s) in 1 object(s) allocated from:
            #0 0x558dbeea33a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x558dbef0fd2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x558dbef0fd2f in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x558dbeea33a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x558dbef0fd2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x558dbef0fd2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x558dbeee702d in adamic_function_by_value_ts_compare_86144a167155d3c4b9fb1cc42659a1f0 /adamic-build/main.c:136:2
            #4 0x558dbeee702d in adamic_initialize_by_value_ts_bcb45c73d87e7954302beb5e57028536 /adamic-build/main.c:367:2
            #5 0x558dbeee6a6d in adamic_initialize_by_value_ts_a2576a49d678491017dd8307a246d519 /adamic-build/main.c:377:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x558dbeea33a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x558dbef0fd2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x558dbef0fd2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x558dbeee6def in adamic_function_by_value_ts_compare_86144a167155d3c4b9fb1cc42659a1f0 /adamic-build/main.c:91:6
            #4 0x558dbeee6def in adamic_initialize_by_value_ts_bcb45c73d87e7954302beb5e57028536 /adamic-build/main.c:367:2
            #5 0x558dbeee6a6d in adamic_initialize_by_value_ts_a2576a49d678491017dd8307a246d519 /adamic-build/main.c:377:2
        
        SUMMARY: AddressSanitizer: 204 byte(s) leaked in 3 allocation(s).
=== RUN   TestCapturedUnionAgrees/by_value.ts/counted
=== RUN   TestCapturedUnionAgrees/object.a/counted
=== NAME  TestCapturedUnionAgrees/nested.ts/counted
    captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.ts/javascript
=== RUN   TestCapturedUnionAgrees/object.ts/counted
=== RUN   TestCapturedUnionAgrees/class.ts/counted
=== NAME  TestCapturedUnionAgrees/by_value.ts/counted
    captured_union_test.go:110: counts: heap values leaked: 3 (allocations 35, frees 32, in regions 0)
    captured_union_test.go:116: counts: allocations 35, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.ts/javascript
=== NAME  TestCapturedUnionAgrees/object.a/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.a/javascript
=== NAME  TestCapturedUnionAgrees/object.ts/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.ts/javascript
=== NAME  TestCapturedUnionAgrees/class.ts/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.ts/javascript
=== NAME  TestCapturedUnionAgrees/object.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 2.42
=== NAME  TestCapturedUnionAgrees/class.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 1.88
=== NAME  TestCapturedUnionAgrees/nested.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 2.40
=== NAME  TestCapturedUnionAgrees/by_value.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 2.53
=== NAME  TestCapturedUnionAgrees/object.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 2.55
=== CONT  TestCapturedUnionAgrees/graph.ts
=== CONT  TestCapturedUnionAgrees/released.ts
=== CONT  TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/class.a
=== CONT  TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/class.a/native
=== RUN   TestCapturedUnionAgrees/released.ts/native
=== RUN   TestCapturedUnionAgrees/cycle.a/native
=== RUN   TestCapturedUnionAgrees/released.a/native
=== RUN   TestCapturedUnionAgrees/graph.ts/native
=== RUN   TestCapturedUnionAgrees/cycle.a/sanitized
=== RUN   TestCapturedUnionAgrees/released.ts/sanitized
=== RUN   TestCapturedUnionAgrees/released.a/sanitized
=== RUN   TestCapturedUnionAgrees/class.a/sanitized
=== NAME  TestCapturedUnionAgrees/released.a/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==11381==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x5654efdcc3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5654efe393ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5654efe393ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5654efe108cc in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:197:85
            #4 0x5654efe0fd4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x5654efe0fd4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x5654efe0fa6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x5654efdcc3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5654efe393ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5654efe393ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5654efe106a6 in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:178:84
            #4 0x5654efe0fd4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x5654efe0fd4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x5654efe0fa6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x5654efdcc3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5654efe393ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5654efe393ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5654efe10431 in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:145:80
            #4 0x5654efe0fd4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x5654efe0fd4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x5654efe0fa6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 216 byte(s) in 3 object(s) allocated from:
            #0 0x5654efdcc3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5654efe393ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5654efe393ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 1218 byte(s) in 18 object(s) allocated from:
            #0 0x5654efdcc3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5654efe393ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5654efe393ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 216 byte(s) in 9 object(s) allocated from:
            #0 0x5654efdcc3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5654efe14daf in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 2406 byte(s) leaked in 39 allocation(s).
=== RUN   TestCapturedUnionAgrees/released.a/counted
=== RUN   TestCapturedUnionAgrees/cycle.a/counted
=== NAME  TestCapturedUnionAgrees/released.ts/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==11393==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55d6eeed43a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55d6eef413ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55d6eef413ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55d6eef188cc in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:197:85
            #4 0x55d6eef17d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x55d6eef17d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x55d6eef17a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55d6eeed43a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55d6eef413ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55d6eef413ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55d6eef186a6 in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:178:84
            #4 0x55d6eef17d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x55d6eef17d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x55d6eef17a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55d6eeed43a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55d6eef413ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55d6eef413ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55d6eef18431 in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:145:80
            #4 0x55d6eef17d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x55d6eef17d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x55d6eef17a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 216 byte(s) in 3 object(s) allocated from:
            #0 0x55d6eeed43a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55d6eef413ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55d6eef413ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 1218 byte(s) in 18 object(s) allocated from:
            #0 0x55d6eeed43a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55d6eef413ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55d6eef413ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 216 byte(s) in 9 object(s) allocated from:
            #0 0x55d6eeed43a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55d6eef1cdaf in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 2406 byte(s) leaked in 39 allocation(s).
=== RUN   TestCapturedUnionAgrees/released.ts/counted
=== RUN   TestCapturedUnionAgrees/class.a/counted
=== NAME  TestCapturedUnionAgrees/cycle.a/counted
    captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== NAME  TestCapturedUnionAgrees/released.a/counted
    captured_union_test.go:110: counts: heap values leaked: 30 (allocations 96, frees 66, in regions 0)
    captured_union_test.go:116: counts: allocations 96, frees 66, in regions 0
=== RUN   TestCapturedUnionAgrees/released.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.06
=== NAME  TestCapturedUnionAgrees/released.ts/counted
    captured_union_test.go:110: counts: heap values leaked: 30 (allocations 96, frees 66, in regions 0)
    captured_union_test.go:116: counts: allocations 96, frees 66, in regions 0
=== RUN   TestCapturedUnionAgrees/released.ts/javascript
=== CONT  TestCapturedUnionAgrees/cycle.ts
=== NAME  TestCapturedUnionAgrees/released.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.05
=== CONT  TestCapturedUnionAgrees/by_value.a
=== CONT  TestCapturedUnionAgrees/graph.a
=== NAME  TestCapturedUnionAgrees/class.a/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.04
=== RUN   TestCapturedUnionAgrees/cycle.ts/native
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/by_value.a/native
=== RUN   TestCapturedUnionAgrees/graph.a/native
=== RUN   TestCapturedUnionAgrees/alternating.ts/native
=== RUN   TestCapturedUnionAgrees/cycle.ts/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.a/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.ts/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.ts/counted
=== NAME  TestCapturedUnionAgrees/by_value.a/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==11560==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 72 byte(s) in 1 object(s) allocated from:
            #0 0x5604b327f3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5604b32ebd2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5604b32ebd2f in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x5604b327f3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5604b32ebd2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5604b32ebd2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5604b32c302d in adamic_function_by_value_a_compare_1954618ce2d9e07b1980a9ac4e4160ae /adamic-build/main.c:136:2
            #4 0x5604b32c302d in adamic_initialize_by_value_a_4c44bd7515c4580a5a4f6d72acde192b /adamic-build/main.c:367:2
            #5 0x5604b32c2a6d in adamic_initialize_by_value_a_e96b858367cf1e592bc28bb5116317f3 /adamic-build/main.c:377:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x5604b327f3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5604b32ebd2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5604b32ebd2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5604b32c2def in adamic_function_by_value_a_compare_1954618ce2d9e07b1980a9ac4e4160ae /adamic-build/main.c:91:6
            #4 0x5604b32c2def in adamic_initialize_by_value_a_4c44bd7515c4580a5a4f6d72acde192b /adamic-build/main.c:367:2
            #5 0x5604b32c2a6d in adamic_initialize_by_value_a_e96b858367cf1e592bc28bb5116317f3 /adamic-build/main.c:377:2
        
        SUMMARY: AddressSanitizer: 204 byte(s) leaked in 3 allocation(s).
=== RUN   TestCapturedUnionAgrees/by_value.a/counted
=== NAME  TestCapturedUnionAgrees/alternating.ts/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==11572==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55df26e6a3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55df26ed6dcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55df26ed6dcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55df26eade1d in adamic_function_alternating_ts_alternate_565c9a204ed396dab50febfce9b6094f /adamic-build/main.c:75:6
            #4 0x55df26eadac0 in adamic_initialize_alternating_ts_798a1070fca02ff27cf4b4600fa7ac9f /adamic-build/main.c:404:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55df26e6a3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55df26ed6dcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55df26ed6dcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55df26eade1d in adamic_function_alternating_ts_alternate_565c9a204ed396dab50febfce9b6094f /adamic-build/main.c:75:6
            #4 0x55df26eada84 in adamic_initialize_alternating_ts_798a1070fca02ff27cf4b4600fa7ac9f /adamic-build/main.c:400:2
        
        Indirect leak of 834 byte(s) in 12 object(s) allocated from:
            #0 0x55df26e6a3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55df26ed6dcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55df26ed6dcf in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 96 byte(s) in 6 object(s) allocated from:
            #0 0x55df26e6a3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55df26eb277f in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 1434 byte(s) leaked in 24 allocation(s).
=== RUN   TestCapturedUnionAgrees/alternating.ts/counted
=== NAME  TestCapturedUnionAgrees/cycle.ts/counted
    captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== CONT  TestCapturedUnionAgrees/nested.a
=== NAME  TestCapturedUnionAgrees/by_value.a/counted
    captured_union_test.go:110: counts: heap values leaked: 3 (allocations 35, frees 32, in regions 0)
    captured_union_test.go:116: counts: allocations 35, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.03
=== CONT  TestCapturedUnionAgrees/alternating.a
=== NAME  TestCapturedUnionAgrees/alternating.ts/counted
    captured_union_test.go:110: counts: heap values leaked: 18 (allocations 54, frees 36, in regions 0)
    captured_union_test.go:116: counts: allocations 54, frees 36, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.06
=== RUN   TestCapturedUnionAgrees/nested.a/native
=== RUN   TestCapturedUnionAgrees/graph.a/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.a/native
=== RUN   TestCapturedUnionAgrees/graph.ts/sanitized
=== RUN   TestCapturedUnionAgrees/nested.a/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.a/sanitized
=== RUN   TestCapturedUnionAgrees/nested.a/counted
=== NAME  TestCapturedUnionAgrees/alternating.a/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==11689==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x555e35a9c3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x555e35b08dcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x555e35b08dcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x555e35adfe1d in adamic_function_alternating_a_alternate_01e607951fdf653584ac75f3f81a1146 /adamic-build/main.c:75:6
            #4 0x555e35adfac0 in adamic_initialize_alternating_a_8aabf12e8e980896d6070e8d5697d4ba /adamic-build/main.c:404:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x555e35a9c3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x555e35b08dcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x555e35b08dcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x555e35adfe1d in adamic_function_alternating_a_alternate_01e607951fdf653584ac75f3f81a1146 /adamic-build/main.c:75:6
            #4 0x555e35adfa84 in adamic_initialize_alternating_a_8aabf12e8e980896d6070e8d5697d4ba /adamic-build/main.c:400:2
        
        Indirect leak of 834 byte(s) in 12 object(s) allocated from:
            #0 0x555e35a9c3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x555e35b08dcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x555e35b08dcf in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 96 byte(s) in 6 object(s) allocated from:
            #0 0x555e35a9c3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x555e35ae477f in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 1434 byte(s) leaked in 24 allocation(s).
=== RUN   TestCapturedUnionAgrees/alternating.a/counted
=== NAME  TestCapturedUnionAgrees/nested.a/counted
    captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== NAME  TestCapturedUnionAgrees/alternating.a/counted
    captured_union_test.go:110: counts: heap values leaked: 18 (allocations 54, frees 36, in regions 0)
    captured_union_test.go:116: counts: allocations 54, frees 36, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== RUN   TestCapturedUnionAgrees/graph.ts/counted
=== RUN   TestCapturedUnionAgrees/graph.a/counted
=== NAME  TestCapturedUnionAgrees/graph.ts/counted
    captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.ts/javascript
=== NAME  TestCapturedUnionAgrees/graph.a/counted
    captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.a/javascript
=== NAME  TestCapturedUnionAgrees/graph.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== NAME  TestCapturedUnionAgrees/graph.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
--- FAIL: TestCapturedUnionAgrees (0.00s)
    --- PASS: TestCapturedUnionAgrees/nested.ts (17.33s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/native (0.73s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/sanitized (1.19s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/counted (0.78s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/javascript (4.49s)
    --- PASS: TestCapturedUnionAgrees/class.ts (17.35s)
        --- PASS: TestCapturedUnionAgrees/class.ts/native (1.10s)
        --- PASS: TestCapturedUnionAgrees/class.ts/sanitized (1.62s)
        --- PASS: TestCapturedUnionAgrees/class.ts/counted (2.15s)
        --- PASS: TestCapturedUnionAgrees/class.ts/javascript (2.13s)
    --- FAIL: TestCapturedUnionAgrees/by_value.ts (17.46s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/native (0.64s)
        --- FAIL: TestCapturedUnionAgrees/by_value.ts/sanitized (1.47s)
        --- FAIL: TestCapturedUnionAgrees/by_value.ts/counted (1.14s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/javascript (4.19s)
    --- PASS: TestCapturedUnionAgrees/object.ts (17.51s)
        --- PASS: TestCapturedUnionAgrees/object.ts/native (1.12s)
        --- PASS: TestCapturedUnionAgrees/object.ts/sanitized (1.77s)
        --- PASS: TestCapturedUnionAgrees/object.ts/counted (1.57s)
        --- PASS: TestCapturedUnionAgrees/object.ts/javascript (2.89s)
    --- PASS: TestCapturedUnionAgrees/object.a (17.51s)
        --- PASS: TestCapturedUnionAgrees/object.a/native (0.89s)
        --- PASS: TestCapturedUnionAgrees/object.a/sanitized (1.77s)
        --- PASS: TestCapturedUnionAgrees/object.a/counted (1.84s)
        --- PASS: TestCapturedUnionAgrees/object.a/javascript (2.94s)
    --- PASS: TestCapturedUnionAgrees/cycle.a (3.76s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/native (0.75s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/sanitized (1.32s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/counted (0.71s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/javascript (0.29s)
    --- FAIL: TestCapturedUnionAgrees/released.a (3.84s)
        --- PASS: TestCapturedUnionAgrees/released.a/native (0.70s)
        --- FAIL: TestCapturedUnionAgrees/released.a/sanitized (1.23s)
        --- FAIL: TestCapturedUnionAgrees/released.a/counted (0.80s)
        --- PASS: TestCapturedUnionAgrees/released.a/javascript (0.34s)
    --- FAIL: TestCapturedUnionAgrees/released.ts (4.22s)
        --- PASS: TestCapturedUnionAgrees/released.ts/native (0.81s)
        --- FAIL: TestCapturedUnionAgrees/released.ts/sanitized (1.43s)
        --- FAIL: TestCapturedUnionAgrees/released.ts/counted (0.81s)
        --- PASS: TestCapturedUnionAgrees/released.ts/javascript (0.36s)
    --- PASS: TestCapturedUnionAgrees/class.a (4.63s)
        --- PASS: TestCapturedUnionAgrees/class.a/native (0.94s)
        --- PASS: TestCapturedUnionAgrees/class.a/sanitized (1.62s)
        --- PASS: TestCapturedUnionAgrees/class.a/counted (1.17s)
        --- PASS: TestCapturedUnionAgrees/class.a/javascript (0.29s)
    --- PASS: TestCapturedUnionAgrees/cycle.ts (3.39s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/native (0.70s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/sanitized (0.96s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/counted (0.74s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/javascript (0.20s)
    --- FAIL: TestCapturedUnionAgrees/by_value.a (3.46s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/native (0.68s)
        --- FAIL: TestCapturedUnionAgrees/by_value.a/sanitized (1.12s)
        --- FAIL: TestCapturedUnionAgrees/by_value.a/counted (0.66s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/javascript (0.20s)
    --- FAIL: TestCapturedUnionAgrees/alternating.ts (3.13s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/native (0.56s)
        --- FAIL: TestCapturedUnionAgrees/alternating.ts/sanitized (1.12s)
        --- FAIL: TestCapturedUnionAgrees/alternating.ts/counted (0.61s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/javascript (0.31s)
    --- PASS: TestCapturedUnionAgrees/nested.a (2.66s)
        --- PASS: TestCapturedUnionAgrees/nested.a/native (0.56s)
        --- PASS: TestCapturedUnionAgrees/nested.a/sanitized (0.87s)
        --- PASS: TestCapturedUnionAgrees/nested.a/counted (0.50s)
        --- PASS: TestCapturedUnionAgrees/nested.a/javascript (0.15s)
    --- FAIL: TestCapturedUnionAgrees/alternating.a (2.59s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/native (0.56s)
        --- FAIL: TestCapturedUnionAgrees/alternating.a/sanitized (0.88s)
        --- FAIL: TestCapturedUnionAgrees/alternating.a/counted (0.43s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/javascript (0.13s)
    --- PASS: TestCapturedUnionAgrees/graph.ts (15.34s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/native (7.09s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/sanitized (4.24s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/counted (2.88s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/javascript (0.12s)
    --- PASS: TestCapturedUnionAgrees/graph.a (11.13s)
        --- PASS: TestCapturedUnionAgrees/graph.a/native (2.95s)
        --- PASS: TestCapturedUnionAgrees/graph.a/sanitized (4.36s)
        --- PASS: TestCapturedUnionAgrees/graph.a/counted (2.89s)
        --- PASS: TestCapturedUnionAgrees/graph.a/javascript (0.13s)
FAIL
gate cache: native hits=0 misses=32
gate cache: node hits=0 misses=32
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	32.738s
FAIL
```


### M4: The overwritten value kept in a graph cell

```text
M4 old text occurs exactly once
M4-a exit=0 wall=51
M4-b exit=1 wall=58
restore: git diff --exit-code = 0; output empty
```


All lowering pins and the maybe-boolean guard pass. Per pin and dialect, sanitized and counted results and the exact logged counts line:

| Pin | Sanitized | Counted | Counts line |
| --- | --- | --- | --- |
| class.ts | FAIL | FAIL | captured_union_test.go:116: counts: allocations 21, frees 18, in regions 0 |
| class.a | FAIL | FAIL | captured_union_test.go:116: counts: allocations 21, frees 18, in regions 0 |
| object.ts | FAIL | FAIL | captured_union_test.go:116: counts: allocations 21, frees 18, in regions 0 |
| object.a | FAIL | FAIL | captured_union_test.go:116: counts: allocations 21, frees 18, in regions 0 |
| by_value.ts | PASS | PASS | captured_union_test.go:116: counts: allocations 35, frees 35, in regions 0 |
| by_value.a | PASS | PASS | captured_union_test.go:116: counts: allocations 35, frees 35, in regions 0 |
| released.ts | PASS | PASS | captured_union_test.go:116: counts: allocations 96, frees 96, in regions 0 |
| released.a | PASS | PASS | captured_union_test.go:116: counts: allocations 96, frees 96, in regions 0 |
| nested.ts | FAIL | FAIL | captured_union_test.go:116: counts: allocations 97, frees 91, in regions 0 |
| nested.a | FAIL | FAIL | captured_union_test.go:116: counts: allocations 97, frees 91, in regions 0 |
| alternating.ts | PASS | PASS | captured_union_test.go:116: counts: allocations 54, frees 54, in regions 0 |
| alternating.a | PASS | PASS | captured_union_test.go:116: counts: allocations 54, frees 54, in regions 0 |
| graph.ts | FAIL | FAIL | captured_union_test.go:116: counts: allocations 42, frees 32, in regions 0 |
| graph.a | FAIL | FAIL | captured_union_test.go:116: counts: allocations 42, frees 32, in regions 0 |
| cycle.ts | PASS | PASS | captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0 |
| cycle.a | PASS | PASS | captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0 |


Every red subtest and its first diagnostic line (parent failures inherit their child failures):

- `TestCapturedUnionAgrees` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/nested.ts` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/nested.ts/sanitized` (b): captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1

- `TestCapturedUnionAgrees/nested.ts/counted` (b): captured_union_test.go:110: counts: heap values leaked: 6 (allocations 97, frees 91, in regions 0)

- `TestCapturedUnionAgrees/class.ts` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/class.ts/sanitized` (b): captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1

- `TestCapturedUnionAgrees/class.ts/counted` (b): captured_union_test.go:110: counts: heap values leaked: 3 (allocations 21, frees 18, in regions 0)

- `TestCapturedUnionAgrees/nested.a` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/nested.a/sanitized` (b): captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1

- `TestCapturedUnionAgrees/nested.a/counted` (b): captured_union_test.go:110: counts: heap values leaked: 6 (allocations 97, frees 91, in regions 0)

- `TestCapturedUnionAgrees/object.ts` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/object.ts/sanitized` (b): captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1

- `TestCapturedUnionAgrees/object.ts/counted` (b): captured_union_test.go:110: counts: heap values leaked: 3 (allocations 21, frees 18, in regions 0)

- `TestCapturedUnionAgrees/object.a` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/object.a/sanitized` (b): captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1

- `TestCapturedUnionAgrees/object.a/counted` (b): captured_union_test.go:110: counts: heap values leaked: 3 (allocations 21, frees 18, in regions 0)

- `TestCapturedUnionAgrees/graph.a` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/graph.a/sanitized` (b): captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1

- `TestCapturedUnionAgrees/graph.a/counted` (b): captured_union_test.go:110: counts: heap values leaked: 10 (allocations 42, frees 32, in regions 0)

- `TestCapturedUnionAgrees/graph.ts` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/graph.ts/sanitized` (b): captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1

- `TestCapturedUnionAgrees/graph.ts/counted` (b): captured_union_test.go:110: counts: heap values leaked: 10 (allocations 42, frees 32, in regions 0)

- `TestCapturedUnionAgrees/class.a` (b): see failed child subtests below/in the verbatim log.

- `TestCapturedUnionAgrees/class.a/sanitized` (b): captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1

- `TestCapturedUnionAgrees/class.a/counted` (b): captured_union_test.go:110: counts: heap values leaked: 3 (allocations 21, frees 18, in regions 0)



Complete verbatim logs retain every first error and, for non-lowering failures, the whole observed leak/count diagnostic. These tests expect no leaks and balanced counts; their pinned stdout expectations are unchanged. No stdout disagreement was reported.

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== CONT  TestCapturedUnionLowersToOneCell
=== CONT  TestCapturedMaybeBooleanStillStops
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== RUN   TestCapturedUnionLowersToOneCell/cycle.ts
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.ts
=== RUN   TestCapturedUnionLowersToOneCell/cycle.a
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.a
=== CONT  TestCapturedUnionLowersToOneCell/cycle.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedUnionLowersToOneCell/cycle.ts
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
--- PASS: TestCapturedMaybeBooleanStillStops (0.67s)
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
=== NAME  TestCapturedUnionLowersToOneCell/cycle.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 1 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
=== NAME  TestCapturedUnionLowersToOneCell/class.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
=== NAME  TestCapturedUnionLowersToOneCell/cycle.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 1 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== NAME  TestCapturedUnionLowersToOneCell/graph.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== NAME  TestCapturedUnionLowersToOneCell/alternating.a
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== NAME  TestCapturedUnionLowersToOneCell/graph.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== NAME  TestCapturedUnionLowersToOneCell/alternating.ts
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
=== NAME  TestCapturedUnionLowersToOneCell/nested.a
    captured_union_test.go:56: extended: graph cell true
    captured_union_test.go:66: extended: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
=== NAME  TestCapturedUnionLowersToOneCell/nested.ts
    captured_union_test.go:56: extended: graph cell true
    captured_union_test.go:66: extended: 2 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/object.a
=== NAME  TestCapturedUnionLowersToOneCell/by_value.a
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 3 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
=== NAME  TestCapturedUnionLowersToOneCell/by_value.ts
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 3 closures carry its cell
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== NAME  TestCapturedUnionLowersToOneCell/released.ts
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/released.a
    captured_union_test.go:56: active: graph cell false
    captured_union_test.go:66: active: 2 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/object.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/object.ts
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
=== NAME  TestCapturedUnionLowersToOneCell/class.a
    captured_union_test.go:56: active: graph cell true
    captured_union_test.go:66: active: 5 closures carry its cell
--- PASS: TestCapturedUnionLowersToOneCell (0.00s)
    --- PASS: TestCapturedUnionLowersToOneCell/cycle.ts (0.72s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.ts (0.76s)
    --- PASS: TestCapturedUnionLowersToOneCell/cycle.a (0.79s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.a (0.80s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.a (0.34s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.ts (0.39s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.ts (0.31s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.a (0.29s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.ts (0.28s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.a (0.16s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.ts (0.15s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.ts (0.17s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.a (0.21s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.a (0.21s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.ts (0.12s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.a (0.12s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	1.414s
```


```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== CONT  TestCapturedUnionAgrees
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/cycle.ts
=== PAUSE TestCapturedUnionAgrees/cycle.ts
=== RUN   TestCapturedUnionAgrees/cycle.a
=== PAUSE TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/nested.ts
=== CONT  TestCapturedUnionAgrees/graph.ts
=== CONT  TestCapturedUnionAgrees/cycle.ts
=== RUN   TestCapturedUnionAgrees/graph.ts/native
=== RUN   TestCapturedUnionAgrees/nested.ts/native
=== RUN   TestCapturedUnionAgrees/cycle.a/native
=== RUN   TestCapturedUnionAgrees/cycle.ts/native
=== RUN   TestCapturedUnionAgrees/class.ts/native
=== RUN   TestCapturedUnionAgrees/nested.ts/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.ts/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.a/sanitized
=== RUN   TestCapturedUnionAgrees/class.ts/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.ts/counted
=== RUN   TestCapturedUnionAgrees/cycle.a/counted
=== NAME  TestCapturedUnionAgrees/class.ts/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==12570==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x5651c5bc03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5651c5c2e2ef in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5651c5c2e2ef in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5651c5c05c41 in adamic_function_class_ts_profiler_enable_anonymous_1_2721b462350c38cee6a81cc0af2cf143 /adamic-build/main.c:538:79
            #4 0x5651c5c04547 in adamic_function_class_ts_profiler_59a94491b94e3e188244d081756ae740 /adamic-build/main.c:163:76
            #5 0x5651c5c04547 in adamic_initialize_class_ts_d1625161018c9574bd4c8cec14b7b9c3 /adamic-build/main.c:627:2
            #6 0x5651c5c03a6d in adamic_initialize_class_ts_4cf89b3161214ac299c03a541426de7e /adamic-build/main.c:637:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x5651c5bc03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5651c5c2e2ef in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5651c5c2e2ef in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5651c5c05c41 in adamic_function_class_ts_profiler_enable_anonymous_1_2721b462350c38cee6a81cc0af2cf143 /adamic-build/main.c:538:79
            #4 0x5651c5c048b6 in adamic_function_class_ts_profiler_59a94491b94e3e188244d081756ae740 /adamic-build/main.c:232:76
            #5 0x5651c5c048b6 in adamic_initialize_class_ts_d1625161018c9574bd4c8cec14b7b9c3 /adamic-build/main.c:627:2
            #6 0x5651c5c03a6d in adamic_initialize_class_ts_4cf89b3161214ac299c03a541426de7e /adamic-build/main.c:637:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x5651c5bc03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5651c5c2e2ef in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5651c5c2e2ef in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5651c5c055d5 in adamic_function_class_ts_profiler_write_anonymous_1_a54d6540e4885f246c5fe301210a0503 /adamic-build/main.c:379:80
            #4 0x5651c5c041a2 in adamic_function_class_ts_profiler_59a94491b94e3e188244d081756ae740 /adamic-build/main.c:115:6
            #5 0x5651c5c041a2 in adamic_initialize_class_ts_d1625161018c9574bd4c8cec14b7b9c3 /adamic-build/main.c:627:2
            #6 0x5651c5c03a6d in adamic_initialize_class_ts_4cf89b3161214ac299c03a541426de7e /adamic-build/main.c:637:2
        
        SUMMARY: AddressSanitizer: 198 byte(s) leaked in 3 allocation(s).
=== RUN   TestCapturedUnionAgrees/class.ts/counted
=== NAME  TestCapturedUnionAgrees/cycle.ts/counted
    captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.ts/javascript
=== NAME  TestCapturedUnionAgrees/cycle.a/counted
    captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.a/javascript
=== NAME  TestCapturedUnionAgrees/class.ts/counted
    captured_union_test.go:110: counts: heap values leaked: 3 (allocations 21, frees 18, in regions 0)
    captured_union_test.go:116: counts: allocations 21, frees 18, in regions 0
=== RUN   TestCapturedUnionAgrees/class.ts/javascript
=== NAME  TestCapturedUnionAgrees/nested.ts/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==12634==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 192 byte(s) in 3 object(s) allocated from:
            #0 0x5604db0b03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5604db11cfcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5604db11cfcf in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 192 byte(s) in 3 object(s) allocated from:
            #0 0x5604db0b03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5604db11cfcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5604db11cfcf in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 32 byte(s) in 1 object(s) allocated from:
            #0 0x5604db0b079c in realloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:81:3
            #1 0x5604db0f8bc5 in adamic_array_push /adamic-runtime/array.c:46:25
        
        SUMMARY: AddressSanitizer: 416 byte(s) leaked in 7 allocation(s).
=== RUN   TestCapturedUnionAgrees/nested.ts/counted
    captured_union_test.go:110: counts: heap values leaked: 6 (allocations 97, frees 91, in regions 0)
    captured_union_test.go:116: counts: allocations 97, frees 91, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.52
=== NAME  TestCapturedUnionAgrees/class.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 1.06
=== CONT  TestCapturedUnionAgrees/graph.a
=== NAME  TestCapturedUnionAgrees/cycle.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 1.76
=== CONT  TestCapturedUnionAgrees/nested.a
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/graph.a/native
=== RUN   TestCapturedUnionAgrees/alternating.ts/native
=== RUN   TestCapturedUnionAgrees/nested.a/native
=== NAME  TestCapturedUnionAgrees/cycle.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 3.33
=== RUN   TestCapturedUnionAgrees/alternating.ts/sanitized
=== RUN   TestCapturedUnionAgrees/graph.ts/sanitized
=== RUN   TestCapturedUnionAgrees/nested.a/sanitized
=== CONT  TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.a/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.ts/counted
=== RUN   TestCapturedUnionAgrees/alternating.a/native
=== NAME  TestCapturedUnionAgrees/alternating.ts/counted
    captured_union_test.go:116: counts: allocations 54, frees 54, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.ts/native
=== RUN   TestCapturedUnionAgrees/alternating.a/sanitized
=== NAME  TestCapturedUnionAgrees/nested.a/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==13005==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 192 byte(s) in 3 object(s) allocated from:
            #0 0x55c1b64b53a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55c1b6521fcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55c1b6521fcf in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 192 byte(s) in 3 object(s) allocated from:
            #0 0x55c1b64b53a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55c1b6521fcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55c1b6521fcf in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 32 byte(s) in 1 object(s) allocated from:
            #0 0x55c1b64b579c in realloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:81:3
            #1 0x55c1b64fdbc5 in adamic_array_push /adamic-runtime/array.c:46:25
        
        SUMMARY: AddressSanitizer: 416 byte(s) leaked in 7 allocation(s).
=== RUN   TestCapturedUnionAgrees/nested.a/counted
=== RUN   TestCapturedUnionAgrees/by_value.ts/sanitized
=== NAME  TestCapturedUnionAgrees/nested.a/counted
    captured_union_test.go:110: counts: heap values leaked: 6 (allocations 97, frees 91, in regions 0)
    captured_union_test.go:116: counts: allocations 97, frees 91, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.04
=== CONT  TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.ts/native
=== RUN   TestCapturedUnionAgrees/alternating.a/counted
=== RUN   TestCapturedUnionAgrees/by_value.ts/counted
    captured_union_test.go:116: counts: allocations 35, frees 35, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.10
=== NAME  TestCapturedUnionAgrees/alternating.a/counted
    captured_union_test.go:116: counts: allocations 54, frees 54, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.09
=== RUN   TestCapturedUnionAgrees/released.ts/sanitized
=== CONT  TestCapturedUnionAgrees/by_value.a
=== CONT  TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/released.a/native
=== RUN   TestCapturedUnionAgrees/released.ts/counted
=== RUN   TestCapturedUnionAgrees/by_value.a/native
=== NAME  TestCapturedUnionAgrees/graph.a/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==13160==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 132 byte(s) in 2 object(s) allocated from:
            #0 0x55b86d7f63a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b86d863d8f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b86d863d8f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55b86d83b514 in adamic_function_graph_a_run_controls_write_a800635d31c0d83b98729b863cf1af33 /adamic-build/main.c:253:82
            #4 0x55b86d83a394 in adamic_closure_call /adamic-build/runtime/adamic.h:359:9
            #5 0x55b86d83a394 in adamic_closure_receiver_call /adamic-build/runtime/adamic.h:367:9
            #6 0x55b86d83a394 in adamic_function_graph_a_run_a261783b039b35063576d66c0726e8ff /adamic-build/main.c:119:150
            #7 0x55b86d839ab5 in adamic_initialize_graph_a_59ce64e44011e73fab583d486e040075 /adamic-build/main.c:435:2
        
        Direct leak of 132 byte(s) in 2 object(s) allocated from:
            #0 0x55b86d7f63a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b86d863d8f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b86d863d8f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55b86d83b514 in adamic_function_graph_a_run_controls_write_a800635d31c0d83b98729b863cf1af33 /adamic-build/main.c:253:82
            #4 0x55b86d83a394 in adamic_closure_call /adamic-build/runtime/adamic.h:359:9
            #5 0x55b86d83a394 in adamic_closure_receiver_call /adamic-build/runtime/adamic.h:367:9
            #6 0x55b86d83a394 in adamic_function_graph_a_run_a261783b039b35063576d66c0726e8ff /adamic-build/main.c:119:150
            #7 0x55b86d839a81 in adamic_initialize_graph_a_59ce64e44011e73fab583d486e040075 /adamic-build/main.c:431:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x55b86d7f63a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b86d863d8f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b86d863d8f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55b86d83b23a in adamic_function_graph_a_run_controls_write_a800635d31c0d83b98729b863cf1af33 /adamic-build/main.c:223:80
            #4 0x55b86d83a394 in adamic_closure_call /adamic-build/runtime/adamic.h:359:9
            #5 0x55b86d83a394 in adamic_closure_receiver_call /adamic-build/runtime/adamic.h:367:9
            #6 0x55b86d83a394 in adamic_function_graph_a_run_a261783b039b35063576d66c0726e8ff /adamic-build/main.c:119:150
            #7 0x55b86d839ab5 in adamic_initialize_graph_a_59ce64e44011e73fab583d486e040075 /adamic-build/main.c:435:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x55b86d7f63a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b86d863d8f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b86d863d8f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55b86d83b23a in adamic_function_graph_a_run_controls_write_a800635d31c0d83b98729b863cf1af33 /adamic-build/main.c:223:80
            #4 0x55b86d83a394 in adamic_closure_call /adamic-build/runtime/adamic.h:359:9
            #5 0x55b86d83a394 in adamic_closure_receiver_call /adamic-build/runtime/adamic.h:367:9
            #6 0x55b86d83a394 in adamic_function_graph_a_run_a261783b039b35063576d66c0726e8ff /adamic-build/main.c:119:150
            #7 0x55b86d839a81 in adamic_initialize_graph_a_59ce64e44011e73fab583d486e040075 /adamic-build/main.c:431:2
        
        Indirect leak of 284 byte(s) in 4 object(s) allocated from:
            #0 0x55b86d7f63a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b86d863d8f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b86d863d8f in adamic_allocate /adamic-runtime/heap.c:267:9
        
        SUMMARY: AddressSanitizer: 680 byte(s) leaked in 10 allocation(s).
=== RUN   TestCapturedUnionAgrees/graph.a/counted
=== RUN   TestCapturedUnionAgrees/released.a/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.a/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.a/counted
=== NAME  TestCapturedUnionAgrees/released.ts/counted
    captured_union_test.go:116: counts: allocations 96, frees 96, in regions 0
=== RUN   TestCapturedUnionAgrees/released.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== NAME  TestCapturedUnionAgrees/graph.ts/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==13205==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 132 byte(s) in 2 object(s) allocated from:
            #0 0x5649f15d33a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5649f1640d8f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5649f1640d8f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5649f1618514 in adamic_function_graph_ts_run_controls_write_f53f58b166562adca36aabea00dd916d /adamic-build/main.c:253:82
            #4 0x5649f1617394 in adamic_closure_call /adamic-build/runtime/adamic.h:359:9
            #5 0x5649f1617394 in adamic_closure_receiver_call /adamic-build/runtime/adamic.h:367:9
            #6 0x5649f1617394 in adamic_function_graph_ts_run_c271e8d5475dbd1e0d78697a11143e51 /adamic-build/main.c:119:150
            #7 0x5649f1616ab5 in adamic_initialize_graph_ts_74e8577b47cb3406b88b61720406a731 /adamic-build/main.c:435:2
        
        Direct leak of 132 byte(s) in 2 object(s) allocated from:
            #0 0x5649f15d33a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5649f1640d8f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5649f1640d8f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5649f1618514 in adamic_function_graph_ts_run_controls_write_f53f58b166562adca36aabea00dd916d /adamic-build/main.c:253:82
            #4 0x5649f1617394 in adamic_closure_call /adamic-build/runtime/adamic.h:359:9
            #5 0x5649f1617394 in adamic_closure_receiver_call /adamic-build/runtime/adamic.h:367:9
            #6 0x5649f1617394 in adamic_function_graph_ts_run_c271e8d5475dbd1e0d78697a11143e51 /adamic-build/main.c:119:150
            #7 0x5649f1616a81 in adamic_initialize_graph_ts_74e8577b47cb3406b88b61720406a731 /adamic-build/main.c:431:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x5649f15d33a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5649f1640d8f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5649f1640d8f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5649f161823a in adamic_function_graph_ts_run_controls_write_f53f58b166562adca36aabea00dd916d /adamic-build/main.c:223:80
            #4 0x5649f1617394 in adamic_closure_call /adamic-build/runtime/adamic.h:359:9
            #5 0x5649f1617394 in adamic_closure_receiver_call /adamic-build/runtime/adamic.h:367:9
            #6 0x5649f1617394 in adamic_function_graph_ts_run_c271e8d5475dbd1e0d78697a11143e51 /adamic-build/main.c:119:150
            #7 0x5649f1616ab5 in adamic_initialize_graph_ts_74e8577b47cb3406b88b61720406a731 /adamic-build/main.c:435:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x5649f15d33a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5649f1640d8f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5649f1640d8f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5649f161823a in adamic_function_graph_ts_run_controls_write_f53f58b166562adca36aabea00dd916d /adamic-build/main.c:223:80
            #4 0x5649f1617394 in adamic_closure_call /adamic-build/runtime/adamic.h:359:9
            #5 0x5649f1617394 in adamic_closure_receiver_call /adamic-build/runtime/adamic.h:367:9
            #6 0x5649f1617394 in adamic_function_graph_ts_run_c271e8d5475dbd1e0d78697a11143e51 /adamic-build/main.c:119:150
            #7 0x5649f1616a81 in adamic_initialize_graph_ts_74e8577b47cb3406b88b61720406a731 /adamic-build/main.c:431:2
        
        Indirect leak of 284 byte(s) in 4 object(s) allocated from:
            #0 0x5649f15d33a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5649f1640d8f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5649f1640d8f in adamic_allocate /adamic-runtime/heap.c:267:9
        
        SUMMARY: AddressSanitizer: 680 byte(s) leaked in 10 allocation(s).
=== RUN   TestCapturedUnionAgrees/graph.ts/counted
=== CONT  TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.ts/native
=== NAME  TestCapturedUnionAgrees/by_value.a/counted
    captured_union_test.go:116: counts: allocations 35, frees 35, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== CONT  TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/released.a/counted
=== RUN   TestCapturedUnionAgrees/object.ts/sanitized
=== RUN   TestCapturedUnionAgrees/class.a/native
=== NAME  TestCapturedUnionAgrees/released.a/counted
    captured_union_test.go:116: counts: allocations 96, frees 96, in regions 0
=== RUN   TestCapturedUnionAgrees/released.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.03
=== RUN   TestCapturedUnionAgrees/class.a/sanitized
=== CONT  TestCapturedUnionAgrees/object.a
=== NAME  TestCapturedUnionAgrees/object.ts/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==13330==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x55b466c763a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b466ce3fbf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b466ce3fbf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55b466cba547 in adamic_function_object_ts_profiler_6fb390cb4c527bc582d38dd08f2415bc /adamic-build/main.c:150:76
            #4 0x55b466cba547 in adamic_initialize_object_ts_3e23f1aa4d0ad0af12e50c80fa485663 /adamic-build/main.c:549:2
            #5 0x55b466cb9a6d in adamic_initialize_object_ts_d880108c4f605c546a617a0e318992a4 /adamic-build/main.c:559:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x55b466c763a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b466ce3fbf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b466ce3fbf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55b466cba8b6 in adamic_function_object_ts_profiler_6fb390cb4c527bc582d38dd08f2415bc /adamic-build/main.c:219:76
            #4 0x55b466cba8b6 in adamic_initialize_object_ts_3e23f1aa4d0ad0af12e50c80fa485663 /adamic-build/main.c:549:2
            #5 0x55b466cb9a6d in adamic_initialize_object_ts_d880108c4f605c546a617a0e318992a4 /adamic-build/main.c:559:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x55b466c763a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b466ce3fbf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b466ce3fbf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55b466cba1a2 in adamic_function_object_ts_profiler_6fb390cb4c527bc582d38dd08f2415bc /adamic-build/main.c:102:6
            #4 0x55b466cba1a2 in adamic_initialize_object_ts_3e23f1aa4d0ad0af12e50c80fa485663 /adamic-build/main.c:549:2
            #5 0x55b466cb9a6d in adamic_initialize_object_ts_d880108c4f605c546a617a0e318992a4 /adamic-build/main.c:559:2
        
        SUMMARY: AddressSanitizer: 198 byte(s) leaked in 3 allocation(s).
=== RUN   TestCapturedUnionAgrees/object.ts/counted
=== RUN   TestCapturedUnionAgrees/object.a/native
=== NAME  TestCapturedUnionAgrees/object.ts/counted
    captured_union_test.go:110: counts: heap values leaked: 3 (allocations 21, frees 18, in regions 0)
    captured_union_test.go:116: counts: allocations 21, frees 18, in regions 0
=== RUN   TestCapturedUnionAgrees/object.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== RUN   TestCapturedUnionAgrees/object.a/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==13365==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x55fe91fc83a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55fe92035fbf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55fe92035fbf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55fe9200c547 in adamic_function_object_a_profiler_86162185be5ce0c466ad01f5424c8c38 /adamic-build/main.c:150:76
            #4 0x55fe9200c547 in adamic_initialize_object_a_6f36c7b5d0a9cb6b9e4c05117fed7d08 /adamic-build/main.c:549:2
            #5 0x55fe9200ba6d in adamic_initialize_object_a_e22c7958b07f44ff41b0e96fdd303f35 /adamic-build/main.c:559:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x55fe91fc83a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55fe92035fbf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55fe92035fbf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55fe9200c8b6 in adamic_function_object_a_profiler_86162185be5ce0c466ad01f5424c8c38 /adamic-build/main.c:219:76
            #4 0x55fe9200c8b6 in adamic_initialize_object_a_6f36c7b5d0a9cb6b9e4c05117fed7d08 /adamic-build/main.c:549:2
            #5 0x55fe9200ba6d in adamic_initialize_object_a_e22c7958b07f44ff41b0e96fdd303f35 /adamic-build/main.c:559:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x55fe91fc83a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55fe92035fbf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55fe92035fbf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55fe9200c1a2 in adamic_function_object_a_profiler_86162185be5ce0c466ad01f5424c8c38 /adamic-build/main.c:102:6
            #4 0x55fe9200c1a2 in adamic_initialize_object_a_6f36c7b5d0a9cb6b9e4c05117fed7d08 /adamic-build/main.c:549:2
            #5 0x55fe9200ba6d in adamic_initialize_object_a_e22c7958b07f44ff41b0e96fdd303f35 /adamic-build/main.c:559:2
        
        SUMMARY: AddressSanitizer: 198 byte(s) leaked in 3 allocation(s).
=== RUN   TestCapturedUnionAgrees/object.a/counted
=== NAME  TestCapturedUnionAgrees/class.a/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==13378==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x5600281233a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5600281912ef in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5600281912ef in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x560028168c41 in adamic_function_class_a_profiler_enable_anonymous_1_ea226c37f4e68e91a097669521c25e5a /adamic-build/main.c:538:79
            #4 0x560028167547 in adamic_function_class_a_profiler_f2f42308525a0bc039a9f6471c4a29fe /adamic-build/main.c:163:76
            #5 0x560028167547 in adamic_initialize_class_a_0bb00bb098737f64bdf63cb15114d431 /adamic-build/main.c:627:2
            #6 0x560028166a6d in adamic_initialize_class_a_a6018fd220e74ffad83cf22f269eaeb8 /adamic-build/main.c:637:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x5600281233a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5600281912ef in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5600281912ef in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x560028168c41 in adamic_function_class_a_profiler_enable_anonymous_1_ea226c37f4e68e91a097669521c25e5a /adamic-build/main.c:538:79
            #4 0x5600281678b6 in adamic_function_class_a_profiler_f2f42308525a0bc039a9f6471c4a29fe /adamic-build/main.c:232:76
            #5 0x5600281678b6 in adamic_initialize_class_a_0bb00bb098737f64bdf63cb15114d431 /adamic-build/main.c:627:2
            #6 0x560028166a6d in adamic_initialize_class_a_a6018fd220e74ffad83cf22f269eaeb8 /adamic-build/main.c:637:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x5600281233a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5600281912ef in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5600281912ef in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5600281685d5 in adamic_function_class_a_profiler_write_anonymous_1_b5fd217a0bf4737a47ba776d789a7d3e /adamic-build/main.c:379:80
            #4 0x5600281671a2 in adamic_function_class_a_profiler_f2f42308525a0bc039a9f6471c4a29fe /adamic-build/main.c:115:6
            #5 0x5600281671a2 in adamic_initialize_class_a_0bb00bb098737f64bdf63cb15114d431 /adamic-build/main.c:627:2
            #6 0x560028166a6d in adamic_initialize_class_a_a6018fd220e74ffad83cf22f269eaeb8 /adamic-build/main.c:637:2
        
        SUMMARY: AddressSanitizer: 198 byte(s) leaked in 3 allocation(s).
=== RUN   TestCapturedUnionAgrees/class.a/counted
=== NAME  TestCapturedUnionAgrees/object.a/counted
    captured_union_test.go:110: counts: heap values leaked: 3 (allocations 21, frees 18, in regions 0)
    captured_union_test.go:116: counts: allocations 21, frees 18, in regions 0
=== RUN   TestCapturedUnionAgrees/object.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== NAME  TestCapturedUnionAgrees/graph.a/counted
    captured_union_test.go:110: counts: heap values leaked: 10 (allocations 42, frees 32, in regions 0)
    captured_union_test.go:116: counts: allocations 42, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== NAME  TestCapturedUnionAgrees/graph.ts/counted
    captured_union_test.go:110: counts: heap values leaked: 10 (allocations 42, frees 32, in regions 0)
    captured_union_test.go:116: counts: allocations 42, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== NAME  TestCapturedUnionAgrees/class.a/counted
    captured_union_test.go:110: counts: heap values leaked: 3 (allocations 21, frees 18, in regions 0)
    captured_union_test.go:116: counts: allocations 21, frees 18, in regions 0
=== RUN   TestCapturedUnionAgrees/class.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
--- FAIL: TestCapturedUnionAgrees (0.00s)
    --- FAIL: TestCapturedUnionAgrees/nested.ts (16.03s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/native (0.54s)
        --- FAIL: TestCapturedUnionAgrees/nested.ts/sanitized (3.27s)
        --- FAIL: TestCapturedUnionAgrees/nested.ts/counted (0.54s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/javascript (1.16s)
    --- FAIL: TestCapturedUnionAgrees/class.ts (16.11s)
        --- PASS: TestCapturedUnionAgrees/class.ts/native (1.50s)
        --- FAIL: TestCapturedUnionAgrees/class.ts/sanitized (1.35s)
        --- FAIL: TestCapturedUnionAgrees/class.ts/counted (0.57s)
        --- PASS: TestCapturedUnionAgrees/class.ts/javascript (2.04s)
    --- PASS: TestCapturedUnionAgrees/cycle.a (16.54s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/native (1.14s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/sanitized (1.28s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/counted (0.66s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/javascript (2.95s)
    --- PASS: TestCapturedUnionAgrees/cycle.ts (18.30s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/native (0.96s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/sanitized (1.39s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/counted (0.62s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/javascript (4.77s)
    --- PASS: TestCapturedUnionAgrees/alternating.ts (3.02s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/native (0.49s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/sanitized (1.11s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/counted (0.30s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/javascript (0.15s)
    --- FAIL: TestCapturedUnionAgrees/nested.a (5.04s)
        --- PASS: TestCapturedUnionAgrees/nested.a/native (0.71s)
        --- FAIL: TestCapturedUnionAgrees/nested.a/sanitized (2.30s)
        --- FAIL: TestCapturedUnionAgrees/nested.a/counted (0.44s)
        --- PASS: TestCapturedUnionAgrees/nested.a/javascript (0.14s)
    --- PASS: TestCapturedUnionAgrees/by_value.ts (3.70s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/native (0.81s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/sanitized (1.12s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/counted (0.51s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/javascript (0.92s)
    --- PASS: TestCapturedUnionAgrees/alternating.a (4.97s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/native (0.65s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/sanitized (1.48s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/counted (1.43s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/javascript (0.44s)
    --- PASS: TestCapturedUnionAgrees/released.ts (4.40s)
        --- PASS: TestCapturedUnionAgrees/released.ts/native (1.56s)
        --- PASS: TestCapturedUnionAgrees/released.ts/sanitized (1.06s)
        --- PASS: TestCapturedUnionAgrees/released.ts/counted (1.42s)
        --- PASS: TestCapturedUnionAgrees/released.ts/javascript (0.11s)
    --- PASS: TestCapturedUnionAgrees/by_value.a (2.83s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/native (0.41s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/sanitized (0.82s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/counted (0.64s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/javascript (0.16s)
    --- PASS: TestCapturedUnionAgrees/released.a (3.78s)
        --- PASS: TestCapturedUnionAgrees/released.a/native (0.77s)
        --- PASS: TestCapturedUnionAgrees/released.a/sanitized (1.66s)
        --- PASS: TestCapturedUnionAgrees/released.a/counted (0.52s)
        --- PASS: TestCapturedUnionAgrees/released.a/javascript (0.44s)
    --- FAIL: TestCapturedUnionAgrees/object.ts (2.56s)
        --- PASS: TestCapturedUnionAgrees/object.ts/native (0.62s)
        --- FAIL: TestCapturedUnionAgrees/object.ts/sanitized (1.12s)
        --- FAIL: TestCapturedUnionAgrees/object.ts/counted (0.48s)
        --- PASS: TestCapturedUnionAgrees/object.ts/javascript (0.11s)
    --- FAIL: TestCapturedUnionAgrees/object.a (2.68s)
        --- PASS: TestCapturedUnionAgrees/object.a/native (0.44s)
        --- FAIL: TestCapturedUnionAgrees/object.a/sanitized (0.52s)
        --- FAIL: TestCapturedUnionAgrees/object.a/counted (0.76s)
        --- PASS: TestCapturedUnionAgrees/object.a/javascript (0.45s)
    --- FAIL: TestCapturedUnionAgrees/graph.a (13.75s)
        --- PASS: TestCapturedUnionAgrees/graph.a/native (1.93s)
        --- FAIL: TestCapturedUnionAgrees/graph.a/sanitized (5.35s)
        --- FAIL: TestCapturedUnionAgrees/graph.a/counted (5.29s)
        --- PASS: TestCapturedUnionAgrees/graph.a/javascript (0.19s)
    --- FAIL: TestCapturedUnionAgrees/graph.ts (29.89s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/native (7.83s)
        --- FAIL: TestCapturedUnionAgrees/graph.ts/sanitized (7.25s)
        --- FAIL: TestCapturedUnionAgrees/graph.ts/counted (4.22s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/javascript (0.21s)
    --- FAIL: TestCapturedUnionAgrees/class.a (4.30s)
        --- PASS: TestCapturedUnionAgrees/class.a/native (0.49s)
        --- FAIL: TestCapturedUnionAgrees/class.a/sanitized (1.98s)
        --- FAIL: TestCapturedUnionAgrees/class.a/counted (0.92s)
        --- PASS: TestCapturedUnionAgrees/class.a/javascript (0.49s)
FAIL
gate cache: native hits=0 misses=32
gate cache: node hits=0 misses=32
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	30.417s
FAIL
```


## Step 4: full lowering suite

Command: `go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1`.

```text
lower exit=0 wall=3012
```


No failing tests; no environment failures; no main reruns required.

Complete lowering log:

```text
ok  	github.com/system-inc/adamic/internal/lower	3003.413s

```


## Step 4b: IR guards

Command: `go test ./internal/ir -count=1 -timeout 30m`. Includes TestCallTargetReadersLower and the IR field guards.

```text
ir exit=0 wall=250
```


```text
ok  	github.com/system-inc/adamic/internal/ir	198.471s
```


## Step 5

The whole internal/oracle suite was deliberately skipped, as requested.

## Step 6: admission delta and isolated timing reruns

Command (run alone, with nothing else heavy running):

```sh
go run ./cmd/adamic-admission-delta --base d43437079a7e7ca723168e67c5efb5bf7ebee216 --head 9d31ad142e9fc1cf22fe0565951e86ccf4af8679 --manifest /tmp/n7proof/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --workers 2 --json
```

Manifest generated with `python3 cloud/admission-corpus/manifest.py --sha 9d31ad142e9fc1cf22fe0565951e86ccf4af8679`. Full corpus, no budget, no filter, no shard. No full-oracle command was running alongside it.

```text
witnesses=2 fixtures=4293 gaps=105 review=20 fuzz=0 stage3=821
admission exit=1 wall=2016
```

Gate exit **3**, verdict **void**; the `go run` wrapper exits **1** after printing `exit status 3`. Wall time **2016 seconds**. This run does not establish an admission-gate pass. Its only void reason is `internal/oracle/testdata/captured_union/alternating.ts: native compile timeout`. No semantic disagreement was reported outside the gate’s ruled negative witnesses/runtime stops.

| Metric | Value |
| --- | --- |
| diff_count | 8 |
| admitted / newly accepted | 8 |
| newly refused | 0 |
| accepted by both | 3282 |
| refused by both | 1951 |
| classification stderr differs | 8 |
| classification compile limit | 45 seconds |
| runtime compile/run limit | 10 seconds |
| initial runtime timeouts | 5 |
| retried alone by gate | 5 |
| still timed out after retry | 1 |

All 8 newly admitted programs are the eight captured_union `.ts` pins. Seven agree with Node in the gate; alternating’s native build timed out, so its agreement is false and the run void. The `.a` pin coverage is in the dedicated pin tests above (the repository commits these fixtures as `.ts`).

```json
{
  "phase_seconds": {
    "base_build": 79.834569151,
    "checkout": 18.634792336,
    "checkout_and_provision": 20.184558531,
    "classification": 1065.533809368,
    "head_build": 34.891230833,
    "manifest_and_verification": 37.269292097,
    "provision": 1.5497605920000002,
    "runtime": 767.871000371,
    "total": 2005.603759479
  },
  "complete_corpus": true,
  "negative_witness_count": 50,
  "accepted_witnesses": 50,
  "node_agreements": 20,
  "runtime_notyet_count": 52,
  "ruled_runtime_stops": 52,
  "backend_divergence_count": 12
}
```

Gate stderr, verbatim:

```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 5 runs timed out, 5 retried alone, 1 timed out again
changed programs: 8
admission delta is void: internal/oracle/testdata/captured_union/alternating.ts: native compile timeout
exit status 3
```

### Every program with differing classification stderr

#### internal/oracle/testdata/captured_union/alternating.ts

Class: `newly-accepted`. Main compile exit 1; HEAD compile exit 0. Main stderr (verbatim):

```text
adamic: /tmp/adamic-gate/admission-delta-3652682472/head/internal/oracle/testdata/captured_union/alternating.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr (verbatim; empty):

```text

```

#### internal/oracle/testdata/captured_union/by_value.ts

Class: `newly-accepted`. Main compile exit 1; HEAD compile exit 0. Main stderr (verbatim):

```text
adamic: /tmp/adamic-gate/admission-delta-3652682472/head/internal/oracle/testdata/captured_union/by_value.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr (verbatim; empty):

```text

```

#### internal/oracle/testdata/captured_union/class.ts

Class: `newly-accepted`. Main compile exit 1; HEAD compile exit 0. Main stderr (verbatim):

```text
adamic: /tmp/adamic-gate/admission-delta-3652682472/head/internal/oracle/testdata/captured_union/class.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr (verbatim; empty):

```text

```

#### internal/oracle/testdata/captured_union/cycle.ts

Class: `newly-accepted`. Main compile exit 1; HEAD compile exit 0. Main stderr (verbatim):

```text
adamic: /tmp/adamic-gate/admission-delta-3652682472/head/internal/oracle/testdata/captured_union/cycle.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr (verbatim; empty):

```text

```

#### internal/oracle/testdata/captured_union/graph.ts

Class: `newly-accepted`. Main compile exit 1; HEAD compile exit 0. Main stderr (verbatim):

```text
adamic: /tmp/adamic-gate/admission-delta-3652682472/head/internal/oracle/testdata/captured_union/graph.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr (verbatim; empty):

```text

```

#### internal/oracle/testdata/captured_union/nested.ts

Class: `newly-accepted`. Main compile exit 1; HEAD compile exit 0. Main stderr (verbatim):

```text
adamic: /tmp/adamic-gate/admission-delta-3652682472/head/internal/oracle/testdata/captured_union/nested.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr (verbatim; empty):

```text

```

#### internal/oracle/testdata/captured_union/object.ts

Class: `newly-accepted`. Main compile exit 1; HEAD compile exit 0. Main stderr (verbatim):

```text
adamic: /tmp/adamic-gate/admission-delta-3652682472/head/internal/oracle/testdata/captured_union/object.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr (verbatim; empty):

```text

```

#### internal/oracle/testdata/captured_union/released.ts

Class: `newly-accepted`. Main compile exit 1; HEAD compile exit 0. Main stderr (verbatim):

```text
adamic: /tmp/adamic-gate/admission-delta-3652682472/head/internal/oracle/testdata/captured_union/released.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```

HEAD stderr (verbatim; empty):

```text

```

### Every gate timeout, including ones whose automatic retry passed

The gate’s JSON retains the final observation after retry. The `timed_out` labels record the first attempt; `retried` records its isolated retry. For the repaired witness, the main record’s native compile is the original witness, while the timed_out label specifically identifies the repair.

```json
{
  "path": "internal/oracle/testdata/captured_union/alternating.ts",
  "blob": "4b580eb148914fd952757e4b14deb1f083910dfc",
  "corpus": "diff",
  "class": "newly-accepted",
  "base_compile": {
    "wall_seconds": 0.741153646,
    "exit": 1,
    "stdout": "",
    "stderr": "adamic: /tmp/adamic-gate/admission-delta-3652682472/head/internal/oracle/testdata/captured_union/alternating.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet\n"
  },
  "compile": {
    "wall_seconds": 0.241775477,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.438673381,
    "exit": 0,
    "stdout": "session 0.0 with 2, stopping, undefined, session 0.3 with 2, stopping, undefined, session 0.6 with 2, stopping\nsession 1.0 with 2, stopping, undefined, session 1.3 with 2, stopping, undefined, session 1.6 with 2, stopping\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.115505646,
    "exit": 0,
    "stdout": "session 0.0 with 2, stopping, undefined, session 0.3 with 2, stopping, undefined, session 0.6 with 2, stopping\nsession 1.0 with 2, stopping, undefined, session 1.3 with 2, stopping, undefined, session 1.6 with 2, stopping\n",
    "stderr": ""
  },
  "native": {
    "wall_seconds": 0,
    "exit": -1,
    "stdout": "",
    "stderr": "",
    "error": "native compilation failed"
  },
  "javascript_compile": {
    "wall_seconds": 0.169379348,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 10.006182103,
    "exit": -1,
    "stdout": "",
    "stderr": "",
    "error": "timeout"
  },
  "agree": false,
  "negative_related": false,
  "negative_witness": false,
  "notyet_related": false,
  "ruled_runtime_stop": false,
  "ruled_backend_divergence": false,
  "sampled": true,
  "void_reason": "native compile timeout",
  "timed_out": [
    "native compile"
  ],
  "retried": true
}
```

```json
{
  "path": "internal/oracle/testdata/captured_union/by_value.ts",
  "blob": "b24ed53fa8a4d070d6c48e453b01ee04c5eadc56",
  "corpus": "diff",
  "class": "newly-accepted",
  "base_compile": {
    "wall_seconds": 0.860300063,
    "exit": 1,
    "stdout": "",
    "stderr": "adamic: /tmp/adamic-gate/admission-delta-3652682472/head/internal/oracle/testdata/captured_union/by_value.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet\n"
  },
  "compile": {
    "wall_seconds": 0.322899148,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.238782608,
    "exit": 0,
    "stdout": "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.088270321,
    "exit": 0,
    "stdout": "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n",
    "stderr": ""
  },
  "native": {
    "wall_seconds": 0.002393067,
    "exit": 0,
    "stdout": "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n",
    "stderr": ""
  },
  "javascript_compile": {
    "wall_seconds": 0.100320653,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 3.381754943,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "agree": true,
  "negative_related": false,
  "negative_witness": false,
  "notyet_related": false,
  "ruled_runtime_stop": false,
  "ruled_backend_divergence": false,
  "sampled": true,
  "timed_out": [
    "native compile"
  ],
  "retried": true
}
```

```json
{
  "path": "internal/oracle/testdata/captured_union/graph.ts",
  "blob": "cdd254124dd70aed8f3820833d50d3fbaae903e7",
  "corpus": "diff",
  "class": "newly-accepted",
  "base_compile": {
    "wall_seconds": 0.153238355,
    "exit": 1,
    "stdout": "",
    "stderr": "adamic: /tmp/adamic-gate/admission-delta-3652682472/head/internal/oracle/testdata/captured_union/graph.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet\n"
  },
  "compile": {
    "wall_seconds": 0.229990445,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.17881846,
    "exit": 0,
    "stdout": "held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.084603854,
    "exit": 0,
    "stdout": "held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n",
    "stderr": ""
  },
  "native": {
    "wall_seconds": 0.005214967,
    "exit": 0,
    "stdout": "held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n",
    "stderr": ""
  },
  "javascript_compile": {
    "wall_seconds": 0.079725589,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 3.182780629,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "agree": true,
  "negative_related": false,
  "negative_witness": false,
  "notyet_related": false,
  "ruled_runtime_stop": false,
  "ruled_backend_divergence": false,
  "sampled": true,
  "timed_out": [
    "native compile"
  ],
  "retried": true
}
```

```json
{
  "path": "internal/oracle/testdata/views_v3/values-lie.ts",
  "blob": "138cb57bfeda723b0f5bb672a1716a02241b07f2",
  "corpus": "fixtures",
  "class": "accepted-by-both",
  "base_compile": {
    "wall_seconds": 0.095822742,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "compile": {
    "wall_seconds": 0.094621948,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.19403717,
    "exit": 0,
    "stdout": "before\nbody\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.076982691,
    "exit": 70,
    "stdout": "before\n",
    "stderr": "adamic: panic: element read failed: view._children's element, in for...of is not a Node; expected Node, found number\n"
  },
  "native": {
    "wall_seconds": 0.003209509,
    "exit": 70,
    "stdout": "before\n",
    "stderr": "adamic: panic: element read failed: view._children's element, in for...of is not a Node; expected Node, found number\n"
  },
  "javascript_compile": {
    "wall_seconds": 0.141899314,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 3.502384873,
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
  "sampled": true,
  "timed_out": [
    "repair native compile"
  ],
  "retried": true
}
```

```json
{
  "path": "stage3/parser-next/speculation/misfit.a",
  "blob": "d9fa197599e9f241d9f22c1e394e2588241b193c",
  "corpus": "stage3",
  "class": "accepted-by-both",
  "base_compile": {
    "wall_seconds": 0.108386314,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "compile": {
    "wall_seconds": 0.130183984,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "node": {
    "wall_seconds": 0.590966156,
    "exit": 0,
    "stdout": "wrong\n",
    "stderr": ""
  },
  "javascript": {
    "wall_seconds": 0.073389111,
    "exit": 70,
    "stdout": "",
    "stderr": "adamic: panic: speculative result failed: callback misfit at /tmp/adamic-gate/admission-delta-3652682472/head/stage3/parser-next/speculation/misfit.a:3:14; expected number\n"
  },
  "native": {
    "wall_seconds": 0.018059156,
    "exit": 70,
    "stdout": "",
    "stderr": "adamic: panic: speculative result failed: callback misfit at /tmp/adamic-gate/admission-delta-3652682472/head/stage3/parser-next/speculation/misfit.a:3:14; expected number\n"
  },
  "javascript_compile": {
    "wall_seconds": 0.143839825,
    "exit": 0,
    "stdout": "",
    "stderr": ""
  },
  "native_compile": {
    "wall_seconds": 5.173656169,
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
  "sampled": true,
  "timed_out": [
    "native compile"
  ],
  "retried": true
}
```



### Explicit isolated reruns at both revisions

Reused the exact revision compiler products built by the gate, with its lowering adapter (the adapter only intercepts `admission-lower`; the ordinary `build` command is unchanged). HEAD product: `/home/agent/.cache/adamic-build/local/b1b2ca290b5459bfa6431dec1261d351c03549905e662580d448d8acdb90fb33/adamic`; main product: `/home/agent/.cache/adamic-build/local/4b4c513e940469f250466a24813d1ce9641cbd0a3da2f8ec5a9d264831bf452f/adamic`. Both consume the HEAD corpus, as the gate does. Verified their identities against class.ts: HEAD admission-lower exit 0, main exit 1 with the old captured-Union stop.

Each command ran alone and sequentially. Native compile command: `<revision-compiler> build <program> -o <external-unique-output>`. A monotonic clock measured wall seconds. Process groups were killed on timeout and reaped before continuing. Every successful compile was followed by a native execution at the gate’s **10-second run limit**. First compile limit: the gate’s **10 seconds**; second compile limit: **120 seconds**. There were no timeouts in these explicit reruns. “Not run” means compilation refused the program and no executable existed.

Included every path with an initial timeout, even when its gate retry passed. Also reran the exact values-lie repair whose native compile had timed out, plus the original witness for context. The repair was created outside the repository using the gate’s bytes replacement `_children:[7]` → `_children:[{kind:7}]`; its SHA-256 was asserted to be `ef4275f69c886d314514fcaa9ec53a253bf4c1ae4f29328deda45614203b83a0`.

| Program | Revision | Compile limit s | Compile exit | Compile wall s | Native run exit | Native wall s |
| --- | --- | --- | --- | --- | --- | --- |
| internal/oracle/testdata/captured_union/alternating.ts | HEAD | 10 | 0 | 9.166890 | 0 | 0.017568 |
| internal/oracle/testdata/captured_union/alternating.ts | main | 10 | 1 | 0.144430 | not run | — |
| internal/oracle/testdata/captured_union/alternating.ts | HEAD | 120 | 0 | 14.594753 | 0 | 0.005320 |
| internal/oracle/testdata/captured_union/alternating.ts | main | 120 | 1 | 0.117267 | not run | — |
| internal/oracle/testdata/captured_union/by_value.ts | HEAD | 10 | 0 | 3.488800 | 0 | 0.003188 |
| internal/oracle/testdata/captured_union/by_value.ts | main | 10 | 1 | 0.141476 | not run | — |
| internal/oracle/testdata/captured_union/by_value.ts | HEAD | 120 | 0 | 3.423448 | 0 | 0.006104 |
| internal/oracle/testdata/captured_union/by_value.ts | main | 120 | 1 | 0.112380 | not run | — |
| internal/oracle/testdata/captured_union/graph.ts | HEAD | 10 | 0 | 4.167592 | 0 | 0.004812 |
| internal/oracle/testdata/captured_union/graph.ts | main | 10 | 1 | 0.153594 | not run | — |
| internal/oracle/testdata/captured_union/graph.ts | HEAD | 120 | 0 | 3.690882 | 0 | 0.008430 |
| internal/oracle/testdata/captured_union/graph.ts | main | 120 | 1 | 0.151740 | not run | — |
| internal/oracle/testdata/views_v3/values-lie.ts | HEAD | 10 | 0 | 3.590788 | 70 | 0.006618 |
| internal/oracle/testdata/views_v3/values-lie.ts | main | 10 | 0 | 3.521618 | 70 | 0.004507 |
| internal/oracle/testdata/views_v3/values-lie.ts | HEAD | 120 | 0 | 3.290091 | 70 | 0.005220 |
| internal/oracle/testdata/views_v3/values-lie.ts | main | 120 | 0 | 3.576456 | 70 | 0.003146 |
| stage3/parser-next/speculation/misfit.a | HEAD | 10 | 0 | 4.119783 | 70 | 0.003360 |
| stage3/parser-next/speculation/misfit.a | main | 10 | 0 | 4.148597 | 70 | 0.003695 |
| stage3/parser-next/speculation/misfit.a | HEAD | 120 | 0 | 4.083444 | 70 | 0.010212 |
| stage3/parser-next/speculation/misfit.a | main | 120 | 0 | 4.871870 | 70 | 0.009118 |
| internal/oracle/testdata/views_v3/values-lie.ts [exact gate repair] | HEAD | 10 | 0 | 4.132230 | 0 | 0.007883 |
| internal/oracle/testdata/views_v3/values-lie.ts [exact gate repair] | main | 10 | 0 | 3.895350 | 0 | 0.003078 |
| internal/oracle/testdata/views_v3/values-lie.ts [exact gate repair] | HEAD | 120 | 0 | 4.078033 | 0 | 0.007251 |
| internal/oracle/testdata/views_v3/values-lie.ts [exact gate repair] | main | 120 | 0 | 4.160362 | 0 | 0.005559 |

Alternating’s isolated 10-second-limit compile passed in 9.166890 seconds, while its 120-second-limit compile took 14.594753 seconds. These separate measurements demonstrate timing variability across the gate’s 10-second cutoff; they do not turn the original void verdict into a pass. The captured-Union fixtures are refused at main, so there is no main native run or successful native compile time to report for them. The original values-lie and misfit witnesses stop with the expected ruled native exit 70 at both revisions; the repaired values-lie exits 0.

### Three isolated HEAD compiles: graph beside class

All use the 120-second compile limit and 10-second native run limit. Each compilation started only after the preceding compile and native run exited. All compile exits and all native run exits are 0. Normal compiler caches remained enabled, as in the gate; no concurrent workload was started.

| Repeat | graph.ts compile wall s | class.ts compile wall s |
| --- | --- | --- |
| 1 | 4.525909 | 4.364017 |
| 2 | 4.403546 | 4.438870 |
| 3 | 4.352784 | 5.044711 |

### Rerun diagnostics, verbatim

internal/oracle/testdata/captured_union/alternating.ts; main; compile limit 10 seconds.

Compile stderr:

```text
adamic: /workspace/adamic/internal/oracle/testdata/captured_union/alternating.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```

internal/oracle/testdata/captured_union/alternating.ts; main; compile limit 120 seconds.

Compile stderr:

```text
adamic: /workspace/adamic/internal/oracle/testdata/captured_union/alternating.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```

internal/oracle/testdata/captured_union/by_value.ts; main; compile limit 10 seconds.

Compile stderr:

```text
adamic: /workspace/adamic/internal/oracle/testdata/captured_union/by_value.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```

internal/oracle/testdata/captured_union/by_value.ts; main; compile limit 120 seconds.

Compile stderr:

```text
adamic: /workspace/adamic/internal/oracle/testdata/captured_union/by_value.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```

internal/oracle/testdata/captured_union/graph.ts; main; compile limit 10 seconds.

Compile stderr:

```text
adamic: /workspace/adamic/internal/oracle/testdata/captured_union/graph.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
```

internal/oracle/testdata/captured_union/graph.ts; main; compile limit 120 seconds.

Compile stderr:

```text
adamic: /workspace/adamic/internal/oracle/testdata/captured_union/graph.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
```

internal/oracle/testdata/views_v3/values-lie.ts; HEAD; compile limit 10 seconds.

Native stderr:

```text
adamic: panic: element read failed: view._children's element, in for...of is not a Node; expected Node, found number
```

internal/oracle/testdata/views_v3/values-lie.ts; main; compile limit 10 seconds.

Native stderr:

```text
adamic: panic: element read failed: view._children's element, in for...of is not a Node; expected Node, found number
```

internal/oracle/testdata/views_v3/values-lie.ts; HEAD; compile limit 120 seconds.

Native stderr:

```text
adamic: panic: element read failed: view._children's element, in for...of is not a Node; expected Node, found number
```

internal/oracle/testdata/views_v3/values-lie.ts; main; compile limit 120 seconds.

Native stderr:

```text
adamic: panic: element read failed: view._children's element, in for...of is not a Node; expected Node, found number
```

stage3/parser-next/speculation/misfit.a; HEAD; compile limit 10 seconds.

Native stderr:

```text
adamic: panic: speculative result failed: callback misfit at /workspace/adamic/stage3/parser-next/speculation/misfit.a:3:14; expected number
```

stage3/parser-next/speculation/misfit.a; main; compile limit 10 seconds.

Native stderr:

```text
adamic: panic: speculative result failed: callback misfit at /workspace/adamic/stage3/parser-next/speculation/misfit.a:3:14; expected number
```

stage3/parser-next/speculation/misfit.a; HEAD; compile limit 120 seconds.

Native stderr:

```text
adamic: panic: speculative result failed: callback misfit at /workspace/adamic/stage3/parser-next/speculation/misfit.a:3:14; expected number
```

stage3/parser-next/speculation/misfit.a; main; compile limit 120 seconds.

Native stderr:

```text
adamic: panic: speculative result failed: callback misfit at /workspace/adamic/stage3/parser-next/speculation/misfit.a:3:14; expected number
```
