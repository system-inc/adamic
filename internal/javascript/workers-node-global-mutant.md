# The Workers Node global check, and its mutant

Three tests refuse a file a Worker ships when it uses a Node global Workers lack (`node:`, `process.` or `Buffer.`):
- TestWorkersCalleeWithoutNodeGlobals, in internal/javascript/workers_test.go;
- TestWorkerRuntimePreservesUTF8AndRefusesIO, in internal/worker/worker_test.go;
- every TestWasmWorkers unit, in internal/worker/wasm_test.go.

Each had its own substring match. Slice 5's adamicPlaceholderNarrow (internal/javascript/readiness.go) is emitted into every module and reads `ArrayBuffer.isView(value)`, a standard global Workers has. A substring match takes its `Buffer.` for Node's. -m12 fixed only internal/javascript's copy, and Loom's run of it went red on internal/worker's two.

There is now one matcher, `javascript.NodeGlobalUse` (internal/javascript/node_globals.go). It matches each spelling only where it starts an identifier, through the regexp `(^|[^A-Za-z0-9_$])(node:|process\.|Buffer\.)`. All three tests call it. internal/worker's two do so through workersUnavailable (internal/worker/unavailable_test.go), which keeps their own `eval(` and `new Function(` checks.

TestNodeGlobalUseMatchesOnlyNodeGlobals pins five strings:
- `Buffer.from(text)`, `(process.env.HOME)` and `import { readFileSync } from "node:fs";` match;
- `!ArrayBuffer.isView(value)` and `adamicProcess.exit(0)` don't.

## The mutant

The boundary is dropped, so the matcher is a substring match again:

```
sed -i 's/(^|\[^A-Za-z0-9_\$\])(node:/()(node:/' internal/javascript/node_globals.go
```

```
-var nodeGlobalSpelling = regexp.MustCompile(`(^|[^A-Za-z0-9_$])(node:|process\.|Buffer\.)`)
+var nodeGlobalSpelling = regexp.MustCompile(`()(node:|process\.|Buffer\.)`)
```

Run on Cloud (Linux):

```
go test -p 1 -count=1 -timeout 20m ./internal/javascript -run 'TestWorkersCalleeWithoutNodeGlobals|TestNodeGlobalUseMatchesOnlyNodeGlobals' -v
```

It exits 1, red on the pin and on both Workers subtests:

```
    workers_test.go:21: NodeGlobalUse("!ArrayBuffer.isView(value)") = "Buffer.", want ""
--- FAIL: TestNodeGlobalUseMatchesOnlyNodeGlobals (0.00s)
    workers_test.go:38: Workers module uses unavailable Buffer.
    workers_test.go:38: Workers module uses unavailable Buffer.
--- FAIL: TestWorkersCalleeWithoutNodeGlobals (0.00s)
    --- FAIL: TestWorkersCalleeWithoutNodeGlobals/argument_count (0.00s)
    --- FAIL: TestWorkersCalleeWithoutNodeGlobals/ordinary (0.00s)
```

internal/worker, with its products built first (`-run '^TestProduct_Worker(Compiler|Checks)$'`, then `-run '^TestProduct_'`, as docs/workers.md says). With the same mutant:

```
go test -p 1 -count=1 -timeout 20m ./internal/worker -run 'TestWorkerRuntimePreservesUTF8AndRefusesIO|TestWasmWorkers_004' -v
```

It exits 1:

```
    worker_test.go:279: handler.mjs uses Workers-unavailable Buffer.
--- FAIL: TestWorkerRuntimePreservesUTF8AndRefusesIO (0.08s)
```

TestWasmWorkers_004 fails on the same run with "Worker product worker wasm words not built ahead". That isn't the matcher: the mutant changes internal/javascript, so the Wasm Worker products' keys change and no product is built for them. Only TestWorkerRuntimePreservesUTF8AndRefusesIO is evidence for internal/worker's call. After the restore, both pass:

```
--- PASS: TestWorkerRuntimePreservesUTF8AndRefusesIO (0.08s)
--- PASS: TestWasmWorkers_004 (0.02s)
ok  	github.com/system-inc/adamic/internal/worker	2.615s
```

internal/worker whole is green on this tree.

## The restore

The file is copied back from the copy saved before the mutant. `cmp` against that copy exits 0, and `git diff --quiet` exits 0. Then the same internal/javascript command passes:

```
--- PASS: TestNodeGlobalUseMatchesOnlyNodeGlobals (0.00s)
--- PASS: TestWorkersCalleeWithoutNodeGlobals (0.00s)
    --- PASS: TestWorkersCalleeWithoutNodeGlobals/ordinary (0.03s)
    --- PASS: TestWorkersCalleeWithoutNodeGlobals/argument_count (0.03s)
ok  	github.com/system-inc/adamic/internal/javascript	0.034s
```
