Built: a full-corpus static read census seeded by all 2,936 audited downcast targets, with per-pair declared families and exact source witnesses.
Commits: measured after parser checkpoint c898009b; Lane 1's pending lazy cast merge remains Lane 1's responsibility.
Commands: Node TypeScript 6.0.3 census passed, matched every audited span/text and required zero diagnostics; 2,825 pairs and 25,848 reads.
Mutants: no compiler change here; c898009b caught seven return-family, two deferred-dispatch and three runtime/signature mutants.
Not covered: allocation/alias or call-result/argument flow, complete per-read backend certification, or an unlocked whole-source lowering count.

The source remains TypeScript commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Every pair in read-census/pairs.json records the receiver checker type id/name,
field, declared type/families, read count, file/line/column witnesses and its cast
or transitive-read provenance. Propagation follows actual source field/element
reads, consumed array elements and union members for narrowed receivers. It never
walks every property of a reachable type. Before the static viewed-type filter,
there are 67,118 source field reads.

These are static potential view reads, not runtime execution counts. Equal static
types elsewhere are conservatively included. Partial family counts are demand;
they do not claim every read is blocked. Families overlap; each pair/read counts
once within a family or lane. Array intrinsics are separate from array-interface
own fields such as pos/end. The intrinsic's declared family is retained beside
its read-contract families.

| Read-contract family | Lane | Distinct pairs | Read sites | Coverage |
| --- | --- | ---: | ---: | --- |
| object or interface contracts | lane 1 | 1101 | 11569 | supported subset |
| scalar contracts | already supported scalar | 982 | 8839 | supported subset |
| nullish members | lane 1 | 865 | 6276 | missing |
| optional properties | lane 1 | 754 | 5728 | missing |
| array contracts | lane 2 | 334 | 3189 | partial; per-read proof required |
| object union | lane 4 or shared union admission | 228 | 2468 | partial; per-read proof required |
| array element or consumer reads | lane 2 | 251 | 1602 | partial; per-read proof required |
| callable contracts | lane 2 | 308 | 1503 | partial; per-read proof required |
| array intrinsic contracts | lane 2 | 179 | 794 | partial; per-read proof required |
| intersection field contract | other unresolved | 45 | 769 | missing |
| mixed primitive union | lane 4 or shared union admission | 23 | 538 | missing |
| object plus primitive union | lane 4 or shared union admission | 42 | 181 | missing |
| dictionary contracts | other unresolved | 18 | 174 | missing |
| any field contract | other unresolved | 55 | 104 | missing |
| array property reads | lane 2 | 30 | 72 | partial; per-read proof required |
| dictionary reads | other unresolved | 14 | 66 | missing |
| nominal class fields | other unresolved | 1 | 42 | missing |
| tuple contracts | lane 2 | 6 | 9 | missing |
| unknown field contract | other unresolved | 1 | 2 | missing |
| generic field contract | other unresolved | 1 | 1 | missing |

| Lane with partial or missing demand | Distinct pairs | Read sites |
| --- | ---: | ---: |
| lane 2 | 920 | 6362 |
| lane 1 | 866 | 6279 |
| lane 4 or shared union admission | 292 | 3186 |
| other unresolved | 130 | 1125 |

Lane 2's remaining families are array/tuple element contracts, producers,
mutations/consumers, array own-property reads and callable implementation
certification. The current conservative admission refuses 35 pairs / 128 reads
across join, indexOf, includes, sort, splice, unshift, concat, lastIndexOf and
shift. Join is largest at 5 pairs / 49 reads, versus 6 tuple pairs / 9 reads.
Array own fields account for 30 pairs / 72 reads. The broader contract/callable
rows are partial demand, not exact missing-read counts. Existing checks do not
certify every producer or element family, so that demand remains recorded.

| Array intrinsic member read | Distinct pairs | Read sites |
| --- | ---: | ---: |
| push | 26 | 352 |
| map | 33 | 82 |
| slice | 25 | 82 |
| join | 5 | 49 |
| forEach | 20 | 37 |
| indexOf | 11 | 34 |
| find | 2 | 29 |
| some | 16 | 28 |
| pop | 6 | 23 |
| filter | 9 | 23 |
| includes | 3 | 16 |
| sort | 4 | 11 |
| every | 5 | 8 |
| splice | 4 | 7 |
| unshift | 4 | 5 |
| concat | 2 | 3 |
| lastIndexOf | 1 | 2 |
| findIndex | 2 | 2 |
| shift | 1 | 1 |

Contract laziness is not complete on c898009b. viewObjectFields/viewArrayFields
and viewContract still visit declared descendants eagerly. Array/callable use
scans remain conservative across the module graph. An unread unsupported
member or unrelated same-name use can still refuse a cast. The parser changes
do not solve that policy. Lane 1 owns the cast merge and lazy admission, and
will merge this tip before pushing. Lane 2 must merge that SHA, leaving the cast
conflict to Lane 1. Afterward, unsupported arrays/callables must stop only at a
reached field/consumer read, naming its field and family. No completed lazy
admission is claimed here.

Reproduce:

```sh
NODE_PATH=/workspace/scratch/lane2-api/node_modules node \
  stage3/interface-downcasts/lane2/view-read-census.cjs \
  /workspace/scratch/lane2-ts-pin \
  stage3/interface-downcasts/blocking-families-sites.json \
  stage3/interface-downcasts/lane2/read-census > /tmp/lane2-read-census.log 2>&1
```

The source checkout includes the generated diagnostic map and Node declarations
needed for a zero-diagnostic program. The script asserts its TypeScript version
and source commit before measuring. summary.json holds these tables; pairs.json
holds every witness. A consistency check verified pair/read totals, provenance
on every pair and each read_count against sites length. Census output is retained
in read-census/census.log. Whole-corpus unlocked sites are not inferred. The prior
0/2,936 complete-source observation remains a baseline, not a new lowering run.

After this table: integrate native storage/producer/mutation paths, then join
before tuples, with Node controls and semantic mutants; callable contracts follow.
Working estimates remain October 9 12:00 MDT for native arrays and October 10
12:00 MDT for callable contracts. These depend on Lane 1's lazy cast handoff and
are estimates, not completion claims. This measurement-only commit needs no full
gate rerun; the preceding compiler checkpoint's focused gate and 12 mutants passed.
