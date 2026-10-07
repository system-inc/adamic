Built makeBreak in one .a file; pushJump/popJump withdrawn because slot 03 claimed earlier.
SHAs: claim 85af6690; withdrawal 2aa00876; rebased withdrawal head 64411850 on main b8fb957a; delivery SHA named in final report.
Commands and outputs: retained Go/source Node/emitted JS/sanitized native PASS 123.847s, 27,648 queries; setup 35s, nproc 5; full prior-helper landing gate recorded separately.
Mutants: all seven compiling semantic variants caught by independent Go comparisons; exact witnesses in evidence/mutants.json and retained.log.
Not covered: external graph dependencies, complete CFG/rule findings and integration, arbitrary callback effects or invalid AST/state representations.

# Observations

Slot 03 reserved pushJump/popJump at 06:00:46 UTC (e4b23dd4), before slot 05 at 06:02:56 (85af6690). They appeared during the fresh batch13 validation after the first branch scan. Slot 05 withdrew them before publication of implementation; no duplicate source is delivered. The retained makeBreak claim is unique across all twenty refreshed helper branches. The old latest trio was reverified PASS 86.290s, including all twelve mutants, before claiming this batch.

Main advanced from c01907a7 to b8fb957a with inherited static-field reads. All 71 owned branch commits rebased cleanly; no upstream compiler diff was changed or reverted. The current-main retained method comparison passed in 123.847s while the earlier-helper landing gate ran concurrently, so this elapsed time is not a performance claim.

# Coverage

27,648 invocations, 803 unique source strings including five explicit controls, 864 parsed source/loop/switch nodes and nine distinct nil/text label controls. Consumer literal occurrences: array-callback-return 166, consistent-return 272, no-unreachable-loop 71, react-hooks/rules-of-hooks 555. All consumer Go test files listed by the frozen inventory are inspected. Labels are taken from parsed breaks plus nil, empty, ASCII case, unmatched and non-ASCII controls. Depths zero through five and 256 masks cover breakability, nil destinations, existing broken flags and loop identities under both current reachability states. Source/configuration/expected-text literals are input coverage, not complete diagnostic replay.

Four dependency occurrences are removed across these four consumers. No final blocker is removed, no inventory status is changed and no rule is declared implemented.

# Mutants

| Mutation | First output line | Go behavior observed |
| --- | ---: | --- |
| ignore unreachable early return | 1 | no dependency call when current unreachable |
| ignore unlabeled breakability | 4610 | nonbreakable target skipped |
| ignore exact label membership | 4612 | unmatched label links no target |
| omit broken marking | 4618 | broken true visible in state and at link |
| omit makeUnreachable | 2 | reachable unmatched input becomes current block 5, false |
| search outermost first | 9228 | repeated a resolves inner target 2 |
| link before broken marking | 4618 | link observes broken true |

Each variant is built in a temporary copy and must compile, exit zero and print no stderr before its wrong output is credited. Compiler refusal, warnings, sanitizer/panic failures are not credited. Baselines match Go stdout bytes on source Node, emitted JavaScript and ASan/UBSan native with leak checking.

The first private-driver run refused a function declaration inside its case loop; an arrow function replaced it within this unit. That superseded attempt was stopped and is not pass evidence. No shared harness or compiler file was changed.

# Commands and limits

All test output goes directly to logs. Setup succeeded: Go/clang/Node/submodules ready 0s; build cache warm 35s; done 35s on five processors (cpu.max 400000 100000), 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Repository-wide vet, type loading and uncached six-fixture input oracle are rerun after rebase and archived with landing evidence. Full repository test gate is not run.

No new rule, rule dispatch, finding serialization, shared registration, shared harness or protected compiler file was changed. Exact regex instructions remain respected: this helper does no pattern matching; it compares identifier labels with array includes.

Current-main landing gate completed: all thirteen actual earlier-helper packages PASS and all 110 prior variants caught. The command had one extra no-Go-files slot root argument and exited 1; no actual test failed. Corrected package pattern plus explicit-refusal check exited 0. Rebased vet/types and filtered oracle PASS 1.380s. See ../LANDING_REPORT_4.md for all commands, timings and the exact argument error.

The two withdrawn slots were replaced by statements and makeContinue, completed at 76ff523314dc72ede98a80bbba2068b04849063a; their report is ../batch15/REPORT.md. This finishes the three-helper batch with twenty compiling semantic mutants caught and 56,102 distinct invocations. All owned helpers are complete on main b8fb957a.

Final publication update: rebased onto main 39638d9e; all 76 patches and all compiler/helper oracle input object IDs preserved. The complete trio rerun PASS: batch14 111.792s, batch15 145.503s, all twenty new variants caught again. Vet/format PASS, filtered uncached oracle PASS 1.143s. Active implementation SHAs: makeBreak 4d1309fe, replacements a03a623b. See ../LANDING_REPORT_4.md and ../landing-evidence-4/final-trio.log for final-main evidence and all 110 earlier witnesses.
