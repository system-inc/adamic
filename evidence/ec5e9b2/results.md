# L2 at ec5e9b2

Exact tested commit: ec5e9b2bf4061eadee1b25df2e0354992594800e, detached worktree /workspace/adamic-rule-g-2594029. Only the requested L2 Go test and its condition-3 mutant were run. Source restored exactly (22-restoration.exit = 0).

## 1. Baseline

`go test -count=1 -run 'TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet' -v ./internal/lower/`

Exit 1: Load rejects Provider as Slot with TS2322, before the relation is exercised.

```text
=== RUN   TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== PAUSE TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== CONT  TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
    rule_g_test.go:71: Load: /tmp/adamic-gate/TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet3373051285/001/main.ts:7:7: error TS2322: Type 'Provider' is not assignable to type 'Slot'.
          Types of property 'run' are incompatible.
            Type '<T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: unknown[]) => void' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ...args: A) => void'.
              Types of parameters 'callback' and 'callback' are incompatible.
                Types of parameters 'args' and 'args' are incompatible.
                  Type 'any[]' is not assignable to type 'A'.
                    'any[]' is assignable to the constraint of type 'A', but 'A' could be instantiated with a different subtype of constraint 'unknown[]'.
--- FAIL: TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet (0.06s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.068s
FAIL

```

## 2. Mutant B

Exact requested signatureAny conditional replaced with `if false {`; see 20-mutant-b.patch. Same test command, exit 1 at the same loader error. It is already red at baseline, so this does not establish condition-3 sensitivity. Source restored exactly afterward.

```text
=== RUN   TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== PAUSE TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
=== CONT  TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet
    rule_g_test.go:71: Load: /tmp/adamic-gate/TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet2212002036/001/main.ts:7:7: error TS2322: Type 'Provider' is not assignable to type 'Slot'.
          Types of property 'run' are incompatible.
            Type '<T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: unknown[]) => void' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ...args: A) => void'.
              Types of parameters 'callback' and 'callback' are incompatible.
                Types of parameters 'args' and 'args' are incompatible.
                  Type 'any[]' is not assignable to type 'A'.
                    'any[]' is assignable to the constraint of type 'A', but 'A' could be instantiated with a different subtype of constraint 'unknown[]'.
--- FAIL: TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet (0.09s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.109s
FAIL

```

## 3. Stock tsc 6.0.3

No tsc command on PATH, but the stock Microsoft TypeScript checkout at 050880ce59e30b356b686bd3144efe24f875ebc8 contains tracked lib/tsc.js and lib/_tsc.js. Its tracked files are unmodified; `node /workspace/typescript-source/lib/tsc.js --version` prints `Version 6.0.3` (exit 0). L2.ts is extracted exactly from the Go test's raw source string.

An initial --noEmit --strict invocation stopped on TS5112 because the repository tsconfig is present. Both isolated checks below use --ignoreConfig to bypass that configuration.

`node /workspace/typescript-source/lib/tsc.js --ignoreConfig --noEmit L2.ts`

Exit 2: rejects the program.

```text
L2.ts(7,7): error TS2322: Type 'Provider' is not assignable to type 'Slot'.
  Types of property 'run' are incompatible.
    Type '<T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: unknown[]) => void' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ...args: A) => void'.
      Types of parameters 'callback' and 'callback' are incompatible.
        Types of parameters 'args' and 'args' are incompatible.
          Type 'any[]' is not assignable to type 'A'.
            'any[]' is assignable to the constraint of type 'A', but 'A' could be instantiated with a different subtype of constraint 'unknown[]'.

```

`node /workspace/typescript-source/lib/tsc.js --ignoreConfig --noEmit --strict L2.ts`

Exit 2: rejects the program with the same diagnostic.

```text
L2.ts(7,7): error TS2322: Type 'Provider' is not assignable to type 'Slot'.
  Types of property 'run' are incompatible.
    Type '<T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: unknown[]) => void' is not assignable to type '<A extends unknown[]>(callback: (...args: A) => void, ...args: A) => void'.
      Types of parameters 'callback' and 'callback' are incompatible.
        Types of parameters 'args' and 'args' are incompatible.
          Type 'any[]' is not assignable to type 'A'.
            'any[]' is assignable to the constraint of type 'A', but 'A' could be instantiated with a different subtype of constraint 'unknown[]'.

```
