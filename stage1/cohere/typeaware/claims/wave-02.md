# Type-aware wave 02

Branch: `codex/typeaware-wave-02`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed rules, positions 4, 5 and 6 after excluding the 26 existing ports from
VOLUME_REPORT.md's combined checker-dependent counts:

| Position | Rule | Compiler | Repository | Total |
| --- | --- | ---: | ---: | ---: |
| 4 | logical-assignment-operators | 143 | 0 | 143 |
| 5 | @typescript-eslint/no-unsafe-return | 126 | 0 | 126 |
| 6 | nexus/correctness-no-nullish-stripping-assertion | 97 | 0 | 97 |

Fetched all origin heads and inspected stage1/cohere claim files and named rule
port files before claiming. No existing claim or named port for these rules was
found. No rule was skipped.

## Continuation 1

Requested on October 7, 2026; same branch, after pushing all original work.
Fetched all 320 origin refs and inspected every typeaware claim blob. Selection
uses VOLUME_REPORT.md's linked validation-volume combined counts descending,
with lexical ties, excluding the ports on origin/codex/tsgo-c-library and
origin/main and rules named in any origin claim file.

The first three eligible rules are reserved here before implementation:

| Rule | Compiler | Repository | Total |
| --- | ---: | ---: | ---: |
| nexus/correctness-no-collection-misuse | 0 | 0 | 0 |
| nexus/correctness-no-discarded-outcome | 0 | 0 | 0 |
| nexus/correctness-no-discarded-pure-result | 0 | 0 | 0 |

All preceding ranked rules were already ported or named in an origin claim.
These three names occur only in inventory/skip-types records on the base and
main, with no native implementation or typeaware claim. Existing reservations
remain excluded even where their claim documents describe incomplete work.
Positive controls and byte-oracle mutants are required despite zero volume.

Continuation status: all three reserved Nexus rules implemented and validated in
cba8883eebe65ae8e3bef3fdb3c54455b00fd1e8, pushed. Later origin fetches revealed
concurrent continuation claims; WAVE_02_CONTINUATION_REPORT.md records them.
Ahra's correction arrived after this reservation and validation; no further
rules were claimed. The original three wave-02 rules were already done and pushed.

## Continuation 2

All six earlier reserved rules are implemented, tested and pushed through
84ecd45cc047337244e0257934bc5aa69d8a0406. After fetching all 325 origin refs,
checked claim contents (including compressed evidence blobs), and main/bridge
implementations against VOLUME_REPORT.md's combined checker-dependent ranking.
The first three eligible rules, all zero compiler and repository volume, are:

1. nexus/correctness-no-process-exit-after-output
2. nexus/correctness-no-uncleared-race-timeout
3. nexus/correctness-require-blocking-standard-streams

These names occur only in inventory/skip-types records on main and the base,
and in none of the fetched typeaware claim files. They are reserved for this
branch before implementation. New native sources remain .a. Existing shared
registration generators and test harness files will not be edited.

Continuation 2 completed in 42724911 and 774a422e, tested and pushed. The report
WAVE_02_CONTINUATION_2_REPORT.md records all nine completed wave-02 ports across
the three reservations, canonical byte comparisons, sanitizers and mutants.
A later fetch found concurrent claims on 17 other branches; the initial tip audit
contains none of these three names. No further reservation was made.

## Continuation 3 claim

Fetched all origin heads on 2026-10-07 after verifying all nine prior ports
were pushed through 0a3a176e. The 197-rule combined compiler/repository
ranking linked by VOLUME_REPORT.md, with descending totals and lexical ties,
was checked against 355 origin refs and 92 distinct claim blobs. The first
three rules absent from claims and from native rule sources on origin/main
and origin/codex/tsgo-c-library are:

- prefer-regex-literals (combined volume 0)
- prefer-rest-params (combined volume 0)
- react-hooks/exhaustive-deps (combined volume 0)

These three are reserved for wave 02. No implementation precedes this claim
commit and push. Base bridge tip: 5afbdb83; main tip: ef3d907e.

### Continuation 3 selection correction

The first scan used substring matches. An exact rule-token audit found that
`prefer-promise-reject-errors` had been incorrectly excluded by a claim for the
different `@typescript-eslint/prefer-promise-reject-errors` rule. The correct
three, against the same initial origin snapshot, are:

- prefer-promise-reject-errors (combined volume 0)
- prefer-regex-literals (combined volume 0)
- prefer-rest-params (combined volume 0)

The core promise-rejection rule is reserved by this correction before any code
for it. The prior exhaustive-deps reservation is released; its already-written
native default implementation and JSX parser failures will be retained as
reviewable supplemental work, not presented as a completed extra allocation.

### Correction refresh and skip

A fresh origin fetch after the correction found exact core-rule reservations on
waves 07, 18, 20, 24 and 26. Therefore `prefer-promise-reject-errors` is skipped
and this attempted reservation is released before any implementation was written.
Wave 01 also mentions skipping it, and wave 04's compressed audit lists its name;
those two mentions are not represented as worker allocations.

The two eligible original reservations, prefer-regex-literals and prefer-rest-params,
remain this wave's implemented allocation. The supplemental exhaustive-deps
implementation remains released and is limited by shared JSX parsing.

## Continuation 4 claim

All prior allocated ports and evidence are pushed through 0ea841d4. Fetched
417 origin references and checked 94 distinct blobs beneath remote claims.
Compressed selection inventories are evidence, not reservations; exact-name
checks against claim documents leave these first three eligible ranked rules:

- react/button-has-type (0 compiler, 0 repository)
- react/checked-requires-onchange-or-readonly (0 compiler, 0 repository)
- react/display-name (0 compiler, 0 repository)

Main and the bridge branch mention these only in count inventories and
skip-types records, not native implementations. They are reserved before code.
The snapshot and document matches are in wave_02_continuation_4/selection-audit.json.
Known shared JSX parsing gaps will be disclosed; other behavior will be ported
inside this rule directory without shared harness or generator changes.
