# u030 plan fixed before mutants
Base: 8171b3173bdbfce1f7982d3c4f731279307ece37.
Code under test: Adamic lowering and internal semantic/storage proofs. Complete observed function list in reached-functions.md, 422 functions. Anonymous closures and zero-statement entries are limited by Go coverage instrumentation.
Oracle: all assigned rows use self-written IR/count/kind expectations or acceptance/refusal expectations, with no cited outside authority. None runs an external semantic comparator. Broad stop-type expectations can accept a different later refusal.
Entries: Lower for 12 rows; clockGenericReturnsT01 directly for RejectsNullBeforeBody and RejectsIndexBeforeBody. P1 applies only to former; P2 only to latter.
Families: distinct test bodies assert distinct properties. Shapes checks Lower; BeforeBody checks signature proof directly. No input-only wrappers grouped.
Fixed menu: 20 production mutants in menu.json, from reached code. Condition flips, changed constants/bounds, dropped statements. M03 drops whole loop. M10 drops guard terms, never edits the checker. P1/P2 are empty-answer probes, excluded from production kills.
Baseline with both inventories enabled fits at 32.623 binary seconds. Run the whole package per mutant, narrowing only for cooked mutants. Every compiler mutant gets ADAMIC_BUILD_CACHE_DIR=/tmp/u030/cache/<id>. Assigned tests do not build native products.
Every standalone diff omits selector scaffolding and is vetted against the starting base. On panic rerun all 14 assigned rows alone; outside unfinished results remain unknown.

Probe validation correction before matrix: unconditional return was rejected by vet as unreachable. P1/P2 now use tautological pointer guards (program==program and result==result) to preserve the unconditional empty value for every input while satisfying vet. No production menu changed. The three assertOverrideParameterRefusal wrappers are one outside-unit family, recorded in families.json.
