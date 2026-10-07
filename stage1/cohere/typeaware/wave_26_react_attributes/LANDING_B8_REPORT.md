Built: rebased all twenty-one ports onto current main; rule implementations unchanged.
Commits: validated rebased tip a479c03899ab61192f068362ff70975bd53f9012 on main b8fb957aa839a9e8cb0b54279dd9864fa317bd30.
Checks: all five Go suites, six React source rules, raw HIR, full bridge, filtered Node oracle and vet passed.
Mutants: all twenty-one native rule mutants caught by complete bytes; question, SSA, released-handle and seven bridge foundation mutants passed.
Not covered: full repository gate, shared registration, checker-linked emitted JavaScript or every configurable-option matrix.

The landing-first rebase completed without conflicts. It includes the upstream
inherited-static-field implementation and oracle fixture. No compiler, shared
harness, registration or rule source was edited by this landing unit.

The frozen populations remain 77 compiler files and 287 repository files.
All twenty-one default-option ports again agree with unchanged production Go
on complete findings, fix edits and suggestions, under native and sanitizers.
The separate positive control suites total 692 files and 457 findings.
Raw HIR independently agrees on 22 controls and 1,046,120 production bytes.
All six React rule sources also match production Go under source Node.
Checker-linked emitted JavaScript still explicitly refuses an unlinked
typescript-go call; no shared implementation was modified to bypass that gap.

Exact commands, exit codes and process durations are retained in
[the command ledger](validation/landing-b8fb957a/runs.json.gz) and
[the sequential runner](validation/landing-b8fb957a/run.py.gz). Each command
redirected stdout and stderr to its own log. All twenty recorded commands exit 0.
The five production Go suites passed in 175.99, 149.55, 142.21, 126.80 and
155.41 seconds. Static, render and effect source suites passed with 14, 25 and
30 control findings; the attribute suite passed with 163 findings/58,921 bytes.
All suites additionally matched both frozen corpora under sanitizers.

The complete bridge gate passed in 109.55 seconds: 1,600 independently answered
positions over four compiler files and 54,982 exact bytes under ASan, UBSan and
LeakSanitizer. Its seven foundation mutants exercise input length, output length,
retained handles, wrong positions, link opt-in, missing C frees and region heap
allocation. Bounds mutations trigger ASan; missing frees/heap allocation trigger
LeakSanitizer; wrong positions fail Go bytes; stale handles and link mutations
fail the explicit contract/refusal expectations.

The filtered Node oracle passed, including inherited_static_field_read.a and
the one-byte mutant. It checked source Node, native sanitizers and emitted
JavaScript for the selected ordinary oracle fixtures. Vet and the compiler/owned-question gofmt logs are empty. The broad gofmt
scan identifies only facts.go: its one-line registration cases follow this
unit's explicit one-line shared-edit constraint and were left untouched. The initial formatting command omitted the toolchain environment and
reported gofmt not found; the unchanged check passed after sourcing the printed
environment. Setup passed: Go 0 s, clang 1 s, Node 1 s, submodules 1 s, cache
warm 180 s, total 180 s; nproc=5, quota=4 cores, memory=17.6 GB.
Named obsolete worker scratch binaries/archives were removed to free disk; all
sources and logs were preserved. No gate command failed due to disk exhaustion.

Every successful rule mutant below compiled, exited 0 and had empty stderr;
only the complete-byte comparison against production Go caught it. Mutation
definitions and rule-name mappings remain in the linked reports from README.md.
Raw-question mutants are also listed below with their observed catcher.

```text
first: wave_26_test.go:136: operation mutant: exit 0, empty stderr, byte oracle catches byte 535
first: wave_26_test.go:136: provider mutant: exit 0, empty stderr, byte oracle catches byte 2046
first: wave_26_test.go:136: relation mutant: exit 0, empty stderr, byte oracle catches byte 2807
first: wave_26_test.go:146: type-arguments mutant: exit 0, empty stderr, byte oracle catches byte 58
first: wave_26_test.go:181: released-registry mutant: exit 0, required panic 70 catches it
next: wave_26_next_test.go:154: listener mutant: exit 0, empty stderr, byte oracle catches byte 1351
next: wave_26_next_test.go:154: render mutant: exit 0, empty stderr, byte oracle catches byte 6484
next: wave_26_next_test.go:154: mock mutant: exit 0, empty stderr, byte oracle catches byte 13857
next: wave_26_next_test.go:170: syntax question mutant: exit 0, empty stderr, byte oracle catches byte 5636
next: wave_26_next_test.go:170: lineage question mutant: exit 0, empty stderr, byte oracle catches byte 57
next: wave_26_next_test.go:170: literal question mutant: exit 0, empty stderr, byte oracle catches byte 6484
next: wave_26_next_test.go:170: transform question mutant: exit 0, empty stderr, byte oracle catches byte 9924
next: wave_26_next_test.go:206: released-registry mutant: exit 0, required panic 70 catches it
third: wave_26_third_test.go:136: eval mutant: exit 0, empty stderr, byte oracle catches byte 3841
third: wave_26_third_test.go:136: extend mutant: exit 0, empty stderr, byte oracle catches byte 11304
third: wave_26_third_test.go:136: function mutant: exit 0, empty stderr, byte oracle catches byte 18330
third: wave_26_third_test.go:150: syntax question mutant: exit 0, empty stderr, byte oracle catches byte 9322
third: wave_26_third_test.go:150: origin question mutant: exit 0, empty stderr, byte oracle catches byte 1299
third: wave_26_third_test.go:186: released-registry mutant: exit 0, required panic 70 catches it
fourth: wave_26_fourth_test.go:142: function mutant: exit 0, empty stderr, byte oracle catches byte 2427
fourth: wave_26_fourth_test.go:142: nonconstructor mutant: exit 0, empty stderr, byte oracle catches byte 7081
fourth: wave_26_fourth_test.go:142: wrapper mutant: exit 0, empty stderr, byte oracle catches byte 9475
fourth: wave_26_fourth_test.go:156: constructor question mutant: exit 0, empty stderr, byte oracle catches byte 556
fourth: wave_26_fourth_test.go:156: origin question mutant: exit 0, empty stderr, byte oracle catches byte 59
fourth: wave_26_fourth_test.go:192: released-registry mutant: exit 0, required panic 70 catches it
fifth: wave_26_fifth_test.go:171: promise mutant: exit 0, empty stderr, byte oracle catches byte 3172
fifth: wave_26_fifth_test.go:171: regex mutant: exit 0, empty stderr, byte oracle catches byte 12273
fifth: wave_26_fifth_test.go:171: rest mutant: exit 0, empty stderr, byte oracle catches byte 8419
fifth: wave_26_fifth_test.go:187: syntax question mutant: exit 0, empty stderr, byte oracle catches byte 1223
fifth: wave_26_fifth_test.go:187: binding question mutant: exit 0, empty stderr, byte oracle catches byte 58
fifth: wave_26_fifth_test.go:187: pattern question mutant: exit 0, empty stderr, byte oracle catches byte 11033
fifth: wave_26_fifth_test.go:187: comments question mutant: exit 0, empty stderr, byte oracle catches byte 25687
fifth: wave_26_fifth_test.go:223: released-registry mutant: exit 0, required panic 70 catches it
raw-hir: omitted SSA mutant: exit 0, empty stderr, production bytes catch it
raw-hir: released data survives; stale question panics 70; registry mutant exits 0
static: assignment-taint mutant: exit 0, empty stderr, byte comparison catches it
render: setter-alias mutant: exit 0, empty stderr, byte comparison catches it
effect: setter-alias mutant: exit 0, empty stderr, byte comparison catches it
attributes: button_has_type mutant: exit 0, empty stderr, complete-byte oracle caught it
attributes: checked_requires_onchange_or_readonly mutant: exit 0, empty stderr, complete-byte oracle caught it
attributes: display_name mutant: exit 0, empty stderr, complete-byte oracle caught it
dependencies: both new questions: copied data survives release, stale queries panic 70, retaining-handle mutant exits 0 and is caught
questions: jsx_structure: compiled, intended contract assertion caught mutant
questions: symbol_locations: compiled, intended contract assertion caught mutant
```

Quiet alternating medians for the final three-rule attribute/display-name suite:

| Corpus | Native | Production Go |
| --- | ---: | ---: |
| Compiler | 7.387989 s | 0.537364 s |
| Repository | 1.222332 s | 0.280945 s |

Every timed sample compared complete diagnostic streams. Native remains slower.
These are three-rule whole-process measurements, not a combined 21-rule runner.

The final explicit all-branch fetch found 554 origin refs. In the original
197-rule by-volume population, the 25 baseline/main ports and 172 claimed names
cover every candidate. No unclaimed rule remains and no additional claim was
made. The complete ref/claim/ranking audit is selection-audit.json.gz beside
the logs. This report stops at the user's exhaustion gate.
