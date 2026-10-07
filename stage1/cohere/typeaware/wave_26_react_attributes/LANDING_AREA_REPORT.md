Built: rebased twenty-one native ports onto area/stage1-lint without rule source changes.
Commits: validated rebased tip ec14ab63b2415288f40b28dd554b36de51cd9aad; area base 7481e0324e34a2537aafa9db7eeacda50405611b.
Checks: twenty-two gate commands passed; all owned rule oracles, merged harness, both corpora, sanitizers, bridge, Node oracle and vet.
Mutants: twenty-one rule byte mutants, raw questions, SSA, handles, bridge foundation and shared emitted-JavaScript mutant caught.
Not covered: shared registration for these existing private runners, checker-linked emitted JavaScript, full repository gate or every option matrix.

The rebase completed without conflicts. The validated area base includes
main 39638d9e278d38bb5aeae887f46d55a70e47aaad and merged harness 41eb6eab2.
Main remained at that SHA through the final fetch. The area advanced to
d65a8f93 during validation; this report claims the tested 7481e032 area base.
No compiler, shared harness, generator, registration or rule implementation
was edited by this landing unit. No new rules were claimed.

The registry package and all five targeted merged-harness checks passed.
The harness compared Go, source Node, emitted JavaScript and native bytes.
Its emitted-JavaScript mismatch mutant triggered the intended nested failure
and the outer test passed. The .ts/.a rename test preserved 663 identical
bytes. Complete suggestion serialization exercised multiple edits, Unicode,
escaped separators and empty suggestions; the mixed case retained unapplied
suggestions beside an applied automatic fix. JSX witness script mode passed.
Shared harness package elapsed: 262.363 seconds. This does not claim that the
existing wave-26 private runners are registered with that harness.

All twenty-one default-option ports again match unchanged production Go
on complete finding, fix and suggestion bytes over their positive controls,
77 compiler files and the frozen 287-file repository manifest. The separate
control suites total 692 files and 457 findings. Normal and sanitized outputs
match, with empty native stderr. The six React source rules also match under
source Node. The CLI's checker-linked emitted-JavaScript path still explicitly
refuses an unlinked typescript-go library call; shared .a loading and suggestion
serialization are now independently green and are not claimed as blockers.

Raw HIR matches 22 controls and 1,046,120 production bytes. The full bridge
gate passed in 107.55 seconds, including 1,600 positions over four compiler
files and 54,982 exact bytes under ASan/UBSan/LeakSanitizer. Seven foundation
mutants exercise input/output byte lengths (ASan), retained handles (stale
handle assertion), wrong positions (Go bytes), removed link opt-in (explicit
refusal), omitted C frees and region heap allocation (LeakSanitizer).
Filtered ordinary Node oracle fixtures, including inherited static fields,
passed source/native/emitted-JavaScript checks and the one-byte mutant.

All twenty-two exact commands, exits and durations are in
[the ledger](validation/landing-area-7481e032/runs.json.gz); the environment,
log destinations and execution order are in
[the runner](validation/landing-area-7481e032/run.py.gz). All test output went
to files. Vet and compiler/owned-question formatting logs are empty. One-line
shared facts.go registration cases remain as required by the unit's edit limit.
Setup passed: Go 0 s, clang 1 s, Node 1 s, submodules 1 s, cache warm 111 s,
total 111 s. nproc=5, quota=4 cores, memory=17.6 GB. Named obsolete worker
binaries and archives were removed to free space; source inputs and logs
were preserved. No gate command failed.

Every native rule mutant below compiled and exited 0 with empty stderr.
Only the complete-byte comparison caught it. Mutation definitions and rule
name mappings are retained in the earlier reports linked from README.md.
The raw-question, SSA and handle observations name their distinct catchers.

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

Quiet alternating whole-process medians for the final three-rule attribute
and display-name suite, with complete stream comparisons for every sample:

| Corpus | Native | Production Go |
| --- | ---: | ---: |
| Compiler | 7.799143 s | 0.601840 s |
| Repository | 1.564761 s | 0.398324 s |

Native remains slower. These are three-rule measurements, not a combined
twenty-one-rule runner. Exact samples and commands are retained with the
current native/Go comparison evidence and cost.log.gz in this landing folder.

The final fetch explicitly refreshed all origin heads and pruned stale local
remote-tracking refs. The audit scanned 433 current origin refs. In the
original 197-rule ranking, 25 baseline/main ports plus 172 claimed names
cover every candidate: zero unclaimed rules. The complete ref SHAs, claim
blobs, rule names and ranking are preserved in selection-audit.json.gz.
Work stops at the requested exhaustion gate without another claim.
