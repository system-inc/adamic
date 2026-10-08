Built: construction presence, V8 first-write key order, inline NodeArray extras, honest undefined observations and checked scalar uses toward step 24.
Commits: rebased contracts onto rehearsal 3d0620c6; preserved published b95d352b through ancestry commit 8ee2057f; delivery SHA is reported after push.
Commands and outputs: focused source oracles, native ASan/UBSan/leaks, JavaScript, counts recording, a-check and go vet pass; logs are in ../evidence.
Mutants: three actual runtime order/JSON mutants disagree with Node; removed declared-T and pending-null guards fail their pins; four completion and four slot mutants fail; payload mutant differs in both backends.
Uncovered: rehearsal results pending codex/placeholder-nonnull; observable null! remains NotYet; Program-region lifetime pending step 06 #7g4qv2b; no whole-parser or check-erasure certificate.

The source reductions run through production lowering, not a stub. The additional
hand-built IR tests isolate runtime mutations. Rehearsal-dependent passing tests
are pending integration of codex/placeholder-nonnull, as required by build-ahead.

The null-placeholder witness observes null followed by undefined on Node. The
rehearsal's literal placeholders still lack observable null transport. Lowering
returns NotYet when such a field collides with a factory field; removing that guard
is caught. It never reports that null! successfully observes undefined.
