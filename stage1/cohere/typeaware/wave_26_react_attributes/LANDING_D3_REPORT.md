Built: twenty-one existing ports rebased and re-greened on lint area d3a37422; no new rule source changes.
Commits: validated rebased tip 54983b294192862c4e959907a0b7b0e510906c4d, containing current main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06.
Checks: all twenty-seven current-base commands passed, zero selected skips, both corpora, sanitizers, merged harness, bridge and expanded Node oracle.
Mutants: all twenty-one rule mutants caught by complete bytes; question, SSA, handles, bridge, typeof, emitted-JavaScript and decoded-options checks passed.
Not covered: full repository gate and seventeen external-input comparisons, shared registration for private runners, checker-linked emitted JavaScript or every option matrix.

The requested b46914832 rebase completed without conflicts and all twenty-six
commands passed. Main advanced during that run with typeof null and slot
presence changes. This unit then rebased again onto d3a37422, which contains
main b6b1538b0, and reran the whole owned gate with the new typeof fixtures
and four typeof mutant tests. Final fetch confirms those same main and area
SHAs. Both rebases and both complete command ledgers and logs are preserved
in validation/landing-b4691483 and validation/landing-d3a37422. No shared
harness, registration generator, compiler or rule implementation was edited.

The exact commands, exits and durations are in
[the current ledger](validation/landing-d3a37422/runs.json.gz) and environment
inputs in [the runner](validation/landing-d3a37422/run.py.gz). All test output
went directly to files. Both runs supply the pinned TypeScript v6.0.3 checkout
and frozen 77-file compiler and 287-file repository manifests. The selected
checks have zero skips. The full repository gate and its seventeen external
TypeScript/postcss/graphql/parser-input comparisons were not run or claimed
green. No skip, assertion, input requirement or option guard was relaxed.

All twenty-one default-option rules match unchanged production Go on complete
finding, fix and suggestion bytes, normally and under sanitizers. Their control
suites total 692 files and 457 findings. The final three-rule suite has 345 valid
controls, 163 findings and 58,921 exact bytes; sixteen nonsources are excluded
by the independent parser, as before. Both frozen corpora yield zero findings
for these rules, with full 5,318 and 18,485 bytes including file headers compared.
All six React rule sources also match under source Node. Checker-linked emitted
JavaScript still explicitly refuses an unlinked typescript-go library call.
These existing private runners remain outside the shared registry. Options
matrices are not covered by these default-option comparisons.

The merged harness's .a loading, full suggestion serialization, automatic fixes
beside unapplied suggestions and JSX witness mode all passed. Its planted
emitted-JavaScript mismatch was caught. The existing decoded-options test also
passed the new oracle guard, matched Go on all backends, and caught its ignored-
option mutant on Node and native. No owned adapter reached the new guard in
the selected checks; the guard was not bypassed.

Raw HIR matches production Go on 22 controls and 1,046,120 bytes. The full bridge
gate passes 1,600 independently answered positions across four compiler files,
54,982 exact bytes under ASan/UBSan/LeakSanitizer, released-buffer survival,
stale-handle rejection and all seven foundation mutants. Those mutants test
input/output bounds with ASan, retained handles with the stale-handle contract,
wrong positions with Go bytes, missing link opt-in with the refusal expectation,
and missing C frees and region allocation with LeakSanitizer.

Proven lowering, runtime records and record mutants passed, as did release-path
and string-equality checks, registry validation, vet and formatting. Expanded
Node comparisons include proven guards and assertions, last-index behavior,
and all seven new typeof fixtures. The four new typeof constructor, literal,
null and slot-presence mutants passed their intended outer assertions. Cache
observations are retained in the Node oracle log; this is a worker gate, not
an uncached integration gate.

Setup passed in 224 s: Go 0 s, clang 1 s, Node 1 s, submodules 1 s, cache warm
224 s. nproc=5, quota=4 cores, memory=17.6 GB. Logged bounded old reproducible
Go cache files were removed for space; source inputs and logs were retained.
The full current bridge command took 114.92 s.

Every current-run mutant and catcher below is copied from the logs. Rule
mutants compile, exit 0 with empty stderr and differ only in findings. The
complete byte oracle catches each. Definitions and full name mappings remain
in the original reports linked from README.md. Bridge, record and typeof
mutant outer assertions are preserved in their respective logs.

```text
first: === RUN   TestWave26AgreementAndMutants
first: wave_26_test.go:136: operation mutant: exit 0, empty stderr, byte oracle catches byte 535
first: wave_26_test.go:136: provider mutant: exit 0, empty stderr, byte oracle catches byte 2046
first: wave_26_test.go:136: relation mutant: exit 0, empty stderr, byte oracle catches byte 2807
first: wave_26_test.go:146: type-arguments mutant: exit 0, empty stderr, byte oracle catches byte 58
first: wave_26_test.go:181: released-registry mutant: exit 0, required panic 70 catches it
first: --- PASS: TestWave26AgreementAndMutants (146.26s)
next: === RUN   TestWave26NextAgreementAndMutants
next: wave_26_next_test.go:154: listener mutant: exit 0, empty stderr, byte oracle catches byte 1351
next: wave_26_next_test.go:154: render mutant: exit 0, empty stderr, byte oracle catches byte 6484
next: wave_26_next_test.go:154: mock mutant: exit 0, empty stderr, byte oracle catches byte 13857
next: wave_26_next_test.go:170: syntax question mutant: exit 0, empty stderr, byte oracle catches byte 5636
next: wave_26_next_test.go:170: lineage question mutant: exit 0, empty stderr, byte oracle catches byte 57
next: wave_26_next_test.go:170: literal question mutant: exit 0, empty stderr, byte oracle catches byte 6484
next: wave_26_next_test.go:170: transform question mutant: exit 0, empty stderr, byte oracle catches byte 9924
next: wave_26_next_test.go:206: released-registry mutant: exit 0, required panic 70 catches it
next: --- PASS: TestWave26NextAgreementAndMutants (178.92s)
third: === RUN   TestWave26ThirdAgreementAndMutants
third: wave_26_third_test.go:136: eval mutant: exit 0, empty stderr, byte oracle catches byte 3841
third: wave_26_third_test.go:136: extend mutant: exit 0, empty stderr, byte oracle catches byte 11304
third: wave_26_third_test.go:136: function mutant: exit 0, empty stderr, byte oracle catches byte 18330
third: wave_26_third_test.go:150: syntax question mutant: exit 0, empty stderr, byte oracle catches byte 9322
third: wave_26_third_test.go:150: origin question mutant: exit 0, empty stderr, byte oracle catches byte 1299
third: wave_26_third_test.go:186: released-registry mutant: exit 0, required panic 70 catches it
third: --- PASS: TestWave26ThirdAgreementAndMutants (163.51s)
fourth: === RUN   TestWave26FourthAgreementAndMutants
fourth: wave_26_fourth_test.go:142: function mutant: exit 0, empty stderr, byte oracle catches byte 2427
fourth: wave_26_fourth_test.go:142: nonconstructor mutant: exit 0, empty stderr, byte oracle catches byte 7081
fourth: wave_26_fourth_test.go:142: wrapper mutant: exit 0, empty stderr, byte oracle catches byte 9475
fourth: wave_26_fourth_test.go:156: constructor question mutant: exit 0, empty stderr, byte oracle catches byte 556
fourth: wave_26_fourth_test.go:156: origin question mutant: exit 0, empty stderr, byte oracle catches byte 59
fourth: wave_26_fourth_test.go:192: released-registry mutant: exit 0, required panic 70 catches it
fourth: --- PASS: TestWave26FourthAgreementAndMutants (133.48s)
fifth: === RUN   TestWave26FifthAgreementAndMutants
fifth: wave_26_fifth_test.go:171: promise mutant: exit 0, empty stderr, byte oracle catches byte 3172
fifth: wave_26_fifth_test.go:171: regex mutant: exit 0, empty stderr, byte oracle catches byte 12273
fifth: wave_26_fifth_test.go:171: rest mutant: exit 0, empty stderr, byte oracle catches byte 8419
fifth: wave_26_fifth_test.go:187: syntax question mutant: exit 0, empty stderr, byte oracle catches byte 1223
fifth: wave_26_fifth_test.go:187: binding question mutant: exit 0, empty stderr, byte oracle catches byte 58
fifth: wave_26_fifth_test.go:187: pattern question mutant: exit 0, empty stderr, byte oracle catches byte 11033
fifth: wave_26_fifth_test.go:187: comments question mutant: exit 0, empty stderr, byte oracle catches byte 25687
fifth: wave_26_fifth_test.go:223: released-registry mutant: exit 0, required panic 70 catches it
fifth: --- PASS: TestWave26FifthAgreementAndMutants (176.90s)
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
decoded-options: === RUN   TestDecodedOptionsAndMutant
decoded-options: registration_test.go:251: ignored decoded-option mutant caught on Node: case 0 line 2: port "/workspace/wave-26-bridge-scratch/TestDecodedOptionsAndMutant1349730401/001/catch.ts:1:26", Go "fixed\ttry { work(); } catch(e) {}\\u000a"
decoded-options: registration_test.go:251: ignored decoded-option mutant caught on native: case 0 line 2: port "/workspace/wave-26-bridge-scratch/TestDecodedOptionsAndMutant1349730401/001/catch.ts:1:26", Go "fixed\ttry { work(); } catch(e) {}\\u000a"
decoded-options: --- PASS: TestDecodedOptionsAndMutant (51.92s)
```

Quiet alternating whole-process medians for the final three-rule attribute/
display-name suite, with full byte comparisons for every timing sample:

```text
compiler native 6.8293635890004225 Go 0.450544389001152
repository native 0.9850099710020004 Go 0.2335545940004522
```

Native remains slower. These timings cover the final three rules, not a
combined twenty-one-rule runner.

Final explicit all-head fetch pruned stale origin refs and scanned 505 current
refs. The original 197-rule ranking has 25 baseline/main ports and 172 claimed
names, covering every candidate. Zero unclaimed rules remain. Ref SHAs, claim
blobs, ranking and names are in selection-audit.json.gz and the claims audit.
No new rule was claimed; work stops at the requested exhaustion gate.
