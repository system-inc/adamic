Built: enter, returnFrame, throwFrame and throwTarget, one helper per .a file; sixteen prerequisite entries across four rules.
Commits: this continuation claims 036c836de and e9d4dc76d pushed before their code; prior helpers refreshed at 2534ca1b8; parked rules d508f5ef2.
Checks: full touched package PASS 39.739s against actual Go on all four consumers plus controls, source Node, emitted JavaScript, ASan/UBSan native; vet and uncached compiler oracle pass.
Mutants: eight compiling state/identity/frame/target mutations caught solely by comparison on all three backends.
Not covered: complete CFG/rule integration, zero complete rules unblocked, invalid nil-frame panic bytes, malformed duplicate-index graphs, full repository gate or speed claims.

Observed calls: array-callback-return 79 (632 bytes), consistent-return 123 (984 bytes), no-unreachable-loop 8,861 (70,936 bytes), react-hooks/rules-of-hooks 495 (4,309 bytes), controls 120 (960 bytes), identical on each backend. Four helper prerequisite edges removed across these exact four rules; zero complete rule blocker sets removed. The helper takes a handed block and reads no syntax kinds.

Base setup timings: Go/clang/Node/submodules ready 0s; cache warm and total 85s; nproc 5. Test output retained in evidence/enter-base.log, setup in evidence/setup.log. Capture runner uses actual Go private function bodies with post-call observation only, validates that each consumer suite ran and produced calls, and fails any Go test failure. No canned finding table or Go expected result is used by the runtime helper.

Landing refresh: rebased only this unit claim/implementation from the required helpers base onto origin/main f8013f0ba. All four consumer observations and both mutants pass again, package PASS 11.021s. Tested source tip 3edde65615aab0a20e23d6a407641e8a040f0dc8. No shared files included by this rebase; helper has no dependency on the absent foundation source files.

## Second helper and final gate

returnFrame claim 7b85efb89 was pushed before code, after enter was rebased, re-green and pushed at 57bc321d5. Wildcard fetch checked 538 origin refs and all 19 distinct claim blobs; returnFrame tied the highest remaining count, four consumers. Go pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db.

Actual Go returnFrame observations: array-callback-return 90 (270 bytes); consistent-return 70 (208 bytes); no-unreachable-loop 426 (1,278 bytes); react-hooks/rules-of-hooks 45 (135 bytes); independent controls 1,022 (2,062 bytes), byte-identical per backend. Original helpers enter observations are unchanged. Combined 11,331 helper observations and 81,774 output bytes per backend. Both helpers remove one prerequisite entry each from each of these four exact rules, eight entries total. None removes a rule's final blocker. No rule findings/fixes/native speed certificate is inferred from helper parity.

Final commands with output redirected to retained logs:

    source /workspace/adamic-tools/env.sh
    go test ./stage1/cohere/lint/helpers/from_wave1_01 -count=1 -timeout 15m -v
    go vet ./stage1/cohere/lint/helpers/from_wave1_01
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -timeout 10m -v

Full touched package PASS 21.371s (enter 11.13s, returnFrame 10.24s); vet silent success; filtered input oracle PASS 0.883s, six probe misses. All twelve mutant/backend observations execute successfully then disagree with Go. Compiler refusals encountered during driver development (numeric console argument and implicit boolean-or-undefined condition) are retained as evidence and not counted as semantic mutant kills; corrected driver emits a string and helper compares the indexed flag explicitly to true.

Branch created from the required helpers base then rebased onto origin/main f8013f0ba; no shared harness files touched. Only this branch's claim and helper files differ from main. Earlier leftover untracked .generated registry output is outside this unit and is not committed. The parked rule branch remains 850fcdd0f with scoped green results and its shared blockers named. No new helper is reserved after this completed pair. Rebase parked work before taking further work when the named harness SHA is supplied.

Continuation refresh: both helpers rebased cleanly onto origin/main c01907a70 and match Go again on all three backends, all four mutants caught, package PASS 20.157s. Parked rule branch revalidated and pushed at d508f5ef2. Main leak-check/developer-tool changes retained without reversions. No new claim before both pushes.

Third helper: throwFrame claim 036c836de pushed before code. Actual private Go observations from all four consumer suites match Node source, emitted JavaScript and sanitized native: array-callback-return 74/206 bytes, consistent-return 62/178 bytes, no-unreachable-loop 5,443/16,217 bytes, react-hooks/rules-of-hooks 722/2,136 bytes, controls 2,067/4,764 bytes per side. Package selection PASS 34.630s. Outermost-handler and inverted catch-finally semantic mutants compile/run and are comparison-only catches on every backend. Exhaustive bounded six-state stacks through depth four plus every uint8 position/finally combination; no nil-frame panic parity. The capture-generation naming/return-statement failure is retained and not counted as a mutant. Four more prerequisite entries removed, twelve cumulative; zero complete rules unblocked. No regex matcher or shared harness changes.

## Fourth helper and second full gate

throwTarget claim e9d4dc76d pushed before implementation, after throwFrame green/pushed at 50a340ca5. Wildcard fetch checked 561 origin refs and all 20 distinct claims blobs; each selected helper tied highest unclaimed fan-out at four. Both own branches are based on origin/main c01907a70; parked rules refreshed at d508f5ef2. No named batch 8 Diagnostic SHA supplied.

Actual private Go throwTarget: array-callback-return 8 calls/48 bytes, consistent-return 4/24, no-unreachable-loop 56/336, react-hooks/rules-of-hooks 12/72, controls 4,096/25,597, identical per backend. Observed block-index/identity/null outputs come from Go's actual return, not a handwritten answer. Eighty real consumer calls plus 4,096 controls. Catch-outside-try and finally-instead-of-catch mutants compile/run then comparison catches both on all three backends.

All four helpers together: 23,875 observations, 131,352 bytes per backend. Each helper removes one prerequisite entry from array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks, four entries/helper and sixteen total. Zero complete rule blocker sets removed. No full rule findings/fixes or throughput claim.

Final commands, output redirected to retained batch2 logs:

    source /workspace/adamic-tools/env.sh
    go test ./stage1/cohere/lint/helpers/from_wave1_01 -count=1 -timeout 15m -v
    go vet ./stage1/cohere/lint/helpers/from_wave1_01
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -timeout 10m -v

PASS full package 39.739s: enter 10.11s, returnFrame 10.00s, throwFrame 9.90s, throwTarget 9.72s. Vet silent success; filtered uncached input oracle PASS 0.880s, six probe misses. All eight semantic mutants execute and are comparison-only catches across 24 backend checks: retained reachability; retained current block; outermost finally; skipped frame zero; outermost exception handler; inverted catch/finally predicate; catch chosen outside try; finally returned instead of catch. Initial throw capture-generation naming error is retained and not credited as a mutant.

Setup Go/clang/Node/submodule ready timings 0s, cache warm and total 55s, nproc 5. Current main's developer-tool/leak-check changes retained through clean rebases; no shared file changed or reverted. Helpers use no regex, so no hand-rolled matcher introduced. Shared harness, graph integration, nil-pointer panic behavior, malformed duplicate-index graphs and full repository/corpus gate remain outside this certificate. Prior rule blockers stay named on the parked branch. Only the own helper branch is pushed; no additional helper reserved after this completed continuation pair.
