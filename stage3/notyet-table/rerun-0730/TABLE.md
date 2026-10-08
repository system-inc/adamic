# 2026-10-08 morning lowering root census

**Before 10,083 roots; after 6,703. Retired 3,961, newly exposed 581, surviving 6,122; net retirement 3,380.**

## Frozen inputs and compiler pins

Fetched and froze tips at `2026-10-08T10:40:28Z` (04:40:28 MDT, 2026-10-08). The [81-file manifest](source-manifest.json) is byte-identical to the original table manifest; all source byte counts and SHA-256 hashes verified before and after the runs.

| Input | SHA |
| --- | --- |
| Before: origin/area/compiler | `b68b2fe1af2dd6ad369ad8dfbfedf66be5a42861` |
| Statics tip | `1db2b32429c7edd8e4b6cfcd36ac98cea8df4a9a` |
| Binary tip | `a8ad5bf0778fe8fd9f30b70bbe3eb807864e6b2a` |
| Checked non-null area change | `c41c0e062e99da37820f822968d4df1b48cdaee7` |
| After scratch merge | `0e5661e4243af95ae3247d066de14ccb2b582a94` |
| Frozen measurement tooling | `57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7` |

The mandatory scratch integration merged statics and binary cleanly. The checked non-null commit is already an ancestor of the after merge: merging it reports `Already up to date.` Both builds use cohere `7945d102a6c18dd36adf9114a758ce646e8b2359` and TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. Compiler merges remain on scratch branches; this report commit changes no production compiler source.

## Measured accounting

| Observation | Before | After |
| --- | ---: | ---: |
| Unique lowering NotYet sites before removing echoes | 14,458 | 10,750 |
| Proven echo-only sites removed | 4,375 | 4,047 |
| NotYet root sites | 10,083 | 6,703 |
| Root reasons | 478 | 472 |
| Untagged reads retained conservatively | 92 | 108 |
| Other governing roots referenced by echoes, outside the NotYet count | 321 | 409 |

| Signature comparison | Sites |
| --- | ---: |
| Retired: before only | 3,961 |
| Newly exposed: after only | 581 |
| Surviving: present in both | 6,122 |
| Net retirement: retired minus newly exposed | 3,380 |

Root signatures are exact `(kind, where, reason, text)` tuples with only the common source-directory prefix normalized. Different attempts and units at the same diagnostic signature are deduplicated, matching the original table. A site is removed as an echo only when every observation carries symbol-linked `blocked_by` provenance. Each tag must point at a real, untagged lowering finding in the same attempt; non-NotYet governing roots are kept in separate ledgers. No worker-reported credits enter these counts.

The entry-root program is checker-rejected in both runs. Both retain identical checker metadata, 324 diagnostic sites, 10,551 attempted units and 147 units split around checker-diagnosed bodies. Unit eligibility changes: zero. Full mode is guarded by `LATENT_ASSERT_NO_OUTPUT=1`; these are capability measurements, not a claim that the compiler accepts or runs tsc.

Of the 581 after-only signatures, 359 appear at locations with no earlier lowering NotYet; 207 replace a different root signature at the same location, and 15 replace a different tagged echo signature there. All 581 count as newly exposed under the requested signature definition. A retired signature is an observed disappearance, not proof of completed compilation; an after-only signature by itself cannot prove the precise causal path through the new lowering. The [observation transition ledger](newly-exposed-observations.csv) preserves these distinctions.

## Top 10 newly exposed root reasons

| Exact reason | Retired | Newly exposed | Surviving |
| --- | ---: | ---: | ---: |
| passing union of differently held members to a function value | 0 | 154 | 2 |
| an ElementAccessExpression | 4 | 65 | 24 |
| assigning to an Identifier | 23 | 30 | 142 |
| a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 0 | 21 | 184 |
| a call through ?. (an optional call) | 0 | 20 | 125 |
| a function value returning union of differently held members | 0 | 20 | 25 |
| an array of T | 0 | 20 | 23 |
| an array of never | 0 | 18 | 72 |
| a value of type __String | 0 | 17 | 170 |
| for...of over an object | 0 | 16 | 108 |

[Every root signature and its three numbers](before-to-after-signatures.csv). [Every reason and its three totals](before-to-after-reasons.csv). [Machine-readable comparison](comparison.json). The CSV includes reasons with zero newly exposed sites.

## Reproduction and validation

[Commands, build workarounds, complete mutant outcomes and setup timings](EVIDENCE.md). [Independent SQL audit](audit.log.txt). Both positive poison witnesses pass; dropping tagging fails the expected root-link assertion on both builds. All ten accounting and provenance-structure mutants are caught, including swapped categories, dropped survivors, echo inflation, changed source hashes, missing provenance roots, changed checker diagnostics and unit eligibility. Both no-output guard mutants are caught.

[Before full annotated JSONL](before/full.jsonl.gz), [before census log](before/census.log.txt), [before roots](before/roots.csv), [before root-to-echo edges](before/echoes.csv), [before run identity](before/run.json).

[After full annotated JSONL](after/full.jsonl.gz), [after census log](after/census.log.txt), [after roots](after/roots.csv), [after root-to-echo edges](after/echoes.csv), [after run identity](after/run.json).

## Optional all-tip run

Not measured. The pinned all-tip scratch integration stopped at `codex/notyet-predicates` (`0ac2a1f5afdf11fc688eebace83cf39857271e8f`), which conflicted in 24 files across view-contract, closure, field-layout and backend implementations. Resolving that overlap requires a separate compiler integration with its semantic controls. The unresolved merge was aborted; no partial all-tip counts or credited reports substitute for a third census. [Frozen tips](pins.json), [merge attempt and resolved hunks](all-merge-attempt.json), [remaining conflicts](optional-status.json).

Remaining coverage limits: checker-diagnosed bodies and statements behind failed compound boundaries are not fully lowered; untagged reads stay conservative. This report adds no registered oracle fixture, so no fixture count refresh or full gate was run.
