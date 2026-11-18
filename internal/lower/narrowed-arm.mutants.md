# narrowedArmReadGap's mutants

Each edit to `narrowed_nullish.go` ran alone on a Linux cloud worker through a Go overlay.
The unmodified arm pins passed in both .a and .ts.

| Mutant | Edit | Red test and failure |
|---|---|---|
| 1. Rule unregistered | Delete the 3350 registration. | `TestNarrowedArmReadOfACallWrittenLetIsNotYet` (all four probes in both dialects), `narrowed_arm_read_test.go:69`, got nil where the read should stop. `TestRefusalRulesRunInOrder`, `refusal_rules_test.go:120`, misses 3350. |
| 2. Cross-function gate removed | Drop `\|\| !l.writtenAcrossFunctions(symbol)` from the arm rule. | `TestNarrowedArmReadOfAnUnwrittenLetLowers` (both dialects), `narrowed_arm_read_test.go:83`, stops at `main.a:7:20` and `main.ts:7:20` instead of lowering. |
| 3. Exemption widened to all storage | Replace the representation test with `if declared != nil {` and drop the unused IR import. The type identity comparison has already dereferenced `declared`. | `TestNarrowedArmReadOfACallWrittenLetIsNotYet` (both original programs and both other probes in both dialects), `narrowed_arm_read_test.go:69`, got nil where the read should stop. |

The first three runs exited 1. Mutant 1 ran the full targeted selection before the never-read
fixtures were updated to pin the two rules. Mutant 2 selected the lowering control, and mutant 3
selected the refusal probes.

## 4. The never rule unregistered

Delete the 3340 registration and run the three never-read tests below. The run exited 1.
Each failing read still stops at the same location with the fallback words
`a read the checker narrowed to never across code that may write it`. The tests require
`a read the checker narrowed to never of a binding a call may write`.

| Red test | Rows | Failure line |
|---|---|---|
| `TestNeverReadOfACallWrittenLetIsNotYet` | nv2, nv6, nv8, nv9 | `narrowed_never_test.go:28` |
| `TestNeverReadAcrossAWriteIsNotYet` | nv2 | `never_read_test.go:104` |
| `TestNeverReadOwnFunctionEdges` | arrow, increment, decrement in both dialects | `never_read_test.go:230` |
