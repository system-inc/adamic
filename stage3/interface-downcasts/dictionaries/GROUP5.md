Built nine more direct dictionary container/read pairs, including mixed named/index-signature CompilerOptions.paths; 30 new compiler-source controls.
Commits: this group follows ce4eeaa4 on codex/views-dictionaries; no shared implementation hook changed.
Checks: uncached full dictionary oracle passed in 29.744s; candidate reproduction and git diff --check passed.
Mutants: parent-container bypass compiled and lost the named refusal in C and JS; skip-check, wrong-shape, dropped-transitive-check and producer-certificate mutants reran successfully.
Not covered: optional dictionary receivers, rich element selection, enumeration and uncertified writes; 18 candidate pairs / 74 candidate reads remain, exact reachability unmeasured.

Whole-family working estimate remains October 11, 2026, 23:00 UTC, as accepted.
Primitive/nullish element selection is assigned to lane 4; object/array and
TsConfigSourceFile alternatives to lane 4b. Lookup, absence and enumeration
remain here. No code from another lane was required for this group.

## Certified read contracts

| Candidate owner | Field | Candidate reads |
| --- | --- | ---: |
| ParsedCommandLine | watchOptions | 9 |
| CompilerOptions | paths | 7 |
| VersionPaths | paths | 6 |
| IncrementalMultiFileEmitBuildInfo | options | 6 |
| IncrementalBundleEmitBuildInfo | options | 3 |
| ParsedCommandLine | wildcardDirectories | 2 |
| IncrementalBuildInfo | options | 2 |
| ReusableBuilderProgramState | compilerOptions | 1 |
| BuilderProgramState | compilerOptions | 1 |

These are original representative read-contract fixtures, not copied compiler
interface definitions. Container checks prove object kind, initialization and
optional absence without demanding unread rich children. They do not certify
CompilerOptionsValue or TsConfigSourceFile element selection.

Every shape has a valid value, a wrong number and an absent property. Optional
absence agrees with Node; required absence stops with the named initialization
message. All valid/allowed-absent controls agree with Node in sanitized C,
release C and the JavaScript backend, and undergo the leak check. Wrong/missing
required values pin their entire exit-70 diagnostic in all three compiled runs.

CompilerOptions.paths specifically carries both its named field and a rich
string index signature. Its controls cover a missing container, absent key,
wrong container, wrong array entry and wrong nested array element. This proves
that lookup uses the selected named field contract rather than trying to admit
the rich index union up front. The exact named wrong-container message is:

    adamic: panic: field read failed: view.paths; expected MapLike<string[]> | undefined, found number

The existing paths fixtures used the ordinary object-field dispatch; this group
uses dictionary-key dispatch for the mixed named/index-signature receiver.

## Mutant evidence

The new parent-container bypass replaces generated checked field reads with a
safe object-valued native adapter and an unchecked JavaScript field read. Both
mutants compile and run successfully with exit 0, printing object and number
respectively; both violate the control's pinned exit-70 refusal. The first C
adapter attempt used a nonexistent pointer typedef and failed compilation. That
attempt is not counted as evidence; the corrected semantic run is recorded in
the final log. This changes test fixtures only, not compiler checks.

Existing dictionary skip-check, accept-wrong-shape and drop-transitive-check
mutants run in both C and JS for fixed objects and shared record producers.
The independent producer storage-certificate mutant also reran. Every final
mutant check passed. No full repository gate is claimed: previously reproduced
baseline failures remain documented in GROUP3.md.

## Remaining inventory audit

The string row (14 reads) and Path row (3) have numeric string-indexing sites,
not string-key object dictionary lookups. The any row (20) is erased dynamic
access, not a declared element contract. Four small rows are finite named-field
cache accesses: IndexedAccessType (3), IterationTypes (2),
IterableOrIteratorType (1), optional signature cache (1). These 44 candidate
reads need explicit disposition evidence; none were removed or credited here.
The remaining inventory therefore stays the accepted static candidate table.

Commands:

    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewDictionary' -count=1 -v
    python3 stage3/interface-downcasts/dictionaries/rank-lazy-demand.py --check
    git diff --check

Raw final gate: logs/group5-oracle.log. Initial failed mutant build:
logs/group5-initial.log. Individual container and mixed-path runs are also kept.
