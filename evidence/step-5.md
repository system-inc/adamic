# Step 5: full lower suite

Revision: d643b42469bd229d461ca025652bf0dc360ee675.

`ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source go test -count=1 -timeout 40m -skip '^TestProduct_' ./internal/lower/`

Pinned TypeScript: 050880ce59e30b356b686bd3144efe24f875ebc8.

Exit 1 after 1751.026 seconds of test execution. Three top-level failures, four failing extension subtests. No other failures reported; no timeout. The complete output, including every failure in full:

```text
--- FAIL: TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet (0.47s)
    rule_g_test.go:90: got /tmp/adamic-gate/TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet3506413439/001/main.ts:5:27: stage 0 can't lower a generic function value without a concrete contextual signature yet, want the unfixed type parameter named as NotYet
--- FAIL: TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers (0.00s)
    --- FAIL: TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers/.a (0.09s)
        rule_g_test.go:67: Load: /tmp/adamic-gate/TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers.a2199723128/001/main.a:6:7: error TS2322: Type 'StringHandler' is not assignable to type 'Handler'.
              Types of property 'handle' are incompatible.
                Type '<T extends string>(x: Animal, t: T) => string' is not assignable to type '<T>(x: Animal, t: T) => string'.
                  Types of parameters 't' and 't' are incompatible.
                    Type 'T' is not assignable to type 'string'.
    --- FAIL: TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers/.ts (0.12s)
        rule_g_test.go:67: Load: /tmp/adamic-gate/TestAGenericMethodWhoseNarrowerConstraintIsReadNeverLowers.ts1336176277/001/main.ts:6:7: error TS2322: Type 'StringHandler' is not assignable to type 'Handler'.
              Types of property 'handle' are incompatible.
                Type '<T extends string>(x: Animal, t: T) => string' is not assignable to type '<T>(x: Animal, t: T) => string'.
                  Types of parameters 't' and 't' are incompatible.
                    Type 'T' is not assignable to type 'string'.
--- FAIL: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot (0.00s)
    --- FAIL: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.a (0.11s)
        rule_g_test.go:22: Load: /tmp/adamic-gate/TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot.a2369395955/001/main.a:4:26: error TS18004: No value exists in scope for the shorthand property 'setTimeout'. Either declare one or provide an initializer.
    --- FAIL: TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot/.ts (0.05s)
        rule_g_test.go:22: Load: /tmp/adamic-gate/TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot.ts2185062375/001/main.ts:4:26: error TS18004: No value exists in scope for the shorthand property 'setTimeout'. Either declare one or provide an initializer.
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	1751.026s
FAIL
```
