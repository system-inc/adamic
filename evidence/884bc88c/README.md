# A–C at 884bc88c

Head: `884bc88cac4d47079039725e840351bfb29db7c2`.
Source worktree: /tmp/adamic-b3. Evidence branch: evidence/compiler/generic-method-view.

| Step | Exit | Result |
|---|---:|---|
| A gofmt -l internal/lower/ | 0 | Empty output. |
| A go vet ./internal/lower/ | 0 | Passed. |
| B combined generic, host-view and NodeWatch pins | 0 | All tests/subtests pass, including setMaxListeners.a and setMaxListeners.ts. |
| C remove ` && l.generic(source)`; host-view test | 1 | Both setMaxListeners rows turn red, exposing the fs.watch argument-form NotYet. |
| Restore byte-for-byte / Git diff | 0 / 0 | Exact; source worktree clean. |

Commands use /workspace/adamic-tools/env.sh (Go 1.27.1, clang 20.1.8 and Node 24.19.0):

```sh
gofmt -l internal/lower/
go vet ./internal/lower/
go test -count=1 -run 'TestAGenericMethod|TestAHostMethod|TestNodeWatch' -v ./internal/lower/
# Remove exactly " && l.generic(source)":
go test -count=1 -run '^TestAHostMethodAViewMakesGenericIsNotYet$' -v ./internal/lower/
# Restore byte-for-byte; verify whole-worktree Git diff.
```

## Exact baseline row diagnostics

The passing test asserts the diagnostic but does not print it. After restoring, the baseline compiler was built with `go build -o /tmp/adamic-884bc88c ./cmd/adamic` (exit 0). Each row's raw source was extracted unchanged from the committed Go test and saved as /tmp/adamic-884bc88c-row/main.a and main.ts (also retained as row-source.*). `adamic c <source>` uses the same loader/lowerer, and captures the diagnostic without editing the test or compiler. Both CLI runs exit 1 because the intended generic-method NotYet rejects the view; these CLI exits are separate from the passing Go test exit.

.a:
```text
adamic: /tmp/adamic-884bc88c-row/main.a:3:67: stage 0 can't lower a generic method setMaxListeners of type (n: number) => FSWatcher seen as <T>(n: string | number) => unknown yet
```

.ts:
```text
adamic: /tmp/adamic-884bc88c-row/main.ts:3:67: stage 0 can't lower a generic method setMaxListeners of type (n: number) => FSWatcher seen as <T>(n: string | number) => unknown yet
```

## Exact mutant row reports

```text
    generic_method_view_test.go:127: got "/tmp/adamic-gate/TestAHostMethodAViewMakesGenericIsNotYetsetMaxListeners.a427344919/001/main.a:2:17: stage 0 can't lower node:fs.watch outside fs.watch(path, options, listener) yet", want the generic method setMaxListeners named as NotYet
    generic_method_view_test.go:127: got "/tmp/adamic-gate/TestAHostMethodAViewMakesGenericIsNotYetsetMaxListeners.ts117251726/001/main.ts:2:17: stage 0 can't lower node:fs.watch outside fs.watch(path, options, listener) yet", want the generic method setMaxListeners named as NotYet
--- FAIL: TestAHostMethodAViewMakesGenericIsNotYet (0.00s)
    --- FAIL: TestAHostMethodAViewMakesGenericIsNotYet/setMaxListeners.a (0.38s)
    --- FAIL: TestAHostMethodAViewMakesGenericIsNotYet/setMaxListeners.ts (0.38s)
```

Thus the test is green at baseline and red when the guard is removed. Under the mutant, lowering reaches `fs.watch('.')`, whose one-argument form is not lowered. The mutant does not compile/run the bad setMaxListeners call; its red result is the loss of the required generic-method NotYet diagnostic.

Raw output, exact patch, all numeric exits and byte-for-byte restoration checks are adjacent. No permanent compiler/test edits. No whole-package or admission rerun was requested or performed. Only evidence/884bc88c/ is committed for this run.
