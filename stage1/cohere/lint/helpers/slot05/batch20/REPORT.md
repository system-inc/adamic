# Batch twenty report

Built forStatement, forInOfStatement and accessOrCall in separate .a files. Each serves array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Twelve prerequisite occurrences removed, zero final blockers; no rule is declared ported. Exact mappings and cumulative counts are in readiness.json and CONSUMERS.md.

## Claim and landing

Claim 3c3cd554 was pushed before source. All twenty origin codex/lint-helpers* branches were fetched and every claim inspected. The larger comments bundle remains owned in shared HELPERS.md. These helpers tie the highest unclaimed concrete count, four consumers each. Subsequent refreshes, including the final publication scan, found no competing claim.

Current main c7991b90 and area b4691483 are already ancestors of the branch and did not advance during this unit. No new rebase was necessary. The area includes the shared harness and all legacy registry migrations. No main or area push is made. All earlier 53 helpers were complete and pushed at 4a65e5a6 before this claim. Their 197 compiling semantic mutant witnesses remain applicable to byte-identical compiler, pinned cohere, Node oracle, options parser, inventory, readiness and retained helper inputs. evidence/retained-input-identity.json records the Git objects compared between the previous publication and the pre-implementation claim commit; only the new batch20 directory adds helper oracle inputs. Prior packages were not rerun in this unit. The final response names the implementation/evidence commit.

## Observed comparisons

The private overlay executes actual original helper bodies from cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. Only dependency calls are renamed for observation. Those dependencies still execute real Go; nested instrumentation is muted. Their graph effects are exported as adapter data. Each dependency is also checked to preserve the incoming optional-chain stack, including the current helper's join where applicable. The driver replays graph effects while retaining the port's own chain stack; it does not reset that stack to the expected Go state.

Source Node, emitted JavaScript and ASan/UBSan native compare byte-for-byte with Go. Comparisons include exact current block, reachability/incoming flags, ordered successor edges, remaining jump and chain prefixes, labels, continue/break/hook destinations, broken flags and operation trace. Numeric node/block arena handles preserve the identity of the real Go objects through the adapter. Helpers consume typed projections rather than numerical AST kinds. No rule, shared harness, compiler or regex matcher is changed.

All four consumers' test files supply 166, 272, 71 and 555 captured Go string occurrences respectively. With targeted controls: 801 distinct strings and 12,570 parsed nodes. The oracle selects 25 for statements, 12 for-in/of statements and 719 member/call/new/tagged-template nodes. Loops each run under reachable/unreachable entry, three existing jump-prefix depths and absent/present throwable handler: 300 for and 144 in/of rows. Access nodes also run with three chain-prefix depths: 25,884 rows. **26,328 total helper inputs**. Captured strings are parsed as test data, not executed as complete rule programs; some test literals are not complete valid source files. Not every string contains each target node kind.

Controls include omitted loop clauses, expression/declaration initializers, constant and short-circuit conditions, disconnected increments, terminal/continue/break bodies, ordered multiple bindings, destructuring defaults and computed keys, nested optional chains, optional calls/elements, generic calls and constructors with ordered arguments, absent argument lists and tagged templates. Compact snapshot serialization retains all observed graph values while reducing JSON node allocation.

## Commands and results

All test output went directly to logs, never through a pipe. Setup succeeded: Go 0s, clang 0s, Node 0s, submodules 0s, build cache 41s, done 41s on nproc 5 (four-core cgroup quota, 17.6 GB). Sourced /workspace/adamic-tools/env.sh. Versions Go 1.27.1, clang 20.1.8, Node 24.19.0.

```sh
bash cloud/setup.sh > /tmp/lint05-batch20-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH20_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch20/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch20 -count=1 -v -timeout=20m > /tmp/lint05-batch20-helpers.log 2>&1
go vet ./... > /tmp/lint05-batch20-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch20-oracle.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05/batch20 > /tmp/lint05-batch20-format.log
```

Final package **PASS 367.839s** with all 25 compiling semantic variants caught. Vet and formatting pass with empty logs. Six external input fixtures PASS 1.076s uncached, zero cache hits and six probe misses. git diff --check passes. Every successful native baseline and mutant must exit zero with empty stderr under sanitizers.

## Every mutant

All 25 variants are independent temporary copies. Each compiles and exits zero with empty stderr before its wrong output is compared to Go. Refusal, sanitizer error, panic or clang warnings cannot earn semantic-mutant credit. evidence/mutants.json retains each exact first differing output line, byte and pair of values.

| Helper | Independent variants caught |
|---|---|
| forStatement (7) | invert initializer declaration routing; lose omitted-clause self-loop; enter incrementor normally rather than disconnected; continue to test instead of incrementor; omit incrementor expression; invert truthy exit routing; backedge to test instead of incrementor |
| forInOfStatement (7) | omit iterable; omit head exit edge; route declarations as assignment targets; omit declaration binding; continue to body instead of head; omit iteration hook; backedge to body instead of head |
| accessOrCall (11) | omit outermost query; omit property traversal; omit computed key; omit call arguments; omit tagged template; omit throwable fork; retain own join after exit; reverse call arguments; reverse constructor arguments; omit call type arguments; omit property optional fork |

The two argument reversals preserve argument count, holding source order separately from presence. All variants are caught by the actual Go output/operation comparator. They do not mutate production files.

## Findings and limits

The preliminary gate passed the three baselines, but one for-in/of mutant anchor matched both the initial head link and backedge. It was replaced with a unique surrounding anchor. A preliminary optional-chain variant panicked because snapshot replay replaced the port's own chain stack with expected Go state. The adapter now proves real dependencies preserve their incoming stack, and replay preserves the port's own mutations. The same variant then compiled and completed successfully with output disagreement. The preliminary failed gate is retained in evidence/initial-mutant-adapter-failure.log and receives no pass or mutant credit. No shared file was edited.

Not covered: complete rule findings, positions, fixes, suggestions or options; shared registry integration; the external dependency implementations; arbitrary forged nodes or inconsistent adapters; arbitrary hooks changing graph state; exhaustive AST combinations or object-handle generic instantiations. The native driver instantiates the helpers with numeric arena handles. Defensive branches unreachable on this captured parser corpus are not claimed exhaustively held.

The full repository gate, including the seventeen required stage1 external-input checks, was not run. This is the authorized bounded touched-package and filtered-oracle gate. No skip is counted as green and no check is relaxed or removed.

## Landing rerun on October 7

The landing-first cap makes this unit the rebase and verification of all 56 existing helpers. No additional helper was claimed or implemented. Previous publication 8f70f184b05de95a990e1c04b6458baa5d457a27 was rebased cleanly (85 commits) onto current origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898, which contains current origin/main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06. Rebased source HEAD is dbad16667fb7bc23340529572e1c1939d5848872. The retained helper subtree remains exactly 1ce3081481308ee62860a269887fe2abaeae1464. Compiler/runtime inputs changed upstream, so every retained helper oracle was rerun rather than carrying earlier witnesses forward. The final fetch confirmed both bases unchanged and ancestors. Only codex/lint-helpers-05 is published.

Setup succeeded: Go, clang, Node and submodules each 0s; build cache 40s; total 40s; nproc 5, four-core quota, 17.6 GB. Sourced /workspace/adamic-tools/env.sh. Exact fresh commands, all output directly to logs:

```sh
bash cloud/setup.sh > /tmp/lint05-landing21-setup.log 2>&1
git rebase origin/area/stage1-lint > /tmp/lint05-landing21-rebase.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test -p 2 ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot05/... -count=1 -v -timeout=30m > /tmp/lint05-landing21-helpers.log 2>&1
go vet ./... > /tmp/lint05-landing21-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05 > /tmp/lint05-landing21-format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-landing21-oracle.log 2>&1
```

All twenty actual helper packages PASS, including batch20 at 388.832s. All 222 compiling semantic mutants were caught again by the real Go output comparators, including the four shared inherited variants and all 218 owned variants. Every exact first differing value is retained in evidence/landing-mutants.json; complete output and per-package timings are in evidence/landing-helpers.log and landing-package-results.json. Vet and formatting pass with empty logs. All six filtered Node input fixtures PASS 8.789s uncached, zero cache hits and six probe misses. No selected check skipped, was relaxed or removed.

Coverage and dependency removals are unchanged: 56 helpers, 319 prerequisite occurrences across 70 consumers, with 50 helper-ready rules under the frozen assumptions. Complete rule diagnostics, positions, fixes, suggestions, options, shared registration integration, and the full repository gate including its seventeen required stage1 external-input checks were not run. No new helper claim is made in this landing unit. Existing helper coverage limits above and in earlier batch reports still apply.
