Built: simple assignment expressions save the RHS once, store it, and yield its value and proofs across all supported value contexts.
Commits: base 4885cec50290686df487b62aac47c85d871ed40c; witness 91c49dcd7259e6eeeb5b4beb3dee6ac48e888bd5; delivery codex/assignment-proofs, tip reported with the push.
Commands/results: focused lower and Node oracle tests, counts update/check, vet and diff check passed; logs below contain exact output.
Mutants: proof-origin, rhs-twice, result-retain, temporary-cleanup, reference-order, effects-writes and alias-proof each failed with its intended catcher.
Not covered: full upstream scanner execution, destructuring or separately represented assignment targets, compound/logical assignment expansion, performance measurement, whole-package tests or the full gate.

This delivers assignment-value lowering toward roadmap step 12 (#cvhj5fk) and step 23 ruling 5. The branch is based directly on origin/compiler/area-next-fixtures; no other worker branch was merged.

**Implementation and hooks**

`internal/lower/assignment_value.go:11` dispatches simple assignment into the common value path. At `:46`, assignmentRight follows simple assignment RHS origins for enum, narrowing and widening analysis. At `:54` and `:61`, hidden declarations save the member receiver and element index before the RHS, then save the RHS, perform the existing validated store, and yield the saved value. The order is ECMA-262 Evaluate AssignmentExpression: evaluate the left reference, GetValue of the RHS, PutValue, return that RHS value. A setter does not run a getter to determine the expression result. The store's Box/Maybe/Weak adapter is removed only when its inner value has the expression's native type; RHS checks remain attached to their single evaluation. Absent results keep the existing destination representation.

`internal/lower/statements.go:195` routes assignment returns through the same expression lowering, removing the separate local-only return implementation. `internal/lower/enum_never.go:11`, `:29`, `:63` and `internal/lower/enum_flags.go:102` use RHS origins. `internal/lower/invariance.go:524` preserves RHS allocation provenance while refusing incompatible mutable aliases: storing a fresh allocation creates an alias and does not prove uniqueness.

`internal/native/taste.go:101` scopes Effects bindings, saves/retains its yielded reference before cleanup, releases internal bindings on each evaluation, and returns the saved result. Throw paths use the existing scope cleanup. `internal/lower/closed_frame_inputs.go:47` admits the Effects container to the existing recursive walker; stores within it remain inspected. A pinned test and mutant prove that internal field stores still invalidate the closed-input proof.

The conservative scope assumption is that this unit extends simple `=` assignments represented as Assign, SetProperty or SetIndex. Unsupported target representations still refuse loudly. No changes were made to emit.go, lower.go, native.go or oracle_test.go.

**Fixtures and correctness**

Seven new `.a` fixtures cover scanner's `return token = keyword`, moduleSpecifiers' assignment in a condition without a hoisting adaptation, chained assignments, RHS side effects, saved reference results across later arguments, assignment in a while condition, receiver/index order, setter semantics, property returns, undefined returns, narrowed presence, checked casts, fresh allocation identity and throwing RHS cleanup. They pass byte for byte against stock Node through both JavaScript and native backends. Native builds run with clang `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`; Linux leak checks rerun with leak detection enabled. Release builds use clang `-O2` without sanitizers. Oracle caching was disabled with ADAMIC_GATE_UNCACHED=1 for all comparisons and mutants.

| Fixture | Allocations | Frees | Retains | Releases | Peak |
|---|---:|---:|---:|---:|---:|
| assignment_scanner | 17 | 17 | 18 | 30 | 5 |
| assignment_module_specifiers | 7 | 7 | 6 | 13 | 3 |
| assignment_chain | 13 | 13 | 10 | 26 | 7 |
| assignment_while | 14 | 14 | 26 | 43 | 3 |
| assignment_order | 22 | 22 | 25 | 49 | 9 |
| assignment_proofs | 17 | 17 | 13 | 30 | 8 |
| assignment_throw | 10 | 10 | 10 | 19 | 5 |

The counts table was regenerated with the required command and checked without update. Additional Effects retains/releases move existing rows; their affected fixtures were held to Node separately. The generator also moves an existing logical-and row to its current inventory position and removes the obsolete binder-flow row that the current inventory does not count.

**Mutants**

All seven final mutants return test exit 1. The runner also asserts each stated catcher and rejects a clang compilation failure as evidence. Its absolute workspace paths describe this run; change root/d when replaying elsewhere.

| Mutant | Fixture or pinned test | Catcher |
|---|---|---|
| Replace assignment RHS proof origin with target origin | assignment_scanner | Lower refuses at 16:24, unproven enum-member-slot value, adamic/enum-literal |
| Evaluate a direct RHS call once extra before saving it | assignment_chain | Node says `1 1 1 1 1` and `heap-2 2`; both mutated backends say `2 2 2 2 2` and `heap-4 4` |
| Omit yielded reference retain | assignment_order | AddressSanitizer heap-use-after-free |
| Omit Effects binding releases | assignment_while | LeakSanitizer detects leaked allocations |
| Evaluate member RHS before its receiver | assignment_order | Node stdout differs |
| Skip inspecting an Effects body in closed-input analysis | TestAssignmentProofsClosedInputStillInspectsEffects | Internal field store hidden by Effects |
| Drop mutable-alias widening guard | TestAssignmentProofsDoNotInventMembersOrUniqueAliases | Expected refusal becomes nil |

The initial duplicate-RHS mutation re-emitted nested Effects declarations and failed C compilation. That version was discarded; the final mutation duplicates leaf calls and is caught by the side-effect counter. The initial counts attempt also exposed the missing pinned @types/node install and an Effects allowlist regression; npm ci installed the lock-pinned dependencies and the recursive-container fix resolved the regression. Final update and verification both passed.

**Native TypeScript scanner first stop**

The unadapted scanner driver imports the existing TypeScript checkout at 050880ce59e30b356b686bd3144efe24f875ebc8. Its normal diagnostic information source was generated with `node scripts/processDiagnosticMessages.mjs src/compiler/diagnosticMessages.json`; no compiler source was copied or rewritten. Compiler CLIs were built separately, with a Go overlay restoring all seven changed production files to the area baseline for the before run.

| Probe | Before | After |
|---|---|---|
| Native C for stock scanner driver | exit 1: binder.ts:1109:17 TS2412, undefined not assignable to FlowNode under exactOptionalPropertyTypes | Identical first stop and diagnostics, exit 1 |
| Reduced scanner assignment | C emitted, exit 0 | C emitted, exit 0; Node oracle passed |
| assignment_order property return | Refuses returning property or element assignment at 21:38 | C emitted, exit 0; Node oracle passed |
| assignment_throw member value | Refuses assignment value to member at 5:19 | C emitted, exit 0; Node oracle passed |

Observation: the historical scanner assignment stop is already absent on this area's tip. No claim is made that this patch removed it or that the full scanner executes natively; the full driver stops in checking before lowering. The witness and before/after results are preserved in logs/compilation-stops.json.

**Exact validation commands**

Run from the repo with `source /workspace/adamic-tools/env.sh`. Every command's output was redirected to a log, never piped.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /workspace/scratch/assignment-proofs/setup.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestAssignmentProofs|TestClosedFrameInputRejectsMutation|TestNumericEnum(LiteralPromises|NeverProof)|TestNumericEnumsAreOpen|TestNativeAgreesWithNode/(internal|stage3)/(oracle|fixtures)/(testdata|nested-functions)/(assignment_|taste_comma|host_optional_intrinsic|nested_assignment_return|taste_optional_join|06_parser_token_state)' -count=1 -v -timeout=10m > /workspace/scratch/assignment-proofs/final-focused.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/(internal|stage3)/(oracle|fixtures)/(testdata|taste)/(library_string_conversion|library_string_raw|method_coverage_conversions|taste_void|taste_stage3_representations|09_literal_cache|13_void_callback|14_relative_complement|22_assignment_once)' -count=1 -v -timeout=10m > /workspace/scratch/assignment-proofs/effects-regression.log 2>&1
go test ./internal/lower -run TestAssignmentProofs -count=1 > /workspace/scratch/assignment-proofs/lower-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout=30m -args -update-counts > /workspace/scratch/assignment-proofs/counts-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout=30m > /workspace/scratch/assignment-proofs/counts-check.log 2>&1
go vet ./internal/lower ./internal/native ./internal/oracle > /workspace/scratch/assignment-proofs/vet.log 2>&1
python3 /workspace/scratch/assignment-proofs/run-mutants.py > /workspace/scratch/assignment-proofs/mutants.log 2>&1
git diff --check > /workspace/scratch/assignment-proofs/diff-check.log 2>&1
```

Focused lower and oracle packages reported PASS (1.319s and 2.815s), the nine additional Effects fixtures passed (9.131s), counts update passed (46.411s), and counts verification passed (33.556s). Vet and diff check returned 0. These are test elapsed times, not performance measurements. No whole-package tests or full gate were run.

Toolchain setup succeeded: Node ready 0.111s; Go 0.140s; Markdown 0.279s (validated install skipped, step 0.018s); clang 0.554s; submodules 3.164s; Go build 204.431s; test binaries deferred 204.618s; cache warm 204.619s; done 204.647s. `nproc` is 5, cgroup cpu.max is 400000/100000, memory 17.6 GB. Versions: Go 1.27.1, Node 24.19.0, clang 20.1.8. No Go module 403 occurred. Setup's environment path was /workspace/adamic-tools/env.sh.
