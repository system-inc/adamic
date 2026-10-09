All three requested rows are defended by a package-unique production mutant.
Main: b8bcadb2c493173855f19d7e5c508b34f5eeb5b6. No test or oracle edits.
Evidence: standalone D1/D2/D3 diffs, full JSON logs, matrix.json, coverage profiles and coverage-differences.json.

Code under test and oracles

October and SharedSlice: internal/fuzz generator, GenerateFeatures/GenerateWithout, regexStatement, allShareCuts/templateShareCut and Program.Source. Oracles are self-written vocabulary and geometry assertions. SharedSlice checks the fuzzer's model, not actual runtime storage sharing.
RuntimeLibrary: internal/fuzz Prepare and checkout execution adapter. Oracles are native.RuntimeLibrary's sanitized product path, source execution on Node versus native/JavaScript backend, and the test's fixed stdout/exit expectations. No native.RuntimeLibrary or oracle code was mutated.

Coverage leads and semantic defenses

October versus SharedSlice covers 20 exclusive source lines; SharedSlice versus October covers two; RuntimeLibrary versus Judge covers 63. Line sets expand executed Go coverage blocks and can include closing lines; these are leads rather than instruction-level measurements.
D1 uses shared regexStatement lines: October explicitly requires /a/g.exec in 200 seeds while SharedSlice checks only sharing vocabulary. Changing the emitted literal to /b/g changes generated behavior and preserves checker/lower acceptance. This proves protection of the specific vocabulary literal, not the necessity of that spelling or absence of regex coverage generally.
D2 uses shared templateShareCut lines: shifting start 96 to 97 shortens the 48-unit template slice to 47, with its actual UTF-8 byte length adjusted from 96 to 94. It remains a valid sharing slice but no longer spans the row's minimum position-index checkpoint length. October generates the scene but does not check this geometry.
D3 changes Prepare's library option from sanitized to unsanitized. Only RuntimeLibrary compares the cached product identity; Judge supplies synthetic outcomes and never calls Prepare. The program subcase still passes, but the library subcase fails.

All 18 current top-level rows ran to completion for every mutant, without panic, timeout, or skip. Each mutant's 17 passing row names are in matrix.json. New rows since the audit are TestBytesSharedRejects63Bytes, TestExactSignatureRejectsSuffix, and TestReduceRejectsDifferentRefusal. The subprocess helper passing without its trigger is not counted as additional substantive coverage.
Every standalone diff applies to this main and passed go vet ./internal/fuzz/. Mutated source restored after each run. No test deleted or weakened.

Timing and friction

Warm toolchain worked, nproc=5, setup skipped. npm ci reported 689ms. Clean baseline binary time 22.767s. Four isolated coverage invocations passed; their commands and wall durations are in coverage-runs.json. Three mutant runs plus vet took 25.77s, 25.89s and 25.77s, totaling 77.43s; binary durations are in the logs. No step exceeded the budget.
Initial free space was /tmp 4.9GB and /workspace 14GB. Removed earlier /tmp/lint-deletion scratch/cache, then checked again: /tmp 5.1GB and /workspace 14GB. /tmp's entire filesystem is only 8.8GB, so the requested 15GB threshold cannot be reached there. No disk error occurred; repository and tools were untouched.
Evidence initially landed beneath stage3/api due to that command's working directory, then was moved to the required repository evidence path before commit. No production consequences.
The original subsumption rested on one broad mutant per row; narrower assertions yielded unique catches on the first defense attempt for each. No undefended row remains to assess a name/assertion mismatch. No repo-wide uniqueness or C-runtime coverage is claimed.
