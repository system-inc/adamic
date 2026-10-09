# TypeScript 6.0.3 allocation observations

Exact constructor calls and snapshot live objects are separate from sampled allocation estimates.

| Input | Observed peak heap MiB | Peak phase | Replay pre-snapshot heap MiB | Replay post-GC heap MiB | After check post-GC heap MiB | Released post-GC heap MiB |
|---|---:|---|---:|---:|---:|---:|
| compiler | 547.66 | check | 547.71 | 422.12 | 422.20 | 4.78 |
| mitt | 116.84 | check | 112.41 | 58.17 | 58.16 | 4.44 |

## compiler

| Phase | Constructor | Exact calls | New live at boundary | New live self KiB | Survive check | Dead before boundary | Gone by check | Survive release |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| parse | Node | 1084449 | 1061497 | 149593.3 | 1061497 | 22952 | 0 | 0 |
| parse | Symbol | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| parse | Type | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| parse | Signature | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| parse | Array | N/A | 163054 | 5095.4 | 162891 | N/A | 163 | 0 |
| parse | Map | N/A | 1017 | 31.8 | 1017 | N/A | 0 | 0 |
| parse | String | N/A | 95224 | 20552.8 | 95063 | N/A | 161 | 2 |
| parse | Array backing | N/A | 151925 | 8410.3 | 151762 | N/A | 163 | 1 |
| bind | Node | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| bind | Symbol | 122977 | 122977 | 16332.9 | 122977 | 0 | 0 | 0 |
| bind | Type | 83 | 83 | 5.2 | 83 | 0 | 0 | 0 |
| bind | Signature | 4 | 4 | 0.4 | 4 | 0 | 0 | 0 |
| bind | Array | N/A | 138135 | 4316.7 | 138133 | N/A | 2 | 0 |
| bind | Map | N/A | 95497 | 2984.3 | 95497 | N/A | 0 | 0 |
| bind | String | N/A | 3523 | 139.9 | 3512 | N/A | 11 | 0 |
| bind | Array backing | N/A | 233822 | 21947.8 | 233794 | N/A | 28 | 0 |
| check | Node | 42411 | 12781 | 1930.5 | 12781 | 29630 | 0 | 0 |
| check | Symbol | 142166 | 137670 | 18284.3 | 137670 | 4496 | 0 | 0 |
| check | Type | 107337 | 104572 | 6535.7 | 104572 | 2765 | 0 | 0 |
| check | Signature | 42622 | 41544 | 4219.3 | 41544 | 1078 | 0 | 0 |
| check | Array | N/A | 179660 | 5614.4 | 179660 | N/A | 0 | 0 |
| check | Map | N/A | 45605 | 1425.2 | 45605 | N/A | 0 | 0 |
| check | String | N/A | 298677 | 9665.5 | 298677 | N/A | 0 | 0 |
| check | Array backing | N/A | 207455 | 58469.9 | 207455 | N/A | 0 | 0 |

Whole live graph by requested constructor after checking (shallow self bytes):

| Constructor | Live count | Self KiB | Live after Program release | Self KiB after release |
|---|---:|---:|---:|---:|
| Node | 1074373 | 151538.1 | 0 | 0.0 |
| Symbol | 260647 | 34617.2 | 0 | 0.0 |
| Type | 104655 | 6540.9 | 0 | 0.0 |
| Signature | 41548 | 4219.7 | 0 | 0.0 |
| Array | 481143 | 15035.8 | 201 | 6.3 |
| Map | 142249 | 4445.3 | 46 | 1.4 |
| String | 426054 | 45776.0 | 14414 | 2569.0 |
| Array backing | 594423 | 89921.8 | 1206 | 807.3 |
| Other | 1555062 | 88463.0 | 49048 | 3587.2 |

Snapshot-free inspector sampling including collected objects: 3690.78 MiB estimated, 220905 sample records.
Unmodified CLI --heap-prof: 408.84 MiB estimated in the exit profile, 24007 sample records.

| Allocation stack attribution | Estimated MiB, including collected | Sample records |
|---|---:|---:|
| unclassified | 3654.52 | 218605 |
| Node-associated stack | 21.41 | 1362 |
| Symbol-associated stack | 7.08 | 451 |
| Signature-associated stack | 1.99 | 127 |
| Type-associated stack | 5.78 | 360 |

## mitt

| Phase | Constructor | Exact calls | New live at boundary | New live self KiB | Survive check | Dead before boundary | Gone by check | Survive release |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| parse | Node | 135372 | 130936 | 17998.4 | 130936 | 4436 | 0 | 0 |
| parse | Symbol | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| parse | Type | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| parse | Signature | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| parse | Array | N/A | 22322 | 697.6 | 22322 | N/A | 0 | 0 |
| parse | Map | N/A | 138 | 4.3 | 138 | N/A | 0 | 0 |
| parse | String | N/A | 12223 | 5472.1 | 12219 | N/A | 4 | 0 |
| parse | Array backing | N/A | 20725 | 1086.8 | 20725 | N/A | 0 | 1 |
| bind | Node | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| bind | Symbol | 30750 | 30750 | 4084.0 | 30750 | 0 | 0 | 0 |
| bind | Type | 85 | 85 | 5.3 | 85 | 0 | 0 | 0 |
| bind | Signature | 4 | 4 | 0.4 | 4 | 0 | 0 | 0 |
| bind | Array | N/A | 30818 | 963.1 | 30816 | N/A | 2 | 0 |
| bind | Map | N/A | 22561 | 705.0 | 22561 | N/A | 0 | 0 |
| bind | String | N/A | 2763 | 115.9 | 2759 | N/A | 4 | 0 |
| bind | Array backing | N/A | 53387 | 4926.9 | 53379 | N/A | 8 | 0 |
| check | Node | 0 | 0 | 0.0 | 0 | 0 | 0 | 0 |
| check | Symbol | 567 | 558 | 74.1 | 558 | 9 | 0 | 0 |
| check | Type | 369 | 369 | 23.1 | 369 | 0 | 0 | 0 |
| check | Signature | 137 | 133 | 13.5 | 133 | 4 | 0 | 0 |
| check | Array | N/A | 709 | 22.2 | 709 | N/A | 0 | 0 |
| check | Map | N/A | 184 | 5.8 | 184 | N/A | 0 | 0 |
| check | String | N/A | 976 | 32.1 | 976 | N/A | 0 | 0 |
| check | Array backing | N/A | 881 | 174.5 | 881 | N/A | 0 | 0 |

Whole live graph by requested constructor after checking (shallow self bytes):

| Constructor | Live count | Self KiB | Live after Program release | Self KiB after release |
|---|---:|---:|---:|---:|
| Node | 130962 | 18002.5 | 0 | 0.0 |
| Symbol | 31308 | 4158.1 | 0 | 0.0 |
| Type | 454 | 28.4 | 0 | 0.0 |
| Signature | 137 | 13.9 | 0 | 0.0 |
| Array | 54285 | 1696.5 | 201 | 6.3 |
| Map | 23012 | 719.1 | 46 | 1.4 |
| String | 44550 | 21029.9 | 14410 | 2378.9 |
| Array backing | 76381 | 7288.1 | 1207 | 807.7 |
| Other | 163523 | 13284.0 | 48846 | 3424.8 |

Snapshot-free inspector sampling including collected objects: 127.43 MiB estimated, 3783 sample records.
Unmodified CLI --heap-prof: 47.26 MiB estimated in the exit profile, 2485 sample records.

| Allocation stack attribution | Estimated MiB, including collected | Sample records |
|---|---:|---:|
| unresolved sample nodeId | 1.10 | 4 |
| unclassified | 120.89 | 3433 |
| Symbol-associated stack | 4.05 | 258 |
| Signature-associated stack | 0.02 | 1 |
| Type-associated stack | 0.06 | 4 |
| Node-associated stack | 1.30 | 83 |

N/A is not zero. Sampling carries allocation call stacks, not allocated-object constructors or exact total counts. Stack attribution is heuristic and is not constructor-byte accounting.

Cohorts are objects newly observed at a post-GC phase boundary. The end is checking complete with the Program still rooted; released is after its compilation frame has unwound. Birth self bytes need not equal end sizes. Backing arrays are V8 storage nodes, separate from JavaScript Array objects; Map storage is not included in Map self bytes.

Peak is the highest heapUsed observation at 256-constructor intervals and phase boundaries in a run without snapshots. The peak snapshot is replayed at the same deterministic constructor event and boundary location; replay GC and heap usage can differ. It is not proof of the continuous global maximum.
