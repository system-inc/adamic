# Completed MultiMap call count

398 of 398 lane units, recorder 3. Original and free helper bodies execute on each input; returned array contents and full ordered Map contents agree after every call.

| Operation and initial bucket | Calls | Differences |
| --- | ---: | ---: |
| add, missing key | 3,479,128 | 0 |
| add, singleton, value absent | 417,178 | 0 |
| add, multiple, value absent | 456,069 | 0 |
| add, singleton, value present | 27 | 0 |
| add, multiple, value present | 10 | 0 |
| remove, singleton, value present | 3,416 | 0 |
| remove, multiple, value present | 9 | 0 |
| Total | 4,355,837 | 0 |

Adds: 4,352,412. Removes: 3,425. Factory Maps registered: 850,706. Own-key enumerations: 0. Sparse buckets: 0. Empty buckets, missing-key removes and absent-value removes: 0 real calls. The 18 Node witnesses cover each occurring bucket class and these zero-count edge classes; all have equal returns and Map contents. The only own-key difference is ["add", "remove"] versus [], witnessed explicitly.

The unchanged lane guards accept 106,366 passing tests, the single sanctioned API failure and zero skips. The final ordinary lane matches this exact test union and baseline diff. Raw per-unit reports, observations, recorder files and excluded attempts are in count-evidence.tar.gz; SHA256SUMS is inside and count-evidence.sha256 authenticates the archive. MEASURE.md explains snapshots, input manifests, retries and exit-listener completeness.
