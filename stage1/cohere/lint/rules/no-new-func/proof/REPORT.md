# Wave 06 shared facts landing

Baseline: `d845dccde413c89643293e808626344d12e3f023` on origin/lint-checker/facts. Branch: lint-rules/facts-wave-06. No rebase, private checker, bridge edit, shared helper edit, harness edit or additional claim.

## Registered rule

`no-new-func` reproduces Go's direct Function calls/construction and static call/apply/bind invocations, including parentheses and local shadows. The descriptor listens only to NewExpression and CallExpression and receives the existing node. It uses the shared same-node node-symbol-details question and SymbolDetails decoder. The first ordered declaration's declaration-file flag reproduces resolvesToAGlobal; it is not replaced with a default-library-only predicate.

Upstream ProgramReads is zero (no_new_func.go:81-146); descriptor programReads is exactly empty. ProgramReads covers explicit ctx.Program reads (rule/rule.go:323-337). This port uses no askFile or foreign-file reads. Its oracle adapter returns the unmodified upstream rule; the rule has no options type. UpstreamTest is TestNoNewFunc, which covers every actual TestNoNewFunc* name in core/no_new_func_test.go.

All 41 unique captured upstream cases passed TestRulesAgree across live Go, sanitized native, source Node and emitted JavaScript replay. `upstream-captured.json` retains the owned capture; `case-counts.json` distinguishes captured counts from complete comparison. The firing witness also covers a local Function shadow. No automatic fixes or suggestions are invented.

## Parked rules and remaining dependencies

`parked-no-throw-literal/BLOCKED.md` records typed recovery metadata being dropped for `function f() { throw; }` before Go comparison. `parked-prefer-arrow-callback/BLOCKED.md` records the shared fix engine refusing an insertion at a removal endpoint. Both have source snapshots, pending descriptors and exact reproducers; neither is registered. No full parity or caught mutant is claimed for either. BLOCKED.md accounts for every original claim, including three rules already completed by other workers and the still missing CFG, flow, reference-tracker, alias/export and React HIR contracts.

## Commands and evidence

Required setup: setup.log, 42.126 seconds, nproc 5 and quota 4 cores. Registry generation: registry.log. Vet: vet.log. Full package command and required inputs: command.json. Output is written directly to gate.jsonl, never piped. Load samples: load.json. Final counts, elapsed wall and load summary: summary.json. Owned mutant comparison lines: mutant-lines.json.

The all-input package run uses -timeout=3h. The baseline's TestCheckerBridgeRefusalPending skips until the unlanded TSGoError library support exists; that is a named shared prerequisite, not a missing input or a relaxed check. Final results follow.

## Final results

Package exit 0. Including subtests: 119 pass, 0 fail, 1 skip. Top-level: 34 pass, 0 fail, 1 skip. The only skip is TestCheckerBridgeRefusalPending, awaiting codex/tsgo-errors-as-values; no required corpus or profile input was absent. Wall 2162.428 seconds, nproc 5, quota 4 cores; one-minute load minimum/median/maximum 0.215/1.208/4.834. TypeScript corpus checkout remained clean.

TestOwnedWitnesses, TestRulesAgree and TestMutants passed. The no-new-func shadow mutant compiled, ran successfully and differed from Go on native (gate.jsonl:4096), Node (4104) and emitted JavaScript (4112). Native comparisons used the shared sanitizer build. The captured upstream case count and complete match count are both 41.

The package's syntax-only 77-file benchmark records best native 15.029627 seconds versus Go 1.787867 seconds at gate.jsonl:1895-1897; it excludes typed rules without a program and is not a per-rule no-new-func measurement. The live typed comparison timings remain in gate.jsonl.

The parked throw rule completed only five upstream comparisons before typed recovery failed; the callback comparison remains incomplete. Their mutant files are preserved as pending inputs, not claimed caught. Missing analyses and unsupported question contracts are named in BLOCKED.md. Only this rule directory is committed.
