# A–C verification at 9b3544dd

Head: `9b3544dd01c6e1b177ba1d180f18374344a37ed9`.
Source worktree: /tmp/adamic-b3. Evidence branch: evidence/compiler/generic-method-view.

| Step | Exit | Result |
|---|---:|---|
| A gofmt -l internal/lower/ | 0 | Empty output. |
| A go vet ./internal/lower/ | 0 | Passed. |
| B combined generic, host-view, NodeWatch pins | 1 | All four host-view rows fail; other generic-method and NodeWatch tests pass. |
| C remove ` && l.generic(source)`; host-view test | 1 | All four rows red with the same namespace-object NotYet as baseline. Not a green-to-red proof. |
| Exact restoration / Git diff | 0 / 0 | Restored byte-for-byte; source worktree clean. |

Commands use /workspace/adamic-tools/env.sh (Go 1.27.1, clang 20.1.8, Node 24.19.0):

```sh
gofmt -l internal/lower/
go vet ./internal/lower/
go test -count=1 -run 'TestAGenericMethod|TestAHostMethod|TestNodeWatch' -v ./internal/lower/
# Remove exactly " && l.generic(source)" from invariance.go:
go test -count=1 -run '^TestAHostMethodAViewMakesGenericIsNotYet$' -v ./internal/lower/
# Restore exactly and check whole-worktree Git diff.
```

## Full baseline failures

```text
    generic_method_view_test.go:128: got "/tmp/adamic-gate/TestAHostMethodAViewMakesGenericIsNotYetchdir.a3603370150/001/main.a:2:59: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet", want the generic method chdir named as NotYet
    generic_method_view_test.go:128: got "/tmp/adamic-gate/TestAHostMethodAViewMakesGenericIsNotYetexit.ts947966648/001/main.ts:2:71: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet", want the generic method exit named as NotYet
    generic_method_view_test.go:128: got "/tmp/adamic-gate/TestAHostMethodAViewMakesGenericIsNotYetexit.a2998800741/001/main.a:2:71: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet", want the generic method exit named as NotYet
    generic_method_view_test.go:128: got "/tmp/adamic-gate/TestAHostMethodAViewMakesGenericIsNotYetchdir.ts2354600076/001/main.ts:2:59: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet", want the generic method chdir named as NotYet
--- FAIL: TestAHostMethodAViewMakesGenericIsNotYet (0.00s)
    --- FAIL: TestAHostMethodAViewMakesGenericIsNotYet/chdir.a (0.48s)
    --- FAIL: TestAHostMethodAViewMakesGenericIsNotYet/exit.ts (0.47s)
    --- FAIL: TestAHostMethodAViewMakesGenericIsNotYet/exit.a (0.41s)
    --- FAIL: TestAHostMethodAViewMakesGenericIsNotYet/chdir.ts (0.37s)
```

## Full mutant failures, each row

```text
    generic_method_view_test.go:128: got "/tmp/adamic-gate/TestAHostMethodAViewMakesGenericIsNotYetchdir.ts462767820/001/main.ts:2:59: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet", want the generic method chdir named as NotYet
    generic_method_view_test.go:128: got "/tmp/adamic-gate/TestAHostMethodAViewMakesGenericIsNotYetexit.ts134436323/001/main.ts:2:71: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet", want the generic method exit named as NotYet
    generic_method_view_test.go:128: got "/tmp/adamic-gate/TestAHostMethodAViewMakesGenericIsNotYetchdir.a1663314267/001/main.a:2:59: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet", want the generic method chdir named as NotYet
    generic_method_view_test.go:128: got "/tmp/adamic-gate/TestAHostMethodAViewMakesGenericIsNotYetexit.a3716606836/001/main.a:2:71: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet", want the generic method exit named as NotYet
--- FAIL: TestAHostMethodAViewMakesGenericIsNotYet (0.00s)
    --- FAIL: TestAHostMethodAViewMakesGenericIsNotYet/chdir.ts (0.61s)
    --- FAIL: TestAHostMethodAViewMakesGenericIsNotYet/exit.ts (0.61s)
    --- FAIL: TestAHostMethodAViewMakesGenericIsNotYet/chdir.a (0.63s)
    --- FAIL: TestAHostMethodAViewMakesGenericIsNotYet/exit.a (0.65s)
```

Both baseline and mutant reach the namespace object used-as-value refusal at the assignment from the imported process object. The explicit Node import resolves the prior missing-name error, but the tests still do not reach the expected generic-method NotYet. The red mutant therefore does not discriminate the source-generic guard. No test/compiler repair was made; only the requested mutant was applied and restored.

Raw logs, numeric exits, exact patch, full failure indexes and restoration records are adjacent. Only evidence/9b3544dd/ is committed for this run. No full-suite or admission rerun was requested or performed.
