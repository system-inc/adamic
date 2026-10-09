# Inheritance defense

Main: 7b9d4272c28f59530ab13daa5c49067e47933b06. All four requested target/subsumer names remain. No tests or harnesses were changed. Source restored after every mutant. All diffs apply to this main and pass go vet ./internal/lower/.

CODE UNDER TEST: Adamic Go lowering, including classGenericCall, nominalTypeArguments, classBases and the generic function instantiation they reach. ORACLE: nominal constraints uses handwritten Refused identity, nominal ancestry and interface repair assertions (self). Factory layouts now compares execution with Node (external-run) and asserts concrete Number and String base layouts in IR (self). The audit used older main 8171b317 and described acceptance/IR checks without external execution.

Clean whole-package baseline passed in 59.472 binary seconds, 274 top-level passes. Four per-test coverage runs passed: constraints 0.095, views 0.066, factory 0.203, monomorphizations 0.193 binary seconds. Each used -coverpkg ./internal/lower and an exact top-level -run with -coverprofile. Coverage-differences.json lists 373 blocks exclusive to constraints versus views and 194 exclusive to factory versus monomorphizations. Coverage is of the lowering code, not native runtime code.

Constraints: D1 drops the generic function's nominal-bound validation while keeping class instantiation validation. Only TestInheritanceGenericNominalConstraints fails, on its generic function input. The class-input subcase remains guarded by the other call to nominalTypeArguments. The views subsumer exercises a class-valued assignment rather than generic function constraint validation. D1 passed 273 other rows; their full names are in D1-results.json. This row is defended.

Factory: D2 drops the mapper/cache assignment in classGenericCall; all 274 rows pass. D3 reverses the mapper's source and target vectors; all 274 rows pass. Generic function instantiation can infer concrete types from its resolved return type, supplying another route for this fixture. Neither survivor has an independent changed-output witness here; treat them as equivalent candidates, not proof of unguarded behavior. D4 disables concrete base substitution in classBases. Factory fails on indexed T["native"], but its subsumer and three other rows fail too. Three honest attempts did not establish a unique catch. This is not defended, not a recommendation to delete. Its name promises concrete generic factory layouts, and its assertions check those layouts and successful execution. No name/assertion mismatch was found.

Every production matrix ran the entire current package under timeout 120, go test -json -count=1 -timeout 90s -run ., with its own ADAMIC_BUILD_CACHE_DIR=/tmp/inheritance-defense/cache/ID. No run cooked or panicked. The outer wall can exceed 90 seconds because it includes Go compilation. Complete failing, passing and skipped row lists appear in matrix.json and individual results files. TestOriginalCycleLedger and TestOptionalWideningCensus skipped in the baseline and all matrices; uniqueness excludes these unavailable corpus/configuration rows. No other package suite was run.

Friction and costs: /tmp has only 8.8 GB total, making the requested 15 GB free threshold impossible. Earlier /tmp/library-defense scratch was removed; afterward /tmp had 5.2 GB free and /workspace 14 GB free on separate filesystems. No disk-full failure occurred. Audit line numbers and oracle descriptions differ from current main; all defense diff references use starting main. No other ambiguity blocked the task. Warm toolchain worked, setup skipped; npm ci stage3/api completed before baseline. nproc 5. D2 and D3 survival required the third factory attempt. Neither target is a cost row or an executor twin.

Matrix wall seconds (including Go build):
- D1: 96.953 wall, 81.003 binary; 1 failing rows.
- D2: 96.583 wall, 82.168 binary; 0 failing rows.
- D3: 50.769 wall, 39.699 binary; 0 failing rows.
- D4: 47.444 wall, 39.745 binary; 5 failing rows.

Publication: the requested remote branch already contained another inheritance defense. This session evidence is isolated under factory-constraints-7b9d4272; existing root evidence and branch history are preserved. The non-fast-forward rejection cost a fetch, path separation and merge. Mutant IDs in this session are local to its subdirectory.
