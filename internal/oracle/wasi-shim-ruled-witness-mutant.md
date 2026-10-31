# The shim's ruled native witness, and its mutant

TestWASIShimAgreesWithNode holds a fixture's wasm32 build to a witness. A fixture with an entry in ruledBackendDivergences (backend_divergences_test.go) is witnessed by that entry's native outcome, through ruledNativeOutcome, since a wasm32 build is a native build. The only such fixture is internal/oracle/testdata/d96d304_try_stack.a:

- native, ruled and pinned by TestTerminalStackStopNative: exit 70, stdout empty, stderr `adamic: panic: RangeError: Maximum call stack size exceeded in depth\n`
- JavaScript, ruled and pinned by TestTerminalStackStopJavaScript: exit 70, stdout empty, stderr `adamic: panic: RangeError: Maximum call stack size exceeded\n`
- wasm32 under the shim and under node:wasi: exit 70, stdout empty, stderr `adamic: panic: RangeError: Maximum call stack size exceeded in depth\n`

## The mutant

ruledNativeOutcome reports no ruling, so the checked fixture's witness is the JavaScript backend again:

```
sed -i 's/\t\t\treturn entry.native, true/\t\t\treturn entry.native, false/' internal/oracle/backend_divergences_test.go
```

```
45c45
< 			return entry.native, true
---
> 			return entry.native, false
```

Run on Home (Linux), with ADAMIC_ORACLE_WASI=1 ADAMIC_TEST_WASI=1:

```
go test -p 1 -count=1 -timeout 30m ./internal/oracle -run 'TestWASIShimAgreesWithNode/internal/oracle/testdata/d96d304_try_stack.a' -v
```

It exits 1, red on the subtest:

```
    wasi_shim_test.go:157: shim vs expected witness: stderr differs
        witness: exit 70 stdout "" stderr "adamic: panic: RangeError: Maximum call stack size exceeded\n"
        shim: exit 70 stdout "" stderr "adamic: panic: RangeError: Maximum call stack size exceeded in depth\n"
    --- FAIL: TestWASIShimAgreesWithNode_105/internal/oracle/testdata/d96d304_try_stack.a (5.30s)
```

## The restore

The file is copied back from the copy saved before the mutant; `cmp` against that copy exits 0 and `git diff --quiet` on the file exits 0. The same command then exits 0:

```
    --- PASS: TestWASIShimAgreesWithNode_105/internal/oracle/testdata/d96d304_try_stack.a (1.87s)
ok  	github.com/system-inc/adamic/internal/oracle	1.985s
```

## The same witness in TestWASIAgreesWithNode and TestWasmtimeAgreesWithNode

Loom's run of -m11 (chg_7a6cdq6g, Eve) found that the node:wasi and wasmtime agreement tests held d96d304_try_stack.a's wasm32 build to the JavaScript backend too. Both now take a ruled divergence's native outcome through ruledNativeOutcome, as the shim does, so one list is read by every wasm backend.

Run on Home (Linux), with ADAMIC_ORACLE_WASI=1 ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASMTIME=1:

```
go test -p 1 -count=1 -timeout 30m ./internal/oracle -run 'Test(WASI|Wasmtime|WASIShim)AgreesWithNode_[0-9]+/internal/oracle/testdata/d96d304_try_stack.a$' -v
```

At the head, all three pass. With the same mutant (`return entry.native, false`), all three go red:

```
    --- FAIL: TestWasmtimeAgreesWithNode_105/internal/oracle/testdata/d96d304_try_stack.a
    wasi_wasmtime_test.go:214: wasmtime vs expected witness: stderr differs
        expected witness/known behavior: exit 70, stdout "", stderr "adamic: panic: RangeError: Maximum call stack size exceeded\n"
        wasmtime: exit 70, stdout "", stderr "adamic: panic: RangeError: Maximum call stack size exceeded in depth\n"
    --- FAIL: TestWASIAgreesWithNode_105/internal/oracle/testdata/d96d304_try_stack.a
    wasi_test.go:104: TestWASIAgreesWithNode_105/internal/oracle/testdata/d96d304_try_stack.a: stderr differs
        node: exit 70, stdout "", stderr "adamic: panic: RangeError: Maximum call stack size exceeded\n"
        wasi: exit 70, stdout "", stderr "adamic: panic: RangeError: Maximum call stack size exceeded in depth\n"
    --- FAIL: TestWASIShimAgreesWithNode_105/internal/oracle/testdata/d96d304_try_stack.a
```

The file was restored from its copy (`cmp` 0), and `git diff --quiet` held.
