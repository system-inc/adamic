# The Workers module's Node global check, and its mutant

TestWorkersCalleeWithoutNodeGlobals (internal/javascript/workers_test.go) refuses an emitted module that uses a Node global Workers lacks: `node:`, `process.` or `Buffer.`. Main ac43d4220a added the check as a substring match. Slice 5's adamicPlaceholderNarrow (internal/javascript/readiness.go), emitted into every module, reads `ArrayBuffer.isView(value)`, a standard global Workers has, and the substring match took its `Buffer.` for Node's.

nodeGlobalUse now matches each spelling only where it starts an identifier, through the regexp `(^|[^A-Za-z0-9_$])(node:|process\.|Buffer\.)`. TestNodeGlobalUseMatchesOnlyNodeGlobals pins five strings: `Buffer.from(text)`, `(process.env.HOME)` and `import { readFileSync } from "node:fs";` match, and `!ArrayBuffer.isView(value)` and `adamicProcess.exit(0)` don't.

## The mutant

The boundary is dropped, so the matcher is a substring match again, as strings.Contains was:

```
sed -i 's/(^|\[^A-Za-z0-9_\$\])(node:/()(node:/' internal/javascript/workers_test.go
```

```
14c14
< var nodeGlobalSpelling = regexp.MustCompile(`(^|[^A-Za-z0-9_$])(node:|process\.|Buffer\.)`)
---
> var nodeGlobalSpelling = regexp.MustCompile(`()(node:|process\.|Buffer\.)`)
```

Run on Home (Linux):

```
go test -p 1 -count=1 -timeout 30m ./internal/javascript -run 'TestWorkersCalleeWithoutNodeGlobals|TestNodeGlobalUseMatchesOnlyNodeGlobals' -v
```

It exits 1, red on the pin and on both Workers subtests:

```
    workers_test.go:33: nodeGlobalUse("!ArrayBuffer.isView(value)") = "Buffer.", want ""
--- FAIL: TestNodeGlobalUseMatchesOnlyNodeGlobals (0.00s)
    workers_test.go:50: Workers module uses unavailable Buffer.
    workers_test.go:50: Workers module uses unavailable Buffer.
--- FAIL: TestWorkersCalleeWithoutNodeGlobals (0.00s)
    --- FAIL: TestWorkersCalleeWithoutNodeGlobals/ordinary (0.00s)
    --- FAIL: TestWorkersCalleeWithoutNodeGlobals/argument_count (0.00s)
```

## The restore

The file is copied back from the copy saved before the mutant; `cmp` against that copy exits 0 and `git diff --quiet` on the file exits 0. The same command then exits 0:

```
--- PASS: TestNodeGlobalUseMatchesOnlyNodeGlobals (0.00s)
--- PASS: TestWorkersCalleeWithoutNodeGlobals (0.00s)
    --- PASS: TestWorkersCalleeWithoutNodeGlobals/argument_count (0.04s)
    --- PASS: TestWorkersCalleeWithoutNodeGlobals/ordinary (0.04s)
ok  	github.com/system-inc/adamic/internal/javascript	0.043s
```
