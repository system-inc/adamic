# Wave 19 shared checker facts port

Base: `origin/lint-checker/facts`, `d845dccde413c89643293e808626344d12e3f023`.
Branch: `lint-rules/facts-wave-19`. Only this rule directory is changed.

## Claimed rules

| Rule | Result on this facts slice |
| --- | --- |
| @typescript-eslint/no-array-constructor | Ported using the shared node-symbol-details question. Validation evidence below. |
| @typescript-eslint/consistent-generic-constructors | Blocked: `ctx.Program.Options().IsolatedDeclarations.IsTrue()`, cohere/internal/lint/rules/typescript/consistent_generic_constructors.go:240. See [reproducer](blocked/typescript-consistent-generic-constructors.md). |
| @typescript-eslint/dot-notation | Blocked: `ctx.Program.Options().NoPropertyAccessFromIndexSignature.IsTrue()`, cohere/internal/lint/rules/typescript/dot_notation.go:88; ordered property modifiers and index infos also remain absent. See [reproducer](blocked/typescript-dot-notation.md). |
| nexus/correctness-no-process-exit-after-output | Blocked: `writers.ctx.TypeChecker.GetResolvedSignature(call)` and `signature.Declaration()`, cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:314,318. See [reproducer](blocked/nexus-correctness-no-process-exit-after-output.md). |
| nexus/correctness-require-blocking-standard-streams | Blocked: `program.ResolveModule(importer, specifier)` and `program.GetSourceFileForResolvedModule(resolved.ResolvedFileName)`, cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams.go:279,283. See [reproducer](blocked/nexus-correctness-require-blocking-standard-streams.md). |
| require-await | Blocked: `checker.Checker_getResolvedSignature(ctx.TypeChecker, call, nil, checker.CheckModeNormal)`, `resolved.Target()` and `declared.TypeParameters()`, cohere/internal/lint/rules/core/require_await.go:988,992,993. See [reproducer](blocked/require-await.md). |
| nexus/correctness-no-uncleared-race-timeout | Already ported by wave 1 per integration instructions; not ported again. |
| symbol-description | Already on the facts base; not ported again. |
| valid-typeof | Already on the facts base; not ported again. |

## Port contract

`rule.json` declares typed, node-based listeners for CallExpression and NewExpression. The rule consumes the node handed to it and scans its provided position for the token start without fetching the root again. Child kinds distinguish callee and spread syntax after listener dispatch.

Cohere declares NeedsTypeChecker and TypeReachShapes, but no ProgramReads for NoArrayConstructor. Accordingly `programReads` is empty. The port consumes the first ordered symbol declaration's declaration-file flag, exactly as cohere's resolvesToAGlobal helper does; it neither asks whether that file is the default library nor opens another file. SymbolDetails is the existing shared decoder, not a private checker or helper copy.

The adapter selects core.NoArrayConstructor. This upstream rule has no options type; its adapter returns nil, and none of its rows carries options. The upstreamTest prefix TestNoArrayConstructor captures all nine real upstream test functions, including suggestion cases. Messages are copied verbatim. Automatic fixes remain empty and suggestions preserve argument trivia and byte positions.

## Evidence

The initial focused run matched 79 upstream cases across Node, emitted JavaScript and sanitized native, passed both owned witnesses, and caught the compiling local-shadow mutant on all three runtimes: [initial summary](evidence/focused-summary.json), [initial log](evidence/focused.jsonl), [initial mutant lines](evidence/mutant.log). These predate the final two equivalent speed edits, which removed a redundant optional-token lookup and a root refetch.

The first full run encountered the shared compiler-corpus native scanner's internal ten-minute execution timeout despite the outer -timeout 3h: [failure events](evidence/corpus-timeout.jsonl). The memory cgroup recorded zero OOM events. This is incomplete corpus comparison, not evidence of a rule mismatch. No timeout, skip, harness or shared checker was changed. Reproduce with ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript and go test -count=1 -timeout 3h ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$', redirecting output to a log.

Workspace restarts terminated two tracked runs and one final-source focused attempt. Their partial output does not certify a completed package run. The runner records every required lint input, nproc, wall time and one-minute load samples. TypeScript is pinned to 050880ce59e30b356b686bd3144efe24f875ebc8.

The shared TestCheckerBridgeRefusalPending test structurally skips when TSGoError is absent from the prelude; there is no input to supply to enable that check. The skip is reported rather than removed or relaxed.

The compiler-corpus test supplies bare source rows, not a program header. Typed rules therefore do not receive a checker in that aggregate test. This port is certified on its upstream cases and typed witnesses, not on a typed compiler-corpus run. A typed corpus comparison remains outside the current shared harness contract.

## Final-source validation

The complete run exited 0: 119 test events passed, 0 failed, 1 skipped (TestCheckerBridgeRefusalPending). Counts include parent tests and subtests. The package event passed. Wall time was 3683.103 seconds; package time was 3668.256 seconds. nproc was 5, CPU quota was 4, and one-minute load samples ranged from 1.000 to 13.431 with median 1.925. See [summary](evidence/full-summary.json), [complete log](evidence/full.jsonl), [inputs and exact command](evidence/full-inputs.json), [load samples](evidence/full-load.jsonl).

TestRulesAgree, TestOwnedWitnesses and TestMutants all passed on the final source: [owned results](evidence/final-owned-results.json). All 79 captured upstream cases match across Go, Node, emitted JavaScript and sanitized native. The prefix covers nine upstream test functions: [capture summary](evidence/upstream-capture-summary.json). The local-shadow mutant compiles, runs and disagrees with Go on native, Node and emitted JavaScript, each at case 1 line 28: [final mutant lines](evidence/final-mutant.log). All 81 descriptor mutants passed. This completed run supersedes the initial focused evidence for certification of the final source.

The 883-file aggregate compiler/stage1 comparison passed on this final attempt (699.25 seconds, identical 30,453,763-byte output); the earlier timeout remains recorded. Aggregate count throughput round 1 was Go 2.193554155 seconds and native 20.399473835 seconds, with 28,312 findings on both sides. These count-only measurements do not isolate this typed rule. Typed program/lint/recording timings remain in the complete log without being described as rule-only execution timings.

Registry, vet, gofmt and whitespace checks passed: [check results](evidence/checks.json). The setup retry completed in 110.651 seconds, with the Go build ready at 110.372 seconds: [setup log](evidence/setup.log). An initial setup attempt was interrupted while the default Go cache filled the workspace; cleaning that regenerable cache and using the scratch cache let setup complete.

All optional lint inputs were supplied. The requested zero-skip bar remains unmet because the base's TSGoError structural prerequisite is absent, not because a corpus or profiling input is missing. That shared prerequisite was not changed. Five claimed rules remain blocked as named above; no approximation, private checker, shared helper, harness or bridge edit was introduced.
