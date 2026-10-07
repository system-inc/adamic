Built: reachedStatement and typeArguments; eight .a helpers total, thirty-two prerequisite entries across four rules.
Commits: rebased onto origin/area/stage1-lint 7481e0324 (contains origin/main 39638d9e2); no shared files changed.
Checks: full touched helper package PASS 97.875s, 40,498 actual Go observations and 316,464 bytes per backend; vet and uncached input oracle PASS 1.064s.
Mutants: all seventeen helper mutants compile/run then comparison alone catches them on source Node, emitted JavaScript and ASan/UBSan native.
Not covered: complete CFG/AST/rule integration or opaque callback bodies, zero complete rules unblocked, invalid panic bytes, full repository/corpus gate or throughput.

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

## Landing refresh after named harness announcement

Current origin/main c01907a70 is already an ancestor of this branch. Full owned helper package revalidated: PASS 39.392s, all 23,875 Go observations and 131,352 bytes per backend remain identical; all eight semantic mutants compile and run before comparison alone catches them on source Node, emitted JavaScript and ASan/UBSan native. Vet passes silently. Filtered uncached compiler input oracle PASS 0.971s, six probe misses. Exact commands are the final gate commands above, redirected to named-harness-refresh.log, named-harness-vet.log and named-harness-input.log.

Rule branch pushed at 092de10ce with the attempted ab70f38d4 merge and ten shared-file conflicts named in its owned LANDING.md. That merge was aborted without editing shared files or reverting developer-tool changes. Named harness integration remains blocked, so no new helper was claimed. Existing four helpers still remove sixteen prerequisites from the four listed rules; zero complete rules unblocked. No full repository/corpus gate or throughput measurement repeated. Untracked generated registry output is not committed.

## Fifth helper

Claim 0fb1cc406 pushed before code after auditing all 575 origin refs and 20 distinct helper claim blobs plus HELPERS.md. Current-main ancestry and pushed fresh oracle evidence on both own branches; rules 092de10ce parked with the named harness conflicts. enterDisconnected ties highest unclaimed fan-out, four consumers.

TestDisconnectedEntryMatchesGo PASS 9.821s. Actual Go calls/output bytes per backend: array-callback-return 5/50, consistent-return 5/50, no-unreachable-loop 610/6,100, react-hooks/rules-of-hooks 6/60, controls 180/1,800. Total 806 observations and 8,060 bytes on source Node, emitted JavaScript and ASan/UBSan native. Existing suites plus five supplementary increment-loop bodies per rule; the supplements are not upstream fixtures. Initial missing-call capture failure is retained; it correctly refused to certify zero observations. Driver optional-number lowering refusal is retained and corrected with an explicit checked current block. Neither is a semantic mutant. First unsourced gofmt invocation reported command not found; sourcing the printed tools environment fixed it.

Ignore-current-reachability and discard-existing-incoming-edge mutants compile and execute successfully, then only Go byte comparison catches both on every backend. This adds four prerequisite entries, twenty total across the same four consumers, zero complete rules unblocked. Source files are .a; no shared file or regex changed. Invalid nil-current panic bytes, full graph/rule integration, complete corpus/full repository gate and speed are not covered.

## Sixth helper

Claim 7fb957e4d pushed before code after enterDisconnected was green/pushed at 7f103900a. Refreshed 578 origin refs, all 20 distinct helper claims and HELPERS.md; isThrowableIdentifier ties the highest remaining four-consumer fan-out. Names, tag names, rest tokens, property-name identity and grandparent kind are handed context views, no string-kind relevance dispatch or AST refetch. Constants are Go parser numbers from its compiler-checked generated table.

TestThrowableIdentifierMatchesGo PASS 35.544s. Actual private Go calls/output bytes on each backend: array-callback-return 52/104, consistent-return 58/116, no-unreachable-loop 5,260/10,520, react-hooks/rules-of-hooks 605/1,210, parsed/factory controls 90/180. Total 6,065 calls and 12,130 bytes. Every upstream matching consumer test runs unchanged, then independent controls call the same Go body. All three mutants compile and run, then only byte comparison catches declaration-name polarity reversal, removed binding-rest exemption and JSX tag identity inversion on all three backends.

Two new helpers add eight prerequisite entries, twenty-four cumulative across the same four exact rules, zero complete rule blocker sets removed. Together all six helpers compare 30,746 calls and 151,542 bytes per backend. Full rule findings/fixes, graph/AST adapter integration, invalid nil-node/current panic bytes and exhaustive arbitrary contexts are outside this certificate. No new rule, shared harness file or regex matcher changed.

## Third full gate and handoff

Exact commands, every output redirected to the retained batch3 logs:

    source /workspace/adamic-tools/env.sh
    go test ./stage1/cohere/lint/helpers/from_wave1_01 -count=1 -timeout 15m -v
    go vet ./stage1/cohere/lint/helpers/from_wave1_01
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -timeout 10m -v

Full owned package PASS 58.437s. Vet silent success; uncached compiler input oracle PASS 0.966s, six probe misses. All thirteen semantic mutants compile and run successfully, then comparison alone kills them on all three backends: retained reachability; retained current block; outermost finally; skipped frame zero; outermost exception handler; inverted catch/finally predicate; catch selected outside try; finally instead of catch; ignored current reachability; discarded incoming edges; reversed declaration-name polarity; removed rest-binding exemption; inverted JSX tag identity. Development failures are retained separately and not credited as mutant kills.

Setup ready timings all 0s; cache warm and total 33s, nproc 5. Both own branches include current origin/main c01907a70 and preserve developer-tool changes. Rule branch 092de10ce is parked under the explicit shared-harness exception: the ab70f38d4 merge was aborted after ten unowned shared-file conflicts, named in its LANDING.md. This helper branch changes only its own helper directory and claim, not finding.ts, context.ts, main.ts, shared generator/oracle/comparison. No regex matcher introduced. Full-source corpus and complete repository gate were not repeated. Temporary captures are regenerated during tests; pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. No further helper reserved after this completed pair.

## Landing refresh onto b8fb957aa

Rebased cleanly onto current origin/main b8fb957aa, preserving inherited-static-field and developer-tool changes. Full six-helper package PASS 63.280s with the same 30,746 observations and 151,542 bytes on source Node, emitted JavaScript and sanitized native; all thirteen compiling mutants remain comparison-only catches. Vet passes; filtered uncached input oracle PASS 1.330s, six probe misses. Commands are the third full gate commands above, redirected to main-b8 logs. Setup ready lines all 0s; cache warm and total 79s, nproc 5. No new helper claim before both own branches are refreshed and pushed. Registry rule.json kind subscriptions remain named; these Go helper views use internal Go enum numbers, not a claimed numeric stage1 parser API. Full repository/corpus gate not repeated.

## Seventh helper

Claim 7c973d706 pushed before code after both own branches rebased, re-green and pushed on main b8fb957aa. Audited 587 origin refs and 20 distinct helper claim blobs plus HELPERS.md; reachedStatement ties maximum available four-consumer fan-out. TestStatementHandoffMatchesGo PASS 35.620s. Actual Go invocation arguments, callback count and builder/node identity match on source Node, emitted JavaScript and sanitized native for all four consumer suites plus 72 independent controls. The actual hook body runs unchanged in Go; Adamic uses an observation hook because opaque consumer callback semantics are outside this helper. Invoking twice and discarding the handed node mutants compile/run successfully, then only byte comparison catches them on every backend. Four additional prerequisite entries removed, twenty-eight cumulative; zero complete rules unblocked. No syntax dispatch, shared harness or regex changes.

## Eighth helper and fourth full gate

Claim a3e8d8b9e pushed before code after statement helper green/pushed at f70eb4dfc. Audit checked 592 origin refs, all 20 distinct helper claim blobs and HELPERS.md reservations. popJump was already claimed, so no claim or implementation was made for it. typeArguments ties highest available fan-out at four, with the same four consumers. Both own branches retain current-main b8fb957aa ancestry and pushed fresh green scoped evidence; rule branch 0a642ff72 stays parked with the ten named unowned ab70f38d4 harness conflicts. No shared file modified or developer-tool change reverted.

Actual Go typeArguments observations/output bytes per backend: array-callback-return 24/148, consistent-return 12/136, no-unreachable-loop 602/726, react-hooks/rules-of-hooks 189/313, controls 122/6,818. Total 949 observations and 8,141 bytes. Existing consumer fixtures only exercised empty lists; the retained initial log records that limit. Four explicitly supplementary generic forms per rule now exercise non-empty lists, and a capture assertion refuses a consumer with no non-empty observations. Controls cover ordered nullable/repeated type-reference identities. Expression callees run unchanged in Go; the Adamic observation callback holds the handoff contract only. Skipped-first and reversed-order mutants compile/run then only comparison kills them on every backend.

Exact final commands, output redirected to retained batch4 logs:

    source /workspace/adamic-tools/env.sh
    go test ./stage1/cohere/lint/helpers/from_wave1_01 -count=1 -timeout 15m -v
    go vet ./stage1/cohere/lint/helpers/from_wave1_01
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -timeout 10m -v

Full eight-helper package PASS 101.194s; vet silent success; uncached input oracle PASS 0.976s, six probe misses. Eight helpers hold 40,498 Go observations and 316,464 bytes per backend. All seventeen semantic mutants compile and execute successfully before comparison catches them across 51 backend checks: retained reachability; retained current block; outermost finally; skipped frame zero; outermost exception handler; inverted catch/finally predicate; catch selected outside try; finally instead of catch; ignored current reachability; discarded incoming edges; reversed declaration-name polarity; removed rest-binding exemption; inverted JSX tag identity; duplicated statement hook; discarded handed statement node; skipped first type argument; reversed type-argument order.

Both new helpers remove four prerequisite entries each from array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Eight new entries, thirty-two cumulative; zero complete rules unblocked. Opaque callback bodies, AST projection and graph adapters remain caller-owned. No regex or node-kind string relevance dispatch introduced. All new Adamic source files are .a. No speed, whole-rule finding/fix or complete repository/corpus claim. Setup ready lines all 0s, cache warm and total 79s, nproc 5. Pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. Regenerating captures and test logs use temporary directories; untracked .generated registry output is not committed. No further helper reserved after this completed pair.

## Landed harness refresh

Rebased onto area 7481e0324, preserving current main and shared harness changes. Repeated the full helper package, vet and uncached TestInputAgreesWithNode commands above; logs are evidence/landed-tests.log, landed-vet.log and landed-input.log. The eight helper contracts and all seventeen mutants remain green. No new helper claimed: the separate rule branch still needs deduplication and landed-harness verification. No complete rule unblocking or throughput inferred.

## Current area d65a8f931 refresh

Rebased onto current area d65a8f931 after an explicit wildcard fetch; the default fetch refspec only updates main. Current main remains 39638d9e2. Preserved all runtime-profile changes; zero shared lint harness edits. Full eight-helper package PASS 85.926s: all 40,498 Go observations and 316,464 bytes/backend identical, all seventeen compiling mutants caught only by comparison on source Node, emitted JavaScript and ASan/UBSan native. Vet PASS; uncached TestInputAgreesWithNode PASS 6.866s, six probe misses. Same final gate commands as above, retained in evidence/d65-tests.log, d65-vet.log and d65-input.log. Setup ready 0s, warm and total 115s, nproc 5. Rule branch refreshed and pushed at 3861e9da9, with the unchanged allowLoop option-routing and live Tailwind witness blockers. No new claims, no complete rule unblocking, no full corpus or repository gate.

## Main c7991b900 and area b84a9d931 refresh

Rebased cleanly onto area b84a9d931, which contains main c7991b900, preserving proven-type/compiler/runtime-record changes. No shared lint files differ from the area. Full owned eight-helper package PASS 89.269s; all 40,498 actual Go observations and 316,464 bytes/backend remain identical; all seventeen mutants compile/run before comparison alone catches them on source Node, emitted JavaScript and ASan/UBSan native. Vet PASS, uncached TestInputAgreesWithNode PASS 6.931s with six probe misses. Same final gate commands above, logs evidence/b84-tests.log, b84-vet.log and b84-input.log. Setup ready lines 0s, warm and total 148s, nproc 5. Rule branch pushed at dda491574 with fresh owned parity and exact shared/corpus failures named. The compiler corpus input was supplied and no check skipped; other mandatory external checks/full repository gate not run. Existing helper boundaries and zero complete rules unblocked remain unchanged. No new claims.

## Registry-only area b46914832 refresh

Rebased cleanly onto area b46914832, retaining main c7991b900 and the registry-only legacy-rule migration. No shared lint edits or new claims. Full eight-helper suite PASS 84.511s: all 40,498 actual Go observations and 316,464 bytes/backend identical; all seventeen mutants compile/run before comparison alone catches them on source Node, emitted JavaScript and ASan/UBSan native. Vet PASS; uncached TestInputAgreesWithNode PASS 1.034s, six probe misses. Same final gate commands above, retained in evidence/b469-tests.log, b469-vet.log and b469-input.log. Setup ready lines 0s, warm and total 47s, nproc 5. Rule branch pushed at bfdf804f4 with two owned root-context driver adaptations and refreshed exact shared/corpus failures. Compiler corpus input supplied; no check skipped or relaxed. Other mandatory external checks/full repository gate not run. Existing helper boundaries and zero complete rules unblocked remain unchanged.

## Main b6b1538b0 and area d3a37422c refresh

Rebased cleanly onto area d3a37422c containing main b6b1538b0, preserving landed typeof compiler/runtime fixes. No shared lint edits or new claims. Full eight-helper suite PASS 86.872s; all 40,498 Go observations and 316,464 bytes/backend identical; all seventeen compiling mutants caught only by comparison on source Node, emitted JavaScript and ASan/UBSan native. Vet PASS, uncached input oracle PASS 6.359s with six probe misses. Same final gate commands above, logs evidence/d3-tests.log, d3-vet.log and d3-input.log. Setup Go ready 0s, clang/Node/submodules 1s, warm and total 146s, nproc 5. Rule branch pushed at 9688a9906 with fresh owned parity and exact shared/corpus failures, mandatory corpus input supplied. Other mandatory external/full repository checks not run. Eight helpers still remove thirty-two prerequisites across the same four consumer rules, zero complete rules unblocked. Existing coverage boundaries remain unchanged.
