# Empty symlink target on WASI

The shared WASI filesystem adapter translates EINVAL to ENOENT only when
symlink fails with an empty target. All other results and errors are preserved.
The combined host branch 08b5b2c4 includes this header before its symlink call,
so the fix applies when that landing is integrated without importing its other
compiler work into p2b.

Node 24.19.0 on Linux reports code ENOENT, errno -2, and
`ENOENT: no such file or directory, symlink '' -> 'adamic-empty-symlink-link'`.
The new source witness runs on Node. A direct C runtime fixture runs on native
and WASI, comparing code, Linux errno projection, and the full message against
that independent witness. p2b does not yet lower symlinkSync, so this is a
runtime fixture rather than an ordinary lowered fixture/counts row. It exercises
the real shared syscall adapter and error formatter. The numeric errno output
explicitly translates libc's symbolic error to libuv's Linux number; this change
does not add an errno property to the existing runtime Error object.

The mutant undefines the symlink adapter in the same WASI fixture. It compiles
and exits zero with no stderr, producing EINVAL, -22, and the corresponding
invalid-argument message. Only the Node output comparison catches it.

Commands (output sent directly to the accompanying logs):

```sh
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASIEmptySymlinkAgreesWithNode$' -count=1 -v -timeout 30m
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^(TestWASIFileAgreesWithNode|TestWASIInputAgreesWithNode|TestWASIHostRuntimeRefusals)$' -count=1 -v -timeout 30m
go test ./internal/native -count=1 -timeout 30m
go vet ./internal/oracle ./internal/native
git diff --check
```

Focused control and mutant test passed (0.969s); existing WASI host tests passed
(66.794s), including the existing explicit permission, timestamp and temporary
creation refusals. Both targets build the complete runtime archive. The full native package passed: ok  	github.com/system-inc/adamic/internal/native	205.647s. Vet and diff
checks passed. Whole repository and whole oracle gates were not rerun for this
small unit; the combined compiler branch was inspected, not rebuilt.

Setup with GOPROXY='https://proxy.golang.org|direct' and --wasi-sdk completed
in 52.107s: Node 0.038s, Go 0.048s, submodules 0.104s, markdown 0.114s,
clang 0.260s, WASI SDK 0.284s, Go build 51.799s, cache warm 52.054s.
Environment sourced: /workspace/adamic-tools/env.sh. nproc=5, CPU quota=4.
Node v24.19.0. npm ci --prefix stage3/api completed successfully.
