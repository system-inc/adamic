# Node library activation (#2nznnt3)

The fixtures exercise the production checker: a lazy `require("fs")`, import
types without static imports, and a program with no Node use. The tests also
check that user package names, relative names, nonliteral arguments, property
calls, unknown `node:` names and `undici-types` do not rebuild with the Node root.
The held Go overlay mutant disables the require branch of the actual hook;
`TestNodeLibraryLazyRequire` must then fail with TS2591.

Validation on Linux, Go 1.27.1, `GOMAXPROCS=4`, cgroup quota four CPUs:

| Test | Seconds |
|---|---:|
| TestNodeLibraryLazyRequire | 1.08 |
| TestNodeLibraryImportTypes | 2.31 |
| TestNodeLibraryActivationStaysNarrow | 0.46 |
| TestNodeLibraryBuiltinSpecifiers | 7.58 |
| TestNodeLibraryRequireMutant | 20.27 |

`timeout 90 go test ./internal/load -count=1 -json -timeout=80s` passed;
package time 22.721 seconds. All tests in that run were under 60 seconds
(maximum 20.27). `go build ./...`, `go vet ./internal/load`, and
`TestEveryTestIsParallelOrSaysWhy` passed (parallel rule: 1.56 seconds).
Every new test calls `t.Parallel()` first.

## Stage 3 measurement

The adapted tree was built using evidence revision `5fe8c05a`'s `stage3/apply.sh`,
locked API dependencies, and TypeScript source pin
`050880ce59e30b356b686bd3144efe24f875ebc8`. All 79 root hashes match its accepted
wave-2 evidence. No trigger root, source repair, checker-option change or prelude
change is included. The before binary uses main's original `node_library.go`
through a Go overlay; the after binary uses the hook change. Both invoke
ordinary `adamic types` on the same 79 roots.

[stage3.json](stage3.json) retains commands, binary and root hashes, complete
parsed diagnostics, log hashes, and exact removed/added diagnostics.

| Observation | Before | After |
|---|---:|---:|
| All diagnostics | 289 | 229 |
| Scoped Node/V8 rows | 61 | 0 |
| source-map-support row | 1 | 1 |

All 61 expected identities disappear. One previously hidden diagnostic becomes
visible with Node types enabled; it also appears in the original control. The
task owner accepts this as satisfying no new rows in substance and will route
the exposed source contract to Correctness separately:

```text
src/compiler/sys.ts:1708:56: error TS2345: Argument of type 'string | undefined' is not assignable to parameter of type 'string | ArrayBufferView<ArrayBufferLike>'.
  Type 'undefined' is not assignable to type 'string | ArrayBufferView<ArrayBufferLike>'.
```

This is `D705b73717985`, also present in the evidence revision's existing
80-root `import "node:fs"` control. All 229 after identities exactly match
that original control (zero identity differences). The optional result of
`JSON.stringify(cleanupPaths(profile))` is rejected by Node's `writeFileSync`
signature once the inspector callback receives its Node types. Fixing that
source contract or changing the prelude would expand this hook-only task.
No such change is included. source-map-support remains untouched.
