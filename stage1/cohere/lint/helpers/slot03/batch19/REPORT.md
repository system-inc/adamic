Built expr, patternBind and switchStatement in three separate .a files; fifty-seven owned helpers total.
Commits: original claim b3f1eb20 pushed before source (rebased fa6fb15c), implementation ce3b63bc; lint area d3a37422 contains current fetched main b6b1538b; final publication SHA is named in the final response.
Commands: full rebased helper gate PASS owned 1,264.599s/shared 91.887s, 3,054,680 comparisons; selected shared harness PASS 290.864s; vet/format clean; six uncached probes PASS 9.502s; setup 58s, nproc 5.
Mutants: all 230 owned variants (twenty-seven new), four inherited variants and the missing-consumer check caught after rebase; semantic variants finish cleanly and fail stdout comparison, the empty-stack guard variant fails its exit comparison.
Not covered: complete rule findings, independent recursive/graph/jump backends, arbitrary malformed ASTs or AST-mutating callbacks, and the full repository gate including its seventeen external comparisons.

# Ownership, landing and readiness

The own branch's fifty-four previous helpers were complete, oracle-green and pushed at 72ed70d1 before this reservation. That run covered 3,046,326 comparisons and all 203 owned variants against lint area b4691483, containing current main c7991b90. A new wildcard fetch found both integration bases unchanged; both are ancestors of this branch. Only codex/lint-helpers-03 belongs to this worker. All twenty origin codex/lint-helpers* branches and nineteen claim files were read before selecting these three helpers, and claim b3f1eb20 was pushed before source. Each ties the highest eligible remaining fan-out at four consumers. Comments remain reserved by their shared bundle; regexp-engine internals are excluded by the JS RegExp instruction. No further helpers are reserved.

Territory: claims/03.md, slot03/batch19 and batch19_test.go. Shared registry, harness, compiler, runtime and actual cohere worktree files are unchanged. Temporary oracle overlays rename Go methods and call their unchanged bodies, and add observation-only files. All new Adamic files have .a extensions. No rule dispatch or numeric AST kinds are introduced.

Every helper removes one dependency for each of these rules:

- array-callback-return
- consistent-return
- no-unreachable-loop
- react-hooks/rules-of-hooks

Twelve dependency edges across four rules; zero final helper blockers removed by this batch alone. readiness.json subtracts only this batch from the frozen 198-rule inventory and lists remaining dependencies. These counts describe helper readiness; they do not assert complete findings parity.

# Actual Go observations

The pinned Go cohere oracle is 715ba94f3608a6500086b1076ce5cb7e51b836db. Regeneration captures all four rules' upstream runtime fixture strings, including dynamically assembled inputs. The upstream core package passes in 6.790s and React in 3.773s. Coverage contains 2,119 deduplicated sources and an asserted exact consumer set. Fixture capture still executes each package's real tests.

For every captured source, the real parser indexes roots and the real CFG builder invokes the helpers. Wrappers execute original Go dependencies and record their direct operations, arguments, return value, current block before/after and jump-stack broken flags. Nested dependency operations are suppressed from the outer helper trace. The source/native/emitted driver generates calls independently while replaying dependency results and state updates. It compares the complete operation trace and final cursor byte for byte. This proves composition; it does not prove the replayed dependency implementations.

An expression observer hook records the exact position Go calls its optional hook without changing graph semantics. Read/write consumer payloads are outside this observer. Controls also cover a nil expression hook. Dependencies and callbacks must preserve the immutable parsed view; arbitrary callback mutation of AST metadata is outside this API contract.

| Consumer | expr | patternBind | switchStatement |
|---|---:|---:|---:|
| array-callback-return | 828 | 80 | 5 |
| consistent-return | 296 | 5 | 0 |
| no-unreachable-loop | 4,905 | 696 | 168 |
| react-hooks/rules-of-hooks | 1,006 | 67 | 2 |

Observed 8,058 live calls and 296 controls, 8,354 comparisons. A zero is a lack of observed reach in that captured corpus, not proof that the rule never reaches the helper. Controls exercise nil nodes/lists, optional hooks, all expression categories, wrapper/default/computed/object/array bindings, empty switches, default before/after cases, abrupt and labelled exits, fallthrough and cumulative cycle barriers. A direct per-node control walk exposes cases normally nested within another helper's direct-call trace.

A temporary AST overlay adds an oracle-only method that changes only a shallow-copied binding node's data to a typed nil *BindingElement. Go's original guard accepts that value and emits no operations. The parsed node and original helper body remain unchanged. Flattened parked fields behind the missing value make ignoring that guard observable. This control does not turn an unrelated Go node type into accepted metadata; Go's type assertion still rejects a wrong type.

# Commands and observations

Every test writes stdout/stderr directly to a log file. Source /workspace/adamic-tools/env.sh first. Native comparisons and variants use ASan/UBSan; successful execution and empty stderr are required before a stdout difference counts as a semantic-mutant catch. A compilation refusal, panic or sanitizer failure is not counted.

```
bash cloud/setup.sh > /tmp/slot03-batch19-setup.log 2>&1
python3 stage1/cohere/lint/helpers/slot03/batch19/testdata/regenerate.py > stage1/cohere/lint/helpers/slot03/batch19/evidence/regeneration.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch19' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch19/evidence/helpers.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot03/batch19/evidence/vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch19/evidence/input-oracle.log 2>&1
go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch19/evidence/shared-helpers.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestConsumerCoverageRejectsMutant$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch19/evidence/coverage-mutant.log 2>&1
```

Setup succeeds: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 58s, done 58s; nproc 5, cgroup cpu.max 400000 100000, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Vet has an empty log. Six uncached input probes pass in 1.845s, zero hits and six misses. Shared helpers pass in 90.073s, including all four compiling variants, ten message refusals and explicit gap/recovery checks. The missing-consumer guard catches removal of better-tailwindcss/no-unknown-classes.

The first owned run fails in 362.776s: 8,124 baseline comparisons pass, but three of its twenty-six variants survive. Missing typed binding metadata, assignment-property binding, and equals binding were nested or absent from that corpus. The failure is preserved in initial-mutant-survivors.log. Actual-Go controls were added and the same three variants retained; no check was relaxed or removed. A twenty-seventh variant proves that an absent optional hook is not called. The expanded baseline passes 8,354 four-way comparisons in 67.18s. The complete selected gate passes in 436.923s; all twenty-seven variants are caught, no skips.

# Every mutant

The owned test records the exact source replacement and each variant's first mismatching output line. evidence/helpers.log names every final variant and its witness. Source Node and emitted JavaScript are baseline comparisons; these semantic variants execute on sanitized native only. The following table describes every owned replacement. Prior batches' 203 variants were already green before reservation; all are rerun in the full rebased landing gate recorded below.

| Variant | Replacement |
|---|---|
| expression.a | `if (!view.present)` -> `if (false)` |
| expression.a#01 | `if (view.hookPresent)` -> `if (false)` |
| expression.a#02 | `if (view.hookPresent)` -> `if (true)` |
| expression.a#03 | `dependencies.read(view.node);` -> `(omit)` |
| expression.a#04 | `if (dependencies.throwable(view.node))` -> `if (false)` |
| expression.a#05 | `dependencies.binaryExpression(view.node);` -> `dependencies.expr(view.node);` |
| expression.a#06 | `if (view.update)` -> `if (false)` |
| expression.a#07 | `dependencies.expr(view.operand); dependencies.makeYield();` -> `dependencies.makeYield(); dependencies.expr(view.operand);` |
| expression.a#08 | `if (dependencies.isRoot(view.node))` -> `if (false)` |
| expression.a#09 | `dependencies.visitUnknown(child);` -> `dependencies.visitUnknown(view.node);` |
| pattern_bind.a | `if (!view.present)` -> `if (false)` |
| pattern_bind.a#01 | `dependencies.write(view.node);` -> `(omit)` |
| pattern_bind.a#02 | `if (!view.elementsPresent)` -> `if (false)` |
| pattern_bind.a#03 | `if (!element.present)` -> `if (false)` |
| pattern_bind.a#04 | `if (element.computed)` -> `if (false)` |
| pattern_bind.a#05 | `dependencies.bindWithDefault(element.target, element.fallback);` -> `dependencies.bindWithDefault(element.fallback, element.target);` |
| pattern_bind.a#06 | `dependencies.patternBind(property.target); break; /           case 'shorthand':` -> `dependencies.patternBind(property.fallback); break; /           case 'shorthand':` |
| pattern_bind.a#07 | `if (view.equals)` -> `if (false)` |
| switch_statement.a | `dependencies.expr(view.expression);` -> `(omit)` |
| switch_statement.a#01 | `if (clause.isDefault) { defaultIndex = i; continue; }` -> `if (false) { defaultIndex = i; continue; }` |
| switch_statement.a#02 | `dependencies.linkWithCycleBarrier(testCur, bodies[defaultIndex] ?? -1, false)` -> `dependencies.linkWithCycleBarrier(testCur, after, false)` |
| switch_statement.a#03 | `const barrier = hadSwitchBreak \|\| beforeFirstCase;` -> `const barrier = false;` |
| switch_statement.a#04 | `hadSwitchBreak = state.jumps[switchJump] ?? false;` -> `hadSwitchBreak = false;` |
| switch_statement.a#05 | `dependencies.statements(clause.statements);` -> `(omit)` |
| switch_statement.a#06 | `dependencies.popJump();` -> `(omit)` |
| switch_statement.a#07 | `dependencies.enter(after); return;` -> `return;` |
| switch_statement.a#08 | `dependencies.linkWithCycleBarrier(fallthroughFrom, bodies[i] ?? -1, barrier);` -> `dependencies.linkWithCycleBarrier(-1, bodies[i] ?? -1, barrier);` |

The four inherited variants permit a raw JSON control character, let oneOf accept overlapping branches, ignore an unknown strict-target field, and omit policy interpolation. Go acceptance/rendered-output comparisons catch all four. The independent coverage mutant removes an expected consumer and is caught by exact-set validation.

# Limits

There is no whole-rule finding/fix/suggestion comparison or integration into a complete native CFG builder. Recursive walkers, real AST adapters, graph allocation/storage/reachability and jump-frame semantics remain backend dependencies. Cases are bounded to the captured corpora and controls. Raw views must obey the named parsed contracts and immutable-callback precondition. Arbitrary malformed metadata, nil lists where Go assumes a valid parsed list, and wrong-kind type assertions are outside the successful-input coverage.

This worker used the bounded touched-package/filtered external oracle gate. The full repository gate, including the seventeen required external correctness comparisons, was not run. No external check was skipped, relaxed or deleted. All previously owned helpers were already green and pushed before new claims; their code is unchanged in this unit.

A later fetch advanced main to b6b1538b and the lint area to d3a37422, which contains that main. The completed selected-gate unit is committed before rebase; the full owned helper gate will be repeated against those bases for landing. No new helpers are claimed.

# Current-main landing

The completed selected-gate implementation accef463 rebased cleanly as ce3b63bc, and original claim b3f1eb20 rebased as fa6fb15c. All sixty-six own commits rebased onto current lint area d3a37422, containing current main b6b1538b. This accepts main's typeof-null changes unchanged. The full fifty-seven-helper gate is rerun on this compiler/runtime; no earlier green result is substituted for that rerun. Vet, six uncached input probes and six selected .a/emitted-JavaScript/suggestion/profile harness tests also run on the new base.

A claim race appeared in the final twenty-branch scan. At reservation, slot04 was 4982546a and its claim file contained neither expr nor patternBind. Slot03 original claim b3f1eb20 is timestamped 2026-10-07T09:11:17Z and was pushed before source. Slot04 later claimed both symbols in 3ab03bc3, timestamped 2026-10-07T09:13:10Z, one minute fifty-three seconds later. Slot03 retains its prior published reservation; the concurrent overlap is recorded in ownership.json for integration. No helper was claimed despite an existing earlier claim.

```
git rebase origin/area/stage1-lint
go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=30m > stage1/cohere/lint/helpers/slot03/batch19/evidence/rebased-helpers.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot03/batch19/evidence/rebased-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch19/evidence/rebased-input-oracle.log 2>&1
go test ./stage1/cohere/lint -run '^(TestEmittedJavaScriptMismatch|TestDotARename|TestCompleteSuggestionSerialization|TestSuggestionAlongsideAutomaticFix|TestWitnessScriptKind|TestProfileCompilation)$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch19/evidence/rebased-harness.log 2>&1
```

Rebased auxiliary results: shared helper package PASS 91.887s, all six selected shared-harness checks PASS 290.864s, six uncached input probes PASS 9.502s with zero hits and six misses, and repository vet output is empty. The profile test reports 139 allocations and 139 frees. Its current-base wire output is 652 bytes, compared to the preceding 654-byte historical run, with Go, source, emitted JavaScript and native still identical. The emitted-JavaScript mismatch test intentionally invokes a failing child comparison; the parent passes only when it catches that mutant. Raw child FAIL text is not an unhandled gate failure.

Final current-base landing result: all fifty-seven owned helpers PASS 1,264.599s, 3,054,680 comparison lines, 229 owned compiling semantic variants plus one owned empty-stack guard variant, four inherited semantic variants and the independent missing-consumer mutant. No selected helper tests skip or fail. evidence/rebased-mutant-witnesses.log names every rerun variant and preserves its first mismatch or guard result; earlier batch reports and tests specify all prior replacements. All 8,354 new cases agree across real Go, source Node, emitted JavaScript and sanitized native. Earlier batches compare Go/source/native, and later ones also emitted JavaScript; the complete line total is not claimed four-way for every earlier batch. Shared harness, vet, formatting and six uncached input probes are green as recorded above. The final publication targets only codex/lint-helpers-03, with an exact lease on the previously pushed claim b3f1eb20. No further helper is claimed.
