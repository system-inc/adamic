# Slot 04 wave 20 report

Built HasJsxOrReactHookCalls, searchForJsxOrHook and descendsForJsxSearch, one helper per `.a` file. All 56 prior retained helpers were complete and pushed before this wave; the retained total is now 59. Claim df5d5097 was pushed before implementation. The final handoff names the implementation/evidence commit.

Current lint area d65a8f931c98655936ae04c6899f38f14862b73e and main 39638d9e278d38bb5aeae887f46d55a70e47aaad are unchanged and ancestors of this worker branch. The earlier area rebase already included the shared harness/finding model and allocator changes; they remain intact. No main or area branch was pushed. All twenty origin helper branches and every claim file were checked; evidence/claims.json pins their pre-reservation tips. The selected helpers were unclaimed and tied the highest available fan-out, four consumers each. The previously delivered comment bundle remains owned.

## Behavior and readiness

The wrapper searches at depth zero. The recursive helper checks presence and depth > 20 before JSX or hook detection, returns early on success, and recurses only through permitted children. It recognizes all three Go JSX forms. The descent helper rejects nil and switch-statement children; passing a switch directly to the search can still find its descendants, matching Go. Depth, the switch blind spot and negative starting depths are behavior, not assumptions repaired by this port.

All three helpers remove a listed prerequisite for each rule below, twelve rule/helper edges total:

| Consumer | New prerequisites removed | Other frozen prerequisites remain |
|---|---:|---:|
| structure/consistency-require-matching-file-name | 3 | 13 |
| structure/react-component-no-destructuring | 3 | 11 |
| structure/react-component-no-separate-named-export | 3 | 6 |
| structure/react-component-require-properties-parameter | 3 | 6 |

This subtracts only the three new helpers from the shared frozen inventory. It does not reconcile other workers' delivered prerequisites or prior waves. No rule becomes completely helper-ready from these three alone. Readiness remains conditional on the common AST adapter and actual ecmascript/react.IsHookCall dependency. No new rule, listener, diagnostic, fix or suggestion is implemented.

## Actual Go evidence

The private overlay exports the actual unexported search/descent methods. The oracle parses consumer and control source with pinned typescript-go, projects actual node classifications and ForEachChild order, and compares the real Go methods with Adamic on the selected AST nodes and nil. Adapter node indices identify slots, not numeric kinds. Helpers compare adapter flags; no rule compares kinds as strings and no handwritten hook/regex matcher is introduced. Hook-call verdicts come from real Go IsHookCall, an explicit separately owned dependency.

Capture ran the real Go structure rule suite, PASS 0.190s, and retained 239 unique asserted fixtures from all four consumers. Entry instrumentation observed 90 HasJsxOrReactHookCalls calls, 890 searchForJsxOrHook calls and 800 descendsForJsxSearch calls. All three helpers were reached in that suite. These counts do not attribute every call to an individual consumer; the separate readiness coverage check establishes each consumer's fixture representation.

The deterministic corpus has 158 controls: 25 source forms, 132 depth-boundary sources across 33 nesting depths and four terminals, and an explicit nil case. Controls cover all JSX forms, hook/non-hook calls, parenthesized/computed/optional calls, Unicode names, switch/if/loop/try/class/nested-function shapes, negative starting depth and depths at/around 20. Ordinary controls and consumers compare every parsed AST node; boundary sources compare roots and nil. Each observation includes wrapper, descent and multiple direct search depths.

Source Node, emitted JavaScript and ASan/UBSan native match actual Go on 533 control observations and 4,986 consumer AST observations, 5,519 total per execution mode. Every native run must finish with exit zero and empty stderr, including compiling semantic mutants; sanitizer/compiler failures do not count as mutant detections.

## Commands and outputs

All test output went directly to evidence logs. Environment Go 1.27.1, clang 20.1.8, Node 24.19.0; source /workspace/adamic-tools/env.sh after setup.

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (2s)
setup: build cache warm (94s)
setup: done in 94s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

nproc printed 5. Setup warms all package/test binaries with no tests selected; this is not the full correctness gate.

- python3 stage1/cohere/lint/helpers/slot04_wave20/testdata/capture.py: 239 consumer fixtures and the helper entry counts above; actual Go structure suite PASS 0.190s.
- python3 stage1/cohere/lint/helpers/slot04_wave20/testdata/regenerate.py: 158 controls; SHA256 unchanged before/after regeneration.
- go test -count=1 -v -timeout=15m ./stage1/cohere/lint/helpers/slot04_wave20: PASS 35.377s. Go/Node/native/emitted-JavaScript parity PASS 15.30s; nine compiling semantic mutants PASS 20.00s; four coverage omissions PASS. No selected check skipped.
- go vet ./stage1/cohere/lint/helpers/slot04_wave20: exit zero, empty vet.log.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout=10m ./internal/oracle -run '^TestRuntimeLastIndexOfMatchesNode$': PASS 0.882s. Node/emitted JavaScript/release native/sanitized native agree on 758 bytes; three native misses, two Node misses, zero cache hits.
- Final fetch and merge-base checks: current main and lint area unchanged, both ancestors of HEAD. No incoming protected/shared file diff from the area base.

## Every mutant

All nine final semantic mutants compiled and ran cleanly. Each was caught by ordinary output comparison against actual Go in source Node, sanitized native and emitted JavaScript, 27 successful semantic detections. Variants exist only in temporary copies.

| Helper | Mutation | Independent check that catches it |
|---|---|---|
| descendsForJsxSearch | remove presence check | Go rejects nil; mutated descent accepts it |
| descendsForJsxSearch | invert switch predicate | parsed switch and ordinary-node descent disagree |
| searchForJsxOrHook | reject depth 20 | JSX/hook nodes at exactly 20 disagree |
| searchForJsxOrHook | omit JSX acceptance | direct/nested JSX controls disagree |
| searchForJsxOrHook | accept every call | ordinary non-hook call controls disagree |
| searchForJsxOrHook | enter excluded switch children | outer subtrees containing JSX only in switch arms disagree |
| searchForJsxOrHook | increment depth by two | nested boundary controls disagree |
| searchForJsxOrHook | return false after a child succeeds | parent subtrees containing JSX/hooks disagree |
| HasJsxOrReactHookCalls | start at depth 21 | real JSX/hook subtrees accepted by Go are missed |

Four further structural mutants omit each consumer's fixtures in turn. The readiness-derived coverage check catches consistency-require-matching-file-name, react-component-no-destructuring, react-component-no-separate-named-export and react-component-require-properties-parameter omissions. These are coverage checks, not compiling semantic mutants.

## Limits and required checks

This unit used the touched helper package plus a filtered uncached external oracle. The full repository gate and the 17 broader stage 1 TypeScript/postcss/graphql/parser correctness checks were not selected. No skip was accepted as a correctness pass, and no test, required input, assertion or harness was relaxed, removed or changed. The helper package requires its controls/consumer fixtures and invokes the real Go parser on every run. A missing fixture or compiler/oracle failure fails the unit.

Earlier retained helpers and shared harness/runtime checks remain green on this same base from their recorded runs; no base or shared code changed here. Arbitrary cyclic or malformed AST adapters, arbitrary Go parser inputs beyond these bounded cases, full IsHookCall implementation/Unicode semantics in Adamic, full rule findings/ranges/fixes/suggestions and native listener registration are not covered. The claimed helper methods are implemented; the common AST and hook predicate dependencies remain explicit handoff work. The suite compares observable boolean results, not exact callback visit counts or arbitrary effectful hook predicates.
