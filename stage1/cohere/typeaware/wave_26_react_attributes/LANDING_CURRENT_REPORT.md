Built: twenty-one ports rebased and re-greened on lint area b84a9d93; no new rule source changes.
Commits: validated rebased tip 30b776687eea02c96f5fc4ac3cf160200b6a33ef, containing current main c7991b900362796aefd111474e65eb5398e91953.
Checks: all twenty-five current-base commands passed, with zero selected test skips; both corpora, sanitizers, merged harness, bridge and expanded Node oracle.
Mutants: all twenty-one rule mutants caught by complete bytes; question, SSA, handle, record, bridge foundation and emitted-JavaScript mutants passed.
Not covered: full repository gate and its seventeen external-input checks, shared registration for private runners, checker-linked emitted JavaScript or every option matrix.

The area rebase completed without conflicts. Main advanced during the first
validation from 39638d9e to c7991b90, adding proven lowering and runtime records.
This unit rebased again onto area b84a9d93 and reran the complete owned gate,
including those new lowering, record and Node fixtures. Current main remained
c7991b90 at the final fetch. The area advanced to b4691483 during validation;
this report claims the tested b84a9d93 area base, which contains current main.
No compiler, shared harness, registration generator or rule implementation was
edited by this landing unit. No new rule was claimed.

The twenty-five exact commands, exits and durations are in
[the command ledger](validation/landing-b84a9d93/runs.json.gz), with environment
inputs and log destinations in [the runner](validation/landing-b84a9d93/run.py.gz).
All test output went directly to files. ADAMIC_TYPESCRIPT_SOURCE points to the
pinned v6.0.3 checkout, with the frozen 77-file compiler and 287-file repository
manifests supplied throughout. No selected test skipped, as recorded in
selected-skip-check.log.gz. The full repository gate and its seventeen external
TypeScript/postcss/graphql/parser-input comparisons were not run or claimed
green. No skip, assertion or input requirement was relaxed, removed or edited.

All twenty-one default-option rules again match unchanged production Go on
complete finding, fix and suggestion bytes, normally and under sanitizers.
The separate control suites total 692 files and 457 findings. The final
attribute/display-name suite has 345 valid controls, 163 findings and 58,921
exact bytes. Six React rule sources additionally match under source Node.
Their checker-linked emitted-JavaScript path still explicitly refuses an
unlinked typescript-go library call. The merged harness's .a loading and
suggestion serialization are independently green and are not named as blockers.
Shared registration for these existing private runners remains outside this unit.

The new proven-lowering tests passed. RecordsAgainstNode, RecordMutants and
RecordReadMutants passed, alongside runtime release-path and string-equality
checks. The expanded Node oracle passed proven guards, class guards, assertions,
satisfies and upcasts, plus the last-index fixture: 758 output bytes agree on
Node, emitted JavaScript, release and sanitizers. The ordinary one-byte oracle
mutant was also caught. Registry and all five targeted merged-harness checks
passed; its intentional nested emitted-JavaScript mismatch is caught by its
outer test. Suggestions remain unapplied beside automatic fixes.

Raw HIR agrees on 22 controls and 1,046,120 production bytes. The full current-
base bridge gate passed in 126.33 seconds: 1,600 independently answered positions
over four compiler files and 54,982 exact bytes under ASan/UBSan/LeakSanitizer.
All seven foundation mutants passed their expectations: input and output length
mutations trigger ASan; retained handles fail the stale-handle assertion; wrong
positions fail production Go bytes; removing link opt-in fails the refusal
expectation; missing C frees and region heap allocation trigger LeakSanitizer.

The earlier d65a8f93 run exhausted disk while constructing the final region
mutant compiler, after healthy bridge answers and preceding mutants passed.
Its exact error was `mkdir /workspace/wave-26-bridge-scratch/go-build3386495449:
no space left on device`. Logged cleanup of reproducible cache entries restored
4.0 GB; the complete unchanged bridge retry passed in 100.37 seconds, including
that region mutant. The original failure, retry, command ledger and resumed
runner are preserved in validation/landing-d65a8f93. This was an infrastructure
interruption, not a skipped or relaxed check. The newer current-base gate above
passed all twenty-five commands on its first attempt.

Setup for the final base passed: Go 0 s, clang 0 s, Node 0 s, submodules 0 s,
cache warm 234 s, total 234 s; nproc=5, quota=4 cores, memory=17.6 GB. Vet,
compiler formatting and owned-question formatting logs are empty. The required
one-line shared facts.go registration cases remain unchanged. Named worker
binaries and logged bounded sets of old reproducible Go cache entries were
removed for space; all source inputs and logs were retained.

Every rule mutant below compiled, exited 0 with empty stderr and differed only
in diagnostic bytes. The independent production Go comparison caught it.
Definitions and complete rule-name mappings remain in the reports linked from
README.md; these lines record every current-run mutant and its observed catcher.

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

New record mutants and their successful outer assertions are recorded in
records.log.gz:

```text
--- PASS: TestRecordsAgainstNode (23.13s)
--- PASS: TestRecordsAgainstNode/semantics (0.25s)
--- PASS: TestRecordsAgainstNode/prototypes (0.08s)
--- PASS: TestRecordsAgainstNode/reads (0.06s)
--- PASS: TestRecordsAgainstNode/references (0.45s)
--- PASS: TestRecordsAgainstNode/iteration (0.07s)
--- PASS: TestRecordsAgainstNode/numeric (1.24s)
--- PASS: TestRecordsAgainstNode/workload-1000000 (10.58s)
--- PASS: TestRecordsAgainstNode/bench-1000 (0.08s)
--- PASS: TestRecordsAgainstNode/missing-get/constructor (0.01s)
--- PASS: TestRecordsAgainstNode/missing-has/constructor (0.01s)
--- PASS: TestRecordsAgainstNode/missing-get/__defineGetter__ (0.01s)
--- PASS: TestRecordsAgainstNode/missing-has/__defineGetter__ (0.01s)
--- PASS: TestRecordsAgainstNode/missing-get/__defineSetter__ (0.01s)
--- PASS: TestRecordsAgainstNode/missing-has/__defineSetter__ (0.01s)
--- PASS: TestRecordsAgainstNode/missing-get/hasOwnProperty (0.01s)
--- PASS: TestRecordsAgainstNode/missing-has/hasOwnProperty (0.01s)
--- PASS: TestRecordsAgainstNode/missing-get/__lookupGetter__ (0.01s)
--- PASS: TestRecordsAgainstNode/missing-has/__lookupGetter__ (0.00s)
--- PASS: TestRecordsAgainstNode/missing-get/__lookupSetter__ (0.01s)
--- PASS: TestRecordsAgainstNode/missing-has/__lookupSetter__ (0.01s)
--- PASS: TestRecordsAgainstNode/missing-get/isPrototypeOf (0.01s)
--- PASS: TestRecordsAgainstNode/missing-has/isPrototypeOf (0.01s)
--- PASS: TestRecordsAgainstNode/missing-get/propertyIsEnumerable (0.01s)
--- PASS: TestRecordsAgainstNode/missing-has/propertyIsEnumerable (0.01s)
--- PASS: TestRecordsAgainstNode/missing-get/toString (0.01s)
--- PASS: TestRecordsAgainstNode/missing-has/toString (0.01s)
--- PASS: TestRecordsAgainstNode/missing-get/valueOf (0.01s)
--- PASS: TestRecordsAgainstNode/missing-has/valueOf (0.01s)
--- PASS: TestRecordsAgainstNode/missing-get/__proto__ (0.01s)
--- PASS: TestRecordsAgainstNode/missing-has/__proto__ (0.01s)
--- PASS: TestRecordsAgainstNode/missing-get/toLocaleString (0.01s)
--- PASS: TestRecordsAgainstNode/missing-has/toLocaleString (0.01s)
--- PASS: TestRecordsAgainstNode/proto-assignment (0.01s)
--- PASS: TestRecordReadMutants (28.73s)
--- PASS: TestRecordReadMutants/prototype-membership-restored (10.89s)
--- PASS: TestRecordReadMutants/missing-read-silent (9.06s)
--- PASS: TestRecordReadMutants/own-read-checked-as-missing (8.78s)
--- PASS: TestRecordMutants (51.89s)
--- PASS: TestRecordMutants/indices-in-insertion-order (10.84s)
--- PASS: TestRecordMutants/uint32-max-as-index (9.18s)
--- PASS: TestRecordMutants/deleted-key-iterated (8.92s)
--- PASS: TestRecordMutants/overwrite-key-leaked (7.67s)
--- PASS: TestRecordMutants/stored-key-freed (6.95s)
--- PASS: TestRecordMutants/own-slot-null-read (8.32s)
```

Quiet alternating whole-process medians for the final three-rule attribute
and display-name suite, with complete byte comparisons for every sample:

| Corpus | Native | Production Go |
| --- | ---: | ---: |
| Compiler | 8.358071 s | 0.617203 s |
| Repository | 1.205950 s | 0.240460 s |

Native remains slower. These are three-rule timings, not a combined twenty-one-
rule runner. Exact samples and command timings are preserved with the evidence.

The final explicit all-head fetch pruned stale remote-tracking refs and scanned
463 current origin refs. In the original 197-rule volume ranking, 25 baseline/
main ports and 172 claimed names cover every candidate: zero unclaimed rules.
All ref SHAs, claim blobs, names and ranking are in selection-audit.json.gz.
Work stops at the requested exhaustion gate without another claim.
