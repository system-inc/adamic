Built: CFG enter and returnFrame, one helper per .a file; eight prerequisite entries removed across four rules.
Commits: enter claim d6fbc1351 and returnFrame claim 7b85efb89 pushed before their code; parked rule branch 850fcdd0f.
Checks: touched package PASS 21.371s against actual Go on all four consumers and controls, source Node, emitted JavaScript and sanitized native; vet and filtered uncached input oracle PASS.
Mutants: retained reachability/current-block identity, outermost finally, omitted first frame; all four compile/run and fail only comparison on all three backends.
Not covered: complete CFG/rule integration, zero complete rules made ready, invalid nil-pointer panic bytes, full repository gate or speed claims.

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
