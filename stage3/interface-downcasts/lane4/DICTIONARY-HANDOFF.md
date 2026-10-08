# Lane 6 primitive/nullish dictionary selection handoff

Built: ranked candidate inventory only; no runtime completion credited.
Commit: this report is committed with the lane 4 handoff checkpoint.
Checks: reproduced source rows and verified pair-id deduplication and 11 pairs / 38 reads pending.
Mutants: none added; this checkpoint changes scheduling metadata, not compiler checks.
Uncovered: dictionary extraction and exact production reachability remain unproved.

Source: codex/views-dictionaries ce4eeaa45750a49d1fd44f0881473513d3d82eb8,
GROUP4.md and candidate-pair-progress.json. Its inventory uses the same lazy census
SHA256 as this lane. These are static candidate counts, not exact viewed allocation
reachability. No individual lane branch was merged.

## Pending queue by candidate reads

| Pair | Declared type | Reads | Ownership |
| --- | --- | ---: | --- |
| 9858:<dynamic-key> | TsConfigSourceFile \| CompilerOptionsValue | 10 | Lane 4 scalar/nullish selector; lanes 6 and 4b also required |
| 6849:<element> | string \| number | 8 | Lane 4 |
| 46428:value | string \| number | 6 | Lane 4 |
| 9761:skippedOn | keyof CompilerOptions \| undefined | 3 | Lane 4 |
| 97934:pendingEmit | IncrementalBuildInfoBundlePendingEmit \| undefined | 3 | Lane 4 |
| 37515:peerDependencies | string \| false \| undefined | 2 | Lane 4 |
| 97892:signature | string \| false \| undefined | 2 | Lane 4 |
| 64997:<dynamic-key> | CompilerOptionsValue | 1 | Lane 4 scalar/nullish selector; lanes 6 and 4b also required |
| 6995:constantValue | string \| number \| undefined | 1 | Lane 4 |
| 97180:skippedOn | keyof CompilerOptions \| undefined | 1 | Lane 4 |
| 97923:forEach | (callbackfn: (value: IncrementalMultiFileEmitBuildInfoFileInfo, index: number, array: readonly IncrementalMultiFileEmitBuildInfoFileInfo[]) => void, thisArg?: any) => void | 1 | Lane 4 plus object/intersection consumer adapters |

The handoff adds two distinct pairs / eleven reads without duplicating existing
(type_id, field) pairs. __String remains complete at fourteen candidate pairs /
five hundred eleven reads. Original mixed primitive unions remain nine pairs /
twenty-seven reads pending; the combined scheduling queue is eleven / thirty-eight.
Family counts overlap and must not be summed as whole-program coverage.

CompilerOptionsValue includes string, number, boolean, null and undefined plus
arrays and dictionaries. CompilerOptions adds TsConfigSourceFile. Lane 4 owns
only scalar/nullish selection at extraction. Lane 6 owns lookup, absence, storage
and enumeration; lane 4b owns object/array selection and descendant contracts.
The 111-read ParsedCommandLine.options container certification supplies no credit
for the ten dynamic child reads. Both handed-off pairs stay pending until their
full extraction path is held to Node in both backends with wrong-value pins.

The twenty any reads, fourteen string reads and three Path reads are not
primitive/nullish union selection. WatchDirectoryFlags is scalar-only; nullish
object and dictionary container rows belong to their existing owners. These rows
are excluded from this handoff, rather than counted as completed.

Delivery estimate: the earlier October 9, 17:00 MDT target was conditional on
array-adapter approval and did not include these shared dictionary pairs. It is
not a promise for the enlarged queue. A full-family date remains conditional on
tested dictionary extraction and object/array alternatives through integration;
no unconditional date is supported by this inventory-only checkpoint.
