# Pair latent censuses

Use this tool when comparing two latent streams measured on the same adapted source bytes, especially when a compiler change relocates a refusal from a property name to its access. Run `python3 stage3/census/pair/pair.py BASE.jsonl.gz HEAD.jsonl.gz ADAPTED --typescript stage3/api/node_modules/typescript --json delta.json --report report.md` from the checkout that supplied the tree's pinned stock TypeScript 6.0.3 package. Plain JSONL also works. No compiler build is needed. Raw findings retain the meter's unique `(kind, source identity, reason, text)` count. A distinct read uses the source path, file SHA-256, UTF-8 byte start of the outermost contiguous property/element access chain, and exact reason; a name and its enclosing access count once even if both are reported on one side. An element access argument starts a separate chain: `a[b.c]` has reads at `a` and `b`, while `a.b.c` and `a[i].b` each remain one receiver chain. Locations use the census's one-based UTF-16 columns. Separate chains and separate reasons remain separate. Both views appear side by side; neither is proof of successful compilation. Only lowering-phase events define unit first stops. Headers, unit identities and attempted eligibility must agree. The caller must supply the identical adapted bytes used for both streams; the JSON records their hashes.

`lowered` means the old finding disappeared and no reason remains at that source identity. `moved` means another reason remains there. `new_findings` includes newly exposed reasons at already blocked sites; `new_sites` counts identities with no base finding. Regressions follow the earlier pairing scripts: a previously finding-free attempted unit now stops, a new panic appears, or a previously finding-free refusal-scan site now stops. All findings, reason transitions, first-stop changes, totals and panics are retained in the JSON.

The complete historical streams give these open-tag results:

| Comparison (head against base) | Raw lowered | Distinct lowered | Distinct moved to another reason | Distinct new findings | Regressions |
| --- | ---: | ---: | ---: | ---: | ---: |
| qeanycw, `5b181782` against `b03fe07e` | 74 | 70 | 2 | 0 | 0 |
| kind-predicates, `b03fe07e` against `df84b974` | 239 | 181 | 5 | 1 | 0 |

Thus qeanycw clears 72 distinct open-tag reasons: 70 reads have no remaining finding and 2 retain another reason. This is the exact result of access-chain normalization, refining the original estimate of about 71. Its full open-tag counts are 1,006 -> 987 raw and 854 -> 782 distinct. The kind-predicates counts are 1,251 -> 1,006 raw and 1,040 -> 854 distinct. Its one new finding is an exposed reason at an already blocked site. Both comparisons have zero first-stop changes and zero panics. With normalization disabled, qeanycw reports 54 new sites, reproducing the erroneous regression count (55 new raw findings, one at an already blocked site).

Run `python3 -E -S -B -m unittest discover -s stage3/census/pair -p 'test_*.py'`. The Go wrapper is `TestCensusPairPython` in `stage3/lane/python_test.go`; `TestEveryStage3PythonSuiteRunsOrSaysWhy` checks its registration. Tests use the checkout's `stage3/api/node_modules/typescript`, installed from its lockfile. They never build or run the compiler.

The fixtures retain findings at every changed raw identity and all findings sharing those normalized chains, including unchanged duplicates needed to pair relocated reports. Unaffected sites are omitted, so fixture before/after totals are deliberately smaller than the full-stream totals above. Each source fixture contains verbatim AST expression slices separated by parentheses and semicolons; dummy identifiers preserve distinct unit locations outside those slices. Finding fields are retained, with locations mapped into slices and text paths made relative. `provenance.json` records original source hashes, the location map, original decompressed stream hashes and fixture stream hashes. Originals are committed under `stage3/census/enum-tag-access/evidence` on `364c536a` and `stage3/census/kind-predicates/evidence` on `33dca393`. The adapted source was generated from `b03fe07e` and checked against the historical source manifest. The normalization-off mutant test requires the same zero-new assertion to turn red with exactly 54 new sites.

The element-argument boundary is covered by synthetic two-read, receiver-chain,
relocation and newly exposed regression cases. `normalize-element-argument.mutant.cjs`
restores the old ancestor walk; the suite requires the two-read test to fail with
`1 != 2` under that mutant.

Re-pairing both recorded fixtures with the old walk and the argument boundary
produced identical complete JSON deltas. For qeanycw, fixture distinct totals
remain 131 before and 59 after, with 70 lowered, 2 moved, 0 new findings,
0 new sites and 0 regressions. For kind-predicates, totals remain 250 before
and 65 after, with 181 lowered, 5 moved, 1 new finding at an existing site,
0 new sites and 0 regressions. Raw counts, reason counts, unit first stops and
panics are also unchanged. These are fixture results, not a remeasurement of
the complete historical streams.
