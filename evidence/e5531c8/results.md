# L2 proof at e5531c8

Exact tested commit: e5531c8c7fd9fa0b5076bb3d0d0961e6954e9ce1, detached worktree /workspace/adamic-rule-g-2594029. Only the requested L2 Go test, its condition-3 mutant, and stock tsc checks were run. Full commands, patches, output, and exit codes are retained beside this report.

## 1. Baseline: PASS, exit 0

`go test -count=1 -run 'TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet' -v ./internal/lower/`

```text
=== RUN   TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== PAUSE TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== CONT  TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
--- PASS: TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet (0.13s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.158s

```

## 2. Mutant B: FAIL, exit 1

Exact requested signatureAny conditional replaced with `if false {` (20-mutant-b.patch). Same test command. Baseline passes and this mutant newly fails, demonstrating condition-3 sensitivity.

```text
=== RUN   TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== PAUSE TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== CONT  TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
    rule_g_test.go:74: got <nil>, want the unfixed type parameter named as NotYet
--- FAIL: TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet (0.25s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.266s
FAIL

```

Restoration: diff against e5531c8 exits 0; final worktree status is empty. No compiler changes retained.

## 3. Stock tsc 6.0.3: accepts

Stock Microsoft TypeScript checkout: 050880ce59e30b356b686bd3144efe24f875ebc8. Tracked files are unmodified, lib/tsc.js and lib/_tsc.js are tracked, and the compiler reports Version 6.0.3. L2.ts is extracted exactly from the Go test's raw string.

Both commands exit 0 with empty output:

```sh
node /workspace/typescript-source/lib/tsc.js --ignoreConfig --noEmit L2.ts
node /workspace/typescript-source/lib/tsc.js --ignoreConfig --noEmit --strict L2.ts
```

--ignoreConfig isolates the file from the repository tsconfig. Default and strict settings both accept the program.
