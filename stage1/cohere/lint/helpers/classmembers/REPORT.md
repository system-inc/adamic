# Classmembers package validation

Claim `2a7d5c0135769fe9bed15cd32edb0c0e42134ca2`; original base
`cd56db1dd7ab7eda7a5f492804dcd1849a1494aa`; triage `d7ab0bc4`.
Fresh ports; no partial production implementation existed to reuse.
The initial three-helper partial was `132effbcdd2b55978e65917875d2a99584b9f880`.

The missing shared property.NameTagged was built under explicit follow-up
authorization, from origin/lint-helpers/property at a72b9f63e66263ea34419bae20d282968e3261b6,
and pushed as its own finished unit at
`ecfd7ffc1e4d37c8b1933c8b339913e2cd3c9a59`. Its [report](../property/testdata/tagged/REPORT.md)
records 1,213 original Go calls, 12,448 identical bytes and its compiling,
running mutant caught on Node, emitted JavaScript and sanitized native.
It is imported by KeyOf; no private copy of Name or NameTagged is built.

[Selected log](testdata/selected.log): all five classmembers helpers agree on
532 actual calls from both unchanged consuming upstream suites, 5,817 output
bytes identical on source Node, emitted JavaScript and ASan/UBSan native.
Per-helper counts: MemberName 144, IsOverloadSignature 144, IsAccessorKind 43,
KeyOf 130, ForEachDuplicate 71. Core and TypeScript captures contain 40 and 35
passing test events respectively, with no failures or skips. These counts
include parents and invocations, rather than unique rule fixtures.
[Coverage](testdata/coverage.json) and [observations](testdata/consumer-calls.json)
are regenerated through an overlay over unchanged original Go bodies.

All five semantic helper mutants compile, execute and disagree with Go on
all three backends. The selected package passes in 35.726 seconds.
The [core proof rule](../../rules/no-dupe-class-members/REPORT.md) matches all
37 unique upstream cases, two owned witnesses and applicable inherited rows;
its static-flag mutant is caught on all three runtimes. The TypeScript extension
is now helper-ready and remains with its earlier claimant. Go's first-accessor
and bodyless-abstract quirks are preserved and named in the README and witnesses.
No shared dependency or language gap remains.

[Complete helpers gate](testdata/helpers-full.jsonl) and
[summary](testdata/helpers-summary.json): all five packages PASS, 50 passing
test events, zero failures and zero skips. Wall time 204.876 seconds.
The WASI SDK, clean pinned TypeScript corpus, benchmark input and one fresh
directory for both profile variables were supplied.

[Complete lint gate](testdata/lint-full.jsonl) and
[machine summary](testdata/full-summary.json): package PASS; 131 passing
test events, zero failures, one skip, TestCheckerBridgeRefusalPending (the
unlanded TSGoError bridge). No missing-input skips. Wall time 1,102.107 seconds;
package elapsed 1,099.927 seconds; nproc 5. Load before 3.629 / 2.188 / 2.006;
after 3.099 / 4.432 / 3.825. All required inputs are recorded in the summary;
both profile variables share one fresh directory. Registry generation passed.
No production changes were made after the selected checks began.
