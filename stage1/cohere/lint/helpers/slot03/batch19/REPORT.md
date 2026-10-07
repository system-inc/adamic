Built expr, patternBind and switchStatement in three separate .a files; fifty-seven owned helpers total.
Commits: claim b3f1eb20 pushed before source; base main c7991b90 and lint area b4691483; final implementation/publication SHAs are named in the final response.
Commands: current baseline 8,354 comparisons; shared helpers PASS 90.073s; vet clean; six uncached probes PASS 1.845s; setup 58s, nproc 5; owned gate PASS 436.923s, all twenty-seven variants caught.
Mutants: all twenty-seven new variants compile and run cleanly before stdout comparisons catch them; three initial binding survivors are caught by the expanded actual-Go controls.
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

The owned test records the exact source replacement and each variant's first mismatching output line. evidence/helpers.log names every final variant and its witness. Source Node and emitted JavaScript are baseline comparisons; these semantic variants execute on sanitized native only. The following table describes every owned replacement. Prior batches' 203 variants were already green on this unchanged integration base; they are not claimed rerun by this selected gate.

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
